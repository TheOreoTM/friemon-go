package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/entities"
	"github.com/theoreotm/friemon/friemon"
)

var character = discord.SlashCommandCreate{
	Name:        "character",
	Description: "Generate a random character",
}

func CharacterHandler(b *friemon.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		randomCharacter := entities.NewCharacter(e.Member().User.ID.String())

		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("IV percent %.2f for %v", randomCharacter.IvPercentage()*100, randomCharacter.IvTotal),
		})
	}
}
