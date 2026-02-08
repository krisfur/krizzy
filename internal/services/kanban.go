package services

import (
	"krizzy/internal/models"
	"krizzy/internal/repository"
	"time"
)

type KanbanService struct {
	BoardRepo     repository.BoardRepository
	ColumnRepo    repository.ColumnRepository
	CardRepo      repository.CardRepository
	PersonRepo    repository.PersonRepository
	CommentRepo   repository.CommentRepository
	ChecklistRepo repository.ChecklistRepository
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
		BoardRepo:     boardRepo,
		ColumnRepo:    columnRepo,
		CardRepo:      cardRepo,
		PersonRepo:    personRepo,
		CommentRepo:   commentRepo,
		ChecklistRepo: checklistRepo,
	}
}

// GetBoardWithData returns a board with all its columns and cards
func (s *KanbanService) GetBoardWithData(boardID int64) (*models.Board, error) {
	board, err := s.BoardRepo.GetByID(boardID)
	if err != nil {
		return nil, err
	}

	columns, err := s.ColumnRepo.GetByBoardID(boardID)
	if err != nil {
		return nil, err
	}

	for i := range columns {
		cards, err := s.CardRepo.GetByColumnID(columns[i].ID)
		if err != nil {
			return nil, err
		}
		// Load assignees and checklist for each card
		for j := range cards {
			assignees, err := s.PersonRepo.GetByCardID(cards[j].ID)
			if err != nil {
				return nil, err
			}
			cards[j].Assignees = assignees

			checklist, err := s.ChecklistRepo.GetByCardID(cards[j].ID)
			if err != nil {
				return nil, err
			}
			cards[j].Checklist = checklist
		}
		columns[i].Cards = cards
	}

	board.Columns = columns
	return board, nil
}

// MoveCard moves a card to a new column/position and handles Done column automation
func (s *KanbanService) MoveCard(cardID int64, newColumnID int64, newPosition int) error {
	column, err := s.ColumnRepo.GetByID(newColumnID)
	if err != nil {
		return err
	}

	card, err := s.CardRepo.GetByID(cardID)
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
	if err := s.CardRepo.Update(card); err != nil {
		return err
	}

	// Move the card to new position
	return s.CardRepo.Move(cardID, newColumnID, newPosition)
}

// GetCardWithDetails returns a card with all its details (assignees, comments, checklist)
func (s *KanbanService) GetCardWithDetails(cardID int64) (*models.Card, error) {
	card, err := s.CardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}

	assignees, err := s.PersonRepo.GetByCardID(cardID)
	if err != nil {
		return nil, err
	}
	card.Assignees = assignees

	comments, err := s.CommentRepo.GetByCardID(cardID)
	if err != nil {
		return nil, err
	}
	card.Comments = comments

	checklist, err := s.ChecklistRepo.GetByCardID(cardID)
	if err != nil {
		return nil, err
	}
	card.Checklist = checklist

	return card, nil
}

// CreateDefaultColumns creates the default columns for a board
func (s *KanbanService) CreateDefaultColumns(boardID int64) error {
	columns := []struct {
		name      string
		isDoneCol bool
	}{
		{"To Do", false},
		{"In Progress", false},
		{"Done", true},
	}

	for _, col := range columns {
		column := &models.Column{
			BoardID:      boardID,
			Name:         col.name,
			IsDoneColumn: col.isDoneCol,
		}
		if err := s.ColumnRepo.Create(column); err != nil {
			return err
		}
	}

	return nil
}
