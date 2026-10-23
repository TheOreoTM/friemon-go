package store

import (
	"context"
	"errors"

	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5"
)

// SetSpawnChannel sets the channel where characters spawn in a server.
func (s *Store) SetSpawnChannel(ctx context.Context, guildID, channelID snowflake.ID) error {
	// ON CONFLICT turns the insert into an update if the guild already has a row (an "upsert").
	_, err := s.db.Exec(ctx, `
		INSERT INTO guilds (id, spawn_channel_id) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET spawn_channel_id = excluded.spawn_channel_id`,
		guildID, channelID)
	return err
}

// SpawnChannel returns a server's spawn channel. ok is false if none is set.
func (s *Store) SpawnChannel(ctx context.Context, guildID snowflake.ID) (channelID snowflake.ID, ok bool, err error) {
	// A pointer can be nil, which is how pgx reports a NULL column.
	var id *snowflake.ID
	err = s.db.QueryRow(ctx, `SELECT spawn_channel_id FROM guilds WHERE id = $1`, guildID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil || id == nil {
		return 0, false, err
	}
	return *id, true, nil
}
