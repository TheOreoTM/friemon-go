package game

import "github.com/TheOreoTM/friemon-go/internal/data"

// EventKind says what happened in a battle. The Discord layer turns events into text,
// and tests check them directly, so the engine never formats messages itself.
type EventKind int

const (
	EvSwitchIn  EventKind = iota // Side sent out Name
	EvSwitchOut                  // Side withdrew Name
	EvMove                       // Name used Move
	EvStruggle                   // Name has no PP left, so it struggles
	EvMiss
	EvFail     // the move did nothing
	EvNoTarget // the target had already fainted
	EvCrit
	EvEffectiveness // Mult is the type multiplier
	EvDamage        // Name lost Amount HP
	EvHits          // the move hit Amount times
	EvHeal          // Name recovered Amount HP
	EvDrain         // Name drained Amount HP
	EvRecoil        // Name lost Amount HP to recoil
	EvFaint
	EvCondition       // Name got Condition
	EvConditionFailed // Name already has a condition
	EvImmune          // Name's type is immune to Condition
	EvConditionDamage // Name lost Amount HP to its Condition
	EvCured           // Name's Condition was cured
	EvFastAsleep
	EvWokeUp
	EvFrozenSolid
	EvThawed
	EvFullyBound // Bound and could not move
	EvRested     // fell asleep and became healthy
	EvConfused
	EvAlreadyConfused
	EvIsConfused
	EvConfusionHurt // hit itself for Amount HP
	EvSnappedOut
	EvFlinched
	EvStatChange // Stat tried to change by Stages; Amount is how much it actually changed
	EvWin        // Side won
)

// Event is one thing that happened. Only the fields that matter for its Kind are set.
type Event struct {
	Kind      EventKind
	Side      int    // the side this is about: 0 or 1
	Name      string // the character this is about
	Move      string
	Amount    int
	Mult      float64
	Condition data.Condition
	Stat      data.Stat
	Stages    int
}
