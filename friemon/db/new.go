package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
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
	return fmt.Sprintf("\n   Host: %s\n   Port: %d\n   Username: %s\n   Password: %s\n   Database: %s\n   SSLMode: %s",
		c.Host,
		c.Port,
		c.Username,
		strings.Repeat("*", len(c.Password)),
		c.Database,
		c.SSLMode,
	)
}

func (c Config) PostgresDataSourceName() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.Username,
		c.Password,
		c.Database,
		c.SSLMode,
	)
}

func NewDB(cfg Config) (*Queries, *pgx.Conn, error) {
	// Parse the PostgreSQL connection configuration
	pgCfg, err := pgx.ParseConfig(cfg.PostgresDataSourceName())
	if err != nil {
		return nil, nil, err
	}

	// Establish a connection to the PostgreSQL database
	conn, err := pgx.ConnectConfig(context.Background(), pgCfg)
	if err != nil {
		return nil, conn, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set a timeout for pinging the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ping the database to ensure the connection is established
	if err = conn.Ping(ctx); err != nil {
		conn.Close(context.Background()) // Close connection on error
		return nil, conn, fmt.Errorf("failed to ping database: %w", err)
	}

	// Return a new instance of Queries with the established connection
	return New(conn), conn, nil
}
