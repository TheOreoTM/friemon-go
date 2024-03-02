package commands

import (
	"github.com/bwmarrin/discordgo"
)

func init() {
	cmds["ping"] = &command{
		Meta: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Replies with Pong!",
		},
		ChatInputRun: chatInputRun,
	}
}

func chatInputRun(s *discordgo.Session, i *discordgo.InteractionCreate) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})
}
