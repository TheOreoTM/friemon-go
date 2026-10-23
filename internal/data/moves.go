package data

// Category decides which stats a move uses in battle.
type Category int

const (
	Physical Category = iota // Attack vs Defense
	Magical                  // Magic Attack vs Magic Defense
	Status                   // no damage; uses Effect instead
)

func (c Category) String() string {
	switch c {
	case Physical:
		return "Physical"
	case Magical:
		return "Magical"
	case Status:
		return "Status"
	}
	return "Unknown"
}

// Effect is what a status move does.
type Effect int

const (
	NoEffect          Effect = iota
	Heal                     // restores half of max HP
	RaiseDefense             // +1 Defense stage
	RaiseMagicDefense        // +1 Magic Defense stage
)

// Move is something a character can do in battle. Characters refer to moves by ID.
type Move struct {
	ID          string // stable key saved in the database; never change it once characters know the move
	Name        string
	Type        Type
	Category    Category
	Power       int    // 0 for status moves
	Accuracy    int    // percent
	Effect      Effect // status moves only
	Description string
}

// Moves is every move in the game.
var Moves = []Move{
	// Destruction
	{ID: "zoltraak", Name: "Zoltraak", Type: Destruction, Category: Magical, Power: 60, Accuracy: 100,
		Description: "Ordinary offensive magic: a beam that once killed demons and is now taught to every mage."},
	{ID: "rapid_barrage", Name: "Rapid Barrage", Type: Destruction, Category: Magical, Power: 75, Accuracy: 90,
		Description: "A flurry of offensive spells fired faster than the target can block."},
	{ID: "lightning_magic", Name: "Lightning Magic", Type: Destruction, Category: Magical, Power: 90, Accuracy: 85,
		Description: "Calls down lightning on a single target."},
	{ID: "cutting_magic", Name: "Cutting Magic", Type: Destruction, Category: Magical, Power: 85, Accuracy: 95,
		Description: "Cuts anything the caster believes can be cut."},
	{ID: "mass_offensive", Name: "Mass Offensive Magic", Type: Destruction, Category: Magical, Power: 110, Accuracy: 75,
		Description: "Hundreds of offensive spells at once. Hard to aim, harder to survive."},

	// Barrier
	{ID: "defensive_magic", Name: "Defensive Magic", Type: Barrier, Category: Status, Accuracy: 100, Effect: RaiseMagicDefense,
		Description: "A hexagonal barrier that blocks offensive magic."},
	{ID: "barrier_shatter", Name: "Barrier Shatter", Type: Barrier, Category: Magical, Power: 70, Accuracy: 95,
		Description: "Breaks the target's defenses with focused mana."},

	// Holy
	{ID: "healing_magic", Name: "Healing Magic", Type: Holy, Category: Status, Accuracy: 100, Effect: Heal,
		Description: "A priest's prayer that closes wounds."},
	{ID: "goddess_light", Name: "Goddess's Light", Type: Holy, Category: Magical, Power: 70, Accuracy: 100,
		Description: "Holy light that burns away curses."},
	{ID: "holy_binding", Name: "Holy Binding", Type: Holy, Category: Magical, Power: 50, Accuracy: 100,
		Description: "Scripture that binds the target in place."},

	// Martial
	{ID: "slash", Name: "Slash", Type: Martial, Category: Physical, Power: 50, Accuracy: 100,
		Description: "A clean sword strike."},
	{ID: "heroic_strike", Name: "Heroic Strike", Type: Martial, Category: Physical, Power: 85, Accuracy: 100,
		Description: "A sword swing with the full weight of a hero's resolve."},
	{ID: "axe_cleave", Name: "Axe Cleave", Type: Martial, Category: Physical, Power: 95, Accuracy: 85,
		Description: "A huge overhead axe swing that splits the ground."},
	{ID: "iron_body", Name: "Iron Body", Type: Martial, Category: Status, Accuracy: 100, Effect: RaiseDefense,
		Description: "Braces to shrug off a blow that would fell anyone else."},
	{ID: "counter_blow", Name: "Counter Blow", Type: Martial, Category: Physical, Power: 65, Accuracy: 100,
		Description: "Steps inside an attack and hits back."},

	// Nature
	{ID: "flower_field", Name: "Flower Field", Type: Nature, Category: Status, Accuracy: 100,
		Description: "Fills the area with flowers. Useless in a fight, and Frieren's favorite spell."},
	{ID: "binding_roots", Name: "Binding Roots", Type: Nature, Category: Magical, Power: 55, Accuracy: 100,
		Description: "Roots burst from the ground and hold the target."},

	// Curse
	{ID: "curse_mark", Name: "Curse Mark", Type: Curse, Category: Magical, Power: 65, Accuracy: 95,
		Description: "Demon magic that eats away at the target."},
}

var movesByID = func() map[string]Move {
	m := make(map[string]Move, len(Moves))
	for _, mv := range Moves {
		m[mv.ID] = mv
	}
	return m
}()

// MoveByID returns the move with that ID, and whether it was found.
func MoveByID(id string) (Move, bool) {
	m, ok := movesByID[id]
	return m, ok
}
