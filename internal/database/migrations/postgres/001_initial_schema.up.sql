-- Columns (board_id is always 1 in per-board Postgres DB)
CREATE TABLE columns (
    id SERIAL PRIMARY KEY,
    board_id INTEGER NOT NULL DEFAULT 1,
    name TEXT NOT NULL,
    position INTEGER NOT NULL,
    is_done_column BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Cards
CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    column_id INTEGER NOT NULL REFERENCES columns(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    position INTEGER NOT NULL,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- People
CREATE TABLE people (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Card Assignees
CREATE TABLE card_assignees (
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    person_id INTEGER NOT NULL REFERENCES people(id) ON DELETE CASCADE,
    PRIMARY KEY (card_id, person_id)
);

-- Comments
CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Checklist Items
CREATE TABLE checklist_items (
    id SERIAL PRIMARY KEY,
    card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_completed BOOLEAN DEFAULT FALSE,
    position INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_columns_board_id ON columns(board_id);
CREATE INDEX idx_cards_column_id ON cards(column_id);
CREATE INDEX idx_card_assignees_card_id ON card_assignees(card_id);
CREATE INDEX idx_card_assignees_person_id ON card_assignees(person_id);
CREATE INDEX idx_comments_card_id ON comments(card_id);
CREATE INDEX idx_checklist_items_card_id ON checklist_items(card_id);
