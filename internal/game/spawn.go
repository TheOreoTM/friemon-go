package game

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/TheOreoTM/friemon-go/internal/data"
)

// SpawnConfig controls how often characters spawn.
type SpawnConfig struct {
	Threshold     int           // counted messages needed for a spawn
	UserCooldown  time.Duration // a user's messages count at most once per this long
	GuildCooldown time.Duration // minimum time between spawns in one server
}

// DefaultSpawnConfig is a starting point; tune it once real servers use the bot.
var DefaultSpawnConfig = SpawnConfig{
	Threshold:     3,
	UserCooldown:  3 * time.Second,
	GuildCooldown: 60 * time.Second,
}

// SpawnCounter counts messages per server and says when to spawn.
// It is in memory only; a restart resets the counts, which is fine.
//
// Discord events arrive on several goroutines (Go's lightweight threads) at once,
// so the mutex makes sure only one goroutine touches the maps at a time.
type SpawnCounter struct {
	cfg    SpawnConfig
	mu     sync.Mutex
	guilds map[string]*guildCount
}

type guildCount struct {
	count       int
	lastSpawn   time.Time
	lastCounted map[string]time.Time // userID -> last time their message counted
}

// NewSpawnCounter creates a counter with the given settings.
func NewSpawnCounter(cfg SpawnConfig) *SpawnCounter {
	return &SpawnCounter{cfg: cfg, guilds: map[string]*guildCount{}}
}

// Record counts one message and reports whether it triggers a spawn.
// It takes the current time as an argument so tests can control the clock.
func (s *SpawnCounter) Record(guildID, userID string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock() // runs when the function returns, on every path

	g, ok := s.guilds[guildID]
	if !ok {
		g = &guildCount{lastCounted: map[string]time.Time{}}
		s.guilds[guildID] = g
	}

	// Per-user cooldown: ignore a user who is spamming.
	if last, ok := g.lastCounted[userID]; ok && now.Sub(last) < s.cfg.UserCooldown {
		return false
	}
	g.lastCounted[userID] = now
	g.count++

	// Per-server cooldown: keep counting, but hold the spawn until it has passed.
	if g.count < s.cfg.Threshold || now.Sub(g.lastSpawn) < s.cfg.GuildCooldown {
		return false
	}
	g.count = 0
	g.lastSpawn = now
	clear(g.lastCounted) // stops the map growing forever with users who left
	return true
}

// rarityWeights is the chance of each tier, out of their total (100 here).
// A tier is picked first, then a character within it, so adding more commons
// does not make Frieren any rarer.
var rarityWeights = map[data.Rarity]int{
	data.Common:    80,
	data.Rare:      17,
	data.Legendary: 3,
}

// PickSpawn chooses a random character, weighted by rarity.
// Taking *rand.Rand as an argument lets tests pass a seeded, repeatable one.
func PickSpawn(r *rand.Rand, chars []data.Character) data.Character {
	byRarity := map[data.Rarity][]data.Character{}
	for _, c := range chars {
		byRarity[c.Rarity] = append(byRarity[c.Rarity], c)
	}

	// Only tiers that have characters can be picked.
	total := 0
	for rarity, w := range rarityWeights {
		if len(byRarity[rarity]) > 0 {
			total += w
		}
	}

	// Walk the tiers in a fixed order (map order is random in Go) until the roll lands.
	roll := r.IntN(total)
	for _, rarity := range []data.Rarity{data.Common, data.Rare, data.Legendary} {
		pool := byRarity[rarity]
		if len(pool) == 0 {
			continue
		}
		if roll < rarityWeights[rarity] {
			return pool[r.IntN(len(pool))]
		}
		roll -= rarityWeights[rarity]
	}
	panic("unreachable: rarity weights do not add up")
}
