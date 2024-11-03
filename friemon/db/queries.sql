-- name: GetCharactersForUser :many
SELECT * FROM characters WHERE owner_id = $1;

-- name: CreateCharacterForUser :one
INSERT INTO characters (id, owner_id, claimed_timestamp, idx, character_id, level, xp, personality, shiny, iv_hp, iv_atk, iv_def, iv_sp_atk, iv_sp_def, iv_spd, iv_total, nickname, favourite, held_item, moves, color)
VALUES (uuid_generate_v4(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20)
RETURNING id, owner_id, claimed_timestamp, idx, character_id, level, xp, personality, shiny, iv_hp, iv_atk, iv_def, iv_sp_atk, iv_sp_def, iv_spd, iv_total, nickname, favourite, held_item, moves, color;
