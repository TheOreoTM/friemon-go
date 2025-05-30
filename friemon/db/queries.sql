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

-- name: updateUser :one
UPDATE users SET balance = $2, selected_id = $3, order_by = $4, order_desc = $5, shinies_caught = $6, next_idx = $7 WHERE id = $1 RETURNING *;

-- name: createUser :one
INSERT INTO users (id) VALUES ($1) RETURNING *;

-- name: getSelectedCharacter :one
SELECT id, owner_id, claimed_timestamp, idx, character_id, level, xp, personality, shiny,
       iv_hp, iv_atk, iv_def, iv_sp_atk, iv_sp_def, iv_spd, iv_total, nickname, favourite,
       held_item, moves, color
FROM characters
WHERE characters.id = (SELECT selected_id FROM users WHERE users.id = $1);

-- name: deleteUsers :exec
DELETE FROM users;

-- name: deleteCharacters :exec
DELETE FROM characters;

-- name: getUserWithSelectedCharacter :one
SELECT 
    u.id as user_id,
    u.balance,
    u.selected_id,
    u.order_by,
    u.order_desc,
    u.shinies_caught,
    u.next_idx,
    c.id as character_id,
    c.owner_id,
    c.claimed_timestamp,
    c.idx,
    c.character_id as char_character_id,
    c.level,
    c.xp,
    c.personality,
    c.shiny,
    c.iv_hp,
    c.iv_atk,
    c.iv_def,
    c.iv_sp_atk,
    c.iv_sp_def,
    c.iv_spd,
    c.iv_total,
    c.nickname,
    c.favourite,
    c.held_item,
    c.moves,
    c.color
FROM users u
LEFT JOIN characters c ON u.selected_id = c.id
WHERE u.id = $1;

-- name: getCharactersWithOwnerInfo :many
SELECT 
    c.*,
    u.balance as owner_balance,
    u.shinies_caught as owner_shinies_caught
FROM characters c
INNER JOIN users u ON c.owner_id = u.id
WHERE c.owner_id = $1
ORDER BY c.idx;

-- name: getCharacterWithOwnerInfo :one
SELECT 
    c.*,
    u.balance as owner_balance,
    u.shinies_caught as owner_shinies_caught,
    u.order_by as owner_order_by,
    u.order_desc as owner_order_desc
FROM characters c
INNER JOIN users u ON c.owner_id = u.id
WHERE c.id = $1;

-- name: getUsersWithCharacterCounts :many
SELECT 
    u.*,
    COUNT(c.id) as character_count,
    COUNT(CASE WHEN c.shiny = true THEN 1 END) as shiny_count,
    AVG(c.level) as avg_level
FROM users u
LEFT JOIN characters c ON u.id = c.owner_id
GROUP BY u.id, u.balance, u.selected_id, u.order_by, u.order_desc, u.shinies_caught, u.next_idx;

-- name: getFavouriteCharactersForUser :many
SELECT 
    c.*,
    u.balance as owner_balance
FROM characters c
INNER JOIN users u ON c.owner_id = u.id
WHERE c.owner_id = $1 AND c.favourite = true
ORDER BY c.idx;

-- name: getShinyCharactersForUser :many
SELECT 
    c.*,
    u.balance as owner_balance
FROM characters c
INNER JOIN users u ON c.owner_id = u.id
WHERE c.owner_id = $1 AND c.shiny = true
ORDER BY c.idx;

-- name: getTopLevelCharactersForUser :many
SELECT 
    c.*,
    u.balance as owner_balance
FROM characters c
INNER JOIN users u ON c.owner_id = u.id
WHERE c.owner_id = $1
ORDER BY c.level DESC, c.xp DESC
LIMIT $2;

-- name: searchCharactersByNickname :many
SELECT 
    c.*,
    u.balance as owner_balance
FROM characters c
INNER JOIN users u ON c.owner_id = u.id
WHERE c.owner_id = $1 AND c.nickname ILIKE '%' || $2 || '%'
ORDER BY c.idx;

-- name: updateUserSelectedCharacter :one
UPDATE users 
SET selected_id = $2 
WHERE id = $1 
RETURNING *;