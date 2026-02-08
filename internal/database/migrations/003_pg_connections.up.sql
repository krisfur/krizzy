CREATE TABLE pg_connections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER NOT NULL DEFAULT 5432,
    username TEXT NOT NULL,
    password TEXT DEFAULT '',
    ssl_mode TEXT NOT NULL DEFAULT 'disable',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE boards ADD COLUMN pg_connection_id INTEGER REFERENCES pg_connections(id);
ALTER TABLE boards ADD COLUMN pg_database_name TEXT DEFAULT '';

-- Backfill people with NULL board_id: assign to the first board
UPDATE people SET board_id = (SELECT id FROM boards ORDER BY id LIMIT 1) WHERE board_id IS NULL;
