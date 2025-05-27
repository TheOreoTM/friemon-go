package components

import (
	"errors"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

var claimMutex sync.Mutex

func init() {
	Components["/claim"] = claimCharacterButton
}

func claimCharacterButton(b *friemon.Bot) handler.ComponentHandler {
	return func(e *handler.ComponentEvent) error {
		claimMutex.Lock()
		defer claimMutex.Unlock()

		characterToClaim, err := b.Cache.GetChannelCharacter(e.Channel().ID())
		if characterToClaim == nil {
			e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetContentf("No character to invite").
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

		if characterToClaim.OwnerID != "" {
			e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetContentf("This character has already been invited.").
					Build())
			return nil
		}

		characterToClaim.OwnerID = e.Member().User.ID.String()
		b.Cache.DeleteChannelCharacter(e.Channel().ID())

		button, exists := e.Message.ButtonByID("/claim")
		if !exists {
			return errors.New("failed to find button")
		}

		e.Client().Rest().UpdateMessage(
			e.Message.ChannelID,
			e.Message.ID,
			discord.NewMessageUpdateBuilder().
				AddActionRow(button.AsDisabled().WithLabel("Invited by "+e.Member().User.Username)).
				Build())

		e.Respond(discord.InteractionResponseTypeCreateMessage,
			discord.NewMessageCreateBuilder().
				SetContentf("Congratulations %v! You invited a %v (%v) to the party!", e.Member(), characterToClaim.Format("l"), characterToClaim.IvPercentage()).
				Build(),
		)

		err = b.DB.CreateCharacter(e.Ctx, e.Member().User.ID, characterToClaim)
		if err != nil {
			return err
		}

		return nil
	}
}
