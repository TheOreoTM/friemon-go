package game

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

func TestAddXP(t *testing.T) {
	tests := []struct {
		name              string
		level, xp, gain   int
		wantLevel, wantXP int
	}{
		{"no level up", 1, 0, 50, 1, 50},
		{"exactly enough", 1, 0, 75, 2, 0},
		{"carries the rest", 1, 70, 20, 2, 15},
		{"several levels at once", 1, 0, 75 + 100 + 10, 3, 10},
		{"stops at max", MaxLevel - 1, 0, 1_000_000, MaxLevel, 0},
		{"nothing past max", MaxLevel, 0, 500, MaxLevel, 0},
	}
	for _, tt := range tests {
		// t.Run makes each case its own sub-test, named in the output if it fails.
		t.Run(tt.name, func(t *testing.T) {
			level, xp := AddXP(tt.level, tt.xp, tt.gain)
			if level != tt.wantLevel || xp != tt.wantXP {
				t.Errorf("AddXP(%d, %d, %d) = %d, %d; want %d, %d",
					tt.level, tt.xp, tt.gain, level, xp, tt.wantLevel, tt.wantXP)
			}
		})
	}
}

func TestCalcStats(t *testing.T) {
	base := data.Stats{HP: 100, Attack: 100, Defense: 100, MagicAttack: 100, MagicDefense: 100, Speed: 100}
	zero := data.Stats{}
	hardy, _ := data.NatureByID("hardy")     // neutral
	adamant, _ := data.NatureByID("adamant") // +Attack, -Magic Attack

	// At level 100 with no IVs or EVs, HP = 200+100+10 and others = 200+5.
	s := CalcStats(base, zero, zero, 100, hardy)
	if s.HP != 310 || s.Attack != 205 || s.Speed != 205 {
		t.Errorf("level 100 stats = %+v", s)
	}
	// A nature changes its two stats by 10%, rounding down.
	s = CalcStats(base, zero, zero, 100, adamant)
	if s.Attack != 225 || s.MagicAttack != 184 || s.Defense != 205 {
		t.Errorf("adamant stats = %+v", s)
	}
	// 252 EVs add 63 points at level 100.
	evs := data.Stats{Speed: 252}
	if got := CalcStats(base, zero, evs, 100, hardy).Speed; got != 268 {
		t.Errorf("Speed with 252 EVs = %d, want 268", got)
	}
	// Higher rolls and levels never make a stat lower.
	perfect := data.Stats{HP: MaxIV, Attack: MaxIV, Defense: MaxIV, MagicAttack: MaxIV, MagicDefense: MaxIV, Speed: MaxIV}
	if CalcStats(base, perfect, zero, 50, hardy).Attack <= CalcStats(base, zero, zero, 50, hardy).Attack {
		t.Error("perfect rolls did not raise Attack")
	}
	if CalcStats(base, zero, zero, 51, hardy).Attack < CalcStats(base, zero, zero, 50, hardy).Attack {
		t.Error("a higher level lowered Attack")
	}
}

func TestAddEVs(t *testing.T) {
	evs := AddEVs(data.Stats{Attack: 250}, data.Stats{Attack: 3, Speed: 2})
	if evs.Attack != MaxEVPerStat || evs.Speed != 2 {
		t.Errorf("per-stat cap: %+v", evs)
	}
	full := data.Stats{HP: 252, Attack: 252, Defense: 5} // 509 total
	evs = AddEVs(full, data.Stats{Speed: 3})
	if evs.Total() != MaxEVTotal || evs.Speed != 1 {
		t.Errorf("total cap: %+v", evs)
	}
}

func TestStartingMoves(t *testing.T) {
	r := newRand()
	for _, c := range data.Characters {
		for _, level := range []int{1, 25, 100} {
			moves := StartingMoves(r, c, level)
			if len(moves) == 0 || len(moves) > MaxMoves {
				t.Fatalf("%s at %d got %d moves", c.ID, level, len(moves))
			}
			learnable := LearnableMoves(c, level)
			for _, m := range moves {
				if !slices.Contains(learnable, m) {
					t.Fatalf("%s at %d got %s, which it cannot learn yet", c.ID, level, m)
				}
			}
		}
	}
}

func TestStartingMovesDoesNotChangeLearnset(t *testing.T) {
	frieren, _ := data.CharacterByID("frieren")
	before := slices.Clone(frieren.Learnset)
	StartingMoves(newRand(), frieren, 100)
	if !slices.Equal(before, frieren.Learnset) {
		t.Fatal("StartingMoves shuffled the character's learnset")
	}
}

func TestCheckLearn(t *testing.T) {
	fern, _ := data.CharacterByID("fern")
	tests := []struct {
		level int
		known []string
		move  string
		want  error
	}{
		{12, []string{"zoltraak"}, "rapid_barrage", nil},
		{11, []string{"zoltraak"}, "rapid_barrage", ErrLevelTooLow},
		{50, []string{"zoltraak"}, "zoltraak", ErrAlreadyKnown},
		{50, nil, "axe_cleave", ErrNotInSet},
		{50, nil, "nope", ErrUnknownMove},
	}
	for _, tt := range tests {
		// errors.Is compares against the exported error values.
		if err := CheckLearn(fern, tt.level, tt.known, tt.move); !errors.Is(err, tt.want) {
			t.Errorf("CheckLearn(fern, %d, %v, %s) = %v, want %v", tt.level, tt.known, tt.move, err, tt.want)
		}
	}
}

func TestCooldown(t *testing.T) {
	c := NewCooldown(10 * time.Second)
	now := time.Now()
	if !c.Allow("a", now) {
		t.Fatal("first use was blocked")
	}
	if c.Allow("a", now.Add(5*time.Second)) {
		t.Fatal("allowed during cooldown")
	}
	if !c.Allow("b", now) {
		t.Fatal("one key's cooldown blocked another")
	}
	if !c.Allow("a", now.Add(10*time.Second)) {
		t.Fatal("blocked after cooldown ended")
	}
}
