package entities

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
)

type OrderBy = int

const (
	OrderByIDX OrderBy = iota
	OrderByIV
	OrderByLevel
)

type OrderOptions struct {
	OrderBy OrderBy
	Desc    bool
}

type User struct {
	ID            snowflake.ID
	Balance       int
	SelectedID    uuid.UUID
	Order         OrderOptions
	NextIdx       int
	ShiniesCaught int
}
