package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/internal/bot"
)

func init() {
	Commands[test.Cmd.CommandName()] = test
}

var test = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "test",
		Description: "test command",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:         "choice",
				Description:  "some autocomplete choice",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	Handler:      handleTest,
	Autocomplete: handleTestAutocomplete,
	Category:     "Friemon",
}

func handleTest(b *bot.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		return e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContentf("test command. Choice: %s", e.SlashCommandInteractionData().String("choice")).
			AddActionRow(discord.NewPrimaryButton("test", "/test-button")).
			Build(),
		)
	}
}

func handleTestAutocomplete(b *bot.Bot) handler.AutocompleteHandler {
	return func(e *handler.AutocompleteEvent) error {
		return e.AutocompleteResult([]discord.AutocompleteChoice{
			discord.AutocompleteChoiceString{
				Name:  "1",
				Value: "1",
			},
			discord.AutocompleteChoiceString{
				Name:  "2",
				Value: "2",
			},
			discord.AutocompleteChoiceString{
				Name:  "3",
				Value: "3",
			},
		})
	}
}
