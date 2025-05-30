package commands

import (
	"fmt"
	"math"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/paginator"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/entities"
)

func init() {
	Commands[cmdList.Cmd.CommandName()] = cmdList
}

const (
	characterPerPage = 20
)

var cmdList = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "list",
		Description: "Get a list of characters you own",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "page",
				Description: "The page you want to view",
				Required:    false,
			},
		},
	},
	Handler:  handlelist,
	Category: "Friemon",
}

func handlelist(b *bot.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		characters, err := b.DB.GetCharactersForUser(e.Ctx, e.Member().User.ID)
		if err != nil {
			return e.CreateMessage(ErrorMessage(err.Error()))
		}

		if len(characters) == 0 {
			return e.CreateMessage(InfoMessage("You don't have any characters"))
		}

		dbUser, err := b.DB.GetUser(e.Ctx, e.Member().User.ID)
		if err != nil {
			return e.CreateMessage(ErrorMessage(err.Error()))
		}

		return b.Paginator.Create(e.Respond, paginator.Pages{
			ID: e.ID().String(),
			PageFunc: func(page int, embed *discord.EmbedBuilder) {
				characterStart := page * characterPerPage
				characterEnd := characterStart + characterPerPage

				if characterEnd > len(characters) {
					characterEnd = len(characters)
				}

				charactersInPage := characters[characterStart:characterEnd]

				highestIdx := maxIDX(charactersInPage)
				if highestIdx == -1 {
					embed.SetDescription("No characters found")
					return
				}

				description := ""
				for i := 0; i < len(charactersInPage); i++ {
					character := charactersInPage[i]
					idx := padIdx(character.IDX, len(fmt.Sprint(highestIdx)))
					name := character.Format("inf")
					if character.ID == dbUser.SelectedID {
						idx = fmt.Sprintf("**`%v`**", idx)
					} else {
						idx = fmt.Sprintf("`%v`", idx)
					}

					description += fmt.Sprintf("%v　%v　•　Lvl. %v　•　%v\n", idx, name, character.Level, character.IvPercentage())
				}
				embed.SetDescription(description)
				embed.SetColor(constants.ColorDefault)
				embed.SetFooterTextf("Showing entries %v-%v out of %v", characterStart+1, characterEnd, len(characters))
			},

			Pages:      int(math.Ceil(float64(len(characters)) / characterPerPage)),
			Creator:    e.User().ID,
			ExpireMode: paginator.ExpireModeAfterLastUsage,
		}, false)
	}
}

func maxIDX(characters []entities.Character) int {
	if len(characters) == 0 {
		return -1
	}

	maxIdx := characters[0].IDX
	for _, character := range characters {
		if character.IDX > maxIdx {
			maxIdx = character.IDX
		}
	}
	return maxIdx
}

func padIdx(idx int, width int) string {
	padding := strings.Repeat(" ", width-len(fmt.Sprint(idx)))
	return padding + fmt.Sprint(idx)
}
