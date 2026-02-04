package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type SQLiteBoardRepository struct {
	db *sql.DB
}

func NewSQLiteBoardRepository(db *sql.DB) *SQLiteBoardRepository {
	return &SQLiteBoardRepository{db: db}
}

func (r *SQLiteBoardRepository) GetByID(id int64) (*models.Board, error) {
	board := &models.Board{}
	err := r.db.QueryRow(
		"SELECT id, name, created_at FROM boards WHERE id = ?",
		id,
	).Scan(&board.ID, &board.Name, &board.CreatedAt)
	if err != nil {
		return nil, err
	}
	return board, nil
}

func (r *SQLiteBoardRepository) GetAll() ([]models.Board, error) {
	rows, err := r.db.Query("SELECT id, name, created_at FROM boards ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var boards []models.Board
	for rows.Next() {
		var board models.Board
		if err := rows.Scan(&board.ID, &board.Name, &board.CreatedAt); err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	return boards, rows.Err()
}

func (r *SQLiteBoardRepository) Create(board *models.Board) error {
	result, err := r.db.Exec(
		"INSERT INTO boards (name) VALUES (?)",
		board.Name,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	board.ID = id
	return nil
}

func (r *SQLiteBoardRepository) Update(board *models.Board) error {
	_, err := r.db.Exec(
		"UPDATE boards SET name = ? WHERE id = ?",
		board.Name, board.ID,
	)
	return err
}

func (r *SQLiteBoardRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM boards WHERE id = ?", id)
	return err
}

func (r *SQLiteBoardRepository) GetDefault() (*models.Board, error) {
	board := &models.Board{}
	err := r.db.QueryRow(
		"SELECT id, name, created_at FROM boards ORDER BY id LIMIT 1",
	).Scan(&board.ID, &board.Name, &board.CreatedAt)
	if err != nil {
		return nil, err
	}
	return board, nil
}
