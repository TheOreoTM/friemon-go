// Package game is pure game logic: no Discord, no database.
// Everything here can be unit-tested on its own.
package game

import (
	"math/rand/v2"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// stripAccents splits letters like "Ü" into "U" + a combining mark (NFD),
// drops the marks (unicode.Mn), and recombines what is left (NFC).
var stripAccents = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// Normalize turns a name into the form used for comparing guesses:
// lowercase, no accents, letters and digits only. "Übel" and "u bel" both become "ubel".
func Normalize(s string) string {
	s, _, _ = transform.String(stripAccents, s)
	var b strings.Builder
	// Ranging over a string yields runes (Unicode code points), not bytes.
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// Matches reports whether guess names the character, by its name or any alias.
func Matches(c data.Character, guess string) bool {
	g := Normalize(guess)
	if g == "" {
		return false
	}
	if g == Normalize(c.Name) {
		return true
	}
	for _, a := range c.Aliases {
		if g == Normalize(a) {
			return true
		}
	}
	return false
}

// Hint returns the name with some letters hidden as "_". Spaces stay visible,
// and at least one letter is always hidden so the hint never gives the full answer.
// The caller must escape "_" before sending it to Discord, where it means italics.
func Hint(r *rand.Rand, name string) string {
	letters := []rune(name) // a rune slice indexes by character, so "Ü" is one element
	var hideable []int
	for i, l := range letters {
		if !unicode.IsSpace(l) {
			hideable = append(hideable, i)
		}
	}
	if len(hideable) == 0 {
		return name
	}

	// Hide about two thirds of the letters, but at least one.
	toHide := max(1, len(hideable)*2/3)
	r.Shuffle(len(hideable), func(i, j int) { hideable[i], hideable[j] = hideable[j], hideable[i] })
	for _, i := range hideable[:toHide] {
		letters[i] = '_'
	}
	return string(letters)
}
