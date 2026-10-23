package store

import (
	"context"
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

func TestSelectXPAndMoves(t *testing.T) {
	s := openTestStore(t)
	ctx := context.Background()
	user := testID(t, s)

	// No characters yet: nothing selected, and XP goes nowhere.
	if _, ok, err := s.SelectedCharacter(ctx, user); err != nil || ok {
		t.Fatalf("SelectedCharacter with no characters: ok=%v err=%v", ok, err)
	}
	if up, err := s.AddXPToSelected(ctx, user, 100); err != nil || up != nil {
		t.Fatalf("AddXPToSelected with no characters = %+v, %v", up, err)
	}

	s.AddCaught(ctx, user, game.Caught{CharacterID: "fern", Level: 1, Moves: []string{"zoltraak"}})
	s.AddCaught(ctx, user, game.Caught{CharacterID: "stark", Level: 1})

	// The first catch is selected automatically.
	sel, ok, err := s.SelectedCharacter(ctx, user)
	if err != nil || !ok || sel.Idx != 1 || !slices.Equal(sel.Moves, []string{"zoltraak"}) {
		t.Fatalf("SelectedCharacter = %+v, %v, %v", sel, ok, err)
	}

	if ok, err := s.SelectCharacter(ctx, user, 99); err != nil || ok {
		t.Fatalf("selecting a missing number: ok=%v err=%v", ok, err)
	}
	if ok, err := s.SelectCharacter(ctx, user, 2); err != nil || !ok {
		t.Fatalf("SelectCharacter(2): ok=%v err=%v", ok, err)
	}

	// 75 XP takes level 1 to level 2, and only the selected character (#2) gets it.
	up, err := s.AddXPToSelected(ctx, user, 80)
	if err != nil || up == nil || up.Idx != 2 || up.NewLevel != 2 {
		t.Fatalf("AddXPToSelected = %+v, %v", up, err)
	}
	stark, _, _ := s.GetCharacter(ctx, user, 2)
	fern, _, _ := s.GetCharacter(ctx, user, 1)
	if stark.Level != 2 || stark.XP != 5 || fern.Level != 1 {
		t.Fatalf("after XP: stark=%+v fern=%+v", stark, fern)
	}

	if err := s.SetMoves(ctx, user, 2, []string{"slash", "counter_blow"}); err != nil {
		t.Fatal(err)
	}
	stark, _, _ = s.GetCharacter(ctx, user, 2)
	if !slices.Equal(stark.Moves, []string{"slash", "counter_blow"}) {
		t.Fatalf("moves = %v", stark.Moves)
	}
	if _, ok, _ := s.GetCharacter(ctx, user, 3); ok {
		t.Fatal("GetCharacter found a number the user does not have")
	}
}
