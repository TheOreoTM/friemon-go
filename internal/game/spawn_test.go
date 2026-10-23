package game

import (
	"testing"
	"time"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

var testCfg = SpawnConfig{Threshold: 3, UserCooldown: 5 * time.Second, GuildCooldown: time.Minute}

func TestSpawnAtThreshold(t *testing.T) {
	c := NewSpawnCounter(testCfg)
	start := time.Now().Add(-time.Hour) // well past any guild cooldown

	// Three different users, one message each: the third should spawn.
	if c.Record("g1", "a", start) || c.Record("g1", "b", start) {
		t.Fatal("spawned before the threshold")
	}
	if !c.Record("g1", "c", start) {
		t.Fatal("did not spawn at the threshold")
	}
}

func TestUserCooldownIgnoresSpam(t *testing.T) {
	c := NewSpawnCounter(testCfg)
	now := time.Now()

	// One user sending many messages within their cooldown only counts once.
	for range 10 {
		if c.Record("g1", "spammer", now) {
			t.Fatal("a single spamming user caused a spawn")
		}
	}
	// After the cooldown, their messages count again.
	c.Record("g1", "spammer", now.Add(6*time.Second))
	if !c.Record("g1", "spammer", now.Add(12*time.Second)) {
		t.Fatal("messages after the cooldown did not count")
	}
}

func TestGuildCooldownDelaysSpawn(t *testing.T) {
	c := NewSpawnCounter(testCfg)
	now := time.Now()
	for _, u := range []string{"a", "b", "c"} {
		c.Record("g1", u, now)
	}

	// The threshold is reached again 10 seconds later, but the server is on cooldown.
	later := now.Add(10 * time.Second)
	for _, u := range []string{"a", "b", "c"} {
		if c.Record("g1", u, later) {
			t.Fatal("spawned during the server cooldown")
		}
	}
	// The next counted message after the cooldown spawns.
	if !c.Record("g1", "d", now.Add(2*time.Minute)) {
		t.Fatal("did not spawn after the server cooldown")
	}
}

func TestServersCountSeparately(t *testing.T) {
	c := NewSpawnCounter(testCfg)
	now := time.Now()
	c.Record("g1", "a", now)
	c.Record("g1", "b", now)
	if c.Record("g2", "c", now) {
		t.Fatal("a message in g2 used g1's count")
	}
}

func TestPickSpawnFollowsRarity(t *testing.T) {
	r := newRand()
	counts := map[data.Rarity]int{}
	const n = 100_000 // underscores in numbers are just for readability
	for range n {
		counts[PickSpawn(r, data.Characters).Rarity]++
	}

	for rarity, weight := range rarityWeights {
		got := float64(counts[rarity]) / n * 100
		if diff := got - float64(weight); diff < -1 || diff > 1 {
			t.Errorf("%s spawned %.1f%% of the time, want about %d%%", rarity, got, weight)
		}
	}
}

func TestPickSpawnSkipsEmptyTiers(t *testing.T) {
	onlyCommon := []data.Character{{ID: "x", Name: "X", Rarity: data.Common}}
	r := newRand()
	for range 1000 {
		if got := PickSpawn(r, onlyCommon); got.ID != "x" {
			t.Fatalf("picked %s from a one-character roster", got.ID)
		}
	}
}

func TestRollCaught(t *testing.T) {
	r := newRand()
	frieren, _ := data.CharacterByID("frieren")
	for range 1000 {
		c := RollCaught(r, frieren)
		if c.CharacterID != "frieren" || c.Level < 1 || c.Level > 25 {
			t.Fatalf("bad roll %+v", c)
		}
		ivs := c.IVs
		for _, iv := range []int{ivs.HP, ivs.Attack, ivs.Defense, ivs.MagicAttack, ivs.MagicDefense, ivs.Speed} {
			if iv < 0 || iv > MaxIV {
				t.Fatalf("IV %d out of range in %+v", iv, c)
			}
		}
	}

	perfect := data.Stats{HP: MaxIV, Attack: MaxIV, Defense: MaxIV, MagicAttack: MaxIV, MagicDefense: MaxIV, Speed: MaxIV}
	if p := IVPercent(perfect); p != 100 {
		t.Errorf("perfect IVPercent = %v, want 100", p)
	}
}
