package data

const (
	Normal    = "Normal"
	Psychic   = "Psychic"
	Water     = "Water"
	Fire      = "Fire"
	Fairy     = "Fairy"
	Ice       = "Ice"
	Earth     = "Earth"
	Steel     = "Steel"
	Fighting  = "Fighting"
	Electric  = "Electric"
	Grass     = "Grass"
	Poison    = "Poison"
	Lightning = "Lightning"
	Dark      = "Dark"
	Flying    = "Flying"
	Dragon    = "Dragon"
	Ground    = "Ground"
	Ghost     = "Ghost"
)

var Himmel = NewBaseCharacter(1, "Himmel", types(Flying, Fairy), 70, 155, 80, 90, 70, 135)
var Frieren = NewBaseCharacter(2, "Frieren", types(Ice, Electric), 70, 90, 55, 155, 135, 95)
var Eisen = NewBaseCharacter(3, "Eisen", types(Steel, Fighting), 110, 125, 130, 80, 95, 60)
var Heiter = NewBaseCharacter(4, "Heiter", types(Normal, Poison), 135, 95, 100, 125, 110, 35)
var Fern = NewBaseCharacter(5, "Fern", types(Water, Electric), 70, 80, 55, 135, 60, 130)
var Stark = NewBaseCharacter(6, "Stark", types(Fire, Steel), 110, 125, 70, 80, 70, 75)
var Sein = NewBaseCharacter(7, "Sein", types(Grass, Poison), 130, 85, 90, 95, 90, 40)
var Ubel = NewBaseCharacter(8, "Übel", types(Dark, ""), 50, 65, 50, 135, 65, 115)
var Land = NewBaseCharacter(9, "Land", types(Ground, Ghost), 55, 50, 80, 110, 105, 90)
var Denken = NewBaseCharacter(10, "Denken", types(Psychic, ""), 110, 85, 80, 120, 85, 30)
var Flamme = NewBaseCharacter(11, "Flamme", types(Fire, Fairy), 100, 100, 90, 150, 140, 90)
var Serie = NewBaseCharacter(12, "Serie", types(Normal, ""), 70, 100, 60, 170, 170, 100)

func types(type0, type1 string) []string {
	if type1 == "" {
		return []string{type0, ""}
	}

	return []string{type0, type1}
}
