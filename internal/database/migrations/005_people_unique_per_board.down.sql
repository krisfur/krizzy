PRAGMA foreign_keys=off;

ALTER TABLE people RENAME TO people_new;
ALTER TABLE card_assignees RENAME TO card_assignees_new;

CREATE TABLE people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    board_id INTEGER REFERENCES boards(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO people (id, name, board_id, created_at)
SELECT id, name, board_id, created_at
FROM people_new;

CREATE TABLE card_assignees (
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    PRIMARY KEY (card_id, person_id)
);

INSERT INTO card_assignees (card_id, person_id)
SELECT card_id, person_id
FROM card_assignees_new;

DROP TABLE card_assignees_new;
DROP TABLE people_new;

PRAGMA foreign_keys=on;
