package commands

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/entities"
	"github.com/theoreotm/friemon/internal/bot"
)

func init() {
	Commands[cmdInfo.Cmd.CommandName()] = cmdInfo
}

var cmdInfo = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "info",
		Description: "Get your current character",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:         "character",
				Description:  "The character you want to get info about",
				Required:     false,
				Autocomplete: true,
			},
		},
	},
	Autocomplete: handleGetCharacterAutocomplete,
	Handler:      handleInfo,
	Category:     "Friemon",
}

func handleInfo(b *bot.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		id := e.SlashCommandInteractionData().String("character")
		var ch *entities.Character

		if id != "" && id != "-1" {
			dbChar, err := b.DB.GetCharacter(e.Ctx, uuid.MustParse(id))
			if err != nil {
				ch = nil
			}
			ch = dbChar
		}

		if ch == nil {
			selectedCh, err := b.DB.GetSelectedCharacter(e.Ctx, e.Member().User.ID)
			if err != nil {
				return e.CreateMessage(discord.MessageCreate{
					Content: fmt.Sprintf("Error: %s", err),
				})
			}
			ch = selectedCh
		}

		var detailFieldValues = [][]string{}
		detailFieldValues = append(detailFieldValues, []string{"XP", fmt.Sprintf("%d/%d", ch.XP, ch.MaxXP())})
		detailFieldValues = append(detailFieldValues, []string{"Personality", ch.Personality.String()})

		detailFieldContent := ""
		for _, v := range detailFieldValues {
			detailFieldContent += fmt.Sprintf("**%s:** %s\n", v[0], v[1])
		}

		var statFieldValues = [][]string{}
		statFieldValues = append(statFieldValues, []string{"HP", fmt.Sprintf("%d – IV: %d/31", ch.MaxHP(), ch.IvHP)})
		statFieldValues = append(statFieldValues, []string{"Attack", fmt.Sprintf("%d – IV: %d/31", ch.Atk(), ch.IvAtk)})
		statFieldValues = append(statFieldValues, []string{"Defense", fmt.Sprintf("%d – IV: %d/31", ch.Def(), ch.IvDef)})
		statFieldValues = append(statFieldValues, []string{"Sp. Atk", fmt.Sprintf("%d – IV: %d/31", ch.SpAtk(), ch.IvSpAtk)})
		statFieldValues = append(statFieldValues, []string{"Sp. Def", fmt.Sprintf("%d – IV: %d/31", ch.SpDef(), ch.IvSpDef)})
		statFieldValues = append(statFieldValues, []string{"Speed", fmt.Sprintf("%d – IV: %d/31", ch.Spd(), ch.IvSpd)})
		statFieldValues = append(statFieldValues, []string{"Total IV", ch.IvPercentage()})
		statFieldContent := ""
		for _, v := range statFieldValues {
			statFieldContent += fmt.Sprintf("**%s:** %s\n", v[0], v[1])
		}

		embedSmallImage := e.Member().EffectiveAvatarURL()

		embed := discord.NewEmbedBuilder().
			SetTitle(fmt.Sprintf("%v", ch)).
			SetThumbnail(embedSmallImage).
			SetImage("attachment://character.png").
			SetFooterTextf("Displaying character %v \nID: %v", ch.IDX, ch.ID).
			SetColor(constants.ColorDefault).
			AddFields(
				discord.EmbedField{
					Name:  "Details",
					Value: detailFieldContent,
				},
				discord.EmbedField{
					Name:  "Stats",
					Value: statFieldContent,
				}).Build()

		image, err := ch.Image()
		if err != nil {
			return e.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Error: %s", err),
			})
		}

		return e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{embed},
			Files:  []*discord.File{image},
		})
	}
}

func handleGetCharacterAutocomplete(b *bot.Bot) handler.AutocompleteHandler {
	return func(e *handler.AutocompleteEvent) error {
		query := e.Data.String("character")
		var results []discord.AutocompleteChoiceString
		chars, err := b.DB.GetCharactersForUser(e.Ctx, e.Member().User.ID)
		if err != nil {
			return e.AutocompleteResult([]discord.AutocompleteChoice{
				discord.AutocompleteChoiceString{
					Name:  "You dont have any characters",
					Value: "-1",
				},
			})
		}

		for _, ch := range chars {
			nameId := strings.ToLower(fmt.Sprintf("%v %v", ch.CharacterName(), ch.IDX))
			if strings.Contains(nameId, query) {
				results = append(results, discord.AutocompleteChoiceString{
					Name:  fmt.Sprintf("%v - Level %v %v", ch.IDX, ch.Level, ch.CharacterName()),
					Value: ch.ID.String(),
				})
			}
		}

		var choices []discord.AutocompleteChoice
		for i, r := range results {
			if i >= 25 {
				break
			}
			choices = append(choices, r)
		}

		return e.AutocompleteResult(choices)
	}
}
