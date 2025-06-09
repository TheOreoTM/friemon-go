package entities

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/theoreotm/friemon/constants"
)

type DamageResult struct {
	Damage              int     `json:"damage"`
	IsCritical          bool    `json:"is_critical"`
	TypeEffectiveness   float64 `json:"type_effectiveness"`
	EffectivenessText   string  `json:"effectiveness_text"`
	Hit                 bool    `json:"hit"`
	StatusEffectApplied bool    `json:"status_effect_applied"`
	StatChangeApplied   bool    `json:"stat_change_applied"`
	CalculationDetails  string  `json:"calculation_details,omitempty"`
}

func CalculateDamage(attacker, defender *Character, move Move, settings GameSettings) DamageResult {
	result := DamageResult{
		Hit: true,
	}

	// Check accuracy first
	if !checkAccuracy(attacker, defender, move, settings) {
		result.Hit = false
		return result
	}

	// Status moves don't deal damage
	if move.Category == MoveCatStatus {
		result.Damage = 0
		return result
	}

	// Fixed damage moves
	if move.Effect != nil && move.Effect.FixedDamage > 0 {
		result.Damage = move.Effect.FixedDamage
		return result
	}

	// Calculate base damage
	level := float64(attacker.Level)
	power := float64(move.Power)

	var attack, defense float64

	// Determine attack and defense stats based on move category
	if move.Category == MoveCatPhysical {
		attack = float64(attacker.Atk())
		defense = float64(defender.Def())

		// Apply stat stages if enabled
		if settings.StatStagesEnabled {
			attack *= attacker.BattleStats.GetStatMultiplier("atk")
			defense *= defender.BattleStats.GetStatMultiplier("def")
		}

		// Burn halves physical attack
		if attacker.BattleStats.HasStatusEffect(constants.StatusBurn) && settings.StatusEffectsEnabled {
			attack *= 0.5
		}

	} else if move.Category == MoveCatSpecial {
		attack = float64(attacker.SpAtk())
		defense = float64(defender.SpDef())

		// Apply stat stages if enabled
		if settings.StatStagesEnabled {
			attack *= attacker.BattleStats.GetStatMultiplier("satk")
			defense *= defender.BattleStats.GetStatMultiplier("sdef")
		}
	}

	// Check for critical hit
	if settings.CriticalHitsEnabled {
		result.IsCritical = checkCriticalHit(attacker, move)
		if result.IsCritical {
			// Critical hits ignore negative stat changes for attacker and positive for defender
			if move.Category == MoveCatPhysical {
				if attacker.BattleStats.AtkStage < 0 {
					attack = float64(attacker.Atk())
				}
				if defender.BattleStats.DefStage > 0 {
					defense = float64(defender.Def())
				}
			} else {
				if attacker.BattleStats.SpAtkStage < 0 {
					attack = float64(attacker.SpAtk())
				}
				if defender.BattleStats.SpDefStage > 0 {
					defense = float64(defender.SpDef())
				}
			}
		}
	}

	// Base damage calculation (Pokemon formula)
	baseDamage := ((((2*level/5 + 2) * power * attack / defense) / 50) + 2)

	// Apply STAB (Same Type Attack Bonus)
	stab := 1.0
	attackerData := attacker.Data()
	if attackerData.Type0 == move.Type || attackerData.Type1 == move.Type {
		stab = 1.5
	}

	// Apply type effectiveness
	typeEffectiveness := 1.0
	if settings.TypeEffectivenessEnabled {
		defenderData := defender.Data()
		typeEffectiveness = GetTypeEffectiveness(move.Type, defenderData.Type0, defenderData.Type1)
		result.TypeEffectiveness = typeEffectiveness
		result.EffectivenessText = GetEffectivenessText(typeEffectiveness)
	}

	// Apply critical hit multiplier
	critMultiplier := 1.0
	if result.IsCritical {
		critMultiplier = 1.5
	}

	// Random factor (85-100%)
	randomFactor := (float64(rand.Intn(16)) + 85) / 100

	// Final damage calculation
	finalDamage := baseDamage * stab * typeEffectiveness * critMultiplier * randomFactor

	// Ensure minimum damage of 1 if move has power
	if power > 0 && finalDamage < 1 {
		finalDamage = 1
	}

	result.Damage = int(math.Round(finalDamage))

	// Build calculation details if requested
	if settings.ShowDamageCalculation {
		result.CalculationDetails = buildCalculationDetails(
			level, power, attack, defense, stab, typeEffectiveness,
			critMultiplier, randomFactor, finalDamage,
		)
	}

	return result
}

func checkAccuracy(attacker, defender *Character, move Move, settings GameSettings) bool {
	accuracy := float64(move.Accuracy)

	// Perfect accuracy moves
	if move.Effect != nil && move.Effect.NeverMiss {
		return true
	}

	// Apply accuracy/evasion stat stages if enabled
	if settings.StatStagesEnabled {
		accMultiplier := attacker.BattleStats.GetStatMultiplier("acc")
		evaMultiplier := defender.BattleStats.GetStatMultiplier("eva")
		accuracy = accuracy * accMultiplier / evaMultiplier
	}

	// Cap accuracy at 100%
	if accuracy > 100 {
		accuracy = 100
	}

	roll := rand.Intn(100) + 1
	return roll <= int(accuracy)
}

func checkCriticalHit(attacker *Character, move Move) bool {
	// Base critical hit rate is 1/24 (about 4.17%)
	critRate := 1.0 / 24.0

	// Move-specific critical hit ratio
	switch move.CritRatio {
	case 2:
		critRate = 1.0 / 8.0 // 12.5%
	case 3:
		critRate = 1.0 / 2.0 // 50%
	case 4:
		critRate = 1.0 // Always crit
	}

	// High critical hit ratio moves
	if move.Effect != nil && move.Effect.HighCritRatio {
		critRate = 1.0 / 8.0
	}

	// Always critical moves
	if move.Effect != nil && move.Effect.AlwaysCritical {
		critRate = 1.0
	}

	// Apply any temporary crit boosts
	critRate += float64(attacker.BattleStats.CritBoost) * (1.0 / 24.0)

	return rand.Float64() < critRate
}

func buildCalculationDetails(level, power, attack, defense, stab, typeEff, crit, random, final float64) string {
	return fmt.Sprintf(
		"Level: %.0f, Power: %.0f, Atk: %.1f, Def: %.1f, STAB: %.1fx, Type: %.1fx, Crit: %.1fx, Random: %.1f%%, Final: %.1f",
		level, power, attack, defense, stab, typeEff, crit, random*100, final,
	)
}
