package game

import (
	"github.com/TheOreoTM/friemon-go/internal/data"
)

const (
	fieldTurns    = 5 // how long weather, terrain, wards and Reversal Field last
	tailwindTurns = 4
)

// setWeather starts a weather, unless it is already active.
func (b *Battle) setWeather(w data.Weather) []Event {
	if b.Weather == w {
		return []Event{{Kind: EvFail}}
	}
	b.Weather, b.WeatherTurns = w, fieldTurns
	return []Event{{Kind: EvWeatherStart, Weather: w}}
}

// setTerrain starts a terrain, unless it is already active.
func (b *Battle) setTerrain(t data.Terrain) []Event {
	if b.Terrain == t {
		return []Event{{Kind: EvFail}}
	}
	b.Terrain, b.TerrainTurns = t, fieldTurns
	return []Event{{Kind: EvTerrainStart, Terrain: t}}
}

// setSideCondition puts a ward or Tailwind on side s, or a hazard on the other side.
func (b *Battle) setSideCondition(s int, c data.SideCondition) []Event {
	if c.Hazard() {
		foe := b.Sides[1-s]
		if foe.layers[c] >= c.MaxLayers() {
			return []Event{{Kind: EvFail}}
		}
		foe.layers[c]++
		return []Event{{Kind: EvSideStart, Side: 1 - s, SideCond: c, Amount: foe.layers[c]}}
	}
	side := b.Sides[s]
	if side.turns[c] > 0 {
		return []Event{{Kind: EvFail}}
	}
	side.turns[c] = fieldTurns
	if c == data.Tailwind {
		side.turns[c] = tailwindTurns
	}
	return []Event{{Kind: EvSideStart, Side: s, SideCond: c}}
}

// clearHazards removes hazards from both sides and the wards on the other side.
func (b *Battle) clearHazards(s int) []Event {
	cleared := false
	for _, side := range b.Sides {
		for _, c := range []data.SideCondition{data.ManaMines, data.Caltrops, data.CurseSeals} {
			if side.layers[c] > 0 {
				side.layers[c] = 0
				cleared = true
			}
		}
	}
	foe := b.Sides[1-s]
	for _, c := range []data.SideCondition{data.PhysicalWard, data.MagicWard} {
		if foe.turns[c] > 0 {
			foe.turns[c] = 0
			cleared = true
		}
	}
	if !cleared {
		return []Event{{Kind: EvFail}}
	}
	return []Event{{Kind: EvHazardsCleared}}
}

// toggleReversal starts Reversal Field, or ends it early if it is already up.
func (b *Battle) toggleReversal() []Event {
	if b.ReversalTurns > 0 {
		b.ReversalTurns = 0
		return []Event{{Kind: EvReversalEnd}}
	}
	b.ReversalTurns = fieldTurns
	return []Event{{Kind: EvReversalStart}}
}

// fieldBlocks reports whether the weather or terrain stops a condition being inflicted,
// and fills in which one did it on the event.
func (b *Battle) fieldBlocks(c data.Condition) (Event, bool) {
	switch {
	case b.Terrain == data.Mist:
		return Event{Kind: EvFieldBlocked, Terrain: data.Mist}, true
	case b.Terrain == data.ManaField && c == data.Sleep:
		return Event{Kind: EvFieldBlocked, Terrain: data.ManaField}, true
	case b.Weather == data.Heatwave && c == data.Freeze:
		return Event{Kind: EvFieldBlocked, Weather: data.Heatwave}, true
	case b.Weather == data.Storm && c == data.Burn:
		return Event{Kind: EvFieldBlocked, Weather: data.Storm}, true
	}
	return Event{}, false
}

// entryHazards hurts or curses side s's fighter as it switches in.
func (b *Battle) entryHazards(s int) []Event {
	side, f := b.Sides[s], b.Active(s)
	var ev []Event
	hurt := func(c data.SideCondition, dmg int) {
		if f.Fainted() {
			return
		}
		dmg = min(max(1, dmg), f.HP)
		f.HP -= dmg
		ev = append(ev, Event{Kind: EvHazardDamage, Side: s, Name: f.Name, SideCond: c, Amount: dmg})
	}

	if side.layers[data.ManaMines] > 0 {
		eff := data.Effectiveness(data.TypeDestruction, f.Types)
		hurt(data.ManaMines, int(float64(f.Stats.HP)*eff/8))
	}
	if n := side.layers[data.Caltrops]; n > 0 {
		hurt(data.Caltrops, f.Stats.HP/[]int{0, 8, 6, 4}[n]) // 1/8, 1/6 or 1/4 of max HP
	}
	if n := side.layers[data.CurseSeals]; n > 0 && !f.Fainted() {
		cond := data.Cursed
		if n >= 2 {
			cond = data.DeeplyCursed
		}
		_, blocked := b.fieldBlocks(cond)
		switch {
		case data.ImmuneToStatus(cond, f.Types):
			side.layers[data.CurseSeals] = 0 // Curse types soak the seals up
			ev = append(ev, Event{Kind: EvSealsAbsorbed, Side: s, Name: f.Name})
		case f.Condition == data.NoCondition && !blocked:
			inflict(nil, f, cond) // no random roll is needed for curses
			ev = append(ev, Event{Kind: EvCondition, Side: s, Name: f.Name, Condition: cond})
		}
	}
	return ev
}

// endOfTurn runs everything that happens after both sides have acted, in the main
// games' order: weather, terrain healing, conditions, traps, then timers running down.
func (b *Battle) endOfTurn(order []int) []Event {
	var ev []Event
	each := func(fn func(s int, f *Fighter)) {
		for _, s := range order {
			if f := b.Active(s); !f.Fainted() && !b.Over() {
				fn(s, f)
			}
		}
	}
	lose := func(kind EventKind, s int, f *Fighter, dmg int, e Event) {
		e.Kind, e.Side, e.Name, e.Amount = kind, s, f.Name, min(max(1, dmg), f.HP)
		f.HP -= e.Amount
		ev = append(ev, e)
		ev = append(ev, b.checkFaints(s)...)
	}

	if shelter, ok := data.WeatherShelter(b.Weather); ok {
		each(func(s int, f *Fighter) {
			for _, t := range f.Types {
				if t == shelter {
					return
				}
			}
			lose(EvWeatherDamage, s, f, f.Stats.HP/16, Event{Weather: b.Weather})
		})
	}

	if b.Terrain == data.Overgrowth {
		each(func(s int, f *Fighter) {
			if healed := heal(f, max(1, f.Stats.HP/16)); healed > 0 {
				ev = append(ev, Event{Kind: EvTerrainHeal, Side: s, Name: f.Name, Amount: healed})
			}
		})
	}

	each(func(s int, f *Fighter) {
		switch f.Condition {
		case data.Burn:
			lose(EvConditionDamage, s, f, f.Stats.HP/16, Event{Condition: data.Burn})
		case data.Cursed:
			lose(EvConditionDamage, s, f, f.Stats.HP/8, Event{Condition: data.Cursed})
		case data.DeeplyCursed:
			lose(EvConditionDamage, s, f, f.Stats.HP*f.curseCounter/16, Event{Condition: data.DeeplyCursed})
			f.curseCounter = min(f.curseCounter+1, 15)
		}
	})

	each(func(s int, f *Fighter) {
		if f.trapTurns == 0 {
			return
		}
		lose(EvTrapDamage, s, f, f.Stats.HP/8, Event{})
		f.trapTurns--
		if f.trapTurns == 0 && !f.Fainted() {
			ev = append(ev, Event{Kind: EvTrapEnd, Side: s, Name: f.Name})
		}
	})

	each(func(s int, f *Fighter) {
		if f.tauntTurns > 0 {
			if f.tauntTurns--; f.tauntTurns == 0 {
				ev = append(ev, Event{Kind: EvTauntEnd, Side: s, Name: f.Name})
			}
		}
		if f.encoreTurns > 0 {
			f.encoreTurns--
			if f.Moves[f.encoreSlot].PP == 0 {
				f.encoreTurns = 0 // the encore ends early once the move runs out
			}
			if f.encoreTurns == 0 {
				ev = append(ev, Event{Kind: EvEncoreEnd, Side: s, Name: f.Name})
			}
		}
	})
	if b.Over() {
		return ev
	}

	for s, side := range b.Sides {
		for _, c := range []data.SideCondition{data.PhysicalWard, data.MagicWard, data.Tailwind} {
			if side.turns[c] > 0 {
				if side.turns[c]--; side.turns[c] == 0 {
					ev = append(ev, Event{Kind: EvSideEnd, Side: s, SideCond: c})
				}
			}
		}
	}
	if b.ReversalTurns > 0 {
		if b.ReversalTurns--; b.ReversalTurns == 0 {
			ev = append(ev, Event{Kind: EvReversalEnd})
		}
	}
	if b.WeatherTurns > 0 {
		if b.WeatherTurns--; b.WeatherTurns == 0 {
			ev = append(ev, Event{Kind: EvWeatherEnd, Weather: b.Weather})
			b.Weather = data.NoWeather
		}
	}
	if b.TerrainTurns > 0 {
		if b.TerrainTurns--; b.TerrainTurns == 0 {
			ev = append(ev, Event{Kind: EvTerrainEnd, Terrain: b.Terrain})
			b.Terrain = data.NoTerrain
		}
	}
	return ev
}
