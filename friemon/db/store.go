package db

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/model"
)

// Store is the database
type Store interface {
	GetCharactersForUser(context.Context, snowflake.ID) (*[]model.Character, error)
	GetCharacter(context.Context, uuid.UUID) (*model.Character, error)
	CreateCharacter(context.Context, snowflake.ID) (*model.Character, error)
	UpdateCharacter(context.Context, uuid.UUID, model.Character) (*model.Character, error)
	DeleteCharacter(context.Context, uuid.UUID) (*model.Character, error)
	Tx(context.Context, func(Store) error) error
}
