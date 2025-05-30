package scheduler

import (
	"log/slog"

	"github.com/theoreotm/friemon/internal/infrastructure/scheduler/tasks"
)

// SetupAsynqScheduler creates and configures the Asynq scheduler
func SetupAsynqScheduler(redisAddr, redisPassword string, redisDB int, logger *slog.Logger, deps tasks.BotDependencies) (*Scheduler, error) {
	config := AsynqConfig{
		RedisAddr:     redisAddr,
		RedisPassword: redisPassword,
		RedisDB:       redisDB,
		Concurrency:   10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
	}

	scheduler := NewAsynqScheduler(config, logger)

	setupTaskHandlers(scheduler, deps)

	return scheduler, nil
}

// setupTaskHandlers registers all task handlers
func setupTaskHandlers(scheduler *Scheduler, deps tasks.BotDependencies) {
	spawnHandlers := tasks.NewSpawnTaskHandlers(deps)

	scheduler.On("disable_spawn_button", spawnHandlers.DisableSpawnButton)
	scheduler.On("cleanup_channel", spawnHandlers.CleanupChannel)

}
