package handlers

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/theoreotm/friemon/friemon"
)

func MessageHandler(b *friemon.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageCreate) {
		// TODO: handle message
	})
}
