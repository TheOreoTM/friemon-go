package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/internal/bot"
)

func init() {
	Commands[version.Cmd.CommandName()] = version
}

var version = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "version",
		Description: "version command",
	},
	Handler:  handleVersion,
	Category: "Bot",
}

func handleVersion(b *bot.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Version: %s\nCommit: %s\nBranch: %s", b.BuildInfo.Version, b.BuildInfo.Commit, b.BuildInfo.Branch),
		})
	}
}
