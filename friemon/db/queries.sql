-- name: getCharactersForUser :many
SELECT * FROM characters WHERE owner_id = $1;

-- name: createCharacter :one
INSERT INTO characters (id, owner_id, claimed_timestamp, idx, character_id, level, xp, personality, shiny, iv_hp, iv_atk, iv_def, iv_sp_atk, iv_sp_def, iv_spd, iv_total, nickname, favourite, held_item, moves, color)
VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
RETURNING id, owner_id, claimed_timestamp, idx, character_id, level, xp, personality, shiny, iv_hp, iv_atk, iv_def, iv_sp_atk, iv_sp_def, iv_spd, iv_total, nickname, favourite, held_item, moves, color;

-- name: getCharacter :one
SELECT * FROM characters WHERE id = $1;

-- name: updateCharacter :one
UPDATE characters SET owner_id = $2, claimed_timestamp = $3, idx = $4, character_id = $5, level = $6, xp = $7, personality = $8, shiny = $9, iv_hp = $10, iv_atk = $11, iv_def = $12, iv_sp_atk = $13, iv_sp_def = $14, iv_spd = $15, iv_total = $16, nickname = $17, favourite = $18, held_item = $19, moves = $20, color = $21 WHERE id = $1 RETURNING id, owner_id, claimed_timestamp, idx, character_id, level, xp, personality, shiny, iv_hp, iv_atk, iv_def, iv_sp_atk, iv_sp_def, iv_spd, iv_total, nickname, favourite, held_item, moves, color;

-- name: deleteCharacter :exec
DELETE FROM characters WHERE id = $1;

-- name: getUser :one
SELECT * FROM users WHERE id = $1;

-- name: DeleteEverything :exec
TRUNCATE TABLE characters;
