package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/core/entities"
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
	user, err := q.GetUser(ctx, ownerID)
	if err != nil {
		return err
	}

	char.IDX = user.NextIdx

	_, err = q.createCharacter(ctx, modelCharToDBChar(char))
	if err != nil {
		return err
	}

	user.NextIdx++
	_, err = q.UpdateUser(ctx, *user)
	if err != nil {
		return err
	}

	return nil
}

func (q *Queries) GetCharactersForUser(ctx context.Context, userID snowflake.ID) ([]entities.Character, error) {
	dbchs, err := q.getCharactersForUser(ctx, userID.String())
	if err != nil {
		return nil, err
	}

	chars := make([]entities.Character, 0, len(dbchs))
	for _, dbch := range dbchs {
		chars = append(chars, *dbCharToModelChar(dbch))
	}

	return chars, nil
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
