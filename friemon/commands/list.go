package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

var list = discord.SlashCommandCreate{
	Name:        "list",
	Description: "Get a list of characters you own",
}

func ListHandler(b *friemon.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {

		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Check console, you have %d characters", len("")),
		})
	}
}
