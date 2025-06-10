package game

// Type effectiveness multipliers
const (
	SuperEffective   = 2.0
	NotVeryEffective = 0.5
	NoEffect         = 0.0
	NormalEffective  = 1.0
)

// TypeEffectiveness represents the effectiveness of attacking type vs defending type
var TypeEffectiveness = map[Type]map[Type]float64{
	TypeNormal: {
		TypeRock:  NotVeryEffective,
		TypeGhost: NoEffect,
		TypeSteel: NotVeryEffective,
	},
	TypeFighting: {
		TypeNormal:   SuperEffective,
		TypeFlying:   NotVeryEffective,
		TypePoison:   NotVeryEffective,
		TypeRock:     SuperEffective,
		TypeBug:      NotVeryEffective,
		TypeGhost:    NoEffect,
		TypeSteel:    SuperEffective,
		TypeFire:     NormalEffective,
		TypeWater:    NormalEffective,
		TypeGrass:    NormalEffective,
		TypeElectric: NormalEffective,
		TypePsychic:  NotVeryEffective,
		TypeIce:      SuperEffective,
		TypeDragon:   NormalEffective,
		TypeDark:     SuperEffective,
		TypeFairy:    NotVeryEffective,
	},
	TypeFlying: {
		TypeFighting: SuperEffective,
		TypeRock:     NotVeryEffective,
		TypeBug:      SuperEffective,
		TypeSteel:    NotVeryEffective,
		TypeGrass:    SuperEffective,
		TypeElectric: NotVeryEffective,
	},
	TypePoison: {
		TypePoison: NotVeryEffective,
		TypeGround: NotVeryEffective,
		TypeRock:   NotVeryEffective,
		TypeGhost:  NotVeryEffective,
		TypeSteel:  NoEffect,
		TypeGrass:  SuperEffective,
		TypeFairy:  SuperEffective,
	},
	TypeGround: {
		TypeFlying:   NoEffect,
		TypePoison:   SuperEffective,
		TypeBug:      NotVeryEffective,
		TypeSteel:    SuperEffective,
		TypeFire:     SuperEffective,
		TypeGrass:    NotVeryEffective,
		TypeElectric: SuperEffective,
	},
	TypeRock: {
		TypeFighting: NotVeryEffective,
		TypeFlying:   SuperEffective,
		TypeGround:   NotVeryEffective,
		TypeBug:      SuperEffective,
		TypeSteel:    NotVeryEffective,
		TypeFire:     SuperEffective,
		TypeIce:      SuperEffective,
	},
	TypeBug: {
		TypeFighting: NotVeryEffective,
		TypeFlying:   NotVeryEffective,
		TypePoison:   NotVeryEffective,
		TypeGhost:    NotVeryEffective,
		TypeSteel:    NotVeryEffective,
		TypeFire:     NotVeryEffective,
		TypeGrass:    SuperEffective,
		TypePsychic:  SuperEffective,
		TypeDark:     SuperEffective,
		TypeFairy:    NotVeryEffective,
	},
	TypeGhost: {
		TypeNormal:  NoEffect,
		TypeGhost:   SuperEffective,
		TypePsychic: SuperEffective,
		TypeDark:    NotVeryEffective,
	},
	TypeSteel: {
		TypeRock:     SuperEffective,
		TypeSteel:    NotVeryEffective,
		TypeFire:     NotVeryEffective,
		TypeWater:    NotVeryEffective,
		TypeElectric: NotVeryEffective,
		TypeIce:      SuperEffective,
		TypeFairy:    SuperEffective,
	},
	TypeFire: {
		TypeRock:   NotVeryEffective,
		TypeBug:    SuperEffective,
		TypeSteel:  SuperEffective,
		TypeFire:   NotVeryEffective,
		TypeWater:  NotVeryEffective,
		TypeGrass:  SuperEffective,
		TypeIce:    SuperEffective,
		TypeDragon: NotVeryEffective,
	},
	TypeWater: {
		TypeGround: SuperEffective,
		TypeRock:   SuperEffective,
		TypeFire:   SuperEffective,
		TypeWater:  NotVeryEffective,
		TypeGrass:  NotVeryEffective,
		TypeDragon: NotVeryEffective,
	},
	TypeGrass: {
		TypeFlying: NotVeryEffective,
		TypePoison: NotVeryEffective,
		TypeGround: SuperEffective,
		TypeRock:   SuperEffective,
		TypeBug:    NotVeryEffective,
		TypeSteel:  NotVeryEffective,
		TypeFire:   NotVeryEffective,
		TypeWater:  SuperEffective,
		TypeGrass:  NotVeryEffective,
		TypeDragon: NotVeryEffective,
	},
	TypeElectric: {
		TypeFlying:   SuperEffective,
		TypeGround:   NoEffect,
		TypeWater:    SuperEffective,
		TypeGrass:    NotVeryEffective,
		TypeElectric: NotVeryEffective,
		TypeDragon:   NotVeryEffective,
	},
	TypePsychic: {
		TypeFighting: SuperEffective,
		TypePoison:   SuperEffective,
		TypeSteel:    NotVeryEffective,
		TypePsychic:  NotVeryEffective,
		TypeDark:     NoEffect,
	},
	TypeIce: {
		TypeFlying: SuperEffective,
		TypeGround: SuperEffective,
		TypeGrass:  SuperEffective,
		TypeSteel:  NotVeryEffective,
		TypeFire:   NotVeryEffective,
		TypeWater:  NotVeryEffective,
		TypeIce:    NotVeryEffective,
		TypeDragon: SuperEffective,
	},
	TypeDragon: {
		TypeSteel:  NotVeryEffective,
		TypeDragon: SuperEffective,
		TypeFairy:  NoEffect,
	},
	TypeDark: {
		TypeFighting: NotVeryEffective,
		TypeGhost:    SuperEffective,
		TypePsychic:  SuperEffective,
		TypeDark:     NotVeryEffective,
		TypeFairy:    NotVeryEffective,
	},
	TypeFairy: {
		TypeFighting: SuperEffective,
		TypePoison:   NotVeryEffective,
		TypeSteel:    NotVeryEffective,
		TypeFire:     NotVeryEffective,
		TypeDragon:   SuperEffective,
		TypeDark:     SuperEffective,
	},
}

// GetTypeEffectiveness returns the effectiveness multiplier for an attacking type vs defending types
func GetTypeEffectiveness(attackingType Type, defendingType1, defendingType2 Type) float64 {
	effectiveness1 := getEffectivenessVsSingleType(attackingType, defendingType1)

	if defendingType2 == TypeNone {
		return effectiveness1
	}

	effectiveness2 := getEffectivenessVsSingleType(attackingType, defendingType2)
	return effectiveness1 * effectiveness2
}

func getEffectivenessVSingleType(attackingType, defendingType Type) float64 {
	if defendingType == TypeNone {
		return NormalEffective
	}

	if typeChart, exists := TypeEffectiveness[attackingType]; exists {
		if effectiveness, exists := typeChart[defendingType]; exists {
			return effectiveness
		}
	}

	return NormalEffective
}

// Helper function (fix typo in the above function)
func getEffectivenessVsSingleType(attackingType, defendingType Type) float64 {
	if defendingType == TypeNone {
		return NormalEffective
	}

	if typeChart, exists := TypeEffectiveness[attackingType]; exists {
		if effectiveness, exists := typeChart[defendingType]; exists {
			return effectiveness
		}
	}

	return NormalEffective
}

// GetEffectivenessText returns a human-readable description of type effectiveness
func GetEffectivenessText(effectiveness float64) string {
	switch effectiveness {
	case SuperEffective:
		return "It's super effective!"
	case NotVeryEffective:
		return "It's not very effective..."
	case NoEffect:
		return "It had no effect!"
	default:
		return ""
	}
}
