package entities

type Type int

const (
	TypeNone     Type = iota // 0
	TypeNormal               // 1
	TypeFighting             // 2
	TypeFlying               // 3
	TypePoison               // 4
	TypeGround               // 5
	TypeRock                 // 6
	TypeBug                  // 7
	TypeGhost                // 8
	TypeSteel                // 9
	TypeFire                 // 10
	TypeWater                // 11
	TypeGrass                // 12
	TypeElectric             // 13
	TypePsychic              // 14
	TypeIce                  // 15
	TypeDragon               // 16
	TypeDark                 // 17
	TypeFairy                // 18
)

var Himmel = NewBaseCharacter(1, "Himmel", types(TypeFlying, TypeFairy), 70, 155, 80, 90, 70, 135)
var Frieren = NewBaseCharacter(2, "Frieren", types(TypeIce, TypeElectric), 70, 90, 55, 155, 135, 95)
var Eisen = NewBaseCharacter(3, "Eisen", types(TypeSteel, TypeFighting), 110, 125, 130, 80, 95, 60)
var Heiter = NewBaseCharacter(4, "Heiter", types(TypeNormal, TypePoison), 135, 95, 100, 125, 110, 35)
var Fern = NewBaseCharacter(5, "Fern", types(TypeWater, TypeElectric), 70, 80, 55, 135, 60, 130)
var Stark = NewBaseCharacter(6, "Stark", types(TypeFire, TypeSteel), 110, 125, 70, 80, 70, 75)
var Sein = NewBaseCharacter(7, "Sein", types(TypeGrass, TypePoison), 130, 85, 90, 95, 90, 40)
var Ubel = NewBaseCharacter(8, "Übel", types(TypeDark, TypeNone), 50, 65, 50, 135, 65, 115)
var Land = NewBaseCharacter(9, "Land", types(TypeGround, TypeGhost), 55, 50, 80, 110, 105, 90)
var Denken = NewBaseCharacter(10, "Denken", types(TypePsychic, TypeNone), 110, 85, 80, 120, 85, 30)
var Flamme = NewBaseCharacter(11, "Flamme", types(TypeFire, TypeFairy), 100, 100, 90, 150, 140, 90)
var Serie = NewBaseCharacter(12, "Serie", types(TypeNormal, TypeNone), 70, 100, 60, 170, 170, 100)

func types(type0, type1 Type) []Type {
	if type1 == 0 {
		return []Type{type0, 0}
	}

	return []Type{type0, type1}
}
