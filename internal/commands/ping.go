package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/theoreotm/gordinal/internal/handler"
)

func init() {
	cmds["ping"] = &handler.Command{
		Meta: &discordgo.ApplicationCommand{
			Name:        "ping",
			Description: "Replies with Pong!",
		},
		ChatInputRun: chatInputRunPing,
		MessageRun:   messageRunPing,
	}
}

func messageRunPing(s *discordgo.Session, m *discordgo.MessageCreate, args *handler.Args) {
	fmt.Println(args)

	latency := s.HeartbeatLatency().Milliseconds()

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Pong! %vms", latency))
}

func chatInputRunPing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	latency := s.HeartbeatLatency().Milliseconds()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Pong! %vms", latency),
		},
	})

}
