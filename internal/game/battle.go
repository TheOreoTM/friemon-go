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

	sleepTurns     int // turns left asleep
	curseCounter   int // Deeply Cursed damage grows with this
	stages         [8]int
	confusionTurns int
	flinched       bool
	faintCounted   bool
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
	for _, id := range s.MoveIDs {
		if m, ok := data.MoveByID(id); ok {
			f.Moves = append(f.Moves, &MoveSlot{Move: m, PP: m.PP})
		}
	}
	return f
}

// Fainted reports whether the fighter is out of HP.
func (f *Fighter) Fainted() bool { return f.HP <= 0 }

// HasPP reports whether any move has PP left. If not, the fighter can only struggle.
func (f *Fighter) HasPP() bool {
	for _, m := range f.Moves {
		if m.PP > 0 {
			return true
		}
	}
	return false
}

// Stage returns a stat stage, from -6 to +6.
func (f *Fighter) Stage(s data.Stat) int { return f.stages[s] }

// clearVolatiles resets everything that ends when a fighter leaves the field.
func (f *Fighter) clearVolatiles() {
	f.stages = [8]int{}
	f.confusionTurns = 0
	f.flinched = false
	if f.Condition == data.DeeplyCursed {
		f.curseCounter = 1 // the main games reset this on switching out
	}
}

// Side is one player's party in a battle.
type Side struct {
	Player string
	Party  []*Fighter
	Active int // index in Party of the fighter on the field
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

	moved [2]bool // who has acted this turn, for flinching
}

// Action is what a player does on their turn: use a move, or switch.
type Action struct {
	Switch bool
	Index  int // a move slot, or a party index when switching
}

var (
	ErrBattleOver      = errors.New("the battle is over")
	ErrNeedReplacement = errors.New("a fainted character must be replaced first")
	ErrNoReplacement   = errors.New("this side does not need a replacement")
	ErrBadSwitch       = errors.New("cannot switch to that character")
	ErrBadMove         = errors.New("that move cannot be used")
)

// NewBattle starts a battle with each side's first party member on the field.
func NewBattle(a, b *Side) (*Battle, []Event) {
	bt := &Battle{Sides: [2]*Side{a, b}, Winner: -1}
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
func (b *Battle) Forfeit(loser int) { b.Winner = 1 - loser }

// NeedsReplacement reports whether side s must send out a new fighter before the next turn.
func (b *Battle) NeedsReplacement(s int) bool {
	return !b.Over() && b.Active(s).Fainted() && b.Sides[s].Alive() > 0
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
	return b.switchIn(s, idx), nil
}

func (b *Battle) validate(s int, a Action) error {
	if a.Switch {
		if !b.CanSwitchTo(s, a.Index) {
			return ErrBadSwitch
		}
		return nil
	}
	f := b.Active(s)
	if !f.HasPP() {
		return nil // any move choice becomes Struggle
	}
	if a.Index < 0 || a.Index >= len(f.Moves) || f.Moves[a.Index].PP == 0 {
		return ErrBadMove
	}
	return nil
}

// PlayTurn runs one turn with both sides' actions and returns what happened.
func (b *Battle) PlayTurn(r *rand.Rand, actions [2]Action) ([]Event, error) {
	if b.Over() {
		return nil, ErrBattleOver
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

	var ev []Event
	order := b.turnOrder(r, actions)
	for _, s := range order {
		if b.Over() {
			break
		}
		if actions[s].Switch {
			ev = append(ev, b.switchIn(s, actions[s].Index)...)
			continue
		}
		if b.Active(s).Fainted() {
			continue
		}
		ev = append(ev, b.useMove(r, s, actions[s].Index)...)
		b.moved[s] = true
		ev = append(ev, b.checkFaints(1-s)...)
	}

	if !b.Over() {
		ev = append(ev, b.endOfTurn(order)...)
	}
	for s := range b.Sides {
		b.Active(s).flinched = false
	}
	return ev, nil
}

// turnOrder decides who acts first: switches before moves, then higher move priority,
// then higher Speed, then a coin flip.
func (b *Battle) turnOrder(r *rand.Rand, actions [2]Action) []int {
	priority := func(s int) int {
		a, f := actions[s], b.Active(s)
		switch {
		case a.Switch:
			return 100
		case !f.HasPP():
			return 0
		}
		return f.Moves[a.Index].Move.Priority
	}
	p0, p1 := priority(0), priority(1)
	s0, s1 := b.Active(0).speed(), b.Active(1).speed()
	switch {
	case p1 > p0, p1 == p0 && s1 > s0, p1 == p0 && s1 == s0 && r.IntN(2) == 1:
		return []int{1, 0}
	}
	return []int{0, 1}
}

func (f *Fighter) speed() int {
	s := applyStage(f.Stats.Speed, f.stages[data.StatSpeed])
	if f.Condition == data.Bound {
		s /= 2
	}
	return s
}

// switchIn puts party member idx on the field for side s.
func (b *Battle) switchIn(s, idx int) []Event {
	side := b.Sides[s]
	var ev []Event
	if old := b.Active(s); !old.Fainted() {
		ev = append(ev, Event{Kind: EvSwitchOut, Side: s, Name: old.Name})
	}
	b.Active(s).clearVolatiles()
	side.Active = idx
	return append(ev, Event{Kind: EvSwitchIn, Side: s, Name: b.Active(s).Name})
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

// endOfTurn applies damage from conditions, in the order the fighters acted.
func (b *Battle) endOfTurn(order []int) []Event {
	var ev []Event
	for _, s := range order {
		f := b.Active(s)
		if f.Fainted() || b.Over() {
			continue
		}
		var dmg int
		switch f.Condition {
		case data.Burn:
			dmg = max(1, f.Stats.HP/16)
		case data.Cursed:
			dmg = max(1, f.Stats.HP/8)
		case data.DeeplyCursed:
			dmg = max(1, f.Stats.HP*f.curseCounter/16)
			f.curseCounter = min(f.curseCounter+1, 15)
		default:
			continue
		}
		dmg = min(dmg, f.HP)
		f.HP -= dmg
		ev = append(ev, Event{Kind: EvConditionDamage, Side: s, Name: f.Name, Condition: f.Condition, Amount: dmg})
		ev = append(ev, b.checkFaints(s)...)
	}
	return ev
}
