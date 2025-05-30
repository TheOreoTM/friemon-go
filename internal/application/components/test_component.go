package components

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/theoreotm/friemon/internal/application/bot"
)

func init() {
	Components["/test-button"] = testComponent
}

func testComponent(b *bot.Bot) handler.ComponentHandler {
	return func(e *handler.ComponentEvent) error {
		return e.UpdateMessage(discord.MessageUpdate{
			Content: json.Ptr("This is a test button update"),
		})
	}
}
