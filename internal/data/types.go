package data

// Type is a character's or move's type. Battles use a type chart built on these.
type Type int

const (
	TypeDestruction Type = iota // offensive magic, like Zoltraak
	TypeBarrier                 // defensive magic
	TypeHoly                    // priests and healing
	TypeMartial                 // warriors and weapons
	TypeCurse                   // demon magic
	TypeNature                  // folk spells, flowers
)

// NoType is for typeless damage, like hitting yourself in confusion. It never gets a
// same-type bonus and is neutral against everything.
const NoType Type = -1

// AllTypes lists every type, in display order.
var AllTypes = []Type{TypeDestruction, TypeBarrier, TypeHoly, TypeMartial, TypeCurse, TypeNature}

func (t Type) String() string {
	switch t {
	case TypeDestruction:
		return "Destruction"
	case TypeBarrier:
		return "Barrier"
	case TypeHoly:
		return "Holy"
	case TypeMartial:
		return "Martial"
	case TypeCurse:
		return "Curse"
	case TypeNature:
		return "Nature"
	case NoType:
		return "Typeless"
	}
	return "Unknown"
}

// typeChart lists the matchups that are not neutral: attacking type -> defending type -> multiplier.
// 2 is super effective, 0.5 is not very effective; anything missing is 1.
var typeChart = map[Type]map[Type]float64{
	TypeDestruction: {TypeCurse: 2, TypeMartial: 2, TypeBarrier: 0.5}, // Zoltraak was made to kill demons; barriers stop it
	TypeBarrier:     {TypeDestruction: 2, TypeMartial: 0.5},           // turns offensive magic back; cannot stop a blade
	TypeHoly:        {TypeCurse: 2, TypeBarrier: 0.5},                 // purifies demon magic
	TypeMartial:     {TypeBarrier: 2, TypeNature: 2, TypeCurse: 0.5},  // physical blows break barriers
	TypeCurse:       {TypeMartial: 2, TypeNature: 2, TypeHoly: 0.5, TypeBarrier: 0.5},
	TypeNature:      {TypeBarrier: 2, TypeHoly: 2, TypeDestruction: 0.5},
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
