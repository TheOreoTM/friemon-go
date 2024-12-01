package memstore

import (
	"errors"
	"sync"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/theoreotm/friemon/entities"
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
	return &memoryCache{
		data: make(map[string]cacheItem),
	}
}

func (c *memoryCache) SetChannelCharacter(channelID snowflake.ID, character *entities.Character) error {
	key := "channel:" + channelID.String() + ":character"
	return c.Set(key, character, 60*60*24)
}

func (c *memoryCache) GetChannelCharacter(channelID snowflake.ID) (*entities.Character, error) {
	key := "channel:" + channelID.String() + ":character"
	value, err := c.Get(key)
	if err != nil {
		return nil, err
	}

	return value.(*entities.Character), nil
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
		expiresAt: time.Now().Add(ttl * time.Second),
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
