package memstore

import (
	"errors"
	"sync"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/theoreotm/friemon/internal/core/game"
)

type memoryCache struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value     interface{}
	expiresAt time.Time
}

func NewMemoryCache() Cache {
	cache := &memoryCache{
		data: make(map[string]cacheItem),
	}
	go cache.cleanup()
	return cache
}

func (c *memoryCache) IncrementInteractionCount(channelId snowflake.ID) error {
	key := "channel:" + channelId.String() + ":interactions"
	return c.Set(key, c.GetInteractionCount(channelId)+1, 3*time.Minute)
}

func (c *memoryCache) GetInteractionCount(channelId snowflake.ID) int {
	key := "channel:" + channelId.String() + ":interactions"
	value, err := c.Get(key)
	if err != nil {
		return 0
	}

	return value.(int)
}

func (c *memoryCache) ResetInteractionCount(channelId snowflake.ID) error {
	key := "channel:" + channelId.String() + ":interactions"
	return c.Set(key, 0, 3*time.Minute)
}

func (c *memoryCache) SetChannelCharacter(channelID snowflake.ID, character *game.Character) error {
	key := "channel:" + channelID.String() + ":character"
	return c.Set(key, character, 3*time.Minute)
}

func (c *memoryCache) GetChannelCharacter(channelID snowflake.ID) (*game.Character, error) {
	key := "channel:" + channelID.String() + ":character"
	value, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	return value.(*game.Character), nil
}

func (c *memoryCache) DeleteChannelCharacter(channelID snowflake.ID) error {
	key := "channel:" + channelID.String() + ":character"
	return c.Delete(key)
}

func (c *memoryCache) Set(key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (c *memoryCache) Get(key string) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists || time.Now().After(item.expiresAt) {
		return nil, errors.New("key not found or expired")
	}
	return item.value, nil
}

func (c *memoryCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

func (c *memoryCache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute) // Cleanup interval
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for key, item := range c.data {
			if now.After(item.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}
