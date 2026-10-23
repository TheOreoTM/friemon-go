package game

import (
	"math/rand/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// MaxIV is the highest hidden roll for a single stat.
const MaxIV = 31

// Caught is a freshly caught copy of a character, before it is saved.
type Caught struct {
	CharacterID string
	Level       int
	IVs         data.Stats // hidden rolls, 0..MaxIV each; higher is better
	Moves       []string   // move IDs, at most MaxMoves
}

// RollCaught creates a new copy with a random level, hidden stat rolls and starting moves.
func RollCaught(r *rand.Rand, c data.Character) Caught {
	iv := func() int { return r.IntN(MaxIV + 1) } // a small helper closure, like an arrow function
	level := 1 + r.IntN(25)
	return Caught{
		CharacterID: c.ID,
		Level:       level,
		IVs: data.Stats{
			HP: iv(), Attack: iv(), Defense: iv(),
			MagicAttack: iv(), MagicDefense: iv(), Speed: iv(),
		},
		Moves: StartingMoves(r, c, level),
	}
}

// IVPercent is the overall quality of the rolls, 0 to 100.
func IVPercent(s data.Stats) float64 {
	sum := s.HP + s.Attack + s.Defense + s.MagicAttack + s.MagicDefense + s.Speed
	return float64(sum) / float64(6*MaxIV) * 100
}

// CalcStats works out a character's real stats from its base stats, hidden rolls and level.
// At level 100 a stat is roughly twice its base; the roll adds up to about 31 on top.
func CalcStats(base, ivs data.Stats, level int) data.Stats {
	stat := func(b, iv int) int { return (2*b+iv)*level/100 + 5 }
	return data.Stats{
		HP:           (2*base.HP+ivs.HP)*level/100 + level + 10,
		Attack:       stat(base.Attack, ivs.Attack),
		Defense:      stat(base.Defense, ivs.Defense),
		MagicAttack:  stat(base.MagicAttack, ivs.MagicAttack),
		MagicDefense: stat(base.MagicDefense, ivs.MagicDefense),
		Speed:        stat(base.Speed, ivs.Speed),
	}
}
