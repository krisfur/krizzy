package models

import "time"

type PgConnection struct {
	ID        int64
	Name      string
	Host      string
	Port      int
	User      string
	Password  string
	SSLMode   string
	CreatedAt time.Time
}

type Board struct {
	ID             int64
	Name           string
	DbType         string // "local" or "postgres"
	PgConnectionID *int64
	PgDatabaseName string
	CreatedAt      time.Time
	Columns        []Column
}

type Column struct {
	ID           int64
	BoardID      int64
	Name         string
	Position     int
	IsDoneColumn bool
	CreatedAt    time.Time
	Cards        []Card
}

type Card struct {
	ID          int64
	ColumnID    int64
	Title       string
	Description string
	Position    int
	CompletedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Assignees   []Person
	Comments    []Comment
	Checklist   []ChecklistItem
}

type Person struct {
	ID        int64
	BoardID   int64
	Name      string
	CreatedAt time.Time
}

type Comment struct {
	ID        int64
	CardID    int64
	Content   string
	CreatedAt time.Time
}

type ChecklistItem struct {
	ID          int64
	CardID      int64
	Content     string
	IsCompleted bool
	Position    int
	CreatedAt   time.Time
}
