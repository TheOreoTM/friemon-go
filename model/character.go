package model

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/theoreotm/friemon/constants"
	data "github.com/theoreotm/friemon/data/character"
)

type BattleStats struct {
	CurrentHP int32
	AtkMod    int32
	DefMod    int32
	SpAtkMod  int32
	SpDefMod  int32
	SpeMod    int32
}

type Character struct {
	ID               string    // Database ID
	OwnerID          string    // Snowflake ID of the owner
	ClaimedTimestamp time.Time // Timestamp when the character was claimed
	IDX              int32     // Index of the character in the list

	CharacterID int32                 // The unique ID of the character
	Level       int32                 // The level of the character
	Xp          int32                 // The current xp of the character
	Personality constants.Personality // The personality of the character
	Shiny       bool                  // Whether the character is shiny or not

	IvHP    int32 // The IV of the character's HP
	IvAtk   int32 // The IV of the character's Attack
	IvDef   int32 // The IV of the character's Defense
	IvSpAtk int32 // The IV of the character's Sp. Attack - NOT_USED
	IvSpDef int32 // The IV of the character's Sp. Defense - NOT_USED
	IvSpd   int32 // The IV of the character's Speed

	IvTotal float64 // The total IV of the character

	Nickname  string  // The nickname of the character
	Favourite bool    // Whether the character is a favourite or not
	HeldItem  int32   // The held item of the character
	Moves     []int32 // The moves of the character TODO: Type this field
	Color     int32

	BattleStats *BattleStats
}

func RandomPersonality() constants.Personality {
	return constants.Personalities[rand.Intn(len(constants.Personalities))]
}

func NewCharacter(ownerID string) *Character {
	c := &Character{}
	c.OwnerID = ownerID
	c.ClaimedTimestamp = time.Now()

	c.Random()
	return c
}

func (c *Character) Random() {
	randomId := (rand.Int31n(int32(len(data.EnabledCharacters()))))

	c.CharacterID = data.EnabledCharacters()[randomId].ID
	ivs := make([]int32, 6)
	for i := range ivs {
		ivs[i] = rand.Int31n(31) + 1
	}

	c.IvHP = ivs[0]
	c.IvAtk = ivs[1]
	c.IvDef = ivs[2]
	c.IvSpAtk = ivs[3]
	c.IvSpDef = ivs[4]
	c.IvSpd = ivs[5]
	c.IvTotal = float64(ivs[0] + ivs[1] + ivs[2] + ivs[3] + ivs[4] + ivs[5])

	c.Personality = RandomPersonality()

	c.Shiny = rand.Intn(1028) == 1
}

func (c *Character) CharacterName() string {
	return c.Data().Name
}

func (c *Character) Data() data.BaseCharacter {
	for _, character := range data.EnabledCharacters() {
		if character.ID == c.CharacterID {
			return character
		}
	}

	return data.BaseCharacter{}
}

func (c *Character) MaxXP() int32 {
	return 250 + 25*c.Level
}

func (c *Character) MaxHP() int32 {
	return (2*c.Data().HP + c.IvHP + 5) * c.Level // TODO: Change 45 to base hp stat
}

func (c *Character) HP() int32 {
	if c.BattleStats != nil && c.BattleStats.CurrentHP > 0 {
		return c.BattleStats.CurrentHP
	} else {
		return c.MaxHP()
	}
}

func (c *Character) Atk() int32 {
	return calcStat(c, "atk")
}

func (c *Character) Def() int32 {
	return calcStat(c, "def")
}

func (c *Character) SpAtk() int32 {
	return calcStat(c, "satk")
}

func (c *Character) SpDef() int32 {
	return calcStat(c, "sdef")
}

func (c *Character) Spd() int32 {
	return calcStat(c, "spd")
}

// func (c Character) String() string {
// 	output := ""
// 	if c.Shiny {
// 		output += "✨ "
// 	}
// 	output += fmt.Sprintf("Level %d ", c.Level)
// 	output += c.CharacterName()
// 	if c.Nickname != "" {
// 		output += fmt.Sprintf(" \"%s\"", c.Nickname)
// 	}
// 	if c.Favourite {
// 		output += " ❤️"
// 	}
// 	return output
// }

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

	output += c.CharacterName() // Assume Species() returns a string

	if contains(spec, 'n') && c.Nickname != "" {
		output += fmt.Sprintf(" \"%s\"", c.Nickname)
	}

	if contains(spec, 'f') && c.Favourite {
		output += " ❤️"
	}

	return output
}

// returns the percentage of the character's IVs. Eg: 0.75 for a character with 75% IV
func (c *Character) IvPercentage() float64 {
	return float64(c.IvTotal / 186)
}

func (c *Character) Sprite() string {
	return "sprite"
}

func (c *Character) SetHP(hp int32) {
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

// contains checks if a rune exists in the spec string.
func contains(spec string, flag rune) bool {
	for _, ch := range spec {
		if ch == flag {
			return true
		}
	}
	return false
}

func calcStat(character *Character, stat string) int32 {
	base := character.Data()

	var iv int32
	var baseStat int32

	switch stat {
	case "atk":
		iv = character.IvAtk
		baseStat = base.Atk
	case "def":
		iv = character.IvDef
		baseStat = base.Def
	case "sdef":
		iv = character.IvSpDef
		baseStat = base.SpDef
	case "satk":
		iv = character.IvSpAtk
		baseStat = base.SpAtk
	case "spd":
		iv = character.IvSpd
		baseStat = base.Spd
	default:
		iv = 0
		baseStat = 0
	}

	calculated := float64((2*baseStat+iv+5)*calcPower(character.Level)) * getPersonalityMultiplier(character.Personality, stat)

	return int32(math.Floor(calculated))
}

func calcPower(level int32) int32 {
	return level/100 + 5
}

// getPersonalityMultiplier returns the multiplier for the given personality and stat.
func getPersonalityMultiplier(p constants.Personality, stat string) float64 {
	// Access the multipliers for the provided personality
	multipliers, exists := constants.PersonalityStatMultipliers[p]
	if !exists {
		return 1.0 // Return a default multiplier of 1 if the personality does not exist
	}

	// Return the multiplier for the specified stat
	multiplier, exists := multipliers[stat]
	if !exists {
		return 1.0 // Return a default multiplier of 1 if the stat does not exist
	}

	return multiplier
}

type createCharacterParams struct {
	OwnerID          string    `json:"owner_id"`
	ClaimedTimestamp time.Time `json:"claimed_timestamp"`
	Idx              int32     `json:"idx"`
	CharacterID      int32     `json:"character_id"`
	Level            int32     `json:"level"`
	Xp               int32     `json:"xp"`
	Personality      string    `json:"personality"`
	Shiny            bool      `json:"shiny"`
	IvHp             int32     `json:"iv_hp"`
	IvAtk            int32     `json:"iv_atk"`
	IvDef            int32     `json:"iv_def"`
	IvSpAtk          int32     `json:"iv_sp_atk"`
	IvSpDef          int32     `json:"iv_sp_def"`
	IvSpd            int32     `json:"iv_spd"`
	IvTotal          float64   `json:"iv_total"`
	Nickname         string    `json:"nickname"`
	Favourite        bool      `json:"favourite"`
	HeldItem         int32     `json:"held_item"`
	Moves            []int32   `json:"moves"`
	Color            int32     `json:"color"`
}
