package components

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

func init() {
	Components["/claim"] = claimCharacterButton
}

func claimCharacterButton(b *friemon.Bot) handler.ComponentHandler {
	return func(e *handler.ComponentEvent) error {
		characterToClaim, err := b.Cache.GetChannelCharacter(e.Channel().ID())
		if err != nil {
			e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetContentf("Error: %s", err).
					Build())
		}

		if characterToClaim == nil {
			e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetContentf("No character to claim").
					Build())

			e.Client().Rest().DeleteMessage(e.Message.ChannelID, e.Message.ID)
			return nil
		}

		e.Respond(discord.InteractionResponseTypeCreateMessage, discord.NewMessageCreateBuilder().
			SetContentf("You basically claimed %v", characterToClaim.CharacterName()).
			Build(),
		)

		return nil
	}
}
