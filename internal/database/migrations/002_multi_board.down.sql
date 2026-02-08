-- SQLite doesn't support DROP COLUMN before 3.35.0, so we recreate the tables

-- Recreate people without board_id
CREATE TABLE people_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO people_backup (id, name, created_at) SELECT id, name, created_at FROM people;
DROP TABLE people;
ALTER TABLE people_backup RENAME TO people;

-- Recreate boards without new columns
CREATE TABLE boards_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO boards_backup (id, name, created_at) SELECT id, name, created_at FROM boards;
DROP TABLE boards;
ALTER TABLE boards_backup RENAME TO boards;
