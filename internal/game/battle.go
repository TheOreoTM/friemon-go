package game

import (
	"errors"
	"math/rand/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// MoveSlot is a move a fighter knows, with the PP it has left this battle.
type MoveSlot struct {
	Move data.Move
	PP   int
}

// Fighter is one character in a battle.
type Fighter struct {
	Name        string
	CharacterID string
	Idx         int // the owner's number for it, so results can be saved
	Level       int
	Types       []data.Type
	Stats       data.Stats // final stats, from CalcStats
	EVs         data.Stats
	EVYield     data.Stats
	HP          int
	Moves       []*MoveSlot
	Condition   data.Condition

	sleepTurns   int // turns left asleep
	curseCounter int // Deeply Cursed damage grows with this
	faintCounted bool

	// Volatile state: all of this resets when the fighter leaves the field.
	stages         [8]int
	confusionTurns int
	flinched       bool
	protected      bool // protected for the rest of this turn
	protectChain   int  // protections in a row; each one makes the next less likely
	substituteHP   int
	chargeSlot     int // the move being charged, or -1
	mustRecharge   bool
	trapped        bool // cannot switch while the trapper is on the field
	trapTurns      int  // partially trapped: damage each turn, and cannot switch
	tauntTurns     int
	encoreTurns    int
	encoreSlot     int
	lastSlot       int // the last move used, or -1
}

// FighterSpec is everything needed to build a fighter.
type FighterSpec struct {
	Character data.Character
	Idx       int
	Level     int
	Nature    data.Nature
	IVs, EVs  data.Stats
	MoveIDs   []string
}

// NewFighter builds a fighter at full HP and full PP. Unknown move IDs are skipped.
func NewFighter(s FighterSpec) *Fighter {
	c := s.Character
	f := &Fighter{
		Name:        c.Name,
		CharacterID: c.ID,
		Idx:         s.Idx,
		Level:       s.Level,
		Types:       c.Types,
		Stats:       CalcStats(c.BaseStats, s.IVs, s.EVs, s.Level, s.Nature),
		EVs:         s.EVs,
		EVYield:     c.EVYield,
	}
	f.HP = f.Stats.HP
	f.clearVolatiles()
	for _, id := range s.MoveIDs {
		if m, ok := data.MoveByID(id); ok {
			f.Moves = append(f.Moves, &MoveSlot{Move: m, PP: m.PP})
		}
	}
	return f
}

// Fainted reports whether the fighter is out of HP.
func (f *Fighter) Fainted() bool { return f.HP <= 0 }

// HasPP reports whether any move has PP left.
func (f *Fighter) HasPP() bool {
	for _, m := range f.Moves {
		if m.PP > 0 {
			return true
		}
	}
	return false
}

// CanUse reports whether the move in a slot can be chosen right now: it needs PP,
// a taunted fighter cannot pick status moves, and an encored one must repeat its move.
func (f *Fighter) CanUse(slot int) bool {
	if slot < 0 || slot >= len(f.Moves) || f.Moves[slot].PP == 0 {
		return false
	}
	if f.Taunted() && f.Moves[slot].Move.Category == data.Status {
		return false
	}
	if e := f.EncoreSlot(); e >= 0 && slot != e {
		return false
	}
	return true
}

// HasUsableMove reports whether any move can be chosen. If not, the fighter struggles.
func (f *Fighter) HasUsableMove() bool {
	for i := range f.Moves {
		if f.CanUse(i) {
			return true
		}
	}
	return false
}

// Stage returns a stat stage, from -6 to +6.
func (f *Fighter) Stage(s data.Stat) int { return f.stages[s] }

// Taunted reports whether the fighter can only use damaging moves.
func (f *Fighter) Taunted() bool { return f.tauntTurns > 0 }

// EncoreSlot returns the move the fighter is forced to repeat, or -1.
func (f *Fighter) EncoreSlot() int {
	if f.encoreTurns > 0 {
		return f.encoreSlot
	}
	return -1
}

// Trapped reports whether the fighter cannot switch out.
func (f *Fighter) Trapped() bool { return f.trapped || f.trapTurns > 0 }

// SubstituteHP is how much HP the fighter's substitute has left, or 0 if it has none.
func (f *Fighter) SubstituteHP() int { return f.substituteHP }

// clearVolatiles resets everything that ends when a fighter leaves the field.
func (f *Fighter) clearVolatiles() {
	f.stages = [8]int{}
	f.confusionTurns = 0
	f.flinched = false
	f.protected = false
	f.protectChain = 0
	f.substituteHP = 0
	f.chargeSlot = -1
	f.mustRecharge = false
	f.trapped = false
	f.trapTurns = 0
	f.tauntTurns = 0
	f.encoreTurns = 0
	f.encoreSlot = -1
	f.lastSlot = -1
	if f.Condition == data.DeeplyCursed {
		f.curseCounter = 1 // the main games reset this on switching out
	}
}

// Side is one player's party in a battle, plus the conditions on their side of the field.
type Side struct {
	Player string
	Party  []*Fighter
	Active int // index in Party of the fighter on the field

	turns  [8]int // turns left for wards and Tailwind, indexed by data.SideCondition
	layers [8]int // layers of each hazard
}

// Alive counts fighters that have not fainted.
func (s *Side) Alive() int {
	n := 0
	for _, f := range s.Party {
		if !f.Fainted() {
			n++
		}
	}
	return n
}

// TurnsLeft returns how long a ward or Tailwind lasts on this side, or 0.
func (s *Side) TurnsLeft(c data.SideCondition) int { return s.turns[c] }

// Layers returns how many layers of a hazard are on this side.
func (s *Side) Layers(c data.SideCondition) int { return s.layers[c] }

// KO records that one fighter knocked out another, for EV rewards.
type KO struct {
	Side    int // the side that scored the knockout
	Scorer  *Fighter
	Fainted *Fighter
}

// Battle is a party-vs-party battle with switching. It holds no Discord state,
// so it can be tested on its own.
type Battle struct {
	Sides  [2]*Side
	Turn   int // turns played so far
	Winner int // index of the winning side, or -1 while the battle is on
	KOs    []KO

	Weather       data.Weather
	WeatherTurns  int
	Terrain       data.Terrain
	TerrainTurns  int
	ReversalTurns int // turns left of Reversal Field

	moved [2]bool    // who has acted this turn, for flinching and protection
	turn  *turnState // the turn in progress, while it is paused for a pivot switch
	pivot int        // the side that must choose who switches in after a pivot move, or -1
}

// turnState lets a turn stop halfway (for a pivot switch) and carry on later.
type turnState struct {
	actions [2]Action
	order   []int
	next    int // index in order of the next side to act
}

// Action is what a player does on their turn: use a move, or switch.
type Action struct {
	Switch bool
	Index  int // a move slot, or a party index when switching; -1 means struggle
}

var (
	ErrBattleOver      = errors.New("the battle is over")
	ErrNeedReplacement = errors.New("a fainted character must be replaced first")
	ErrNoReplacement   = errors.New("this side does not need a replacement")
	ErrBadSwitch       = errors.New("cannot switch to that character")
	ErrBadMove         = errors.New("that move cannot be used")
	ErrTrapped         = errors.New("a trapped character cannot switch out")
	ErrPivotPending    = errors.New("waiting for a pivot switch")
	ErrNoPivot         = errors.New("this side is not switching after a pivot move")
)

// NewBattle starts a battle with each side's first party member on the field.
func NewBattle(a, b *Side) (*Battle, []Event) {
	bt := &Battle{Sides: [2]*Side{a, b}, Winner: -1, pivot: -1}
	var ev []Event
	for s := range bt.Sides {
		ev = append(ev, Event{Kind: EvSwitchIn, Side: s, Name: bt.Active(s).Name})
	}
	return bt, ev
}

// Active returns the fighter side s has on the field.
func (b *Battle) Active(s int) *Fighter { return b.Sides[s].Party[b.Sides[s].Active] }

// Over reports whether someone has won.
func (b *Battle) Over() bool { return b.Winner >= 0 }

// Forfeit ends the battle with the other side winning.
func (b *Battle) Forfeit(loser int) {
	b.Winner = 1 - loser
	b.turn, b.pivot = nil, -1
}

// Forced reports whether side s has no choice this turn, because its fighter is
// charging a move or recharging. The engine picks its action itself.
func (b *Battle) Forced(s int) bool {
	f := b.Active(s)
	return !f.Fainted() && (f.chargeSlot >= 0 || f.mustRecharge)
}

// PivotPending returns the side that must choose who switches in after a pivot move, or -1.
func (b *Battle) PivotPending() int { return b.pivot }

// NeedsReplacement reports whether side s must send out a new fighter before the next turn.
func (b *Battle) NeedsReplacement(s int) bool {
	return !b.Over() && b.turn == nil && b.Active(s).Fainted() && b.Sides[s].Alive() > 0
}

// CanSwitchTo reports whether side s may send out party member idx.
func (b *Battle) CanSwitchTo(s, idx int) bool {
	side := b.Sides[s]
	return idx >= 0 && idx < len(side.Party) && idx != side.Active && !side.Party[idx].Fainted()
}

// Replace sends out a new fighter after the active one fainted.
func (b *Battle) Replace(s, idx int) ([]Event, error) {
	if !b.NeedsReplacement(s) {
		return nil, ErrNoReplacement
	}
	if !b.CanSwitchTo(s, idx) {
		return nil, ErrBadSwitch
	}
	ev := b.switchIn(s, idx)
	return append(ev, b.checkFaints(s)...), nil
}

// ChoosePivot finishes a pivot move: side s sends out idx, then the rest of the turn plays.
func (b *Battle) ChoosePivot(r *rand.Rand, s, idx int) ([]Event, error) {
	if b.pivot != s {
		return nil, ErrNoPivot
	}
	if !b.CanSwitchTo(s, idx) {
		return nil, ErrBadSwitch
	}
	b.pivot = -1
	ev := b.switchIn(s, idx)
	ev = append(ev, b.checkFaints(s)...)
	return append(ev, b.continueTurn(r)...), nil
}

func (b *Battle) validate(s int, a Action) error {
	f := b.Active(s)
	if b.Forced(s) {
		return nil // the engine picks the action
	}
	if a.Switch {
		if f.Trapped() {
			return ErrTrapped
		}
		if !b.CanSwitchTo(s, a.Index) {
			return ErrBadSwitch
		}
		return nil
	}
	if !f.HasUsableMove() {
		return nil // any move choice becomes Struggle
	}
	if !f.CanUse(a.Index) {
		return ErrBadMove
	}
	return nil
}

// PlayTurn runs one turn with both sides' actions and returns what happened.
// If a pivot move needs a switch-in choice, the turn pauses: PivotPending says who
// must choose, and ChoosePivot carries the turn on.
func (b *Battle) PlayTurn(r *rand.Rand, actions [2]Action) ([]Event, error) {
	switch {
	case b.Over():
		return nil, ErrBattleOver
	case b.turn != nil:
		return nil, ErrPivotPending
	}
	for s := range b.Sides {
		if b.NeedsReplacement(s) {
			return nil, ErrNeedReplacement
		}
		if err := b.validate(s, actions[s]); err != nil {
			return nil, err
		}
	}
	b.Turn++
	b.moved = [2]bool{}

	for s := range actions {
		f := b.Active(s)
		f.protected = false
		switch {
		case f.chargeSlot >= 0:
			actions[s] = Action{Index: f.chargeSlot}
		case f.mustRecharge:
			actions[s] = Action{Index: -1}
		case actions[s].Switch:
		case !f.HasUsableMove():
			actions[s] = Action{Index: -1} // struggle
		}
	}
	b.turn = &turnState{actions: actions, order: b.turnOrder(r, actions)}
	return b.continueTurn(r), nil
}

// continueTurn plays the rest of the turn in progress.
func (b *Battle) continueTurn(r *rand.Rand) []Event {
	t := b.turn
	var ev []Event
	for t.next < len(t.order) && !b.Over() {
		s := t.order[t.next]
		t.next++
		a, f := t.actions[s], b.Active(s)
		if a.Switch {
			ev = append(ev, b.switchIn(s, a.Index)...)
			ev = append(ev, b.checkFaints(s)...)
			continue
		}
		if f.Fainted() {
			continue
		}
		if f.mustRecharge {
			f.mustRecharge = false
			b.moved[s] = true
			ev = append(ev, Event{Kind: EvMustRecharge, Side: s, Name: f.Name})
			continue
		}
		ev = append(ev, b.useMove(r, s, a.Index)...)
		b.moved[s] = true
		ev = append(ev, b.checkFaints(1-s)...)
		if b.Over() {
			b.pivot = -1
		}
		if b.pivot >= 0 {
			return ev // wait for ChoosePivot
		}
	}

	if !b.Over() {
		ev = append(ev, b.endOfTurn(t.order)...)
	}
	for s := range b.Sides {
		f := b.Active(s)
		f.flinched = false
		f.protected = false
	}
	b.turn = nil
	return ev
}

// turnOrder decides who acts first: switches before moves, then higher move priority,
// then Speed (slower first under Reversal Field), then a coin flip.
func (b *Battle) turnOrder(r *rand.Rand, actions [2]Action) []int {
	priority := func(s int) int {
		a, f := actions[s], b.Active(s)
		switch {
		case a.Switch:
			return 100
		case a.Index < 0:
			return 0
		}
		return f.Moves[a.Index].Move.Priority
	}
	p0, p1 := priority(0), priority(1)
	s0, s1 := b.speed(0), b.speed(1)
	if b.ReversalTurns > 0 {
		s0, s1 = s1, s0
	}
	switch {
	case p1 > p0, p1 == p0 && s1 > s0, p1 == p0 && s1 == s0 && r.IntN(2) == 1:
		return []int{1, 0}
	}
	return []int{0, 1}
}

// speed is the active fighter's Speed after stages, Bound and Tailwind.
func (b *Battle) speed(s int) int {
	f := b.Active(s)
	v := applyStage(f.Stats.Speed, f.stages[data.StatSpeed])
	if f.Condition == data.Bound {
		v /= 2
	}
	if b.Sides[s].turns[data.Tailwind] > 0 {
		v *= 2
	}
	return v
}

// switchIn puts party member idx on the field for side s, then applies hazards.
func (b *Battle) switchIn(s, idx int) []Event {
	side := b.Sides[s]
	var ev []Event
	if old := b.Active(s); !old.Fainted() {
		ev = append(ev, Event{Kind: EvSwitchOut, Side: s, Name: old.Name})
	}
	b.Active(s).clearVolatiles()
	b.freeFoe(s) // a trap ends when the trapper leaves
	side.Active = idx
	ev = append(ev, Event{Kind: EvSwitchIn, Side: s, Name: b.Active(s).Name})
	return append(ev, b.entryHazards(s)...)
}

// freeFoe ends any trap side s's fighter had on the other side.
func (b *Battle) freeFoe(s int) {
	foe := b.Active(1 - s)
	foe.trapped = false
	foe.trapTurns = 0
}

// checkFaints reports newly fainted fighters, checking side first first, and ends the
// battle when a side has nobody left. Checking the target before the user means a move
// that knocks out the target and its own user (through recoil) wins the battle.
func (b *Battle) checkFaints(first int) []Event {
	var ev []Event
	for _, s := range []int{first, 1 - first} {
		f := b.Active(s)
		if !f.Fainted() || f.faintCounted {
			continue
		}
		f.faintCounted = true
		b.freeFoe(s)
		ev = append(ev, Event{Kind: EvFaint, Side: s, Name: f.Name})
		if scorer := b.Active(1 - s); !scorer.Fainted() {
			b.KOs = append(b.KOs, KO{Side: 1 - s, Scorer: scorer, Fainted: f})
		}
		if b.Sides[s].Alive() == 0 && !b.Over() {
			b.Winner = 1 - s
			ev = append(ev, Event{Kind: EvWin, Side: 1 - s})
		}
	}
	return ev
}
