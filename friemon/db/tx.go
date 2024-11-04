package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Tx executes a function in a transaction.
func (q *Queries) Tx(ctx context.Context, fn func(s Store) error) error {
	return q.asTx(ctx, func(txQ *Queries) error {
		return fn(txQ)
	})
}

// asTx wraps a function in a database transaction.
func (q *Queries) asTx(ctx context.Context, fn func(txQ *Queries) error) error {
	conn, ok := q.db.(*pgx.Conn)
	if !ok {
		return errors.New("invalid database connection")
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	txQ := q.WithTx(tx)

	if err := fn(txQ); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return errors.New("transaction rollback failed: " + rbErr.Error())
		}
		return err
	}

	return tx.Commit(ctx)
}
