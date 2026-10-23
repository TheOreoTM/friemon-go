package data

// Stats is one number per battle stat. It is used for base stats, IVs, EVs and final stats.
type Stats struct {
	HP           int
	Attack       int
	Defense      int
	MagicAttack  int
	MagicDefense int
	Speed        int
}

// Stat names one stat. Accuracy and Evasion only exist as battle stages.
type Stat int

const (
	StatHP Stat = iota
	StatAttack
	StatDefense
	StatMagicAttack
	StatMagicDefense
	StatSpeed
	StatAccuracy
	StatEvasion
)

// MainStats are the six stats every character has, in display order.
var MainStats = []Stat{StatHP, StatAttack, StatDefense, StatMagicAttack, StatMagicDefense, StatSpeed}

func (s Stat) String() string {
	switch s {
	case StatHP:
		return "HP"
	case StatAttack:
		return "Attack"
	case StatDefense:
		return "Defense"
	case StatMagicAttack:
		return "Magic Attack"
	case StatMagicDefense:
		return "Magic Defense"
	case StatSpeed:
		return "Speed"
	case StatAccuracy:
		return "Accuracy"
	case StatEvasion:
		return "Evasion"
	}
	return "Unknown"
}

// Get returns one stat by name. Accuracy and Evasion return 0.
func (s Stats) Get(stat Stat) int {
	switch stat {
	case StatHP:
		return s.HP
	case StatAttack:
		return s.Attack
	case StatDefense:
		return s.Defense
	case StatMagicAttack:
		return s.MagicAttack
	case StatMagicDefense:
		return s.MagicDefense
	case StatSpeed:
		return s.Speed
	}
	return 0
}

// Set changes one stat by name. A pointer receiver (*Stats) lets the method change
// the caller's value; a plain Stats receiver would only change a copy.
func (s *Stats) Set(stat Stat, v int) {
	switch stat {
	case StatHP:
		s.HP = v
	case StatAttack:
		s.Attack = v
	case StatDefense:
		s.Defense = v
	case StatMagicAttack:
		s.MagicAttack = v
	case StatMagicDefense:
		s.MagicDefense = v
	case StatSpeed:
		s.Speed = v
	}
}

// Total adds up all six stats.
func (s Stats) Total() int {
	return s.HP + s.Attack + s.Defense + s.MagicAttack + s.MagicDefense + s.Speed
}
