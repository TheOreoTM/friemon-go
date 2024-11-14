package handlers

import (
	"database/sql"
	"fmt"
	"log/slog"
	"math/rand"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/friemon"
)

func MessageHandler(b *friemon.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageCreate) {
		if e.Message.Author.Bot {
			return
		}

		leveledUp := false
		selectedCharacter, err := b.DB.GetSelectedCharacter(b.Context, e.Message.Author.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return
			}

			slog.Error("Failed to get selected character", slog.Any("err", err))
			return
		}

		if selectedCharacter == nil {
			return
		}

		if selectedCharacter.Level == 100 {
			return
		}

		selectedCharacter.XP += randomInt(10, 40)
		fmt.Println(selectedCharacter.XP)
		if selectedCharacter.XP > selectedCharacter.MaxXP() {
			selectedCharacter.XP = 0

			if selectedCharacter.Level+1 >= 100 {
				selectedCharacter.Level = 100
			} else {
				selectedCharacter.Level++
				leveledUp = true
			}

		}

		char, err := b.DB.UpdateCharacter(b.Context, selectedCharacter.ID, selectedCharacter)
		if err != nil {
			return
		}

		if leveledUp {
			embedB := discord.NewEmbedBuilder().
				SetTitlef("Congratulations %v!", e.Message.Author.EffectiveName()).
				SetDescriptionf("Your %v is now level %v!", char.Format("n"), char.Level).
				SetColor(constants.ColorDefault)

			image, err := char.Image()
			if err == nil {
				embedB.SetThumbnail("attachment://character.png")
			}

			e.Client().Rest().CreateMessage(e.ChannelID, discord.MessageCreate{
				Embeds: []discord.Embed{embedB.Build()},
				Files:  []*discord.File{image},
			})
		}
	})
}

func randomInt(min, max int) int {
	return rand.Intn(max-min) + min
}
