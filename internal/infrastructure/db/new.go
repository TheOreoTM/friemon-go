package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	SSLMode  string `toml:"ssl_mode"`
}

func (c Config) String() string {
	return fmt.Sprintf("Host: %s, Port: %d, Database: %s, Username: %s, SSLMode: %s",
		c.Host, c.Port, c.Database, c.Username, c.SSLMode)
}

func (c Config) PostgresDataSourceName() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("host=%s", c.Host))
	parts = append(parts, fmt.Sprintf("port=%d", c.Port))
	parts = append(parts, fmt.Sprintf("user=%s", c.Username))
	if c.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", c.Password))
	}
	parts = append(parts, fmt.Sprintf("dbname=%s", c.Database))
	if c.SSLMode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", c.SSLMode))
	}
	parts = append(parts, "TimeZone=UTC")
	return strings.Join(parts, " ")
}

type DB struct {
	*gorm.DB
}

func NewDB(cfg Config) (*DB, error) {
	gormLogger := logger.Default.LogMode(logger.Info)

	db, err := gorm.Open(postgres.Open(cfg.PostgresDataSourceName()), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) AutoMigrate() error {

	err := db.DB.AutoMigrate(&User{}, &Character{})
	if err != nil {
		return err
	}

	return nil
}
