package handlers

import (
	"log/slog"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/entities"
	"github.com/theoreotm/friemon/friemon"
)

const (
	spawnThreshold = 4 // Change accordingly
)

func spawnCharacter(b *friemon.Bot, e *events.MessageCreate) {
	slog.Info("Interaction count", slog.Int("count", b.Cache.GetInteractionCount(e.ChannelID)))
	err := b.Cache.IncrementInteractionCount(e.ChannelID)
	if err != nil {
		slog.Error("Failed to increment interaction count", slog.Any("err", err))
		return
	}

	if b.Cache.GetInteractionCount(e.ChannelID) <= spawnThreshold {
		return
	}

	randomCharacter := entities.RandomCharacterSpawn()

	spawnEmbed := discord.NewEmbedBuilder().
		SetTitlef("A wandering %v appeared!", randomCharacter.CharacterName()).
		SetDescriptionf("Click the button below to add %v to your characters!", randomCharacter.CharacterName()).
		SetColor(constants.ColorDefault)

	spawnImage, err := randomCharacter.Image()
	if err != nil {
		slog.Error("Failed to get character image", slog.Any("err", err))
		err := b.Cache.ResetInteractionCount(e.ChannelID)
		if err != nil {
			slog.Error("Failed to reset interaction count", slog.Any("err", err))
			return
		}
		return
	} else {
		spawnEmbed.SetImage("attachment://character.png")
	}

	_, err = e.Client().Rest().CreateMessage(e.ChannelID,
		discord.NewMessageCreateBuilder().
			AddEmbeds(spawnEmbed.Build()).
			AddFiles(spawnImage).
			AddActionRow(discord.NewPrimaryButton("Invite to party", "/claim")).
			Build(),
	)

	if err != nil {
		slog.Error("Failed to send spawn message",
			slog.String("channel_id", e.ChannelID.String()),
			slog.String("guild_id", e.GuildID.String()),
			slog.Int("character_id", randomCharacter.CharacterID),
			slog.Any("err", err))

		return
	}

	err = b.Cache.SetChannelCharacter(e.ChannelID, randomCharacter)
	if err != nil {
		slog.Error("Failed to set channel character", slog.Any("err", err))
		return
	}

	err = b.Cache.ResetInteractionCount(e.ChannelID)
	if err != nil {
		slog.Error("Failed to reset interaction count", slog.Any("err", err))
		return
	}
}
