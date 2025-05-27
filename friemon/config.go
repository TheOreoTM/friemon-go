package friemon

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/disgoorg/snowflake/v2"
	"github.com/pelletier/go-toml/v2"
	"github.com/theoreotm/friemon/friemon/db"
)

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config: %w", err)
	}

	var cfg Config
	if err = toml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	// Override with environment variables if set
	cfg.overrideWithEnv()

	return &cfg, nil
}

func (c *Config) overrideWithEnv() {
	// Log config
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		var l slog.Level
		switch level {
		case "debug":
			l = slog.LevelDebug
		case "info":
			l = slog.LevelInfo
		case "warn":
			l = slog.LevelWarn
		case "error":
			l = slog.LevelError
		}
		c.Log.Level = l
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		c.Log.Format = format
	}
	if addSource := os.Getenv("LOG_ADD_SOURCE"); addSource != "" {
		c.Log.AddSource, _ = strconv.ParseBool(addSource)
	}

	if timezone := os.Getenv("TZ"); timezone != "" {
		c.Timezone = timezone
	}
	if assetsDir := os.Getenv("ASSETS_DIR"); assetsDir != "" {
		c.AssetsDir = assetsDir
	}

	// Bot config
	if token := os.Getenv("BOT_TOKEN"); token != "" {
		c.Bot.Token = token
	}
	if syncCommands := os.Getenv("BOT_SYNC_COMMANDS"); syncCommands != "" {
		c.Bot.SyncCommands, _ = strconv.ParseBool(syncCommands)
	}
	if devMode := os.Getenv("BOT_DEV_MODE"); devMode != "" {
		c.Bot.DevMode, _ = strconv.ParseBool(devMode)
	}
	if version := os.Getenv("BOT_VERSION"); version != "" {
		c.Bot.Version = version
	}
	if devGuilds := os.Getenv("BOT_DEV_GUILDS"); devGuilds != "" {
		c.Bot.DevGuilds = stringsToSnowflakeIDs(devGuilds)
	}

	// Database config
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		c.Database.Port, _ = strconv.Atoi(port)
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
		c.Redis.DB, _ = strconv.Atoi(db)
	}
}

type Config struct {
	Timezone  string      `toml:"timezone"`
	AssetsDir string      `toml:"assets_dir"`
	Log       LogConfig   `toml:"log"`
	Bot       BotConfig   `toml:"bot"`
	Database  db.Config   `toml:"database"`
	Redis     RedisConfig `toml:"redis"`
}

type BotConfig struct {
	DevGuilds    []snowflake.ID `toml:"dev_guilds"`
	Token        string         `toml:"token"`
	SyncCommands bool           `toml:"sync_commands"`
	DevMode      bool           `toml:"dev_mode"`
	Version      string         `toml:"version"`
}

type LogConfig struct {
	Level     slog.Level `toml:"level"`
	Format    string     `toml:"format"`
	AddSource bool       `toml:"add_source"`
}

type RedisConfig struct {
	Addr     string `toml:"addr"`
	Password string `toml:"password"`
	DB       int    `toml:"db"`
}

func stringsToSnowflakeIDs(ids string) []snowflake.ID {
	ids = strings.TrimSpace(ids)
	if ids == "" {
		return nil
	}
	var snowflakeIDs []snowflake.ID
	for _, id := range strings.Split(ids, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		snowflakeID, err := snowflake.Parse(id)
		if err != nil {
			slog.Error("Failed to parse snowflake ID", slog.String("id", id), slog.Any("err", err))
			continue
		}
		snowflakeIDs = append(snowflakeIDs, snowflakeID)
	}
	return snowflakeIDs
}
