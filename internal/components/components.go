package components

import (
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/internal/bot"
)

type componentHandler func(b *bot.Bot) handler.ComponentHandler

var Components = map[string]componentHandler{}
