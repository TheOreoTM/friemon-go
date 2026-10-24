package data

// Weather affects the whole battlefield for 5 turns.
type Weather int

const (
	NoWeather Weather = iota
	Heatwave          // (Sun) Destruction x1.5, Nature x0.5; no freezing; stronger heals
	Storm             // (Rain) Nature x1.5, Destruction x0.5; no burns
	Sandstorm         // 1/16 damage a turn except to Martial; Martial x1.5 Magic Defense
	Blizzard          // (Hail/Snow) 1/16 damage a turn except to Nature; Nature x1.5 Defense
)

func (w Weather) String() string {
	switch w {
	case Heatwave:
		return "Heatwave"
	case Storm:
		return "Storm"
	case Sandstorm:
		return "Sandstorm"
	case Blizzard:
		return "Blizzard"
	}
	return "None"
}

// Terrain affects the whole battlefield for 5 turns.
type Terrain int

const (
	NoTerrain    Terrain = iota
	ManaField            // Destruction x1.3; nobody can fall asleep
	SacredGround         // Holy x1.3, Curse x0.5; priority moves cannot hit
	Overgrowth           // Nature x1.3; everyone heals 1/16 a turn; ground moves x0.5
	Mist                 // no new conditions or confusion; Curse x0.5
)

func (t Terrain) String() string {
	switch t {
	case ManaField:
		return "Mana Field"
	case SacredGround:
		return "Sacred Ground"
	case Overgrowth:
		return "Overgrowth"
	case Mist:
		return "Mist"
	}
	return "None"
}

// SideCondition affects one side of the field.
type SideCondition int

const (
	NoSideCondition SideCondition = iota
	PhysicalWard                  // (Reflect) halves physical damage to this side, 5 turns
	MagicWard                     // (Light Screen) halves magical damage to this side, 5 turns
	Tailwind                      // doubles this side's Speed, 4 turns
	ManaMines                     // (Stealth Rock) hurts switch-ins by how weak they are to Destruction
	Caltrops                      // (Spikes) hurts switch-ins; up to 3 layers
	CurseSeals                    // (Toxic Spikes) curses switch-ins; 2 layers deeply curse
)

func (c SideCondition) String() string {
	switch c {
	case PhysicalWard:
		return "Physical Ward"
	case MagicWard:
		return "Magic Ward"
	case Tailwind:
		return "Tailwind"
	case ManaMines:
		return "Mana Mines"
	case Caltrops:
		return "Caltrops"
	case CurseSeals:
		return "Curse Seals"
	}
	return "None"
}

// Hazard reports whether the condition is set on the opponent's side to hurt switch-ins.
func (c SideCondition) Hazard() bool {
	return c == ManaMines || c == Caltrops || c == CurseSeals
}

// MaxLayers is how many times a hazard can be stacked.
func (c SideCondition) MaxLayers() int {
	switch c {
	case Caltrops:
		return 3
	case CurseSeals:
		return 2
	}
	return 1
}

var weatherBoosts = map[Weather]map[Type]float64{
	Heatwave: {TypeDestruction: 1.5, TypeNature: 0.5},
	Storm:    {TypeNature: 1.5, TypeDestruction: 0.5},
}

var terrainBoosts = map[Terrain]map[Type]float64{
	ManaField:    {TypeDestruction: 1.3},
	SacredGround: {TypeHoly: 1.3, TypeCurse: 0.5},
	Overgrowth:   {TypeNature: 1.3},
	Mist:         {TypeCurse: 0.5},
}

// WeatherMultiplier is the damage multiplier weather gives a move of type t.
func WeatherMultiplier(w Weather, t Type) float64 { return lookup(weatherBoosts[w], t) }

// TerrainMultiplier is the damage multiplier terrain gives a move of type t.
func TerrainMultiplier(tr Terrain, t Type) float64 { return lookup(terrainBoosts[tr], t) }

func lookup(m map[Type]float64, t Type) float64 {
	if x, ok := m[t]; ok {
		return x
	}
	return 1
}

// weatherShelter is the type that takes no damage from a weather, and gets a defense boost in it.
var weatherShelter = map[Weather]Type{
	Sandstorm: TypeMartial,
	Blizzard:  TypeNature,
}

// WeatherShelter returns the type sheltered from w, if any.
func WeatherShelter(w Weather) (Type, bool) {
	t, ok := weatherShelter[w]
	return t, ok
}
