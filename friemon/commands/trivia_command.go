package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
	"github.com/theoreotm/friemon/friemon/handlers/trivia"
)

var TriviaCommand = Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "trivia",
		Description: "Start a trivia game",
	},
	Handler: func(b *friemon.Bot) handler.CommandHandler {
		return func(e *handler.CommandEvent) error {
			triviaHandler := trivia.NewTriviaHandler(b, b.Redis)
			return triviaHandler.StartTrivia(e.ApplicationCommandInteractionCreate)
		}
	},
}
