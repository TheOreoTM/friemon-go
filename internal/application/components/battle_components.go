package components

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/core/entities"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

func init() {
	Components["battle_accept"] = battleAcceptHandler
	Components["battle_decline"] = battleDeclineHandler
	Components["team_select"] = teamSelectHandler
	Components["battle_move"] = battleMoveHandler
	Components["battle_switch"] = battleSwitchHandler
	Components["battle_forfeit"] = battleForfeitHandler
}

func battleAcceptHandler(b *bot.Bot) handler.ComponentHandler {
	return func(event *handler.ComponentEvent) error {
		battleID := strings.TrimPrefix(event.ButtonInteractionData().CustomID(), "battle_accept_")
		userID := event.User().ID
		
		// Get battle from database
		battle, err := b.DB.GetBattle(b.Context, uuid.MustParse(battleID))
		if err != nil {
			return event.CreateMessage(ErrorMessage("Battle not found or expired"))
		}
		
		// Check if user is the opponent
		if userID != battle.OpponentID {
			return event.CreateMessage(ErrorMessage("Only the challenged user can accept this battle"))
		}
		
		// Check if battle is still pending
		if battle.Status != entities.BattleStatusPending {
			return event.CreateMessage(ErrorMessage("This battle challenge has expired"))
		}
		
		// Update battle status
		battle.Status = entities.BattleStatusActive
		err = b.DB.UpdateBattle(b.Context, battle)
		if err != nil {
			logger.Error("Failed to update battle status", logger.ErrorField(err))
			return event.CreateMessage(ErrorMessage("Failed to start battle"))
		}
		
		// Create threads for team selection
		err = b.createBattleThreads(event.ChannelID(), battle)
		if err != nil {
			logger.Error("Failed to create battle threads", logger.ErrorField(err))
			return event.CreateMessage(ErrorMessage("Failed to create battle threads"))
		}
		
		// Update original message
		embed := discord.NewEmbedBuilder().
			SetTitle("⚔️ Battle Accepted!").
			SetDescription("Battle threads have been created. Please select your teams!").
			SetColor(constants.ColorSuccess).
			Build()
		
		_, err = event.UpdateMessage(discord.NewMessageUpdateBuilder().
			SetEmbeds(embed).
			ClearComponents().
			Build())
		
		if err != nil {
			logger.Error("Failed to update battle message", logger.ErrorField(err))
		}
		
		// Start team selection phase
		return b.startTeamSelection(battle)
	}
}

func battleDeclineHandler(b *bot.Bot) handler.ComponentHandler {
	return func(event *handler.ComponentEvent) error {
		battleID := strings.TrimPrefix(event.ButtonInteractionData().CustomID(), "battle_decline_")
		userID := event.User().ID
		
		// Get battle from database
		battle, err := b.DB.GetBattle(b.Context, uuid.MustParse(battleID))
		if err != nil {
			return event.CreateMessage(ErrorMessage("Battle not found"))
		}
		
		// Check if user is the opponent
		if userID != battle.OpponentID {
			return event.CreateMessage(ErrorMessage("Only the challenged user can decline this battle"))
		}
		
		// Update battle status
		battle.Status = entities.BattleStatusCancelled
		err = b.DB.UpdateBattle(b.Context, battle)
		if err != nil {
			logger.Error("Failed to update battle status", logger.ErrorField(err))
		}
		
		// Update message
		embed := discord.NewEmbedBuilder().
			SetTitle("❌ Battle Declined").
			SetDescription("The battle challenge was declined.").
			SetColor(constants.ColorFail).
			Build()
		
		_, err = event.UpdateMessage(discord.NewMessageUpdateBuilder().
			SetEmbeds(embed).
			ClearComponents().
			Build())
		
		return err
	}
}

func teamSelectHandler(b *bot.Bot) handler.ComponentHandler {
	return func(event *handler.ComponentEvent) error {
		data := event.StringSelectMenuInteractionData()
		battleID := strings.TrimPrefix(data.CustomID(), "team_select_")
		selectedChars := data.Values
		userID := event.User().ID
		
		if len(selectedChars) != 3 {
			return event.CreateMessage(ErrorMessage("You must select exactly 3 characters"))
		}
		
		// Get battle
		battle, err := b.DB.GetBattle(b.Context, uuid.MustParse(battleID))
		if err != nil {
			return event.CreateMessage(ErrorMessage("Battle not found"))
		}
		
		// Get characters and validate
		var characters []*entities.Character
		for i, charIDStr := range selectedChars {
			charID := uuid.MustParse(charIDStr)
			char, err := b.DB.GetCharacter(b.Context, charID)
			if err != nil || char.OwnerID != userID.String() {
				return event.CreateMessage(ErrorMessage("Invalid character selection"))
			}
			characters = append(characters, char)
			
			// Create battle team member
			teamMember := entities.NewBattleTeamMember(battle.ID, userID, i+1, char)
			err = b.DB.CreateBattleTeamMember(b.Context, teamMember)
			if err != nil {
				logger.Error("Failed to create battle team member", logger.ErrorField(err))
				return event.CreateMessage(ErrorMessage("Failed to save team"))
			}
		}
		
		// Update message to show team was selected
		embed := discord.NewEmbedBuilder().
			SetTitle("✅ Team Selected!").
			SetDescription("Your team has been saved. Waiting for opponent...").
			SetColor(constants.ColorSuccess).
			Build()
		
		err = event.UpdateMessage(discord.NewMessageUpdateBuilder().
			SetEmbeds(embed).
			ClearComponents().
			Build())
		
		if err != nil {
			logger.Error("Failed to update team selection message", logger.ErrorField(err))
		}
		
		// Check if both teams are ready
		return b.checkBattleReady(battle)
	}
}

func battleMoveHandler(b *bot.Bot) handler.ComponentHandler {
	return func(event *handler.ComponentEvent) error {
		data := event.StringSelectMenuInteractionData()
		parts := strings.Split(data.CustomID(), "_")
		if len(parts) != 3 {
			return event.CreateMessage(ErrorMessage("Invalid move selection"))
		}
		
		battleID := uuid.MustParse(parts[2])
		moveID := data.Values[0]
		userID := event.User().ID
		
		// Process move action
		return b.processBattleAction(battleID, userID, entities.ActionMove, map[string]interface{}{
			"move_id": moveID,
		})
	}
}

func battleSwitchHandler(b *bot.Bot) handler.ComponentHandler {
	return func(event *handler.ComponentEvent) error {
		data := event.StringSelectMenuInteractionData()
		parts := strings.Split(data.CustomID(), "_")
		if len(parts) != 3 {
			return event.CreateMessage(ErrorMessage("Invalid switch selection"))
		}
		
		battleID := uuid.MustParse(parts[2])
		characterIndex := data.Values[0]
		userID := event.User().ID
		
		// Process switch action
		return b.processBattleAction(battleID, userID, entities.ActionSwitch, map[string]interface{}{
			"character_index": characterIndex,
		})
	}
}

func battleForfeitHandler(b *bot.Bot) handler.ComponentHandler {
	return func(event *handler.ComponentEvent) error {
		battleID := strings.TrimPrefix(event.ButtonInteractionData().CustomID(), "battle_forfeit_")
		userID := event.User().ID
		
		// Process forfeit action
		return b.processBattleAction(uuid.MustParse(battleID), userID, entities.ActionForfeit, map[string]interface{}{})
	}
}

// Helper methods for bot
func (b *bot.Bot) createBattleThreads(channelID snowflake.ID, battle *entities.Battle) error {
	// Create main battle thread
	mainThread, err := b.Client.Rest().CreateThread(channelID, discord.ThreadFromMessage{
		Name:                fmt.Sprintf("⚔️ Battle: %s vs %s", battle.ChallengerID.String()[:8], battle.OpponentID.String()[:8]),
		AutoArchiveDuration: discord.AutoArchiveDurationOneHour,
		Type:                discord.ChannelTypeGuildPublicThread,
	})
	if err != nil {
		return fmt.Errorf("failed to create main thread: %w", err)
	}
	
	// Create challenger thread
	challengerThread, err := b.Client.Rest().CreateThread(channelID, discord.ThreadFromMessage{
		Name:                fmt.Sprintf("🔒 %s's Battle Actions", battle.ChallengerID.String()[:8]),
		AutoArchiveDuration: discord.AutoArchiveDurationOneHour,
		Type:                discord.ChannelTypeGuildPrivateThread,
	})
	if err != nil {
		return fmt.Errorf("failed to create challenger thread: %w", err)
	}
	
	// Create opponent thread
	opponentThread, err := b.Client.Rest().CreateThread(channelID, discord.ThreadFromMessage{
		Name:                fmt.Sprintf("🔒 %s's Battle Actions", battle.OpponentID.String()[:8]),
		AutoArchiveDuration: discord.AutoArchiveDurationOneHour,
		Type:                discord.ChannelTypeGuildPrivateThread,
	})
	if err != nil {
		return fmt.Errorf("failed to create opponent thread: %w", err)
	}
	
	// Update battle with thread IDs
	battle.MainThreadID = &mainThread.ID
	battle.ChallengerThreadID = &challengerThread.ID
	battle.OpponentThreadID = &opponentThread.ID
	
	return b.DB.UpdateBattle(b.Context, battle)
}

func (b *bot.Bot) startTeamSelection(battle *entities.Battle) error {
	// Get characters for both users
	challengerChars, err := b.DB.GetCharactersForUser(b.Context, battle.ChallengerID)
	if err != nil {
		return err
	}
	
	opponentChars, err := b.DB.GetCharactersForUser(b.Context, battle.OpponentID)
	if err != nil {
		return err
	}
	
	// Send team selection to challenger thread
	err = b.sendTeamSelection(*battle.ChallengerThreadID, battle.ID, challengerChars)
	if err != nil {
		return err
	}
	
	// Send team selection to opponent thread
	return b.sendTeamSelection(*battle.OpponentThreadID, battle.ID, opponentChars)
}

func (b *bot.Bot) sendTeamSelection(threadID snowflake.ID, battleID uuid.UUID, characters []entities.Character) error {
	embed := discord.NewEmbedBuilder().
		SetTitle("🔧 Select Your Battle Team").
		SetDescription("Choose 3 characters for your battle team. Order matters - first character will be sent out first!").
		SetColor(constants.ColorInfo).
		Build()
	
	// Create select menu with characters
	options := make([]discord.StringSelectMenuOption, 0, len(characters))
	for _, char := range characters {
		name := char.CharacterName()
		if char.Nickname != "" {
			name = char.Nickname
		}
		
		options = append(options, discord.StringSelectMenuOption{
			Label:       fmt.Sprintf("#%d %s (Lv.%d)", char.IDX, name, char.Level),
			Value:       char.ID.String(),
			Description: fmt.Sprintf("HP: %d | %s", char.MaxHP(), char.IvPercentage()),
		})
	}
	
	selectMenu := discord.NewStringSelectMenu(
		fmt.Sprintf("team_select_%s", battleID.String()),
		"Choose 3 characters for your team...",
		options...,
	).WithMinValues(3).WithMaxValues(3)
	
	components := []discord.ContainerComponent{
		discord.NewActionRow(selectMenu),
	}
	
	message := discord.NewMessageCreateBuilder().
		SetEmbeds(embed).
		SetComponents(components...).
		Build()
	
	_, err := b.Client.Rest().CreateMessage(threadID, message)
	return err
}

func (b *bot.Bot) checkBattleReady(battle *entities.Battle) error {
	// Check if both teams have 3 members
	challengerTeam, err := b.DB.GetBattleTeam(b.Context, battle.ID, battle.ChallengerID)
	if err != nil || len(challengerTeam) != 3 {
		return nil // Not ready yet
	}
	
	opponentTeam, err := b.DB.GetBattleTeam(b.Context, battle.ID, battle.OpponentID)
	if err != nil || len(opponentTeam) != 3 {
		return nil // Not ready yet
	}
	
	// Both teams ready, start battle!
	return b.startBattle(battle)
}

func (b *bot.Bot) startBattle(battle *entities.Battle) error {
	// Load teams into battle object
	challengerTeam, err := b.DB.GetBattleTeam(b.Context, battle.ID, battle.ChallengerID)
	if err != nil {
		return err
	}
	
	opponentTeam, err := b.DB.GetBattleTeam(b.Context, battle.ID, battle.OpponentID)
	if err != nil {
		return err
	}
	
	battle.ChallengerTeam = challengerTeam
	battle.OpponentTeam = opponentTeam
	
	// Determine first turn (faster character goes first)
	challengerActive := battle.GetActiveCharacter(battle.ChallengerID)
	opponentActive := battle.GetActiveCharacter(battle.OpponentID)
	
	if challengerActive.CalculateEffectiveStat("spd") >= opponentActive.CalculateEffectiveStat("spd") {
		battle.CurrentTurnPlayer = &battle.ChallengerID
	} else {
		battle.CurrentTurnPlayer = &battle.OpponentID
	}
	
	// Update battle in database
	err = b.DB.UpdateBattle(b.Context, battle)
	if err != nil {
		return err
	}
	
	// Send battle start message to main thread
	embed := discord.NewEmbedBuilder().
		SetTitle("⚔️ Battle Begin!").
		SetDescription("The battle has started! Trainers, send out your first characters!").
		SetColor(constants.ColorInfo).
		AddField("Turn Order", fmt.Sprintf("%s goes first!", battle.CurrentTurnPlayer.String()), false).
		Build()
	
	_, err = b.Client.Rest().CreateMessage(*battle.MainThreadID, discord.NewMessageCreateBuilder().
		SetEmbeds(embed).
		Build())
	
	if err != nil {
		logger.Error("Failed to send battle start message", logger.ErrorField(err))
	}
	
	// Start first turn
	return b.startBattleTurn(battle)
}

func (b *bot.Bot) startBattleTurn(battle *entities.Battle) error {
	if battle.CurrentTurnPlayer == nil {
		return fmt.Errorf("no current turn player")
	}
	
	userID := *battle.CurrentTurnPlayer
	threadID := battle.ChallengerThreadID
	if userID == battle.OpponentID {
		threadID = battle.OpponentThreadID
	}
	
	// Get current active character
	activeChar := battle.GetActiveCharacter(userID)
	if activeChar == nil {
		return fmt.Errorf("no active character for user %s", userID)
	}
	
	// Send turn prompt
	embed := discord.NewEmbedBuilder().
		SetTitle("🎯 Your Turn!").
		SetDescription(fmt.Sprintf("**%s** (HP: %d/%d)\nChoose your action:", 
			activeChar.CharacterData.CharacterName(),
			activeChar.CurrentHP,
			activeChar.CharacterData.MaxHP())).
		SetColor(constants.ColorInfo).
		Build()
	
	// Create action components
	components := b.createBattleActionComponents(battle, userID)
	
	message := discord.NewMessageCreateBuilder().
		SetEmbeds(embed).
		SetComponents(components...).
		Build()
	
	_, err := b.Client.Rest().CreateMessage(*threadID, message)
	if err != nil {
		return err
	}
	
	// Schedule turn timeout
	_, err = b.Scheduler.After(time.Duration(battle.Settings.TurnTimeoutSeconds) * time.Second).
		Type("battle_turn_timeout").
		With("battle_id", battle.ID.String()).
		With("user_id", userID.String()).
		Execute("battle_turn_timeout")
	
	return err
}

func (b *bot.Bot) createBattleActionComponents(battle *entities.Battle, userID snowflake.ID) []discord.ContainerComponent {
	activeChar := battle.GetActiveCharacter(userID)
	if activeChar == nil {
		return nil
	}
	
	var components []discord.ContainerComponent
	
	// Move selection
	if len(activeChar.CharacterData.Moves) > 0 {
		moveOptions := make([]discord.StringSelectMenuOption, 0, len(activeChar.CharacterData.Moves))
		for i, moveID := range activeChar.CharacterData.Moves {
			// Get move data (you'll need to implement move lookup)
			moveOptions = append(moveOptions, discord.StringSelectMenuOption{
				Label:       fmt.Sprintf("Move %d", moveID),
				Value:       fmt.Sprintf("%d", moveID),
				Description: fmt.Sprintf("Slot %d", i+1),
			})
		}
		
		moveSelect := discord.NewStringSelectMenu(
			fmt.Sprintf("battle_move_%s", battle.ID.String()),
			"Choose a move...",
			moveOptions...,
		)
		
		components = append(components, discord.NewActionRow(moveSelect))
	}
	
	// Switch options (only if there are non-fainted characters)
	team := battle.GetTeam(userID)
	var switchOptions []discord.StringSelectMenuOption
	for i, member := range team {
		if !member.IsFainted && !member.IsActive {
			switchOptions = append(switchOptions, discord.StringSelectMenuOption{
				Label:       member.CharacterData.CharacterName(),
				Value:       fmt.Sprintf("%d", i),
				Description: fmt.Sprintf("HP: %d/%d", member.CurrentHP, member.CharacterData.MaxHP()),
			})
		}
	}
	
	if len(switchOptions) > 0 {
		switchSelect := discord.NewStringSelectMenu(
			fmt.Sprintf("battle_switch_%s", battle.ID.String()),
			"Switch character...",
			switchOptions...,
		)
		
		components = append(components, discord.NewActionRow(switchSelect))
	}
	
	// Forfeit button
	forfeitButton := discord.NewDangerButton(
		"Forfeit Battle",
		fmt.Sprintf("battle_forfeit_%s", battle.ID.String()),
	)
	
	components = append(components, discord.NewActionRow(forfeitButton))
	
	return components
}

func (b *bot.Bot) processBattleAction(battleID uuid.UUID, userID snowflake.ID, actionType entities.ActionType, actionData map[string]interface{}) error {
	// Get battle
	battle, err := b.DB.GetBattle(b.Context, battleID)
	if err != nil {
		return fmt.Errorf("battle not found: %w", err)
	}
	
	// Validate it's the user's turn
	if !battle.IsPlayersTurn(userID) {
		return fmt.Errorf("it's not your turn")
	}
	
	// Process the action based on type
	switch actionType {
	case entities.ActionMove:
		return b.processMoveAction(battle, userID, actionData)
	case entities.ActionSwitch:
		return b.processSwitchAction(battle, userID, actionData)
	case entities.ActionForfeit:
		return b.processForfeitAction(battle, userID)
	default:
		return fmt.Errorf("unknown action type: %s", actionType)
	}
}

func (b *bot.Bot) processMoveAction(battle *entities.Battle, userID snowflake.ID, actionData map[string]interface{}) error {
	// Implementation for processing move actions
	// This is quite complex and involves damage calculation, type effectiveness, status effects, etc.
	// For brevity, I'll provide a simplified version
	
	logger.Info("Processing move action", 
		logger.UserID(userID.String()),
		zap.String("battle_id", battle.ID.String()),
		zap.Any("action_data", actionData))
	
	// TODO: Implement full move processing
	return nil
}

func (b *bot.Bot) processSwitchAction(battle *entities.Battle, userID snowflake.ID, actionData map[string]interface{}) error {
	// Implementation for processing switch actions
	logger.Info("Processing switch action",
		logger.UserID(userID.String()),
		zap.String("battle_id", battle.ID.String()),
		zap.Any("action_data", actionData))
	
	// TODO: Implement full switch processing
	return nil
}

func (b *bot.Bot) processForfeitAction(battle *entities.Battle, userID snowflake.ID) error {
	// Set winner as opponent
	opponentID := battle.GetOpponentID(userID)
	battle.WinnerID = &opponentID
	battle.Status = entities.BattleStatusCompleted
	battle.CompletedAt = &time.Time{}
	*battle.CompletedAt = time.Now()
	
	// Update battle in database
	err := b.DB.UpdateBattle(b.Context, battle)
	if err != nil {
		return err
	}
	
	// Update ELO ratings
	return b.updateEloRatings(battle)
}

func (b *bot.Bot) updateEloRatings(battle *entities.Battle) error {
	if battle.WinnerID == nil {
		return fmt.Errorf("no winner determined")
	}
	
	winnerID := *battle.WinnerID
	loserID := battle.GetOpponentID(winnerID)
	
	// Get current ELO ratings
	winnerElo, err := b.DB.GetUserElo(b.Context, winnerID)
	if err != nil {
		// Create new ELO record if doesn't exist
		winnerElo = &entities.UserElo{
			UserID:    winnerID,
			EloRating: 1000,
		}
	}
	
	loserElo, err := b.DB.GetUserElo(b.Context, loserID)
	if err != nil {
		loserElo = &entities.UserElo{
			UserID:    loserID,
			EloRating: 1000,
		}
	}
	
	// Calculate ELO changes
	winnerChange, loserChange := entities.CalculateEloChange(winnerElo.EloRating, loserElo.EloRating)
	
	// Update ratings
	winnerElo.EloRating += winnerChange
	winnerElo.BattlesWon++
	winnerElo.BattlesTotal++
	if winnerElo.EloRating > winnerElo.HighestElo {
		winnerElo.HighestElo = winnerElo.EloRating
	}
	
	loserElo.EloRating += loserChange // loserChange is negative
	loserElo.BattlesLost++
	loserElo.BattlesTotal++
	
	// Save to database
	err = b.DB.UpdateUserElo(b.Context, winnerElo)
	if err != nil {
		return err
	}
	
	return b.DB.UpdateUserElo(b.Context, loserElo)
}