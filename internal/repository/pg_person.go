package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type PgPersonRepository struct {
	db      *sql.DB
	boardID int64
}

func NewPgPersonRepository(db *sql.DB, boardID int64) *PgPersonRepository {
	return &PgPersonRepository{db: db, boardID: boardID}
}

func (r *PgPersonRepository) GetByID(id int64) (*models.Person, error) {
	person := &models.Person{}
	err := r.db.QueryRow(
		"SELECT id, name, created_at FROM people WHERE id = $1",
		id,
	).Scan(&person.ID, &person.Name, &person.CreatedAt)
	if err != nil {
		return nil, err
	}
	person.BoardID = r.boardID
	return person, nil
}

func (r *PgPersonRepository) GetByBoardID(boardID int64) ([]models.Person, error) {
	// In Postgres mode, all people belong to this board's DB
	rows, err := r.db.Query("SELECT id, name, created_at FROM people ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var people []models.Person
	for rows.Next() {
		var person models.Person
		if err := rows.Scan(&person.ID, &person.Name, &person.CreatedAt); err != nil {
			return nil, err
		}
		person.BoardID = r.boardID
		people = append(people, person)
	}
	return people, rows.Err()
}

func (r *PgPersonRepository) Create(person *models.Person) error {
	err := r.db.QueryRow(
		"INSERT INTO people (name) VALUES ($1) RETURNING id",
		person.Name,
	).Scan(&person.ID)
	if err != nil {
		return err
	}
	person.BoardID = r.boardID
	return nil
}

func (r *PgPersonRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM people WHERE id = $1", id)
	return err
}

func (r *PgPersonRepository) GetByCardID(cardID int64) ([]models.Person, error) {
	rows, err := r.db.Query(
		`SELECT p.id, p.name, p.created_at
		FROM people p
		JOIN card_assignees ca ON p.id = ca.person_id
		WHERE ca.card_id = $1
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
		if err := rows.Scan(&person.ID, &person.Name, &person.CreatedAt); err != nil {
			return nil, err
		}
		person.BoardID = r.boardID
		people = append(people, person)
	}
	return people, rows.Err()
}

func (r *PgPersonRepository) SetCardAssignees(cardID int64, personIDs []int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM card_assignees WHERE card_id = $1", cardID)
	if err != nil {
		return err
	}

	for _, personID := range personIDs {
		_, err = tx.Exec(
			"INSERT INTO card_assignees (card_id, person_id) VALUES ($1, $2)",
			cardID, personID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
