// Command seed is a developer tool: it gives players ready-to-battle characters.
//
//	go run ./cmd/seed -list                       show users and how many characters they have
//	go run ./cmd/seed -users 123,456 -level 50    give each user 4 random characters as their party
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"

	"github.com/disgoorg/snowflake/v2"
	"github.com/joho/godotenv"

	"github.com/TheOreoTM/friemon-go/internal/data"
	"github.com/TheOreoTM/friemon-go/internal/game"
	"github.com/TheOreoTM/friemon-go/internal/store"
)

func main() {
	list := flag.Bool("list", false, "list users and their character counts")
	users := flag.String("users", "", "comma-separated Discord user IDs to give characters to")
	level := flag.Int("level", 50, "level of the new characters")
	count := flag.Int("count", game.MaxParty, "characters per user")
	flag.Parse()

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("load .env: %v", err)
	}
	ctx := context.Background()
	st, err := store.Open(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	if *list {
		ids, err := st.UserIDs(ctx)
		if err != nil {
			log.Fatal(err)
		}
		for _, id := range ids {
			n, err := st.CountCharacters(ctx, id)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s\t%d characters\n", id, n)
		}
		return
	}

	if *users == "" {
		flag.Usage()
		os.Exit(2)
	}
	r := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	for _, raw := range strings.Split(*users, ",") {
		userID, err := snowflake.Parse(strings.TrimSpace(raw))
		if err != nil {
			log.Fatalf("bad user ID %q: %v", raw, err)
		}
		var party []int
		for _, i := range r.Perm(len(data.Characters))[:min(*count, len(data.Characters))] {
			c := data.Characters[i]
			caught := game.RollCaught(r, c)
			caught.Level = *level
			caught.Moves = game.StartingMoves(r, c, *level)
			idx, err := st.AddCaught(ctx, userID, caught)
			if err != nil {
				log.Fatal(err)
			}
			if len(party) < game.MaxParty {
				party = append(party, idx)
			}
			fmt.Printf("%s: #%d level %d %s (%s)\n", userID, idx, caught.Level, c.Name, strings.Join(caught.Moves, ", "))
		}
		if err := st.SetParty(ctx, userID, party); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: party set to %v\n", userID, party)
	}
}
