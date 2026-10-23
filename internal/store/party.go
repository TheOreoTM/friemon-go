package store

import (
	"context"
	"errors"

	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5"

	"github.com/TheOreoTM/friemon-go/internal/game"
)

// ErrNotOwned means a party change named a character the user does not have.
var ErrNotOwned = errors.New("you do not own one of those characters")

// PartyIdxs returns the numbers of the user's party members, leader first.
func (s *Store) PartyIdxs(ctx context.Context, userID snowflake.ID) ([]int, error) {
	var party []int
	err := s.db.QueryRow(ctx, `SELECT party FROM users WHERE id = $1`, userID).Scan(&party)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return party, err
}

// Party returns the user's party members, leader first.
func (s *Store) Party(ctx context.Context, userID snowflake.ID) ([]OwnedCharacter, error) {
	// unnest turns the party array into rows; WITH ORDINALITY numbers them so we keep the order.
	rows, err := s.db.Query(ctx, `
		SELECT `+characterColumns+`
		FROM users u
		CROSS JOIN LATERAL unnest(u.party) WITH ORDINALITY AS p(idx, slot)
		JOIN characters c ON c.owner_id = u.id AND c.idx = p.idx
		WHERE u.id = $1
		ORDER BY p.slot`,
		userID)
	if err != nil {
		return nil, err
	}
	return collectCharacters(rows)
}

// SetParty saves a new party. Every number must be a character the user owns.
func (s *Store) SetParty(ctx context.Context, userID snowflake.ID, party []int) error {
	return pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		var owned int
		err := tx.QueryRow(ctx, `SELECT count(*) FROM characters WHERE owner_id = $1 AND idx = ANY($2)`,
			userID, party).Scan(&owned)
		if err != nil {
			return err
		}
		if owned != len(party) {
			return ErrNotOwned
		}
		_, err = tx.Exec(ctx, `UPDATE users SET party = $2 WHERE id = $1`, userID, party)
		return err
	})
}

// LevelUp describes a character that just gained one or more levels.
type LevelUp struct {
	Idx         int
	CharacterID string
	OldLevel    int
	NewLevel    int
}

// AddPartyXP gives chat XP to the user's party: the leader gets leaderXP and the
// others get game.PartyXPShare percent of it. It returns every member that levelled up.
func (s *Store) AddPartyXP(ctx context.Context, userID snowflake.ID, leaderXP int) ([]LevelUp, error) {
	var ups []LevelUp
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		// FOR UPDATE locks the rows until the transaction ends, so two messages
		// at once cannot both read the old XP and lose one of the gains.
		rows, err := tx.Query(ctx, `
			SELECT c.idx, c.character_id, c.level, c.xp, p.slot
			FROM users u
			CROSS JOIN LATERAL unnest(u.party) WITH ORDINALITY AS p(idx, slot)
			JOIN characters c ON c.owner_id = u.id AND c.idx = p.idx
			WHERE u.id = $1
			FOR UPDATE OF c`,
			userID)
		if err != nil {
			return err
		}
		type member struct {
			idx, level, xp int
			slot           int64
			characterID    string
		}
		members, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (member, error) {
			var m member
			err := row.Scan(&m.idx, &m.characterID, &m.level, &m.xp, &m.slot)
			return m, err
		})
		if err != nil {
			return err
		}

		for _, m := range members {
			if m.level >= game.MaxLevel {
				continue
			}
			gain := leaderXP
			if m.slot != 1 {
				gain = max(1, leaderXP*game.PartyXPShare/100)
			}
			level, xp := game.AddXP(m.level, m.xp, gain)
			_, err := tx.Exec(ctx, `UPDATE characters SET level = $3, xp = $4 WHERE owner_id = $1 AND idx = $2`,
				userID, m.idx, level, xp)
			if err != nil {
				return err
			}
			if level > m.level {
				ups = append(ups, LevelUp{Idx: m.idx, CharacterID: m.characterID, OldLevel: m.level, NewLevel: level})
			}
		}
		return nil
	})
	return ups, err
}

// UserIDs returns every player who has caught something, for developer tools.
func (s *Store) UserIDs(ctx context.Context) ([]snowflake.ID, error) {
	rows, err := s.db.Query(ctx, `SELECT id FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[snowflake.ID])
}
