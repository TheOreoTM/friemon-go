package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

var selected = discord.SlashCommandCreate{
	Name:        "selected",
	Description: "Generate a random character",
}

func SelectedHandler(b *friemon.Bot) handler.CommandHandler {
	return func(e *handler.CommandEvent) error {
		ch, err := b.DB.GetSelectedCharacter(e.Ctx, e.Member().User.ID)
		if err != nil {
			return e.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Error: %s", err),
			})
		}

		var detailFieldValues = [][]string{}
		detailFieldValues = append(detailFieldValues, []string{"XP", fmt.Sprintf("%d", ch.Xp)})
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
		statFieldValues = append(statFieldValues, []string{"Total IV", fmt.Sprintf("%v", fmt.Sprintf("%.2f", ch.IvPercentage()*100)+"%")})
		statFieldContent := ""
		for _, v := range statFieldValues {
			statFieldContent += fmt.Sprintf("**%s:** %s\n", v[0], v[1])
		}

		embedSmallImage := e.Member().EffectiveAvatarURL()

		embed := discord.NewEmbedBuilder().
			SetTitle(fmt.Sprintf("%v", ch)).
			SetThumbnail(embedSmallImage).
			SetFooter(fmt.Sprintf("Displaying character %v. \nID: %v", ch.IDX, ch.ID), "").
			SetImage("attachment://character.png").
			AddFields(
				discord.EmbedField{
					Name:  "Details",
					Value: detailFieldContent,
				},
				discord.EmbedField{
					Name:  "Stats",
					Value: statFieldContent,
				}).Build()

		// read the character image from the assets folder

		loa, err := loadImage(fmt.Sprintf("./assets/characters/%v.png", ch.CharacterID))
		fmt.Println()
		if err != nil {
			return e.CreateMessage(discord.MessageCreate{
				Content: fmt.Sprintf("Error: %s", err),
			})
		}
		embedImage := discord.NewFile("character.png", "", loa)

		return e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{embed},
			Files:  []*discord.File{embedImage},
		})
	}
}

func loadImage(filePath string) (io.Reader, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
