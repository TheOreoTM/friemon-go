package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

func init() {
	Commands[cmdSelect.Cmd.CommandName()] = cmdSelect
}

var cmdSelect = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "select",
		Description: "Select a character",
	},
	Handler: handleSelect,
}

func handleSelect(b *friemon.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Not implemented",
		})
	}
}
