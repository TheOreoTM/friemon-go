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
