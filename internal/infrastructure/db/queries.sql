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

-- name: getGameSetting :one
SELECT setting_value FROM game_settings WHERE setting_key = $1;

-- name: getAllGameSettings :many
SELECT setting_key, setting_value FROM game_settings;

-- name: updateGameSetting :exec
UPDATE game_settings SET setting_value = $2, updated_at = CURRENT_TIMESTAMP WHERE setting_key = $1;

-- name: createGameSetting :exec
INSERT INTO game_settings (setting_key, setting_value) VALUES ($1, $2);

-- Battle queries
-- name: createBattle :one
INSERT INTO battles (id, challenger_id, opponent_id, status, battle_settings, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, challenger_id, opponent_id, winner_id, status, turn_count, current_turn_player, main_thread_id, challenger_thread_id, opponent_thread_id, battle_settings, created_at, updated_at, completed_at;

-- name: getBattle :one
SELECT id, challenger_id, opponent_id, winner_id, status, turn_count, current_turn_player, main_thread_id, challenger_thread_id, opponent_thread_id, battle_settings, created_at, updated_at, completed_at
FROM battles WHERE id = $1;

-- name: updateBattle :one
UPDATE battles 
SET winner_id = $2, status = $3, turn_count = $4, current_turn_player = $5, main_thread_id = $6, challenger_thread_id = $7, opponent_thread_id = $8, battle_settings = $9, updated_at = $10, completed_at = $11
WHERE id = $1
RETURNING id, challenger_id, opponent_id, winner_id, status, turn_count, current_turn_player, main_thread_id, challenger_thread_id, opponent_thread_id, battle_settings, created_at, updated_at, completed_at;

-- name: getActiveBattleForUser :one
SELECT id, challenger_id, opponent_id, winner_id, status, turn_count, current_turn_player, main_thread_id, challenger_thread_id, opponent_thread_id, battle_settings, created_at, updated_at, completed_at
FROM battles 
WHERE (challenger_id = $1 OR opponent_id = $1) AND status IN ('pending', 'active')
ORDER BY created_at DESC LIMIT 1;

-- name: getUserBattleHistory :many
SELECT id, challenger_id, opponent_id, winner_id, status, turn_count, current_turn_player, main_thread_id, challenger_thread_id, opponent_thread_id, battle_settings, created_at, updated_at, completed_at
FROM battles 
WHERE (challenger_id = $1 OR opponent_id = $1) AND status = 'completed'
ORDER BY completed_at DESC
LIMIT $2 OFFSET $3;

-- name: getBattlesByStatus :many
SELECT id, challenger_id, opponent_id, winner_id, status, turn_count, current_turn_player, main_thread_id, challenger_thread_id, opponent_thread_id, battle_settings, created_at, updated_at, completed_at
FROM battles 
WHERE status = $1
ORDER BY created_at DESC;

-- Battle Team queries
-- name: createBattleTeamMember :one
INSERT INTO battle_teams (id, battle_id, user_id, team_position, character_id, character_data, current_hp, status_effects, stat_stages, is_active, is_fainted, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING id, battle_id, user_id, team_position, character_id, character_data, current_hp, status_effects, stat_stages, is_active, is_fainted, created_at;

-- name: getBattleTeam :many
SELECT id, battle_id, user_id, team_position, character_id, character_data, current_hp, status_effects, stat_stages, is_active, is_fainted, created_at
FROM battle_teams 
WHERE battle_id = $1 AND user_id = $2
ORDER BY team_position;

-- name: updateBattleTeamMember :one
UPDATE battle_teams 
SET current_hp = $3, status_effects = $4, stat_stages = $5, is_active = $6, is_fainted = $7
WHERE id = $1 AND battle_id = $2
RETURNING id, battle_id, user_id, team_position, character_id, character_data, current_hp, status_effects, stat_stages, is_active, is_fainted, created_at;

-- name: getBattleTeamMember :one
SELECT id, battle_id, user_id, team_position, character_id, character_data, current_hp, status_effects, stat_stages, is_active, is_fainted, created_at
FROM battle_teams 
WHERE id = $1;

-- name: createBattleTurn :one
INSERT INTO battle_turns (id, battle_id, turn_number, user_id, action_type, action_data, result_data, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, battle_id, turn_number, user_id, action_type, action_data, result_data, created_at;

-- name: getBattleTurns :many
SELECT id, battle_id, turn_number, user_id, action_type, action_data, result_data, created_at
FROM battle_turns 
WHERE battle_id = $1
ORDER BY turn_number, created_at;

-- name: getLastBattleTurn :one
SELECT id, battle_id, turn_number, user_id, action_type, action_data, result_data, created_at
FROM battle_turns 
WHERE battle_id = $1
ORDER BY turn_number DESC, created_at DESC
LIMIT 1;

-- name: getUserElo :one
SELECT user_id, elo_rating, battles_won, battles_lost, battles_total, highest_elo, created_at, updated_at
FROM user_elo WHERE user_id = $1;

-- name: createUserElo :one
INSERT INTO user_elo (user_id, elo_rating, battles_won, battles_lost, battles_total, highest_elo, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING user_id, elo_rating, battles_won, battles_lost, battles_total, highest_elo, created_at, updated_at;

-- name: updateUserElo :one
UPDATE user_elo 
SET elo_rating = $2, battles_won = $3, battles_lost = $4, battles_total = $5, highest_elo = $6, updated_at = $7
WHERE user_id = $1
RETURNING user_id, elo_rating, battles_won, battles_lost, battles_total, highest_elo, created_at, updated_at;

-- name: getEloLeaderboard :many
SELECT user_id, elo_rating, battles_won, battles_lost, battles_total, highest_elo, created_at, updated_at
FROM user_elo 
WHERE battles_total >= $1
ORDER BY elo_rating DESC
LIMIT $2 OFFSET $3;

-- name: getUserEloRank :one
SELECT COUNT(*) + 1 as rank
FROM user_elo u1
WHERE u1.elo_rating > (
    SELECT u2.elo_rating 
    FROM user_elo u2 
    WHERE u2.user_id = $1
)
AND u1.battles_total >= $2;