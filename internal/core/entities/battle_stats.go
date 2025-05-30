package entities

import "github.com/theoreotm/friemon/constants"

type BattleStats struct {
	CurrentHP     int
	AtkMod        int
	DefMod        int
	SpAtkMod      int
	SpDefMod      int
	SpeMod        int
	StatusEffects []constants.StatusEffect // Conditions like Burn, Paralysis
	TurnPriority  int                      // Used to determine move order
	TempModifiers map[string]int           // Temporary stat boosts
}

func (b *BattleStats) HasStatusEffect(effect constants.StatusEffect) bool {
	for _, e := range b.StatusEffects {
		if e == effect {
			return true
		}
	}
	return false
}

func (b *BattleStats) AddStatusEffect(effect constants.StatusEffect) {
	if !b.HasStatusEffect(effect) {
		b.StatusEffects = append(b.StatusEffects, effect)
	}
}

func (b *BattleStats) RemoveStatusEffect(effect constants.StatusEffect) {
	for i, e := range b.StatusEffects {
		if e == effect {
			b.StatusEffects = append(b.StatusEffects[:i], b.StatusEffects[i+1:]...)
			break
		}
	}
}
