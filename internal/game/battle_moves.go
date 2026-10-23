package game

import (
	"math/rand/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

const maxStage = 6

// confusionHit is the typeless attack a confused fighter uses on itself.
var confusionHit = data.Move{Name: "confusion", Type: data.NoType, Category: data.Physical, Power: 40}

// useMove runs side s's move: checks whether it can act, spends PP, rolls accuracy,
// then deals damage and applies effects.
func (b *Battle) useMove(r *rand.Rand, s, slot int) []Event {
	user, target := b.Active(s), b.Active(1-s)
	ev, ok := b.canMove(r, s, user)
	if !ok {
		return ev
	}

	var move data.Move
	if user.HasPP() {
		ms := user.Moves[slot]
		ms.PP--
		move = ms.Move
	} else {
		ev = append(ev, Event{Kind: EvStruggle, Side: s, Name: user.Name})
		move = data.Struggle
	}
	ev = append(ev, Event{Kind: EvMove, Side: s, Name: user.Name, Move: move.Name})

	if move.Has(data.FlagRest) {
		return append(ev, rest(s, user)...)
	}

	hitsFoe := move.Category != data.Status
	for _, e := range move.Effects {
		if !e.Self {
			hitsFoe = true
		}
	}
	if hitsFoe {
		if target.Fainted() {
			return append(ev, Event{Kind: EvNoTarget})
		}
		if move.Accuracy > 0 && !accuracyHits(r, user, target, move.Accuracy) {
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
		if r.IntN(100) >= 20 {
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
				dmg, _ := damage(r, f, f, confusionHit, 1, false)
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

	total, landed := 0, 0
	for range hits {
		if target.Fainted() {
			break
		}
		dmg, crit := damage(r, user, target, m, eff, true)
		dmg = min(dmg, target.HP)
		target.HP -= dmg
		total += dmg
		landed++
		if crit {
			ev = append(ev, Event{Kind: EvCrit})
		}
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

	// Effects on the target only happen if it is still standing, and fail quietly.
	// Effects on the user (like a drop in Magic Attack) always happen.
	for _, e := range m.Effects {
		if e.Self && !user.Fainted() {
			ev = append(ev, b.applyEffect(r, s, e, true)...)
		} else if !e.Self && !target.Fainted() {
			ev = append(ev, b.applyEffect(r, s, e, false)...)
		}
	}
	return ev
}

func (b *Battle) statusMove(r *rand.Rand, s int, m data.Move) []Event {
	user := b.Active(s)
	var ev []Event
	if m.Heal > 0 {
		if healed := heal(user, max(1, user.Stats.HP*m.Heal/100)); healed > 0 {
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

	if e.CureStatus && t.Condition != data.NoCondition {
		cured := ev(EvCured)
		cured.Condition = t.Condition
		out = append(out, cured)
		t.Condition = data.NoCondition
	}

	if e.Status != data.NoCondition {
		switch {
		case t.Condition != data.NoCondition:
			if loud {
				out = append(out, ev(EvConditionFailed))
			}
		case data.ImmuneToStatus(e.Status, t.Types):
			if loud {
				immune := ev(EvImmune)
				immune.Condition = e.Status
				out = append(out, immune)
			}
		default:
			inflict(r, t, e.Status)
			got := ev(EvCondition)
			got.Condition = e.Status
			out = append(out, got)
		}
	}

	switch e.Volatile {
	case data.Confusion:
		if t.confusionTurns > 0 {
			if loud {
				out = append(out, ev(EvAlreadyConfused))
			}
		} else {
			t.confusionTurns = 2 + r.IntN(4) // counts down before each move: confused for 1-4 turns
			out = append(out, ev(EvConfused))
		}
	case data.Flinch:
		if !b.moved[ts] { // flinching only matters if the target has not acted yet
			t.flinched = true
		}
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
func rest(s int, f *Fighter) []Event {
	if f.HP == f.Stats.HP || f.Condition == data.Sleep {
		return []Event{{Kind: EvFail}}
	}
	healed := f.Stats.HP - f.HP
	f.HP = f.Stats.HP
	f.Condition = data.Sleep
	f.sleepTurns = 2
	return []Event{{Kind: EvRested, Side: s, Name: f.Name, Amount: healed}}
}

// heal restores up to amount HP and returns how much was actually restored.
func heal(f *Fighter, amount int) int {
	healed := min(amount, f.Stats.HP-f.HP)
	f.HP += healed
	return healed
}

// damage works out one hit, using the main games' formula: level, power and attack
// over defense, then critical hit, random spread, same-type bonus, type matchup and burn.
func damage(r *rand.Rand, user, target *Fighter, m data.Move, eff float64, canCrit bool) (dmg int, crit bool) {
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
	def := max(1, applyStage(target.Stats.Get(defStat), defStage))

	base := float64((2*user.Level/5+2)*m.Power*atk/def)/50 + 2
	mult := 0.85 + r.Float64()*0.15
	if crit {
		mult *= 1.5
	}
	for _, t := range user.Types {
		if t == m.Type {
			mult *= 1.5 // same-type attack bonus
			break
		}
	}
	mult *= eff
	if m.Category == data.Physical && user.Condition == data.Burn {
		mult *= 0.5
	}
	return max(1, int(base*mult)), crit
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
