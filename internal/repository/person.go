package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type SQLitePersonRepository struct {
	db *sql.DB
}

func NewSQLitePersonRepository(db *sql.DB) *SQLitePersonRepository {
	return &SQLitePersonRepository{db: db}
}

func (r *SQLitePersonRepository) GetByID(id int64) (*models.Person, error) {
	person := &models.Person{}
	err := r.db.QueryRow(
		"SELECT id, board_id, name, created_at FROM people WHERE id = ?",
		id,
	).Scan(&person.ID, &person.BoardID, &person.Name, &person.CreatedAt)
	if err != nil {
		return nil, err
	}
	return person, nil
}

func (r *SQLitePersonRepository) GetByBoardID(boardID int64) ([]models.Person, error) {
	rows, err := r.db.Query("SELECT id, board_id, name, created_at FROM people WHERE board_id = ? ORDER BY name", boardID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var person models.Person
		if err := rows.Scan(&person.ID, &person.BoardID, &person.Name, &person.CreatedAt); err != nil {
			return nil, err
		}
		people = append(people, person)
	}
	return people, rows.Err()
}

func (r *SQLitePersonRepository) Create(person *models.Person) error {
	result, err := r.db.Exec(
		"INSERT INTO people (name, board_id) VALUES (?, ?)",
		person.Name, person.BoardID,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	person.ID = id
	return nil
}

func (r *SQLitePersonRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM people WHERE id = ?", id)
	return err
}

func (r *SQLitePersonRepository) GetByCardID(cardID int64) ([]models.Person, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.board_id, p.name, p.created_at
		FROM people p
		JOIN card_assignees ca ON p.id = ca.person_id
		WHERE ca.card_id = ?
		ORDER BY p.name`,
		cardID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var person models.Person
		if err := rows.Scan(&person.ID, &person.BoardID, &person.Name, &person.CreatedAt); err != nil {
			return nil, err
		}
		people = append(people, person)
	}
	return people, rows.Err()
}

func (r *SQLitePersonRepository) SetCardAssignees(cardID int64, personIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Remove all existing assignees
	_, err = tx.Exec("DELETE FROM card_assignees WHERE card_id = ?", cardID)
	if err != nil {
		return err
	}

	// Add new assignees
	for _, personID := range personIDs {
		_, err = tx.Exec(
			"INSERT INTO card_assignees (card_id, person_id) VALUES (?, ?)",
			cardID, personID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
