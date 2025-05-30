package memstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/redis/go-redis/v9"
	"github.com/theoreotm/friemon/internal/core/entities"
)

// RedisCache implements the Cache interface using Redis.
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisCache creates a new RedisCache instance.
// It takes the Redis server address, password, and database number as arguments.
func NewRedisCache(addr string, password string, db int) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()
	// Test connection
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Helper function to generate a standardized key for channel interactions.
func channelInteractionKey(channelID snowflake.ID) string {
	return fmt.Sprintf("channel:%s:interactions", channelID.String())
}

// Helper function to generate a standardized key for channel characters.
func channelCharacterKey(channelID snowflake.ID) string {
	return fmt.Sprintf("channel:%s:character", channelID.String())
}

// Set stores a value in Redis with a given key and TTL (time-to-live).
// Complex types are marshalled to JSON. Integers are stored as strings.
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	var dataToStore []byte // Use []byte for SET command

	switch v := value.(type) {
	case *entities.Character:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal character to JSON: %w", err)
		}
		dataToStore = jsonData
	case int:
		dataToStore = []byte(strconv.Itoa(v))
	case string:
		dataToStore = []byte(v)
	default:
		// For other types, attempt to marshal to JSON
		jsonData, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value to JSON for key %s: %w", key, err)
		}
		dataToStore = jsonData
	}

	// Use SET with EX for TTL in seconds
	return c.client.Set(c.ctx, key, dataToStore, ttl).Err()
}

// Get retrieves a value from Redis by its key.
// It attempts to infer the type based on the key or by trying to unmarshal as JSON.
// This generic Get is tricky with Redis due to lack of stored type info for simple strings.
// Specific getters like GetChannelCharacter are preferred for typed retrieval.
func (c *RedisCache) Get(key string) (interface{}, error) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("key not found or expired")
		}
		return nil, fmt.Errorf("failed to get key %s from Redis: %w", key, err)
	}

	// Attempt to unmarshal as *entities.Character if it's a character key
	// This is a heuristic; a more robust system might involve type hints or specific methods.
	if _, err := snowflake.Parse(key[len("channel:"):strings.LastIndex(key, ":")]); err == nil && strings.HasSuffix(key, ":character") {
		var char entities.Character
		if jsonErr := json.Unmarshal([]byte(val), &char); jsonErr == nil {
			return &char, nil
		}
	}

	// Attempt to parse as int if it's an interaction key
	if _, err := snowflake.Parse(key[len("channel:"):strings.LastIndex(key, ":")]); err == nil && strings.HasSuffix(key, ":interactions") {
		num, intErr := strconv.Atoi(val)
		if intErr == nil {
			return num, nil
		}
	}

	// As a fallback, try to unmarshal as a generic JSON object
	var jsonData interface{}
	if json.Unmarshal([]byte(val), &jsonData) == nil {
		return jsonData, nil
	}

	// If all else fails, return as string
	return val, nil
}

// Delete removes a key from Redis.
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// IncrementInteractionCount increments the interaction counter for a channel.
// The counter expires after 3 minutes of inactivity.
func (c *RedisCache) IncrementInteractionCount(channelID snowflake.ID) error {
	key := channelInteractionKey(channelID)
	pipe := c.client.Pipeline()
	pipe.Incr(c.ctx, key)
	pipe.Expire(c.ctx, key, 3*time.Minute) // Set/update expiration
	_, err := pipe.Exec(c.ctx)
	return err
}

// GetInteractionCount retrieves the interaction count for a channel.
// Returns 0 if the key doesn't exist or an error occurs.
func (c *RedisCache) GetInteractionCount(channelID snowflake.ID) int {
	key := channelInteractionKey(channelID)
	val, err := c.client.Get(c.ctx, key).Int()
	if err != nil {
		// If key doesn't exist (redis.Nil) or other error, return 0
		return 0
	}
	return val
}

// ResetInteractionCount resets the interaction counter for a channel to 0.
// The counter expires after 3 minutes.
func (c *RedisCache) ResetInteractionCount(channelID snowflake.ID) error {
	key := channelInteractionKey(channelID)
	// Set to 0 with a 3-minute TTL
	return c.Set(key, 0, 3*time.Minute)
}

// SetChannelCharacter stores a character associated with a channel.
// The character data is marshalled to JSON and expires after 3 minutes.
func (c *RedisCache) SetChannelCharacter(channelID snowflake.ID, character *entities.Character) error {
	key := channelCharacterKey(channelID)
	return c.Set(key, character, 3*time.Minute)
}

// GetChannelCharacter retrieves a character associated with a channel.
// Returns nil if the key doesn't exist, is expired, or an error occurs during unmarshalling.
func (c *RedisCache) GetChannelCharacter(channelID snowflake.ID) (*entities.Character, error) {
	key := channelCharacterKey(channelID)
	val, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("character not found in cache or expired")
		}
		return nil, fmt.Errorf("failed to get character from Redis for channel %s: %w", channelID, err)
	}

	var character entities.Character
	if err := json.Unmarshal(val, &character); err != nil {
		return nil, fmt.Errorf("failed to unmarshal character JSON for channel %s: %w", channelID, err)
	}
	return &character, nil
}

// DeleteChannelCharacter removes a character associated with a channel from the cache.
func (c *RedisCache) DeleteChannelCharacter(channelID snowflake.ID) error {
	key := channelCharacterKey(channelID)
	return c.Delete(key)
}

// Close closes the Redis client connection.
func (c *RedisCache) Close() error {
	return c.client.Close()
}
