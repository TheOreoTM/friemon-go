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

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/theoreotm/friemon/friemon"
	"github.com/theoreotm/friemon/friemon/commands"
	"github.com/theoreotm/friemon/friemon/components"
	"github.com/theoreotm/friemon/friemon/handlers"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	shouldSyncCommands := flag.Bool("sync-commands", false, "Whether to sync commands to discord")
	shouldNuke := flag.Bool("nuke", false, "Whether to nuke the database")
	path := flag.String("config", "config.toml", "path to config")
	flag.Parse()

	cfg, err := friemon.LoadConfig(*path)
	setupLogger(cfg.Log)

	if err != nil {
		slog.Error("Failed to read config", slog.Any("err", err))
		os.Exit(-1)
	}

	slog.Info("Starting friemon...", slog.String("version", version), slog.String("commit", commit))
	slog.Info("Syncing commands", slog.Bool("sync", *shouldSyncCommands))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := friemon.New(*cfg, version, commit, ctx)

	h := handler.New()
	for _, cmd := range commands.Commands {
		slog.Info("Registering command", slog.String("command", cmd.Cmd.CommandName()))
		h.Command(fmt.Sprintf("/%s", cmd.Cmd.CommandName()), cmd.Handler(b))

		if cmd.Autocomplete != nil {
			h.Autocomplete(fmt.Sprintf("/%s", cmd.Cmd.CommandName()), cmd.Autocomplete(b))
		}
	}

	h.Component("/test-button", components.TestComponent)

	if err = b.SetupBot(h, bot.NewListenerFunc(b.OnReady), handlers.MessageHandler(b)); err != nil {
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

	if *shouldSyncCommands {
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

func setupLogger(cfg friemon.LogConfig) {
	opts := &slog.HandlerOptions{
		AddSource: cfg.AddSource,
		Level:     cfg.Level,
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
