package game

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// newRand returns a seeded generator so test results are repeatable.
func newRand() *rand.Rand {
	return rand.New(rand.NewPCG(1, 2))
}

// A table-driven test: a slice of cases, then one loop that checks them all.
// This is the standard Go testing style.
func TestMatches(t *testing.T) {
	ubel, _ := data.CharacterByID("ubel")
	himmel, _ := data.CharacterByID("himmel")

	tests := []struct {
		char  data.Character
		guess string
		want  bool
	}{
		{himmel, "Himmel", true},
		{himmel, "himmel", true},
		{himmel, "  HIMMEL  ", true},
		{himmel, "himmel the hero", true}, // alias
		{himmel, "Himmel-the-Hero", true}, // punctuation ignored
		{himmel, "himmelx", false},
		{himmel, "himme", false},
		{himmel, "", false},
		{ubel, "Übel", true},
		{ubel, "ubel", true}, // accent-insensitive
		{ubel, "UBEL", true},
		{ubel, "ub", false},
	}
	for _, tt := range tests {
		if got := Matches(tt.char, tt.guess); got != tt.want {
			t.Errorf("Matches(%s, %q) = %v, want %v", tt.char.ID, tt.guess, got, tt.want)
		}
	}
}

// Two characters must never accept the same guess, or catch would be ambiguous.
func TestNamesDoNotClash(t *testing.T) {
	owner := map[string]string{} // normalized name -> character ID
	for _, c := range data.Characters {
		for _, name := range append([]string{c.Name}, c.Aliases...) {
			n := Normalize(name)
			if n == "" {
				t.Errorf("%s has a name %q that normalizes to nothing", c.ID, name)
			}
			if other, ok := owner[n]; ok && other != c.ID {
				t.Errorf("%q is used by both %s and %s", n, other, c.ID)
			}
			owner[n] = c.ID
		}
	}
}

func TestHint(t *testing.T) {
	r := newRand()
	for _, name := range []string{"Frieren", "Himmel the Hero", "Übel", "A", ""} {
		for range 50 { // ranging over an int repeats the loop that many times
			h := Hint(r, name)
			want, got := []rune(name), []rune(h)
			if len(got) != len(want) {
				t.Fatalf("Hint(%q) = %q, wrong length", name, h)
			}
			hidden := 0
			for i := range want {
				switch {
				case got[i] == '_':
					hidden++
				case got[i] != want[i]:
					t.Fatalf("Hint(%q) = %q, changed a letter", name, h)
				}
			}
			if strings.TrimSpace(name) != "" && hidden == 0 {
				t.Fatalf("Hint(%q) = %q, hid nothing", name, h)
			}
		}
	}
}
