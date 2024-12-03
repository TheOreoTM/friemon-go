package components

import (
	"errors"

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

		if characterToClaim == nil {
			e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetContentf("No character to claim").
					Build())
			return nil
		}

		if err != nil {
			e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetContentf("Error: %s", err).
					Build())
			return err
		}

		b.Cache.DeleteChannelCharacter(e.Channel().ID())

		button, exists := e.Message.ButtonByID("/claim")
		if !exists {
			return errors.New("failed to find button")
		}

		e.Client().Rest().UpdateMessage(
			e.Message.ChannelID,
			e.Message.ID,
			discord.NewMessageUpdateBuilder().
				AddActionRow(button.AsDisabled()).
				Build())

		e.Respond(discord.InteractionResponseTypeCreateMessage, discord.NewMessageCreateBuilder().
			SetContentf("Congratulations %v! You claimed a %v (%v)", e.Member(), characterToClaim.Format("l"), characterToClaim.IvPercentage()).
			Build(),
		)

		characterToClaim.OwnerID = e.Member().User.ID.String()
		b.DB.CreateCharacter(e.Ctx, e.Member().User.ID, characterToClaim)

		return nil
	}
}
