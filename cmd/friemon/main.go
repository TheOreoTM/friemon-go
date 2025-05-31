package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
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
	// Parse command line flags
	var configPath string
	flag.StringVar(&configPath, "config", "config.toml", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := bot.LoadConfig(configPath)
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

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Warn("Config file not found, using environment variables and defaults",
			zap.String("config_path", configPath),
		)
	}

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

	// Prepare command list
	var cmds []discord.ApplicationCommandCreate
	for _, cmd := range commands.Commands {
		cmds = append(cmds, cmd.Cmd)
	}

	// Setup bot with event listeners
	if err := b.SetupBot(handlers.OnMessage(b)); err != nil {
		log.Fatal("Failed to setup bot", logger.ErrorField(err))
	}

	// Setup command and component handlers
	h := handler.New()
	for _, cmd := range commands.Commands {
		h.Command(fmt.Sprintf("/%s", cmd.Cmd.CommandName()), cmd.Handler(b))
		if cmd.Autocomplete != nil {
			h.Autocomplete(fmt.Sprintf("/%s", cmd.Cmd.CommandName()), cmd.Autocomplete(b))
		}
	}

	for name, comp := range components.Components {
		h.Component(name, comp(b))
	}

	b.Client.AddEventListeners(h)

	// Connect to Discord
	if err := b.Client.OpenGateway(ctx); err != nil {
		log.Fatal("Failed to connect to Discord", logger.ErrorField(err))
	}

	log.Info("Bot connected successfully", logger.Component("main"))

	// Sync commands if enabled
	if cfg.Bot.SyncCommands {
		log.Info("Syncing commands...", logger.Component("main"))
		if _, err := b.Client.Rest().SetGlobalCommands(b.Client.ApplicationID(), cmds); err != nil {
			log.Error("Failed to sync commands", logger.ErrorField(err))
		} else {
			log.Info("Commands synced successfully",
				logger.Component("main"),
				zap.Int("command_count", len(cmds)),
			)
		}
	}

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
