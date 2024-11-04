package db

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/model"
)

// var _ Store = (*Queries)(nil)

func (q *Queries) CreateCharacter(ctx context.Context, ownerID snowflake.ID) (*model.Character, error) {
	randomChar := model.NewCharacter(ownerID.String())
	_, err := q.createCharacter(ctx, modelCharToDBChar(randomChar))
	if err != nil {
		return &model.Character{}, err
	}

	return randomChar, nil
}

func (q *Queries) GetCharactersForUser(ctx context.Context, userID snowflake.ID) ([]model.Character, error) {
	dbchs, err := q.getCharactersForUser(ctx, userID.String())
	if err != nil {
		return nil, err
	}

	chars := make([]model.Character, 0, len(dbchs))
	for _, dbch := range dbchs {
		chars = append(chars, dbCharToModelChar(dbch))
	}

	return chars, nil
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

func modelCharToDBChar(ch *model.Character) createCharacterParams {
	return createCharacterParams{
		OwnerID:          ch.OwnerID,
		ClaimedTimestamp: ch.ClaimedTimestamp,
		Idx:              ch.IDX,
		CharacterID:      ch.CharacterID,
		Level:            ch.Level,
		Xp:               ch.Xp,
		Personality:      ch.Personality.String(),
		Shiny:            ch.Shiny,
		IvHp:             ch.IvHP,
		IvAtk:            ch.IvAtk,
		IvDef:            ch.IvDef,
		IvSpAtk:          ch.IvSpAtk,
		IvSpDef:          ch.IvSpDef,
		IvSpd:            ch.IvSpd,
		IvTotal:          ch.IvTotal,
		Nickname:         ch.Nickname,
		Favourite:        ch.Favourite,
		HeldItem:         ch.HeldItem,
		Moves:            ch.Moves,
		Color:            ch.Color,
	}
}

func dbCharToModelChar(dbch Character) model.Character {
	return model.Character{
		ID:               dbch.ID.String(),
		OwnerID:          dbch.OwnerID,
		ClaimedTimestamp: dbch.ClaimedTimestamp,
		IDX:              dbch.Idx,
		CharacterID:      dbch.CharacterID,
		Level:            dbch.Level,
		Xp:               dbch.Xp,
		Personality:      stringToPersonality(dbch.Personality),
		Shiny:            dbch.Shiny,
		IvHP:             dbch.IvHp,
		IvAtk:            dbch.IvAtk,
		IvDef:            dbch.IvDef,
		IvSpAtk:          dbch.IvSpAtk,
		IvSpDef:          dbch.IvSpDef,
		IvSpd:            dbch.IvSpd,
		IvTotal:          dbch.IvTotal,
		Nickname:         dbch.Nickname,
		Favourite:        dbch.Favourite,
		HeldItem:         dbch.HeldItem,
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
