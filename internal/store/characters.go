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
	XP          int // towards the next level
	IVs         data.Stats
	Moves       []string
	Favorite    bool
	CaughtAt    time.Time
}

// characterColumns and scanCharacter go together: every query that returns
// whole characters selects these columns, in this order.
const characterColumns = `c.idx, c.character_id, c.level, c.xp,
	c.iv_hp, c.iv_attack, c.iv_defense, c.iv_magic_attack, c.iv_magic_defense, c.iv_speed,
	c.moves, c.favorite, c.caught_at`

// pgx.Row is an interface satisfied by both a single QueryRow result and each row
// of a Query, so this one function scans both.
func scanCharacter(row pgx.Row) (OwnedCharacter, error) {
	var c OwnedCharacter
	err := row.Scan(&c.Idx, &c.CharacterID, &c.Level, &c.XP,
		&c.IVs.HP, &c.IVs.Attack, &c.IVs.Defense, &c.IVs.MagicAttack, &c.IVs.MagicDefense, &c.IVs.Speed,
		&c.Moves, &c.Favorite, &c.CaughtAt)
	return c, err
}

// AddCaught saves a newly caught character for a user and returns its number (#1, #2, ...).
// The user's first character becomes their selected one.
func (s *Store) AddCaught(ctx context.Context, userID snowflake.ID, c game.Caught) (int, error) {
	// pgx.BeginFunc commits if the function returns nil and rolls back if it returns an error.
	var idx int
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		// Create the user on their first catch, and take the next number atomically,
		// so two catches at the same moment can never get the same number.
		err := tx.QueryRow(ctx, `
			INSERT INTO users (id, next_idx, selected_idx) VALUES ($1, 1, 1)
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
			INSERT INTO characters (owner_id, idx, character_id, level, moves,
				iv_hp, iv_attack, iv_defense, iv_magic_attack, iv_magic_defense, iv_speed)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			userID, idx, c.CharacterID, c.Level, moves,
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
	// CollectRows calls the function once per row and gathers the results into a slice.
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

// SelectedCharacter returns the user's selected character. ok is false if they have none.
func (s *Store) SelectedCharacter(ctx context.Context, userID snowflake.ID) (c OwnedCharacter, ok bool, err error) {
	row := s.db.QueryRow(ctx, `
		SELECT `+characterColumns+` FROM characters c
		JOIN users u ON u.id = c.owner_id AND u.selected_idx = c.idx
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

// SelectCharacter makes one of the user's characters their selected one.
// ok is false if they have no character with that number.
func (s *Store) SelectCharacter(ctx context.Context, userID snowflake.ID, idx int) (ok bool, err error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE users SET selected_idx = $2
		WHERE id = $1 AND EXISTS (SELECT 1 FROM characters WHERE owner_id = $1 AND idx = $2)`,
		userID, idx)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// LevelUp describes a character that just gained one or more levels.
type LevelUp struct {
	Idx         int
	CharacterID string
	OldLevel    int
	NewLevel    int
}

// AddXPToSelected gives XP to the user's selected character.
// It returns a LevelUp if the character levelled up, or nil if not (or if there is no selected character).
func (s *Store) AddXPToSelected(ctx context.Context, userID snowflake.ID, gain int) (*LevelUp, error) {
	var up *LevelUp
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		var idx, level, xp int
		var characterID string
		// FOR UPDATE locks the row until the transaction ends, so two messages
		// at once cannot both read the old XP and lose one of the gains.
		err := tx.QueryRow(ctx, `
			SELECT c.idx, c.character_id, c.level, c.xp FROM characters c
			JOIN users u ON u.id = c.owner_id AND u.selected_idx = c.idx
			WHERE u.id = $1
			FOR UPDATE OF c`,
			userID).Scan(&idx, &characterID, &level, &xp)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // no characters yet
		}
		if err != nil {
			return err
		}
		if level >= game.MaxLevel {
			return nil
		}

		newLevel, newXP := game.AddXP(level, xp, gain)
		_, err = tx.Exec(ctx, `UPDATE characters SET level = $3, xp = $4 WHERE owner_id = $1 AND idx = $2`,
			userID, idx, newLevel, newXP)
		if err != nil {
			return err
		}
		if newLevel > level {
			up = &LevelUp{Idx: idx, CharacterID: characterID, OldLevel: level, NewLevel: newLevel}
		}
		return nil
	})
	return up, err
}

// SetMoves replaces the moves a user's character knows.
func (s *Store) SetMoves(ctx context.Context, userID snowflake.ID, idx int, moves []string) error {
	_, err := s.db.Exec(ctx, `UPDATE characters SET moves = $3 WHERE owner_id = $1 AND idx = $2`,
		userID, idx, moves)
	return err
}
