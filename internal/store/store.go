// Package store reads and writes per-player data in Postgres.
package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a pool of database connections. A pool is safe to use
// from many goroutines at once and reconnects by itself.
type Store struct {
	db *pgxpool.Pool
}

// Open connects to the database and applies any pending migrations.
func Open(ctx context.Context, url string) (*Store, error) {
	db, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close releases all connections.
func (s *Store) Close() {
	s.db.Close()
}
