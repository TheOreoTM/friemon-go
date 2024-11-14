package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/friemon"
)

var Commands = map[string]*Command{}

type Command struct {
	Cmd     discord.ApplicationCommandCreate
	Handler func(b *friemon.Bot) handler.CommandHandler
}

func SuccessMessage(title, desc string) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			SimpleEmbed(title, desc, constants.ColorSuccess),
		},
	}
}

func ErrorMessage(err string) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			SimpleEmbed("Error", err, constants.ColorFail),
		},
	}
}

func InfoMessage(message string) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			SimpleEmbed("", message, constants.ColorInfo),
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
