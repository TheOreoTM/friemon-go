package game

import (
	"sync"
	"time"
)

// Cooldown lets each key (such as a user ID) through at most once per period.
type Cooldown struct {
	period time.Duration
	mu     sync.Mutex
	last   map[string]time.Time
}

// NewCooldown creates a cooldown with the given period.
func NewCooldown(period time.Duration) *Cooldown {
	return &Cooldown{period: period, last: map[string]time.Time{}}
}

// Allow reports whether key may act now, and if so starts its cooldown.
func (c *Cooldown) Allow(key string, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if t, ok := c.last[key]; ok && now.Sub(t) < c.period {
		return false
	}
	c.last[key] = now

	// Now and then, forget keys whose cooldown is over so the map does not grow forever.
	if len(c.last) > 10_000 {
		for k, t := range c.last {
			if now.Sub(t) >= c.period {
				delete(c.last, k)
			}
		}
	}
	return true
}
