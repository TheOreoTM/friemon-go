package components

import (
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

type componentHandler func(b *friemon.Bot) handler.ComponentHandler

var Components = map[string]componentHandler{}
