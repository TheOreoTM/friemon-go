package game

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// fighter builds a test fighter from a real character with average rolls and a neutral nature.
func fighter(t *testing.T, id string, level int, moves ...string) *Fighter {
	t.Helper()
	c, ok := data.CharacterByID(id)
	if !ok {
		t.Fatalf("no character %q", id)
	}
	hardy, _ := data.NatureByID("hardy")
	ivs := data.Stats{HP: 15, Attack: 15, Defense: 15, MagicAttack: 15, MagicDefense: 15, Speed: 15}
	return NewFighter(FighterSpec{Character: c, Level: level, Nature: hardy, IVs: ivs, MoveIDs: moves})
}

// withMove gives a fighter a single custom move, for tests that must not depend on chance.
func withMove(f *Fighter, m data.Move) *Fighter {
	if m.PP == 0 {
		m.PP = 10
	}
	f.Moves = []*MoveSlot{{Move: m, PP: m.PP}}
	return f
}

func side(name string, fs ...*Fighter) *Side { return &Side{Player: name, Party: fs} }

func duel(a, b *Fighter) *Battle {
	bt, _ := NewBattle(side("A", a), side("B", b))
	return bt
}

func count(ev []Event, kind EventKind) int {
	n := 0
	for _, e := range ev {
		if e.Kind == kind {
			n++
		}
	}
	return n
}

// first returns the index of the first event of a kind, or -1.
func first(ev []Event, kind EventKind) int {
	for i, e := range ev {
		if e.Kind == kind {
			return i
		}
	}
	return -1
}

var (
	useMove0 = [2]Action{{Index: 0}, {Index: 0}}
	tackle   = data.Move{ID: "test_tackle", Name: "Tackle", Type: data.NoType, Category: data.Physical, Power: 40}
	nothing  = data.Move{ID: "test_wait", Name: "Wait", Type: data.NoType, Category: data.Status, PP: 99999,
		Effects: []data.MoveEffect{{Self: true, Stats: []data.StatChange{{Stat: data.StatEvasion, Stages: 1}}}}}
)

func play(t *testing.T, b *Battle, r *rand.Rand, actions [2]Action) []Event {
	t.Helper()
	ev, err := b.PlayTurn(r, actions)
	if err != nil {
		t.Fatalf("PlayTurn: %v", err)
	}
	return ev
}

func TestRandomBattlesAlwaysFinish(t *testing.T) {
	r := newRand()
	ids := []string{"frieren", "fern", "stark", "eisen", "himmel", "heiter", "sein", "ubel"}
	for game := range 300 {
		var sides [2]*Side
		for s := range sides {
			var party []*Fighter
			for range 1 + r.IntN(MaxParty) {
				c, _ := data.CharacterByID(ids[r.IntN(len(ids))])
				party = append(party, fighter(t, c.ID, 5+r.IntN(60), StartingMoves(r, c, 100)...))
			}
			sides[s] = side("P", party...)
		}
		b, _ := NewBattle(sides[0], sides[1])

		for turn := 0; !b.Over(); turn++ {
			if turn > 2000 {
				t.Fatalf("game %d did not finish", game)
			}
			for s := range b.Sides {
				if b.NeedsReplacement(s) {
					for i := range b.Sides[s].Party {
						if b.CanSwitchTo(s, i) {
							if _, err := b.Replace(s, i); err != nil {
								t.Fatal(err)
							}
							break
						}
					}
				}
			}
			var actions [2]Action
			for s := range actions {
				f := b.Active(s)
				actions[s] = Action{Index: r.IntN(len(f.Moves))}
				for f.HasPP() && f.Moves[actions[s].Index].PP == 0 {
					actions[s].Index = r.IntN(len(f.Moves))
				}
				if r.IntN(6) == 0 { // switch sometimes
					for i := range b.Sides[s].Party {
						if b.CanSwitchTo(s, i) {
							actions[s] = Action{Switch: true, Index: i}
						}
					}
				}
			}
			play(t, b, r, actions)
			for s := range b.Sides {
				for _, f := range b.Sides[s].Party {
					if f.HP < 0 || f.HP > f.Stats.HP {
						t.Fatalf("%s has %d/%d HP", f.Name, f.HP, f.Stats.HP)
					}
				}
			}
		}
		if b.Sides[b.Winner].Alive() == 0 && b.Sides[1-b.Winner].Alive() != 0 {
			t.Fatalf("winner %d has nobody left but the loser does", b.Winner)
		}
	}
}

func TestPriorityBeatsSpeed(t *testing.T) {
	r := newRand()
	for range 50 {
		eisen := fighter(t, "eisen", 50, "counter_blow") // slow, but +1 priority
		fern := fighter(t, "fern", 50, "zoltraak")       // fast
		ev := play(t, duel(fern, eisen), r, useMove0)
		if ev[0].Kind != EvMove || ev[0].Name != "Eisen" {
			t.Fatalf("first event %+v, want Eisen to move first", ev[0])
		}
	}
}

func TestSwitchHappensBeforeMoves(t *testing.T) {
	r := newRand()
	stark := fighter(t, "stark", 50, "slash")
	eisen := fighter(t, "eisen", 50, "slash")
	fern := withMove(fighter(t, "fern", 50), tackle)
	b, _ := NewBattle(side("A", stark, eisen), side("B", fern))

	ev := play(t, b, r, [2]Action{{Switch: true, Index: 1}, {Index: 0}})
	if ev[0].Kind != EvSwitchOut || ev[1].Kind != EvSwitchIn || ev[1].Name != "Eisen" {
		t.Fatalf("events %+v, want the switch first", ev[:2])
	}
	if eisen.HP == eisen.Stats.HP || stark.HP != stark.Stats.HP {
		t.Error("the attack should hit Eisen, who switched in, not Stark")
	}
}

func TestFaintedFighterMustBeReplaced(t *testing.T) {
	r := newRand()
	strong := withMove(fighter(t, "frieren", 100), data.Move{ID: "nuke", Name: "Nuke", Type: data.NoType, Category: data.Magical, Power: 250})
	weak1, weak2 := fighter(t, "sein", 5, "holy_binding"), fighter(t, "fern", 5, "zoltraak")
	b, _ := NewBattle(side("A", strong), side("B", weak1, weak2))

	ev := play(t, b, r, useMove0)
	if count(ev, EvFaint) != 1 || !b.NeedsReplacement(1) || b.Over() {
		t.Fatalf("after the knockout: faint events %d, needs replacement %v", count(ev, EvFaint), b.NeedsReplacement(1))
	}
	if len(b.KOs) != 1 || b.KOs[0].Scorer != strong || b.KOs[0].Fainted != weak1 {
		t.Errorf("KOs = %+v", b.KOs)
	}
	if _, err := b.PlayTurn(r, useMove0); !errors.Is(err, ErrNeedReplacement) {
		t.Errorf("playing before replacing: %v", err)
	}
	if _, err := b.Replace(1, 0); !errors.Is(err, ErrBadSwitch) {
		t.Errorf("replacing with the fainted fighter: %v", err)
	}
	if _, err := b.Replace(1, 1); err != nil {
		t.Fatal(err)
	}

	ev = play(t, b, r, useMove0)
	if !b.Over() || b.Winner != 0 || count(ev, EvWin) != 1 {
		t.Errorf("after the last knockout: over %v, winner %d", b.Over(), b.Winner)
	}
}

func TestValidation(t *testing.T) {
	a := fighter(t, "fern", 20, "zoltraak")
	a.Moves[0].PP = 0
	a.Moves = append(a.Moves, &MoveSlot{Move: tackle, PP: 5})
	b, _ := NewBattle(side("A", a), side("B", fighter(t, "stark", 20, "slash")))
	r := newRand()

	if _, err := b.PlayTurn(r, [2]Action{{Index: 0}, {Index: 0}}); !errors.Is(err, ErrBadMove) {
		t.Errorf("a move with no PP: %v", err)
	}
	if _, err := b.PlayTurn(r, [2]Action{{Index: 5}, {Index: 0}}); !errors.Is(err, ErrBadMove) {
		t.Errorf("a move slot that does not exist: %v", err)
	}
	if _, err := b.PlayTurn(r, [2]Action{{Switch: true, Index: 0}, {Index: 0}}); !errors.Is(err, ErrBadSwitch) {
		t.Errorf("switching to the active fighter: %v", err)
	}
	if b.Turn != 0 {
		t.Errorf("invalid actions still played %d turns", b.Turn)
	}
}

func TestStruggle(t *testing.T) {
	a := fighter(t, "stark", 50, "slash")
	a.Moves[0].PP = 0
	b := duel(a, withMove(fighter(t, "eisen", 50), nothing))

	ev := play(t, b, newRand(), useMove0)
	if count(ev, EvStruggle) != 1 || count(ev, EvRecoil) != 1 {
		t.Fatalf("events %+v, want a struggle with recoil", ev)
	}
	if lost := a.Stats.HP - a.HP; lost != a.Stats.HP/4 {
		t.Errorf("struggle recoil = %d, want %d", lost, a.Stats.HP/4)
	}
}

func TestEndOfTurnConditionDamage(t *testing.T) {
	r := newRand()
	for _, tt := range []struct {
		cond data.Condition
		want []int // sixteenths of max HP lost on each of three turns
	}{
		{data.Burn, []int{1, 1, 1}},
		{data.Cursed, []int{2, 2, 2}},
		{data.DeeplyCursed, []int{1, 2, 3}},
	} {
		f := withMove(fighter(t, "eisen", 100), nothing)
		f.Stats.HP, f.HP = 1600, 1600 // easy sixteenths
		inflict(r, f, tt.cond)
		b := duel(f, withMove(fighter(t, "stark", 100), nothing))
		for turn, want := range tt.want {
			before := f.HP
			play(t, b, r, useMove0)
			if lost := before - f.HP; lost != want*100 {
				t.Errorf("%s turn %d: lost %d, want %d", tt.cond, turn+1, lost, want*100)
			}
		}
	}
}

func TestDeepCurseResetsOnSwitch(t *testing.T) {
	r := newRand()
	f := withMove(fighter(t, "eisen", 100), nothing)
	inflict(r, f, data.DeeplyCursed)
	f.curseCounter = 5
	b, _ := NewBattle(side("A", f, fighter(t, "stark", 100, "slash")), side("B", withMove(fighter(t, "sein", 100), nothing)))
	play(t, b, r, [2]Action{{Switch: true, Index: 1}, {Index: 0}})
	if f.curseCounter != 1 || f.Condition != data.DeeplyCursed {
		t.Errorf("after switching out: counter %d, condition %s", f.curseCounter, f.Condition)
	}
}

func TestBurnHalvesPhysicalDamage(t *testing.T) {
	user, target := fighter(t, "stark", 50), fighter(t, "eisen", 50)
	normal, _ := damage(rand.New(rand.NewPCG(1, 1)), user, target, tackle, 1, false)
	user.Condition = data.Burn
	burned, _ := damage(rand.New(rand.NewPCG(1, 1)), user, target, tackle, 1, false)
	if burned > normal/2+1 || burned < normal/2-1 {
		t.Errorf("burned damage %d, normal %d; want about half", burned, normal)
	}
}

func TestSleepLastsOneToThreeTurns(t *testing.T) {
	r := newRand()
	seen := map[int]bool{}
	for range 200 {
		f := withMove(fighter(t, "fern", 50), nothing)
		inflict(r, f, data.Sleep)
		b := duel(f, withMove(fighter(t, "stark", 50), nothing))
		asleep := 0
		for f.Condition == data.Sleep {
			ev := play(t, b, r, useMove0)
			asleep += count(ev, EvFastAsleep)
			if asleep > 3 {
				t.Fatal("slept for more than 3 turns")
			}
		}
		seen[asleep] = true
	}
	if !seen[1] || !seen[2] || !seen[3] || seen[0] {
		t.Errorf("sleep lengths seen: %v, want exactly 1, 2 and 3", seen)
	}
}

func TestRestSleepsExactlyTwoTurns(t *testing.T) {
	r := newRand()
	f := fighter(t, "heiter", 50, "blessed_rest")
	f.HP = 1
	b := duel(f, withMove(fighter(t, "stark", 50), nothing))
	ev := play(t, b, r, useMove0)
	if count(ev, EvRested) != 1 || f.HP != f.Stats.HP {
		t.Fatalf("rest did not heal: %+v", ev)
	}
	asleep := 0
	for range 3 {
		asleep += count(play(t, b, r, useMove0), EvFastAsleep)
	}
	if asleep != 2 || f.Condition == data.Sleep {
		t.Errorf("asleep %d turns, condition now %s; want 2 turns then awake", asleep, f.Condition)
	}
}

func TestBound(t *testing.T) {
	r := newRand()
	f := withMove(fighter(t, "fern", 50), nothing)
	speed := f.speed()
	inflict(r, f, data.Bound)
	if f.speed() != speed/2 {
		t.Errorf("bound speed %d, want %d", f.speed(), speed/2)
	}
	b := duel(f, withMove(fighter(t, "eisen", 50), nothing))
	stuck := 0
	const turns = 2000
	for range turns {
		stuck += count(play(t, b, r, useMove0), EvFullyBound)
	}
	if pct := stuck * 100 / turns; pct < 20 || pct > 30 {
		t.Errorf("fully bound %d%% of turns, want about 25%%", pct)
	}
}

func TestConfusionLastsOneToFourTurns(t *testing.T) {
	r := newRand()
	for range 200 {
		f := withMove(fighter(t, "eisen", 100), nothing)
		f.confusionTurns = 2 + r.IntN(4)
		b := duel(f, withMove(fighter(t, "stark", 100), nothing))
		confused := 0
		for f.confusionTurns > 0 {
			confused += count(play(t, b, r, useMove0), EvIsConfused)
		}
		if confused < 1 || confused > 4 {
			t.Fatalf("confused for %d turns", confused)
		}
	}
}

func TestFlinchOnlyWhenTargetHasNotMoved(t *testing.T) {
	r := newRand()
	flinchHit := data.Move{ID: "test_flinch", Name: "Flinch Hit", Type: data.NoType, Category: data.Physical, Power: 10,
		Effects: []data.MoveEffect{{Volatile: data.Flinch}}}

	// The faster fighter makes the slower one flinch.
	fast, slow := withMove(fighter(t, "fern", 50), flinchHit), withMove(fighter(t, "eisen", 50), tackle)
	if ev := play(t, duel(fast, slow), r, useMove0); count(ev, EvFlinched) != 1 {
		t.Errorf("faster user: %d flinches, want 1", count(ev, EvFlinched))
	}
	// The slower fighter's flinch comes too late to matter.
	fast, slow = withMove(fighter(t, "fern", 50), tackle), withMove(fighter(t, "eisen", 50), flinchHit)
	b := duel(fast, slow)
	if ev := play(t, b, r, useMove0); count(ev, EvFlinched) != 0 {
		t.Error("slower user made the target flinch after it had already moved")
	}
	if ev := play(t, b, r, useMove0); count(ev, EvFlinched) != 0 {
		t.Error("a flinch carried over to the next turn")
	}
}

func TestStatStages(t *testing.T) {
	r := newRand()
	f := fighter(t, "stark", 50, "battle_stance")
	f.Moves[0].PP = 99
	b := duel(f, withMove(fighter(t, "eisen", 50), nothing))
	for range 3 { // +2 three times reaches the +6 cap
		play(t, b, r, useMove0)
	}
	if f.Stage(data.StatAttack) != maxStage {
		t.Fatalf("attack stage %d, want %d", f.Stage(data.StatAttack), maxStage)
	}
	ev := play(t, b, r, useMove0)
	i := first(ev, EvStatChange)
	if i < 0 || ev[i].Amount != 0 {
		t.Errorf("at the cap the move should report no change: %+v", ev)
	}

	for stage, want := range map[int]int{0: 100, 1: 150, 2: 200, 6: 400, -1: 66, -2: 50, -6: 25} {
		if got := applyStage(100, stage); got != want {
			t.Errorf("applyStage(100, %d) = %d, want %d", stage, got, want)
		}
	}
}

func TestStatusMovesFailLoudly(t *testing.T) {
	r := newRand()
	healer := fighter(t, "heiter", 50, "healing_magic")
	ev := play(t, duel(healer, withMove(fighter(t, "eisen", 50), nothing)), r, useMove0)
	if count(ev, EvFail) != 1 {
		t.Errorf("healing at full HP: %+v, want a failure", ev)
	}

	sleeper := withMove(fighter(t, "fern", 50), data.Move{ID: "test_sleep", Name: "Sleep", Type: data.NoType, Category: data.Status,
		Effects: []data.MoveEffect{{Status: data.Sleep}}})
	target := withMove(fighter(t, "eisen", 50), nothing)
	b := duel(sleeper, target)
	play(t, b, r, useMove0)
	ev = play(t, b, r, useMove0)
	if count(ev, EvConditionFailed) != 1 {
		t.Errorf("sleeping an already sleeping target: %+v", ev)
	}
}

func TestCurseTypesAreImmuneToCurses(t *testing.T) {
	r := newRand()
	demon := withMove(fighter(t, "eisen", 50), nothing)
	demon.Types = []data.Type{data.TypeCurse}
	ev := play(t, duel(fighter(t, "ubel", 50, "deep_curse"), demon), r, useMove0)
	if demon.Condition != data.NoCondition || (count(ev, EvImmune) == 0 && count(ev, EvMiss) == 0) {
		t.Errorf("a Curse type got %s: %+v", demon.Condition, ev)
	}
}

func TestDrainAndRecoil(t *testing.T) {
	r := newRand()
	drainer := fighter(t, "ubel", 50, "soul_drain")
	drainer.HP = 1
	ev := play(t, duel(drainer, withMove(fighter(t, "eisen", 50), nothing)), r, useMove0)
	dmg, drained := ev[first(ev, EvDamage)].Amount, ev[first(ev, EvDrain)].Amount
	if drained != max(1, dmg/2) {
		t.Errorf("drained %d from %d damage, want half", drained, dmg)
	}

	rammer := fighter(t, "stark", 50, "axe_cleave")
	ev = play(t, duel(rammer, withMove(fighter(t, "eisen", 50), nothing)), r, useMove0)
	if i := first(ev, EvRecoil); i < 0 || ev[i].Amount != max(1, ev[first(ev, EvDamage)].Amount*33/100) {
		t.Errorf("recoil events %+v", ev)
	}
}

func TestMultiHit(t *testing.T) {
	r := newRand()
	counts := map[int]int{}
	for range 1000 {
		counts[rollHits(r, 2, 5)]++
	}
	if counts[2] < 300 || counts[3] < 300 || counts[4] < 100 || counts[5] < 100 || counts[1]+counts[6] > 0 {
		t.Errorf("hit counts %v, want roughly 35/35/15/15%%", counts)
	}

	// Hits stop once the target faints.
	lash := data.Move{ID: "test_lash", Name: "Lash", Type: data.NoType, Category: data.Physical, Power: 25, MinHits: 5, MaxHits: 5}
	target := withMove(fighter(t, "sein", 50), nothing)
	target.HP = 1
	ev := play(t, duel(withMove(fighter(t, "stark", 50), lash), target), r, useMove0)
	if count(ev, EvDamage) != 1 || ev[first(ev, EvHits)].Amount != 1 {
		t.Errorf("hitting a fainted target: %+v", ev)
	}
}

func TestCritStages(t *testing.T) {
	r := newRand()
	for range 100 {
		if !rollCrit(r, 3) {
			t.Fatal("crit stage 3 should always crit")
		}
	}
}
