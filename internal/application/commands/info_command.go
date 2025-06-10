package commands

import (
	"database/sql" // Import for sql.ErrNoRows
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/game"
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
				Description:  "The ID of the character to get info about (optional)",
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
	log := logger.NewLogger("commands.info") // Create a logger for this command

	return func(e *handler.CommandEvent) error {
		start := time.Now()
		userID := e.User().ID
		commandName := e.SlashCommandInteractionData().CommandName()

		log.Info("Info command started",
			logger.Command(commandName),
			logger.DiscordUserID(userID),
		)

		defer func() {
			duration := time.Since(start)
			if r := recover(); r != nil {
				log.Error("Panic in info command",
					logger.Command(commandName),
					logger.DiscordUserID(userID),
					logger.Duration(duration),
					zap.Any("panic", r),
				)
				// Attempt to send an ephemeral error message to the user
				_ = e.CreateMessage(ErrorMessage("An unexpected error occurred. Please try again later."))
			}
			log.Debug("Info command completed", // Changed to Debug for completion
				logger.Command(commandName),
				logger.DiscordUserID(userID),
				logger.Duration(duration),
			)
		}()

		characterIDParam := e.SlashCommandInteractionData().String("character")
		var ch *game.Character
		var err error

		if characterIDParam != "" && characterIDParam != "-1" { // -1 might be from autocomplete placeholder
			log.Debug("Character ID parameter provided",
				logger.DiscordUserID(userID),
				zap.String("character_id_param", characterIDParam),
			)
			parsedID, parseErr := uuid.Parse(characterIDParam)
			if parseErr != nil {
				log.Warn("Invalid UUID format provided for character ID",
					logger.DiscordUserID(userID),
					zap.String("invalid_uuid", characterIDParam),
					logger.ErrorField(parseErr),
				)
				return e.CreateMessage(ErrorMessage(fmt.Sprintf("The provided character ID '%s' is not valid. Please select from the list.", characterIDParam)))
			}

			log.Debug("Fetching character by ID from database",
				logger.DiscordUserID(userID),
				logger.CharacterID(parsedID),
			)
			ch, err = b.DB.GetCharacter(e.Ctx, parsedID)
			if err != nil {
				log.Warn("Failed to get character by ID",
					logger.DiscordUserID(userID),
					logger.CharacterID(parsedID),
					logger.ErrorField(err),
				)
				// Don't assign to ch if error, ch will remain nil
				// Let it fall through to selected character logic or error if no selected.
			} else if ch != nil && ch.OwnerID != userID.String() {
				log.Warn("User attempted to view info for a character they don't own",
					logger.DiscordUserID(userID),
					logger.CharacterID(ch.ID),
					logger.CharacterOwner(ch.OwnerID),
				)
				return e.CreateMessage(ErrorMessage("You can only view information for characters you own."))
			} else if ch != nil {
				log.Info("Successfully fetched character by ID",
					logger.DiscordUserID(userID),
					logger.CharacterID(ch.ID),
					logger.CharacterName(ch.CharacterName()),
				)
			}
		}

		// If ch is still nil (no valid ID provided, or DB error for that ID)
		// try to get the user's selected character.
		if ch == nil {
			log.Debug("No specific character ID provided or fetch failed, attempting to get selected character",
				logger.DiscordUserID(userID),
			)
			selectedCh, selectedErr := b.DB.GetSelectedCharacter(e.Ctx, userID)
			if selectedErr != nil {
				if selectedErr == sql.ErrNoRows {
					log.Info("User has no selected character and no valid specific character ID was provided",
						logger.DiscordUserID(userID),
					)
					return e.CreateMessage(InfoMessage("You don't have a character selected, and no specific character was found. Use `/select` or provide a valid character ID."))
				}
				log.Error("Error fetching selected character",
					logger.DiscordUserID(userID),
					logger.ErrorField(selectedErr),
				)
				return e.CreateMessage(ErrorMessage(fmt.Sprintf("Error fetching your selected character: %s", selectedErr.Error())))
			}
			ch = selectedCh
			log.Info("Successfully fetched user's selected character",
				logger.DiscordUserID(userID),
				logger.CharacterID(ch.ID),
				logger.CharacterName(ch.CharacterName()),
			)
		}

		// At this point, ch should be non-nil if a character was found.
		if ch == nil {
			// This case should ideally be caught by the logic above, but as a safeguard:
			log.Error("Character is unexpectedly nil after all fetch attempts", logger.DiscordUserID(userID))
			return e.CreateMessage(ErrorMessage("Could not determine which character to display information for."))
		}

		// --- Your Embed Building Logic ---
		var detailFieldValues = [][]string{}
		detailFieldValues = append(detailFieldValues, []string{"XP", fmt.Sprintf("%d/%d", ch.XP, ch.MaxXP())})
		detailFieldValues = append(detailFieldValues, []string{"Personality", ch.Personality.String()})
		// Add more details if needed, e.g., Shiny, Claimed Date
		shinyStr := "No"
		if ch.Shiny {
			shinyStr = "Yes ✨"
		}
		detailFieldValues = append(detailFieldValues, []string{"Shiny", shinyStr})
		detailFieldValues = append(detailFieldValues, []string{"Claimed", ch.ClaimedTimestamp.Format("Jan 02, 2006")})

		detailFieldContent := ""
		for _, v := range detailFieldValues {
			detailFieldContent += fmt.Sprintf("**%s:** %s\n", v[0], v[1])
		}

		var statFieldValues = [][]string{}
		statFieldValues = append(statFieldValues, []string{"HP", fmt.Sprintf("%d – IV: %d/31", ch.HP(), ch.IvHP)}) // Changed from MaxHP to HP
		statFieldValues = append(statFieldValues, []string{"Attack", fmt.Sprintf("%d – IV: %d/31", ch.Atk(), ch.IvAtk)})
		statFieldValues = append(statFieldValues, []string{"Defense", fmt.Sprintf("%d – IV: %d/31", ch.Def(), ch.IvDef)})
		statFieldValues = append(statFieldValues, []string{"Sp. Atk", fmt.Sprintf("%d – IV: %d/31", ch.SpAtk(), ch.IvSpAtk)})
		statFieldValues = append(statFieldValues, []string{"Sp. Def", fmt.Sprintf("%d – IV: %d/31", ch.SpDef(), ch.IvSpDef)})
		statFieldValues = append(statFieldValues, []string{"Speed", fmt.Sprintf("%d – IV: %d/31", ch.Spd(), ch.IvSpd)})
		statFieldValues = append(statFieldValues, []string{"Total IV", ch.IvPercentage()})

		statFieldContent := ""
		for _, v := range statFieldValues {
			statFieldContent += fmt.Sprintf("**%s:** %s\n", v[0], v[1])
		}

		titleString := ch.CharacterName()
		if ch.Nickname != "" {
			titleString = fmt.Sprintf("%s (%s)", ch.Nickname, ch.CharacterName())
		}
		title := fmt.Sprintf("%s %s - Lvl %d", ch.Data().Emoji, titleString, ch.Level)

		embedBuilder := discord.NewEmbedBuilder().
			SetTitle(title).
			SetThumbnail(e.User().EffectiveAvatarURL()).
			SetFooterTextf("Displaying character #%d | ID: %s", ch.IDX, ch.ID.String()).
			SetColor(constants.ColorDefault).
			SetTimestamp(time.Now()).
			AddFields(
				discord.EmbedField{
					Name:  "Details",
					Value: strings.TrimSpace(detailFieldContent),
				},
				discord.EmbedField{
					Name:  "Stats & IVs",
					Value: strings.TrimSpace(statFieldContent),
				},
			)

		var messageFiles []*discord.File
		image, imgErr := ch.Image()
		if imgErr != nil {
			log.Warn("Failed to get character image for embed",
				logger.CharacterID(ch.ID),
				logger.CharacterName(ch.CharacterName()),
				logger.ErrorField(imgErr),
			)
		} else {
			embedBuilder.SetImage("attachment://" + image.Name)
			messageFiles = append(messageFiles, image)
		}

		log.Debug("Sending info embed",
			logger.DiscordUserID(userID),
			logger.CharacterID(ch.ID),
		)
		return e.CreateMessage(discord.MessageCreate{
			Embeds: []discord.Embed{embedBuilder.Build()},
			Files:  messageFiles,
		})
	}
}

func handleGetCharacterAutocomplete(b *bot.Bot) handler.AutocompleteHandler {
	log := logger.NewLogger("autocomplete.info_character") // Specific logger

	return func(e *handler.AutocompleteEvent) error {
		start := time.Now()
		userID := e.User().ID
		query := e.Data.String("character")

		log.Debug("Autocomplete request for info command received",
			logger.DiscordUserID(userID),
			zap.String("query", query),
		)

		var results []discord.AutocompleteChoiceString
		chars, err := b.DB.GetCharactersForUser(e.Ctx, userID)
		if err != nil {
			log.Error("Failed to get characters for autocomplete (info cmd)",
				logger.DiscordUserID(userID),
				logger.ErrorField(err),
			)
			return e.AutocompleteResult([]discord.AutocompleteChoice{
				discord.AutocompleteChoiceString{Name: "Error fetching your characters", Value: "db_error"},
			})
		}

		if len(chars) == 0 {
			log.Debug("User has no characters for autocomplete (info cmd)", logger.DiscordUserID(userID))
			return e.AutocompleteResult([]discord.AutocompleteChoice{
				discord.AutocompleteChoiceString{Name: "You don't have any characters yet", Value: "no_chars_owned"},
			})
		}

		for _, ch := range chars {
			displayName := ch.CharacterName()
			if ch.Nickname != "" {
				displayName = ch.Nickname
			}
			// Match against IDX, Nickname (if exists), and CharacterName
			searchText := strings.ToLower(fmt.Sprintf("%d %s %s", ch.IDX, displayName, ch.CharacterName()))
			queryLower := strings.ToLower(query)

			if query == "" || strings.Contains(searchText, queryLower) {
				choiceName := fmt.Sprintf("#%d: %s (Lvl %d)", ch.IDX, displayName, ch.Level)
				if ch.Shiny {
					choiceName += " ✨"
				}
				results = append(results, discord.AutocompleteChoiceString{
					Name:  choiceName,
					Value: ch.ID.String(), // Value is the Character's UUID
				})
			}
		}

		var choices []discord.AutocompleteChoice
		for i, r := range results {
			if i >= 25 { // Discord limit
				break
			}
			choices = append(choices, r)
		}

		if len(choices) == 0 && query != "" {
			log.Debug("No matching characters found for autocomplete query (info cmd)",
				logger.DiscordUserID(userID),
				zap.String("query", query),
			)
			// Provide a "no results" option if user typed something but no matches
			choices = append(choices, discord.AutocompleteChoiceString{
				Name:  fmt.Sprintf("No characters found matching '%s'", query),
				Value: "no_match_found",
			})
		}

		log.Debug("Autocomplete results prepared for info command",
			logger.DiscordUserID(userID),
			zap.Int("result_count", len(choices)),
			logger.Duration(time.Since(start)),
		)
		return e.AutocompleteResult(choices)
	}
}
