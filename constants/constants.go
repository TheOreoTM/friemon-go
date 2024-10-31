package constants

type Personality int

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

// String method to get the string representation of each Personality
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
