package repository

import (
	"database/sql"
	"krizzy/internal/models"
)

type SQLitePgConnectionRepository struct {
	db *sql.DB
}

func NewSQLitePgConnectionRepository(db *sql.DB) *SQLitePgConnectionRepository {
	return &SQLitePgConnectionRepository{db: db}
}

func (r *SQLitePgConnectionRepository) GetByID(id int64) (*models.PgConnection, error) {
	conn := &models.PgConnection{}
	err := r.db.QueryRow(
		"SELECT id, name, host, port, username, password, ssl_mode, created_at FROM pg_connections WHERE id = ?",
		id,
	).Scan(&conn.ID, &conn.Name, &conn.Host, &conn.Port, &conn.User, &conn.Password, &conn.SSLMode, &conn.CreatedAt)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (r *SQLitePgConnectionRepository) GetAll() ([]models.PgConnection, error) {
	rows, err := r.db.Query("SELECT id, name, host, port, username, password, ssl_mode, created_at FROM pg_connections ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conns []models.PgConnection
	for rows.Next() {
		var conn models.PgConnection
		if err := rows.Scan(&conn.ID, &conn.Name, &conn.Host, &conn.Port, &conn.User, &conn.Password, &conn.SSLMode, &conn.CreatedAt); err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}
	return conns, rows.Err()
}

func (r *SQLitePgConnectionRepository) Create(conn *models.PgConnection) error {
	if conn.Port == 0 {
		conn.Port = 5432
	}
	if conn.SSLMode == "" {
		conn.SSLMode = "disable"
	}
	result, err := r.db.Exec(
		"INSERT INTO pg_connections (name, host, port, username, password, ssl_mode) VALUES (?, ?, ?, ?, ?, ?)",
		conn.Name, conn.Host, conn.Port, conn.User, conn.Password, conn.SSLMode,
	)
	if err != nil {
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	conn.ID = id
	return nil
}

func (r *SQLitePgConnectionRepository) Update(conn *models.PgConnection) error {
	_, err := r.db.Exec(
		"UPDATE pg_connections SET name = ?, host = ?, port = ?, username = ?, password = ?, ssl_mode = ? WHERE id = ?",
		conn.Name, conn.Host, conn.Port, conn.User, conn.Password, conn.SSLMode, conn.ID,
	)
	return err
}

func (r *SQLitePgConnectionRepository) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM pg_connections WHERE id = ?", id)
	return err
}
