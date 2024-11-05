package data

// StatChanges maps a stat to the modifier, e.g., "Attack" +1
type StatChanges map[string]int
type TargetType int
type MoveCategory int

const (
	TargetSingleOpponent TargetType = iota // Targets a single opponent
	TargetAllOpponents                     // Targets all opponents
	TargetUser                             // Targets the user
	TargetSingleAlly                       // Targets a single ally
	TargerAllAllies                        // Targets all allies
	TargetAll                              // Targets everyone (useful for certain area moves)
)

const (
	MoveCatPhysical MoveCategory = iota // Physical moves
	MoveCatSpecial                      // Special moves
	MoveCatStatus                       // Status moves
)

type EffectType struct {
	StatusCondition string      // Status condition inflicted (e.g., "Paralyze", "Burn", "Sleep"), empty if none
	StatModifiers   StatChanges // Changes to stats (e.g., {"Attack": +1}), could be nil if no stat change
	Recoil          int         // Percentage recoil damage, if any (e.g., 25 for moves like "Double-Edge")
	Flinch          bool        // Whether the move has a chance to flinch the target
	DrainPercentage int         // Percentage of damage healed back to the user (e.g., 50 for "Giga Drain")
	ProtectsUser    bool        // If true, the move protects the user (e.g., "Protect", "Detect")
	SelfDestruct    bool        // If true, the move faints the user after use (e.g., "Explosion")
	Chance          int         // Chance for secondary effect, if any (e.g., 10)
}

type Move struct {
	Name            string       // Name of the move (e.g., "Thunderbolt")
	Type            Type         // Type of the move (e.g., "Electric")
	Category        MoveCategory // Physical, Special, or Status
	Power           int          // Power of the move (e.g., 90 for Thunderbolt)
	Accuracy        int          // Accuracy percentage (e.g., 100 for Thunderbolt, or -1 for never-miss moves)
	PP              int          // Power Points, the number of times the move can be used
	CurrentPP       int          // Tracks remaining PP in a battle
	Priority        int          // Priority level (e.g., 1 for Quick Attack)
	Effect          EffectType   // The main effect, if any, (e.g., Burn, Paralysis)
	SecondaryEffect *EffectType  // Any additional effect (e.g., 10% chance to paralyze)
	Target          TargetType   // Determines if it targets an opponent, ally, or all
}
