package commands

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

func init() {
	Commands[cmdInfo.Cmd.CommandName()] = cmdInfo
}

var cmdInfo = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "info",
		Description: "Get your current character",
	},
	Handler: handleInfo,
}

func handleInfo(b *friemon.Bot) handler.CommandHandler {
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
		statFieldValues = append(statFieldValues, []string{"Total IV", ch.IvPercentage()})
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
