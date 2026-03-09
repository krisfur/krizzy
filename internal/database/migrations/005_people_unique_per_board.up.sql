PRAGMA foreign_keys=off;

ALTER TABLE people RENAME TO people_old;
ALTER TABLE card_assignees RENAME TO card_assignees_old;

CREATE TABLE people (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    board_id INTEGER NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (board_id, name)
);

INSERT INTO people (id, name, board_id, created_at)
SELECT id,
       name,
       COALESCE(board_id, (SELECT id FROM boards ORDER BY id LIMIT 1)),
       created_at
FROM people_old;

CREATE TABLE card_assignees (
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    PRIMARY KEY (card_id, person_id)
);

INSERT INTO card_assignees (card_id, person_id)
SELECT card_id, person_id
FROM card_assignees_old;

DROP TABLE card_assignees_old;
DROP TABLE people_old;

CREATE INDEX idx_people_board_id ON people(board_id);
CREATE INDEX idx_card_assignees_card_id ON card_assignees(card_id);
CREATE INDEX idx_card_assignees_person_id ON card_assignees(person_id);

PRAGMA foreign_keys=on;
