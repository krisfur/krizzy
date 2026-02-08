-- Backfill people with NULL board_id: assign to the first board
UPDATE people SET board_id = (SELECT id FROM boards ORDER BY id LIMIT 1) WHERE board_id IS NULL;
