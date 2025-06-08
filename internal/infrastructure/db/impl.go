package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/core/entities"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

var _ Store = (*Queries)(nil)

func stringToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgTextToString(pt pgtype.Text) *string {
	if !pt.Valid {
		return nil
	}
	return &pt.String
}

func stringToPgTextRequired(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func (q *Queries) DeleteEverything(ctx context.Context) error {
	err := q.deleteUsers(ctx)
	if err != nil {
		return err
	}

	err = q.deleteCharacters(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queries) UpdateUser(ctx context.Context, user entities.User) (*entities.User, error) {
	dbUser, err := q.updateUser(ctx, updateUserParams{
		ID:            user.ID.String(),
		Balance:       int32(user.Balance),
		SelectedID:    user.SelectedID,
		OrderBy:       int32(user.Order.OrderBy),
		OrderDesc:     user.Order.Desc,
		ShiniesCaught: int32(user.ShiniesCaught),
		NextIdx:       int32(user.NextIdx),
	})

	if err != nil {
		return &entities.User{}, err
	}
	return dbUserToModelUser(dbUser), nil
}

func (q *Queries) GetSelectedCharacter(ctx context.Context, id snowflake.ID) (*entities.Character, error) {
	dbch, err := q.getSelectedCharacter(ctx, id.String())
	if err != nil {
		return &entities.Character{}, err
	}

	return dbCharToModelChar(dbch), nil
}

func (q *Queries) CreateUser(ctx context.Context, id snowflake.ID) (*entities.User, error) {
	dbUser, err := q.createUser(ctx, id.String())
	if err != nil {
		return &entities.User{}, err
	}

	return dbUserToModelUser(dbUser), nil
}

func (q *Queries) GetUser(ctx context.Context, id snowflake.ID) (*entities.User, error) {
	dbUser, err := q.getUser(ctx, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			dbUser, err := q.createUser(ctx, id.String())
			if err != nil {
				return &entities.User{}, err
			}
			return dbUserToModelUser(dbUser), nil
		}
		return &entities.User{}, err
	}
	return dbUserToModelUser(dbUser), nil // Ugly, somehow fix
}

func (q *Queries) DeleteCharacter(ctx context.Context, id uuid.UUID) (*entities.Character, error) {
	dbch, err := q.getCharacter(ctx, id)
	if err != nil {
		return &entities.Character{}, err
	}

	err = q.deleteCharacter(ctx, id)
	if err != nil {
		return &entities.Character{}, err
	}

	return dbCharToModelChar(dbch), nil
}

func (q *Queries) UpdateCharacter(ctx context.Context, id uuid.UUID, ch *entities.Character) (*entities.Character, error) {

	dbch, err := q.updateCharacter(ctx, updateCharacterParams{
		ID:               (ch.ID),
		OwnerID:          ch.OwnerID,
		ClaimedTimestamp: ch.ClaimedTimestamp,
		Idx:              int32(ch.IDX),
		CharacterID:      int32(ch.CharacterID),
		Level:            int32(ch.Level),
		Xp:               int32(ch.XP),
		Personality:      ch.Personality.String(),
		Shiny:            ch.Shiny,
		IvHp:             int32(ch.IvHP),
		IvAtk:            int32(ch.IvAtk),
		IvDef:            int32(ch.IvDef),
		IvSpAtk:          int32(ch.IvSpAtk),
		IvSpDef:          int32(ch.IvSpDef),
		IvSpd:            int32(ch.IvSpd),
		IvTotal:          ch.IvTotal,
		Nickname:         ch.Nickname,
		Favourite:        ch.Favourite,
		HeldItem:         int32(ch.HeldItem),
		Moves:            ch.Moves,
		Color:            ch.Color,
	})
	if err != nil {
		return &entities.Character{}, err
	}

	return dbCharToModelChar(dbch), nil
}

func (q *Queries) GetCharacter(ctx context.Context, id uuid.UUID) (*entities.Character, error) {
	dbch, err := q.getCharacter(ctx, id)
	if err != nil {
		return &entities.Character{}, err
	}

	return dbCharToModelChar(dbch), nil
}

func (q *Queries) CreateCharacter(ctx context.Context, ownerID snowflake.ID, char *entities.Character) error {
	log := logger.NewLogger("database.characters")
	start := time.Now()

	log.Debug("Creating character",
		logger.Operation("create_character"),
		logger.DiscordUserID(ownerID),
		logger.CharacterName(char.CharacterName()),
		logger.CharacterLevel(char.Level),
	)

	defer func() {
		log.Debug("Create character operation completed",
			logger.Operation("create_character"),
			logger.Duration(time.Since(start)),
		)
	}()

	params := modelCharToDBChar(char)
	params.OwnerID = ownerID.String()

	dbChar, err := q.createCharacter(ctx, params)
	if err != nil {
		log.Error("Failed to create character",
			logger.Operation("create_character"),
			logger.DiscordUserID(ownerID),
			logger.CharacterName(char.CharacterName()),
			logger.ErrorField(err),
		)
		return err
	}

	// Update the character with the generated ID
	char.ID = dbChar.ID

	log.Info("Character created successfully",
		logger.Operation("create_character"),
		logger.DiscordUserID(ownerID),
		logger.CharacterID(char.ID),
		logger.CharacterName(char.CharacterName()),
	)

	return nil
}

func (q *Queries) GetCharactersForUser(ctx context.Context, userID snowflake.ID) ([]entities.Character, error) {
	log := logger.NewLogger("database.characters")
	start := time.Now()

	log.Debug("Getting characters for user",
		logger.Operation("get_user_characters"),
		logger.DiscordUserID(userID),
	)

	defer func() {
		log.Debug("Get user characters operation completed",
			logger.Operation("get_user_characters"),
			logger.Duration(time.Since(start)),
		)
	}()

	dbChars, err := q.getCharactersForUser(ctx, userID.String())
	if err != nil {
		log.Error("Failed to get characters for user",
			logger.Operation("get_user_characters"),
			logger.DiscordUserID(userID),
			logger.ErrorField(err),
		)
		return nil, err
	}

	characters := make([]entities.Character, len(dbChars))
	for i, dbChar := range dbChars {
		characters[i] = *dbCharToModelChar(dbChar)
	}

	log.Info("Characters retrieved for user",
		logger.Operation("get_user_characters"),
		logger.DiscordUserID(userID),
		zap.Int("character_count", len(characters)),
	)

	return characters, nil
}

func dbUserToModelUser(dbUser User) *entities.User {
	return &entities.User{
		ID:         snowflake.MustParse(dbUser.ID),
		Balance:    int(dbUser.Balance),
		SelectedID: dbUser.SelectedID,
		Order: entities.OrderOptions{
			OrderBy: int(dbUser.OrderBy),
			Desc:    dbUser.OrderDesc,
		},
		ShiniesCaught: int(dbUser.ShiniesCaught),
		NextIdx:       int(dbUser.NextIdx),
	}
}

func modelCharToDBChar(ch *entities.Character) createCharacterParams {
	return createCharacterParams{
		OwnerID:          ch.OwnerID,
		ClaimedTimestamp: ch.ClaimedTimestamp,
		CharacterID:      int32(ch.CharacterID),
		Level:            int32(ch.Level),
		Xp:               int32(ch.XP),
		Personality:      ch.Personality.String(),
		Shiny:            ch.Shiny,
		IvHp:             int32(ch.IvHP),
		IvAtk:            int32(ch.IvAtk),
		IvDef:            int32(ch.IvDef),
		IvSpAtk:          int32(ch.IvSpAtk),
		IvSpDef:          int32(ch.IvSpDef),
		IvSpd:            int32(ch.IvSpd),
		IvTotal:          ch.IvTotal,
		Nickname:         ch.Nickname,
		Favourite:        ch.Favourite,
		HeldItem:         int32(ch.HeldItem),
		Moves:            ch.Moves,
		Color:            ch.Color,
		Idx:              int32(ch.IDX),
	}
}

func dbCharToModelChar(dbch Character) *entities.Character {
	return &entities.Character{
		ID:               dbch.ID,
		OwnerID:          dbch.OwnerID,
		ClaimedTimestamp: dbch.ClaimedTimestamp,
		IDX:              int(dbch.Idx),
		CharacterID:      int(dbch.CharacterID),
		Level:            int(dbch.Level),
		XP:               int(dbch.Xp),
		Personality:      stringToPersonality(dbch.Personality),
		Shiny:            dbch.Shiny,
		IvHP:             int(dbch.IvHp),
		IvAtk:            int(dbch.IvAtk),
		IvDef:            int(dbch.IvDef),
		IvSpAtk:          int(dbch.IvSpAtk),
		IvSpDef:          int(dbch.IvSpDef),
		IvSpd:            int(dbch.IvSpd),
		IvTotal:          dbch.IvTotal,
		Nickname:         dbch.Nickname,
		Favourite:        dbch.Favourite,
		HeldItem:         int(dbch.HeldItem),
		Moves:            dbch.Moves,
		Color:            dbch.Color,
		BattleStats:      nil, // TODO: Load battle stats once a system is in place
	}
}

func stringToPersonality(s string) constants.Personality {
	for _, p := range constants.Personalities {
		if p.String() == s {
			return p
		}
	}
	return constants.PersonalityAloof
}

func (q *Queries) GetGameSettings(ctx context.Context) (entities.GameSettings, error) {
	settings := entities.GameSettings{
		TurnLimit:          25,
		TurnTimeoutSeconds: 60,
		SwitchCostsTurn:    false,
		MaxTeamSize:        3,
	}

	rows, err := q.getAllGameSettings(ctx)
	if err != nil {
		return settings, err
	}

	for _, row := range rows {
		switch row.SettingKey {
		case "battle_turn_limit":
			if val, err := strconv.Atoi(row.SettingValue); err == nil {
				settings.TurnLimit = val
			}
		case "battle_turn_timeout_seconds":
			if val, err := strconv.Atoi(row.SettingValue); err == nil {
				settings.TurnTimeoutSeconds = val
			}
		case "battle_switch_costs_turn":
			settings.SwitchCostsTurn = row.SettingValue == "true"
		case "battle_max_team_size":
			if val, err := strconv.Atoi(row.SettingValue); err == nil {
				settings.MaxTeamSize = val
			}
		}
	}

	return settings, nil
}

func (q *Queries) GetGameSetting(ctx context.Context, key string) (string, error) {
	return q.getGameSetting(ctx, key)
}

func (q *Queries) UpdateGameSetting(ctx context.Context, key, value string) error {
	return q.updateGameSetting(ctx, updateGameSettingParams{SettingKey: key, SettingValue: value})
}

func (q *Queries) CreateGameSetting(ctx context.Context, key, value string) error {
	return q.createGameSetting(ctx, createGameSettingParams{SettingKey: key, SettingValue: value})
}

// Battle methods
func (q *Queries) CreateBattle(ctx context.Context, battle *entities.Battle) error {
	settingsJSON, err := json.Marshal(battle.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal battle settings: %w", err)
	}

	_, err = q.createBattle(ctx, createBattleParams{
		ID:             battle.ID,
		ChallengerID:   battle.ChallengerID.String(),
		OpponentID:     battle.OpponentID.String(),
		Status:         string(battle.Status),
		BattleSettings: settingsJSON,
		CreatedAt:      battle.CreatedAt,
		UpdatedAt:      battle.UpdatedAt,
	})

	return err
}

func (q *Queries) GetBattle(ctx context.Context, id uuid.UUID) (*entities.Battle, error) {
	row, err := q.getBattle(ctx, id)
	if err != nil {
		return nil, err
	}

	return dbBattleToModelBattle(row)
}

func (q *Queries) UpdateBattle(ctx context.Context, battle *entities.Battle) error {
	settingsJSON, err := json.Marshal(battle.Settings)
	if err != nil {
		return fmt.Errorf("failed to marshal battle settings: %w", err)
	}

	var winnerID pgtype.Text
	if battle.WinnerID != nil {
		winnerID = stringToPgTextRequired(battle.WinnerID.String())
	} else {
		winnerID = pgtype.Text{Valid: false}
	}

	var currentPlayer pgtype.Text
	if battle.CurrentTurnPlayer != nil {
		currentPlayer = stringToPgTextRequired(battle.CurrentTurnPlayer.String())
	} else {
		currentPlayer = pgtype.Text{Valid: false}
	}

	var mainThreadID, challengerThreadID, opponentThreadID pgtype.Text
	if battle.MainThreadID != nil {
		mainThreadID = stringToPgTextRequired(battle.MainThreadID.String())
	} else {
		mainThreadID = pgtype.Text{Valid: false}
	}

	if battle.ChallengerThreadID != nil {
		challengerThreadID = stringToPgTextRequired(battle.ChallengerThreadID.String())
	} else {
		challengerThreadID = pgtype.Text{Valid: false}
	}

	if battle.OpponentThreadID != nil {
		opponentThreadID = stringToPgTextRequired(battle.OpponentThreadID.String())
	} else {
		opponentThreadID = pgtype.Text{Valid: false}
	}

	var completedAt pgtype.Timestamptz
	if battle.CompletedAt != nil {
		completedAt = pgtype.Timestamptz{Time: *battle.CompletedAt, Valid: true}
	} else {
		completedAt = pgtype.Timestamptz{Valid: false}
	}

	_, err = q.updateBattle(ctx, updateBattleParams{
		ID:                 battle.ID,
		WinnerID:           winnerID,
		Status:             string(battle.Status),
		TurnCount:          int32(battle.TurnCount),
		CurrentTurnPlayer:  currentPlayer,
		MainThreadID:       mainThreadID,
		ChallengerThreadID: challengerThreadID,
		OpponentThreadID:   opponentThreadID,
		BattleSettings:     settingsJSON,
		UpdatedAt:          time.Now(),
		CompletedAt:        completedAt,
	})

	return err
}

func (q *Queries) GetActiveBattleForUser(ctx context.Context, userID snowflake.ID) (*entities.Battle, error) {
	row, err := q.getActiveBattleForUser(ctx, userID.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return dbBattleToModelBattle(row)
}

func (q *Queries) GetUserBattleHistory(ctx context.Context, userID snowflake.ID, limit, offset int) ([]*entities.Battle, error) {
	rows, err := q.getUserBattleHistory(ctx, getUserBattleHistoryParams{ChallengerID: userID.String(), Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, err
	}

	battles := make([]*entities.Battle, len(rows))
	for i, row := range rows {
		battle, err := dbBattleToModelBattle(row)
		if err != nil {
			return nil, err
		}
		battles[i] = battle
	}

	return battles, nil
}

func (q *Queries) GetBattlesByStatus(ctx context.Context, status entities.BattleStatus) ([]*entities.Battle, error) {
	rows, err := q.getBattlesByStatus(ctx, string(status))
	if err != nil {
		return nil, err
	}

	battles := make([]*entities.Battle, len(rows))
	for i, row := range rows {
		battle, err := dbBattleToModelBattle(row)
		if err != nil {
			return nil, err
		}
		battles[i] = battle
	}

	return battles, nil
}

// Battle Team methods
func (q *Queries) CreateBattleTeamMember(ctx context.Context, member *entities.BattleTeamMember) error {
	characterDataJSON, err := json.Marshal(member.CharacterData)
	if err != nil {
		return fmt.Errorf("failed to marshal character data: %w", err)
	}

	statStagesJSON, err := json.Marshal(member.StatStages)
	if err != nil {
		return fmt.Errorf("failed to marshal stat stages: %w", err)
	}

	statusEffectsStr := make([]string, len(member.StatusEffects))
	for i, effect := range member.StatusEffects {
		statusEffectsStr[i] = string(effect)
	}

	_, err = q.createBattleTeamMember(ctx, createBattleTeamMemberParams{
		ID:            member.ID,
		BattleID:      member.BattleID,
		UserID:        member.UserID.String(),
		TeamPosition:  int32(member.TeamPosition),
		CharacterID:   member.CharacterID,
		CharacterData: characterDataJSON,
		CurrentHp:     int32(member.CurrentHP),
		StatusEffects: statusEffectsStr,
		StatStages:    statStagesJSON,
		IsActive:      member.IsActive,
		IsFainted:     member.IsFainted,
		CreatedAt:     member.CreatedAt,
	})

	return err
}

func (q *Queries) GetBattleTeam(ctx context.Context, battleID uuid.UUID, userID snowflake.ID) ([]*entities.BattleTeamMember, error) {
	rows, err := q.getBattleTeam(ctx, getBattleTeamParams{
		BattleID: battleID,
		UserID:   userID.String(),
	})
	if err != nil {
		return nil, err
	}

	members := make([]*entities.BattleTeamMember, len(rows))
	for i, row := range rows {
		member, err := dbBattleTeamMemberToModel(row)
		if err != nil {
			return nil, err
		}
		members[i] = member
	}

	return members, nil
}

func (q *Queries) UpdateBattleTeamMember(ctx context.Context, member *entities.BattleTeamMember) error {
	statStagesJSON, err := json.Marshal(member.StatStages)
	if err != nil {
		return fmt.Errorf("failed to marshal stat stages: %w", err)
	}

	statusEffectsStr := make([]string, len(member.StatusEffects))
	for i, effect := range member.StatusEffects {
		statusEffectsStr[i] = string(effect)
	}

	_, err = q.updateBattleTeamMember(ctx, updateBattleTeamMemberParams{
		ID:            member.ID,
		BattleID:      member.BattleID,
		CurrentHp:     int32(member.CurrentHP),
		StatusEffects: statusEffectsStr,
		StatStages:    statStagesJSON,
		IsActive:      member.IsActive,
		IsFainted:     member.IsFainted,
	})

	return err
}

func (q *Queries) GetBattleTeamMember(ctx context.Context, id uuid.UUID) (*entities.BattleTeamMember, error) {
	row, err := q.getBattleTeamMember(ctx, id)
	if err != nil {
		return nil, err
	}

	return dbBattleTeamMemberToModel(row)
}

// Battle Turn methods
func (q *Queries) CreateBattleTurn(ctx context.Context, turn *entities.BattleTurn) error {
	actionDataJSON, err := json.Marshal(turn.ActionData)
	if err != nil {
		return fmt.Errorf("failed to marshal action data: %w", err)
	}

	resultDataJSON, err := json.Marshal(turn.ResultData)
	if err != nil {
		return fmt.Errorf("failed to marshal result data: %w", err)
	}

	_, err = q.createBattleTurn(ctx, createBattleTurnParams{
		ID:         turn.ID,
		BattleID:   turn.BattleID,
		TurnNumber: int32(turn.TurnNumber),
		UserID:     turn.UserID.String(),
		ActionType: string(turn.ActionType),
		ActionData: actionDataJSON,
		ResultData: resultDataJSON,
		CreatedAt:  turn.CreatedAt,
	})

	return err
}

func (q *Queries) GetBattleTurns(ctx context.Context, battleID uuid.UUID) ([]*entities.BattleTurn, error) {
	rows, err := q.getBattleTurns(ctx, battleID)
	if err != nil {
		return nil, err
	}

	turns := make([]*entities.BattleTurn, len(rows))
	for i, row := range rows {
		turn, err := dbBattleTurnToModel(row)
		if err != nil {
			return nil, err
		}
		turns[i] = turn
	}

	return turns, nil
}

func (q *Queries) GetLastBattleTurn(ctx context.Context, battleID uuid.UUID) (*entities.BattleTurn, error) {
	row, err := q.getLastBattleTurn(ctx, battleID)
	if err != nil {
		return nil, err
	}

	return dbBattleTurnToModel(row)
}

// ELO methods
func (q *Queries) GetUserElo(ctx context.Context, userID snowflake.ID) (*entities.UserElo, error) {
	row, err := q.getUserElo(ctx, userID.String())
	if err != nil {
		return nil, err
	}

	return dbUserEloToModel(row), nil
}

func (q *Queries) CreateUserElo(ctx context.Context, elo *entities.UserElo) error {
	_, err := q.createUserElo(ctx, createUserEloParams{
		UserID:       elo.UserID.String(),
		EloRating:    int32(elo.EloRating),
		BattlesWon:   int32(elo.BattlesWon),
		BattlesLost:  int32(elo.BattlesLost),
		BattlesTotal: int32(elo.BattlesTotal),
		HighestElo:   int32(elo.HighestElo),
		CreatedAt:    elo.CreatedAt,
		UpdatedAt:    elo.UpdatedAt,
	})

	return err
}

func (q *Queries) UpdateUserElo(ctx context.Context, elo *entities.UserElo) error {
	_, err := q.updateUserElo(ctx, updateUserEloParams{
		UserID:       elo.UserID.String(),
		EloRating:    int32(elo.EloRating),
		BattlesWon:   int32(elo.BattlesWon),
		BattlesLost:  int32(elo.BattlesLost),
		BattlesTotal: int32(elo.BattlesTotal),
		HighestElo:   int32(elo.HighestElo),
		UpdatedAt:    time.Now(),
	})

	return err
}

func (q *Queries) GetEloLeaderboard(ctx context.Context, minBattles, limit, offset int) ([]*entities.UserElo, error) {
	rows, err := q.getEloLeaderboard(ctx, getEloLeaderboardParams{
		BattlesTotal: int32(minBattles),
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, err
	}

	leaderboard := make([]*entities.UserElo, len(rows))
	for i, row := range rows {
		leaderboard[i] = dbUserEloToModel(row)
	}

	return leaderboard, nil
}

func (q *Queries) GetUserEloRank(ctx context.Context, userID snowflake.ID, minBattles int) (int, error) {
	rank, err := q.getUserEloRank(ctx, getUserEloRankParams{UserID: userID.String(), BattlesTotal: int32(minBattles)})
	if err != nil {
		return 0, err
	}

	return int(rank), nil
}

// Helper conversion functions
func dbBattleToModelBattle(row Battle) (*entities.Battle, error) {
	var settings entities.GameSettings
	if err := json.Unmarshal(row.BattleSettings, &settings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal battle settings: %w", err)
	}

	challengerID, err := snowflake.Parse(row.ChallengerID)
	if err != nil {
		return nil, fmt.Errorf("invalid challenger ID: %w", err)
	}

	opponentID, err := snowflake.Parse(row.OpponentID)
	if err != nil {
		return nil, fmt.Errorf("invalid opponent ID: %w", err)
	}

	battle := &entities.Battle{
		ID:           row.ID,
		ChallengerID: challengerID,
		OpponentID:   opponentID,
		Status:       entities.BattleStatus(row.Status),
		TurnCount:    int(row.TurnCount),
		Settings:     settings,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}

	// Handle nullable fields with pgtype.Text
	if row.WinnerID.Valid {
		winnerID, err := snowflake.Parse(row.WinnerID.String)
		if err == nil {
			battle.WinnerID = &winnerID
		}
	}

	if row.CurrentTurnPlayer.Valid {
		currentPlayer, err := snowflake.Parse(row.CurrentTurnPlayer.String)
		if err == nil {
			battle.CurrentTurnPlayer = &currentPlayer
		}
	}

	if row.MainThreadID.Valid {
		mainThreadID, err := snowflake.Parse(row.MainThreadID.String)
		if err == nil {
			battle.MainThreadID = &mainThreadID
		}
	}

	if row.ChallengerThreadID.Valid {
		challengerThreadID, err := snowflake.Parse(row.ChallengerThreadID.String)
		if err == nil {
			battle.ChallengerThreadID = &challengerThreadID
		}
	}

	if row.OpponentThreadID.Valid {
		opponentThreadID, err := snowflake.Parse(row.OpponentThreadID.String)
		if err == nil {
			battle.OpponentThreadID = &opponentThreadID
		}
	}

	if row.CompletedAt.Valid {
		battle.CompletedAt = &row.CompletedAt.Time
	}

	return battle, nil
}

func dbBattleTeamMemberToModel(row BattleTeam) (*entities.BattleTeamMember, error) {
	var characterData entities.Character
	if err := json.Unmarshal(row.CharacterData, &characterData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal character data: %w", err)
	}

	var statStages map[string]int
	if err := json.Unmarshal(row.StatStages, &statStages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stat stages: %w", err)
	}

	userID, err := snowflake.Parse(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	statusEffects := make([]constants.StatusEffect, len(row.StatusEffects))
	for i, effect := range row.StatusEffects {
		statusEffects[i] = constants.StatusEffect(effect)
	}

	return &entities.BattleTeamMember{
		ID:            row.ID,
		BattleID:      row.BattleID,
		UserID:        userID,
		TeamPosition:  int(row.TeamPosition),
		CharacterID:   row.CharacterID,
		CharacterData: characterData,
		CurrentHP:     int(row.CurrentHp),
		StatusEffects: statusEffects,
		StatStages:    statStages,
		IsActive:      row.IsActive,
		IsFainted:     row.IsFainted,
		CreatedAt:     row.CreatedAt,
	}, nil
}

func dbBattleTurnToModel(row BattleTurn) (*entities.BattleTurn, error) {
	var actionData map[string]interface{}
	if err := json.Unmarshal(row.ActionData, &actionData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal action data: %w", err)
	}

	var resultData map[string]interface{}
	if err := json.Unmarshal(row.ResultData, &resultData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result data: %w", err)
	}

	userID, err := snowflake.Parse(row.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return &entities.BattleTurn{
		ID:         row.ID,
		BattleID:   row.BattleID,
		TurnNumber: int(row.TurnNumber),
		UserID:     userID,
		ActionType: entities.ActionType(row.ActionType),
		ActionData: actionData,
		ResultData: resultData,
		CreatedAt:  row.CreatedAt,
	}, nil
}

func dbUserEloToModel(row UserElo) *entities.UserElo {
	userID, _ := snowflake.Parse(row.UserID)

	return &entities.UserElo{
		UserID:       userID,
		EloRating:    int(row.EloRating),
		BattlesWon:   int(row.BattlesWon),
		BattlesLost:  int(row.BattlesLost),
		BattlesTotal: int(row.BattlesTotal),
		HighestElo:   int(row.HighestElo),
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}
