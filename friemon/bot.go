package friemon

import (
	"context"
	"fmt"
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
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/theoreotm/friemon/friemon/db"
	"github.com/theoreotm/friemon/friemon/memstore"
)

func New(cfg Config, buildInfo BuildInfo, ctx context.Context) *Bot {
	db, conn, err := db.NewDB(cfg.Database)
	if err != nil {
		slog.Error("failed to initialize database: %v", slog.String("err", err.Error()))
		os.Exit(-1)
	}

	slog.Info("Connected to database", slog.String("database", cfg.Database.String()))

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		slog.Error("failed to connect to Redis: %v", slog.String("err", err.Error()))
		os.Exit(-1)
	}

	slog.Info("Connected to Redis", slog.String("addr", cfg.Redis.Addr))

	b := &Bot{
		Cfg:       cfg,
		Paginator: paginator.New(),
		BuildInfo: buildInfo,
		Context:   ctx,
		conn:      conn,
		Redis:     redisClient,
	}

	b.DB = db
	b.Cache, err = memstore.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		slog.Error("Failed to create Redis cache", slog.Any("err", err))
		os.Exit(-1)
	}

	return b
}

type Bot struct {
	Cfg       Config
	Client    bot.Client
	Paginator *paginator.Manager
	DB        *db.Queries
	Cache     memstore.Cache
	BuildInfo BuildInfo
	Context   context.Context
	conn      *pgx.Conn
	Redis     *redis.Client
}

type BuildInfo struct {
	Version string
	Commit  string
	Branch  string
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

func (b *Bot) Close(ctx context.Context) error {
	slog.Info("Closing friemon...")
	b.Client.Close(ctx)

	// Close Redis connection
	slog.Info("Closing Redis connection...")
	if err := b.Redis.Close(); err != nil {
		return fmt.Errorf("error closing Redis connection: %w", err)
	}

	// Close pgx connection
	slog.Info("Closing pgx connection...")
	if err := b.conn.Close(ctx); err != nil {
		return fmt.Errorf("error closing pgx connection: %w", err)
	}

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
