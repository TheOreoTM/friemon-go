package bot

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo"
	dbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/paginator"
	"github.com/redis/go-redis/v9"

	"github.com/theoreotm/friemon/internal/core/game"
	"github.com/theoreotm/friemon/internal/infrastructure/db"
	"github.com/theoreotm/friemon/internal/infrastructure/memstore"
	"github.com/theoreotm/friemon/internal/infrastructure/scheduler"
	"github.com/theoreotm/friemon/internal/infrastructure/scheduler/tasks"
)

type Bot struct {
	Cfg           Config
	Client        dbot.Client
	Paginator     *paginator.Manager
	DB            *db.DB
	Cache         memstore.Cache
	BuildInfo     BuildInfo
	Context       context.Context
	Redis         *redis.Client
	Scheduler     *scheduler.Scheduler
	BattleManager *game.BattleManager
}

func (b *Bot) GetRestClient() tasks.RestClient {
	return b.Client.Rest()
}

func (b *Bot) GetCache() tasks.CacheClient {
	return b.Cache
}

func New(cfg Config, buildInfo BuildInfo, ctx context.Context) *Bot {
	return &Bot{
		Cfg:           cfg,
		BuildInfo:     buildInfo,
		Context:       ctx,
		BattleManager: game.NewBattleManager(),
	}
}

type BuildInfo struct {
	Version string
	Commit  string
	Branch  string
}

func (b *Bot) SetupBot(listeners ...dbot.EventListener) error {
	// Database setup
	dbConn, err := db.NewDB(b.Cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	b.DB = dbConn

	// Auto-migrate tables
	if err := b.DB.AutoMigrate(); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	// Redis setup
	redisClient := redis.NewClient(&redis.Options{
		Addr:     b.Cfg.Redis.Addr,
		Password: b.Cfg.Redis.Password,
		DB:       b.Cfg.Redis.DB,
	})
	b.Redis = redisClient

	// Cache setup
	redisCache, err := memstore.NewRedisCache(b.Cfg.Redis.Addr, b.Cfg.Redis.Password, b.Cfg.Redis.DB)
	if err != nil {
		slog.Warn("Failed to connect to Redis, falling back to memory cache", "error", err)
		redisCache = memstore.NewMemoryCache()
	}
	b.Cache = redisCache

	// Scheduler setup
	scheduler, err := scheduler.SetupAsynqScheduler(
		b.Cfg.Redis.Addr,
		b.Cfg.Redis.Password,
		b.Cfg.Redis.DB,
		slog.Default(),
		b,
	)
	if err != nil {
		return fmt.Errorf("failed to setup scheduler: %w", err)
	}
	b.Scheduler = scheduler

	// Discord bot setup
	client, err := disgo.New(b.Cfg.Bot.Token,
		dbot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentGuilds,
				gateway.IntentGuildMessages,
				gateway.IntentMessageContent,
			),
		),
		dbot.WithCacheConfigOpts(
			cache.WithCaches(cache.FlagChannels, cache.FlagGuilds),
		),
		dbot.WithEventListeners(append(listeners, &events.ListenerAdapter{
			OnReady: b.OnReady,
		})...),
	)
	if err != nil {
		return fmt.Errorf("error while building client: %w", err)
	}
	b.Client = client

	// Paginator setup
	b.Paginator = paginator.New()

	return nil
}

func (b *Bot) Close(ctx context.Context) error {
	if b.Client != nil {
		b.Client.Close(ctx)
	}

	if b.DB != nil {
		if err := b.DB.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	if b.Redis != nil {
		if err := b.Redis.Close(); err != nil {
			return fmt.Errorf("failed to close Redis: %w", err)
		}
	}

	if b.Scheduler != nil {
		if err := b.Scheduler.Stop(); err != nil {
			return fmt.Errorf("failed to stop scheduler: %w", err)
		}
	}

	return nil
}

func (b *Bot) OnReady(_ *events.Ready) {
	slog.Info("Bot is ready!")
}
