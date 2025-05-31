package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/core/entities"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

var _ Store = (*Queries)(nil)

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
