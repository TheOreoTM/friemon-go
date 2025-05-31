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
		Description: "Get your current character's information",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:         "character",
				Description:  "The ID of the character you want to get info about (optional)",
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
				_ = e.CreateMessage(ErrorMessage("An unexpected error occurred while processing your request."))
			}
			log.Debug("Info command completed",
				logger.Command("info"),
				logger.DiscordUserID(userID),
				logger.Duration(duration),
			)
		}()

		characterParam := e.SlashCommandInteractionData().String("character")
		var ch *entities.Character
		var err error

		if characterParam != "" {
			log.Debug("Character parameter provided",
				logger.DiscordUserID(userID),
				zap.String("character_param", characterParam),
			)
			var characterID uuid.UUID
			characterID, parseErr := uuid.Parse(characterParam)
			if parseErr == nil {
				log.Debug("Getting specific character by ID", logger.DiscordUserID(userID), logger.CharacterID(characterID))
				ch, err = b.DB.GetCharacter(e.Ctx, characterID)
				if err != nil {
					log.Warn("Failed to get specific character by ID", logger.DiscordUserID(userID), logger.CharacterID(characterID), logger.ErrorField(err))
					return e.CreateMessage(ErrorMessage("Character not found with that ID!"))
				}
				if ch.OwnerID != userID.String() {
					log.Warn("User attempted to access character not owned by them", logger.DiscordUserID(userID), logger.CharacterID(characterID), logger.CharacterOwner(ch.OwnerID))
					return e.CreateMessage(ErrorMessage("This character doesn't belong to you!"))
				}
			} else {
				log.Warn("Invalid character ID format provided in parameter", logger.DiscordUserID(userID), zap.String("invalid_param_value", characterParam), logger.ErrorField(parseErr))
				return e.CreateMessage(ErrorMessage(fmt.Sprintf("Invalid character ID: '%s'. Please select from the list or provide a valid ID.", characterParam)))
			}
		} else {
			log.Debug("No character parameter, getting selected character for user", logger.DiscordUserID(userID))
			ch, err = b.DB.GetSelectedCharacter(e.Ctx, userID)
			if err != nil {
				log.Info("User has no selected character", logger.DiscordUserID(userID), logger.ErrorField(err))
				return e.CreateMessage(InfoMessage("You don't have a selected character! Use `/select` to choose one, or specify a character ID with this command."))
			}
		}

		log.Info("Character info retrieved successfully",
			logger.DiscordUserID(userID),
			logger.CharacterID(ch.ID),
			logger.CharacterName(ch.CharacterName()),
			logger.CharacterLevel(ch.Level),
		)

		// --- Build the Embed ---
		embedBuilder := discord.NewEmbedBuilder()
		characterData := ch.Data() // Get base character data

		// Title: Emoji Nickname (Original Name) - Lvl X
		titleName := ch.CharacterName()
		if ch.Nickname != "" {
			titleName = fmt.Sprintf("%s (%s)", ch.Nickname, ch.CharacterName())
		}
		embedBuilder.SetTitle(fmt.Sprintf("%s %s - Lvl %d", characterData.Emoji, titleName, ch.Level))
		embedBuilder.SetColor(int(ch.Color))

		// Thumbnail/Image
		// If you have a smaller "thumbnail" sprite and a larger "image" sprite, you can choose here.
		// For now, let's assume ch.Image() provides the main visual.
		var messageFiles []*discord.File
		if img, imgErr := ch.Image(); imgErr == nil {
			embedBuilder.SetThumbnail("attachment://" + img.Name) // Use SetThumbnail for a side image
			messageFiles = append(messageFiles, img)
		} else {
			log.Warn("Failed to get image for info command thumbnail",
				logger.CharacterID(ch.ID),
				logger.ErrorField(imgErr),
			)
		}

		// Section 1: Details
		embedBuilder.AddField("Details", "──────────────────", false) // Separator
		embedBuilder.AddField("ID", fmt.Sprintf("`%s`", ch.ID.String()), true)
		embedBuilder.AddField("Owner", fmt.Sprintf("<@%s>", ch.OwnerID), true)
		embedBuilder.AddField("Claimed", ch.ClaimedTimestamp.Format("Jan 02, 2006"), true)

		embedBuilder.AddField("Personality", ch.Personality.String(), true)
		shinyStr := "No"
		if ch.Shiny {
			shinyStr = "Yes ✨"
		}
		embedBuilder.AddField("Shiny", shinyStr, true)
		embedBuilder.AddField("IV Total", fmt.Sprintf("**%s**", ch.IvPercentage()), true)

		// Section 2: Stats
		embedBuilder.AddField("Base Stats", "──────────────────", false) // Separator
		embedBuilder.AddField("HP", fmt.Sprintf("%d", ch.HP()), true)
		embedBuilder.AddField("Attack", fmt.Sprintf("%d", ch.Atk()), true)
		embedBuilder.AddField("Defense", fmt.Sprintf("%d", ch.Def()), true)
		embedBuilder.AddField("Sp. Atk", fmt.Sprintf("%d", ch.SpAtk()), true)
		embedBuilder.AddField("Sp. Def", fmt.Sprintf("%d", ch.SpDef()), true)
		embedBuilder.AddField("Speed", fmt.Sprintf("%d", ch.Spd()), true)

		// Section 3: IVs (Individual Values)
		embedBuilder.AddField("Individual Values (IVs)", "──────────────────", false)
		embedBuilder.AddField("IV HP", fmt.Sprintf("%d/31", ch.IvHP), true)
		embedBuilder.AddField("IV Atk", fmt.Sprintf("%d/31", ch.IvAtk), true)
		embedBuilder.AddField("IV Def", fmt.Sprintf("%d/31", ch.IvDef), true)
		embedBuilder.AddField("IV Sp.Atk", fmt.Sprintf("%d/31", ch.IvSpAtk), true)
		embedBuilder.AddField("IV Sp.Def", fmt.Sprintf("%d/31", ch.IvSpDef), true)
		embedBuilder.AddField("IV Spd", fmt.Sprintf("%d/31", ch.IvSpd), true)

		// Footer
		embedBuilder.SetFooterText(fmt.Sprintf("Friemon Bot | Character IDX: %d", ch.IDX))
		embedBuilder.SetTimestamp(time.Now())

		messageCreate := discord.MessageCreate{
			Embeds: []discord.Embed{embedBuilder.Build()},
			Files:  messageFiles,
		}

		return e.CreateMessage(messageCreate)
	}
}

// handleGetCharacterAutocomplete function remains the same
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
