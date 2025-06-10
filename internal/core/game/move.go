package game

import (
	"fmt"

	"github.com/theoreotm/friemon/constants"
)

type StatChanges map[string]int

type TargetType int
type MoveCategory int

const (
	TargetSingleFoe TargetType = iota
	TargetAllFoes
	TargetAllAdjacentFoes
	TargetUser
	TargetSingleAlly
	TargetAllAllies
	TargetAny
	TargetAllAdjacent
)

const (
	MoveCatPhysical MoveCategory = iota
	MoveCatSpecial
	MoveCatStatus
)

func (mc MoveCategory) String() string {
	switch mc {
	case MoveCatPhysical:
		return "Physical"
	case MoveCatSpecial:
		return "Special"
	case MoveCatStatus:
		return "Status"
	default:
		return "Unknown"
	}
}

type EffectType struct {
	// Status effects
	StatusCondition constants.StatusEffect `json:"status_condition"`
	StatusChance    int                    `json:"status_chance"` // Percentage chance

	// Stat modifications (for target)
	StatModifiers StatChanges `json:"stat_modifiers"`
	StatChance    int         `json:"stat_chance"` // Percentage chance

	// Self stat modifications
	SelfStatModifiers StatChanges `json:"self_stat_modifiers"`
	SelfStatChance    int         `json:"self_stat_chance"`

	// Damage modifications
	Recoil          int  `json:"recoil"`           // Percentage of damage dealt as recoil
	DrainPercentage int  `json:"drain_percentage"` // Percentage of damage healed
	FixedDamage     int  `json:"fixed_damage"`     // Fixed damage amount
	Flinch          bool `json:"flinch"`
	FlinchChance    int  `json:"flinch_chance"`

	// Special effects
	ProtectsUser     bool `json:"protects_user"`
	SelfDestruct     bool `json:"self_destruct"`
	IgnoreDefense    bool `json:"ignore_defense"`
	IgnoreEvasion    bool `json:"ignore_evasion"`
	AlwaysCritical   bool `json:"always_critical"`
	HighCritRatio    bool `json:"high_crit_ratio"`
	NeverMiss        bool `json:"never_miss"`
	ChargeRequired   bool `json:"charge_required"`   // Two-turn move
	SemiInvulnerable bool `json:"semi_invulnerable"` // User becomes semi-invulnerable
	MultipleTurns    int  `json:"multiple_turns"`    // Attack for X turns
	RandomTurns      bool `json:"random_turns"`      // Random 2-5 turns
	TrapTarget       bool `json:"trap_target"`       // Trap the target
	ConfuseAfter     bool `json:"confuse_after"`     // Confuse self after use
	RequiresRecharge bool `json:"requires_recharge"` // Must recharge next turn

	// Healing
	HealPercentage int `json:"heal_percentage"` // Percentage of max HP to heal
	HealFixed      int `json:"heal_fixed"`      // Fixed HP amount to heal

	// Weather/field effects (for future expansion)
	WeatherEffect string `json:"weather_effect,omitempty"`
	FieldEffect   string `json:"field_effect,omitempty"`
}

type Move struct {
	ID              int          `json:"id"`
	Name            string       `json:"name"`
	Type            Type         `json:"type"`
	Category        MoveCategory `json:"category"`
	Power           int          `json:"power"`            // Base power (0 for status moves)
	Accuracy        int          `json:"accuracy"`         // Accuracy percentage
	PP              int          `json:"pp"`               // Power Points
	Priority        int          `json:"priority"`         // Move priority (-7 to +7)
	CritRatio       int          `json:"crit_ratio"`       // Critical hit ratio (1-4)
	Description     string       `json:"description"`      // Move description
	Effect          *EffectType  `json:"effect"`           // Primary effect
	SecondaryEffect *EffectType  `json:"secondary_effect"` // Secondary effect (chance-based)
	Target          TargetType   `json:"target"`           // Who this move targets

	// Move flags
	MakesContact      bool `json:"makes_contact"`       // Physical contact
	AffectedByProtect bool `json:"affected_by_protect"` // Blocked by Protect
	AffectedByMirror  bool `json:"affected_by_mirror"`  // Reflected by Magic Coat
	Copyable          bool `json:"copyable"`            // Can be copied by moves like Mimic
	Punching          bool `json:"punching"`            // Punching move
	Sound             bool `json:"sound"`               // Sound-based move
	Pulse             bool `json:"pulse"`               // Pulse move
	Bite              bool `json:"bite"`                // Biting move
	Ballistic         bool `json:"ballistic"`           // Ballistic move
}

// MoveInstance represents a move that a character knows with current PP
type MoveInstance struct {
	Move      Move `json:"move"`
	CurrentPP int  `json:"current_pp"`
}

func NewMoveInstance(move Move) MoveInstance {
	return MoveInstance{
		Move:      move,
		CurrentPP: move.PP,
	}
}

func (mi *MoveInstance) CanUse() bool {
	return mi.CurrentPP > 0
}

func (mi *MoveInstance) Use() error {
	if !mi.CanUse() {
		return fmt.Errorf("move %s has no PP remaining", mi.Move.Name)
	}
	mi.CurrentPP--
	return nil
}

func (mi *MoveInstance) RestorePP(amount int) {
	mi.CurrentPP += amount
	if mi.CurrentPP > mi.Move.PP {
		mi.CurrentPP = mi.Move.PP
	}
}

func (mi *MoveInstance) FullRestore() {
	mi.CurrentPP = mi.Move.PP
}

// Move creation helper
type moveCreateParams struct {
	Type              Type
	Category          MoveCategory
	Power             int
	Accuracy          int
	PP                int
	Priority          int
	CritRatio         int
	Description       string
	Effect            *EffectType
	SecondaryEffect   *EffectType
	Target            TargetType
	MakesContact      bool
	AffectedByProtect bool
	AffectedByMirror  bool
	Copyable          bool
	Punching          bool
	Sound             bool
	Pulse             bool
	Bite              bool
	Ballistic         bool
}

func NewMove(id int, name string, params moveCreateParams) Move {
	return Move{
		ID:                id,
		Name:              name,
		Type:              params.Type,
		Category:          params.Category,
		Power:             params.Power,
		Accuracy:          params.Accuracy,
		PP:                params.PP,
		Priority:          params.Priority,
		CritRatio:         max(1, params.CritRatio), // Default crit ratio is 1
		Description:       params.Description,
		Effect:            params.Effect,
		SecondaryEffect:   params.SecondaryEffect,
		Target:            params.Target,
		MakesContact:      params.MakesContact,
		AffectedByProtect: params.AffectedByProtect,
		AffectedByMirror:  params.AffectedByMirror,
		Copyable:          params.Copyable,
		Punching:          params.Punching,
		Sound:             params.Sound,
		Pulse:             params.Pulse,
		Bite:              params.Bite,
		Ballistic:         params.Ballistic,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
