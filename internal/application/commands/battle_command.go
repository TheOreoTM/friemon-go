package commands

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/entities"
	"github.com/theoreotm/friemon/internal/pkg/logger"
)

func init() {
	Commands["battle"] = cmdBattle
}

var cmdBattle = &Command{
	Cmd: discord.SlashCommandCreate{
		Name:        "battle",
		Description: "Challenge another user to a battle",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{
				Name:        "opponent",
				Description: "The user you want to battle",
				Required:    true,
			},
		},
	},
	Handler:  handleBattle,
	Category: "Battle",
}

func handleBattle(b *bot.Bot) handler.CommandHandler {
	return func(event *handler.CommandEvent) error {
		challengerID := event.User().ID
		opponentUser, ok := event.SlashCommandInteractionData().OptUser("opponent")

		if !ok {
			return event.CreateMessage(ErrorMessage("Invalid opponent specified"))
		}

		opponentID := opponentUser.ID

		// Validation checks
		if challengerID == opponentID {
			return event.CreateMessage(ErrorMessage("You cannot battle yourself!"))
		}

		if opponentUser.Bot {
			return event.CreateMessage(ErrorMessage("You cannot battle bots!"))
		}

		// Check if either user is already in a battle
		existingBattle, err := b.DB.GetActiveBattleForUser(b.Context, challengerID)
		if err == nil && existingBattle != nil {
			return event.CreateMessage(ErrorMessage("You are already in an active battle!"))
		}

		existingBattle, err = b.DB.GetActiveBattleForUser(b.Context, opponentID)
		if err == nil && existingBattle != nil {
			return event.CreateMessage(ErrorMessage("That user is already in an active battle!"))
		}

		// Check if both users have enough characters
		challengerChars, err := b.DB.GetCharactersForUser(b.Context, challengerID)
		if err != nil || len(challengerChars) < 3 {
			return event.CreateMessage(ErrorMessage("You need at least 3 characters to battle!"))
		}

		opponentChars, err := b.DB.GetCharactersForUser(b.Context, opponentID)
		if err != nil || len(opponentChars) < 3 {
			return event.CreateMessage(ErrorMessage("Your opponent needs at least 3 characters to battle!"))
		}

		// Get default game settings
		settings, err := b.DB.GetGameSettings(b.Context)
		if err != nil {
			logger.Error("Failed to get game settings", logger.ErrorField(err))
			// Use defaults
			settings = entities.GameSettings{
				TurnLimit:          25,
				TurnTimeoutSeconds: 60,
				SwitchCostsTurn:    false,
				MaxTeamSize:        3,
			}
		}

		// Create battle challenge
		battle := entities.NewBattle(challengerID, opponentID, settings)

		// Save to database
		err = b.DB.CreateBattle(b.Context, battle)
		if err != nil {
			logger.Error("Failed to create battle", logger.ErrorField(err))
			return event.CreateMessage(ErrorMessage("Failed to create battle. Please try again."))
		}

		// Create challenge embed
		embed := discord.NewEmbedBuilder().
			SetTitle("⚔️ Battle Challenge!").
			SetDescription(fmt.Sprintf("%s has challenged %s to a battle!",
				event.User().Mention(), opponentUser.Mention())).
			SetColor(constants.ColorInfo).
			AddField("Battle Settings", fmt.Sprintf(
				"• Turn Limit: %d\n• Turn Timeout: %d seconds\n• Switch Costs Turn: %t",
				settings.TurnLimit, settings.TurnTimeoutSeconds, settings.SwitchCostsTurn), false).
			SetFooter("The challenge will expire in 2 minutes", "").
			SetTimestamp(time.Now()).
			Build()

		components := []discord.ContainerComponent{
			discord.NewActionRow(
				discord.NewPrimaryButton("Accept Battle", fmt.Sprintf("battle_accept_%s", battle.ID.String())),
				discord.NewDangerButton("Decline Battle", fmt.Sprintf("battle_decline_%s", battle.ID.String())),
			),
		}

		response := discord.NewMessageCreateBuilder().
			SetEmbeds(embed).
			SetContainerComponents(components...).
			Build()

		// Schedule timeout for challenge
		_, err = b.Scheduler.After(2*time.Minute).
			Type("battle_timeout").
			With("battle_id", battle.ID.String()).
			Execute("battle_timeout")

		if err != nil {
			logger.Warn("Failed to schedule battle timeout", logger.ErrorField(err))
		}

		return event.CreateMessage(response)
	}
}
