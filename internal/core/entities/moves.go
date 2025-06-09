package entities

import "github.com/theoreotm/friemon/constants"

// Moves registry
var MovesRegistry = map[int]Move{}

// Initialize all moves
func init() {
	initializeMoves()
}

func initializeMoves() {
	// Physical Moves
	MovesRegistry[1] = SacredSword
	MovesRegistry[2] = Tackle
	MovesRegistry[3] = Scratch
	MovesRegistry[4] = QuickAttack
	MovesRegistry[5] = SwordsDance
	MovesRegistry[6] = BodySlam
	MovesRegistry[7] = Earthquake
	MovesRegistry[8] = RockSlide
	MovesRegistry[9] = IronHead
	MovesRegistry[10] = FlameWheel

	// Special Moves
	MovesRegistry[11] = Aeroblast
	MovesRegistry[12] = AirCutter
	MovesRegistry[13] = Flamethrower
	MovesRegistry[14] = Hydropump
	MovesRegistry[15] = SolarBeam
	MovesRegistry[16] = Thunderbolt
	MovesRegistry[17] = Psychic
	MovesRegistry[18] = IceBeam
	MovesRegistry[19] = DragonPulse
	MovesRegistry[20] = ShadowBall

	// Status Moves
	MovesRegistry[21] = ThunderWave
	MovesRegistry[22] = Toxic
	MovesRegistry[23] = WillOWisp
	MovesRegistry[24] = SleepPowder
	MovesRegistry[25] = Recover
	MovesRegistry[26] = DoubleTeam
	MovesRegistry[27] = Protect
	MovesRegistry[28] = Substitute
	MovesRegistry[29] = Rest
	MovesRegistry[30] = Confuse
}

// Physical Moves
var SacredSword = NewMove(1, "Sacred Sword", moveCreateParams{
	Type:              TypeFighting,
	Category:          MoveCatPhysical,
	Power:             90,
	Accuracy:          100,
	PP:                15,
	Priority:          0,
	Description:       "An attack that ignores the target's stat changes.",
	Target:            TargetSingleFoe,
	MakesContact:      true,
	AffectedByProtect: true,
	Effect: &EffectType{
		IgnoreDefense: true,
	},
})

var Tackle = NewMove(2, "Tackle", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatPhysical,
	Power:             40,
	Accuracy:          100,
	PP:                35,
	Priority:          0,
	Description:       "A physical attack in which the user charges and slams into the target.",
	Target:            TargetSingleFoe,
	MakesContact:      true,
	AffectedByProtect: true,
})

var Scratch = NewMove(3, "Scratch", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatPhysical,
	Power:             40,
	Accuracy:          100,
	PP:                35,
	Priority:          0,
	Description:       "Hard, pointed, sharp claws rake the target to inflict damage.",
	Target:            TargetSingleFoe,
	MakesContact:      true,
	AffectedByProtect: true,
})

var QuickAttack = NewMove(4, "Quick Attack", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatPhysical,
	Power:             40,
	Accuracy:          100,
	PP:                30,
	Priority:          1,
	Description:       "The user lunges at the target at a speed that makes it almost invisible.",
	Target:            TargetSingleFoe,
	MakesContact:      true,
	AffectedByProtect: true,
})

var SwordsDance = NewMove(5, "Swords Dance", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          100,
	PP:                20,
	Priority:          0,
	Description:       "A frenetic dance to uplift the fighting spirit. This sharply raises the user's Attack stat.",
	Target:            TargetUser,
	AffectedByProtect: false,
	Effect: &EffectType{
		SelfStatModifiers: StatChanges{"atk": 2},
		SelfStatChance:    100,
	},
})

var BodySlam = NewMove(6, "Body Slam", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatPhysical,
	Power:             85,
	Accuracy:          100,
	PP:                15,
	Priority:          0,
	Description:       "The user drops onto the target with its full body weight. This may also leave the target with paralysis.",
	Target:            TargetSingleFoe,
	MakesContact:      true,
	AffectedByProtect: true,
	SecondaryEffect: &EffectType{
		StatusCondition: constants.StatusParalyze,
		StatusChance:    30,
	},
})

var Earthquake = NewMove(7, "Earthquake", moveCreateParams{
	Type:              TypeGround,
	Category:          MoveCatPhysical,
	Power:             100,
	Accuracy:          100,
	PP:                10,
	Priority:          0,
	Description:       "The user sets off an earthquake that strikes every Pokémon around it.",
	Target:            TargetAllAdjacent,
	AffectedByProtect: true,
})

var RockSlide = NewMove(8, "Rock Slide", moveCreateParams{
	Type:              TypeRock,
	Category:          MoveCatPhysical,
	Power:             75,
	Accuracy:          90,
	PP:                10,
	Priority:          0,
	Description:       "Large boulders are hurled at the opposing Pokémon to inflict damage. This may also make the opposing Pokémon flinch.",
	Target:            TargetAllFoes,
	AffectedByProtect: true,
	SecondaryEffect: &EffectType{
		Flinch:       true,
		FlinchChance: 30,
	},
})

var IronHead = NewMove(9, "Iron Head", moveCreateParams{
	Type:              TypeSteel,
	Category:          MoveCatPhysical,
	Power:             80,
	Accuracy:          100,
	PP:                15,
	Priority:          0,
	Description:       "The user slams the target with its steel-hard head. This may also make the target flinch.",
	Target:            TargetSingleFoe,
	MakesContact:      true,
	AffectedByProtect: true,
	SecondaryEffect: &EffectType{
		Flinch:       true,
		FlinchChance: 30,
	},
})

var FlameWheel = NewMove(10, "Flame Wheel", moveCreateParams{
	Type:              TypeFire,
	Category:          MoveCatPhysical,
	Power:             60,
	Accuracy:          100,
	PP:                25,
	Priority:          0,
	Description:       "The user cloaks itself in fire and charges at the target. This may also leave the target with a burn.",
	Target:            TargetSingleFoe,
	MakesContact:      true,
	AffectedByProtect: true,
	SecondaryEffect: &EffectType{
		StatusCondition: constants.StatusBurn,
		StatusChance:    10,
	},
})

// Special Moves
var Aeroblast = NewMove(11, "Aeroblast", moveCreateParams{
	Type:              TypeFlying,
	Category:          MoveCatSpecial,
	Power:             100,
	Accuracy:          95,
	PP:                5,
	Priority:          0,
	CritRatio:         2,
	Description:       "A vortex of air is shot at the target to inflict damage. Critical hits land more easily.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
})

var AirCutter = NewMove(12, "Air Cutter", moveCreateParams{
	Type:              TypeFlying,
	Category:          MoveCatSpecial,
	Power:             60,
	Accuracy:          95,
	PP:                25,
	Priority:          0,
	CritRatio:         2,
	Description:       "The user attacks with razor-like wind. Critical hits land more easily.",
	Target:            TargetAllFoes,
	AffectedByProtect: true,
})

var Flamethrower = NewMove(13, "Flamethrower", moveCreateParams{
	Type:              TypeFire,
	Category:          MoveCatSpecial,
	Power:             90,
	Accuracy:          100,
	PP:                15,
	Priority:          0,
	Description:       "The target is scorched with an intense blast of fire. This may also leave the target with a burn.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	SecondaryEffect: &EffectType{
		StatusCondition: constants.StatusBurn,
		StatusChance:    10,
	},
})

var Hydropump = NewMove(14, "Hydro Pump", moveCreateParams{
	Type:              TypeWater,
	Category:          MoveCatSpecial,
	Power:             110,
	Accuracy:          80,
	PP:                5,
	Priority:          0,
	Description:       "The target is blasted by a huge volume of water launched under great pressure.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
})

var SolarBeam = NewMove(15, "Solar Beam", moveCreateParams{
	Type:              TypeGrass,
	Category:          MoveCatSpecial,
	Power:             120,
	Accuracy:          100,
	PP:                10,
	Priority:          0,
	Description:       "A two-turn attack. The user gathers light, then blasts a bundled beam on the next turn.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	Effect: &EffectType{
		ChargeRequired: true,
	},
})

var Thunderbolt = NewMove(16, "Thunderbolt", moveCreateParams{
	Type:              TypeElectric,
	Category:          MoveCatSpecial,
	Power:             90,
	Accuracy:          100,
	PP:                15,
	Priority:          0,
	Description:       "A strong electric blast crashes down on the target. This may also leave the target with paralysis.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	SecondaryEffect: &EffectType{
		StatusCondition: constants.StatusParalyze,
		StatusChance:    10,
	},
})

var Psychic = NewMove(17, "Psychic", moveCreateParams{
	Type:              TypePsychic,
	Category:          MoveCatSpecial,
	Power:             90,
	Accuracy:          100,
	PP:                10,
	Priority:          0,
	Description:       "The target is hit by a strong telekinetic force. This may also lower the target's Sp. Def stat.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	SecondaryEffect: &EffectType{
		StatModifiers: StatChanges{"sdef": -1},
		StatChance:    10,
	},
})

var IceBeam = NewMove(18, "Ice Beam", moveCreateParams{
	Type:              TypeIce,
	Category:          MoveCatSpecial,
	Power:             90,
	Accuracy:          100,
	PP:                10,
	Priority:          0,
	Description:       "The target is struck with an icy-cold beam of energy. This may also leave the target frozen.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	SecondaryEffect: &EffectType{
		StatusCondition: constants.StatusFreeze,
		StatusChance:    10,
	},
})

var DragonPulse = NewMove(19, "Dragon Pulse", moveCreateParams{
	Type:              TypeDragon,
	Category:          MoveCatSpecial,
	Power:             85,
	Accuracy:          100,
	PP:                10,
	Priority:          0,
	Description:       "The target is attacked with a shock wave generated by the user's gaping mouth.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	Pulse:             true,
})

var ShadowBall = NewMove(20, "Shadow Ball", moveCreateParams{
	Type:              TypeGhost,
	Category:          MoveCatSpecial,
	Power:             80,
	Accuracy:          100,
	PP:                15,
	Priority:          0,
	Description:       "The user hurls a shadowy blob at the target. This may also lower the target's Sp. Def stat.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	Ballistic:         true,
	SecondaryEffect: &EffectType{
		StatModifiers: StatChanges{"sdef": -1},
		StatChance:    20,
	},
})

// Status Moves
var ThunderWave = NewMove(21, "Thunder Wave", moveCreateParams{
	Type:              TypeElectric,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          90,
	PP:                20,
	Priority:          0,
	Description:       "The user launches a weak jolt of electricity that paralyzes the target.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	Effect: &EffectType{
		StatusCondition: constants.StatusParalyze,
		StatusChance:    100,
	},
})

var Toxic = NewMove(22, "Toxic", moveCreateParams{
	Type:              TypePoison,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          90,
	PP:                10,
	Priority:          0,
	Description:       "A move that leaves the target badly poisoned. Its poison damage worsens every turn.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	Effect: &EffectType{
		StatusCondition: constants.StatusPoison,
		StatusChance:    100,
	},
})

var WillOWisp = NewMove(23, "Will-O-Wisp", moveCreateParams{
	Type:              TypeFire,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          85,
	PP:                15,
	Priority:          0,
	Description:       "The user shoots a sinister, bluish-white flame at the target to inflict a burn.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	Effect: &EffectType{
		StatusCondition: constants.StatusBurn,
		StatusChance:    100,
	},
})

var SleepPowder = NewMove(24, "Sleep Powder", moveCreateParams{
	Type:              TypeGrass,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          75,
	PP:                15,
	Priority:          0,
	Description:       "The user scatters a big cloud of sleep-inducing dust around the target.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	Effect: &EffectType{
		StatusCondition: constants.StatusSleep,
		StatusChance:    100,
	},
})

var Recover = NewMove(25, "Recover", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          100,
	PP:                5,
	Priority:          0,
	Description:       "Restoring its own cells, the user restores its own HP by half of its max HP.",
	Target:            TargetUser,
	AffectedByProtect: false,
	Effect: &EffectType{
		HealPercentage: 50,
	},
})

var DoubleTeam = NewMove(26, "Double Team", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          100,
	PP:                15,
	Priority:          0,
	Description:       "By moving rapidly, the user makes illusory copies of itself to raise its evasiveness.",
	Target:            TargetUser,
	AffectedByProtect: false,
	Effect: &EffectType{
		SelfStatModifiers: StatChanges{"eva": 1},
		SelfStatChance:    100,
	},
})

var Protect = NewMove(27, "Protect", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          100,
	PP:                10,
	Priority:          4,
	Description:       "Enables the user to evade all attacks. Its chance of failing rises if it is used in succession.",
	Target:            TargetUser,
	AffectedByProtect: false,
	Effect: &EffectType{
		ProtectsUser: true,
	},
})

var Substitute = NewMove(28, "Substitute", moveCreateParams{
	Type:              TypeNormal,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          100,
	PP:                10,
	Priority:          0,
	Description:       "The user makes a copy of itself using some of its HP. The copy serves as the user's decoy.",
	Target:            TargetUser,
	AffectedByProtect: false,
	Effect: &EffectType{
		FixedDamage: 25, // 25% of max HP
	},
})

var Rest = NewMove(29, "Rest", moveCreateParams{
	Type:              TypePsychic,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          100,
	PP:                5,
	Priority:          0,
	Description:       "The user goes to sleep for two turns. This fully restores the user's HP and heals any status conditions.",
	Target:            TargetUser,
	AffectedByProtect: false,
	Effect: &EffectType{
		HealPercentage:  100,
		StatusCondition: constants.StatusSleep,
		StatusChance:    100,
	},
})

var Confuse = NewMove(30, "Confuse Ray", moveCreateParams{
	Type:              TypeGhost,
	Category:          MoveCatStatus,
	Power:             0,
	Accuracy:          100,
	PP:                10,
	Priority:          0,
	Description:       "The target is exposed to a sinister ray that triggers confusion.",
	Target:            TargetSingleFoe,
	AffectedByProtect: true,
	Effect: &EffectType{
		StatusCondition: constants.StatusConfuse,
		StatusChance:    100,
	},
})

// Helper function to get a move by ID
func GetMoveByID(id int) (Move, bool) {
	move, exists := MovesRegistry[id]
	return move, exists
}

// Helper function to get all moves
func GetAllMoves() []Move {
	moves := make([]Move, 0, len(MovesRegistry))
	for _, move := range MovesRegistry {
		moves = append(moves, move)
	}
	return moves
}

// Helper function to get moves by type
func GetMovesByType(moveType Type) []Move {
	var moves []Move
	for _, move := range MovesRegistry {
		if move.Type == moveType {
			moves = append(moves, move)
		}
	}
	return moves
}

// Helper function to get moves by category
func GetMovesByCategory(category MoveCategory) []Move {
	var moves []Move
	for _, move := range MovesRegistry {
		if move.Category == category {
			moves = append(moves, move)
		}
	}
	return moves
}
