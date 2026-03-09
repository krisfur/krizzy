PRAGMA foreign_keys=off;

ALTER TABLE people RENAME TO people_with_color;

CREATE TABLE people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (board_id, name)
);

INSERT INTO people (id, name, board_id, created_at)
SELECT id, name, board_id, created_at
FROM people_with_color;

DROP TABLE people_with_color;

CREATE INDEX idx_people_board_id ON people(board_id);

PRAGMA foreign_keys=on;
