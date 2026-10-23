package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// commands is the list registered with Discord on startup.
var commands = []*discordgo.ApplicationCommand{
	{Name: "ping", Description: "Check that the bot is alive"},
}

// handlers maps a command name to the function that answers it.
var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"ping": handlePing,
}

func handlePing(s *discordgo.Session, i *discordgo.InteractionCreate) {
	reply(s, i, fmt.Sprintf("Pong! Gateway latency: %dms", s.HeartbeatLatency().Milliseconds()))
}

// reply sends a plain text response to a slash command.
func reply(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Content: content},
	})
	if err != nil {
		log.Printf("reply failed: %v", err)
	}
}
