package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func init() {
	cmds["ping"] = &Command{
		Meta: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Replies with Pong!",
		},
		ChatInputRun: chatInputRun,
		MessageRun:   messageRun,
	}
}

func messageRun(s *discordgo.Session, m *discordgo.MessageCreate, args ...string) {
	fmt.Println(args)

	latency := s.HeartbeatLatency().Milliseconds()

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Pong! %vms", latency))
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("These are the args provided to the command %v", args))
}

func chatInputRun(s *discordgo.Session, i *discordgo.InteractionCreate) {
	latency := s.HeartbeatLatency().Milliseconds()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong! %vms", latency),
		},
	})

}
