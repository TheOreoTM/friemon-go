package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/game"
)

func init() {
	Commands["character"] = cmdCharacter
}

var cmdCharacter = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "character",
		Description: "Generate a random character",
	},
	Handler:  HandleCharacter,
	Category: "Friemon",
}

func HandleCharacter(b *bot.Bot) handler.CommandHandler {

	return func(e *handler.CommandEvent) error {
		character := game.RandomCharacterSpawn()
		user, err := b.DB.EnsureUser(e.Ctx, e.User().ID)
		if err != nil {
			return e.CreateMessage(ErrorMessage(err.Error()))
		}

		fmt.Printf("user: %+v\n", user)

		dbCharacter, err := b.DB.CreateCharacter(e.Ctx, user.ID, character)
		if err != nil {
			return e.CreateMessage(ErrorMessage(err.Error()))
		}

		// Update the user's selected character
		user.SelectedID = dbCharacter.ID
		if _, err := b.DB.UpdateUser(e.Ctx, *user); err != nil {
			return e.CreateMessage(ErrorMessage(err.Error()))
		}

		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("Character created: %s", character.Format("l")),
		})
	}
}
