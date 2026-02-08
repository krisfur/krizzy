package repository

import (
	"database/sql"
	"krizzy/internal/models"
	"time"
)

type PgCardRepository struct {
	db *sql.DB
}

func NewPgCardRepository(db *sql.DB) *PgCardRepository {
	return &PgCardRepository{db: db}
}

func (r *PgCardRepository) GetByID(id int64) (*models.Card, error) {
	card := &models.Card{}
	var completedAt sql.NullTime
	var description sql.NullString
	err := r.db.QueryRow(
		"SELECT id, column_id, title, description, position, completed_at, created_at, updated_at FROM cards WHERE id = $1",
		id,
	).Scan(&card.ID, &card.ColumnID, &card.Title, &description, &card.Position, &completedAt, &card.CreatedAt, &card.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if completedAt.Valid {
		card.CompletedAt = &completedAt.Time
	}
	if description.Valid {
		card.Description = description.String
	}
	return card, nil
}

func (r *PgCardRepository) GetByColumnID(columnID int64) ([]models.Card, error) {
	rows, err := r.db.Query(
		"SELECT id, column_id, title, description, position, completed_at, created_at, updated_at FROM cards WHERE column_id = $1 ORDER BY position",
		columnID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var card models.Card
		var completedAt sql.NullTime
		var description sql.NullString
		if err := rows.Scan(&card.ID, &card.ColumnID, &card.Title, &description, &card.Position, &completedAt, &card.CreatedAt, &card.UpdatedAt); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			card.CompletedAt = &completedAt.Time
		}
		if description.Valid {
			card.Description = description.String
		}
		cards = append(cards, card)
	}
	return cards, rows.Err()
}

func (r *PgCardRepository) Create(card *models.Card) error {
	maxPos, err := r.GetMaxPosition(card.ColumnID)
	if err != nil {
		return err
	}
	card.Position = maxPos + 1

	err = r.db.QueryRow(
		"INSERT INTO cards (column_id, title, description, position) VALUES ($1, $2, $3, $4) RETURNING id",
		card.ColumnID, card.Title, card.Description, card.Position,
	).Scan(&card.ID)
	return err
}

func (r *PgCardRepository) Update(card *models.Card) error {
	_, err := r.db.Exec(
		"UPDATE cards SET title = $1, description = $2, completed_at = $3, updated_at = $4 WHERE id = $5",
		card.Title, card.Description, card.CompletedAt, time.Now(), card.ID,
	)
	return err
}

func (r *PgCardRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM cards WHERE id = $1", id)
	return err
}

func (r *PgCardRepository) Move(cardID int64, newColumnID int64, newPosition int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var oldColumnID int64
	var oldPosition int
	err = tx.QueryRow(
		"SELECT column_id, position FROM cards WHERE id = $1",
		cardID,
	).Scan(&oldColumnID, &oldPosition)
	if err != nil {
		return err
	}

	if oldColumnID == newColumnID {
		if oldPosition < newPosition {
			_, err = tx.Exec(
				"UPDATE cards SET position = position - 1 WHERE column_id = $1 AND position > $2 AND position <= $3",
				oldColumnID, oldPosition, newPosition,
			)
		} else if oldPosition > newPosition {
			_, err = tx.Exec(
				"UPDATE cards SET position = position + 1 WHERE column_id = $1 AND position >= $2 AND position < $3",
				oldColumnID, newPosition, oldPosition,
			)
		}
		if err != nil {
			return err
		}
	} else {
		_, err = tx.Exec(
			"UPDATE cards SET position = position - 1 WHERE column_id = $1 AND position > $2",
			oldColumnID, oldPosition,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(
			"UPDATE cards SET position = position + 1 WHERE column_id = $1 AND position >= $2",
			newColumnID, newPosition,
		)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(
		"UPDATE cards SET column_id = $1, position = $2, updated_at = $3 WHERE id = $4",
		newColumnID, newPosition, time.Now(), cardID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PgCardRepository) GetMaxPosition(columnID int64) (int, error) {
	var maxPos sql.NullInt64
	err := r.db.QueryRow(
		"SELECT MAX(position) FROM cards WHERE column_id = $1",
		columnID,
	).Scan(&maxPos)
	if err != nil {
		return -1, err
	}
	if maxPos.Valid {
		return int(maxPos.Int64), nil
	}
	return -1, nil
}
