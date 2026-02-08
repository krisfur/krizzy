package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type PgColumnRepository struct {
	db      *sql.DB
	boardID int64
}

func NewPgColumnRepository(db *sql.DB, boardID int64) *PgColumnRepository {
	return &PgColumnRepository{db: db, boardID: boardID}
}

func (r *PgColumnRepository) GetByID(id int64) (*models.Column, error) {
	column := &models.Column{}
	err := r.db.QueryRow(
		"SELECT id, board_id, name, position, is_done_column, created_at FROM columns WHERE id = $1",
		id,
	).Scan(&column.ID, &column.BoardID, &column.Name, &column.Position, &column.IsDoneColumn, &column.CreatedAt)
	if err != nil {
		return nil, err
	}
	return column, nil
}

func (r *PgColumnRepository) GetByBoardID(boardID int64) ([]models.Column, error) {
	rows, err := r.db.Query(
		"SELECT id, board_id, name, position, is_done_column, created_at FROM columns WHERE board_id = $1 ORDER BY position",
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

func (r *PgColumnRepository) Create(column *models.Column) error {
	var maxPos sql.NullInt64
	err := r.db.QueryRow(
		"SELECT MAX(position) FROM columns WHERE board_id = $1",
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

	err = r.db.QueryRow(
		"INSERT INTO columns (board_id, name, position, is_done_column) VALUES ($1, $2, $3, $4) RETURNING id",
		column.BoardID, column.Name, column.Position, column.IsDoneColumn,
	).Scan(&column.ID)
	return err
}

func (r *PgColumnRepository) Update(column *models.Column) error {
	_, err := r.db.Exec(
		"UPDATE columns SET name = $1, is_done_column = $2 WHERE id = $3",
		column.Name, column.IsDoneColumn, column.ID,
	)
	return err
}

func (r *PgColumnRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM columns WHERE id = $1", id)
	return err
}

func (r *PgColumnRepository) Reorder(boardID int64, columnIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range columnIDs {
		_, err := tx.Exec(
			"UPDATE columns SET position = $1 WHERE id = $2 AND board_id = $3",
			i, id, boardID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
