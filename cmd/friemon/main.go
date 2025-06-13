package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/handler/middleware"
	"github.com/joho/godotenv"
	"github.com/theoreotm/friemon/internal/application/bot"
	"github.com/theoreotm/friemon/internal/application/commands"
	"github.com/theoreotm/friemon/internal/application/components"
	"github.com/theoreotm/friemon/internal/application/handlers"
	"github.com/theoreotm/friemon/internal/pkg/logger"
	"go.uber.org/zap"
)

var (
	commit = "unknown"
	branch = "unknown"
	dev    = false
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Don't fail if .env doesn't exist (for production)
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	flag.BoolVar(&dev, "dev", false, "Enable development mode")
	flag.Parse()
	// Load configuration
	cfg, err := bot.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logging
	if err := logger.Initialize(cfg.Log); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Ensure we sync logs before exiting
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Printf("Failed to sync logger: %v\n", err)
		}
	}()

	// Get a logger for main
	log := logger.NewLogger("main")

	log.Info("Starting Friemon Bot",
		logger.Component("main"),
		zap.String("commit", commit),
		zap.String("branch", branch),
		zap.Bool("dev", dev),
	)

	// Create build info
	buildInfo := bot.BuildInfo{
		Version: cfg.Bot.Version,
		Commit:  commit,
		Branch:  branch,
	}

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create bot instance
	b := bot.New(*cfg, buildInfo, ctx)
	r := handler.New()

	// Setup bot with event listeners
	if err := b.SetupBot(handlers.OnMessage(b)); err != nil {
		log.Fatal("Failed to setup bot", logger.ErrorField(err))
	}

	r.Use(middleware.Logger)
	r.Command("/character", commands.HandleCharacter(b))
	r.Command("/info", commands.HandleInfo(b))
	r.Command("/list", commands.HandleList(b))
	r.Command("/battle", commands.HandleBattle(b))
	r.Component("/battle_challenge_accept/{challenge_id}", components.HandleChallengeAccept(b))
	r.Component("/battle_challenge_decline/{challenge_id}", components.HandleChallengeDecline(b))

	var applicationCmds []discord.ApplicationCommandCreate
	for _, cmd := range commands.Commands {
		applicationCmds = append(applicationCmds, cmd.Cmd)
	}

	if _, err := b.Client.Rest().SetGlobalCommands(b.Client.ApplicationID(), applicationCmds); err != nil {
		slog.Error("error while registering global commands", slog.Any("err", err))
	}
	slog.Info("registered global commands", slog.Any("commands", applicationCmds))

	// Connect to Discord
	if err := b.Client.OpenGateway(ctx); err != nil {
		log.Fatal("Failed to connect to Discord", logger.ErrorField(err))
	}

	log.Info("Bot connected successfully", logger.Component("main"))

	// Wait for interrupt signal
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Bot is now running. Press CTRL+C to exit.", logger.Component("main"))
	<-s

	log.Info("Shutting down...", logger.Component("main"))

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := b.Close(shutdownCtx); err != nil {
		log.Error("Error during shutdown", logger.ErrorField(err))
	}

	log.Info("Bot shutdown complete", logger.Component("main"))
}
