package entities

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/constants"
)

type Character struct {
	ID               uuid.UUID // Database ID
	OwnerID          string    // Snowflake ID of the owner
	ClaimedTimestamp time.Time // Timestamp when the character was claimed
	IDX              int       // Index of the character in the list

	CharacterID int                   // The associated with entities.Character
	Level       int                   // The level of the character
	XP          int                   // The current xp of the character
	Personality constants.Personality // The personality of the character
	Shiny       bool                  // Whether the character is shiny or not

	IvHP    int // The IV of the character's HP
	IvAtk   int // The IV of the character's Attack
	IvDef   int // The IV of the character's Defense
	IvSpAtk int // The IV of the character's Sp. Attack
	IvSpDef int // The IV of the character's Sp. Defense
	IvSpd   int // The IV of the character's Speed

	IvTotal float64 // The total IV of the character

	Nickname  string  // The nickname of the character
	Favourite bool    // Whether the character is a favourite or not
	HeldItem  int     // The held item of the character
	Moves     []int32 // The moves of the character TODO: Type this field
	Color     int32

	// Battle relevent fields
	BattleStats *BattleStats // The battle stats of the character
	ActiveMoves []int        // The active moves of the character
	IsInBattle  bool         // Whether the character is in a battle or not
}

func RandomPersonality() constants.Personality {
	return constants.Personalities[rand.Intn(len(constants.Personalities))]
}

func NewCharacter(ownerID string) *Character {
	c := &Character{}
	c.OwnerID = ownerID
	c.ClaimedTimestamp = time.Now()
	c.Level = 1
	c.HeldItem = -1

	c.Randomize()
	return c
}

// RandomCharacterSpawn returns an unclaimed character
func RandomCharacterSpawn() *Character {
	c := &Character{}
	c.Randomize()

	return c
}

// Returns the embed image for the character
func (c *Character) Image() (*discord.File, error) {
	loadImage := func(filePath string) (io.Reader, error) {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		return file, nil
	}

	assetsDir := os.Getenv("ASSETS_DIR")
	if assetsDir == "" {
		assetsDir = "./assets" //  Default if not set (for local runs?)
	}
	fmt.Print(assetsDir)
	filePath := fmt.Sprintf("%s/characters/%v.png", assetsDir, c.CharacterID)
	loa, err := loadImage(filePath)
	if err != nil {
		return nil, err
	}

	return discord.NewFile("character.png", "", loa), nil
}

func (c *Character) Randomize() {
	randomId := randomInt(1, len(Characters))

	c.CharacterID = Characters[randomId].ID
	ivs := make([]int, 6)
	for i := range ivs {
		ivs[i] = randomInt(1, 31)
	}

	c.IvHP = ivs[0]
	c.IvAtk = ivs[1]
	c.IvDef = ivs[2]
	c.IvSpAtk = ivs[3]
	c.IvSpDef = ivs[4]
	c.IvSpd = ivs[5]
	c.IvTotal = float64(ivs[0] + ivs[1] + ivs[2] + ivs[3] + ivs[4] + ivs[5])

	c.Personality = RandomPersonality()
	c.Level = int(math.Min(math.Max(float64(int(normalRandom(20, 10))), 1), 100))
	c.Shiny = rand.Intn(1028-1) == 1
}

func (c *Character) CharacterName() string {
	return c.Data().Name
}

func (c *Character) Data() BaseCharacter {
	for _, character := range Characters {
		if character.ID == c.CharacterID {
			return character
		}
	}

	return BaseCharacter{}
}

func (c *Character) MaxXP() int {
	return 250 + 25*c.Level
}

func (c *Character) MaxHP() int {
	return (2*c.Data().HP+c.IvHP+5)*int(c.Level/100) + c.Level + 10
}

func (c *Character) HP() int {
	if c.BattleStats != nil && c.BattleStats.CurrentHP > 0 {
		return c.BattleStats.CurrentHP
	} else {
		return c.MaxHP()
	}
}

func (c *Character) Atk() int {
	return calcStat(c, "atk")
}

func (c *Character) Def() int {
	return calcStat(c, "def")
}

func (c *Character) SpAtk() int {
	return calcStat(c, "satk")
}

func (c *Character) SpDef() int {
	return calcStat(c, "sdef")
}

func (c *Character) Spd() int {
	return calcStat(c, "spd")
}

func (c Character) String() string {
	output := ""
	if c.Shiny {
		output += "✨ "
	}
	output += fmt.Sprintf("Level %d ", c.Level)
	output += c.CharacterName()
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
		output += c.IvPercentage()
	}

	if contains(spec, 'i') && c.Sprite() != "" {
		output += fmt.Sprintf("%s ", c.Sprite())
	}

	output += c.CharacterName()

	if contains(spec, 'n') && c.Nickname != "" {
		output += fmt.Sprintf(" \"%s\"", c.Nickname)
	}

	if contains(spec, 'f') && c.Favourite {
		output += " ❤️"
	}

	return output
}

// returns the percentage of the character's IVs. Eg: 0.75 for a character with 75% IV
func (c *Character) IvPercentage() string {
	percentage := float64((c.IvTotal / 186) * 100)
	return fmt.Sprintf("%.2f", percentage) + "%"
}

func (c *Character) Sprite() string {
	emoji, ok := CharacterSprites[c.CharacterID]
	if ok {
		return fmt.Sprintf("<:character:%v>", emoji)
	}

	return "❔"
}

func (c *Character) SetHP(hp int) {
	if c.BattleStats == nil {
		c.InitializeBattleStats()
	}
	c.BattleStats.CurrentHP = hp
}

func (c *Character) CanUseMove(moveID int) bool {
	return c.ActiveMoves[moveID] > 0
}

func (c *Character) UseMove(moveID int) error {
	if c.CanUseMove(moveID) {
		c.ActiveMoves[moveID]--
		return nil
	}
	return fmt.Errorf("Move %d is not usable", moveID)
}

func (c *Character) CalculateTurnPriority() {
	baseSpeed := c.Spd()
	c.BattleStats.TurnPriority = baseSpeed + rand.Intn(10) // Add random factor
}

func (c *Character) InitializeBattleStats() {
	c.BattleStats = &BattleStats{}
}

func (c *Character) ResetAfterBattle() {
	c.BattleStats = &BattleStats{
		CurrentHP: c.MaxHP(),
	}
	c.IsInBattle = false
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

func calcStat(character *Character, stat string) int {
	base := character.Data() // Fetch the base stats of the character

	var iv int
	var baseStat int

	// Set IVs and base stats based on the requested stat
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

	// Calculate the stat with the formula and apply personality multiplier
	calculated := math.Floor((float64(((2*baseStat+iv+5)*character.Level)/100 + 5)) * getPersonalityMultiplier(character.Personality, stat))

	return int(calculated)
}

// func calcPower(level int) int {
// 	return level/100 + 5
// }

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

func randomInt(min, max int) int {
	return rand.Intn(max-min) + min
}

func normalRandom(mean, stddev float64) float64 {
	return rand.NormFloat64()*stddev + mean
}
