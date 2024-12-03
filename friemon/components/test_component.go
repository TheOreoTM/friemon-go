package components

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/json"
	"github.com/theoreotm/friemon/friemon"
)

func init() {
	Components["/test-button"] = testComponent
}

func testComponent(b *friemon.Bot) handler.ComponentHandler {
	return func(e *handler.ComponentEvent) error {
		return e.UpdateMessage(discord.MessageUpdate{
			Content: json.Ptr("This is a test button update"),
		})
	}
}
