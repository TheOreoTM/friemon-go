package entities

type GameSettings struct {
	// Turn settings
	MaxTurns      int  `json:"max_turns"`
	TurnTimeLimit int  `json:"turn_time_limit"` // seconds
	SwitchingCost bool `json:"switching_cost"`  // costs a turn to switch

	// Battle mechanics
	CriticalHitsEnabled      bool `json:"critical_hits_enabled"`
	StatusEffectsEnabled     bool `json:"status_effects_enabled"`
	TypeEffectivenessEnabled bool `json:"type_effectiveness_enabled"`
	StatStagesEnabled        bool `json:"stat_stages_enabled"`

	// Team settings
	TeamSize        int  `json:"team_size"`
	AllowDuplicates bool `json:"allow_duplicates"`
	LevelCap        int  `json:"level_cap"`

	// ELO settings
	ELOEnabled bool `json:"elo_enabled"`
	ELOKFactor int  `json:"elo_k_factor"`

	// Debug settings
	ShowDamageCalculation bool `json:"show_damage_calculation"`
	ShowAccuracyRolls     bool `json:"show_accuracy_rolls"`
}

func DefaultGameSettings() GameSettings {
	return GameSettings{
		MaxTurns:                 25,
		TurnTimeLimit:            60,
		SwitchingCost:            true,
		CriticalHitsEnabled:      true,
		StatusEffectsEnabled:     true,
		TypeEffectivenessEnabled: true,
		StatStagesEnabled:        true,
		TeamSize:                 3,
		AllowDuplicates:          false,
		LevelCap:                 100,
		ELOEnabled:               true,
		ELOKFactor:               32,
		ShowDamageCalculation:    false,
		ShowAccuracyRolls:        false,
	}
}
