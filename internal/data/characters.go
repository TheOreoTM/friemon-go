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
	EVYield     Stats // EVs a character gains for knocking this one out in battle (1-3 total)
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
		Types:     []Type{TypeDestruction, TypeNature},
		BaseStats: Stats{HP: 80, Attack: 45, Defense: 70, MagicAttack: 150, MagicDefense: 135, Speed: 120},
		EVYield:   Stats{MagicAttack: 3},
		Learnset: []LearnsetEntry{
			{1, "zoltraak"}, {1, "flower_field"}, {1, "defensive_magic"},
			{8, "binding_roots"}, {12, "quick_cast"}, {18, "frost_magic"}, {22, "bloom"},
			{26, "mana_focus"}, {32, "lightning_magic"}, {38, "hellfire"}, {44, "rapid_barrage"},
			{50, "barrier_shatter"}, {58, "petal_storm"}, {70, "mass_offensive"},
			{15, "storm_call"}, {28, "heatwave"}, {34, "overgrowth"}, {40, "full_barrier"}, {46, "blizzard"}, {52, "flash_retreat"}, {55, "clearing_wind"}, {62, "sunlight_lance"}, {66, "mana_mines"}, {80, "annihilation"},
		},
		Description: "An elf mage who has lived for over a thousand years and collects spells as a hobby.",
	},
	{
		ID:        "himmel",
		Name:      "Himmel",
		Aliases:   []string{"Himmel the Hero"},
		Rarity:    Rare,
		Types:     []Type{TypeMartial},
		BaseStats: Stats{HP: 85, Attack: 115, Defense: 90, MagicAttack: 50, MagicDefense: 80, Speed: 110},
		EVYield:   Stats{Attack: 2},
		Learnset: []LearnsetEntry{
			{1, "slash"}, {1, "counter_blow"},
			{10, "war_cry"}, {18, "heroic_strike"}, {24, "holy_smite"}, {32, "battle_stance"},
			{40, "flower_field"}, {48, "axe_cleave"},
			{14, "provoke"}, {28, "hit_and_run"}, {36, "tailwind"}, {44, "overgrowth"},
		},
		Description: "The hero who defeated the Demon King, remembered by statues in every village.",
	},
	{
		ID:        "heiter",
		Name:      "Heiter",
		Rarity:    Rare,
		Types:     []Type{TypeHoly},
		BaseStats: Stats{HP: 90, Attack: 40, Defense: 70, MagicAttack: 115, MagicDefense: 125, Speed: 60},
		EVYield:   Stats{MagicDefense: 2},
		Learnset: []LearnsetEntry{
			{1, "healing_magic"}, {1, "holy_binding"},
			{8, "holy_smite"}, {14, "purify"}, {20, "defensive_magic"}, {26, "goddess_light"},
			{32, "sleep_magic"}, {38, "hexagon_ward"}, {44, "blessed_rest"},
			{18, "mist"}, {24, "binding_chains"}, {30, "sacred_ground"}, {36, "magic_ward"}, {48, "physical_ward"},
		},
		Description: "A priest of the hero's party, fond of drink and wiser than he lets on.",
	},
	{
		ID:        "eisen",
		Name:      "Eisen",
		Rarity:    Rare,
		Types:     []Type{TypeMartial},
		BaseStats: Stats{HP: 120, Attack: 115, Defense: 135, MagicAttack: 20, MagicDefense: 80, Speed: 40},
		EVYield:   Stats{Defense: 2},
		Learnset: []LearnsetEntry{
			{1, "slash"}, {1, "iron_body"},
			{8, "war_cry"}, {14, "counter_blow"}, {20, "mana_wall"}, {26, "ground_splitter"},
			{34, "battle_stance"}, {42, "axe_cleave"},
			{12, "provoke"}, {18, "sandstorm"}, {30, "physical_ward"}, {38, "caltrops"}, {46, "full_barrier"},
		},
		Description: "A dwarf warrior of the hero's party who can take almost any blow.",
	},
	{
		ID:        "ubel",
		Name:      "Übel",
		Rarity:    Rare,
		Types:     []Type{TypeDestruction},
		BaseStats: Stats{HP: 70, Attack: 50, Defense: 60, MagicAttack: 130, MagicDefense: 70, Speed: 115},
		EVYield:   Stats{MagicAttack: 2},
		Learnset: []LearnsetEntry{
			{1, "zoltraak"}, {1, "cutting_magic"},
			{8, "quick_cast"}, {14, "binding_roots"}, {20, "curse_mark"}, {26, "illusion"},
			{32, "mana_focus"}, {38, "soul_drain"}, {44, "rapid_barrage"}, {52, "demons_whisper"},
			{16, "mana_field"}, {24, "vine_snare"}, {30, "compulsion"}, {36, "illusory_double"}, {42, "curse_seals"}, {48, "reversal_field"},
		},
		Description: "A mage who can cut anything she imagines she can cut.",
	},
	{
		ID:        "fern",
		Name:      "Fern",
		Rarity:    Common,
		Types:     []Type{TypeDestruction},
		BaseStats: Stats{HP: 65, Attack: 35, Defense: 50, MagicAttack: 110, MagicDefense: 75, Speed: 100},
		EVYield:   Stats{MagicAttack: 1},
		Learnset: []LearnsetEntry{
			{1, "zoltraak"}, {1, "defensive_magic"},
			{6, "quick_cast"}, {12, "rapid_barrage"}, {18, "mana_focus"}, {24, "barrier_shatter"},
			{30, "hexagon_ward"}, {36, "lightning_magic"}, {44, "hellfire"}, {54, "mass_offensive"},
			{10, "full_barrier"}, {16, "flash_retreat"}, {28, "heatwave"}, {33, "mana_field"}, {40, "mana_mines"}, {60, "annihilation"},
		},
		Description: "Frieren's apprentice, a fast and precise caster of basic offensive magic.",
	},
	{
		ID:        "stark",
		Name:      "Stark",
		Rarity:    Common,
		Types:     []Type{TypeMartial},
		BaseStats: Stats{HP: 95, Attack: 110, Defense: 85, MagicAttack: 20, MagicDefense: 55, Speed: 70},
		EVYield:   Stats{Attack: 1},
		Learnset: []LearnsetEntry{
			{1, "slash"}, {1, "war_cry"},
			{8, "counter_blow"}, {15, "iron_body"}, {22, "battle_stance"}, {30, "ground_splitter"},
			{38, "barrier_ram"}, {46, "axe_cleave"},
			{12, "hit_and_run"}, {18, "sandstorm"}, {26, "provoke"}, {34, "caltrops"},
		},
		Description: "Eisen's former pupil, a warrior far braver in battle than he believes.",
	},
	{
		ID:        "sein",
		Name:      "Sein",
		Rarity:    Common,
		Types:     []Type{TypeHoly},
		BaseStats: Stats{HP: 80, Attack: 45, Defense: 60, MagicAttack: 80, MagicDefense: 95, Speed: 65},
		EVYield:   Stats{MagicDefense: 1},
		Learnset: []LearnsetEntry{
			{1, "healing_magic"}, {1, "holy_binding"},
			{10, "holy_smite"}, {16, "purify"}, {22, "defensive_magic"}, {30, "goddess_light"},
			{38, "blessed_rest"},
			{12, "storm_call"}, {18, "binding_chains"}, {26, "sacred_ground"}, {34, "magic_ward"},
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
