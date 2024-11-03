package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
	"github.com/theoreotm/friemon/friemon/db"
	"github.com/theoreotm/friemon/model"
)

var character = discord.SlashCommandCreate{
	Name:        "character",
	Description: "Generate a random character",
}

func CharacterHandler(b *friemon.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		randomCharacter := model.NewCharacter(e.Member().User.ID.String())
		c, err := b.Database.CreateCharacterForUser(e.Ctx, db.CreateCharacterForUserParams{
			
			OwnerID:          randomCharacter.OwnerID,
			ClaimedTimestamp: randomCharacter.ClaimedTimestamp,
			Idx:              randomCharacter.IDX,
			CharacterID:      randomCharacter.CharacterID,
			Level:            randomCharacter.Level,
			Xp:               randomCharacter.Xp,
			Personality:      randomCharacter.Personality.String(),
			Shiny:            randomCharacter.Shiny,
			IvHp:             randomCharacter.IvHP,
			IvAtk:            randomCharacter.IvAtk,
			IvDef:            randomCharacter.IvDef,
			IvSpAtk:          randomCharacter.IvSpAtk,
			IvSpDef:          randomCharacter.IvSpDef,
			IvSpd:            randomCharacter.IvSpd,
			IvTotal:          randomCharacter.IvTotal,
			Nickname:         randomCharacter.Nickname,
			Favourite:        randomCharacter.Favourite,
			HeldItem:         randomCharacter.HeldItem,
			Moves:            randomCharacter.Moves,
			Color:            randomCharacter.Color,
		})

		if err != nil {
			return err
		}

		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("IV percent %.2f for %v", randomCharacter.IvPercentage()*100, c.IvTotal),
		})
	}
}
