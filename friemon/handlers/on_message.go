package handlers

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/theoreotm/friemon/friemon"
)

func OnMessage(b *friemon.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageCreate) {
		if e.Message.Author.Bot {
			return
		}

		spawnCharacter(b, e)
		incrementXp(b, e)
	})
}
