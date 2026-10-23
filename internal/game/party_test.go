package game

import (
	"errors"
	"slices"
	"testing"
)

func TestPartyAddRemove(t *testing.T) {
	p, err := PartyAdd([]int{1, 2}, 5)
	if err != nil || !slices.Equal(p, []int{1, 2, 5}) {
		t.Fatalf("PartyAdd = %v, %v", p, err)
	}
	if _, err := PartyAdd(p, 2); !errors.Is(err, ErrAlreadyInParty) {
		t.Errorf("adding a member twice: %v", err)
	}
	if _, err := PartyAdd([]int{1, 2, 3, 4}, 5); !errors.Is(err, ErrPartyFull) {
		t.Errorf("adding to a full party: %v", err)
	}

	p, err = PartyRemove([]int{1, 2, 3}, 1)
	if err != nil || !slices.Equal(p, []int{2, 3}) {
		t.Fatalf("PartyRemove = %v, %v", p, err)
	}
	if _, err := PartyRemove([]int{1}, 1); !errors.Is(err, ErrLastMember) {
		t.Errorf("removing the last member: %v", err)
	}
	if _, err := PartyRemove([]int{1, 2}, 9); !errors.Is(err, ErrNotInParty) {
		t.Errorf("removing a non-member: %v", err)
	}
}

func TestPartyWithLeader(t *testing.T) {
	tests := []struct {
		party       []int
		idx         int
		want        []int
		wantDropped int
	}{
		{[]int{1, 2, 3}, 3, []int{3, 2, 1}, 0},       // already in: swaps with the leader
		{[]int{1, 2}, 7, []int{7, 1, 2}, 0},          // room left: joins at the front
		{[]int{1, 2, 3, 4}, 7, []int{7, 2, 3, 4}, 1}, // full: replaces the leader
		{nil, 7, []int{7}, 0},
	}
	for _, tt := range tests {
		original := slices.Clone(tt.party)
		got, dropped := PartyWithLeader(tt.party, tt.idx)
		if !slices.Equal(got, tt.want) || dropped != tt.wantDropped {
			t.Errorf("PartyWithLeader(%v, %d) = %v, %d; want %v, %d", tt.party, tt.idx, got, dropped, tt.want, tt.wantDropped)
		}
		if !slices.Equal(tt.party, original) {
			t.Errorf("PartyWithLeader changed its input to %v", tt.party)
		}
	}
}
