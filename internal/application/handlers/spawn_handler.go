package handlers

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/entities"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

const (
	spawnThreshold = 4
)

func spawnCharacter(b *bot.Bot, e *events.MessageCreate) {
	log := logger.NewLogger("handlers.spawn")

	start := time.Now()
	channelID := e.ChannelID
	guildID := e.GuildID

	log.Info("Spawn handler triggered",
		logger.Handler("spawn"),
		logger.DiscordChannelID(channelID),
		logger.DiscordGuildID(*guildID),
	)

	defer func() {
		log.Info("Spawn handler completed",
			logger.Handler("spawn"),
			logger.DiscordChannelID(channelID),
			logger.Duration(time.Since(start)),
		)
	}()

	// Get and log current interaction count
	count := b.Cache.GetInteractionCount(channelID)
	log.Info("Checking interaction count",
		logger.DiscordChannelID(channelID),
		zap.Int("current_count", count),
		zap.Int("required_threshold", spawnThreshold),
		logger.CacheHit(count > 0),
	)

	if count < spawnThreshold {
		log.Info("Threshold not met, skipping spawn",
			logger.DiscordChannelID(channelID),
			zap.Int("count", count),
			zap.Int("needed", spawnThreshold-count),
		)
		return
	}

	// Check for existing character
	if existingChar, err := b.Cache.GetChannelCharacter(channelID); err == nil && existingChar != nil {
		log.Info("Character already exists in channel",
			logger.DiscordChannelID(channelID),
			logger.CharacterID(existingChar.ID),
			logger.CharacterName(existingChar.CharacterName()),
		)
		return
	}

	// Generate character
	character := entities.RandomCharacterSpawn()

	log.Info("Spawning character",
		logger.Handler("spawn"),
		logger.DiscordChannelID(channelID),
		logger.CharacterName(character.CharacterName()),
		logger.CharacterLevel(character.Level),
		zap.Bool("shiny", character.Shiny),
		zap.String("personality", character.Personality.String()),
		zap.String("iv_percentage", character.IvPercentage()),
	)

	// Cache character
	if err := b.Cache.SetChannelCharacter(channelID, character); err != nil {
		log.Error("Failed to cache spawned character",
			logger.DiscordChannelID(channelID),
			logger.CharacterID(character.ID),
			logger.ErrorField(err),
		)
		return
	}

	log.Info("Character cached successfully",
		logger.DiscordChannelID(channelID),
		logger.CharacterID(character.ID),
		logger.CacheKey(fmt.Sprintf("channel:%s:character", channelID)),
	)

	// Reset interaction count
	if err := b.Cache.ResetInteractionCount(channelID); err != nil {
		log.Warn("Failed to reset interaction count",
			logger.DiscordChannelID(channelID),
			logger.ErrorField(err),
		)
	} else {
		log.Debug("Interaction count reset",
			logger.DiscordChannelID(channelID),
		)
	}

	// Build spawn message
	embed := discord.NewEmbedBuilder().
		SetTitle(fmt.Sprintf("A wild %s appeared!", character.CharacterName())).
		SetDescription("Click the button below to claim it!").
		SetColor(constants.ColorInfo).
		Build()

	components := []discord.ContainerComponent{
		discord.NewActionRow(
			discord.NewPrimaryButton("Claim", "claim_character"),
		),
	}

	// Handle image
	var files []*discord.File
	if img, err := character.Image(); err == nil {
		embed.Image = &discord.EmbedResource{URL: "attachment://" + img.Name}
		files = []*discord.File{img}

		log.Info("Character image loaded",
			logger.CharacterName(character.CharacterName()),
			zap.String("image_file", img.Name),
		)
	} else {
		log.Warn("Failed to load character image",
			logger.CharacterName(character.CharacterName()),
			logger.ErrorField(err),
		)
	}

	// Send spawn message
	msg, err := b.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
		Embeds:     []discord.Embed{embed},
		Components: components,
		Files:      files,
	})

	if err != nil {
		log.Error("Failed to send spawn message",
			logger.DiscordChannelID(channelID),
			logger.CharacterID(character.ID),
			logger.ErrorField(err),
		)
		return
	}

	log.Info("Character spawned successfully",
		logger.Handler("spawn"),
		logger.DiscordChannelID(channelID),
		logger.DiscordMessageID(msg.ID),
		logger.CharacterName(character.CharacterName()),
		zap.String("message_url", func() string {
			if guildID == nil {
				return ""
			}
			return fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guildID, channelID, msg.ID)
		}()),
	)

	// Schedule cleanup
	if _, err := b.Scheduler.After(3*time.Minute).
		Type("cleanup_character").
		With("channel_id", channelID.String()).
		With("message_id", msg.ID.String()).
		Emit("cleanup_character"); err != nil {

		log.Error("Failed to schedule cleanup task",
			logger.DiscordChannelID(channelID),
			logger.DiscordMessageID(msg.ID),
			logger.ErrorField(err),
		)
	} else {
		log.Info("Cleanup task scheduled",
			logger.DiscordChannelID(channelID),
			logger.DiscordMessageID(msg.ID),
			zap.Duration("cleanup_in", 3*time.Minute),
		)
	}
}
