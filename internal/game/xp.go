package game

import (
	"math/rand/v2"
	"time"
)

// MaxLevel is the highest level a character can reach.
const MaxLevel = 100

// XPToNext is how much XP a character needs to go from level to level+1.
// Level 1 needs 75, level 50 needs 1300; reaching 100 takes about 128,000 XP in total.
func XPToNext(level int) int {
	return 50 + 25*level
}

// RollMessageXP is the XP the selected character gets for one counted message.
func RollMessageXP(r *rand.Rand) int {
	return 10 + r.IntN(11) // 10 to 20
}

// AddXP adds gain to a character and returns its new level and XP towards the next level.
// One big gain can level up several times. At MaxLevel, XP stays at 0.
func AddXP(level, xp, gain int) (newLevel, newXP int) {
	if level >= MaxLevel {
		return MaxLevel, 0
	}
	xp += gain
	for xp >= XPToNext(level) {
		xp -= XPToNext(level)
		level++
		if level == MaxLevel {
			return MaxLevel, 0
		}
	}
	return level, xp
}

// XPCooldown is how often a user's messages can give XP.
const XPCooldown = 1 * time.Second
