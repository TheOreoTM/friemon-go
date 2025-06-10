package game

import (
	"math/rand"

	"github.com/theoreotm/friemon/constants"
)

type BattleStats struct {
	CurrentHP int
	MaxHP     int

	// Stat stages (-6 to +6)
	AtkStage   int
	DefStage   int
	SpAtkStage int
	SpDefStage int
	SpeStage   int
	AccStage   int // Accuracy stage
	EvaStage   int // Evasion stage

	// Status effects
	StatusEffects   []constants.StatusEffect
	StatusDurations map[constants.StatusEffect]int // Turns remaining

	// Battle state
	TurnPriority     int
	ProtectedTurns   int   // Turns protected (Protect, etc.)
	ChargingMove     *Move // Move being charged
	TrappedBy        *Move // Move that's trapping this character
	TrappedTurns     int
	ConfusedTurns    int
	FlinchThisTurn   bool
	MustRecharge     bool
	SemiInvulnerable bool

	// Temporary modifiers for this battle
	TempModifiers map[string]int
	CritBoost     int              // Temporary crit rate boost
	TypeBoosts    map[Type]float64 // Temporary type effectiveness boosts

	// Multi-turn moves
	MultiTurnMove   *Move
	MultiTurnLeft   int
	MultiTurnDamage int // Stored damage for moves like Fury Swipes

	// Move usage tracking
	LastMoveUsed  *Move
	MovesUsed     []Move // History of moves used this battle
	Disabled      bool   // Can't use moves
	DisabledTurns int
}

func NewBattleStats(maxHP int) *BattleStats {
	return &BattleStats{
		CurrentHP:       maxHP,
		MaxHP:           maxHP,
		StatusDurations: make(map[constants.StatusEffect]int),
		TempModifiers:   make(map[string]int),
		TypeBoosts:      make(map[Type]float64),
		MovesUsed:       make([]Move, 0),
	}
}

func (b *BattleStats) HasStatusEffect(effect constants.StatusEffect) bool {
	for _, status := range b.StatusEffects {
		if status == effect {
			return true
		}
	}
	return false
}

func (b *BattleStats) AddStatusEffect(effect constants.StatusEffect, duration int) bool {
	// Can't apply major status if already has one (except None)
	if b.HasMajorStatus() && isMajorStatus(effect) {
		return false
	}

	if !b.HasStatusEffect(effect) {
		b.StatusEffects = append(b.StatusEffects, effect)
	}

	if duration > 0 {
		b.StatusDurations[effect] = duration
	}

	return true
}

func (b *BattleStats) RemoveStatusEffect(effect constants.StatusEffect) {
	for i, status := range b.StatusEffects {
		if status == effect {
			b.StatusEffects = append(b.StatusEffects[:i], b.StatusEffects[i+1:]...)
			break
		}
	}
	delete(b.StatusDurations, effect)
}

func (b *BattleStats) HasMajorStatus() bool {
	majorStatuses := []constants.StatusEffect{
		constants.StatusPoison,
		constants.StatusBurn,
		constants.StatusParalyze,
		constants.StatusSleep,
		constants.StatusFreeze,
	}

	for _, status := range majorStatuses {
		if b.HasStatusEffect(status) {
			return true
		}
	}
	return false
}

func isMajorStatus(effect constants.StatusEffect) bool {
	majorStatuses := []constants.StatusEffect{
		constants.StatusPoison,
		constants.StatusBurn,
		constants.StatusParalyze,
		constants.StatusSleep,
		constants.StatusFreeze,
	}

	for _, status := range majorStatuses {
		if effect == status {
			return true
		}
	}
	return false
}

func (b *BattleStats) ModifyStat(stat string, stages int) {
	switch stat {
	case "atk":
		b.AtkStage = clampStage(b.AtkStage + stages)
	case "def":
		b.DefStage = clampStage(b.DefStage + stages)
	case "satk":
		b.SpAtkStage = clampStage(b.SpAtkStage + stages)
	case "sdef":
		b.SpDefStage = clampStage(b.SpDefStage + stages)
	case "spe":
		b.SpeStage = clampStage(b.SpeStage + stages)
	case "acc":
		b.AccStage = clampStage(b.AccStage + stages)
	case "eva":
		b.EvaStage = clampStage(b.EvaStage + stages)
	}
}

func clampStage(stage int) int {
	if stage > 6 {
		return 6
	}
	if stage < -6 {
		return -6
	}
	return stage
}

func (b *BattleStats) GetStatMultiplier(stat string) float64 {
	var stage int
	switch stat {
	case "atk":
		stage = b.AtkStage
	case "def":
		stage = b.DefStage
	case "satk":
		stage = b.SpAtkStage
	case "sdef":
		stage = b.SpDefStage
	case "spe":
		stage = b.SpeStage
	case "acc":
		stage = b.AccStage
	case "eva":
		stage = b.EvaStage
	default:
		return 1.0
	}

	// Pokemon-style stat stage multipliers
	if stage >= 0 {
		return float64(2+stage) / 2.0
	} else {
		return 2.0 / float64(2+(-stage))
	}
}

func (b *BattleStats) ProcessTurnEnd() {
	// Process status effect durations
	for effect, duration := range b.StatusDurations {
		if duration > 0 {
			b.StatusDurations[effect] = duration - 1
			if b.StatusDurations[effect] <= 0 {
				b.RemoveStatusEffect(effect)
			}
		}
	}

	// Reset turn-based flags
	b.FlinchThisTurn = false
	b.ProtectedTurns = max(0, b.ProtectedTurns-1)
	b.TrappedTurns = max(0, b.TrappedTurns-1)
	b.ConfusedTurns = max(0, b.ConfusedTurns-1)
	b.DisabledTurns = max(0, b.DisabledTurns-1)

	if b.TrappedTurns <= 0 {
		b.TrappedBy = nil
	}
	if b.ConfusedTurns <= 0 {
		b.RemoveStatusEffect(constants.StatusConfuse)
	}
	if b.DisabledTurns <= 0 {
		b.Disabled = false
	}

	// Handle multi-turn moves
	if b.MultiTurnMove != nil {
		b.MultiTurnLeft--
		if b.MultiTurnLeft <= 0 {
			b.MultiTurnMove = nil
		}
	}

	// Reset recharge
	b.MustRecharge = false
}

func (b *BattleStats) CanMove() bool {
	// Check if character can move this turn
	if b.MustRecharge {
		return false
	}

	if b.Disabled {
		return false
	}

	if b.HasStatusEffect(constants.StatusSleep) || b.HasStatusEffect(constants.StatusFreeze) {
		return false
	}

	if b.HasStatusEffect(constants.StatusParalyze) {
		// 25% chance to be paralyzed
		return rand.Intn(100) >= 25
	}

	return true
}

func (b *BattleStats) TakeDamage(damage int) {
	b.CurrentHP -= damage
	if b.CurrentHP < 0 {
		b.CurrentHP = 0
	}
}

func (b *BattleStats) Heal(amount int) {
	b.CurrentHP += amount
	if b.CurrentHP > b.MaxHP {
		b.CurrentHP = b.MaxHP
	}
}

func (b *BattleStats) IsFainted() bool {
	return b.CurrentHP <= 0
}

func (b *BattleStats) Reset() {
	b.CurrentHP = b.MaxHP
	b.StatusEffects = make([]constants.StatusEffect, 0)
	b.StatusDurations = make(map[constants.StatusEffect]int)
	b.AtkStage = 0
	b.DefStage = 0
	b.SpAtkStage = 0
	b.SpDefStage = 0
	b.SpeStage = 0
	b.AccStage = 0
	b.EvaStage = 0
	b.ProtectedTurns = 0
	b.ChargingMove = nil
	b.TrappedBy = nil
	b.TrappedTurns = 0
	b.ConfusedTurns = 0
	b.FlinchThisTurn = false
	b.MustRecharge = false
	b.SemiInvulnerable = false
	b.TempModifiers = make(map[string]int)
	b.CritBoost = 0
	b.TypeBoosts = make(map[Type]float64)
	b.MultiTurnMove = nil
	b.MultiTurnLeft = 0
	b.MultiTurnDamage = 0
	b.LastMoveUsed = nil
	b.MovesUsed = make([]Move, 0)
	b.Disabled = false
	b.DisabledTurns = 0
}
