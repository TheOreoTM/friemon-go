package data

import "testing"

func TestMovesAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, m := range Moves {
		if m.ID == "" || m.Name == "" || m.Description == "" {
			t.Errorf("move %+v is missing an ID, name or description", m)
		}
		if seen[m.ID] {
			t.Errorf("duplicate move ID %q", m.ID)
		}
		seen[m.ID] = true
		if m.Type.String() == "Unknown" || m.Category.String() == "Unknown" {
			t.Errorf("%s has an unknown type or category", m.ID)
		}
		if m.Accuracy < 1 || m.Accuracy > 100 {
			t.Errorf("%s has accuracy %d, want 1-100", m.ID, m.Accuracy)
		}
		// Status moves deal no damage; every other move must.
		if (m.Category == Status) != (m.Power == 0) {
			t.Errorf("%s is %s with power %d", m.ID, m.Category, m.Power)
		}
	}
}

func TestOnlyStatusMovesHaveEffects(t *testing.T) {
	for _, m := range Moves {
		if m.Category != Status && m.Effect != NoEffect {
			t.Errorf("%s deals damage but also has an effect", m.ID)
		}
	}
}

func TestEffectiveness(t *testing.T) {
	tests := []struct {
		attack   Type
		defender []Type
		want     float64
	}{
		{Destruction, []Type{Curse}, 2},
		{Destruction, []Type{Barrier}, 0.5},
		{Destruction, []Type{Holy}, 1},
		{Martial, []Type{Barrier, Nature}, 4}, // both weaknesses multiply
		{Martial, []Type{Barrier, Curse}, 1},  // a weakness and a resistance cancel out
	}
	for _, tt := range tests {
		if got := Effectiveness(tt.attack, tt.defender); got != tt.want {
			t.Errorf("Effectiveness(%s, %v) = %v, want %v", tt.attack, tt.defender, got, tt.want)
		}
	}
}
