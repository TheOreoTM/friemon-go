package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/disgoorg/snowflake/v2"
	"github.com/pelletier/go-toml/v2"
	"github.com/theoreotm/friemon/internal/infrastructure/db"
	"github.com/theoreotm/friemon/internal/pkg/logger"
)

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := toml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	cfg.overrideWithEnv()

	return &cfg, nil
}

func (c *Config) overrideWithEnv() {
	// Bot config
	if token := os.Getenv("BOT_TOKEN"); token != "" {
		c.Bot.Token = token
	}
	if devMode := os.Getenv("DEV_MODE"); devMode != "" {
		c.Bot.DevMode = strings.ToLower(devMode) == "true"
	}
	if syncCommands := os.Getenv("SYNC_COMMANDS"); syncCommands != "" {
		c.Bot.SyncCommands = strings.ToLower(syncCommands) == "true"
	}
	if devGuilds := os.Getenv("DEV_GUILDS"); devGuilds != "" {
		c.Bot.DevGuilds = stringsToSnowflakeIDs(devGuilds)
	}

	// Log config
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Log.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		c.Log.Format = format
	}
	if addSource := os.Getenv("LOG_ADD_SOURCE"); addSource != "" {
		c.Log.AddSource = strings.ToLower(addSource) == "true"
	}
	if outputPath := os.Getenv("LOG_OUTPUT_PATH"); outputPath != "" {
		c.Log.OutputPath = outputPath
	}

	// Database config
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.Database.Username = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		c.Database.Password = password
	}
	if dbname := os.Getenv("DB_NAME"); dbname != "" {
		c.Database.Database = dbname
	}
	if sslmode := os.Getenv("DB_SSL_MODE"); sslmode != "" {
		c.Database.SSLMode = sslmode
	}

	// Redis config
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		c.Redis.Addr = addr
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		c.Redis.Password = password
	}
	if db := os.Getenv("REDIS_DB"); db != "" {
		if d, err := strconv.Atoi(db); err == nil {
			c.Redis.DB = d
		}
	}
}

type Config struct {
	Timezone  string        `toml:"timezone"`
	AssetsDir string        `toml:"assets_dir"`
	Log       logger.Config `toml:"log"`
	Bot       BotConfig     `toml:"bot"`
	Database  db.Config     `toml:"database"`
	Redis     RedisConfig   `toml:"redis"`
}

type BotConfig struct {
	DevGuilds    []snowflake.ID `toml:"dev_guilds"`
	Token        string         `toml:"token"`
	SyncCommands bool           `toml:"sync_commands"`
	DevMode      bool           `toml:"dev_mode"`
	Version      string         `toml:"version"`
}

type RedisConfig struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

func stringsToSnowflakeIDs(ids string) []snowflake.ID {
	var snowflakeIDs []snowflake.ID
	for _, id := range strings.Split(ids, ",") {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			if snowflakeID, err := snowflake.Parse(trimmed); err == nil {
				snowflakeIDs = append(snowflakeIDs, snowflakeID)
			}
		}
	}
	return snowflakeIDs
}
