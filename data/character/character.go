package data

import "log/slog"

type IBaseCharacter interface {
	GetBaseStats() (int, int, int, int, int, int)
}

type BaseCharacter struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	HP    int    `json:"hp"`
	Atk   int    `json:"atk"`
	Def   int    `json:"def"`
	SpAtk int    `json:"satk"`
	SpDef int    `json:"sdef"`
	Spd   int    `json:"spd"`

	Type0 string `json:"type0"`
	Type1 string `json:"type1"`
}

func (bc BaseCharacter) Disabled() {
	delete(Characters, bc.ID)
}

// GetBaseStats retrieves the character's base stats.
func (bc BaseCharacter) GetBaseStats() (int, int, int, int, int, int) {
	return bc.HP, bc.Atk, bc.Def, bc.SpAtk, bc.SpDef, bc.Spd
}

// NewBaseCharacter is a constructor for creating a new BaseCharacter.
func NewBaseCharacter(id int, name string, types []string, hp, atk, def, spAtk, spDef, spd int) BaseCharacter {
	for _, ch := range Characters {
		if ch.ID == id {
			slog.Error("Character already exists, character wasnt added", slog.Int("existing_id", ch.ID), slog.Int("new_id", id))
			return BaseCharacter{}
		}

		if ch.Name == name {
			slog.Error("Character already exists, character wasnt added", slog.String("existing_name", ch.Name), slog.String("new_name", name))
			return BaseCharacter{}
		}
	}

	c := BaseCharacter{
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

	Characters[id] = c
	return c
}

var Characters = map[int]BaseCharacter{}
