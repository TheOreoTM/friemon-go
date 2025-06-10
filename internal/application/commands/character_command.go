package commands

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/game"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

func init() {
	Commands["character"] = cmdCharacter
}

var cmdCharacter = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "character",
		Description: "Generate a random character",
	},
	Handler:  handleCharacter,
	Category: "Friemon",
}

func handleCharacter(b *bot.Bot) handler.CommandHandler {
	// Create a logger for this command
	log := logger.NewLogger("commands.character")

	return func(e *handler.CommandEvent) error {
		start := time.Now()
		userID := e.User().ID
		channelID := e.Channel().ID()

		// Log command start with context
		log.Info("Character command started",
			logger.Command("character"),
			logger.DiscordUserID(userID),
			logger.DiscordChannelID(channelID),
		)

		// Add defer for completion logging and panic recovery
		defer func() {
			duration := time.Since(start)
			if r := recover(); r != nil {
				log.Error("Panic in character command",
					logger.Command("character"),
					logger.DiscordUserID(userID),
					logger.Duration(duration),
					zap.Any("panic", r),
				)
				panic(r) // Re-panic after logging
			}

			log.Info("Character command completed",
				logger.Command("character"),
				logger.DiscordUserID(userID),
				logger.Duration(duration),
			)
		}()

		// Generate character with logging
		log.Debug("Generating random character",
			logger.DiscordUserID(userID),
		)

		character := game.RandomCharacterSpawn()

		log.Info("Random character generated",
			logger.CharacterName(character.CharacterName()),
			logger.CharacterLevel(character.Level),
			zap.Bool("shiny", character.Shiny),
			zap.String("personality", character.Personality.String()),
			zap.String("iv_percentage", character.IvPercentage()),
		)

		// Create embed
		embed := discord.NewEmbedBuilder().
			SetTitle(fmt.Sprintf("Random %s", character.CharacterName())).
			SetDescription(fmt.Sprintf("Level: %d\nPersonality: %s\nIV: %s",
				character.Level,
				character.Personality.String(),
				character.IvPercentage(),
			)).
			SetColor(int(character.Color)).
			Build()

		err := b.DB.CreateCharacter(e.Ctx, userID, character)
		if err != nil {
			log.Error("Failed to create character",
				logger.DiscordUserID(userID),
				logger.CharacterName(character.CharacterName()),
				logger.ErrorField(err),
			)
			return e.CreateMessage(ErrorMessage("Failed to create character!"))
		}

		log.Info("Character created successfully",
			logger.DiscordUserID(userID),
			logger.CharacterID(character.ID),
			logger.CharacterName(character.CharacterName()),
		)

		// Handle image with error logging
		var response discord.MessageCreate
		if img, err := character.Image(); err == nil {
			embed.Image = &discord.EmbedResource{URL: "attachment://" + img.Name}
			response = discord.MessageCreate{
				Embeds: []discord.Embed{embed},
				Files:  []*discord.File{img},
			}

			log.Debug("Character image attached",
				logger.CharacterName(character.CharacterName()),
				zap.String("image_file", img.Name),
			)
		} else {
			log.Warn("Failed to get character image",
				logger.CharacterName(character.CharacterName()),
				logger.ErrorField(err),
			)

			response = discord.MessageCreate{
				Embeds: []discord.Embed{embed},
			}
		}

		// Send response with error handling
		if err := e.CreateMessage(response); err != nil {
			log.Error("Failed to send character response",
				logger.Command("character"),
				logger.DiscordUserID(userID),
				logger.ErrorField(err),
			)
			return err
		}

		log.Info("Character command response sent successfully",
			logger.DiscordUserID(userID),
			logger.CharacterName(character.CharacterName()),
		)

		return nil
	}
}
