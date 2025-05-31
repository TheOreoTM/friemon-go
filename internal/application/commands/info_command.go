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
				_ = e.CreateMessage(ErrorMessage("An unexpected error occurred."))
			}
			log.Debug("Info command completed",
				logger.Command("info"),
				logger.DiscordUserID(userID),
				logger.Duration(duration),
			)
		}()

		// Directly get the "character" option string.
		// If the option is not provided, this will return an empty string.
		characterParam := e.SlashCommandInteractionData().String("character")

		if characterParam != "" {
			log.Debug("Character parameter provided",
				logger.DiscordUserID(userID),
				zap.String("character_param", characterParam),
			)
		}

		var ch *entities.Character
		var err error

		if characterParam != "" {
			// Attempt to parse the parameter as a UUID (character ID)
			var characterID uuid.UUID
			characterID, parseErr := uuid.Parse(characterParam)
			if parseErr == nil {
				// Parameter is a valid UUID, try to fetch this specific character
				log.Debug("Getting specific character by ID",
					logger.DiscordUserID(userID),
					logger.CharacterID(characterID),
				)

				ch, err = b.DB.GetCharacter(e.Ctx, characterID)
				if err != nil {
					log.Warn("Failed to get specific character by ID",
						logger.DiscordUserID(userID),
						logger.CharacterID(characterID),
						logger.ErrorField(err),
					)
					return e.CreateMessage(ErrorMessage("Character not found with that ID!"))
				}

				// Important: Check if the fetched character belongs to the user
				if ch.OwnerID != userID.String() {
					log.Warn("User attempted to access character not owned by them",
						logger.DiscordUserID(userID),
						logger.CharacterID(characterID),
						logger.CharacterOwner(ch.OwnerID),
					)
					return e.CreateMessage(ErrorMessage("This character doesn't belong to you!"))
				}
			} else {
				// Parameter was not a valid UUID.
				// This can happen if autocomplete fails or user types something random.
				log.Warn("Invalid character ID format provided in parameter",
					logger.DiscordUserID(userID),
					zap.String("invalid_param_value", characterParam),
					logger.ErrorField(parseErr),
				)
				return e.CreateMessage(ErrorMessage(fmt.Sprintf("Invalid character ID: '%s'. Please select from the list or provide a valid ID.", characterParam)))
			}
		} else {
			// No character parameter provided, get the user's selected character
			log.Debug("No character parameter, getting selected character for user",
				logger.DiscordUserID(userID),
			)

			ch, err = b.DB.GetSelectedCharacter(e.Ctx, userID)
			if err != nil {
				log.Info("User has no selected character",
					logger.DiscordUserID(userID),
					logger.ErrorField(err),
				)
				return e.CreateMessage(InfoMessage("You don't have a selected character! Use `/select` to choose one, or specify a character ID with this command."))
			}
		}

		// At this point, 'ch' should hold the character to display info for.
		log.Info("Character info retrieved successfully",
			logger.DiscordUserID(userID),
			logger.CharacterID(ch.ID),
			logger.CharacterName(ch.CharacterName()),
			logger.CharacterLevel(ch.Level),
		)

		// --- Build embed and send response ---
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

		embedBuilder := discord.NewEmbedBuilder().
			SetTitle(fmt.Sprintf("%s %s (Lvl %d)", ch.Data().Emoji, ch.CharacterName(), ch.Level)).
			SetColor(int(ch.Color))

		for _, detail := range detailFieldValues {
			embedBuilder.AddField(detail[0], detail[1], true)
		}
		// if len(detailFieldValues)%2 != 0 { // Ensure alignment for stats
		// 	embedBuilder.AddField("0", "0", true)
		// }
		for _, stat := range statFieldValues {
			embedBuilder.AddField(stat[0], stat[1], true)
		}

		messageCreate := discord.MessageCreate{}

		if img, imgErr := ch.Image(); imgErr == nil {
			embedBuilder.SetImage("attachment://" + img.Name)
			messageCreate.Files = []*discord.File{img}
		} else {
			log.Warn("Failed to get image for info command",
				logger.CharacterID(ch.ID),
				logger.ErrorField(imgErr),
			)
		}
		messageCreate.Embeds = []discord.Embed{embedBuilder.Build()}

		return e.CreateMessage(messageCreate)
		// --- End of embed building ---
	}
}

// handleGetCharacterAutocomplete remains the same as your previous version
func handleGetCharacterAutocomplete(b *bot.Bot) handler.AutocompleteHandler {
	log := logger.NewLogger("autocomplete.character")

	return func(e *handler.AutocompleteEvent) error {
		start := time.Now()
		userID := e.User().ID
		query := e.Data.String("character")

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
					Value: ch.ID.String(),
				})
			}
		}

		var choices []discord.AutocompleteChoice
		for i, r := range results {
			if i >= 25 {
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
