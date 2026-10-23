-- Discord IDs are 64-bit snowflakes; they always fit in a signed BIGINT.

CREATE TABLE guilds (
    id               BIGINT PRIMARY KEY,
    spawn_channel_id BIGINT
);

CREATE TABLE users (
    id           BIGINT PRIMARY KEY,
    next_idx     INT NOT NULL DEFAULT 0, -- last number given to a caught character
    selected_idx INT,                    -- the character's per-user number, not its row id
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One row per caught copy. character_id points at a Character in internal/data.
CREATE TABLE characters (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    owner_id         BIGINT NOT NULL REFERENCES users (id),
    idx              INT NOT NULL, -- the owner's #1, #2, ... shown in /collection
    character_id     TEXT NOT NULL,
    level            INT NOT NULL,
    xp               INT NOT NULL DEFAULT 0,
    iv_hp            SMALLINT NOT NULL,
    iv_attack        SMALLINT NOT NULL,
    iv_defense       SMALLINT NOT NULL,
    iv_magic_attack  SMALLINT NOT NULL,
    iv_magic_defense SMALLINT NOT NULL,
    iv_speed         SMALLINT NOT NULL,
    favorite         BOOLEAN NOT NULL DEFAULT false,
    caught_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (owner_id, idx)
);
