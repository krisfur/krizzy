package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type SQLiteCommentRepository struct {
	db *sql.DB
}

func NewSQLiteCommentRepository(db *sql.DB) *SQLiteCommentRepository {
	return &SQLiteCommentRepository{db: db}
}

func (r *SQLiteCommentRepository) GetByID(id int64) (*models.Comment, error) {
	comment := &models.Comment{}
	err := r.db.QueryRow(
		"SELECT id, card_id, content, created_at FROM comments WHERE id = ?",
		id,
	).Scan(&comment.ID, &comment.CardID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *SQLiteCommentRepository) GetByCardID(cardID int64) ([]models.Comment, error) {
	rows, err := r.db.Query(
		"SELECT id, card_id, content, created_at FROM comments WHERE card_id = ? ORDER BY created_at DESC",
		cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.CardID, &comment.Content, &comment.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, rows.Err()
}

func (r *SQLiteCommentRepository) Create(comment *models.Comment) error {
	result, err := r.db.Exec(
		"INSERT INTO comments (card_id, content) VALUES (?, ?)",
		comment.CardID, comment.Content,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	comment.ID = id
	return nil
}

func (r *SQLiteCommentRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM comments WHERE id = ?", id)
	return err
}
