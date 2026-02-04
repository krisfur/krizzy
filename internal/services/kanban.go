package services

import (
	"krizzy/internal/models"
	"krizzy/internal/repository"
	"time"
)

type KanbanService struct {
	boardRepo     repository.BoardRepository
	columnRepo    repository.ColumnRepository
	cardRepo      repository.CardRepository
	personRepo    repository.PersonRepository
	commentRepo   repository.CommentRepository
	checklistRepo repository.ChecklistRepository
}

func NewKanbanService(
	boardRepo repository.BoardRepository,
	columnRepo repository.ColumnRepository,
	cardRepo repository.CardRepository,
	personRepo repository.PersonRepository,
	commentRepo repository.CommentRepository,
	checklistRepo repository.ChecklistRepository,
) *KanbanService {
	return &KanbanService{
		boardRepo:     boardRepo,
		columnRepo:    columnRepo,
		cardRepo:      cardRepo,
		personRepo:    personRepo,
		commentRepo:   commentRepo,
		checklistRepo: checklistRepo,
	}
}

// GetBoardWithData returns a board with all its columns and cards
func (s *KanbanService) GetBoardWithData(boardID int64) (*models.Board, error) {
	board, err := s.boardRepo.GetByID(boardID)
	if err != nil {
		return nil, err
	}

	columns, err := s.columnRepo.GetByBoardID(boardID)
	if err != nil {
		return nil, err
	}

	for i := range columns {
		cards, err := s.cardRepo.GetByColumnID(columns[i].ID)
		if err != nil {
			return nil, err
		}
		// Load assignees for each card
		for j := range cards {
			assignees, err := s.personRepo.GetByCardID(cards[j].ID)
			if err != nil {
				return nil, err
			}
			cards[j].Assignees = assignees
		}
		columns[i].Cards = cards
	}

	board.Columns = columns
	return board, nil
}

// GetDefaultBoard returns the default board with all data
func (s *KanbanService) GetDefaultBoard() (*models.Board, error) {
	board, err := s.boardRepo.GetDefault()
	if err != nil {
		return nil, err
	}
	return s.GetBoardWithData(board.ID)
}

// MoveCard moves a card to a new column/position and handles Done column automation
func (s *KanbanService) MoveCard(cardID int64, newColumnID int64, newPosition int) error {
	column, err := s.columnRepo.GetByID(newColumnID)
	if err != nil {
		return err
	}

	card, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return err
	}

	// Handle Done column automation
	if column.IsDoneColumn {
		now := time.Now()
		card.CompletedAt = &now
	} else {
		card.CompletedAt = nil
	}

	// Update the card's completed_at
	if err := s.cardRepo.Update(card); err != nil {
		return err
	}

	// Move the card to new position
	return s.cardRepo.Move(cardID, newColumnID, newPosition)
}

// GetCardWithDetails returns a card with all its details (assignees, comments, checklist)
func (s *KanbanService) GetCardWithDetails(cardID int64) (*models.Card, error) {
	card, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}

	assignees, err := s.personRepo.GetByCardID(cardID)
	if err != nil {
		return nil, err
	}
	card.Assignees = assignees

	comments, err := s.commentRepo.GetByCardID(cardID)
	if err != nil {
		return nil, err
	}
	card.Comments = comments

	checklist, err := s.checklistRepo.GetByCardID(cardID)
	if err != nil {
		return nil, err
	}
	card.Checklist = checklist

	return card, nil
}

// CreateDefaultBoard creates the default board with initial columns
func (s *KanbanService) CreateDefaultBoard() (*models.Board, error) {
	board := &models.Board{Name: "My Board"}
	if err := s.boardRepo.Create(board); err != nil {
		return nil, err
	}

	columns := []struct {
		name       string
		isDoneCol  bool
	}{
		{"To Do", false},
		{"In Progress", false},
		{"Done", true},
	}

	for _, col := range columns {
		column := &models.Column{
			BoardID:      board.ID,
			Name:         col.name,
			IsDoneColumn: col.isDoneCol,
		}
		if err := s.columnRepo.Create(column); err != nil {
			return nil, err
		}
	}

	return s.GetBoardWithData(board.ID)
}

// EnsureDefaultBoard ensures a default board exists, creating one if needed
func (s *KanbanService) EnsureDefaultBoard() (*models.Board, error) {
	board, err := s.boardRepo.GetDefault()
	if err != nil {
		// Create default board if none exists
		return s.CreateDefaultBoard()
	}
	return s.GetBoardWithData(board.ID)
}
