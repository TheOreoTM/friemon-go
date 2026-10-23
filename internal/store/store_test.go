package store

import (
	"context"
	"errors"
	"math/rand/v2"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/disgoorg/snowflake/v2"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
)

// openTestStore connects to TEST_DATABASE_URL, or skips the test if it is not set.
// Point it at a separate Neon branch, never the live database.
func openTestStore(t *testing.T) *Store {
	t.Helper() // makes failures point at the calling test, not this line
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	s, err := Open(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s
}

// beforeLaunch is a few days after Discord's ID epoch (2015-01-01) but months before
// Discord launched (May 2015), so no real account, server or channel has an ID from then.
var beforeLaunch = snowflake.New(time.Date(2015, 1, 2, 0, 0, 0, 0, time.UTC))

// testID returns a random ID that can never belong to a real user, and deletes its rows after the test.
func testID(t *testing.T, s *Store) snowflake.ID {
	id := beforeLaunch + snowflake.ID(rand.Uint64N(1<<40)) // spread over a few minutes of IDs
	t.Cleanup(func() {
		ctx := context.Background()
		s.db.Exec(ctx, `DELETE FROM characters WHERE owner_id = $1`, id)
		s.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
		s.db.Exec(ctx, `DELETE FROM guilds WHERE id = $1`, id)
	})
	return id
}

func TestSpawnChannel(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	guild := testID(t, s)

	if _, ok, err := s.SpawnChannel(ctx, guild); err != nil || ok {
		t.Fatalf("new guild: ok=%v err=%v, want no channel", ok, err)
	}
	for _, channel := range []snowflake.ID{beforeLaunch + 1, beforeLaunch + 2} { // the second call overwrites the first
		if err := s.SetSpawnChannel(ctx, guild, channel); err != nil {
			t.Fatal(err)
		}
		got, ok, err := s.SpawnChannel(ctx, guild)
		if err != nil || !ok || got != channel {
			t.Fatalf("SpawnChannel = %d, %v, %v; want %d", got, ok, err, channel)
		}
	}
}

func TestAddAndListCharacters(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	user := testID(t, s)

	first := game.Caught{CharacterID: "fern", Level: 5, IVs: data.Stats{
		HP: 1, Attack: 2, Defense: 3, MagicAttack: 4, MagicDefense: 5, Speed: 6,
	}}
	second := game.Caught{CharacterID: "frieren", Level: 20}
	for want, c := range []game.Caught{first, second} {
		idx, err := s.AddCaught(ctx, user, c)
		if err != nil {
			t.Fatal(err)
		}
		if idx != want+1 {
			t.Errorf("catch %d got number #%d", want+1, idx)
		}
	}

	if n, err := s.CountCharacters(ctx, user); err != nil || n != 2 {
		t.Fatalf("CountCharacters = %d, %v; want 2", n, err)
	}
	list, err := s.ListCharacters(ctx, user, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].CharacterID != "fern" || list[1].CharacterID != "frieren" {
		t.Fatalf("ListCharacters = %+v", list)
	}
	if list[0].IVs != first.IVs || list[1].Level != 20 {
		t.Errorf("stored values do not match: %+v", list)
	}

	// Paging: offset 1 skips the first character.
	page, err := s.ListCharacters(ctx, user, 10, 1)
	if err != nil || len(page) != 1 || page[0].Idx != 2 {
		t.Fatalf("second page = %+v, %v", page, err)
	}
}

func TestPartyXPAndMoves(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	user := testID(t, s)

	// No characters yet: no party, and XP goes nowhere.
	if _, ok, err := s.SelectedCharacter(ctx, user); err != nil || ok {
		t.Fatalf("SelectedCharacter with no characters: ok=%v err=%v", ok, err)
	}
	if ups, err := s.AddPartyXP(ctx, user, 100); err != nil || len(ups) != 0 {
		t.Fatalf("AddPartyXP with no characters = %+v, %v", ups, err)
	}

	for _, id := range []string{"fern", "stark", "eisen"} {
		if _, err := s.AddCaught(ctx, user, game.Caught{CharacterID: id, Level: 1, Nature: "jolly", Moves: []string{"slash"}}); err != nil {
			t.Fatal(err)
		}
	}

	// The first catch leads the party automatically.
	sel, ok, err := s.SelectedCharacter(ctx, user)
	if err != nil || !ok || sel.Idx != 1 || sel.Nature != "jolly" || !slices.Equal(sel.Moves, []string{"slash"}) {
		t.Fatalf("SelectedCharacter = %+v, %v, %v", sel, ok, err)
	}

	if err := s.SetParty(ctx, user, []int{2, 99}); !errors.Is(err, ErrNotOwned) {
		t.Fatalf("a party with a number the user does not own: %v", err)
	}
	if err := s.SetParty(ctx, user, []int{2, 3}); err != nil {
		t.Fatal(err)
	}
	party, err := s.Party(ctx, user)
	if err != nil || len(party) != 2 || party[0].Idx != 2 || party[1].Idx != 3 {
		t.Fatalf("Party = %+v, %v", party, err)
	}

	// The leader (#2) gets all 150 XP: level 2 needs 75, so it reaches level 2 with 75 left over.
	// #3 gets half, 75, exactly enough for level 2. #1 is not in the party and gets nothing.
	ups, err := s.AddPartyXP(ctx, user, 150)
	if err != nil || len(ups) != 2 {
		t.Fatalf("AddPartyXP = %+v, %v", ups, err)
	}
	leader, _, _ := s.GetCharacter(ctx, user, 2)
	second, _, _ := s.GetCharacter(ctx, user, 3)
	outside, _, _ := s.GetCharacter(ctx, user, 1)
	if leader.Level != 2 || leader.XP != 75 || second.Level != 2 || second.XP != 0 || outside.Level != 1 {
		t.Fatalf("after XP: leader=%d/%d second=%d/%d outside=%d",
			leader.Level, leader.XP, second.Level, second.XP, outside.Level)
	}

	if err := s.SetMoves(ctx, user, 2, []string{"slash", "counter_blow"}); err != nil {
		t.Fatal(err)
	}
	evs := data.Stats{Attack: 10, Speed: 3}
	if err := s.SetEVs(ctx, user, 2, evs); err != nil {
		t.Fatal(err)
	}
	leader, _, _ = s.GetCharacter(ctx, user, 2)
	if !slices.Equal(leader.Moves, []string{"slash", "counter_blow"}) || leader.EVs != evs {
		t.Fatalf("moves %v, EVs %+v", leader.Moves, leader.EVs)
	}
}
