package db

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/entities"
)

var _ Store = (*Queries)(nil)

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

func (q *Queries) UpdateCharacter(ctx context.Context, id uuid.UUID, ch entities.Character) (*entities.Character, error) {

	dbch, err := q.updateCharacter(ctx, updateCharacterParams{
		ID:               uuid.MustParse(ch.ID),
		OwnerID:          ch.OwnerID,
		ClaimedTimestamp: ch.ClaimedTimestamp,
		Idx:              int32(ch.IDX),
		CharacterID:      int32(ch.CharacterID),
		Level:            int32(ch.Level),
		Xp:               int32(ch.Xp),
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

func (q *Queries) CreateCharacter(ctx context.Context, ownerID snowflake.ID) (*entities.Character, error) {
	randomChar := entities.NewCharacter(ownerID.String())
	_, err := q.createCharacter(ctx, modelCharToDBChar(randomChar))
	if err != nil {
		return &entities.Character{}, err
	}

	return randomChar, nil
}

func (q *Queries) GetCharactersForUser(ctx context.Context, userID snowflake.ID) (*[]entities.Character, error) {
	dbchs, err := q.getCharactersForUser(ctx, userID.String())
	if err != nil {
		return nil, err
	}

	chars := make([]entities.Character, 0, len(dbchs))
	for _, dbch := range dbchs {
		chars = append(chars, *dbCharToModelChar(dbch))
	}

	return &chars, nil
}

// func isValidOrderBy(orderBy string) bool {
// 	return orderBy == "idx" || orderBy == "level" || orderBy == "xp" || orderBy == "iv_total"
// }

// func isValidFilter(filter []string) bool {
// 	for _, f := range filter {
// 		if f != "shiny" && f != "favourite" && f != "held_item" && f != "moves" {
// 			return false
// 		}
// 	}
// 	return true
// }

// func isValidSort(sort string) bool {
// 	return sort == "asc" || sort == "desc"
// }

func modelCharToDBChar(ch *entities.Character) createCharacterParams {
	return createCharacterParams{
		OwnerID:          ch.OwnerID,
		ClaimedTimestamp: ch.ClaimedTimestamp,
		CharacterID:      int32(ch.CharacterID),
		Level:            int32(ch.Level),
		Xp:               int32(ch.Xp),
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
		ID:               dbch.ID.String(),
		OwnerID:          dbch.OwnerID,
		ClaimedTimestamp: dbch.ClaimedTimestamp,
		IDX:              int(dbch.Idx),
		CharacterID:      int(dbch.CharacterID),
		Level:            int(dbch.Level),
		Xp:               int(dbch.Xp),
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
