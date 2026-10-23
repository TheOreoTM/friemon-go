package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
)

// OwnedCharacter is one caught copy as stored in the database.
type OwnedCharacter struct {
	Idx         int // the owner's #1, #2, ...
	CharacterID string
	Level       int
	XP          int    // towards the next level
	Nature      string // a nature ID
	IVs         data.Stats
	EVs         data.Stats
	Moves       []string
	Favorite    bool
	CaughtAt    time.Time
}

// characterColumns and scanCharacter go together: every query that returns
// whole characters selects these columns, in this order.
const characterColumns = `c.idx, c.character_id, c.level, c.xp, c.nature,
	c.iv_hp, c.iv_attack, c.iv_defense, c.iv_magic_attack, c.iv_magic_defense, c.iv_speed,
	c.ev_hp, c.ev_attack, c.ev_defense, c.ev_magic_attack, c.ev_magic_defense, c.ev_speed,
	c.moves, c.favorite, c.caught_at`

// pgx.Row is an interface satisfied by both a single QueryRow result and each row
// of a Query, so this one function scans both.
func scanCharacter(row pgx.Row) (OwnedCharacter, error) {
	var c OwnedCharacter
	iv, ev := &c.IVs, &c.EVs
	err := row.Scan(&c.Idx, &c.CharacterID, &c.Level, &c.XP, &c.Nature,
		&iv.HP, &iv.Attack, &iv.Defense, &iv.MagicAttack, &iv.MagicDefense, &iv.Speed,
		&ev.HP, &ev.Attack, &ev.Defense, &ev.MagicAttack, &ev.MagicDefense, &ev.Speed,
		&c.Moves, &c.Favorite, &c.CaughtAt)
	return c, err
}

// AddCaught saves a newly caught character for a user and returns its number (#1, #2, ...).
// The user's first character becomes their party leader.
func (s *Store) AddCaught(ctx context.Context, userID snowflake.ID, c game.Caught) (int, error) {
	// pgx.BeginFunc commits if the function returns nil and rolls back if it returns an error.
	var idx int
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		// Create the user on their first catch, and take the next number atomically,
		// so two catches at the same moment can never get the same number.
		err := tx.QueryRow(ctx, `
			INSERT INTO users (id, next_idx, party) VALUES ($1, 1, '{1}')
			ON CONFLICT (id) DO UPDATE SET next_idx = users.next_idx + 1
			RETURNING next_idx`,
			userID).Scan(&idx)
		if err != nil {
			return fmt.Errorf("next number: %w", err)
		}

		moves := c.Moves
		if moves == nil {
			moves = []string{} // a nil slice would be saved as NULL, which the column does not allow
		}
		iv := c.IVs
		_, err = tx.Exec(ctx, `
			INSERT INTO characters (owner_id, idx, character_id, level, nature, moves,
				iv_hp, iv_attack, iv_defense, iv_magic_attack, iv_magic_defense, iv_speed)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			userID, idx, c.CharacterID, c.Level, c.Nature, moves,
			iv.HP, iv.Attack, iv.Defense, iv.MagicAttack, iv.MagicDefense, iv.Speed)
		if err != nil {
			return fmt.Errorf("insert character: %w", err)
		}
		return nil
	})
	return idx, err
}

// ListCharacters returns one page of a user's characters, ordered by number.
func (s *Store) ListCharacters(ctx context.Context, userID snowflake.ID, limit, offset int) ([]OwnedCharacter, error) {
	rows, err := s.db.Query(ctx, `
		SELECT `+characterColumns+` FROM characters c
		WHERE c.owner_id = $1
		ORDER BY c.idx LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return collectCharacters(rows)
}

// collectCharacters scans every row into a slice. CollectRows calls the function once per row.
func collectCharacters(rows pgx.Rows) ([]OwnedCharacter, error) {
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (OwnedCharacter, error) {
		return scanCharacter(row)
	})
}

// CountCharacters returns how many characters a user owns.
func (s *Store) CountCharacters(ctx context.Context, userID snowflake.ID) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, `SELECT count(*) FROM characters WHERE owner_id = $1`, userID).Scan(&n)
	return n, err
}

// GetCharacter returns a user's character by number. ok is false if they have no such number.
func (s *Store) GetCharacter(ctx context.Context, userID snowflake.ID, idx int) (c OwnedCharacter, ok bool, err error) {
	row := s.db.QueryRow(ctx, `
		SELECT `+characterColumns+` FROM characters c
		WHERE c.owner_id = $1 AND c.idx = $2`,
		userID, idx)
	return found(scanCharacter(row))
}

// SelectedCharacter returns the user's party leader. ok is false if they have none.
func (s *Store) SelectedCharacter(ctx context.Context, userID snowflake.ID) (c OwnedCharacter, ok bool, err error) {
	row := s.db.QueryRow(ctx, `
		SELECT `+characterColumns+` FROM characters c
		JOIN users u ON u.id = c.owner_id AND c.idx = u.party[1]
		WHERE u.id = $1`,
		userID)
	return found(scanCharacter(row))
}

// found turns pgx's "no rows" error into ok=false, so callers do not have to check for it.
func found(c OwnedCharacter, err error) (OwnedCharacter, bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return OwnedCharacter{}, false, nil
	}
	if err != nil {
		return OwnedCharacter{}, false, err
	}
	return c, true, nil
}

// SetMoves replaces the moves a user's character knows.
func (s *Store) SetMoves(ctx context.Context, userID snowflake.ID, idx int, moves []string) error {
	_, err := s.db.Exec(ctx, `UPDATE characters SET moves = $3 WHERE owner_id = $1 AND idx = $2`,
		userID, idx, moves)
	return err
}

// SetEVs replaces a character's effort values.
func (s *Store) SetEVs(ctx context.Context, userID snowflake.ID, idx int, ev data.Stats) error {
	_, err := s.db.Exec(ctx, `
		UPDATE characters SET ev_hp = $3, ev_attack = $4, ev_defense = $5,
			ev_magic_attack = $6, ev_magic_defense = $7, ev_speed = $8
		WHERE owner_id = $1 AND idx = $2`,
		userID, idx, ev.HP, ev.Attack, ev.Defense, ev.MagicAttack, ev.MagicDefense, ev.Speed)
	return err
}
