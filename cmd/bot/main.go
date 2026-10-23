package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/TheOreoTM/friemon-go/internal/bot"
	"github.com/TheOreoTM/friemon-go/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env into the environment if it exists. Variables already set in the shell win.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("load .env: %v", err)
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_TOKEN is not set")
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	// Optional: register commands to one test server so changes show up instantly.
	guildID := os.Getenv("DISCORD_GUILD_ID")

	ctx := context.Background()
	st, err := store.Open(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	b, err := bot.New(token, guildID, st)
	if err != nil {
		log.Fatal(err)
	}
	if err := b.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer b.Close(ctx)

	// Block until Ctrl+C or a termination signal.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("shutting down")
}
