package game

import (
	"math/rand/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

const maxStage = 6

// confusionHit is the typeless attack a confused fighter uses on itself.
var confusionHit = data.Move{Name: "confusion", Type: data.NoType, Category: data.Physical, Power: 40}

// useMove runs side s's move: checks whether it can act, spends PP, handles charging,
// field moves and protection, rolls accuracy, then deals damage and applies effects.
func (b *Battle) useMove(r *rand.Rand, s, slot int) []Event {
	user, target := b.Active(s), b.Active(1-s)
	ev, ok := b.canMove(r, s, user)
	if !ok {
		user.chargeSlot = -1 // losing a turn also loses a charge
		user.protectChain = 0
		return ev
	}

	charged := user.chargeSlot >= 0 // second turn of a charge move: the PP is already spent
	var move data.Move
	switch {
	case charged:
		move = user.Moves[user.chargeSlot].Move
		user.chargeSlot = -1
	case slot < 0 || !user.HasPP():
		ev = append(ev, Event{Kind: EvStruggle, Side: s, Name: user.Name})
		move = data.Struggle
		user.lastSlot = -1
	default:
		ms := user.Moves[slot]
		move = ms.Move
		if user.Taunted() && move.Category == data.Status {
			return append(ev, Event{Kind: EvTauntBlocked, Side: s, Name: user.Name, Move: move.Name})
		}
		ms.PP--
		user.lastSlot = slot
	}
	if !move.Has(data.FlagProtect) {
		user.protectChain = 0
	}

	if move.Has(data.FlagCharge) && !charged && !(move.NoChargeIn != data.NoWeather && b.Weather == move.NoChargeIn) {
		user.chargeSlot = slot
		return append(ev, Event{Kind: EvCharging, Side: s, Name: user.Name, Move: move.Name})
	}
	ev = append(ev, Event{Kind: EvMove, Side: s, Name: user.Name, Move: move.Name})

	// Moves that act on the user or the field, not the target.
	switch {
	case move.Has(data.FlagRest):
		return append(ev, b.rest(s, user)...)
	case move.Has(data.FlagProtect):
		return append(ev, b.protect(r, s, user)...)
	case move.Has(data.FlagSubstitute):
		return append(ev, substitute(s, user)...)
	case move.Weather != data.NoWeather:
		return append(ev, b.setWeather(move.Weather)...)
	case move.Terrain != data.NoTerrain:
		return append(ev, b.setTerrain(move.Terrain)...)
	case move.Side != data.NoSideCondition:
		return append(ev, b.setSideCondition(s, move.Side)...)
	case move.Has(data.FlagClearHazards):
		return append(ev, b.clearHazards(s)...)
	case move.Has(data.FlagReversal):
		return append(ev, b.toggleReversal()...)
	}

	hitsFoe := move.Category != data.Status
	for _, e := range move.Effects {
		if !e.Self {
			hitsFoe = true
		}
	}
	if hitsFoe {
		switch {
		case target.Fainted():
			return append(ev, Event{Kind: EvNoTarget})
		case target.protected:
			return append(ev, Event{Kind: EvProtected, Side: 1 - s, Name: target.Name})
		case b.Terrain == data.SacredGround && move.Priority > 0:
			return append(ev, Event{Kind: EvFieldBlocked, Side: 1 - s, Name: target.Name, Terrain: data.SacredGround})
		case move.Category == data.Status && target.substituteHP > 0 && !move.Has(data.FlagSound):
			return append(ev, Event{Kind: EvFail}) // the substitute takes it instead
		}
		sureHit := move.SureHitIn != data.NoWeather && b.Weather == move.SureHitIn
		if move.Accuracy > 0 && !sureHit && !accuracyHits(r, user, target, move.Accuracy) {
			return append(ev, Event{Kind: EvMiss})
		}
	}

	if move.Category == data.Status {
		return append(ev, b.statusMove(r, s, move)...)
	}
	return append(ev, b.damagingMove(r, s, move)...)
}

// canMove runs the checks before a fighter acts, in the main games' order.
// ok is false if the fighter loses its turn.
func (b *Battle) canMove(r *rand.Rand, s int, f *Fighter) (ev []Event, ok bool) {
	e := func(kind EventKind) Event { return Event{Kind: kind, Side: s, Name: f.Name} }

	switch f.Condition {
	case data.Sleep:
		if f.sleepTurns > 0 {
			f.sleepTurns--
			return []Event{e(EvFastAsleep)}, false
		}
		f.Condition = data.NoCondition
		ev = append(ev, e(EvWokeUp))
	case data.Freeze:
		if b.Weather != data.Heatwave && r.IntN(100) >= 20 {
			return []Event{e(EvFrozenSolid)}, false
		}
		f.Condition = data.NoCondition
		ev = append(ev, e(EvThawed))
	}

	if f.flinched {
		return append(ev, e(EvFlinched)), false
	}

	if f.confusionTurns > 0 {
		f.confusionTurns--
		if f.confusionTurns == 0 {
			ev = append(ev, e(EvSnappedOut))
		} else {
			ev = append(ev, e(EvIsConfused))
			if r.IntN(3) == 0 {
				dmg, _ := b.damage(r, f, f, s, confusionHit, 1, false, false)
				dmg = min(dmg, f.HP)
				f.HP -= dmg
				hurt := e(EvConfusionHurt)
				hurt.Amount = dmg
				return append(ev, hurt), false
			}
		}
	}

	if f.Condition == data.Bound && r.IntN(100) < 25 {
		return append(ev, e(EvFullyBound)), false
	}
	return ev, true
}

func (b *Battle) damagingMove(r *rand.Rand, s int, m data.Move) []Event {
	user, target := b.Active(s), b.Active(1-s)
	var ev []Event

	hits := 1
	if m.MinHits > 0 {
		hits = rollHits(r, m.MinHits, m.MaxHits)
	}
	eff := data.Effectiveness(m.Type, target.Types)

	total, landed, hitSubstitute := 0, 0, false
	for range hits {
		if target.Fainted() {
			break
		}
		dmg, crit := b.damage(r, user, target, 1-s, m, eff, true, true)
		landed++
		if crit {
			ev = append(ev, Event{Kind: EvCrit})
		}
		if target.substituteHP > 0 && !m.Has(data.FlagSound) {
			hitSubstitute = true
			dmg = min(dmg, target.substituteHP)
			target.substituteHP -= dmg
			total += dmg
			ev = append(ev, Event{Kind: EvSubDamage, Side: 1 - s, Name: target.Name, Amount: dmg})
			if target.substituteHP == 0 {
				ev = append(ev, Event{Kind: EvSubBroke, Side: 1 - s, Name: target.Name})
			}
			continue
		}
		dmg = min(dmg, target.HP)
		target.HP -= dmg
		total += dmg
		ev = append(ev, Event{Kind: EvDamage, Side: 1 - s, Name: target.Name, Amount: dmg})
	}
	if eff != 1 {
		ev = append(ev, Event{Kind: EvEffectiveness, Mult: eff})
	}
	if m.MaxHits > 1 {
		ev = append(ev, Event{Kind: EvHits, Amount: landed})
	}

	if m.Drain > 0 && total > 0 && !user.Fainted() {
		if healed := heal(user, max(1, total*m.Drain/100)); healed > 0 {
			ev = append(ev, Event{Kind: EvDrain, Side: s, Name: user.Name, Amount: healed})
		}
	}
	recoil := 0
	switch {
	case m.ID == data.Struggle.ID:
		recoil = max(1, user.Stats.HP/4)
	case m.Recoil > 0 && total > 0:
		recoil = max(1, total*m.Recoil/100)
	}
	if recoil > 0 && !user.Fainted() {
		recoil = min(recoil, user.HP)
		user.HP -= recoil
		ev = append(ev, Event{Kind: EvRecoil, Side: s, Name: user.Name, Amount: recoil})
	}

	// Effects on the target only happen if it is still standing and no substitute took the
	// hit; they fail quietly. Effects on the user (like a drop in Magic Attack) always happen.
	for _, e := range m.Effects {
		if e.Self && !user.Fainted() {
			ev = append(ev, b.applyEffect(r, s, e, true)...)
		} else if !e.Self && !target.Fainted() && !hitSubstitute {
			ev = append(ev, b.applyEffect(r, s, e, false)...)
		}
	}

	if m.Has(data.FlagRecharge) && !user.Fainted() {
		user.mustRecharge = true
	}
	if m.Has(data.FlagPivot) && total > 0 && !user.Fainted() {
		for i := range b.Sides[s].Party {
			if b.CanSwitchTo(s, i) {
				b.pivot = s // PlayTurn pauses until ChoosePivot
				break
			}
		}
	}
	return ev
}

func (b *Battle) statusMove(r *rand.Rand, s int, m data.Move) []Event {
	user := b.Active(s)
	var ev []Event
	if m.Heal > 0 {
		amount := max(1, user.Stats.HP*m.Heal/100)
		if b.Weather == data.Heatwave {
			amount = amount * 4 / 3 // heals are a third stronger in a Heatwave
		}
		if healed := heal(user, amount); healed > 0 {
			ev = append(ev, Event{Kind: EvHeal, Side: s, Name: user.Name, Amount: healed})
		}
	}
	for _, e := range m.Effects {
		ev = append(ev, b.applyEffect(r, s, e, true)...)
	}
	// Status moves report every failure, so if nothing at all happened, say so once.
	if len(ev) == 0 {
		ev = append(ev, Event{Kind: EvFail})
	}
	return ev
}

// applyEffect applies one move effect from side s. Loud effects report why they failed;
// quiet ones (chance-based extras on attacks) fail silently.
func (b *Battle) applyEffect(r *rand.Rand, s int, e data.MoveEffect, loud bool) []Event {
	if e.Chance > 0 && r.IntN(100) >= e.Chance {
		return nil
	}
	ts := 1 - s
	if e.Self {
		ts = s
	}
	t := b.Active(ts)
	ev := func(kind EventKind) Event { return Event{Kind: kind, Side: ts, Name: t.Name} }
	var out []Event
	fail := func(e Event) {
		if loud {
			out = append(out, e)
		}
	}

	if e.CureStatus && t.Condition != data.NoCondition {
		cured := ev(EvCured)
		cured.Condition = t.Condition
		out = append(out, cured)
		t.Condition = data.NoCondition
	}

	if e.Status != data.NoCondition {
		blockedBy, blocked := b.fieldBlocks(e.Status)
		switch {
		case t.Condition != data.NoCondition:
			fail(ev(EvConditionFailed))
		case data.ImmuneToStatus(e.Status, t.Types):
			immune := ev(EvImmune)
			immune.Condition = e.Status
			fail(immune)
		case blocked:
			blockedBy.Side, blockedBy.Name = ts, t.Name
			fail(blockedBy)
		default:
			inflict(r, t, e.Status)
			got := ev(EvCondition)
			got.Condition = e.Status
			out = append(out, got)
		}
	}

	switch e.Volatile {
	case data.Confusion:
		switch {
		case t.confusionTurns > 0:
			fail(ev(EvAlreadyConfused))
		case b.Terrain == data.Mist:
			blocked := ev(EvFieldBlocked)
			blocked.Terrain = data.Mist
			fail(blocked)
		default:
			t.confusionTurns = 2 + r.IntN(4) // counts down before each move: confused for 1-4 turns
			out = append(out, ev(EvConfused))
		}
	case data.Flinch:
		if !b.moved[ts] { // flinching only matters if the target has not acted yet
			t.flinched = true
		}
	case data.Trap:
		if t.Trapped() {
			fail(Event{Kind: EvFail})
		} else {
			t.trapped = true
			out = append(out, ev(EvTrapped))
		}
	case data.PartialTrap:
		if t.trapTurns == 0 && !t.trapped {
			t.trapTurns = 4 + r.IntN(2)
			out = append(out, ev(EvPartiallyTrapped))
		}
	case data.Taunt:
		if t.Taunted() {
			fail(Event{Kind: EvFail})
		} else {
			t.tauntTurns = 3
			out = append(out, ev(EvTaunted))
		}
	case data.Encore:
		out = append(out, b.encore(ts, t, loud)...)
	}

	for _, sc := range e.Stats {
		old := t.stages[sc.Stat]
		t.stages[sc.Stat] = max(-maxStage, min(maxStage, old+sc.Stages))
		changed := t.stages[sc.Stat] - old
		if changed != 0 || loud {
			c := ev(EvStatChange)
			c.Stat, c.Stages, c.Amount = sc.Stat, sc.Stages, changed
			out = append(out, c)
		}
	}
	return out
}

// encore forces the target to repeat its last move for 3 turns. If it has not acted yet
// this turn, its move this turn changes too.
func (b *Battle) encore(ts int, t *Fighter, loud bool) []Event {
	if t.encoreTurns > 0 || t.lastSlot < 0 || t.Moves[t.lastSlot].PP == 0 {
		if loud {
			return []Event{{Kind: EvFail}}
		}
		return nil
	}
	t.encoreTurns, t.encoreSlot = 3, t.lastSlot
	if b.turn != nil && !b.moved[ts] && !b.turn.actions[ts].Switch {
		b.turn.actions[ts].Index = t.encoreSlot
	}
	return []Event{{Kind: EvEncored, Side: ts, Name: t.Name, Move: t.Moves[t.encoreSlot].Move.Name}}
}

// inflict gives f a condition. r is only needed for Sleep.
func inflict(r *rand.Rand, f *Fighter, c data.Condition) {
	f.Condition = c
	switch c {
	case data.Sleep:
		f.sleepTurns = 1 + r.IntN(3)
	case data.DeeplyCursed:
		f.curseCounter = 1
	}
}

// rest fully heals the user and puts it to sleep for exactly 2 turns.
func (b *Battle) rest(s int, f *Fighter) []Event {
	if blocked, ok := b.fieldBlocks(data.Sleep); ok {
		blocked.Side, blocked.Name = s, f.Name
		return []Event{blocked}
	}
	if f.HP == f.Stats.HP || f.Condition == data.Sleep {
		return []Event{{Kind: EvFail}}
	}
	healed := f.Stats.HP - f.HP
	f.HP = f.Stats.HP
	f.Condition = data.Sleep
	f.sleepTurns = 2
	return []Event{{Kind: EvRested, Side: s, Name: f.Name, Amount: healed}}
}

// protect shields the user for the rest of the turn. Each protection in a row is three
// times less likely to work, and it always fails if the user is the last to move.
func (b *Battle) protect(r *rand.Rand, s int, f *Fighter) []Event {
	odds := 1
	for range f.protectChain {
		odds *= 3
	}
	if b.moved[1-s] || r.IntN(odds) != 0 {
		f.protectChain = 0
		return []Event{{Kind: EvFail}}
	}
	f.protected = true
	f.protectChain++
	return []Event{{Kind: EvProtecting, Side: s, Name: f.Name}}
}

// substitute spends a quarter of max HP to make a decoy with that much HP.
func substitute(s int, f *Fighter) []Event {
	cost := f.Stats.HP / 4
	if f.substituteHP > 0 || f.HP <= cost {
		return []Event{{Kind: EvFail}}
	}
	f.HP -= cost
	f.substituteHP = cost
	return []Event{{Kind: EvSubstitute, Side: s, Name: f.Name}}
}

// heal restores up to amount HP and returns how much was actually restored.
func heal(f *Fighter, amount int) int {
	healed := min(amount, f.Stats.HP-f.HP)
	f.HP += healed
	return healed
}

// damage works out one hit, using the main games' formula: level, power and attack
// over defense, then critical hit, random spread, same-type bonus, type matchup and burn.
// With field set, weather, terrain and wards on side ts also apply.
func (b *Battle) damage(r *rand.Rand, user, target *Fighter, ts int, m data.Move, eff float64, canCrit, field bool) (dmg int, crit bool) {
	crit = canCrit && rollCrit(r, m.CritStage)

	atkStat, defStat := data.StatAttack, data.StatDefense
	if m.Category == data.Magical {
		atkStat, defStat = data.StatMagicAttack, data.StatMagicDefense
	}
	atkStage, defStage := user.stages[atkStat], target.stages[defStat]
	if crit { // critical hits ignore the attacker's drops and the defender's boosts
		atkStage, defStage = max(atkStage, 0), min(defStage, 0)
	}
	atk := applyStage(user.Stats.Get(atkStat), atkStage)
	def := applyStage(target.Stats.Get(defStat), defStage)
	if field {
		// Sandstorm raises Martial types' Magic Defense; Blizzard raises Nature types' Defense.
		if shelter, ok := data.WeatherShelter(b.Weather); ok && hasType(target, shelter) {
			if (b.Weather == data.Sandstorm && defStat == data.StatMagicDefense) ||
				(b.Weather == data.Blizzard && defStat == data.StatDefense) {
				def = def * 3 / 2
			}
		}
	}
	def = max(def, 1)

	base := float64((2*user.Level/5+2)*m.Power*atk/def)/50 + 2
	mult := 0.85 + r.Float64()*0.15
	if crit {
		mult *= 1.5
	}
	if hasType(user, m.Type) {
		mult *= 1.5 // same-type attack bonus
	}
	mult *= eff
	if m.Category == data.Physical && user.Condition == data.Burn {
		mult *= 0.5
	}
	if field {
		mult *= data.WeatherMultiplier(b.Weather, m.Type) * data.TerrainMultiplier(b.Terrain, m.Type)
		if b.Terrain == data.Overgrowth && m.Has(data.FlagGround) {
			mult *= 0.5
		}
		side := b.Sides[ts]
		if !crit && ((m.Category == data.Physical && side.turns[data.PhysicalWard] > 0) ||
			(m.Category == data.Magical && side.turns[data.MagicWard] > 0)) {
			mult *= 0.5 // wards halve damage, but critical hits go straight through
		}
	}
	return max(1, int(base*mult)), crit
}

func hasType(f *Fighter, t data.Type) bool {
	for _, ft := range f.Types {
		if ft == t {
			return true
		}
	}
	return false
}

// critOdds is 1 in N for each crit stage.
var critOdds = []int{24, 8, 2}

func rollCrit(r *rand.Rand, stage int) bool {
	if stage >= len(critOdds) {
		return true
	}
	return r.IntN(critOdds[stage]) == 0
}

// rollHits picks how many times a multi-hit move lands. For 2-5 hit moves this uses the
// main games' odds: 2 or 3 hits 35% each, 4 or 5 hits 15% each.
func rollHits(r *rand.Rand, lo, hi int) int {
	if lo == 2 && hi == 5 {
		switch n := r.IntN(100); {
		case n < 35:
			return 2
		case n < 70:
			return 3
		case n < 85:
			return 4
		default:
			return 5
		}
	}
	return lo + r.IntN(hi-lo+1)
}

// accuracyHits rolls whether a move lands, after the user's Accuracy and the target's
// Evasion stages.
func accuracyHits(r *rand.Rand, user, target *Fighter, accuracy int) bool {
	stage := max(-maxStage, min(maxStage, user.stages[data.StatAccuracy]-target.stages[data.StatEvasion]))
	chance := float64(accuracy)
	if stage >= 0 {
		chance = chance * float64(3+stage) / 3
	} else {
		chance = chance * 3 / float64(3-stage)
	}
	return r.Float64()*100 < chance
}

// applyStage scales a stat by its stage: +1 is x1.5, +2 is x2 ... +6 is x4,
// and -1 is x2/3 ... -6 is x1/4.
func applyStage(stat, stage int) int {
	if stage >= 0 {
		return stat * (2 + stage) / 2
	}
	return stat * 2 / (2 - stage)
}
