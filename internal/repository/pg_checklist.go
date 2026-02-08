package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type PgChecklistRepository struct {
	db *sql.DB
}

func NewPgChecklistRepository(db *sql.DB) *PgChecklistRepository {
	return &PgChecklistRepository{db: db}
}

func (r *PgChecklistRepository) GetByID(id int64) (*models.ChecklistItem, error) {
	item := &models.ChecklistItem{}
	err := r.db.QueryRow(
		"SELECT id, card_id, content, is_completed, position, created_at FROM checklist_items WHERE id = $1",
		id,
	).Scan(&item.ID, &item.CardID, &item.Content, &item.IsCompleted, &item.Position, &item.CreatedAt)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PgChecklistRepository) GetByCardID(cardID int64) ([]models.ChecklistItem, error) {
	rows, err := r.db.Query(
		"SELECT id, card_id, content, is_completed, position, created_at FROM checklist_items WHERE card_id = $1 ORDER BY position",
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

func (r *PgChecklistRepository) Create(item *models.ChecklistItem) error {
	maxPos, err := r.GetMaxPosition(item.CardID)
	if err != nil {
		return err
	}
	item.Position = maxPos + 1

	err = r.db.QueryRow(
		"INSERT INTO checklist_items (card_id, content, is_completed, position) VALUES ($1, $2, $3, $4) RETURNING id",
		item.CardID, item.Content, item.IsCompleted, item.Position,
	).Scan(&item.ID)
	return err
}

func (r *PgChecklistRepository) Update(item *models.ChecklistItem) error {
	_, err := r.db.Exec(
		"UPDATE checklist_items SET content = $1, is_completed = $2 WHERE id = $3",
		item.Content, item.IsCompleted, item.ID,
	)
	return err
}

func (r *PgChecklistRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM checklist_items WHERE id = $1", id)
	return err
}

func (r *PgChecklistRepository) Reorder(cardID int64, itemIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range itemIDs {
		_, err := tx.Exec(
			"UPDATE checklist_items SET position = $1 WHERE id = $2 AND card_id = $3",
			i, id, cardID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PgChecklistRepository) GetMaxPosition(cardID int64) (int, error) {
	var maxPos sql.NullInt64
	err := r.db.QueryRow(
		"SELECT MAX(position) FROM checklist_items WHERE card_id = $1",
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
