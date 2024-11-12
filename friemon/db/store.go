package db

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/entities"
)

// Store is the database
type Store interface {
	GetCharactersForUser(context.Context, snowflake.ID) ([]entities.Character, error)
	GetCharacter(context.Context, uuid.UUID) (*entities.Character, error)
	CreateCharacter(context.Context, snowflake.ID) (*entities.Character, error)
	UpdateCharacter(context.Context, uuid.UUID, entities.Character) (*entities.Character, error)
	DeleteCharacter(context.Context, uuid.UUID) (*entities.Character, error)

	GetUser(context.Context, snowflake.ID) (*entities.User, error)
	UpdateUser(context.Context, entities.User) (*entities.User, error)
	CreateUser(context.Context, snowflake.ID) (*entities.User, error)
	GetSelectedCharacter(context.Context, snowflake.ID) (*entities.Character, error)

	DeleteEverything(context.Context) error

	Tx(context.Context, func(Store) error) error
}
