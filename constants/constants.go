package constants

type Personality int
type Color int

// Define the multipliers for each personality.
var PersonalityStatMultipliers = map[Personality]map[string]float64{
	PersonalityAloof: {
		"hp":   1.0,
		"atk":  1.0,
		"def":  1.0,
		"satk": 1.0,
		"sdef": 1.0,
		"spd":  1.0,
	},
	PersonalityStoic: {
		"hp":   1.0,
		"atk":  1.0,
		"def":  1.0,
		"satk": 1.0,
		"sdef": 1.0,
		"spd":  1.0,
	},
	PersonalityMerry: {
		"hp":   1.0,
		"atk":  1.0,
		"def":  1.0,
		"satk": 1.1,
		"sdef": 1.0,
		"spd":  1.1,
	},
	PersonalityResolute: {
		"hp":   1.0,
		"atk":  1.0,
		"def":  1.0,
		"satk": 1.1,
		"sdef": 1.0,
		"spd":  1.0,
	},
	PersonalitySkeptical: {
		"hp":   1.0,
		"atk":  0.9,
		"def":  1.0,
		"satk": 1.0,
		"sdef": 1.0,
		"spd":  1.0,
	},
	PersonalityBrooding: {
		"hp":   1.05,
		"atk":  1.0,
		"def":  1.0,
		"satk": 1.0,
		"sdef": 1.0,
		"spd":  1.0,
	},
	PersonalityBrave: {
		"hp":   1.0,
		"atk":  1.1,
		"def":  1.0,
		"satk": 1.0,
		"sdef": 1.0,
		"spd":  0.9,
	},
	PersonalityInsightful: {
		"hp":   1.0,
		"atk":  1.0,
		"def":  1.0,
		"satk": 1.1,
		"sdef": 1.0,
		"spd":  1.0,
	},
	PersonalityPlayful: {
		"hp":   1.0,
		"atk":  1.0,
		"def":  1.0,
		"satk": 1.0,
		"sdef": 1.0,
		"spd":  1.1,
	},
	PersonalityRash: {
		"hp":   1.0,
		"atk":  1.1,
		"def":  1.0,
		"satk": 1.0,
		"sdef": 0.9,
		"spd":  1.0,
	},
}

// String method to get the string representation of each Personality.
func (p Personality) String() string {
	return [...]string{
		"Aloof",
		"Stoic",
		"Merry",
		"Resolute",
		"Skeptical",
		"Brooding",
		"Brave",
		"Insightful",
		"Playful",
		"Rash",
	}[p]
}

// Personalities is an array of all available personalities.
var Personalities = []Personality{
	PersonalityAloof,
	PersonalityStoic,
	PersonalityMerry,
	PersonalityResolute,
	PersonalitySkeptical,
	PersonalityBrooding,
	PersonalityBrave,
	PersonalityInsightful,
	PersonalityPlayful,
	PersonalityRash,
}

const (
	PersonalityAloof Personality = iota // Frieren
	PersonalityStoic                    // Himmel
	PersonalityMerry                    // Heiter
	PersonalityResolute
	PersonalitySkeptical
	PersonalityBrooding
	PersonalityBrave
	PersonalityInsightful
	PersonalityPlayful
	PersonalityRash // Stark
)

const (
	ColorSuccess int = 0x46b485
	ColorFail    int = 0xf05050
	ColorWarn    int = 0xfee65c
	ColorInfo    int = 0x297bd1
	ColorLoading int = 0x23272a
	ColorDefault int = 0x2b2d31
)

// StatusEffect represents possible status effects for a character.
type StatusEffect string

const (
	StatusNone     StatusEffect = "None"     // No status effect
	StatusPoison   StatusEffect = "Poison"   // Gradual HP loss each turn
	StatusBurn     StatusEffect = "Burn"     // Gradual HP loss + reduced attack
	StatusParalyze StatusEffect = "Paralyze" // Reduced speed + chance to skip turn
	StatusSleep    StatusEffect = "Sleep"    // Cannot act for a few turns
	StatusFreeze   StatusEffect = "Freeze"   // Cannot act until thawed
	StatusConfuse  StatusEffect = "Confuse"  // Chance to hurt self
)
