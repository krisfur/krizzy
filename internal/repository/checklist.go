package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type SQLiteChecklistRepository struct {
	db *sql.DB
}

func NewSQLiteChecklistRepository(db *sql.DB) *SQLiteChecklistRepository {
	return &SQLiteChecklistRepository{db: db}
}

func (r *SQLiteChecklistRepository) GetByID(id int64) (*models.ChecklistItem, error) {
	item := &models.ChecklistItem{}
	err := r.db.QueryRow(
		"SELECT id, card_id, content, is_completed, position, created_at FROM checklist_items WHERE id = ?",
		id,
	).Scan(&item.ID, &item.CardID, &item.Content, &item.IsCompleted, &item.Position, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *SQLiteChecklistRepository) GetByCardID(cardID int64) ([]models.ChecklistItem, error) {
	rows, err := r.db.Query(
		"SELECT id, card_id, content, is_completed, position, created_at FROM checklist_items WHERE card_id = ? ORDER BY position",
		cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ChecklistItem
	for rows.Next() {
		var item models.ChecklistItem
		if err := rows.Scan(&item.ID, &item.CardID, &item.Content, &item.IsCompleted, &item.Position, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SQLiteChecklistRepository) Create(item *models.ChecklistItem) error {
	maxPos, err := r.GetMaxPosition(item.CardID)
	if err != nil {
		return err
	}
	item.Position = maxPos + 1

	result, err := r.db.Exec(
		"INSERT INTO checklist_items (card_id, content, is_completed, position) VALUES (?, ?, ?, ?)",
		item.CardID, item.Content, item.IsCompleted, item.Position,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	item.ID = id
	return nil
}

func (r *SQLiteChecklistRepository) Update(item *models.ChecklistItem) error {
	_, err := r.db.Exec(
		"UPDATE checklist_items SET content = ?, is_completed = ? WHERE id = ?",
		item.Content, item.IsCompleted, item.ID,
	)
	return err
}

func (r *SQLiteChecklistRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM checklist_items WHERE id = ?", id)
	return err
}

func (r *SQLiteChecklistRepository) Reorder(cardID int64, itemIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range itemIDs {
		_, err := tx.Exec(
			"UPDATE checklist_items SET position = ? WHERE id = ? AND card_id = ?",
			i, id, cardID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SQLiteChecklistRepository) GetMaxPosition(cardID int64) (int, error) {
	var maxPos sql.NullInt64
	err := r.db.QueryRow(
		"SELECT MAX(position) FROM checklist_items WHERE card_id = ?",
		cardID,
	).Scan(&maxPos)
	if err != nil {
		return -1, err
	}
	if maxPos.Valid {
		return int(maxPos.Int64), nil
	}
	return -1, nil
}
