package components

import (
	"errors"
	"sync"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
)

// Mutex to protect claim logic
var claimMutex sync.Mutex

func init() {
	Components["/claim"] = claimCharacterButton
}

func claimCharacterButton(b *friemon.Bot) handler.ComponentHandler {
	return func(e *handler.ComponentEvent) error {
		claimMutex.Lock()
		defer claimMutex.Unlock()

		// Get the character to claim
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

		// Check if the character is already claimed
		if characterToClaim.OwnerID != "" {
			e.Respond(discord.InteractionResponseTypeCreateMessage,
				discord.NewMessageCreateBuilder().
					SetContentf("This character has already been claimed.").
					Build())
			return nil
		}

		// Mark the character as claimed
		characterToClaim.OwnerID = e.Member().User.ID.String()
		b.Cache.DeleteChannelCharacter(e.Channel().ID())

		// Disable the claim button
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

		// Respond to the user
		e.Respond(discord.InteractionResponseTypeCreateMessage,
			discord.NewMessageCreateBuilder().
				SetContentf("Congratulations %v! You claimed a %v (%v)", e.Member(), characterToClaim.Format("l"), characterToClaim.IvPercentage()).
				Build(),
		)

		// Save the character to the database
		err = b.DB.CreateCharacter(e.Ctx, e.Member().User.ID, characterToClaim)
		if err != nil {
			return err
		}

		return nil
	}
}
