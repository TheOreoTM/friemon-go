package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

func init() {
	Commands[cmdCharacter.Cmd.CommandName()] = cmdCharacter
}

var cmdCharacter = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "character",
		Description: "Generate a random character",
	},
	Handler: handleCharacter,
}

func handleCharacter(b *friemon.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		ch, err := b.DB.CreateCharacter(e.Ctx, e.Member().User.ID)
		if err != nil {
			return e.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Error: %s", err),
			})
		}

		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("IV percent %v for %v | %v", ch.IvPercentage(), ch.IvTotal, ch.Data().Name),
		})
	}
}
