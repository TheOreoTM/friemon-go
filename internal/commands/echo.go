package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/theoreotm/gordinal/internal/handler"
)

func init() {
	cmds["echo"] = &handler.Command{
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

func messageRunEcho(s *discordgo.Session, m *discordgo.MessageCreate, args *handler.Args) {
	s.ChannelMessageSend(m.ChannelID, m.Content)
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
