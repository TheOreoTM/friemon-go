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
