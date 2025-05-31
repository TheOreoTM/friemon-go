package components

import (
	"fmt"
	"sync"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

var claimMutex sync.Mutex

func init() {
	Components["claim_character"] = claimCharacterButton
}

func claimCharacterButton(b *bot.Bot) handler.ComponentHandler {
	log := logger.NewLogger("components.claim")

	return func(e *handler.ComponentEvent) error {
		start := time.Now()
		userID := e.User().ID
		channelID := e.Channel().ID()
		messageID := e.Message.ID

		log.Info("Claim button pressed",
			logger.Component("claim_character"),
			logger.DiscordUserID(userID),
			logger.DiscordChannelID(channelID),
			logger.DiscordMessageID(messageID),
		)

		defer func() {
			duration := time.Since(start)
			if r := recover(); r != nil {
				log.Error("Panic in claim button handler",
					logger.Component("claim_character"),
					logger.DiscordUserID(userID),
					logger.Duration(duration),
					zap.Any("panic", r),
				)
				panic(r)
			}

			log.Debug("Claim button handler completed",
				logger.Component("claim_character"),
				logger.DiscordUserID(userID),
				logger.Duration(duration),
			)
		}()

		// Acquire mutex to prevent race conditions
		log.Debug("Acquiring claim mutex",
			logger.DiscordUserID(userID),
			logger.DiscordChannelID(channelID),
		)

		claimMutex.Lock()
		defer claimMutex.Unlock()

		log.Debug("Claim mutex acquired",
			logger.DiscordUserID(userID),
		)

		// Get cached character
		character, err := b.Cache.GetChannelCharacter(channelID)
		if err != nil || character == nil {
			log.Warn("No character found in channel cache",
				logger.DiscordUserID(userID),
				logger.DiscordChannelID(channelID),
				logger.ErrorField(err),
			)

			return e.CreateMessage(discord.MessageCreate{
				Content: "❌ No character available to claim!",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		log.Info("Character found for claiming",
			logger.DiscordUserID(userID),
			logger.CharacterID(character.ID),
			logger.CharacterName(character.CharacterName()),
			logger.CharacterLevel(character.Level),
			zap.Bool("shiny", character.Shiny),
		)

		// Check if character is already claimed
		if character.OwnerID != "" {
			log.Warn("Character already claimed",
				logger.DiscordUserID(userID),
				logger.CharacterID(character.ID),
				logger.CharacterOwner(character.OwnerID),
			)

			return e.CreateMessage(discord.MessageCreate{
				Content: "❌ This character has already been claimed!",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		// Get or create user
		user, err := b.DB.GetUser(e.Ctx, userID)
		if err != nil {
			log.Info("Creating new user",
				logger.DiscordUserID(userID),
			)

			user, err = b.DB.CreateUser(e.Ctx, userID)
			if err != nil {
				log.Error("Failed to create user",
					logger.DiscordUserID(userID),
					logger.ErrorField(err),
				)

				return e.CreateMessage(discord.MessageCreate{
					Content: "❌ Failed to create user profile!",
					Flags:   discord.MessageFlagEphemeral,
				})
			}

			log.Info("New user created",
				logger.DiscordUserID(userID),
			)
		}

		// Set character ownership
		character.OwnerID = userID.String()
		character.IDX = user.NextIdx
		character.ClaimedTimestamp = time.Now()

		log.Debug("Assigning character to user",
			logger.DiscordUserID(userID),
			logger.CharacterID(character.ID),
			zap.Int("character_index", character.IDX),
		)

		// Save character to database
		if err := b.DB.CreateCharacter(e.Ctx, userID, character); err != nil {
			log.Error("Failed to save character to database",
				logger.DiscordUserID(userID),
				logger.CharacterID(character.ID),
				logger.ErrorField(err),
			)

			return e.CreateMessage(discord.MessageCreate{
				Content: "❌ Failed to save character!",
				Flags:   discord.MessageFlagEphemeral,
			})
		}

		log.Info("Character saved to database",
			logger.DiscordUserID(userID),
			logger.CharacterID(character.ID),
			logger.CharacterName(character.CharacterName()),
		)

		// Update user's next index
		user.NextIdx++
		if _, err := b.DB.UpdateUser(e.Ctx, *user); err != nil {
			log.Error("Failed to update user next index",
				logger.DiscordUserID(userID),
				logger.ErrorField(err),
			)
			// Don't return error here as character was already saved
		}

		// Clean up cache
		if err := b.Cache.DeleteChannelCharacter(channelID); err != nil {
			log.Warn("Failed to clean up character cache",
				logger.DiscordChannelID(channelID),
				logger.ErrorField(err),
			)
		}

		log.Debug("Character cache cleaned up",
			logger.DiscordChannelID(channelID),
		)

		// Send success response
		embed := discord.NewEmbedBuilder().
			SetTitle("🎉 Character Claimed!").
			SetDescription(fmt.Sprintf("You claimed **%s** (Level %d)!\nIV: %s | %s",
				character.CharacterName(),
				character.Level,
				character.IvPercentage(),
				character.Personality.String(),
			)).
			SetColor(int(character.Color)).
			Build()

		if err := e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{embed},
			Flags:  discord.MessageFlagEphemeral,
		}); err != nil {
			log.Error("Failed to send claim success message",
				logger.DiscordUserID(userID),
				logger.ErrorField(err),
			)
			return err
		}

		log.Info("Character claimed successfully",
			logger.DiscordUserID(userID),
			logger.CharacterID(character.ID),
			logger.CharacterName(character.CharacterName()),
			zap.Int("user_character_count", user.NextIdx),
		)

		// Update original message to show it's claimed
		// ... (rest of the logic to disable the button)

		return nil
	}
}
