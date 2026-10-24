package bot

import (
	"fmt"
	"strings"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
)

// battleLog turns engine events into the lines shown in a battle message.
func battleLog(ab *activeBattle, events []game.Event) string {
	var lines []string
	for _, e := range events {
		if line := eventText(ab, e); line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func eventText(ab *activeBattle, e game.Event) string {
	player := ab.players[e.Side].name
	who := fmt.Sprintf("%s's **%s**", player, e.Name)

	switch e.Kind {
	case game.EvSwitchIn:
		return fmt.Sprintf("%s sent out **%s**!", player, e.Name)
	case game.EvSwitchOut:
		return fmt.Sprintf("%s withdrew **%s**.", player, e.Name)
	case game.EvMove:
		return fmt.Sprintf("%s used **%s**!", who, e.Move)
	case game.EvStruggle:
		return who + " has no moves left!"
	case game.EvMiss:
		return "But it missed!"
	case game.EvFail:
		return "But it failed!"
	case game.EvNoTarget:
		return "But there was no target..."
	case game.EvCrit:
		return "A critical hit!"
	case game.EvEffectiveness:
		if e.Mult > 1 {
			return "It's super effective!"
		}
		return "It's not very effective..."
	case game.EvDamage:
		return fmt.Sprintf("%s took %d damage.", who, e.Amount)
	case game.EvHits:
		if e.Amount == 1 {
			return "Hit 1 time!"
		}
		return fmt.Sprintf("Hit %d times!", e.Amount)
	case game.EvHeal:
		return fmt.Sprintf("%s recovered %d HP.", who, e.Amount)
	case game.EvDrain:
		return fmt.Sprintf("%s drained %d HP.", who, e.Amount)
	case game.EvRecoil:
		return fmt.Sprintf("%s is hurt by recoil! (%d)", who, e.Amount)
	case game.EvFaint:
		return who + " fainted!"
	case game.EvCondition:
		return conditionInflicted(who, e.Condition)
	case game.EvConditionFailed:
		return who + " already has a status condition."
	case game.EvImmune:
		return "It doesn't affect " + who + "..."
	case game.EvConditionDamage:
		if e.Condition == data.Burn {
			return fmt.Sprintf("%s is hurt by its burn! (%d)", who, e.Amount)
		}
		return fmt.Sprintf("%s is hurt by the curse! (%d)", who, e.Amount)
	case game.EvCured:
		return fmt.Sprintf("%s is no longer %s.", who, strings.ToLower(conditionWord(e.Condition)))
	case game.EvFastAsleep:
		return who + " is fast asleep."
	case game.EvWokeUp:
		return who + " woke up!"
	case game.EvFrozenSolid:
		return who + " is frozen solid!"
	case game.EvThawed:
		return who + " thawed out!"
	case game.EvFullyBound:
		return who + " is bound and can't move!"
	case game.EvRested:
		return fmt.Sprintf("%s slept and became healthy! (+%d HP)", who, e.Amount)
	case game.EvConfused:
		return who + " became confused!"
	case game.EvAlreadyConfused:
		return who + " is already confused!"
	case game.EvIsConfused:
		return who + " is confused!"
	case game.EvConfusionHurt:
		return fmt.Sprintf("It hurt itself in its confusion! (%d)", e.Amount)
	case game.EvSnappedOut:
		return who + " snapped out of its confusion!"
	case game.EvFlinched:
		return who + " flinched and couldn't move!"
	case game.EvStatChange:
		return statChangeText(who, e)
	case game.EvWin:
		return fmt.Sprintf("**%s wins!**", player)

	case game.EvWeatherStart:
		return weatherStart[e.Weather]
	case game.EvWeatherEnd:
		return weatherEnd[e.Weather]
	case game.EvWeatherDamage:
		return fmt.Sprintf("%s is buffeted by the %s! (%d)", who, strings.ToLower(e.Weather.String()), e.Amount)
	case game.EvTerrainStart:
		return terrainStart[e.Terrain]
	case game.EvTerrainEnd:
		return fmt.Sprintf("The %s faded.", e.Terrain)
	case game.EvTerrainHeal:
		return fmt.Sprintf("%s was healed by the Overgrowth. (+%d)", who, e.Amount)
	case game.EvFieldBlocked:
		if e.Terrain != data.NoTerrain {
			return fmt.Sprintf("%s is protected by the %s!", who, e.Terrain)
		}
		return fmt.Sprintf("The %s protects %s!", e.Weather, who)
	case game.EvSideStart:
		return sideStartText(player, e)
	case game.EvSideEnd:
		return fmt.Sprintf("%s's %s wore off.", player, e.SideCond)
	case game.EvHazardsCleared:
		return "The hazards and wards were blown away!"
	case game.EvHazardDamage:
		return fmt.Sprintf("%s was hurt by the %s! (%d)", who, e.SideCond, e.Amount)
	case game.EvSealsAbsorbed:
		return who + " absorbed the Curse Seals!"
	case game.EvReversalStart:
		return "Time twisted! Slower characters will move first."
	case game.EvReversalEnd:
		return "Time returned to normal."
	case game.EvProtecting:
		return who + " protected itself!"
	case game.EvProtected:
		return who + " protected itself from the attack!"
	case game.EvSubstitute:
		return who + " made an illusory double!"
	case game.EvSubDamage:
		return fmt.Sprintf("The double took the hit for %s! (%d)", who, e.Amount)
	case game.EvSubBroke:
		return who + "'s double vanished!"
	case game.EvCharging:
		return fmt.Sprintf("%s is charging **%s**!", who, e.Move)
	case game.EvMustRecharge:
		return who + " must recharge!"
	case game.EvTrapped:
		return who + " can no longer escape!"
	case game.EvPartiallyTrapped:
		return who + " is caught in the vines!"
	case game.EvTrapDamage:
		return fmt.Sprintf("%s is hurt by the vines! (%d)", who, e.Amount)
	case game.EvTrapEnd:
		return who + " was freed!"
	case game.EvTaunted:
		return who + " fell for the provocation!"
	case game.EvTauntEnd:
		return who + " calmed down."
	case game.EvTauntBlocked:
		return fmt.Sprintf("%s can't use **%s** after the provocation!", who, e.Move)
	case game.EvEncored:
		return fmt.Sprintf("%s is compelled to repeat **%s**!", who, e.Move)
	case game.EvEncoreEnd:
		return who + " is free of the compulsion."
	}
	return ""
}

var weatherStart = map[data.Weather]string{
	data.Heatwave:  "A scorching heatwave began!",
	data.Storm:     "A storm rolled in!",
	data.Sandstorm: "A sandstorm kicked up!",
	data.Blizzard:  "A blizzard began to blow!",
}

var weatherEnd = map[data.Weather]string{
	data.Heatwave:  "The heatwave faded.",
	data.Storm:     "The storm passed.",
	data.Sandstorm: "The sandstorm died down.",
	data.Blizzard:  "The blizzard stopped.",
}

var terrainStart = map[data.Terrain]string{
	data.ManaField:    "Mana surged through the ground!",
	data.SacredGround: "The ground became sacred!",
	data.Overgrowth:   "Plants overran the battlefield!",
	data.Mist:         "Mist covered the battlefield!",
}

func sideStartText(player string, e game.Event) string {
	switch e.SideCond {
	case data.PhysicalWard:
		return fmt.Sprintf("A Physical Ward guards %s's side!", player)
	case data.MagicWard:
		return fmt.Sprintf("A Magic Ward guards %s's side!", player)
	case data.Tailwind:
		return fmt.Sprintf("A tailwind blows behind %s's side!", player)
	case data.ManaMines:
		return fmt.Sprintf("Mana mines were hidden around %s's side!", player)
	case data.Caltrops:
		return fmt.Sprintf("Caltrops were scattered around %s's side! (%d)", player, e.Amount)
	case data.CurseSeals:
		return fmt.Sprintf("Curse seals were planted around %s's side! (%d)", player, e.Amount)
	}
	return ""
}

func conditionInflicted(who string, c data.Condition) string {
	switch c {
	case data.Burn:
		return who + " was burned!"
	case data.Bound:
		return who + " is bound! It may be unable to move!"
	case data.Sleep:
		return who + " fell asleep!"
	case data.Freeze:
		return who + " was frozen solid!"
	case data.Cursed:
		return who + " was cursed!"
	case data.DeeplyCursed:
		return who + " was deeply cursed!"
	}
	return ""
}

// conditionWord describes a condition as a state, for status lines.
func conditionWord(c data.Condition) string {
	switch c {
	case data.Burn:
		return "Burned"
	case data.Bound:
		return "Bound"
	case data.Sleep:
		return "Asleep"
	case data.Freeze:
		return "Frozen"
	case data.Cursed:
		return "Cursed"
	case data.DeeplyCursed:
		return "Deeply cursed"
	}
	return ""
}

func statChangeText(who string, e game.Event) string {
	stat := e.Stat.String()
	if e.Amount == 0 {
		if e.Stages > 0 {
			return fmt.Sprintf("%s's %s won't go any higher!", who, stat)
		}
		return fmt.Sprintf("%s's %s won't go any lower!", who, stat)
	}
	n := e.Amount
	switch {
	case n >= 3:
		return fmt.Sprintf("%s's %s rose drastically!", who, stat)
	case n == 2:
		return fmt.Sprintf("%s's %s rose sharply!", who, stat)
	case n == 1:
		return fmt.Sprintf("%s's %s rose!", who, stat)
	case n == -1:
		return fmt.Sprintf("%s's %s fell!", who, stat)
	case n == -2:
		return fmt.Sprintf("%s's %s harshly fell!", who, stat)
	}
	return fmt.Sprintf("%s's %s severely fell!", who, stat)
}
