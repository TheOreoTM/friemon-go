package friemon

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/paginator"
	"github.com/theoreotm/friemon/friemon/db"
)

func New(cfg Config, version string, commit string) *Bot {
	b := &Bot{
		Cfg:       cfg,
		Paginator: paginator.New(),
		Version:   version,
		Commit:    commit,
	}

	db, conn, err := db.NewDB(cfg.Database)
	if err != nil {
		slog.Error("failed to initialize database: %v", slog.String("err", err.Error()))
		os.Exit(-1)
	}
	defer conn.Close(context.Background())

	slog.Info("Connected to database", slog.String("database", cfg.Database.String()))

	b.Database = db

	return b
}

type Bot struct {
	Cfg       Config
	Client    bot.Client
	Paginator *paginator.Manager
	Database  *db.Queries
	Version   string
	Commit    string
}

func (b *Bot) SetupBot(listeners ...bot.EventListener) error {
	client, err := disgo.New(b.Cfg.Bot.Token,
		bot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildMessages, gateway.IntentMessageContent)),
		bot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
		bot.WithEventListeners(b.Paginator),
		bot.WithEventListeners(listeners...),
	)
	if err != nil {
		return err
	}

	b.Client = client
	return nil
}

func (b *Bot) OnReady(_ *events.Ready) {
	slog.Info("friemon ready")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := b.Client.SetPresence(ctx, gateway.WithListeningActivity("you"), gateway.WithOnlineStatus(discord.OnlineStatusOnline)); err != nil {
		slog.Error("Failed to set presence", slog.Any("err", err))
	}
}
