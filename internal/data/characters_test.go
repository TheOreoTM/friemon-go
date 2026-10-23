package data

import "testing"

func TestCharactersAreValid(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Characters {
		if c.ID == "" || c.Name == "" || c.Description == "" {
			t.Errorf("character %+v is missing an ID, name or description", c)
		}
		if seen[c.ID] {
			t.Errorf("duplicate character ID %q", c.ID)
		}
		seen[c.ID] = true
		if c.Rarity.String() == "Unknown" {
			t.Errorf("%s has an unknown rarity %d", c.ID, c.Rarity)
		}
	}
}

func TestEveryRarityHasACharacter(t *testing.T) {
	for _, r := range []Rarity{Common, Rare, Legendary} {
		found := false
		for _, c := range Characters {
			if c.Rarity == r {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("no character has rarity %s", r)
		}
	}
}

func TestCharacterByID(t *testing.T) {
	if c, ok := CharacterByID("frieren"); !ok || c.Name != "Frieren" {
		t.Errorf("CharacterByID(frieren) = %+v, %v", c, ok)
	}
	if _, ok := CharacterByID("nope"); ok {
		t.Error("CharacterByID(nope) should not be found")
	}
}

func TestCharacterTypesStatsAndLearnsets(t *testing.T) {
	for _, c := range Characters {
		if len(c.Types) < 1 || len(c.Types) > 2 {
			t.Errorf("%s has %d types, want 1 or 2", c.ID, len(c.Types))
		}
		for _, ty := range c.Types {
			if ty.String() == "Unknown" {
				t.Errorf("%s has an unknown type %d", c.ID, ty)
			}
		}

		b := c.BaseStats
		for _, s := range []int{b.HP, b.Attack, b.Defense, b.MagicAttack, b.MagicDefense, b.Speed} {
			if s <= 0 {
				t.Errorf("%s has a base stat of %d", c.ID, s)
			}
		}

		startMoves := 0
		seen := map[string]bool{}
		for _, e := range c.Learnset {
			if _, ok := MoveByID(e.MoveID); !ok {
				t.Errorf("%s learns unknown move %q", c.ID, e.MoveID)
			}
			if e.Level < 1 || e.Level > 100 {
				t.Errorf("%s learns %s at level %d, want 1-100", c.ID, e.MoveID, e.Level)
			}
			if seen[e.MoveID] {
				t.Errorf("%s learns %s twice", c.ID, e.MoveID)
			}
			seen[e.MoveID] = true
			if e.Level == 1 {
				startMoves++
			}
		}
		if startMoves == 0 {
			t.Errorf("%s has no level 1 moves, so a fresh catch could know nothing", c.ID)
		}
	}
}

func TestEVYields(t *testing.T) {
	for _, c := range Characters {
		if total := c.EVYield.Total(); total < 1 || total > 3 {
			t.Errorf("%s yields %d EVs, want 1-3", c.ID, total)
		}
	}
}
