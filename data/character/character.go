package data

type IBaseCharacter interface {
	GetBaseStats() (int, int, int, int, int, int)
}

type BaseCharacter struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	HP    int32  `json:"hp"`
	Atk   int32  `json:"atk"`
	Def   int32  `json:"def"`
	SpAtk int32  `json:"satk"`
	SpDef int32  `json:"sdef"`
	Spd   int32  `json:"spd"`

	Type0 string `json:"type0"`
	Type1 string `json:"type1"`
}

// GetBaseStats retrieves the character's base stats.
func (bc BaseCharacter) GetBaseStats() (int32, int32, int32, int32, int32, int32) {
	return bc.HP, bc.Atk, bc.Def, bc.SpAtk, bc.SpDef, bc.Spd
}

// NewBaseCharacter is a constructor for creating a new BaseCharacter.
func NewBaseCharacter(id int32, name string, types []string, hp, atk, def, spAtk, spDef, spd int32) BaseCharacter {
	return BaseCharacter{
		ID:    id,
		Name:  name,
		HP:    hp,
		Atk:   atk,
		Def:   def,
		SpAtk: spAtk,
		SpDef: spDef,
		Spd:   spd,

		Type0: types[0],
		Type1: types[1],
	}
}

func EnabledCharacters() []BaseCharacter {
	return []BaseCharacter{
		Frieren,
		Himmel,
		Eisen,
		Heiter,
		Stark,
		Fern,
		Flamme,
		Serie,
		Sein,
	}
}
