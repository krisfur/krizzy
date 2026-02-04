package repository

import (
	"database/sql"
	"krizzy/internal/models"
	"time"
)

type SQLiteCardRepository struct {
	db *sql.DB
}

func NewSQLiteCardRepository(db *sql.DB) *SQLiteCardRepository {
	return &SQLiteCardRepository{db: db}
}

func (r *SQLiteCardRepository) GetByID(id int64) (*models.Card, error) {
	card := &models.Card{}
	var completedAt sql.NullTime
	var description sql.NullString
	err := r.db.QueryRow(
		"SELECT id, column_id, title, description, position, completed_at, created_at, updated_at FROM cards WHERE id = ?",
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

func (r *SQLiteCardRepository) GetByColumnID(columnID int64) ([]models.Card, error) {
	rows, err := r.db.Query(
		"SELECT id, column_id, title, description, position, completed_at, created_at, updated_at FROM cards WHERE column_id = ? ORDER BY position",
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

func (r *SQLiteCardRepository) Create(card *models.Card) error {
	maxPos, err := r.GetMaxPosition(card.ColumnID)
	if err != nil {
		return err
	}
	card.Position = maxPos + 1

	result, err := r.db.Exec(
		"INSERT INTO cards (column_id, title, description, position) VALUES (?, ?, ?, ?)",
		card.ColumnID, card.Title, card.Description, card.Position,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	card.ID = id
	return nil
}

func (r *SQLiteCardRepository) Update(card *models.Card) error {
	_, err := r.db.Exec(
		"UPDATE cards SET title = ?, description = ?, completed_at = ?, updated_at = ? WHERE id = ?",
		card.Title, card.Description, card.CompletedAt, time.Now(), card.ID,
	)
	return err
}

func (r *SQLiteCardRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM cards WHERE id = ?", id)
	return err
}

func (r *SQLiteCardRepository) Move(cardID int64, newColumnID int64, newPosition int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current card info
	var oldColumnID int64
	var oldPosition int
	err = tx.QueryRow(
		"SELECT column_id, position FROM cards WHERE id = ?",
		cardID,
	).Scan(&oldColumnID, &oldPosition)
	if err != nil {
		return err
	}

	// If moving within same column
	if oldColumnID == newColumnID {
		if oldPosition < newPosition {
			// Moving down
			_, err = tx.Exec(
				"UPDATE cards SET position = position - 1 WHERE column_id = ? AND position > ? AND position <= ?",
				oldColumnID, oldPosition, newPosition,
			)
		} else if oldPosition > newPosition {
			// Moving up
			_, err = tx.Exec(
				"UPDATE cards SET position = position + 1 WHERE column_id = ? AND position >= ? AND position < ?",
				oldColumnID, newPosition, oldPosition,
			)
		}
		if err != nil {
			return err
		}
	} else {
		// Moving to different column
		// Close gap in old column
		_, err = tx.Exec(
			"UPDATE cards SET position = position - 1 WHERE column_id = ? AND position > ?",
			oldColumnID, oldPosition,
		)
		if err != nil {
			return err
		}

		// Make space in new column
		_, err = tx.Exec(
			"UPDATE cards SET position = position + 1 WHERE column_id = ? AND position >= ?",
			newColumnID, newPosition,
		)
		if err != nil {
			return err
		}
	}

	// Update the card
	_, err = tx.Exec(
		"UPDATE cards SET column_id = ?, position = ?, updated_at = ? WHERE id = ?",
		newColumnID, newPosition, time.Now(), cardID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteCardRepository) GetMaxPosition(columnID int64) (int, error) {
	var maxPos sql.NullInt64
	err := r.db.QueryRow(
		"SELECT MAX(position) FROM cards WHERE column_id = ?",
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
