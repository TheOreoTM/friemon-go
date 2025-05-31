package logger

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Common field helpers for Discord bot logging

// Discord related fields
func DiscordUserID(id snowflake.ID) zap.Field {
	return zap.String("discord_user_id", id.String())
}

func DiscordChannelID(id snowflake.ID) zap.Field {
	return zap.String("discord_channel_id", id.String())
}

func DiscordGuildID(id snowflake.ID) zap.Field {
	return zap.String("discord_guild_id", id.String())
}

func DiscordMessageID(id snowflake.ID) zap.Field {
	return zap.String("discord_message_id", id.String())
}

// Character related fields
func CharacterID(id uuid.UUID) zap.Field {
	return zap.String("character_id", id.String())
}

func CharacterName(name string) zap.Field {
	return zap.String("character_name", name)
}

func CharacterLevel(level int) zap.Field {
	return zap.Int("character_level", level)
}

func CharacterOwner(ownerID string) zap.Field {
	return zap.String("character_owner", ownerID)
}

// Bot related fields
func Component(name string) zap.Field {
	return zap.String("component", name)
}

func Command(name string) zap.Field {
	return zap.String("command", name)
}

func Event(name string) zap.Field {
	return zap.String("event", name)
}

func Handler(name string) zap.Field {
	return zap.String("handler", name)
}

// General fields
func Duration(d time.Duration) zap.Field {
	return zap.Duration("duration", d)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func UserID(id string) zap.Field {
	return zap.String("user_id", id)
}

func RequestID(id string) zap.Field {
	return zap.String("request_id", id)
}

func Operation(op string) zap.Field {
	return zap.String("operation", op)
}

// Database related fields
func DatabaseQuery(query string) zap.Field {
	return zap.String("db_query", query)
}

func DatabaseTable(table string) zap.Field {
	return zap.String("db_table", table)
}

func RowsAffected(count int64) zap.Field {
	return zap.Int64("rows_affected", count)
}

// Cache related fields
func CacheKey(key string) zap.Field {
	return zap.String("cache_key", key)
}

func CacheHit(hit bool) zap.Field {
	return zap.Bool("cache_hit", hit)
}

func TTL(ttl time.Duration) zap.Field {
	return zap.Duration("ttl", ttl)
}
