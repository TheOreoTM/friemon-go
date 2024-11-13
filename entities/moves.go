package entities

var SacredSword = NewMove(1, "Sacred Sword", moveCreateParams{
	Type:            Fighting,
	Category:        MoveCatPhysical,
	Power:           90,
	Accuracy:        100,
	PP:              15,
	Priority:        0,
	Effect:          nil,
	SecondaryEffect: nil,
	Target:          TargetSingleFoes,
	IgnoreDefensive: true,
	IgnoreEvasion:   true,
})

var Aeroblast = NewMove(2, "Aeroblast", moveCreateParams{
	Type:            Flying,
	Category:        MoveCatSpecial,
	Power:           100,
	Accuracy:        95,
	PP:              5,
	Priority:        0,
	CritRatio:       2,
	Target:          TargetAny,
	Effect:          nil,
	SecondaryEffect: nil,
})

var AirCutter = NewMove(3, "Air Cutter", moveCreateParams{
	Type:            Flying,
	Category:        MoveCatSpecial,
	Power:           60,
	PP:              25,
	Accuracy:        95,
	Priority:        0,
	SecondaryEffect: nil,
	Target:          TargetAllAdjacentFoes,
	CritRatio:       2,
})
