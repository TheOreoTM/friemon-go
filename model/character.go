package model

import (
	"math/rand"
	"time"
)

type Personality int

const (
	PersonalityPensive Personality = iota // Frieren
	PersonalityStoic                      // Himmel
	PersonalityMerry                      // Heiter
	PersonalityResolute
	PersonalitySkeptical
	PersonalityBrooding
	PersonalityBrave
	PersonalityInsightful
	PersonalityPlayful
	PersonalityRash // Stark
)

type BattleStats struct {
	CurrentHP int
	AtkMod    int
	DefMod    int
	SpAtkMod  int
	SpDefMod  int
	SpeMod    int
}

type Character struct {
	ID               string    // Database ID
	OwnerID          string    // Snowflake ID of the owner
	ClaimedTimestamp time.Time // Timestamp when the character was claimed
	IDX              int       // Index of the character in the list

	CharacterID int         // The unique ID of the character
	Level       int         // The level of the character
	Xp          int         // The current xp of the character
	Personality Personality // The personality of the character
	Shiny       bool        // Whether the character is shiny or not

	IvHP    int // The IV of the character's HP
	IvAtk   int // The IV of the character's Attack
	IvDef   int // The IV of the character's Defense
	IvSpAtk int // The IV of the character's Sp. Attack
	IvSpDef int // The IV of the character's Sp. Defense
	IvSpe   int // The IV of the character's Speed

	IvTotal int // The total IV of the character

	Nickname  string // The nickname of the character
	Favourite bool   // Whether the character is a favourite or not
	HeldItem  int    // The held item of the character
	Moves     []int  // The moves of the character TODO: Type this field
	Color     int32

	BattleStats *BattleStats
}

// String method to get the string representation of each Personality
func (p Personality) String() string {
	return [...]string{
		"Pensive",
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

func RandomPersonality() Personality {
	return Personality(rand.Intn(int(PersonalityRash)))
}

func (c *Character) Random() {
	ivs := make([]int, 6)
	for i := range ivs {
		ivs[i] = rand.Intn(31) + 1
	}

	c.IvHP = ivs[0]
	c.IvAtk = ivs[1]
	c.IvDef = ivs[2]
	c.IvSpAtk = ivs[3]
	c.IvSpDef = ivs[4]
	c.IvSpe = ivs[5]
	c.IvTotal = ivs[0] + ivs[1] + ivs[2] + ivs[3] + ivs[4] + ivs[5]

	c.Personality = RandomPersonality()

	c.Shiny = rand.Intn(1028) == 1
}

func (c *Character) Species() {

}

func (c *Character) MaxXP() int {
	return 250 + 25*c.Level
}

func (c *Character) MaxHP() int {
	return (2*45 + c.IvHP + 5) * c.Level // TODO: Change 45 to base hp stat
}

func (c *Character) HP() int {
	if c.BattleStats != nil && c.BattleStats.CurrentHP > 0 {
		return c.BattleStats.CurrentHP
	} else {
		return c.MaxHP()
	}
}

func (c *Character) SetHP(hp int) {
	if c.BattleStats == nil {
		c.InitializeBattleStats()
	}
	c.BattleStats.CurrentHP = hp
}

func (c *Character) InitializeBattleStats() {
	c.BattleStats = &BattleStats{}
}

func (c *Character) ResetBattleStats() {
	c.BattleStats = nil
}
