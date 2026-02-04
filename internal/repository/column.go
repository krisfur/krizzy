package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type SQLiteColumnRepository struct {
	db *sql.DB
}

func NewSQLiteColumnRepository(db *sql.DB) *SQLiteColumnRepository {
	return &SQLiteColumnRepository{db: db}
}

func (r *SQLiteColumnRepository) GetByID(id int64) (*models.Column, error) {
	column := &models.Column{}
	err := r.db.QueryRow(
		"SELECT id, board_id, name, position, is_done_column, created_at FROM columns WHERE id = ?",
		id,
	).Scan(&column.ID, &column.BoardID, &column.Name, &column.Position, &column.IsDoneColumn, &column.CreatedAt)
	if err != nil {
		return nil, err
	}
	return column, nil
}

func (r *SQLiteColumnRepository) GetByBoardID(boardID int64) ([]models.Column, error) {
	rows, err := r.db.Query(
		"SELECT id, board_id, name, position, is_done_column, created_at FROM columns WHERE board_id = ? ORDER BY position",
		boardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []models.Column
	for rows.Next() {
		var column models.Column
		if err := rows.Scan(&column.ID, &column.BoardID, &column.Name, &column.Position, &column.IsDoneColumn, &column.CreatedAt); err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}
	return columns, rows.Err()
}

func (r *SQLiteColumnRepository) Create(column *models.Column) error {
	// Get max position for the board
	var maxPos sql.NullInt64
	err := r.db.QueryRow(
		"SELECT MAX(position) FROM columns WHERE board_id = ?",
		column.BoardID,
	).Scan(&maxPos)
	if err != nil {
		return err
	}

	newPos := 0
	if maxPos.Valid {
		newPos = int(maxPos.Int64) + 1
	}
	column.Position = newPos

	result, err := r.db.Exec(
		"INSERT INTO columns (board_id, name, position, is_done_column) VALUES (?, ?, ?, ?)",
		column.BoardID, column.Name, column.Position, column.IsDoneColumn,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	column.ID = id
	return nil
}

func (r *SQLiteColumnRepository) Update(column *models.Column) error {
	_, err := r.db.Exec(
		"UPDATE columns SET name = ?, is_done_column = ? WHERE id = ?",
		column.Name, column.IsDoneColumn, column.ID,
	)
	return err
}

func (r *SQLiteColumnRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM columns WHERE id = ?", id)
	return err
}

func (r *SQLiteColumnRepository) Reorder(boardID int64, columnIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range columnIDs {
		_, err := tx.Exec(
			"UPDATE columns SET position = ? WHERE id = ? AND board_id = ?",
			i, id, boardID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
