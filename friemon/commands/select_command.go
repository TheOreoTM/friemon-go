package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/entities"
	"github.com/theoreotm/friemon/friemon"
)

func init() {
	Commands[cmdSelect.Cmd.CommandName()] = cmdSelect
}

var cmdSelect = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "select",
		Description: "Select a character",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:         "character",
				Description:  "The character you want select",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	Handler:      handleSelect,
	Autocomplete: handleGetCharacterAutocomplete,
}

func handleSelect(b *friemon.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		id := e.SlashCommandInteractionData().String("character")
		if id == "" || id == "-1" {
			return e.CreateMessage(ErrorMessage("Select a valid character"))
		}

		targetChar, err := b.DB.GetCharacter(b.Context, uuid.MustParse(id))
		if err != nil {
			return e.CreateMessage(ErrorMessage(err.Error()))
		}

		_, err = b.DB.UpdateUser(b.Context, entities.User{
			ID:         e.Member().User.ID,
			SelectedID: targetChar.ID,
		})

		if err != nil {
			return e.CreateMessage(ErrorMessage(err.Error()))
		}

		return e.CreateMessage(discord.MessageCreate{
			Content: fmt.Sprintf("You selected your **%v (%v) No. %v**", targetChar.Format("l"), targetChar.IvPercentage(), targetChar.IDX),
		})
	}
}
