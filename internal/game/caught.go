package game

import (
	"math/rand/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// Caught is a freshly caught copy of a character, before it is saved.
type Caught struct {
	CharacterID string
	Level       int
	Nature      string     // a nature ID
	IVs         data.Stats // hidden rolls, 0..MaxIV each; higher is better
	Moves       []string   // move IDs, at most MaxMoves
}

// RollCaught creates a new copy with a random level, nature, hidden stat rolls and starting moves.
func RollCaught(r *rand.Rand, c data.Character) Caught {
	iv := func() int { return r.IntN(MaxIV + 1) } // a small helper closure, like an arrow function
	level := 1 + r.IntN(25)
	return Caught{
		CharacterID: c.ID,
		Level:       level,
		Nature:      data.Natures[r.IntN(len(data.Natures))].ID,
		IVs: data.Stats{
			HP: iv(), Attack: iv(), Defense: iv(),
			MagicAttack: iv(), MagicDefense: iv(), Speed: iv(),
		},
		Moves: StartingMoves(r, c, level),
	}
}
