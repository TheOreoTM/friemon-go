package handlers

import (
	"database/sql"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

func incrementXp(b *bot.Bot, e *events.MessageCreate) {
	log := logger.NewLogger("handlers.xp")

	userID := e.Message.Author.ID
	channelID := e.ChannelID

	log.Debug("XP increment triggered",
		logger.Handler("xp"),
		logger.DiscordUserID(userID),
		logger.DiscordChannelID(channelID),
	)

	// Get user's selected character
	character, err := b.DB.GetSelectedCharacter(b.Context, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("User has no selected character for XP",
				logger.DiscordUserID(userID),
			)
		} else {
			log.Error("Failed to get selected character for XP",
				logger.DiscordUserID(userID),
				logger.ErrorField(err),
			)
		}
		return
	}

	oldLevel := character.Level
	oldXP := character.XP

	log.Debug("Current character stats",
		logger.DiscordUserID(userID),
		logger.CharacterID(character.ID),
		logger.CharacterName(character.CharacterName()),
		logger.CharacterLevel(oldLevel),
		zap.Int("current_xp", oldXP),
		zap.Int("max_xp", character.MaxXP()),
	)

	// Increment XP (you can adjust this logic)
	xpGain := 1 // Base XP gain
	character.XP += xpGain

	// Check for level up
	leveledUp := false
	newLevel := character.Level

	for character.XP >= character.MaxXP() && character.Level < 100 {
		character.XP -= character.MaxXP()
		character.Level++
		leveledUp = true
		newLevel = character.Level

		log.Info("Character leveled up!",
			logger.DiscordUserID(userID),
			logger.CharacterID(character.ID),
			logger.CharacterName(character.CharacterName()),
			zap.Int("old_level", oldLevel),
			zap.Int("new_level", newLevel),
		)
	}

	// Update character in database
	updatedChar, err := b.DB.UpdateCharacter(b.Context, character.ID, character)
	if err != nil {
		log.Error("Failed to update character XP",
			logger.DiscordUserID(userID),
			logger.CharacterID(character.ID),
			logger.ErrorField(err),
		)
		return
	}

	log.Debug("Character XP updated",
		logger.DiscordUserID(userID),
		logger.CharacterID(character.ID),
		zap.Int("xp_gained", xpGain),
		zap.Int("new_xp", updatedChar.XP),
		zap.Int("new_level", updatedChar.Level),
	)

	// Send level up notification if applicable
	if leveledUp {
		embed := discord.NewEmbedBuilder().
			SetTitle("🎉 Level Up!").
			SetDescription(fmt.Sprintf("Your %s reached level %d!",
				character.CharacterName(), newLevel)).
			SetColor(constants.ColorSuccess).
			Build()

		if _, err := b.Client.Rest().CreateMessage(channelID, discord.MessageCreate{
			Embeds: []discord.Embed{embed},
		}); err != nil {
			log.Error("Failed to send level up notification",
				logger.DiscordUserID(userID),
				logger.DiscordChannelID(channelID),
				logger.CharacterID(character.ID),
				logger.ErrorField(err),
			)
		} else {
			log.Info("Level up notification sent",
				logger.DiscordUserID(userID),
				logger.DiscordChannelID(channelID),
				logger.CharacterName(character.CharacterName()),
				zap.Int("new_level", newLevel),
			)
		}
	}
}
