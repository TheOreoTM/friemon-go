package bot

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/disgoorg/disgo"
	disgobot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/paginator"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/theoreotm/friemon/internal/infrastructure/db"
	"github.com/theoreotm/friemon/internal/infrastructure/memstore"
	"github.com/theoreotm/friemon/internal/infrastructure/scheduler"
)

type Bot struct {
	Cfg       Config
	Client    disgobot.Client
	Paginator *paginator.Manager
	DB        *db.Queries
	Cache     memstore.Cache
	BuildInfo BuildInfo
	Context   context.Context
	conn      *pgx.Conn
	Redis     *redis.Client
	Scheduler *scheduler.Scheduler
}

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

// Implement the dependencies interface for scheduler tasks
func (b *Bot) GetRestClient() interface {
	GetMessage(channelID, messageID snowflake.ID, opts ...rest.RequestOpt) (*discord.Message, error)
	UpdateMessage(channelID, messageID snowflake.ID, messageUpdate discord.MessageUpdate, opts ...rest.RequestOpt) (*discord.Message, error)
} {
	return b.Client.Rest()
}

func (b *Bot) GetCache() interface {
	DeleteChannelCharacter(channelID snowflake.ID) error
	ResetInteractionCount(channelID snowflake.ID) error
} {
	return b.Cache
}

type BuildInfo struct {
	Version string
	Commit  string
	Branch  string
}

func (b *Bot) SetupBot(listeners ...disgobot.EventListener) error {
	client, err := disgo.New(b.Cfg.Bot.Token,
		disgobot.WithGatewayConfigOpts(gateway.WithIntents(gateway.IntentGuilds, gateway.IntentGuildMessages, gateway.IntentMessageContent)),
		disgobot.WithCacheConfigOpts(cache.WithCaches(cache.FlagGuilds)),
		disgobot.WithEventListeners(b.Paginator),
		disgobot.WithEventListeners(listeners...),
	)
	if err != nil {
		return err
	}

	b.Client = client
	return nil
}

func (b *Bot) Close(ctx context.Context) error {
	slog.Info("Closing friemon...")

	// Stop scheduler first
	if b.Scheduler != nil {
		if err := b.Scheduler.Stop(); err != nil {
			slog.Error("Failed to stop scheduler", slog.Any("err", err))
		}
	}

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
