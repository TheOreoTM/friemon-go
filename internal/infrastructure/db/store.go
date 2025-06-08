package db

import (
	"context"

	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/theoreotm/friemon/internal/core/entities"
)

type Store interface {
	// Character methods
	GetCharactersForUser(context.Context, snowflake.ID) ([]entities.Character, error)
	GetCharacter(context.Context, uuid.UUID) (*entities.Character, error)
	CreateCharacter(context.Context, snowflake.ID, *entities.Character) error
	UpdateCharacter(context.Context, uuid.UUID, *entities.Character) (*entities.Character, error)
	DeleteCharacter(context.Context, uuid.UUID) (*entities.Character, error)

	// User methods
	GetUser(context.Context, snowflake.ID) (*entities.User, error)
	UpdateUser(context.Context, entities.User) (*entities.User, error)
	CreateUser(context.Context, snowflake.ID) (*entities.User, error)
	GetSelectedCharacter(context.Context, snowflake.ID) (*entities.Character, error)

	// Game Settings methods
	GetGameSettings(context.Context) (entities.GameSettings, error)
	GetGameSetting(context.Context, string) (string, error)
	UpdateGameSetting(context.Context, string, string) error
	CreateGameSetting(context.Context, string, string) error

	// Battle methods
	CreateBattle(context.Context, *entities.Battle) error
	GetBattle(context.Context, uuid.UUID) (*entities.Battle, error)
	UpdateBattle(context.Context, *entities.Battle) error
	GetActiveBattleForUser(context.Context, snowflake.ID) (*entities.Battle, error)
	GetUserBattleHistory(context.Context, snowflake.ID, int, int) ([]*entities.Battle, error)
	GetBattlesByStatus(context.Context, entities.BattleStatus) ([]*entities.Battle, error)

	// Battle Team methods
	CreateBattleTeamMember(context.Context, *entities.BattleTeamMember) error
	GetBattleTeam(context.Context, uuid.UUID, snowflake.ID) ([]*entities.BattleTeamMember, error)
	UpdateBattleTeamMember(context.Context, *entities.BattleTeamMember) error
	GetBattleTeamMember(context.Context, uuid.UUID) (*entities.BattleTeamMember, error)

	// Battle Turn methods
	CreateBattleTurn(context.Context, *entities.BattleTurn) error
	GetBattleTurns(context.Context, uuid.UUID) ([]*entities.BattleTurn, error)
	GetLastBattleTurn(context.Context, uuid.UUID) (*entities.BattleTurn, error)

	// ELO methods
	GetUserElo(context.Context, snowflake.ID) (*entities.UserElo, error)
	CreateUserElo(context.Context, *entities.UserElo) error
	UpdateUserElo(context.Context, *entities.UserElo) error
	GetEloLeaderboard(context.Context, int, int, int) ([]*entities.UserElo, error)
	GetUserEloRank(context.Context, snowflake.ID, int) (int, error)

	// Utility methods
	DeleteEverything(context.Context) error
	Tx(context.Context, func(Store) error) error
}
