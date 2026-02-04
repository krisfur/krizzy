package database

import "database/sql"

type Database interface {
	DB() *sql.DB
	Close() error
	Migrate() error
}
