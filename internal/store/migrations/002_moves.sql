-- Move IDs a character knows (at most 4), pointing at Moves in internal/data.
ALTER TABLE characters ADD COLUMN moves TEXT[] NOT NULL DEFAULT '{}';
