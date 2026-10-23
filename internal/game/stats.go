package game

import "github.com/TheOreoTM/friemon-go/internal/data"

const (
	MaxIV         = 31  // highest hidden roll for one stat
	MaxEVPerStat  = 252 // most EVs one stat can hold
	MaxEVTotal    = 510 // most EVs a character can hold across all stats
	natureBoost   = 110 // percent for the stat a nature raises
	natureLowered = 90  // percent for the stat a nature lowers
)

// CalcStats works out a character's real stats, using the main games' formula:
// base stats, hidden rolls (IVs), effort values (EVs, 4 EVs = 1 point at level 100),
// level, and nature (+10% to one stat, -10% to another).
func CalcStats(base, ivs, evs data.Stats, level int, nature data.Nature) data.Stats {
	var out data.Stats
	for _, s := range data.MainStats {
		core := (2*base.Get(s) + ivs.Get(s) + evs.Get(s)/4) * level / 100
		if s == data.StatHP {
			out.Set(s, core+level+10)
			continue
		}
		v := core + 5
		if !nature.Neutral() {
			switch s {
			case nature.Up:
				v = v * natureBoost / 100
			case nature.Down:
				v = v * natureLowered / 100
			}
		}
		out.Set(s, v)
	}
	return out
}

// AddEVs adds gained EVs, respecting the per-stat and total caps.
func AddEVs(evs, gain data.Stats) data.Stats {
	for _, s := range data.MainStats {
		room := min(MaxEVPerStat-evs.Get(s), MaxEVTotal-evs.Total())
		add := max(0, min(gain.Get(s), room))
		evs.Set(s, evs.Get(s)+add)
	}
	return evs
}

// IVPercent is the overall quality of the hidden rolls, 0 to 100.
func IVPercent(ivs data.Stats) float64 {
	return float64(ivs.Total()) / float64(6*MaxIV) * 100
}
