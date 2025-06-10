package db

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/internal/core/game"
)

type Store interface {
	// Character operations
	GetCharactersForUser(context.Context, snowflake.ID) ([]game.Character, error)
	GetCharacter(context.Context, uuid.UUID) (*game.Character, error)
	CreateCharacter(context.Context, snowflake.ID, *game.Character) error
	UpdateCharacter(context.Context, uuid.UUID, *game.Character) (*game.Character, error)
	DeleteCharacter(context.Context, uuid.UUID) (*game.Character, error)

	// User operations
	GetUser(context.Context, snowflake.ID) (*game.User, error)
	UpdateUser(context.Context, game.User) (*game.User, error)
	CreateUser(context.Context, snowflake.ID) (*game.User, error)
	GetSelectedCharacter(context.Context, snowflake.ID) (*game.Character, error)

	// Utility operations
	DeleteEverything(context.Context) error
	Tx(context.Context, func(Store) error) error
}
