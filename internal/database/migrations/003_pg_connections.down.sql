-- SQLite doesn't support DROP COLUMN, so we recreate the boards table
CREATE TABLE boards_backup (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    db_type TEXT NOT NULL DEFAULT 'local',
    connection_string TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO boards_backup (id, name, db_type, created_at)
SELECT id, name, db_type, created_at FROM boards;

DROP TABLE boards;
ALTER TABLE boards_backup RENAME TO boards;

DROP TABLE IF EXISTS pg_connections;
