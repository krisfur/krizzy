package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type PgCommentRepository struct {
	db *sql.DB
}

func NewPgCommentRepository(db *sql.DB) *PgCommentRepository {
	return &PgCommentRepository{db: db}
}

func (r *PgCommentRepository) GetByID(id int64) (*models.Comment, error) {
	comment := &models.Comment{}
	err := r.db.QueryRow(
		"SELECT id, card_id, content, created_at FROM comments WHERE id = $1",
		id,
	).Scan(&comment.ID, &comment.CardID, &comment.Content, &comment.CreatedAt)
	if err != nil {
		return nil, err
	}
	return comment, nil
}

func (r *PgCommentRepository) GetByCardID(cardID int64) ([]models.Comment, error) {
	rows, err := r.db.Query(
		"SELECT id, card_id, content, created_at FROM comments WHERE card_id = $1 ORDER BY created_at DESC",
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

func (r *PgCommentRepository) Create(comment *models.Comment) error {
	err := r.db.QueryRow(
		"INSERT INTO comments (card_id, content) VALUES ($1, $2) RETURNING id",
		comment.CardID, comment.Content,
	).Scan(&comment.ID)
	return err
}

func (r *PgCommentRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM comments WHERE id = $1", id)
	return err
}
