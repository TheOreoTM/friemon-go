package scheduler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/theoreotm/friemon/internal/infrastructure/scheduler/tasks"
)

// Dependencies interface that the bot must implement
type BotDeps interface {
	GetRestClient() tasks.RestClient
	GetCache() tasks.CacheClient
	GetRedis() interface {
		// Add Redis interface methods as needed
	}
}

// SetupScheduler initializes and configures the scheduler
func SetupScheduler(deps BotDeps, redisClient redis.Client) (*Scheduler, error) {
	backend, err := NewRedisBackendWithClient(&redisClient, "friemon_scheduler")
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis backend: %w", err)
	}

	// Create scheduler
	sched := New(backend)

	// Create task handlers
	spawnHandlers := tasks.NewSpawnTaskHandlers(deps)

	// Register task handlers
	// Convert tasks.TaskData to scheduler.TaskData
	sched.On("disable_spawn_button", func(ctx context.Context, data TaskData) error {
		return spawnHandlers.DisableSpawnButton(ctx, tasks.TaskData(data))
	})
	sched.On("cleanup_channel", func(ctx context.Context, data TaskData) error {
		return spawnHandlers.CleanupChannel(ctx, tasks.TaskData(data))
	})

	// Start the scheduler
	if err := sched.Start(); err != nil {
		return nil, fmt.Errorf("failed to start scheduler: %w", err)
	}

	slog.Info("Scheduler initialized and started")
	return sched, nil
}
