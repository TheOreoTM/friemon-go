package game

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// moveByID fetches a real move for a test, failing if it does not exist.
func moveByID(t *testing.T, id string) data.Move {
	t.Helper()
	m, ok := data.MoveByID(id)
	if !ok {
		t.Fatalf("no move %q", id)
	}
	m.PP = 99
	return m
}

// withMoves gives a fighter several real moves.
func withMoves(t *testing.T, f *Fighter, ids ...string) *Fighter {
	f.Moves = nil
	for _, id := range ids {
		m := moveByID(t, id)
		f.Moves = append(f.Moves, &MoveSlot{Move: m, PP: m.PP})
	}
	return f
}

func waiter(t *testing.T, id string) *Fighter { return withMove(fighter(t, id, 50), nothing) }

// sameSeedDamage compares the damage of a move with and without a battle change,
// using the same random numbers so only the change differs.
func sameSeedDamage(b *Battle, user, target *Fighter, m data.Move, change func()) (before, after int) {
	eff := data.Effectiveness(m.Type, target.Types)
	before, _ = b.damage(rand.New(rand.NewPCG(7, 7)), user, target, 1, m, eff, false, true)
	change()
	after, _ = b.damage(rand.New(rand.NewPCG(7, 7)), user, target, 1, m, eff, false, true)
	return before, after
}

func TestWeatherLastsFiveTurns(t *testing.T) {
	r := newRand()
	b := duel(withMoves(t, fighter(t, "fern", 50), "heatwave"), waiter(t, "eisen"))
	ev := play(t, b, r, useMove0)
	if b.Weather != data.Heatwave || count(ev, EvWeatherStart) != 1 {
		t.Fatalf("weather %s after Heatwave", b.Weather)
	}
	if ev := play(t, b, r, useMove0); count(ev, EvFail) != 1 {
		t.Error("starting the same weather again should fail")
	}
	for range 3 {
		play(t, b, r, [2]Action{{Index: 0}, {Index: 0}})
	}
	if b.Weather != data.NoWeather {
		t.Errorf("weather %s still up after 5 turns", b.Weather)
	}
}

func TestWeatherAndTerrainChangeDamage(t *testing.T) {
	fern, stark := fighter(t, "fern", 50), fighter(t, "stark", 50)
	b := duel(fern, stark)
	zoltraak := moveByID(t, "zoltraak") // Destruction

	before, after := sameSeedDamage(b, fern, stark, zoltraak, func() { b.Weather = data.Heatwave })
	if after != before*3/2 && after != before*3/2+1 && after != before*3/2-1 {
		t.Errorf("Heatwave: %d -> %d, want about x1.5", before, after)
	}
	before, after = sameSeedDamage(b, fern, stark, zoltraak, func() { b.Weather = data.Storm })
	if after > before/2+1 {
		t.Errorf("Storm should halve Destruction: %d -> %d", before, after)
	}
	b.Weather = data.NoWeather
	before, after = sameSeedDamage(b, fern, stark, zoltraak, func() { b.Terrain = data.ManaField })
	if after <= before {
		t.Errorf("Mana Field should boost Destruction: %d -> %d", before, after)
	}

	b.Terrain = data.NoTerrain
	splitter := moveByID(t, "ground_splitter")
	before, after = sameSeedDamage(b, stark, fern, splitter, func() { b.Terrain = data.Overgrowth })
	if after > before/2+1 {
		t.Errorf("Overgrowth should halve ground moves: %d -> %d", before, after)
	}
}

func TestWeatherChipAndShelter(t *testing.T) {
	r := newRand()
	stark, fern := waiter(t, "stark"), waiter(t, "fern") // Martial, Destruction
	b := duel(stark, fern)
	b.Weather, b.WeatherTurns = data.Sandstorm, 5
	play(t, b, r, useMove0)
	if stark.HP != stark.Stats.HP {
		t.Error("Sandstorm hurt a Martial type")
	}
	if lost := fern.Stats.HP - fern.HP; lost != fern.Stats.HP/16 {
		t.Errorf("Sandstorm took %d from Fern, want %d", lost, fern.Stats.HP/16)
	}

	// Sandstorm also raises Martial types' Magic Defense.
	zoltraak := moveByID(t, "zoltraak")
	b.Weather = data.NoWeather
	before, after := sameSeedDamage(b, fern, stark, zoltraak, func() { b.Weather = data.Sandstorm })
	if after >= before {
		t.Errorf("Sandstorm should cut magic damage to Martial types: %d -> %d", before, after)
	}
}

func TestOvergrowthHeals(t *testing.T) {
	r := newRand()
	f := waiter(t, "eisen")
	f.HP = 1
	b := duel(f, waiter(t, "stark"))
	b.Terrain, b.TerrainTurns = data.Overgrowth, 5
	play(t, b, r, useMove0)
	if f.HP != 1+f.Stats.HP/16 {
		t.Errorf("HP after Overgrowth = %d, want %d", f.HP, 1+f.Stats.HP/16)
	}
}

func TestFieldBlocksConditions(t *testing.T) {
	r := newRand()
	tests := []struct {
		name    string
		setup   func(b *Battle)
		cond    data.Condition
		blocked bool
	}{
		{"mist blocks burns", func(b *Battle) { b.Terrain = data.Mist }, data.Burn, true},
		{"mana field blocks sleep", func(b *Battle) { b.Terrain = data.ManaField }, data.Sleep, true},
		{"mana field allows burns", func(b *Battle) { b.Terrain = data.ManaField }, data.Burn, false},
		{"heatwave blocks freezing", func(b *Battle) { b.Weather = data.Heatwave }, data.Freeze, true},
		{"storm blocks burns", func(b *Battle) { b.Weather = data.Storm }, data.Burn, true},
	}
	for _, tt := range tests {
		inflicter := withMove(fighter(t, "fern", 50), data.Move{ID: "test_inflict", Name: "Inflict",
			Type: data.NoType, Category: data.Status, Effects: []data.MoveEffect{{Status: tt.cond}}})
		target := waiter(t, "stark")
		b := duel(inflicter, target)
		tt.setup(b)
		b.WeatherTurns, b.TerrainTurns = 5, 5
		ev := play(t, b, r, useMove0)
		if got := target.Condition == data.NoCondition; got != tt.blocked {
			t.Errorf("%s: condition %s, events %+v", tt.name, target.Condition, ev)
		}
	}

	// Mist also stops confusion.
	b := duel(withMoves(t, fighter(t, "ubel", 50), "illusion"), waiter(t, "stark"))
	b.Terrain, b.TerrainTurns = data.Mist, 5
	if ev := play(t, b, r, useMove0); count(ev, EvConfused) != 0 || count(ev, EvFieldBlocked) != 1 {
		t.Errorf("Mist did not stop confusion: %+v", ev)
	}
}

func TestSacredGroundBlocksPriority(t *testing.T) {
	r := newRand()
	target := waiter(t, "fern")
	b := duel(withMoves(t, fighter(t, "eisen", 50), "counter_blow"), target)
	b.Terrain, b.TerrainTurns = data.SacredGround, 5
	ev := play(t, b, r, useMove0)
	if target.HP != target.Stats.HP || count(ev, EvFieldBlocked) != 1 {
		t.Errorf("a priority move hit through Sacred Ground: %+v", ev)
	}
}

func TestWards(t *testing.T) {
	stark, eisen := fighter(t, "stark", 50), fighter(t, "eisen", 50)
	b := duel(stark, eisen)
	slash := moveByID(t, "slash")
	before, after := sameSeedDamage(b, stark, eisen, slash, func() { b.Sides[1].turns[data.PhysicalWard] = 5 })
	if after > before/2+1 {
		t.Errorf("Physical Ward: %d -> %d, want half", before, after)
	}
	zoltraak := moveByID(t, "zoltraak")
	before, after = sameSeedDamage(b, stark, eisen, zoltraak, func() {})
	if before != after {
		t.Error("Physical Ward should not affect magical damage")
	}

	// Critical hits go straight through wards.
	eff := data.Effectiveness(slash.Type, eisen.Types)
	slash.CritStage = 3
	warded, _ := b.damage(rand.New(rand.NewPCG(7, 7)), stark, eisen, 1, slash, eff, true, true)
	b.Sides[1].turns[data.PhysicalWard] = 0
	open, _ := b.damage(rand.New(rand.NewPCG(7, 7)), stark, eisen, 1, slash, eff, true, true)
	if warded != open {
		t.Errorf("crit through a ward did %d, without %d", warded, open)
	}

	// A ward wears off after 5 turns.
	r := newRand()
	b = duel(withMoves(t, fighter(t, "heiter", 50), "physical_ward"), waiter(t, "stark"))
	play(t, b, r, useMove0)
	for range 4 {
		play(t, b, r, [2]Action{{Index: 0}, {Index: 0}})
	}
	if b.Sides[0].TurnsLeft(data.PhysicalWard) != 0 {
		t.Errorf("ward still has %d turns", b.Sides[0].TurnsLeft(data.PhysicalWard))
	}
}

func TestTailwindAndReversal(t *testing.T) {
	r := newRand()
	stark, fern := withMove(fighter(t, "stark", 50), tackle), withMove(fighter(t, "fern", 50), tackle) // Fern is faster
	b := duel(stark, fern)
	b.Sides[0].turns[data.Tailwind] = 4
	if b.speed(0) <= b.speed(1) {
		t.Fatalf("Tailwind Stark %d should outspeed Fern %d", b.speed(0), b.speed(1))
	}

	b.Sides[0].turns[data.Tailwind] = 0
	b.ReversalTurns = 5
	ev := play(t, b, r, useMove0)
	if ev[0].Name != "Stark" {
		t.Errorf("under Reversal Field the slower Stark should move first, got %+v", ev[0])
	}

	// Using Reversal Field again ends it.
	b = duel(withMoves(t, fighter(t, "ubel", 50), "reversal_field"), waiter(t, "stark"))
	play(t, b, r, useMove0)
	if ev := play(t, b, r, useMove0); b.ReversalTurns != 0 || count(ev, EvReversalEnd) != 1 {
		t.Errorf("second Reversal Field should end it; turns %d", b.ReversalTurns)
	}
}

func TestProtect(t *testing.T) {
	r := newRand()
	protector := withMoves(t, fighter(t, "eisen", 50), "full_barrier")
	b := duel(protector, withMove(fighter(t, "stark", 50), tackle))
	ev := play(t, b, r, useMove0)
	if protector.HP != protector.Stats.HP || count(ev, EvProtected) != 1 {
		t.Fatalf("Full Barrier did not block: %+v", ev)
	}

	// Protecting over and over fails more and more often.
	fails := 0
	for range 300 {
		f := withMoves(t, fighter(t, "eisen", 50), "full_barrier")
		bt := duel(f, waiter(t, "stark"))
		play(t, bt, r, useMove0)
		fails += count(play(t, bt, r, useMove0), EvFail)
	}
	if fails < 150 || fails > 250 {
		t.Errorf("second protection failed %d/300 times, want about 2/3", fails)
	}
}

func TestSubstitute(t *testing.T) {
	r := newRand()
	user := withMoves(t, fighter(t, "eisen", 50), "illusory_double")
	user.Moves = append(user.Moves, &MoveSlot{Move: nothing, PP: nothing.PP})
	attacker := withMove(fighter(t, "fern", 50), tackle)
	attacker.Moves[0].PP = 99
	b := duel(user, attacker)
	play(t, b, r, useMove0)
	cost := user.Stats.HP / 4
	if user.SubstituteHP() == 0 || user.HP > user.Stats.HP-cost {
		t.Fatalf("no substitute made: sub %d, HP %d", user.SubstituteHP(), user.HP)
	}
	hpWithSub := user.HP
	for user.SubstituteHP() > 0 {
		play(t, b, r, [2]Action{{Index: 1}, {Index: 0}}) // wait while the substitute takes hits
	}
	if user.HP != hpWithSub {
		t.Errorf("damage leaked through the substitute: %d -> %d", hpWithSub, user.HP)
	}

	// Status moves aimed at a substitute fail.
	sleeper := withMoves(t, fighter(t, "heiter", 50), "sleep_magic")
	decoy := withMoves(t, fighter(t, "eisen", 50), "illusory_double")
	b = duel(decoy, sleeper)
	b.Active(0).substituteHP = 10
	for range 5 {
		play(t, b, r, useMove0)
	}
	if decoy.Condition != data.NoCondition {
		t.Error("a status move got through a substitute")
	}
}

func TestChargeAndRecharge(t *testing.T) {
	r := newRand()
	lance := withMoves(t, fighter(t, "frieren", 60), "sunlight_lance")
	target := waiter(t, "eisen")
	b := duel(lance, target)
	ev := play(t, b, r, useMove0)
	if count(ev, EvCharging) != 1 || target.HP != target.Stats.HP || !b.Forced(0) {
		t.Fatalf("first turn should only charge: %+v", ev)
	}
	ev = play(t, b, r, useMove0)
	if count(ev, EvDamage) != 1 || b.Forced(0) {
		t.Errorf("second turn should fire: %+v", ev)
	}
	if lance.Moves[0].PP != 98 {
		t.Errorf("a charge move should spend PP once, PP is %d", lance.Moves[0].PP)
	}

	// In a Heatwave, Sunlight Lance fires at once.
	b = duel(withMoves(t, fighter(t, "frieren", 60), "sunlight_lance"), waiter(t, "eisen"))
	b.Weather, b.WeatherTurns = data.Heatwave, 5
	if ev := play(t, b, r, useMove0); count(ev, EvCharging) != 0 || count(ev, EvDamage) != 1 {
		t.Errorf("Sunlight Lance should not charge in a Heatwave: %+v", ev)
	}

	beam := withMoves(t, fighter(t, "fern", 70), "annihilation")
	beam.Moves[0].Move.Accuracy = 0 // never miss, for the test
	tank := waiter(t, "eisen")
	tank.Stats.HP, tank.HP = 10000, 10000
	b = duel(beam, tank)
	play(t, b, r, useMove0)
	if !b.Forced(0) {
		t.Fatal("Annihilation Beam should force a recharge")
	}
	if ev := play(t, b, r, useMove0); count(ev, EvMustRecharge) != 1 || b.Forced(0) {
		t.Errorf("recharge turn: %+v", ev)
	}
}

func TestTraps(t *testing.T) {
	r := newRand()
	chainer := withMoves(t, fighter(t, "heiter", 50), "binding_chains")
	victim := waiter(t, "stark")
	b, _ := NewBattle(side("A", chainer, waiter(t, "sein")), side("B", victim, waiter(t, "eisen")))
	play(t, b, r, useMove0)
	if !victim.Trapped() {
		t.Fatal("Binding Chains did not trap")
	}
	if _, err := b.PlayTurn(r, [2]Action{{Index: 0}, {Switch: true, Index: 1}}); !errors.Is(err, ErrTrapped) {
		t.Errorf("a trapped character switched: %v", err)
	}
	// The trap ends when the trapper leaves.
	play(t, b, r, [2]Action{{Switch: true, Index: 1}, {Index: 0}})
	if victim.Trapped() {
		t.Error("the trap should end when the trapper switches out")
	}

	snarer := withMoves(t, fighter(t, "frieren", 50), "vine_snare")
	snarer.Moves[0].Move.Accuracy = 0
	wrapped := waiter(t, "eisen")
	b = duel(snarer, wrapped)
	play(t, b, r, useMove0)
	turns := 0
	for wrapped.Trapped() {
		before := wrapped.HP
		ev := play(t, b, r, [2]Action{{Index: 0}, {Index: 0}})
		turns++
		if i := first(ev, EvTrapDamage); i >= 0 && ev[i].Amount != wrapped.Stats.HP/8 && ev[i].Amount != before {
			t.Errorf("trap damage %d, want 1/8 of max HP", ev[i].Amount)
		}
		snarer.Moves[0].PP = 99
		if turns > 10 {
			t.Fatal("partial trap never ended")
		}
	}
}

func TestTaunt(t *testing.T) {
	r := newRand()
	taunter := withMoves(t, fighter(t, "stark", 50), "provoke")
	victim := withMoves(t, fighter(t, "eisen", 50), "iron_body", "slash")
	b := duel(taunter, victim)

	// Stark is faster, so Eisen's Iron Body this turn is blocked by the fresh taunt.
	ev := play(t, b, r, useMove0)
	if count(ev, EvTaunted) != 1 || count(ev, EvTauntBlocked) != 1 {
		t.Fatalf("taunt turn: %+v", ev)
	}
	if victim.CanUse(0) || !victim.CanUse(1) {
		t.Error("a taunted character should only be able to pick damaging moves")
	}
	if _, err := b.PlayTurn(r, [2]Action{{Index: 0}, {Index: 0}}); !errors.Is(err, ErrBadMove) {
		t.Errorf("picking a status move while taunted: %v", err)
	}
	for range 2 {
		play(t, b, r, [2]Action{{Index: 0}, {Index: 1}})
	}
	if victim.Taunted() {
		t.Error("taunt should wear off after 3 turns")
	}
}

func TestEncore(t *testing.T) {
	r := newRand()
	encorer := withMoves(t, fighter(t, "eisen", 50), "compulsion") // slow: moves after the target
	encorer.Moves = append(encorer.Moves, &MoveSlot{Move: nothing, PP: nothing.PP})
	victim := withMoves(t, fighter(t, "fern", 50), "defensive_magic", "zoltraak")
	b := duel(encorer, victim)
	ev := play(t, b, r, [2]Action{{Index: 0}, {Index: 0}})
	if count(ev, EvEncored) != 1 || victim.EncoreSlot() != 0 {
		t.Fatalf("encore: %+v", ev)
	}
	if victim.CanUse(1) {
		t.Error("an encored character picked a different move")
	}
	for range 3 {
		play(t, b, r, [2]Action{{Index: 1}, {Index: 0}})
	}
	if victim.EncoreSlot() >= 0 {
		t.Error("encore should wear off")
	}
}

func TestHazards(t *testing.T) {
	r := newRand()
	setter := withMoves(t, fighter(t, "stark", 50), "caltrops", "mana_mines", "curse_seals")
	lead, back := waiter(t, "sein"), waiter(t, "eisen") // Eisen is Martial: weak to Destruction
	b, _ := NewBattle(side("A", setter), side("B", lead, back))
	for _, slot := range []int{0, 0, 1} { // 2 layers of caltrops, then mines
		play(t, b, r, [2]Action{{Index: slot}, {Index: 0}})
	}
	if b.Sides[1].Layers(data.Caltrops) != 2 || b.Sides[1].Layers(data.ManaMines) != 1 {
		t.Fatalf("layers: caltrops %d, mines %d", b.Sides[1].Layers(data.Caltrops), b.Sides[1].Layers(data.ManaMines))
	}

	ev := play(t, b, r, [2]Action{{Index: 0}, {Switch: true, Index: 1}})
	mines := back.Stats.HP * 2 / 8 // Martial takes double from Destruction
	caltrops := back.Stats.HP / 6
	var got []int
	for _, e := range ev {
		if e.Kind == EvHazardDamage {
			got = append(got, e.Amount)
		}
	}
	if len(got) != 2 || got[0] != mines || got[1] != caltrops {
		t.Errorf("hazard damage %v, want [%d %d]", got, mines, caltrops)
	}

	// Curse Seals curse switch-ins, and Curse types absorb them.
	b, _ = NewBattle(side("A", setter), side("B", waiter(t, "sein"), waiter(t, "fern")))
	b.Sides[1].layers[data.CurseSeals] = 1
	play(t, b, r, [2]Action{{Index: 0}, {Switch: true, Index: 1}})
	if b.Active(1).Condition != data.Cursed {
		t.Errorf("Curse Seals did not curse the switch-in: %s", b.Active(1).Condition)
	}
	demon := waiter(t, "eisen")
	demon.Types = []data.Type{data.TypeCurse}
	b, _ = NewBattle(side("A", setter), side("B", waiter(t, "sein"), demon))
	b.Sides[1].layers[data.CurseSeals] = 2
	play(t, b, r, [2]Action{{Index: 0}, {Switch: true, Index: 1}})
	if demon.Condition != data.NoCondition || b.Sides[1].Layers(data.CurseSeals) != 0 {
		t.Error("a Curse type should absorb Curse Seals")
	}

	// Clearing Wind removes hazards on both sides.
	b = duel(withMoves(t, fighter(t, "frieren", 60), "clearing_wind"), waiter(t, "stark"))
	b.Sides[0].layers[data.Caltrops] = 3
	b.Sides[1].layers[data.ManaMines] = 1
	if ev := play(t, b, r, useMove0); count(ev, EvHazardsCleared) != 1 || b.Sides[0].Layers(data.Caltrops) != 0 {
		t.Errorf("Clearing Wind: %+v", ev)
	}
}

func TestPivot(t *testing.T) {
	r := newRand()
	pivoter := withMoves(t, fighter(t, "stark", 50), "hit_and_run")
	partner := waiter(t, "eisen")
	slowAttacker := withMove(fighter(t, "eisen", 50), tackle) // moves after Stark
	b, _ := NewBattle(side("A", pivoter, partner), side("B", slowAttacker))

	ev := play(t, b, r, useMove0)
	if b.PivotPending() != 0 || count(ev, EvMove) != 1 {
		t.Fatalf("the turn should pause after Hit and Run: pivot %d, %+v", b.PivotPending(), ev)
	}
	if _, err := b.PlayTurn(r, useMove0); !errors.Is(err, ErrPivotPending) {
		t.Errorf("playing during a pivot: %v", err)
	}
	ev, err := b.ChoosePivot(r, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if b.Active(0) != partner || partner.HP == partner.Stats.HP || pivoter.HP != pivoter.Stats.HP {
		t.Errorf("the rest of the turn should hit the switched-in partner: %+v", ev)
	}
	if b.PivotPending() != -1 || b.turn != nil {
		t.Error("the turn should be finished")
	}
}

func TestSureHitInWeather(t *testing.T) {
	r := newRand()
	caster := withMoves(t, fighter(t, "fern", 50), "lightning_magic")
	target := waiter(t, "eisen")
	target.stages[data.StatEvasion] = 6
	b := duel(caster, target)
	b.Weather, b.WeatherTurns = data.Storm, 99
	for range 20 {
		if ev := play(t, b, r, useMove0); count(ev, EvMiss) > 0 {
			t.Fatal("Lightning Magic missed in a Storm")
		}
		target.HP = target.Stats.HP
	}
}
