package data

// Condition is a lasting status condition. A character has at most one, and it stays when switching out.
type Condition int

const (
	NoCondition  Condition = iota
	Burn                   // loses 1/16 HP each turn; physical damage halved
	Bound                  // (Paralysis) Speed halved; 25% chance to be unable to move
	Sleep                  // cannot move for 1-3 turns
	Freeze                 // cannot move; 20% chance to thaw each turn
	Cursed                 // (Poison) loses 1/8 HP each turn
	DeeplyCursed           // (Badly poisoned) loses 1/16 HP, then 2/16, 3/16... each turn
)

func (s Condition) String() string {
	switch s {
	case Burn:
		return "Burn"
	case Bound:
		return "Bound"
	case Sleep:
		return "Sleep"
	case Freeze:
		return "Freeze"
	case Cursed:
		return "Cursed"
	case DeeplyCursed:
		return "Deeply Cursed"
	}
	return "None"
}

// Volatile is a short-lived condition that ends when the character switches out.
type Volatile int

const (
	NoVolatile  Volatile = iota
	Confusion            // 1-4 turns; 33% chance to hit itself instead of moving
	Flinch               // skips its move this turn; only works if it had not moved yet
	Trap                 // cannot switch out while the user stays on the field
	PartialTrap          // cannot switch out and loses 1/8 HP a turn, for 4-5 turns
	Taunt                // can only use damaging moves, for 3 turns
	Encore               // must repeat its last move, for 3 turns
)

// statusImmunities lists types that can never get a status.
var statusImmunities = map[Condition][]Type{
	Cursed:       {TypeCurse}, // demons cannot be cursed
	DeeplyCursed: {TypeCurse},
}

// ImmuneToStatus reports whether a character with these types can never get s.
func ImmuneToStatus(s Condition, types []Type) bool {
	for _, immune := range statusImmunities[s] {
		for _, t := range types {
			if t == immune {
				return true
			}
		}
	}
	return false
}
