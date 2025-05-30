package friemon

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/internal/bot"
	"github.com/theoreotm/friemon/internal/commands"
	"github.com/theoreotm/friemon/internal/components"
	"github.com/theoreotm/friemon/internal/handlers"
)

var (
	commit = "unknown"
	branch = "unknown"
	dev    = false
)

func main() {
	// Set up initial basic logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	shouldNuke := flag.Bool("nuke", false, "Whether to nuke the database")
	path := flag.String("config", "config.toml", "path to config")
	flag.StringVar(&commit, "commit", "unknown", "commit")
	flag.StringVar(&branch, "branch", "unknown", "branch")
	flag.Parse()

	// Log the working directory and config path
	wd, _ := os.Getwd()
	slog.Info("Current working directory", "path", wd)
	slog.Info("Loading config from", "path", *path)

	// Check if config file exists
	if _, err := os.Stat(*path); os.IsNotExist(err) {
		slog.Error("Config file does not exist", "path", *path)
		os.Exit(-1)
	}

	cfg, err := bot.LoadConfig(*path)
	if err != nil {
		slog.Error("Failed to read config", slog.Any("err", err))
		os.Exit(-1)
	}

	setupLogger(cfg.Log)
	dev = cfg.Bot.DevMode
	shouldSyncCommands := cfg.Bot.SyncCommands

	// Log environment variables
	slog.Debug("Environment variables",
		"BOT_TOKEN", os.Getenv("BOT_TOKEN") != "",
		"DB_HOST", os.Getenv("DB_HOST"),
		"DB_USER", os.Getenv("DB_USER"),
		"REDIS_ADDR", os.Getenv("REDIS_ADDR"),
	)

	slog.Info("Starting friemon...", slog.String("version", cfg.Bot.Version), slog.String("commit", commit), slog.String("branch", branch))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buildInfo := bot.BuildInfo{
		Version: cfg.Bot.Version,
		Commit:  commit,
		Branch:  branch,
	}
	b := bot.New(*cfg, buildInfo, ctx)
	h := handler.New()

	for _, cmd := range commands.Commands {
		slog.Debug("Registering command", slog.String("command", cmd.Cmd.CommandName()))
		h.Command(fmt.Sprintf("/%s", cmd.Cmd.CommandName()), cmd.Handler(b))

		if cmd.Autocomplete != nil {
			h.Autocomplete(fmt.Sprintf("/%s", cmd.Cmd.CommandName()), cmd.Autocomplete(b))
		}
	}

	for key, comp := range components.Components {
		h.Component(key, comp(b))
	}

	if err = b.SetupBot(h, disgobot.NewListenerFunc(b.OnReady), handlers.OnMessage(b)); err != nil {
		slog.Error("Failed to setup bot", slog.Any("err", err))
		os.Exit(-1)
	}

	defer func() {
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := b.Close(shutdownCtx); err != nil {
			slog.Error("Failed to close friemon", slog.Any("err", err))
		}
	}()

	if shouldSyncCommands {
		var cmds []discord.ApplicationCommandCreate
		for _, cmd := range commands.Commands {
			cmds = append(cmds, cmd.Cmd)
		}

		slog.Info("Syncing commands", slog.Any("guild_ids", cfg.Bot.DevGuilds))
		if err = handler.SyncCommands(b.Client, cmds, cfg.Bot.DevGuilds); err != nil {
			slog.Error("Failed to sync commands", slog.Any("err", err))
		}
	}

	if *shouldNuke {
		slog.Info("Nuking database")
		if err = b.DB.DeleteEverything(ctx); err != nil {
			slog.Error("Failed to nuke database", slog.Any("err", err))
		}
	}

	if err = b.Client.OpenGateway(ctx); err != nil {
		slog.Error("Failed to open gateway", slog.Any("err", err))
		os.Exit(-1)
	}

	slog.Info("Bot is running. Press CTRL-C to exit.")
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
	slog.Info("Shutting down bot...")
}

func setupLogger(cfg bot.LogConfig) {
	level := cfg.Level
	if dev {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     level,
	}

	var sHandler slog.Handler
	switch cfg.Format {
	case "json":
		sHandler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		sHandler = slog.NewTextHandler(os.Stdout, opts)
	default:
		slog.Error("Unknown log format", slog.String("format", cfg.Format))
		os.Exit(-1)
	}
	slog.SetDefault(slog.New(sHandler))
}
