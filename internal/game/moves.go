package game

import (
	"errors"
	"math/rand/v2"
	"slices"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// MaxMoves is how many moves a character can know at once.
const MaxMoves = 4

// LearnableMoves returns the IDs of every move c can know at this level, in learnset order.
func LearnableMoves(c data.Character, level int) []string {
	var ids []string
	for _, e := range c.Learnset {
		if e.Level <= level {
			ids = append(ids, e.MoveID)
		}
	}
	return ids
}

// StartingMoves picks up to MaxMoves random moves from those learnable at level.
func StartingMoves(r *rand.Rand, c data.Character, level int) []string {
	ids := LearnableMoves(c, level)
	r.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	return ids[:min(len(ids), MaxMoves)]
}

// Errors from CheckLearn. Exported error values like these let callers
// tell the cases apart with errors.Is, instead of comparing message strings.
var (
	ErrUnknownMove  = errors.New("that move does not exist")
	ErrNotInSet     = errors.New("this character cannot learn that move")
	ErrLevelTooLow  = errors.New("this character's level is too low for that move")
	ErrAlreadyKnown = errors.New("this character already knows that move")
)

// CheckLearn reports whether a character at level, knowing known, may learn moveID.
// It does not check the 4-move limit; the caller asks which move to replace.
func CheckLearn(c data.Character, level int, known []string, moveID string) error {
	if _, ok := data.MoveByID(moveID); !ok {
		return ErrUnknownMove
	}
	if slices.Contains(known, moveID) {
		return ErrAlreadyKnown
	}
	for _, e := range c.Learnset {
		if e.MoveID == moveID {
			if level < e.Level {
				return ErrLevelTooLow
			}
			return nil
		}
	}
	return ErrNotInSet
}

// NewMovesBetween returns moves c unlocks after oldLevel, up to and including newLevel.
func NewMovesBetween(c data.Character, oldLevel, newLevel int) []string {
	var ids []string
	for _, e := range c.Learnset {
		if e.Level > oldLevel && e.Level <= newLevel {
			ids = append(ids, e.MoveID)
		}
	}
	return ids
}
