package game

import (
	"fmt"
	"math/rand/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// maxStage caps how far Defense and Magic Defense can be raised.
const maxStage = 6

// Fighter is one side of a battle.
type Fighter struct {
	Name  string // shown in the battle log, e.g. "Oreo's Fern"
	Level int
	Types []data.Type
	Stats data.Stats // final stats, from CalcStats
	HP    int
	Moves []data.Move

	defenseStage      int
	magicDefenseStage int
}

// NewFighter builds a fighter at full HP. Unknown move IDs are skipped.
func NewFighter(name string, c data.Character, level int, ivs data.Stats, moveIDs []string) *Fighter {
	f := &Fighter{
		Name:  name,
		Level: level,
		Types: c.Types,
		Stats: CalcStats(c.BaseStats, ivs, level),
	}
	f.HP = f.Stats.HP
	for _, id := range moveIDs {
		if m, ok := data.MoveByID(id); ok {
			f.Moves = append(f.Moves, m)
		}
	}
	return f
}

// Fainted reports whether the fighter is out of HP.
func (f *Fighter) Fainted() bool { return f.HP <= 0 }

// Battle is a 1v1 battle. It holds no Discord state, so it can be tested on its own.
type Battle struct {
	Fighters [2]*Fighter
	Turn     int // turns played so far
	Winner   int // index of the winner, or -1 while the battle is on
}

// NewBattle starts a battle between two fighters.
func NewBattle(a, b *Fighter) *Battle {
	return &Battle{Fighters: [2]*Fighter{a, b}, Winner: -1}
}

// Over reports whether someone has won.
func (b *Battle) Over() bool { return b.Winner >= 0 }

// Forfeit ends the battle with the other side winning.
func (b *Battle) Forfeit(loser int) {
	b.Winner = 1 - loser
}

// PlayTurn runs one turn. choices[i] is the index of the move fighter i picked.
// It returns what happened, one line per event, for the battle log.
func (b *Battle) PlayTurn(r *rand.Rand, choices [2]int) []string {
	if b.Over() {
		return nil
	}
	b.Turn++

	// The faster fighter goes first; a tie is a coin flip.
	first := 0
	s0, s1 := b.Fighters[0].Stats.Speed, b.Fighters[1].Stats.Speed
	if s1 > s0 || (s1 == s0 && r.IntN(2) == 1) {
		first = 1
	}

	var log []string
	for _, i := range []int{first, 1 - first} {
		attacker, defender := b.Fighters[i], b.Fighters[1-i]
		log = append(log, useMove(r, attacker, defender, attacker.Moves[choices[i]])...)
		if defender.Fainted() {
			log = append(log, fmt.Sprintf("%s fainted! **%s wins!**", defender.Name, attacker.Name))
			b.Winner = i
			break
		}
	}
	return log
}

// useMove applies one move and returns the log lines it produced.
func useMove(r *rand.Rand, attacker, defender *Fighter, m data.Move) []string {
	log := []string{fmt.Sprintf("%s used **%s**!", attacker.Name, m.Name)}
	if r.IntN(100) >= m.Accuracy {
		return append(log, "But it missed!")
	}

	if m.Category == data.Status {
		return append(log, applyEffect(attacker, m.Effect))
	}

	dmg, eff := Damage(r, attacker, defender, m)
	defender.HP = max(0, defender.HP-dmg)
	switch {
	case eff > 1:
		log = append(log, "It's super effective!")
	case eff < 1:
		log = append(log, "It's not very effective...")
	}
	return append(log, fmt.Sprintf("%s took %d damage.", defender.Name, dmg))
}

func applyEffect(f *Fighter, e data.Effect) string {
	switch e {
	case data.Heal:
		healed := min(f.Stats.HP/2, f.Stats.HP-f.HP)
		f.HP += healed
		return fmt.Sprintf("%s recovered %d HP.", f.Name, healed)
	case data.RaiseDefense:
		if f.defenseStage == maxStage {
			return fmt.Sprintf("%s's Defense cannot go any higher.", f.Name)
		}
		f.defenseStage++
		return fmt.Sprintf("%s's Defense rose!", f.Name)
	case data.RaiseMagicDefense:
		if f.magicDefenseStage == maxStage {
			return fmt.Sprintf("%s's Magic Defense cannot go any higher.", f.Name)
		}
		f.magicDefenseStage++
		return fmt.Sprintf("%s's Magic Defense rose!", f.Name)
	}
	return "Nothing happened... but the flowers are lovely."
}

// Damage works out how much damage m does, and the type multiplier it got.
// It uses the familiar Pokémon-style formula: level, power and attack over defense,
// then same-type bonus, type matchup and a small random spread.
func Damage(r *rand.Rand, attacker, defender *Fighter, m data.Move) (dmg int, effectiveness float64) {
	atk, def := attacker.Stats.Attack, stageStat(defender.Stats.Defense, defender.defenseStage)
	if m.Category == data.Magical {
		atk, def = attacker.Stats.MagicAttack, stageStat(defender.Stats.MagicDefense, defender.magicDefenseStage)
	}

	base := float64((2*attacker.Level/5+2)*m.Power*atk/max(def, 1))/50 + 2

	effectiveness = data.Effectiveness(m.Type, defender.Types)
	stab := 1.0 // "same type attack bonus" for using a move of your own type
	for _, t := range attacker.Types {
		if t == m.Type {
			stab = 1.5
		}
	}
	spread := 0.85 + r.Float64()*0.15

	dmg = int(base * stab * effectiveness * spread)
	return max(dmg, 1), effectiveness
}

// stageStat applies a stat stage: +1 is x1.5, +2 is x2, up to x4 at +6.
func stageStat(stat, stage int) int {
	return stat * (2 + stage) / 2
}
