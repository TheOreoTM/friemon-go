// Package data holds static game data: characters, moves, relics and items.
// It is plain Go values so the compiler checks it and tests can validate it.
package data

// Rarity is a character's spawn tier.
type Rarity int

// iota counts up from 0 inside a const block, so Common=0, Rare=1, Legendary=2.
// This is Go's usual way to write an enum.
const (
	Common Rarity = iota
	Rare
	Legendary
)

// String makes Rarity print as a word. Any type with a String() method
// satisfies fmt.Stringer, so fmt.Println and %v use it automatically.
func (r Rarity) String() string {
	switch r {
	case Common:
		return "Common"
	case Rare:
		return "Rare"
	case Legendary:
		return "Legendary"
	}
	return "Unknown"
}

// Character is a catchable character species. Each caught copy is stored
// separately in the database and points back here by ID.
type Character struct {
	ID          string // stable key saved in the database; never change it once players own copies
	Name        string
	Aliases     []string // other names accepted by catch
	Rarity      Rarity
	Types       []Type // one or two
	BaseStats   Stats
	Learnset    []LearnsetEntry
	Description string
	Emoji       string // shown in lists; a Unicode emoji or a custom one like "<:frieren:123>"
}

// LearnsetEntry says a character can learn a move once it reaches a level.
type LearnsetEntry struct {
	Level  int
	MoveID string
}

// DisplayEmoji returns the character's emoji, or a placeholder if it has none yet.
func (c Character) DisplayEmoji() string {
	if c.Emoji == "" {
		return "❓"
	}
	return c.Emoji
}

// Characters is the full roster. Base stats total roughly 425 for commons,
// 500 for rares and 600 for legendaries.
var Characters = []Character{
	{
		ID:        "frieren",
		Name:      "Frieren",
		Aliases:   []string{"Frieren the Slayer"},
		Rarity:    Legendary,
		Types:     []Type{Destruction, Nature},
		BaseStats: Stats{HP: 80, Attack: 45, Defense: 70, MagicAttack: 150, MagicDefense: 135, Speed: 120},
		Learnset: []LearnsetEntry{
			{1, "zoltraak"}, {1, "flower_field"}, {1, "defensive_magic"},
			{15, "binding_roots"}, {25, "barrier_shatter"}, {35, "lightning_magic"},
			{50, "rapid_barrage"}, {70, "mass_offensive"},
		},
		Description: "An elf mage who has lived for over a thousand years and collects spells as a hobby.",
	},
	{
		ID:        "himmel",
		Name:      "Himmel",
		Aliases:   []string{"Himmel the Hero"},
		Rarity:    Rare,
		Types:     []Type{Martial},
		BaseStats: Stats{HP: 85, Attack: 115, Defense: 90, MagicAttack: 50, MagicDefense: 80, Speed: 110},
		Learnset: []LearnsetEntry{
			{1, "slash"}, {1, "counter_blow"},
			{20, "goddess_light"}, {30, "heroic_strike"}, {45, "flower_field"},
		},
		Description: "The hero who defeated the Demon King, remembered by statues in every village.",
	},
	{
		ID:        "heiter",
		Name:      "Heiter",
		Rarity:    Rare,
		Types:     []Type{Holy},
		BaseStats: Stats{HP: 90, Attack: 40, Defense: 70, MagicAttack: 115, MagicDefense: 125, Speed: 60},
		Learnset: []LearnsetEntry{
			{1, "healing_magic"}, {1, "holy_binding"},
			{15, "defensive_magic"}, {30, "goddess_light"}, {45, "barrier_shatter"},
		},
		Description: "A priest of the hero's party, fond of drink and wiser than he lets on.",
	},
	{
		ID:        "eisen",
		Name:      "Eisen",
		Rarity:    Rare,
		Types:     []Type{Martial},
		BaseStats: Stats{HP: 120, Attack: 115, Defense: 135, MagicAttack: 20, MagicDefense: 80, Speed: 40},
		Learnset: []LearnsetEntry{
			{1, "slash"}, {1, "iron_body"},
			{20, "counter_blow"}, {35, "axe_cleave"},
		},
		Description: "A dwarf warrior of the hero's party who can take almost any blow.",
	},
	{
		ID:        "ubel",
		Name:      "Übel",
		Rarity:    Rare,
		Types:     []Type{Destruction},
		BaseStats: Stats{HP: 70, Attack: 50, Defense: 60, MagicAttack: 130, MagicDefense: 70, Speed: 115},
		Learnset: []LearnsetEntry{
			{1, "zoltraak"}, {1, "cutting_magic"},
			{20, "binding_roots"}, {35, "rapid_barrage"}, {50, "curse_mark"},
		},
		Description: "A mage who can cut anything she imagines she can cut.",
	},
	{
		ID:        "fern",
		Name:      "Fern",
		Rarity:    Common,
		Types:     []Type{Destruction},
		BaseStats: Stats{HP: 65, Attack: 35, Defense: 50, MagicAttack: 110, MagicDefense: 75, Speed: 100},
		Learnset: []LearnsetEntry{
			{1, "zoltraak"}, {1, "defensive_magic"},
			{20, "rapid_barrage"}, {35, "barrier_shatter"}, {55, "mass_offensive"},
		},
		Description: "Frieren's apprentice, a fast and precise caster of basic offensive magic.",
	},
	{
		ID:        "stark",
		Name:      "Stark",
		Rarity:    Common,
		Types:     []Type{Martial},
		BaseStats: Stats{HP: 95, Attack: 110, Defense: 85, MagicAttack: 20, MagicDefense: 55, Speed: 70},
		Learnset: []LearnsetEntry{
			{1, "slash"}, {1, "counter_blow"},
			{20, "iron_body"}, {40, "axe_cleave"},
		},
		Description: "Eisen's former pupil, a warrior far braver in battle than he believes.",
	},
	{
		ID:        "sein",
		Name:      "Sein",
		Rarity:    Common,
		Types:     []Type{Holy},
		BaseStats: Stats{HP: 80, Attack: 45, Defense: 60, MagicAttack: 80, MagicDefense: 95, Speed: 65},
		Learnset: []LearnsetEntry{
			{1, "healing_magic"}, {1, "holy_binding"},
			{20, "defensive_magic"}, {35, "goddess_light"},
		},
		Description: "A priest who joins Frieren's party for a while and is good at healing.",
	},
}

// charactersByID is built once at startup for fast lookups.
var charactersByID = func() map[string]Character {
	m := make(map[string]Character, len(Characters))
	for _, c := range Characters {
		m[c.ID] = c
	}
	return m
}()

// CharacterByID returns the character with that ID. The second result reports
// whether it was found: the "comma ok" idiom, used instead of returning null.
func CharacterByID(id string) (Character, bool) {
	c, ok := charactersByID[id]
	return c, ok
}
