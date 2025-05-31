package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/entities"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

func init() {
	Commands[cmdInfo.Cmd.CommandName()] = cmdInfo
}

var cmdInfo = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "info",
		Description: "Get your current character",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:         "character",
				Description:  "The character you want to get info about",
				Required:     false,
				Autocomplete: true,
			},
		},
	},
	Autocomplete: handleGetCharacterAutocomplete,
	Handler:      handleInfo,
	Category:     "Friemon",
}

func handleInfo(b *bot.Bot) handler.CommandHandler {
	log := logger.NewLogger("commands.info")

	return func(e *handler.CommandEvent) error {
		start := time.Now()
		userID := e.User().ID

		log.Info("Info command started",
			logger.Command("info"),
			logger.DiscordUserID(userID),
		)

		defer func() {
			duration := time.Since(start)
			if r := recover(); r != nil {
				log.Error("Panic in info command",
					logger.Command("info"),
					logger.DiscordUserID(userID),
					logger.Duration(duration),
					zap.Any("panic", r),
				)
				panic(r) // Re-panic after logging
			}
			log.Debug("Info command completed", // Changed to Debug for completion
				logger.Command("info"),
				logger.DiscordUserID(userID),
				logger.Duration(duration),
			)
		}()

		// Get character parameter if provided
		characterParam := ""
		data := e.SlashCommandInteractionData()

		// Correct way to access an option from the map
		if opt, ok := data.Options["character"]; ok {
			characterParam = string(opt.Value)
			log.Debug("Character parameter provided",
				logger.DiscordUserID(userID),
				zap.String("character_param", characterParam),
			)
		}

		var ch *entities.Character
		var err error

		if characterParam != "" {
			// Parse UUID and get specific character
			var characterID uuid.UUID
			characterID, parseErr := uuid.Parse(characterParam)
			if parseErr == nil {
				log.Debug("Getting specific character by ID",
					logger.DiscordUserID(userID),
					logger.CharacterID(characterID),
				)

				ch, err = b.DB.GetCharacter(e.Ctx, characterID)
				if err != nil {
					log.Warn("Failed to get specific character by ID",
						logger.DiscordUserID(userID),
						logger.CharacterID(characterID),
						logger.ErrorField(err), // Using ErrorField as you mentioned
					)
					// Check if the character belongs to the user or if it's a general not found
					// For simplicity, keeping the original error message
					return e.CreateMessage(ErrorMessage("Character not found or doesn't belong to you!"))
				}
				// Ensure the character belongs to the user if fetched by ID
				if ch.OwnerID != userID.String() {
					log.Warn("User attempted to access character not owned by them",
						logger.DiscordUserID(userID),
						logger.CharacterID(characterID),
						logger.CharacterOwner(ch.OwnerID),
					)
					return e.CreateMessage(ErrorMessage("This character doesn't belong to you!"))
				}

			} else {
				log.Warn("Invalid character UUID provided",
					logger.DiscordUserID(userID),
					zap.String("invalid_uuid", characterParam),
					logger.ErrorField(parseErr), // Using ErrorField
				)
				return e.CreateMessage(ErrorMessage("Invalid character ID format! Please select from the list or provide a valid ID."))
			}
		} else {
			// Get user's selected character
			log.Debug("Getting selected character for user",
				logger.DiscordUserID(userID),
			)

			ch, err = b.DB.GetSelectedCharacter(e.Ctx, userID)
			if err != nil {
				log.Info("User has no selected character", // This is an expected scenario, so Info level
					logger.DiscordUserID(userID),
					logger.ErrorField(err), // Using ErrorField
				)
				return e.CreateMessage(InfoMessage("You don't have a selected character! Use `/select` to choose one."))
			}
		}

		log.Info("Character info retrieved successfully",
			logger.DiscordUserID(userID),
			logger.CharacterID(ch.ID),
			logger.CharacterName(ch.CharacterName()),
			logger.CharacterLevel(ch.Level),
		)

		// --- Build embed and send response ---
		// (This part of your code would go here)
		// Example:
		var detailFieldValues = [][]string{
			{"ID", ch.ID.String()},
			{"Owner", fmt.Sprintf("<@%s>", ch.OwnerID)},
			{"Claimed", ch.ClaimedTimestamp.Format("Jan 02, 2006")},
			{"Personality", ch.Personality.String()},
			{"Shiny", fmt.Sprintf("%t", ch.Shiny)},
			{"IV Total", ch.IvPercentage()},
		}

		var statFieldValues = [][]string{
			{"HP", fmt.Sprintf("%d", ch.HP())},
			{"Attack", fmt.Sprintf("%d", ch.Atk())},
			{"Defense", fmt.Sprintf("%d", ch.Def())},
			{"Sp. Atk", fmt.Sprintf("%d", ch.SpAtk())},
			{"Sp. Def", fmt.Sprintf("%d", ch.SpDef())},
			{"Speed", fmt.Sprintf("%d", ch.Spd())},
		}

		embed := discord.NewEmbedBuilder().
			SetTitle(fmt.Sprintf("%s %s (Lvl %d)", ch.Data().Emoji, ch.CharacterName(), ch.Level)).
			SetColor(int(ch.Color))

		for _, detail := range detailFieldValues {
			embed.AddField(detail[0], detail[1], true)
		}
		if len(detailFieldValues)%2 != 0 {
			embed.AddField("", "", true) // Add empty field for spacing
		}
		for _, stat := range statFieldValues {
			embed.AddField(stat[0], stat[1], true)
		}

		messageCreate := discord.MessageCreate{Embeds: []discord.Embed{embed.Build()}}

		if img, imgErr := ch.Image(); imgErr == nil {
			embed.SetImage("attachment://" + img.Name) // Update image URL in embed
			messageCreate.Files = []*discord.File{img}
			messageCreate.Embeds = []discord.Embed{embed.Build()} // Rebuild embed with image
		} else {
			log.Warn("Failed to get image for info command",
				logger.CharacterID(ch.ID),
				logger.ErrorField(imgErr),
			)
		}

		return e.CreateMessage(messageCreate)
		// --- End of embed building ---
	}
}

func handleGetCharacterAutocomplete(b *bot.Bot) handler.AutocompleteHandler {
	log := logger.NewLogger("autocomplete.character") // Logger for autocomplete

	return func(e *handler.AutocompleteEvent) error {
		start := time.Now()
		userID := e.User().ID
		query := e.Data.String("character") // Assuming 'character' is the option name

		log.Debug("Autocomplete request received",
			logger.DiscordUserID(userID),
			zap.String("query", query),
		)

		var results []discord.AutocompleteChoiceString
		chars, err := b.DB.GetCharactersForUser(e.Ctx, userID)
		if err != nil {
			log.Error("Failed to get characters for autocomplete",
				logger.DiscordUserID(userID),
				logger.ErrorField(err),
			)
			// It's better to return an empty result or a generic error choice
			// than to show "You dont have any characters" if there's a DB error.
			return e.AutocompleteResult([]discord.AutocompleteChoice{
				discord.AutocompleteChoiceString{Name: "Error fetching characters", Value: "error"},
			})
		}

		if len(chars) == 0 {
			log.Debug("User has no characters for autocomplete", logger.DiscordUserID(userID))
			return e.AutocompleteResult([]discord.AutocompleteChoice{
				discord.AutocompleteChoiceString{Name: "You don't have any characters", Value: "no_characters"},
			})
		}

		for _, ch := range chars {
			// Improved matching: check nickname, character name, and IDX
			displayName := ch.CharacterName()
			if ch.Nickname != "" {
				displayName = ch.Nickname
			}

			searchText := strings.ToLower(fmt.Sprintf("%s %s %d", displayName, ch.CharacterName(), ch.IDX))
			queryLower := strings.ToLower(query)

			if query == "" || strings.Contains(searchText, queryLower) {
				choiceName := fmt.Sprintf("%d: %s (Lvl %d)", ch.IDX, displayName, ch.Level)
				if ch.Shiny {
					choiceName += " ✨"
				}
				results = append(results, discord.AutocompleteChoiceString{
					Name:  choiceName,
					Value: ch.ID.String(), // Value should be the ID for selection
				})
			}
		}

		var choices []discord.AutocompleteChoice
		for i, r := range results {
			if i >= 25 { // Discord limit for autocomplete choices
				break
			}
			choices = append(choices, r)
		}

		if len(choices) == 0 && query != "" {
			log.Debug("No matching characters found for autocomplete query",
				logger.DiscordUserID(userID),
				zap.String("query", query),
			)
			choices = append(choices, discord.AutocompleteChoiceString{
				Name:  fmt.Sprintf("No characters found matching '%s'", query),
				Value: "no_match",
			})
		}

		log.Debug("Autocomplete results prepared",
			logger.DiscordUserID(userID),
			zap.Int("result_count", len(choices)),
			logger.Duration(time.Since(start)),
		)
		return e.AutocompleteResult(choices)
	}
}
