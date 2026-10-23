package game

import (
	"errors"
	"slices"
)

const (
	MaxParty     = 4  // characters in a party, which is also the battle team
	PartyXPShare = 50 // percent of the leader's chat XP the other members get
)

// A party is a list of character numbers (#1, #2, ...). Index 0 is the leader.
// These functions never change the slice they are given; they return a new one.

var (
	ErrPartyFull      = errors.New("your party is full")
	ErrAlreadyInParty = errors.New("that character is already in your party")
	ErrNotInParty     = errors.New("that character is not in your party")
	ErrLastMember     = errors.New("your party needs at least one character")
)

// PartyAdd puts idx at the end of the party.
func PartyAdd(party []int, idx int) ([]int, error) {
	switch {
	case slices.Contains(party, idx):
		return nil, ErrAlreadyInParty
	case len(party) >= MaxParty:
		return nil, ErrPartyFull
	}
	return append(slices.Clone(party), idx), nil
}

// PartyRemove takes idx out of the party. The next member becomes leader if needed.
func PartyRemove(party []int, idx int) ([]int, error) {
	i := slices.Index(party, idx)
	switch {
	case i < 0:
		return nil, ErrNotInParty
	case len(party) == 1:
		return nil, ErrLastMember
	}
	return slices.Delete(slices.Clone(party), i, i+1), nil
}

// PartyWithLeader makes idx the leader. If idx is already in the party, it swaps
// places with the old leader. Otherwise it joins at the front; if that would make the
// party too big, it replaces the old leader, which is returned as dropped (or 0).
func PartyWithLeader(party []int, idx int) (newParty []int, dropped int) {
	p := slices.Clone(party)
	if i := slices.Index(p, idx); i >= 0 {
		p[0], p[i] = p[i], p[0]
		return p, 0
	}
	if len(p) < MaxParty {
		return slices.Insert(p, 0, idx), 0
	}
	dropped, p[0] = p[0], idx
	return p, dropped
}
