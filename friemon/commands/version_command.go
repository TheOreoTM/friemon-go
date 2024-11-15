package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

func init() {
	Commands[version.Cmd.CommandName()] = version
}

var version = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "version",
		Description: "version command",
	},
	Handler: handleVersion,
}

func handleVersion(b *friemon.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Version: %s\nCommit: %s", b.Version, b.Commit),
		})
	}
}
