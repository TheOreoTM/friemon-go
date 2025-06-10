package game

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
)

type BattleState int

const (
	BattleStateWaitingForPlayers BattleState = iota
	BattleStateTeamSelection
	BattleStateInProgress
	BattleStateFinished
	BattleStateCancelled
)

func (bs BattleState) String() string {
	switch bs {
	case BattleStateWaitingForPlayers:
		return "Waiting for Players"
	case BattleStateTeamSelection:
		return "Team Selection"
	case BattleStateInProgress:
		return "In Progress"
	case BattleStateFinished:
		return "Finished"
	case BattleStateCancelled:
		return "Cancelled"
	default:
		return "Unknown"
	}
}

type BattleAction int

const (
	ActionAttack BattleAction = iota
	ActionSwitch
	ActionSkip
)

type PlayerAction struct {
	PlayerID snowflake.ID `json:"player_id"`
	Action   BattleAction `json:"action"`
	MoveID   int          `json:"move_id,omitempty"`
	TargetID uuid.UUID    `json:"target_id,omitempty"`
	SwitchTo int          `json:"switch_to,omitempty"` // Index in team
}

type BattlePlayer struct {
	ID              snowflake.ID   `json:"id"`
	Team            []*Character   `json:"team"`
	ActiveCharacter int            `json:"active_character"` // Index in team
	ELORating       int            `json:"elo_rating"`
	ActionsThisTurn []PlayerAction `json:"actions_this_turn"`
}

func (bp *BattlePlayer) GetActiveCharacter() *Character {
	if bp.ActiveCharacter >= 0 && bp.ActiveCharacter < len(bp.Team) {
		return bp.Team[bp.ActiveCharacter]
	}
	return nil
}

func (bp *BattlePlayer) HasAlivePokemon() bool {
	for _, char := range bp.Team {
		if !char.BattleStats.IsFainted() {
			return true
		}
	}
	return false
}

func (bp *BattlePlayer) GetAlivePokemonCount() int {
	count := 0
	for _, char := range bp.Team {
		if !char.BattleStats.IsFainted() {
			count++
		}
	}
	return count
}

type Battle struct {
	ID          uuid.UUID              `json:"id"`
	ChannelID   snowflake.ID           `json:"channel_id"`
	ThreadID    snowflake.ID           `json:"thread_id"`
	Player1     *BattlePlayer          `json:"player1"`
	Player2     *BattlePlayer          `json:"player2"`
	State       BattleState            `json:"state"`
	CurrentTurn int                    `json:"current_turn"`
	TurnOrder   []snowflake.ID         `json:"turn_order"`
	Settings    GameSettings           `json:"settings"`
	StartedAt   time.Time              `json:"started_at"`
	FinishedAt  *time.Time             `json:"finished_at,omitempty"`
	Winner      *snowflake.ID          `json:"winner,omitempty"`
	BattleLog   []string               `json:"battle_log"`
	Weather     string                 `json:"weather,omitempty"`
	Field       map[string]interface{} `json:"field,omitempty"`
}

func NewBattle(channelID snowflake.ID, player1ID, player2ID snowflake.ID) *Battle {
	return &Battle{
		ID:        uuid.New(),
		ChannelID: channelID,
		Player1: &BattlePlayer{
			ID:              player1ID,
			Team:            make([]*Character, 0, 3),
			ActiveCharacter: 0,
			ActionsThisTurn: make([]PlayerAction, 0),
		},
		Player2: &BattlePlayer{
			ID:              player2ID,
			Team:            make([]*Character, 0, 3),
			ActiveCharacter: 0,
			ActionsThisTurn: make([]PlayerAction, 0),
		},
		State:       BattleStateWaitingForPlayers,
		CurrentTurn: 1,
		TurnOrder:   make([]snowflake.ID, 0),
		Settings:    DefaultGameSettings(),
		StartedAt:   time.Now(),
		BattleLog:   make([]string, 0),
		Field:       make(map[string]interface{}),
	}
}

func (b *Battle) GetPlayer(playerID snowflake.ID) *BattlePlayer {
	if b.Player1.ID == playerID {
		return b.Player1
	}
	if b.Player2.ID == playerID {
		return b.Player2
	}
	return nil
}

func (b *Battle) GetOpponent(playerID snowflake.ID) *BattlePlayer {
	if b.Player1.ID == playerID {
		return b.Player2
	}
	if b.Player2.ID == playerID {
		return b.Player1
	}
	return nil
}

func (b *Battle) IsPlayerInBattle(playerID snowflake.ID) bool {
	return b.Player1.ID == playerID || b.Player2.ID == playerID
}

func (b *Battle) AddToLog(message string) {
	b.BattleLog = append(b.BattleLog, message)
}

func (b *Battle) CanStart() bool {
	return b.State == BattleStateTeamSelection &&
		len(b.Player1.Team) == b.Settings.TeamSize &&
		len(b.Player2.Team) == b.Settings.TeamSize
}

func (b *Battle) Start() error {
	if !b.CanStart() {
		return fmt.Errorf("battle cannot start: invalid state or incomplete teams")
	}

	// Initialize battle stats for all characters
	for _, char := range b.Player1.Team {
		char.InitializeBattleStats()
	}
	for _, char := range b.Player2.Team {
		char.InitializeBattleStats()
	}

	b.State = BattleStateInProgress
	b.AddToLog(fmt.Sprintf("Battle between %s and %s has begun!", b.Player1.ID, b.Player2.ID))

	return nil
}

func (b *Battle) CalculateTurnOrder() {
	b.TurnOrder = make([]snowflake.ID, 0, 2)

	char1 := b.Player1.GetActiveCharacter()
	char2 := b.Player2.GetActiveCharacter()

	if char1 == nil || char2 == nil {
		return
	}

	// Calculate effective speed (including stat stages)
	speed1 := float64(char1.Spd()) * char1.BattleStats.GetStatMultiplier("spe")
	speed2 := float64(char2.Spd()) * char2.BattleStats.GetStatMultiplier("spe")

	// Handle paralysis (halves speed)
	if char1.BattleStats.HasStatusEffect(constants.StatusParalyze) {
		speed1 *= 0.5
	}
	if char2.BattleStats.HasStatusEffect(constants.StatusParalyze) {
		speed2 *= 0.5
	}

	// Determine order
	if speed1 > speed2 {
		b.TurnOrder = []snowflake.ID{b.Player1.ID, b.Player2.ID}
	} else if speed2 > speed1 {
		b.TurnOrder = []snowflake.ID{b.Player2.ID, b.Player1.ID}
	} else {
		// Speed tie - random
		if rand.Intn(2) == 0 {
			b.TurnOrder = []snowflake.ID{b.Player1.ID, b.Player2.ID}
		} else {
			b.TurnOrder = []snowflake.ID{b.Player2.ID, b.Player1.ID}
		}
	}
}

func (b *Battle) ProcessTurn() error {
	if b.State != BattleStateInProgress {
		return fmt.Errorf("battle is not in progress")
	}

	b.AddToLog(fmt.Sprintf("--- Turn %d ---", b.CurrentTurn))

	// Process priority moves first, then regular moves
	actions := b.collectAllActions()
	sortedActions := b.sortActionsByPriority(actions)

	for _, action := range sortedActions {
		player := b.GetPlayer(action.PlayerID)
		if player == nil {
			continue
		}

		char := player.GetActiveCharacter()
		if char == nil || char.BattleStats.IsFainted() {
			continue
		}

		err := b.executeAction(action)
		if err != nil {
			b.AddToLog(fmt.Sprintf("Error executing action: %v", err))
			continue
		}

		// Check for battle end after each action
		if b.checkBattleEnd() {
			return nil
		}
	}

	// Process end-of-turn effects
	b.processEndOfTurnEffects()

	// Check battle end conditions
	if b.checkBattleEnd() {
		return nil
	}

	// Check turn limit
	if b.CurrentTurn >= b.Settings.MaxTurns {
		b.endBattleDueToTurnLimit()
		return nil
	}

	b.CurrentTurn++

	// Clear actions for next turn
	b.Player1.ActionsThisTurn = make([]PlayerAction, 0)
	b.Player2.ActionsThisTurn = make([]PlayerAction, 0)

	return nil
}

func (b *Battle) collectAllActions() []PlayerAction {
	actions := make([]PlayerAction, 0)
	actions = append(actions, b.Player1.ActionsThisTurn...)
	actions = append(actions, b.Player2.ActionsThisTurn...)
	return actions
}

func (b *Battle) sortActionsByPriority(actions []PlayerAction) []PlayerAction {
	// Sort by move priority, then by speed
	sorted := make([]PlayerAction, len(actions))
	copy(sorted, actions)

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			priority1 := b.getActionPriority(sorted[i])
			priority2 := b.getActionPriority(sorted[j])

			if priority1 < priority2 {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			} else if priority1 == priority2 {
				// Same priority, check speed
				speed1 := b.getPlayerSpeed(sorted[i].PlayerID)
				speed2 := b.getPlayerSpeed(sorted[j].PlayerID)

				if speed1 < speed2 {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
	}

	return sorted
}

func (b *Battle) getActionPriority(action PlayerAction) int {
	if action.Action == ActionSwitch {
		return 6 // Switching has highest priority
	}

	if action.Action == ActionAttack {
		if move, exists := GetMoveByID(action.MoveID); exists {
			return move.Priority
		}
	}

	return 0
}

func (b *Battle) getPlayerSpeed(playerID snowflake.ID) float64 {
	player := b.GetPlayer(playerID)
	if player == nil {
		return 0
	}

	char := player.GetActiveCharacter()
	if char == nil {
		return 0
	}

	speed := float64(char.Spd()) * char.BattleStats.GetStatMultiplier("spe")

	if char.BattleStats.HasStatusEffect(constants.StatusParalyze) {
		speed *= 0.5
	}

	return speed
}

func (b *Battle) executeAttack(player *BattlePlayer, action PlayerAction) error {
	attacker := player.GetActiveCharacter()
	if attacker == nil || attacker.BattleStats.IsFainted() {
		return fmt.Errorf("no active character or character is fainted")
	}

	// Check if character can move
	if !attacker.BattleStats.CanMove() {
		if attacker.BattleStats.MustRecharge {
			b.AddToLog(fmt.Sprintf("%s must recharge!", attacker.CharacterName()))
		} else if attacker.BattleStats.HasStatusEffect(constants.StatusSleep) {
			b.AddToLog(fmt.Sprintf("%s is asleep!", attacker.CharacterName()))
		} else if attacker.BattleStats.HasStatusEffect(constants.StatusFreeze) {
			b.AddToLog(fmt.Sprintf("%s is frozen solid!", attacker.CharacterName()))
		} else if attacker.BattleStats.HasStatusEffect(constants.StatusParalyze) {
			b.AddToLog(fmt.Sprintf("%s is paralyzed and can't move!", attacker.CharacterName()))
		}
		return nil
	}

	// Check for confusion
	if attacker.BattleStats.HasStatusEffect(constants.StatusConfuse) {
		if rand.Intn(100) < 33 { // 33% chance to hurt self in confusion
			damage := b.calculateConfusionDamage(attacker)
			attacker.BattleStats.TakeDamage(damage)
			b.AddToLog(fmt.Sprintf("%s hurt itself in confusion for %d damage!", attacker.CharacterName(), damage))

			if attacker.BattleStats.IsFainted() {
				b.AddToLog(fmt.Sprintf("%s fainted!", attacker.CharacterName()))
			}
			return nil
		} else {
			b.AddToLog(fmt.Sprintf("%s snapped out of confusion for this turn!", attacker.CharacterName()))
		}
	}

	// Get the move
	move, exists := GetMoveByID(action.MoveID)
	if !exists {
		return fmt.Errorf("move not found")
	}

	// Check if character knows this move
	if !b.characterKnowsMove(attacker, action.MoveID) {
		return fmt.Errorf("character doesn't know this move")
	}

	b.AddToLog(fmt.Sprintf("%s used %s!", attacker.CharacterName(), move.Name))

	// Handle charging moves
	if move.Effect != nil && move.Effect.ChargeRequired && attacker.BattleStats.ChargingMove == nil {
		attacker.BattleStats.ChargingMove = &move
		b.AddToLog(fmt.Sprintf("%s is charging power!", attacker.CharacterName()))
		return nil
	}

	// If this is the second turn of a charging move, clear the charging state
	if attacker.BattleStats.ChargingMove != nil {
		attacker.BattleStats.ChargingMove = nil
	}

	// Execute the move based on its target
	opponent := b.GetOpponent(player.ID)
	if opponent == nil {
		return fmt.Errorf("opponent not found")
	}

	switch move.Target {
	case TargetSingleFoe:
		return b.executeSingleTargetMove(attacker, opponent.GetActiveCharacter(), move)
	case TargetAllFoes:
		return b.executeMultiTargetMove(attacker, opponent.Team, move, false)
	case TargetUser:
		return b.executeSelfTargetMove(attacker, move)
	case TargetSingleAlly:
		// In 1v1 battles, this targets self
		return b.executeSelfTargetMove(attacker, move)
	default:
		return b.executeSingleTargetMove(attacker, opponent.GetActiveCharacter(), move)
	}
}

func (b *Battle) characterKnowsMove(char *Character, moveID int) bool {
	for _, knownMoveID := range char.Moves {
		if int(knownMoveID) == moveID {
			return true
		}
	}
	return false
}

func (b *Battle) calculateConfusionDamage(char *Character) int {
	// Confusion damage is based on a 40 power physical move against self
	attack := float64(char.Atk())
	defense := float64(char.Def())
	level := float64(char.Level)

	damage := ((((2*level/5 + 2) * 40 * attack / defense) / 50) + 2)
	return int(damage)
}

func (b *Battle) executeSingleTargetMove(attacker, target *Character, move Move) error {
	if target == nil || target.BattleStats.IsFainted() {
		b.AddToLog("But there was no target!")
		return nil
	}

	// Check if target is protected
	if target.BattleStats.ProtectedTurns > 0 && move.AffectedByProtect {
		b.AddToLog(fmt.Sprintf("%s protected itself!", target.CharacterName()))
		return nil
	}

	// Calculate damage
	damageResult := CalculateDamage(attacker, target, move, b.Settings)

	if !damageResult.Hit {
		b.AddToLog(fmt.Sprintf("%s's attack missed!", attacker.CharacterName()))
		return nil
	}

	// Apply damage
	if damageResult.Damage > 0 {
		target.BattleStats.TakeDamage(damageResult.Damage)
		b.AddToLog(fmt.Sprintf("%s took %d damage!", target.CharacterName(), damageResult.Damage))

		if damageResult.IsCritical {
			b.AddToLog("A critical hit!")
		}

		if damageResult.EffectivenessText != "" {
			b.AddToLog(damageResult.EffectivenessText)
		}

		if b.Settings.ShowDamageCalculation && damageResult.CalculationDetails != "" {
			b.AddToLog(fmt.Sprintf("Calculation: %s", damageResult.CalculationDetails))
		}
	}

	// Apply move effects
	b.applyMoveEffects(attacker, target, move, damageResult.Damage)

	// Handle recoil
	if move.Effect != nil && move.Effect.Recoil > 0 {
		recoilDamage := (damageResult.Damage * move.Effect.Recoil) / 100
		if recoilDamage > 0 {
			attacker.BattleStats.TakeDamage(recoilDamage)
			b.AddToLog(fmt.Sprintf("%s took %d recoil damage!", attacker.CharacterName(), recoilDamage))
		}
	}

	// Handle drain
	if move.Effect != nil && move.Effect.DrainPercentage > 0 {
		healAmount := (damageResult.Damage * move.Effect.DrainPercentage) / 100
		if healAmount > 0 {
			attacker.BattleStats.Heal(healAmount)
			b.AddToLog(fmt.Sprintf("%s restored %d HP!", attacker.CharacterName(), healAmount))
		}
	}

	// Check if target fainted
	if target.BattleStats.IsFainted() {
		b.AddToLog(fmt.Sprintf("%s fainted!", target.CharacterName()))
	}

	// Update last move used
	attacker.BattleStats.LastMoveUsed = &move
	attacker.BattleStats.MovesUsed = append(attacker.BattleStats.MovesUsed, move)

	return nil
}

func (b *Battle) executeMultiTargetMove(attacker *Character, targets []*Character, move Move, includeAllies bool) error {
	hitCount := 0

	for _, target := range targets {
		if target == nil || target.BattleStats.IsFainted() {
			continue
		}

		// Skip allies if not included (for moves like Earthquake)
		if !includeAllies && target.OwnerID == attacker.OwnerID {
			continue
		}

		// Execute move on each target (reduced power for multi-target moves)
		modifiedMove := move
		if len(targets) > 1 {
			modifiedMove.Power = int(float64(move.Power) * 0.75) // 75% power for multi-target
		}

		err := b.executeSingleTargetMove(attacker, target, modifiedMove)
		if err == nil {
			hitCount++
		}
	}

	if hitCount == 0 {
		b.AddToLog("But it failed!")
	}

	return nil
}

func (b *Battle) executeSelfTargetMove(attacker *Character, move Move) error {
	b.applyMoveEffects(attacker, attacker, move, 0)
	return nil
}

func (b *Battle) applyMoveEffects(attacker, target *Character, move Move, damage int) {
	// Apply primary effect
	if move.Effect != nil {
		b.applySingleEffect(attacker, target, move.Effect, 100, damage)
	}

	// Apply secondary effect (chance-based)
	if move.SecondaryEffect != nil {
		chance := 100 // Default chance if not specified
		if move.SecondaryEffect.StatusChance > 0 {
			chance = move.SecondaryEffect.StatusChance
		} else if move.SecondaryEffect.StatChance > 0 {
			chance = move.SecondaryEffect.StatChance
		} else if move.SecondaryEffect.FlinchChance > 0 {
			chance = move.SecondaryEffect.FlinchChance
		}

		if rand.Intn(100) < chance {
			b.applySingleEffect(attacker, target, move.SecondaryEffect, chance, damage)
		}
	}
}

func (b *Battle) applySingleEffect(attacker, target *Character, effect *EffectType, chance int, damage int) {
	// Status conditions
	if effect.StatusCondition != constants.StatusNone && effect.StatusChance > 0 {
		if rand.Intn(100) < effect.StatusChance {
			duration := 0
			switch effect.StatusCondition {
			case constants.StatusSleep:
				duration = rand.Intn(3) + 1 // 1-3 turns
			case constants.StatusConfuse:
				duration = rand.Intn(4) + 1 // 1-4 turns
			}

			if target.BattleStats.AddStatusEffect(effect.StatusCondition, duration) {
				b.AddToLog(fmt.Sprintf("%s was %s!", target.CharacterName(), effect.StatusCondition))
			} else {
				b.AddToLog(fmt.Sprintf("%s is already affected by a status condition!", target.CharacterName()))
			}
		}
	}

	// Stat modifications (target)
	if len(effect.StatModifiers) > 0 && effect.StatChance > 0 {
		if rand.Intn(100) < effect.StatChance {
			for stat, stages := range effect.StatModifiers {
				oldStage := b.getStatStage(target, stat)
				target.BattleStats.ModifyStat(stat, stages)
				newStage := b.getStatStage(target, stat)

				if newStage != oldStage {
					if stages > 0 {
						b.AddToLog(fmt.Sprintf("%s's %s rose!", target.CharacterName(), stat))
					} else {
						b.AddToLog(fmt.Sprintf("%s's %s fell!", target.CharacterName(), stat))
					}
				} else {
					if stages > 0 {
						b.AddToLog(fmt.Sprintf("%s's %s won't go higher!", target.CharacterName(), stat))
					} else {
						b.AddToLog(fmt.Sprintf("%s's %s won't go lower!", target.CharacterName(), stat))
					}
				}
			}
		}
	}

	// Self stat modifications
	if len(effect.SelfStatModifiers) > 0 && effect.SelfStatChance > 0 {
		if rand.Intn(100) < effect.SelfStatChance {
			for stat, stages := range effect.SelfStatModifiers {
				oldStage := b.getStatStage(attacker, stat)
				attacker.BattleStats.ModifyStat(stat, stages)
				newStage := b.getStatStage(attacker, stat)

				if newStage != oldStage {
					if stages > 0 {
						b.AddToLog(fmt.Sprintf("%s's %s rose!", attacker.CharacterName(), stat))
					} else {
						b.AddToLog(fmt.Sprintf("%s's %s fell!", attacker.CharacterName(), stat))
					}
				}
			}
		}
	}

	// Flinch
	if effect.Flinch && effect.FlinchChance > 0 {
		if rand.Intn(100) < effect.FlinchChance {
			target.BattleStats.FlinchThisTurn = true
			b.AddToLog(fmt.Sprintf("%s flinched!", target.CharacterName()))
		}
	}

	// Healing
	if effect.HealPercentage > 0 {
		healAmount := (attacker.BattleStats.MaxHP * effect.HealPercentage) / 100
		oldHP := attacker.BattleStats.CurrentHP
		attacker.BattleStats.Heal(healAmount)
		actualHeal := attacker.BattleStats.CurrentHP - oldHP

		if actualHeal > 0 {
			b.AddToLog(fmt.Sprintf("%s restored %d HP!", attacker.CharacterName(), actualHeal))
		} else {
			b.AddToLog(fmt.Sprintf("%s's HP is already full!", attacker.CharacterName()))
		}
	}

	if effect.HealFixed > 0 {
		oldHP := attacker.BattleStats.CurrentHP
		attacker.BattleStats.Heal(effect.HealFixed)
		actualHeal := attacker.BattleStats.CurrentHP - oldHP

		if actualHeal > 0 {
			b.AddToLog(fmt.Sprintf("%s restored %d HP!", attacker.CharacterName(), actualHeal))
		}
	}

	// Protection
	if effect.ProtectsUser {
		attacker.BattleStats.ProtectedTurns = 1
		b.AddToLog(fmt.Sprintf("%s protected itself!", attacker.CharacterName()))
	}

	// Recharge requirement
	if effect.RequiresRecharge {
		attacker.BattleStats.MustRecharge = true
	}

	// Self-destruct
	if effect.SelfDestruct {
		attacker.BattleStats.TakeDamage(attacker.BattleStats.CurrentHP)
		b.AddToLog(fmt.Sprintf("%s fainted!", attacker.CharacterName()))
	}
}

func (b *Battle) getStatStage(char *Character, stat string) int {
	switch stat {
	case "atk":
		return char.BattleStats.AtkStage
	case "def":
		return char.BattleStats.DefStage
	case "satk":
		return char.BattleStats.SpAtkStage
	case "sdef":
		return char.BattleStats.SpDefStage
	case "spe":
		return char.BattleStats.SpeStage
	case "acc":
		return char.BattleStats.AccStage
	case "eva":
		return char.BattleStats.EvaStage
	default:
		return 0
	}
}

func (b *Battle) executeAction(action PlayerAction) error {
	player := b.GetPlayer(action.PlayerID)
	if player == nil {
		return fmt.Errorf("player not found")
	}

	switch action.Action {
	case ActionAttack:
		return b.executeAttack(player, action)
	case ActionSwitch:
		return b.executeSwitch(player, action)
	case ActionSkip:
		b.AddToLog(fmt.Sprintf("%s skipped their turn", player.ID))
		return nil
	default:
		return fmt.Errorf("unknown action type")
	}
}

func (b *Battle) executeSwitch(player *BattlePlayer, action PlayerAction) error {
	if action.SwitchTo < 0 || action.SwitchTo >= len(player.Team) {
		return fmt.Errorf("invalid switch target")
	}

	newChar := player.Team[action.SwitchTo]
	if newChar.BattleStats.IsFainted() {
		return fmt.Errorf("cannot switch to fainted character")
	}

	if player.ActiveCharacter == action.SwitchTo {
		return fmt.Errorf("character is already active")
	}

	oldChar := player.GetActiveCharacter()
	player.ActiveCharacter = action.SwitchTo

	if oldChar != nil {
		b.AddToLog(fmt.Sprintf("%s recalled %s!", player.ID, oldChar.CharacterName()))
	}
	b.AddToLog(fmt.Sprintf("%s sent out %s!", player.ID, newChar.CharacterName()))

	return nil
}

func (b *Battle) processEndOfTurnEffects() {
	// Process status effects for both players' active characters
	for _, player := range []*BattlePlayer{b.Player1, b.Player2} {
		char := player.GetActiveCharacter()
		if char == nil || char.BattleStats.IsFainted() {
			continue
		}

		b.processStatusEffects(char, player.ID)
		char.BattleStats.ProcessTurnEnd()
	}
}

func (b *Battle) processStatusEffects(char *Character, playerID snowflake.ID) {
	stats := char.BattleStats

	// Poison damage
	if stats.HasStatusEffect(constants.StatusPoison) {
		damage := stats.MaxHP / 8
		stats.TakeDamage(damage)
		b.AddToLog(fmt.Sprintf("%s took %d poison damage!", char.CharacterName(), damage))

		if stats.IsFainted() {
			b.AddToLog(fmt.Sprintf("%s fainted from poison!", char.CharacterName()))
		}
	}

	// Burn damage
	if stats.HasStatusEffect(constants.StatusBurn) {
		damage := stats.MaxHP / 16
		stats.TakeDamage(damage)
		b.AddToLog(fmt.Sprintf("%s took %d burn damage!", char.CharacterName(), damage))

		if stats.IsFainted() {
			b.AddToLog(fmt.Sprintf("%s fainted from burn!", char.CharacterName()))
		}
	}

	// Sleep countdown
	if stats.HasStatusEffect(constants.StatusSleep) {
		if duration, exists := stats.StatusDurations[constants.StatusSleep]; exists && duration <= 1 {
			stats.RemoveStatusEffect(constants.StatusSleep)
			b.AddToLog(fmt.Sprintf("%s woke up!", char.CharacterName()))
		}
	}

	// Freeze chance to thaw
	if stats.HasStatusEffect(constants.StatusFreeze) {
		if rand.Intn(100) < 20 { // 20% chance to thaw each turn
			stats.RemoveStatusEffect(constants.StatusFreeze)
			b.AddToLog(fmt.Sprintf("%s thawed out!", char.CharacterName()))
		}
	}
}

func (b *Battle) checkBattleEnd() bool {
	if !b.Player1.HasAlivePokemon() {
		b.endBattle(b.Player2.ID)
		return true
	}

	if !b.Player2.HasAlivePokemon() {
		b.endBattle(b.Player1.ID)
		return true
	}

	return false
}

func (b *Battle) endBattle(winnerID snowflake.ID) {
	b.State = BattleStateFinished
	b.Winner = &winnerID
	now := time.Now()
	b.FinishedAt = &now

	winner := b.GetPlayer(winnerID)
	loser := b.GetOpponent(winnerID)

	if winner != nil && loser != nil {
		b.AddToLog(fmt.Sprintf("%s wins the battle!", winner.ID))
		b.AddToLog(fmt.Sprintf("Battle lasted %d turns", b.CurrentTurn))
	}
}

func (b *Battle) endBattleDueToTurnLimit() {
	b.State = BattleStateFinished
	now := time.Now()
	b.FinishedAt = &now

	// Determine winner by remaining HP percentage
	p1HP := b.calculateTeamHPPercentage(b.Player1)
	p2HP := b.calculateTeamHPPercentage(b.Player2)

	if p1HP > p2HP {
		b.Winner = &b.Player1.ID
		b.AddToLog(fmt.Sprintf("%s wins by HP percentage! (%0.1f%% vs %0.1f%%)", b.Player1.ID, p1HP, p2HP))
	} else if p2HP > p1HP {
		b.Winner = &b.Player2.ID
		b.AddToLog(fmt.Sprintf("%s wins by HP percentage! (%0.1f%% vs %0.1f%%)", b.Player2.ID, p2HP, p1HP))
	} else {
		b.AddToLog("Battle ended in a draw due to turn limit!")
	}
}

func (b *Battle) calculateTeamHPPercentage(player *BattlePlayer) float64 {
	totalHP := 0
	currentHP := 0

	for _, char := range player.Team {
		totalHP += char.BattleStats.MaxHP
		currentHP += char.BattleStats.CurrentHP
	}

	if totalHP == 0 {
		return 0
	}

	return (float64(currentHP) / float64(totalHP)) * 100
}

func (b *Battle) CanAddAction(playerID snowflake.ID, action PlayerAction) error {
	if b.State != BattleStateInProgress {
		return fmt.Errorf("battle is not in progress")
	}

	player := b.GetPlayer(playerID)
	if player == nil {
		return fmt.Errorf("player not in battle")
	}

	// Check if player already has an action this turn
	if len(player.ActionsThisTurn) > 0 {
		return fmt.Errorf("action already submitted for this turn")
	}

	// Validate action based on type
	switch action.Action {
	case ActionAttack:
		return b.validateAttackAction(player, action)
	case ActionSwitch:
		return b.validateSwitchAction(player, action)
	}

	return nil
}

func (b *Battle) validateAttackAction(player *BattlePlayer, action PlayerAction) error {
	char := player.GetActiveCharacter()
	if char == nil || char.BattleStats.IsFainted() {
		return fmt.Errorf("no active character or character is fainted")
	}

	if !char.BattleStats.CanMove() {
		return fmt.Errorf("character cannot move this turn")
	}

	_, exists := GetMoveByID(action.MoveID)
	if !exists {
		return fmt.Errorf("move not found")
	}

	// Check if character knows this move
	canUse := false
	for _, moveID := range char.Moves {
		if int(moveID) == action.MoveID {
			canUse = true
			break
		}
	}

	if !canUse {
		return fmt.Errorf("character doesn't know this move")
	}

	// Check PP (would need to track move instances)
	// This would require expanding the character model to track individual move PP

	return nil
}

func (b *Battle) validateSwitchAction(player *BattlePlayer, action PlayerAction) error {
	if action.SwitchTo < 0 || action.SwitchTo >= len(player.Team) {
		return fmt.Errorf("invalid switch target index")
	}

	targetChar := player.Team[action.SwitchTo]
	if targetChar.BattleStats.IsFainted() {
		return fmt.Errorf("cannot switch to fainted character")
	}

	if player.ActiveCharacter == action.SwitchTo {
		return fmt.Errorf("character is already active")
	}

	return nil
}

func (b *Battle) AddAction(playerID snowflake.ID, action PlayerAction) error {
	if err := b.CanAddAction(playerID, action); err != nil {
		return err
	}

	player := b.GetPlayer(playerID)
	player.ActionsThisTurn = append(player.ActionsThisTurn, action)

	return nil
}

func (b *Battle) BothPlayersHaveActions() bool {
	return len(b.Player1.ActionsThisTurn) > 0 && len(b.Player2.ActionsThisTurn) > 0
}

func (b *Battle) GetBattleSummary() string {
	summary := fmt.Sprintf("Battle ID: %s\n", b.ID.String()[:8])
	summary += fmt.Sprintf("Turn: %d/%d\n", b.CurrentTurn, b.Settings.MaxTurns)
	summary += fmt.Sprintf("State: %s\n", b.State.String())

	if b.Player1 != nil {
		char1 := b.Player1.GetActiveCharacter()
		if char1 != nil {
			summary += fmt.Sprintf("Player 1: %s (HP: %d/%d)\n",
				char1.CharacterName(),
				char1.BattleStats.CurrentHP,
				char1.BattleStats.MaxHP)
		}
	}

	if b.Player2 != nil {
		char2 := b.Player2.GetActiveCharacter()
		if char2 != nil {
			summary += fmt.Sprintf("Player 2: %s (HP: %d/%d)\n",
				char2.CharacterName(),
				char2.BattleStats.CurrentHP,
				char2.BattleStats.MaxHP)
		}
	}

	return summary
}
