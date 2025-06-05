package bot

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/disgoorg/snowflake/v2"
	"github.com/theoreotm/friemon/internal/infrastructure/db"
	"github.com/theoreotm/friemon/internal/pkg/logger"
)

// LoadConfig loads configuration from environment variables only
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// Set defaults
		Timezone:  getEnvWithDefault("TZ", "UTC"),
		AssetsDir: getEnvWithDefault("ASSETS_DIR", "./assets"),
		Log: logger.Config{
			Level:      getEnvWithDefault("LOG_LEVEL", "info"),
			Format:     getEnvWithDefault("LOG_FORMAT", "console"),
			AddSource:  getEnvBool("LOG_ADD_SOURCE", true),
			OutputPath: getEnvWithDefault("LOG_OUTPUT_PATH", "stdout"),
		},
		Bot: BotConfig{
			Token:        getEnvRequired("BOT_TOKEN"),
			DevMode:      getEnvBool("DEV_MODE", false),
			SyncCommands: getEnvBool("SYNC_COMMANDS", true),
			Version:      getEnvWithDefault("BOT_VERSION", "1.0.0"),
			DevGuilds:    parseSnowflakes(os.Getenv("DEV_GUILDS")),
			AdminUsers:   parseSnowflakes(os.Getenv("ADMIN_USERS")),
		},
		Database: db.Config{
			Host:     getEnvWithDefault("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			Username: getEnvWithDefault("DB_USER", "friemon"),
			Password: getEnvRequired("DB_PASSWORD"),
			Database: getEnvWithDefault("DB_NAME", "friemon"),
			SSLMode:  getEnvWithDefault("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Addr:     getEnvWithDefault("REDIS_ADDR", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"), // Optional, can be empty
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}

	// Validate required fields
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Config represents the application configuration
type Config struct {
	Timezone  string
	AssetsDir string
	Log       logger.Config
	Bot       BotConfig
	Database  db.Config
	Redis     RedisConfig
}

// BotConfig holds Discord bot specific configuration
type BotConfig struct {
	DevGuilds    []snowflake.ID
	AdminUsers   []snowflake.ID
	Token        string
	SyncCommands bool
	DevMode      bool
	Version      string
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// validate checks if all required configuration is present and valid
func (c *Config) validate() error {
	if c.Bot.Token == "" {
		return fmt.Errorf("BOT_TOKEN is required")
	}

	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLevels, strings.ToLower(c.Log.Level)) {
		return fmt.Errorf("invalid LOG_LEVEL: %s (valid: %v)", c.Log.Level, validLevels)
	}

	// Validate log format
	validFormats := []string{"json", "console"}
	if !contains(validFormats, strings.ToLower(c.Log.Format)) {
		return fmt.Errorf("invalid LOG_FORMAT: %s (valid: %v)", c.Log.Format, validFormats)
	}

	return nil
}

// Helper functions for environment variable parsing

// getEnvRequired gets a required environment variable or panics
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("Required environment variable %s is not set", key))
	}
	return value
}

// getEnvWithDefault gets an environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvBool gets a boolean environment variable with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch strings.ToLower(value) {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// parseSnowflakes parses comma-separated snowflake IDs
func parseSnowflakes(snowflakeStr string) []snowflake.ID {
	if snowflakeStr == "" {
		return nil
	}

	var snowflakes []snowflake.ID
	for _, idStr := range strings.Split(snowflakeStr, ",") {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}

		if id, err := snowflake.Parse(idStr); err == nil {
			snowflakes = append(snowflakes, id)
		}
	}

	return snowflakes
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// String returns a string representation of the config (without sensitive data)
func (c *Config) String() string {
	return fmt.Sprintf("Config{Timezone: %s, Bot: {DevMode: %t, SyncCommands: %t}, Database: {Host: %s, Port: %d}, Redis: {Addr: %s}}",
		c.Timezone,
		c.Bot.DevMode,
		c.Bot.SyncCommands,
		c.Database.Host,
		c.Database.Port,
		c.Redis.Addr,
	)
}
