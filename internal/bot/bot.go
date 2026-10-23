// Package bot wires Discord events to game logic.
package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// Bot holds the Discord session and config.
type Bot struct {
	session *discordgo.Session
	guildID string // if set, commands register to this guild only (instant updates while developing)
}

// New creates a bot. It does not connect yet; call Start for that.
func New(token, guildID string) (*Bot, error) {
	s, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	// Slash commands need no privileged intents.
	s.Identify.Intents = discordgo.IntentsGuilds

	b := &Bot{session: s, guildID: guildID}
	s.AddHandler(b.onInteraction)
	return b, nil
}

// Start opens the gateway connection and registers slash commands.
func (b *Bot) Start() error {
	if err := b.session.Open(); err != nil {
		return fmt.Errorf("open gateway: %w", err)
	}

	// BulkOverwrite replaces the full command list, so removed commands disappear too.
	appID := b.session.State.User.ID
	if _, err := b.session.ApplicationCommandBulkOverwrite(appID, b.guildID, commands); err != nil {
		return fmt.Errorf("register commands: %w", err)
	}

	log.Printf("logged in as %s, %d command(s) registered", b.session.State.User.Username, len(commands))
	return nil
}

// Close disconnects from Discord.
func (b *Bot) Close() error {
	return b.session.Close()
}

func (b *Bot) onInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	name := i.ApplicationCommandData().Name
	if h, ok := handlers[name]; ok {
		h(s, i)
	}
}
