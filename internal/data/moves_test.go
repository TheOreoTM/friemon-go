package data

import "testing"

func TestMovesAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, m := range append([]Move{Struggle}, Moves...) {
		if m.ID == "" || m.Name == "" || m.Description == "" {
			t.Errorf("move %+v is missing an ID, name or description", m)
		}
		if seen[m.ID] {
			t.Errorf("duplicate move ID %q", m.ID)
		}
		seen[m.ID] = true
		if m.Category.String() == "Unknown" || (m.Type.String() == "Unknown") {
			t.Errorf("%s has an unknown type or category", m.ID)
		}
		if m.Accuracy < 0 || m.Accuracy > 100 {
			t.Errorf("%s has accuracy %d, want 0 (never misses) to 100", m.ID, m.Accuracy)
		}
		if m.PP <= 0 {
			t.Errorf("%s has %d PP", m.ID, m.PP)
		}
		// Status moves deal no damage; every other move must.
		if (m.Category == Status) != (m.Power == 0) {
			t.Errorf("%s is %s with power %d", m.ID, m.Category, m.Power)
		}
		if m.Category == Status && (m.Drain != 0 || m.Recoil != 0 || m.MinHits != 0) {
			t.Errorf("%s is a status move with drain, recoil or multiple hits", m.ID)
		}
		if m.Category != Status && m.Heal != 0 {
			t.Errorf("%s deals damage and also has Heal; use Drain instead", m.ID)
		}
		if m.MinHits > m.MaxHits || (m.MinHits == 0) != (m.MaxHits == 0) {
			t.Errorf("%s has hits %d-%d", m.ID, m.MinHits, m.MaxHits)
		}
		for _, e := range m.Effects {
			if e.Chance < 0 || e.Chance > 100 {
				t.Errorf("%s has an effect with chance %d", m.ID, e.Chance)
			}
			for _, sc := range e.Stats {
				if sc.Stat == StatHP || sc.Stages == 0 || sc.Stages < -6 || sc.Stages > 6 {
					t.Errorf("%s has an invalid stat change %+v", m.ID, sc)
				}
			}
			if e.Self && (e.Status != NoCondition || e.Volatile != NoVolatile) {
				t.Errorf("%s inflicts a status on its own user", m.ID)
			}
		}
		// A status move with nothing to do would always fail.
		if m.Category == Status && m.Heal == 0 && len(m.Effects) == 0 && !m.Has(FlagRest) {
			t.Errorf("%s is a status move that does nothing", m.ID)
		}
	}
}

func TestEffectiveness(t *testing.T) {
	tests := []struct {
		attack   Type
		defender []Type
		want     float64
	}{
		{TypeDestruction, []Type{TypeCurse}, 2},
		{TypeDestruction, []Type{TypeBarrier}, 0.5},
		{TypeDestruction, []Type{TypeHoly}, 1},
		{TypeMartial, []Type{TypeBarrier, TypeNature}, 4}, // both weaknesses multiply
		{TypeMartial, []Type{TypeBarrier, TypeCurse}, 1},  // a weakness and a resistance cancel out
		{NoType, []Type{TypeCurse}, 1},                    // typeless is always neutral
	}
	for _, tt := range tests {
		if got := Effectiveness(tt.attack, tt.defender); got != tt.want {
			t.Errorf("Effectiveness(%s, %v) = %v, want %v", tt.attack, tt.defender, got, tt.want)
		}
	}
}

func TestNatures(t *testing.T) {
	if len(Natures) != 25 {
		t.Fatalf("got %d natures, want 25", len(Natures))
	}
	neutral := 0
	seen := map[string]bool{}
	for _, n := range Natures {
		if seen[n.ID] {
			t.Errorf("duplicate nature %q", n.ID)
		}
		seen[n.ID] = true
		if n.Neutral() {
			neutral++
		}
		if n.Up == StatHP || n.Down == StatHP {
			t.Errorf("%s changes HP, which natures never do", n.Name)
		}
	}
	if neutral != 5 {
		t.Errorf("got %d neutral natures, want 5", neutral)
	}
	if n, ok := NatureByID("adamant"); !ok || n.Up != StatAttack || n.Down != StatMagicAttack {
		t.Errorf("adamant = %+v", n)
	}
}

func TestStatusImmunity(t *testing.T) {
	if !ImmuneToStatus(Cursed, []Type{TypeCurse}) {
		t.Error("Curse types should be immune to Cursed")
	}
	if ImmuneToStatus(Burn, []Type{TypeCurse}) {
		t.Error("Curse types should not be immune to Burn")
	}
}
