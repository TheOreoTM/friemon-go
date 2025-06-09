package components

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
)

type componentHandler func(b *bot.Bot) handler.ComponentHandler

var Components = map[string]componentHandler{}

func ErrorMessage(err string) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			SimpleEmbed("Error", err, constants.ColorFail),
		},
	}
}
func SimpleEmbed(title, desc string, color int) discord.Embed {
	return discord.Embed{
		Title:       title,
		Description: desc,
		Color:       int(color),
	}
}
