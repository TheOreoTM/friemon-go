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
	Dragon    = "Dragon"
)

var Frieren = NewBaseCharacter(1, "Frieren", types(Ice, Electric), 73, 67, 75, 81, 100, 109)
var Himmel = NewBaseCharacter(2, "Himmel", types(Fairy, Steel), 55, 150, 115, 80, 115, 148)
var Heiter = NewBaseCharacter(3, "Heiter", types(Fairy, ""), 95, 70, 73, 95, 90, 60)
var Flamme = NewBaseCharacter(4, "Flamme", types(Dragon, Fire), 100, 120, 100, 150, 120, 90)
var Serie = NewBaseCharacter(5, "Serie", types(Normal, ""), 120, 120, 120, 120, 120, 120)
var Eisen = NewBaseCharacter(6, "Eisen", types(Fighting, Steel), 92, 120, 140, 80, 140, 128)
var Sein = NewBaseCharacter(7, "Sein", types(Grass, Poison), 114, 85, 70, 85, 80, 30)
var Stark = NewBaseCharacter(8, "Stark", types(Fighting, ""), 105, 140, 95, 55, 65, 45)
var Fern = NewBaseCharacter(9, "Fern", types(Psychic, ""), 95, 60, 60, 101, 60, 105)

func types(type0, type1 string) []string {
	if type1 == "" {
		return []string{type0, ""}
	}

	return []string{type0, type1}
}
