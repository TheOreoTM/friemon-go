package data

// Category decides which stats a move uses in battle.
type Category int

const (
	Physical Category = iota // Attack vs Defense
	Magical                  // Magic Attack vs Magic Defense
	Status                   // no damage; works through Heal and Effects
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

// Flag marks a move property. Flags are bits, so a move can have several: FlagContact | FlagSound.
type Flag uint

const (
	FlagContact      Flag = 1 << iota // touches the target
	FlagSound                         // a sound-based move
	FlagRest                          // special: the user sleeps for 2 turns and fully heals
	FlagProtect                       // blocks moves aimed at the user this turn; fails more often if repeated
	FlagSubstitute                    // spends 1/4 HP to make a decoy that takes hits
	FlagCharge                        // spends a turn charging, then hits on the next
	FlagRecharge                      // the user must rest on the next turn
	FlagPivot                         // the user switches out after hitting
	FlagGround                        // shakes the ground; weaker in Overgrowth
	FlagClearHazards                  // removes hazards from both sides and the target's wards
	FlagReversal                      // starts (or ends) Reversal Field: slower characters move first
)

// StatChange raises (positive) or lowers (negative) a stat by a number of stages.
type StatChange struct {
	Stat   Stat
	Stages int
}

// MoveEffect is one thing a move does besides damage. Everything in it happens together,
// rolled once against Chance.
type MoveEffect struct {
	Chance     int  // percent; 0 means it always happens
	Self       bool // true: affects the user; false: affects the target
	Status     Condition
	Volatile   Volatile
	Stats      []StatChange
	CureStatus bool // removes the affected character's status
}

// Move is something a character can do in battle. Characters refer to moves by ID.
type Move struct {
	ID          string // stable key saved in the database; never change it once characters know the move
	Name        string
	Type        Type
	Category    Category
	Power       int // 0 for status moves
	Accuracy    int // percent; 0 means it never misses
	PP          int // how many times it can be used in one battle
	Priority    int // higher goes first, before Speed is compared
	CritStage   int // 0 normal, 1 high, 3 always
	Flags       Flag
	MinHits     int // multi-hit moves; 0 means a single hit
	MaxHits     int
	Drain       int // percent of damage dealt that heals the user
	Recoil      int // percent of damage dealt that hurts the user
	Heal        int // percent of max HP the user recovers (status moves)
	Effects     []MoveEffect
	Weather     Weather       // weather this move starts
	Terrain     Terrain       // terrain this move starts
	Side        SideCondition // ward or Tailwind on the user's side, or a hazard on the target's
	SureHitIn   Weather       // never misses in this weather
	NoChargeIn  Weather       // a charge move that fires at once in this weather
	Description string
}

// Has reports whether the move has a flag.
func (m Move) Has(f Flag) bool { return m.Flags&f != 0 }

// Struggle is used when a character has no PP left in any move. It is not learnable.
var Struggle = Move{
	ID: "struggle", Name: "Struggle", Type: NoType, Category: Physical, Power: 50, PP: 1,
	Flags:       FlagContact,
	Description: "Used only when nothing else is left. Hurts the user by 1/4 of its max HP.",
}

// shorthand helpers keep the move list readable.
func raise(stat Stat, stages int) []StatChange { return []StatChange{{stat, stages}} }
func lower(stat Stat, stages int) []StatChange { return []StatChange{{stat, -stages}} }

// Moves is every learnable move.
var Moves = []Move{
	// Destruction: offensive magic.
	{ID: "zoltraak", Name: "Zoltraak", Type: TypeDestruction, Category: Magical, Power: 60, Accuracy: 100, PP: 25,
		Description: "Ordinary offensive magic: a beam once made to kill demons, now taught to every mage."},
	{ID: "quick_cast", Name: "Quick Cast", Type: TypeDestruction, Category: Magical, Power: 40, Accuracy: 100, PP: 30, Priority: 1,
		Description: "A spell fired before the target can react. Usually goes first."},
	{ID: "rapid_barrage", Name: "Rapid Barrage", Type: TypeDestruction, Category: Magical, Power: 20, Accuracy: 95, PP: 20, MinHits: 2, MaxHits: 5,
		Description: "A flurry of small offensive spells. Hits 2 to 5 times."},
	{ID: "cutting_magic", Name: "Cutting Magic", Type: TypeDestruction, Category: Magical, Power: 70, Accuracy: 100, PP: 20, CritStage: 1,
		Description: "Cuts anything the caster believes can be cut. Lands critical hits more often."},
	{ID: "lightning_magic", Name: "Lightning Magic", Type: TypeDestruction, Category: Magical, Power: 90, Accuracy: 100, PP: 15,
		Effects: []MoveEffect{{Chance: 10, Status: Bound}}, SureHitIn: Storm,
		Description: "Calls down lightning. 10% chance to leave the target Bound. Never misses in a Storm."},
	{ID: "hellfire", Name: "Hellfire Magic", Type: TypeDestruction, Category: Magical, Power: 90, Accuracy: 100, PP: 15,
		Effects:     []MoveEffect{{Chance: 10, Status: Burn}},
		Description: "Summons the fires of hell. 10% chance to burn the target."},
	{ID: "mass_offensive", Name: "Mass Offensive Magic", Type: TypeDestruction, Category: Magical, Power: 130, Accuracy: 90, PP: 5,
		Effects:     []MoveEffect{{Self: true, Stats: lower(StatMagicAttack, 2)}},
		Description: "Hundreds of spells at once. Sharply lowers the user's Magic Attack afterwards."},
	{ID: "mana_focus", Name: "Mana Focus", Type: TypeDestruction, Category: Status, PP: 20,
		Effects:     []MoveEffect{{Self: true, Stats: raise(StatMagicAttack, 2)}},
		Description: "Gathers mana for the next spell. Sharply raises Magic Attack."},

	// Barrier: defensive magic.
	{ID: "defensive_magic", Name: "Defensive Magic", Type: TypeBarrier, Category: Status, PP: 20,
		Effects:     []MoveEffect{{Self: true, Stats: raise(StatMagicDefense, 2)}},
		Description: "A hexagonal barrier against offensive magic. Sharply raises Magic Defense."},
	{ID: "mana_wall", Name: "Mana Wall", Type: TypeBarrier, Category: Status, PP: 15,
		Effects:     []MoveEffect{{Self: true, Stats: raise(StatDefense, 2)}},
		Description: "Hardens the air with mana. Sharply raises Defense."},
	{ID: "hexagon_ward", Name: "Hexagon Ward", Type: TypeBarrier, Category: Status, PP: 20,
		Effects:     []MoveEffect{{Self: true, Stats: []StatChange{{StatDefense, 1}, {StatMagicDefense, 1}}}},
		Description: "Layers of small barriers. Raises Defense and Magic Defense."},
	{ID: "barrier_shatter", Name: "Barrier Shatter", Type: TypeBarrier, Category: Magical, Power: 70, Accuracy: 100, PP: 15,
		Effects:     []MoveEffect{{Chance: 50, Stats: lower(StatMagicDefense, 1)}},
		Description: "Cracks the target's defenses. 50% chance to lower its Magic Defense."},
	{ID: "barrier_ram", Name: "Barrier Ram", Type: TypeBarrier, Category: Physical, Power: 90, Accuracy: 100, PP: 15, Flags: FlagContact, Recoil: 25,
		Description: "Charges with a barrier raised. The user takes 1/4 of the damage dealt."},

	// Holy: priests and healing.
	{ID: "healing_magic", Name: "Healing Magic", Type: TypeHoly, Category: Status, PP: 10, Heal: 50,
		Description: "A priest's prayer. Restores half of the user's max HP."},
	{ID: "blessed_rest", Name: "Blessed Rest", Type: TypeHoly, Category: Status, PP: 10, Flags: FlagRest,
		Description: "Sleeps for 2 turns to fully restore HP and cure any status."},
	{ID: "holy_binding", Name: "Holy Binding", Type: TypeHoly, Category: Magical, Power: 20, Accuracy: 100, PP: 20,
		Effects:     []MoveEffect{{Status: Bound}},
		Description: "Scripture that binds the target in place. Always leaves it Bound."},
	{ID: "holy_smite", Name: "Holy Smite", Type: TypeHoly, Category: Physical, Power: 75, Accuracy: 100, PP: 15, Flags: FlagContact,
		Effects:     []MoveEffect{{Chance: 10, Status: Burn}},
		Description: "A blow wreathed in holy light. 10% chance to burn."},
	{ID: "goddess_light", Name: "Goddess's Light", Type: TypeHoly, Category: Magical, Power: 90, Accuracy: 100, PP: 10,
		Effects:     []MoveEffect{{Chance: 10, Stats: lower(StatMagicDefense, 1)}},
		Description: "Holy light that burns away curses. 10% chance to lower Magic Defense."},
	{ID: "purify", Name: "Purify", Type: TypeHoly, Category: Status, PP: 20,
		Effects:     []MoveEffect{{Self: true, CureStatus: true}},
		Description: "Cleanses the user of any status condition."},

	// Martial: warriors and weapons.
	{ID: "slash", Name: "Slash", Type: TypeMartial, Category: Physical, Power: 70, Accuracy: 100, PP: 20, Flags: FlagContact, CritStage: 1,
		Description: "A clean sword strike. Lands critical hits more often."},
	{ID: "counter_blow", Name: "Counter Blow", Type: TypeMartial, Category: Physical, Power: 40, Accuracy: 100, PP: 30, Flags: FlagContact, Priority: 1,
		Description: "Steps inside an attack and hits back. Usually goes first."},
	{ID: "heroic_strike", Name: "Heroic Strike", Type: TypeMartial, Category: Physical, Power: 90, Accuracy: 100, PP: 15, Flags: FlagContact,
		Description: "A sword swing with the full weight of a hero's resolve."},
	{ID: "ground_splitter", Name: "Ground Splitter", Type: TypeMartial, Category: Physical, Power: 100, Accuracy: 90, PP: 10,
		Flags: FlagGround, Effects: []MoveEffect{{Chance: 30, Volatile: Flinch}},
		Description: "Smashes the earth with an axe. 30% chance to make the target flinch. Weaker in Overgrowth."},
	{ID: "axe_cleave", Name: "Axe Cleave", Type: TypeMartial, Category: Physical, Power: 120, Accuracy: 100, PP: 15, Flags: FlagContact, Recoil: 33,
		Description: "An all-out overhead swing. The user takes 1/3 of the damage dealt."},
	{ID: "iron_body", Name: "Iron Body", Type: TypeMartial, Category: Status, PP: 20,
		Effects:     []MoveEffect{{Self: true, Stats: []StatChange{{StatAttack, 1}, {StatDefense, 1}}}},
		Description: "Tenses every muscle. Raises Attack and Defense."},
	{ID: "battle_stance", Name: "Battle Stance", Type: TypeMartial, Category: Status, PP: 20,
		Effects:     []MoveEffect{{Self: true, Stats: raise(StatAttack, 2)}},
		Description: "Settles into a fighting stance. Sharply raises Attack."},
	{ID: "war_cry", Name: "War Cry", Type: TypeMartial, Category: Status, Accuracy: 100, PP: 40, Flags: FlagSound,
		Effects:     []MoveEffect{{Stats: lower(StatAttack, 1)}},
		Description: "A shout that shakes the target's nerve. Lowers its Attack."},

	// Curse: demon magic.
	{ID: "curse_mark", Name: "Curse Mark", Type: TypeCurse, Category: Magical, Power: 65, Accuracy: 100, PP: 20,
		Effects:     []MoveEffect{{Chance: 30, Status: Cursed}},
		Description: "Demon magic that eats away at the target. 30% chance to curse it."},
	{ID: "deep_curse", Name: "Deep Curse", Type: TypeCurse, Category: Status, Accuracy: 90, PP: 10,
		Effects:     []MoveEffect{{Status: DeeplyCursed}},
		Description: "A curse that grows worse every turn."},
	{ID: "soul_drain", Name: "Soul Drain", Type: TypeCurse, Category: Magical, Power: 75, Accuracy: 100, PP: 10, Drain: 50,
		Description: "Steals the target's life force. Heals the user by half the damage dealt."},
	{ID: "illusion", Name: "Illusion", Type: TypeCurse, Category: Status, Accuracy: 100, PP: 10,
		Effects:     []MoveEffect{{Volatile: Confusion}},
		Description: "Shows the target things that are not there. Confuses it."},
	{ID: "sleep_magic", Name: "Sleep Magic", Type: TypeCurse, Category: Status, Accuracy: 75, PP: 15,
		Effects:     []MoveEffect{{Status: Sleep}},
		Description: "Puts the target to sleep."},
	{ID: "demons_whisper", Name: "Demon's Whisper", Type: TypeCurse, Category: Magical, Power: 80, Accuracy: 100, PP: 15, Flags: FlagSound,
		Effects:     []MoveEffect{{Chance: 20, Volatile: Flinch}},
		Description: "Words that freeze the heart. 20% chance to make the target flinch."},

	// Nature: folk spells.
	{ID: "flower_field", Name: "Flower Field", Type: TypeNature, Category: Status, PP: 10, Heal: 25,
		Effects:     []MoveEffect{{Self: true, CureStatus: true}},
		Description: "Frieren's favorite spell. The flowers soothe the user: cures its status and restores 1/4 of its HP."},
	{ID: "binding_roots", Name: "Binding Roots", Type: TypeNature, Category: Magical, Power: 55, Accuracy: 95, PP: 15,
		Effects:     []MoveEffect{{Stats: lower(StatSpeed, 1)}},
		Description: "Roots burst from the ground and hold the target. Lowers its Speed."},
	{ID: "thorn_lash", Name: "Thorn Lash", Type: TypeNature, Category: Physical, Power: 25, Accuracy: 100, PP: 20, Flags: FlagContact, MinHits: 2, MaxHits: 5,
		Description: "Whips the target with thorny vines. Hits 2 to 5 times."},
	{ID: "frost_magic", Name: "Frost Magic", Type: TypeNature, Category: Magical, Power: 90, Accuracy: 100, PP: 10,
		Effects: []MoveEffect{{Chance: 10, Status: Freeze}}, SureHitIn: Blizzard,
		Description: "Freezing wind from the northern lands. 10% chance to freeze the target. Never misses in a Blizzard."},
	{ID: "drowsy_pollen", Name: "Drowsy Pollen", Type: TypeNature, Category: Status, Accuracy: 75, PP: 15,
		Effects:     []MoveEffect{{Status: Sleep}},
		Description: "A cloud of sleepy pollen. Puts the target to sleep."},
	{ID: "bloom", Name: "Bloom", Type: TypeNature, Category: Status, PP: 20,
		Effects:     []MoveEffect{{Self: true, Stats: []StatChange{{StatAttack, 1}, {StatMagicAttack, 1}}}},
		Description: "Draws strength from growing things. Raises Attack and Magic Attack."},
	{ID: "petal_storm", Name: "Petal Storm", Type: TypeNature, Category: Magical, Power: 110, Accuracy: 85, PP: 10,
		Description: "A storm of razor-sharp petals. Powerful but hard to aim."},

	// Weather.
	{ID: "heatwave", Name: "Heatwave", Type: TypeDestruction, Category: Status, PP: 5, Weather: Heatwave,
		Description: "Scorches the battlefield for 5 turns. Destruction moves are stronger, Nature moves weaker, and nobody can be frozen."},
	{ID: "storm_call", Name: "Storm Call", Type: TypeNature, Category: Status, PP: 5, Weather: Storm,
		Description: "Summons a storm for 5 turns. Nature moves are stronger, Destruction moves weaker, and nobody can be burned."},
	{ID: "sandstorm", Name: "Sandstorm", Type: TypeMartial, Category: Status, PP: 10, Weather: Sandstorm,
		Description: "Whips up sand for 5 turns. Everyone but Martial types is hurt each turn; Martial types gain Magic Defense."},
	{ID: "blizzard", Name: "Blizzard", Type: TypeNature, Category: Status, PP: 10, Weather: Blizzard,
		Description: "Calls a blizzard for 5 turns. Everyone but Nature types is hurt each turn; Nature types gain Defense."},

	// Terrain.
	{ID: "mana_field", Name: "Mana Field", Type: TypeDestruction, Category: Status, PP: 10, Terrain: ManaField,
		Description: "Charges the ground with mana for 5 turns. Destruction moves are stronger and nobody can fall asleep."},
	{ID: "sacred_ground", Name: "Sacred Ground", Type: TypeHoly, Category: Status, PP: 10, Terrain: SacredGround,
		Description: "Blesses the ground for 5 turns. Holy moves are stronger, Curse moves weaker, and priority moves cannot hit."},
	{ID: "overgrowth", Name: "Overgrowth", Type: TypeNature, Category: Status, PP: 10, Terrain: Overgrowth,
		Description: "Covers the ground in plants for 5 turns. Nature moves are stronger and everyone heals a little each turn."},
	{ID: "mist", Name: "Mist", Type: TypeBarrier, Category: Status, PP: 10, Terrain: Mist,
		Description: "Fills the field with mist for 5 turns. Nobody can get a new condition or be confused, and Curse moves are weaker."},

	// Wards and field control.
	{ID: "physical_ward", Name: "Physical Ward", Type: TypeBarrier, Category: Status, PP: 20, Side: PhysicalWard,
		Description: "A barrier that halves physical damage to your side for 5 turns."},
	{ID: "magic_ward", Name: "Magic Ward", Type: TypeBarrier, Category: Status, PP: 20, Side: MagicWard,
		Description: "A barrier that halves magical damage to your side for 5 turns."},
	{ID: "tailwind", Name: "Tailwind", Type: TypeNature, Category: Status, PP: 15, Side: Tailwind,
		Description: "A wind at your back doubles your side's Speed for 4 turns."},
	{ID: "reversal_field", Name: "Reversal Field", Type: TypeCurse, Category: Status, PP: 5, Priority: -7, Flags: FlagReversal,
		Description: "Twists time for 5 turns so slower characters move first. Using it again ends it. Always goes last."},
	{ID: "full_barrier", Name: "Full Barrier", Type: TypeBarrier, Category: Status, PP: 10, Priority: 4, Flags: FlagProtect,
		Description: "Blocks every move aimed at the user this turn. Likely to fail if used twice in a row."},
	{ID: "illusory_double", Name: "Illusory Double", Type: TypeCurse, Category: Status, PP: 10, Flags: FlagSubstitute,
		Description: "Spends 1/4 of max HP to make an illusion that takes hits in the user's place."},

	// Hazards.
	{ID: "mana_mines", Name: "Mana Mines", Type: TypeDestruction, Category: Status, PP: 20, Side: ManaMines,
		Description: "Hides mana mines on the opponent's side. They hurt anyone switching in, more so if weak to Destruction."},
	{ID: "caltrops", Name: "Caltrops", Type: TypeMartial, Category: Status, PP: 20, Side: Caltrops,
		Description: "Scatters spikes on the opponent's side that hurt anyone switching in. Stacks up to 3 times."},
	{ID: "curse_seals", Name: "Curse Seals", Type: TypeCurse, Category: Status, PP: 20, Side: CurseSeals,
		Description: "Plants seals on the opponent's side that curse anyone switching in. Two layers deeply curse."},
	{ID: "clearing_wind", Name: "Clearing Wind", Type: TypeNature, Category: Status, PP: 15, Flags: FlagClearHazards,
		Description: "A gust that blows away hazards on both sides and the opponent's wards."},

	// Charge, recharge, traps, control and pivots.
	{ID: "sunlight_lance", Name: "Sunlight Lance", Type: TypeNature, Category: Magical, Power: 120, Accuracy: 100, PP: 10,
		Flags: FlagCharge, NoChargeIn: Heatwave,
		Description: "Gathers light for a turn, then fires. Fires at once in a Heatwave."},
	{ID: "annihilation", Name: "Annihilation Beam", Type: TypeDestruction, Category: Magical, Power: 150, Accuracy: 90, PP: 5,
		Flags:       FlagRecharge,
		Description: "A beam that erases everything in its path. The user must rest on the next turn."},
	{ID: "binding_chains", Name: "Binding Chains", Type: TypeHoly, Category: Status, Accuracy: 0, PP: 5,
		Effects:     []MoveEffect{{Volatile: Trap}},
		Description: "Holy chains stop the target from switching out while the user stays in."},
	{ID: "vine_snare", Name: "Vine Snare", Type: TypeNature, Category: Physical, Power: 35, Accuracy: 90, PP: 15, Flags: FlagContact,
		Effects:     []MoveEffect{{Volatile: PartialTrap}},
		Description: "Vines wrap the target for 4-5 turns. It cannot switch out and loses 1/8 of its HP each turn."},
	{ID: "provoke", Name: "Provoke", Type: TypeMartial, Category: Status, Accuracy: 100, PP: 20,
		Effects:     []MoveEffect{{Volatile: Taunt}},
		Description: "Goads the target so it can only use damaging moves for 3 turns."},
	{ID: "compulsion", Name: "Compulsion", Type: TypeCurse, Category: Status, Accuracy: 100, PP: 5,
		Effects:     []MoveEffect{{Volatile: Encore}},
		Description: "Demon magic that forces the target to repeat its last move for 3 turns."},
	{ID: "hit_and_run", Name: "Hit and Run", Type: TypeMartial, Category: Physical, Power: 70, Accuracy: 100, PP: 20,
		Flags:       FlagContact | FlagPivot,
		Description: "Strikes, then switches out for another party member."},
	{ID: "flash_retreat", Name: "Flash Retreat", Type: TypeDestruction, Category: Magical, Power: 70, Accuracy: 100, PP: 20,
		Flags:       FlagPivot,
		Description: "Fires a spell, then flashes back to switch out for another party member."},
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
