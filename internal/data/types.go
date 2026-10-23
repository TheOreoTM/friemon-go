package data

// Type is a character's or move's type. Battles use a type chart built on these.
type Type int

const (
	Destruction Type = iota // offensive magic, like Zoltraak
	Barrier                 // defensive magic
	Holy                    // priests and healing
	Martial                 // warriors and weapons
	Curse                   // demon magic
	Nature                  // folk spells, flowers
)

// AllTypes lists every type, in display order.
var AllTypes = []Type{Destruction, Barrier, Holy, Martial, Curse, Nature}

func (t Type) String() string {
	switch t {
	case Destruction:
		return "Destruction"
	case Barrier:
		return "Barrier"
	case Holy:
		return "Holy"
	case Martial:
		return "Martial"
	case Curse:
		return "Curse"
	case Nature:
		return "Nature"
	}
	return "Unknown"
}

// Stats is one number per battle stat. It is used for base stats, hidden rolls and final stats.
type Stats struct {
	HP           int
	Attack       int
	Defense      int
	MagicAttack  int
	MagicDefense int
	Speed        int
}

// typeChart lists the matchups that are not neutral: attacking type -> defending type -> multiplier.
// 2 is super effective, 0.5 is not very effective; anything missing is 1.
var typeChart = map[Type]map[Type]float64{
	Destruction: {Curse: 2, Martial: 2, Barrier: 0.5}, // Zoltraak was made to kill demons; barriers stop it
	Barrier:     {Destruction: 2, Martial: 0.5},       // turns offensive magic back; cannot stop a blade
	Holy:        {Curse: 2, Barrier: 0.5},             // purifies demon magic
	Martial:     {Barrier: 2, Nature: 2, Curse: 0.5},  // physical blows break barriers
	Curse:       {Martial: 2, Nature: 2, Holy: 0.5, Barrier: 0.5},
	Nature:      {Barrier: 2, Holy: 2, Destruction: 0.5},
}

// Effectiveness is the damage multiplier for a move of type attack against a defender's types.
func Effectiveness(attack Type, defender []Type) float64 {
	m := 1.0
	for _, t := range defender {
		if x, ok := typeChart[attack][t]; ok {
			m *= x
		}
	}
	return m
}
