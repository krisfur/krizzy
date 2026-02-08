package services

import (
	"database/sql"
	"fmt"
	"regexp"
	"sync"

	"krizzy/internal/database"
	"krizzy/internal/models"
	"krizzy/internal/repository"

	_ "github.com/lib/pq"
)

var pgDbNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,62}$`)

type BoardManager struct {
	localDB    database.Database
	boardRepo  repository.BoardRepository
	pgConnRepo repository.PgConnectionRepository
	mu         sync.RWMutex
	services   map[int64]*KanbanService
	pgDBs      map[int64]database.Database
}

func NewBoardManager(localDB database.Database, boardRepo repository.BoardRepository, pgConnRepo repository.PgConnectionRepository) *BoardManager {
	return &BoardManager{
		localDB:    localDB,
		boardRepo:  boardRepo,
		pgConnRepo: pgConnRepo,
		services:   make(map[int64]*KanbanService),
		pgDBs:      make(map[int64]database.Database),
	}
}

func (bm *BoardManager) BoardRepo() repository.BoardRepository {
	return bm.boardRepo
}

func (bm *BoardManager) PgConnRepo() repository.PgConnectionRepository {
	return bm.pgConnRepo
}

func (bm *BoardManager) GetAllBoards() ([]models.Board, error) {
	return bm.boardRepo.GetAll()
}

func (bm *BoardManager) GetBoard(id int64) (*models.Board, error) {
	return bm.boardRepo.GetByID(id)
}

func (bm *BoardManager) GetServiceForBoard(boardID int64) (*KanbanService, error) {
	bm.mu.RLock()
	if svc, ok := bm.services[boardID]; ok {
		bm.mu.RUnlock()
		return svc, nil
	}
	bm.mu.RUnlock()

	bm.mu.Lock()
	defer bm.mu.Unlock()

	// Double-check after acquiring write lock
	if svc, ok := bm.services[boardID]; ok {
		return svc, nil
	}

	board, err := bm.boardRepo.GetByID(boardID)
	if err != nil {
		return nil, fmt.Errorf("board not found: %w", err)
	}

	var svc *KanbanService
	if board.DbType == "postgres" {
		svc, err = bm.createPostgresService(board)
	} else {
		svc, err = bm.createLocalService(board)
	}
	if err != nil {
		return nil, err
	}

	bm.services[boardID] = svc
	return svc, nil
}

func (bm *BoardManager) createLocalService(board *models.Board) (*KanbanService, error) {
	db := bm.localDB.DB()
	return NewKanbanService(
		bm.boardRepo,
		repository.NewSQLiteColumnRepository(db),
		repository.NewSQLiteCardRepository(db),
		repository.NewSQLitePersonRepository(db),
		repository.NewSQLiteCommentRepository(db),
		repository.NewSQLiteChecklistRepository(db),
	), nil
}

func (bm *BoardManager) buildConnString(conn *models.PgConnection, dbName string) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		conn.Host, conn.Port, conn.User, conn.Password, dbName, conn.SSLMode)
}

func (bm *BoardManager) ensureDatabase(conn *models.PgConnection, dbName string) error {
	// Connect to the "postgres" database to create the target database
	adminConn := bm.buildConnString(conn, "postgres")
	adminDB, err := sql.Open("postgres", adminConn)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres server: %w", err)
	}
	defer adminDB.Close()

	// Check if database exists
	var exists bool
	err = adminDB.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		// Database names can't be parameterised, but we've validated the name is alphanumeric+underscores
		_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %q", dbName))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
	}

	return nil
}

func (bm *BoardManager) createPostgresService(board *models.Board) (*KanbanService, error) {
	if board.PgConnectionID == nil {
		return nil, fmt.Errorf("board %d has no postgres connection configured", board.ID)
	}

	conn, err := bm.pgConnRepo.GetByID(*board.PgConnectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to load postgres connection: %w", err)
	}

	// Ensure the database exists
	if err := bm.ensureDatabase(conn, board.PgDatabaseName); err != nil {
		return nil, fmt.Errorf("failed to ensure database for board %d: %w", board.ID, err)
	}

	connString := bm.buildConnString(conn, board.PgDatabaseName)
	pgDB, err := database.NewPostgres(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres for board %d: %w", board.ID, err)
	}

	if err := pgDB.Migrate(); err != nil {
		pgDB.Close()
		return nil, fmt.Errorf("failed to migrate postgres for board %d: %w", board.ID, err)
	}

	bm.pgDBs[board.ID] = pgDB

	db := pgDB.DB()
	return NewKanbanService(
		bm.boardRepo,
		repository.NewPgColumnRepository(db, board.ID),
		repository.NewPgCardRepository(db),
		repository.NewPgPersonRepository(db, board.ID),
		repository.NewPgCommentRepository(db),
		repository.NewPgChecklistRepository(db),
	), nil
}

func (bm *BoardManager) CreateBoard(name, dbType string, pgConnectionID *int64, pgDatabaseName string) (*models.Board, error) {
	board := &models.Board{
		Name:           name,
		DbType:         dbType,
		PgConnectionID: pgConnectionID,
		PgDatabaseName: pgDatabaseName,
	}

	if dbType == "postgres" {
		if pgConnectionID == nil {
			return nil, fmt.Errorf("postgres connection is required")
		}
		if !pgDbNameRegex.MatchString(pgDatabaseName) {
			return nil, fmt.Errorf("invalid database name: must be alphanumeric with underscores, max 63 chars")
		}
		// Verify connection exists
		if _, err := bm.pgConnRepo.GetByID(*pgConnectionID); err != nil {
			return nil, fmt.Errorf("postgres connection not found")
		}
	}

	if err := bm.boardRepo.Create(board); err != nil {
		return nil, err
	}

	svc, err := bm.GetServiceForBoard(board.ID)
	if err != nil {
		// Clean up the board entry if service creation fails
		bm.boardRepo.Delete(board.ID)
		return nil, err
	}

	if err := svc.CreateDefaultColumns(board.ID); err != nil {
		return nil, err
	}

	return board, nil
}

func (bm *BoardManager) RenameBoard(id int64, name string) error {
	board, err := bm.boardRepo.GetByID(id)
	if err != nil {
		return err
	}
	board.Name = name
	return bm.boardRepo.Update(board)
}

func (bm *BoardManager) DeleteBoard(id int64) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if pgDB, ok := bm.pgDBs[id]; ok {
		pgDB.Close()
		delete(bm.pgDBs, id)
	}
	delete(bm.services, id)

	return bm.boardRepo.Delete(id)
}

// InvalidateCache removes the cached service for a board
func (bm *BoardManager) InvalidateCache(boardID int64) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.services, boardID)
}

// Close cleans up all Postgres connections
func (bm *BoardManager) Close() {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	for _, pgDB := range bm.pgDBs {
		pgDB.Close()
	}
}

// TestConnection tests connectivity to a PG server
func (bm *BoardManager) TestConnection(conn *models.PgConnection) error {
	connString := bm.buildConnString(conn, "postgres")
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return fmt.Errorf("failed to open connection: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping: %w", err)
	}

	return nil
}

// HasBoardsUsingConnection checks if any boards reference the given connection
func (bm *BoardManager) HasBoardsUsingConnection(connID int64) (bool, error) {
	boards, err := bm.boardRepo.GetAll()
	if err != nil {
		return false, err
	}
	for _, b := range boards {
		if b.PgConnectionID != nil && *b.PgConnectionID == connID {
			return true, nil
		}
	}
	return false, nil
}
