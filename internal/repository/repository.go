package repository

import "krizzy/internal/models"

type BoardRepository interface {
	GetByID(id int64) (*models.Board, error)
	GetAll() ([]models.Board, error)
	Create(board *models.Board) error
	Update(board *models.Board) error
	Delete(id int64) error
	GetDefault() (*models.Board, error)
}

type ColumnRepository interface {
	GetByID(id int64) (*models.Column, error)
	GetByBoardID(boardID int64) ([]models.Column, error)
	Create(column *models.Column) error
	Update(column *models.Column) error
	Delete(id int64) error
	Reorder(boardID int64, columnIDs []int64) error
}

type CardRepository interface {
	GetByID(id int64) (*models.Card, error)
	GetByColumnID(columnID int64) ([]models.Card, error)
	Create(card *models.Card) error
	Update(card *models.Card) error
	Delete(id int64) error
	Move(cardID int64, newColumnID int64, newPosition int) error
	GetMaxPosition(columnID int64) (int, error)
}

type PersonRepository interface {
	GetByID(id int64) (*models.Person, error)
	GetByBoardID(boardID int64) ([]models.Person, error)
	Create(person *models.Person) error
	Delete(id int64) error
	GetByCardID(cardID int64) ([]models.Person, error)
	SetCardAssignees(cardID int64, personIDs []int64) error
}

type CommentRepository interface {
	GetByID(id int64) (*models.Comment, error)
	GetByCardID(cardID int64) ([]models.Comment, error)
	Create(comment *models.Comment) error
	Delete(id int64) error
}

type ChecklistRepository interface {
	GetByID(id int64) (*models.ChecklistItem, error)
	GetByCardID(cardID int64) ([]models.ChecklistItem, error)
	Create(item *models.ChecklistItem) error
	Update(item *models.ChecklistItem) error
	Delete(id int64) error
	Reorder(cardID int64, itemIDs []int64) error
	GetMaxPosition(cardID int64) (int, error)
}

type PgConnectionRepository interface {
	GetByID(id int64) (*models.PgConnection, error)
	GetAll() ([]models.PgConnection, error)
	Create(conn *models.PgConnection) error
	Update(conn *models.PgConnection) error
	Delete(id int64) error
}
