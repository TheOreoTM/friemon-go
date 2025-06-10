package memstore

import (
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/theoreotm/friemon/internal/core/game"
)

type Cache interface {
	Set(key string, value interface{}, ttl time.Duration) error
	Get(key string) (interface{}, error)
	Delete(key string) error

	IncrementInteractionCount(channelId snowflake.ID) error
	GetInteractionCount(channelId snowflake.ID) int
	ResetInteractionCount(channelId snowflake.ID) error

	SetChannelCharacter(channelID snowflake.ID, character *game.Character) error
	GetChannelCharacter(channelID snowflake.ID) (*game.Character, error)
	DeleteChannelCharacter(channelID snowflake.ID) error
}
