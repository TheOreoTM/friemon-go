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
	Target:          TargetSingleOpponent,
})
