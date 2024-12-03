package handlers

import (
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/entities"
	"github.com/theoreotm/friemon/friemon"
)

const (
	spawnThreshold = 4
)

func SpawnHandler(b *friemon.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.MessageCreate) {
		slog.Info("Interaction count", slog.Int("count", b.Cache.GetInteractionCount(e.ChannelID)))
		b.Cache.IncrementInteractionCount(e.ChannelID)

		if b.Cache.GetInteractionCount(e.ChannelID) <= spawnThreshold {
			return
		}

		randomCharacter := entities.RandomCharacterSpawn()
		b.Cache.SetChannelCharacter(e.ChannelID, randomCharacter)

		spawnEmbed := discord.NewEmbedBuilder().
			SetTitlef("A wandering %v appeared!", randomCharacter.CharacterName()).
			SetDescriptionf("Click the button below to add %v to your characters!", randomCharacter.CharacterName()).
			SetColor(constants.ColorDefault)

		spawnImage, err := randomCharacter.Image()
		if err != nil {
			slog.Error("Failed to get character image", slog.Any("err", err))
		} else {
			spawnEmbed.SetImage("attachment://character.png")
		}

		_, err = e.Client().Rest().CreateMessage(e.ChannelID, discord.MessageCreate{
			Embeds: []discord.Embed{spawnEmbed.Build()},
			Files:  []*discord.File{spawnImage},
		})

		if err != nil {
			slog.Error("Failed to send spawn message",
				slog.String("channel_id", e.ChannelID.String()),
				slog.String("guild_id", e.GuildID.String()),
				slog.Int("character_id", randomCharacter.CharacterID),
				slog.Any("err", err))

			return
		}

		b.Cache.ResetInteractionCount(e.ChannelID)
	})
}
