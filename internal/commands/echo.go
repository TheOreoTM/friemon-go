package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	cmds["echo"] = &Command{
		Meta: &discordgo.ApplicationCommand{
			Name:        "echo",
			Description: "Echos your message!",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "message",
					Description: "Message to echo",
					Required:    true,
				},
			},
		},
		MessageRun:   messageRunEcho,
		ChatInputRun: chatInputRunEcho,
	}
}

func messageRunEcho(s *discordgo.Session, m *discordgo.MessageCreate, args ...string) {
	s.ChannelMessageSend(m.ChannelID, strings.Join(args, " "))
}

func chatInputRunEcho(s *discordgo.Session, i *discordgo.InteractionCreate) {
	messageToEcho := i.ApplicationCommandData().Options[0].StringValue()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: messageToEcho,
		},
	})
}
