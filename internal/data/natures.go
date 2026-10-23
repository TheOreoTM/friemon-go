package data

import "strings"

// Nature raises one stat by 10% and lowers another by 10%. Neutral natures change nothing.
type Nature struct {
	ID   string // stable key saved in the database
	Name string
	Up   Stat // equal to Down for neutral natures
	Down Stat
}

// Neutral reports whether the nature changes no stats.
func (n Nature) Neutral() bool { return n.Up == n.Down }

// natureGrid lays out the 25 natures as in the main games: the row is the raised stat,
// the column the lowered one, in the order Attack, Defense, Magic Attack, Magic Defense, Speed.
var natureGrid = [5][5]string{
	{"Hardy", "Lonely", "Adamant", "Naughty", "Brave"},
	{"Bold", "Docile", "Impish", "Lax", "Relaxed"},
	{"Modest", "Mild", "Bashful", "Rash", "Quiet"},
	{"Calm", "Gentle", "Careful", "Quirky", "Sassy"},
	{"Timid", "Hasty", "Jolly", "Naive", "Serious"},
}

// Natures is all 25 natures, built from natureGrid.
var Natures = func() []Nature {
	order := []Stat{StatAttack, StatDefense, StatMagicAttack, StatMagicDefense, StatSpeed}
	var ns []Nature
	for up, row := range natureGrid {
		for down, name := range row {
			ns = append(ns, Nature{ID: strings.ToLower(name), Name: name, Up: order[up], Down: order[down]})
		}
	}
	return ns
}()

var naturesByID = func() map[string]Nature {
	m := make(map[string]Nature, len(Natures))
	for _, n := range Natures {
		m[n.ID] = n
	}
	return m
}()

// NatureByID returns the nature with that ID, and whether it was found.
func NatureByID(id string) (Nature, bool) {
	n, ok := naturesByID[id]
	return n, ok
}
