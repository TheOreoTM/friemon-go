-- Natures: give every existing character a random one. random() runs once per row.
ALTER TABLE characters ADD COLUMN nature TEXT NOT NULL DEFAULT 'hardy';
UPDATE characters SET nature = (ARRAY[
    'hardy', 'lonely', 'adamant', 'naughty', 'brave',
    'bold', 'docile', 'impish', 'lax', 'relaxed',
    'modest', 'mild', 'bashful', 'rash', 'quiet',
    'calm', 'gentle', 'careful', 'quirky', 'sassy',
    'timid', 'hasty', 'jolly', 'naive', 'serious'
])[1 + floor(random() * 25)::int];

-- Effort values, earned by knocking out opponents in battle.
ALTER TABLE characters
    ADD COLUMN ev_hp            SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN ev_attack        SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN ev_defense       SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN ev_magic_attack  SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN ev_magic_defense SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN ev_speed         SMALLINT NOT NULL DEFAULT 0;

-- The party replaces the single selected character: an ordered list of up to 4
-- character numbers, where the first is the leader.
ALTER TABLE users ADD COLUMN party INT[] NOT NULL DEFAULT '{}';
UPDATE users SET party = ARRAY[selected_idx] WHERE selected_idx IS NOT NULL;
ALTER TABLE users DROP COLUMN selected_idx;
