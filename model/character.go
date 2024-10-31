package model

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/theoreotm/gordinal/constants"
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

	CharacterID int                   // The unique ID of the character
	Level       int                   // The level of the character
	Xp          int                   // The current xp of the character
	Personality constants.Personality // The personality of the character
	Shiny       bool                  // Whether the character is shiny or not

	IvHP  int // The IV of the character's HP
	IvAtk int // The IV of the character's Attack
	IvDef int // The IV of the character's Defense
	// IvSpAtk int // The IV of the character's Sp. Attack - DISABLED
	// IvSpDef int // The IV of the character's Sp. Defense - DISABLED
	IvSpd int // The IV of the character's Speed

	IvTotal int // The total IV of the character

	Nickname  string // The nickname of the character
	Favourite bool   // Whether the character is a favourite or not
	HeldItem  int    // The held item of the character
	Moves     []int  // The moves of the character TODO: Type this field
	Color     int32

	BattleStats *BattleStats
}

func RandomPersonality() constants.Personality {
	return constants.Personalities[rand.Intn(len(constants.Personalities))]
}

func (c *Character) Random() {
	ivs := make([]int, 6)
	for i := range ivs {
		ivs[i] = rand.Intn(31) + 1
	}

	c.IvHP = ivs[0]
	c.IvAtk = ivs[1]
	c.IvDef = ivs[2]
	// c.IvSpAtk = ivs[3]
	// c.IvSpDef = ivs[4]
	c.IvSpd = ivs[5]
	c.IvTotal = ivs[0] + ivs[1] + ivs[2] + ivs[5]

	c.Personality = RandomPersonality()

	c.Shiny = rand.Intn(1028) == 1
}

func (c *Character) Species() string {
	return "placeholder"
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

func (c Character) String() string {
	output := ""
	if c.Shiny {
		output += "✨ "
	}
	output += fmt.Sprintf("Level %d ", c.Level)
	output += c.Species()
	if c.Nickname != "" {
		output += fmt.Sprintf(" \"%s\"", c.Nickname)
	}
	if c.Favourite {
		output += " ❤️"
	}
	return output
}

func (c Character) Format(spec string) string {
	var output string

	if c.Shiny {
		output += "✨ "
	}

	if contains(spec, 'l') {
		output += fmt.Sprintf("Level %d ", c.Level)
	} else if contains(spec, 'L') {
		output += fmt.Sprintf("L%d ", c.Level)
	}

	if contains(spec, 'p') {
		output += fmt.Sprintf("%.2f%% ", c.IvPercentage())
	}

	if contains(spec, 'i') && c.Sprite() != "" {
		output += fmt.Sprintf("%s ", c.Sprite())
	}

	output += c.Species() // Assume Species() returns a string

	if contains(spec, 'n') && c.Nickname != "" {
		output += fmt.Sprintf(" \"%s\"", c.Nickname)
	}

	if contains(spec, 'f') && c.Favourite {
		output += " ❤️"
	}

	return output
}

// contains checks if a rune exists in the spec string.
func contains(spec string, flag rune) bool {
	for _, ch := range spec {
		if ch == flag {
			return true
		}
	}
	return false
}
func (c *Character) IvPercentage() float32 {
	ivPercentage := c.IvHP/31 + c.IvAtk/31 + c.IvDef/31 + c.IvSpd/31
	return float32(ivPercentage) / 4
}

func (c *Character) Sprite() string {
	return "sprite"
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
