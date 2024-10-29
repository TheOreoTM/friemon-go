package discord

import "github.com/Karitham/corde"

type Bot struct {
	mux       *corde.Mux
	AppID     corde.Snowflake
	GuildID   *corde.Snowflake
	PublicKey string
	BotToken  string
}

// New runs the bot
func New(b *Bot) *corde.Mux {
	b.mux = corde.NewMux(b.PublicKey, b.AppID, b.BotToken)

	return b.mux
}
