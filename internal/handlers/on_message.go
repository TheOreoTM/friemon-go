package handlers

import (
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/theoreotm/friemon/internal/bot"
)

func OnMessage(b *bot.Bot) disgobot.EventListener {
	return disgobot.NewListenerFunc(func(e *events.MessageCreate) {
		if e.Message.Author.Bot {
			return
		}

		spawnCharacter(b, e)
		incrementXp(b, e)
	})
}
