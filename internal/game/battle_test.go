package game

import (
	"strings"
	"testing"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// fighter builds a test fighter from a real character with average rolls.
func fighter(t *testing.T, id string, level int, moves ...string) *Fighter {
	t.Helper()
	c, ok := data.CharacterByID(id)
	if !ok {
		t.Fatalf("no character %q", id)
	}
	ivs := data.Stats{HP: 15, Attack: 15, Defense: 15, MagicAttack: 15, MagicDefense: 15, Speed: 15}
	return NewFighter(id, c, level, ivs, moves)
}

func TestBattleAlwaysEnds(t *testing.T) {
	r := newRand()
	for range 200 {
		a := fighter(t, "fern", 30, "zoltraak", "defensive_magic")
		b := fighter(t, "stark", 30, "slash", "iron_body")
		battle := NewBattle(a, b)
		for turn := 0; !battle.Over(); turn++ {
			if turn > 500 {
				t.Fatal("battle did not end after 500 turns")
			}
			battle.PlayTurn(r, [2]int{r.IntN(len(a.Moves)), r.IntN(len(b.Moves))})
		}
		loser := battle.Fighters[1-battle.Winner]
		if !loser.Fainted() || battle.Fighters[battle.Winner].Fainted() {
			t.Fatalf("winner %d, but HP is %d vs %d", battle.Winner, a.HP, b.HP)
		}
	}
}

func TestFasterFighterGoesFirst(t *testing.T) {
	// Fern is much faster than Eisen, so her move is always logged first.
	r := newRand()
	for range 50 {
		fern := fighter(t, "fern", 50, "zoltraak")
		eisen := fighter(t, "eisen", 50, "slash")
		log := NewBattle(eisen, fern).PlayTurn(r, [2]int{0, 0})
		if !strings.HasPrefix(log[0], "fern used") {
			t.Fatalf("first line was %q, want Fern to move first", log[0])
		}
	}
}

func TestDamageUsesTypes(t *testing.T) {
	r := newRand()
	zoltraak, _ := data.MoveByID("zoltraak") // Destruction: strong vs Martial
	fern := fighter(t, "fern", 50)
	stark := fighter(t, "stark", 50)   // Martial
	heiter := fighter(t, "heiter", 50) // Holy, neutral

	_, effStark := Damage(r, fern, stark, zoltraak)
	_, effHeiter := Damage(r, fern, heiter, zoltraak)
	if effStark != 2 || effHeiter != 1 {
		t.Errorf("effectiveness vs Stark = %v, vs Heiter = %v; want 2 and 1", effStark, effHeiter)
	}
}

func TestDefenseStageReducesDamage(t *testing.T) {
	slash, _ := data.MoveByID("slash")
	stark := fighter(t, "stark", 50)

	before := fighter(t, "eisen", 50)
	after := fighter(t, "eisen", 50)
	applyEffect(after, data.RaiseDefense)
	applyEffect(after, data.RaiseDefense)

	// Same seed for both, so only the stage differs.
	d1, _ := Damage(newRand(), stark, before, slash)
	d2, _ := Damage(newRand(), stark, after, slash)
	if d2 >= d1 {
		t.Errorf("damage after +2 Defense = %d, before = %d; want less", d2, d1)
	}
}

func TestStagesAreCapped(t *testing.T) {
	f := fighter(t, "eisen", 50)
	for range maxStage + 3 {
		applyEffect(f, data.RaiseDefense)
	}
	if f.defenseStage != maxStage {
		t.Errorf("defense stage = %d, want %d", f.defenseStage, maxStage)
	}
}

func TestHealNeverOverfills(t *testing.T) {
	f := fighter(t, "sein", 50)
	f.HP = f.Stats.HP - 3
	applyEffect(f, data.Heal)
	if f.HP != f.Stats.HP {
		t.Errorf("HP after heal = %d, want max %d", f.HP, f.Stats.HP)
	}
}

func TestForfeit(t *testing.T) {
	b := NewBattle(fighter(t, "fern", 5, "zoltraak"), fighter(t, "stark", 5, "slash"))
	b.Forfeit(0)
	if !b.Over() || b.Winner != 1 {
		t.Errorf("after player 0 forfeits, winner = %d", b.Winner)
	}
	if log := b.PlayTurn(newRand(), [2]int{0, 0}); log != nil {
		t.Errorf("a finished battle played another turn: %v", log)
	}
}
