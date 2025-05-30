package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBackend implements the Backend interface using Redis
type RedisBackend struct {
	client   *redis.Client
	pubsub   *redis.PubSub
	ctx      context.Context
	onExpire func(task Task)
	done     chan struct{}
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisBackend creates a new Redis backend
func NewRedisBackend(config RedisConfig) (*RedisBackend, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Enable keyspace notifications for expired events
	if err := client.ConfigSet(ctx, "notify-keyspace-events", "Ex").Err(); err != nil {
		return nil, fmt.Errorf("failed to enable keyspace notifications: %w", err)
	}

	return &RedisBackend{
		client: client,
		ctx:    ctx,
		done:   make(chan struct{}),
	}, nil
}

// StoreTask stores a task in Redis with expiration
func (r *RedisBackend) StoreTask(task Task, delay time.Duration) error {
	// Serialize task data
	taskData, err := task.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Store task metadata that won't expire (cleaned up manually)
	metadataKey := r.metadataKey(task.ID)
	if err := r.client.Set(r.ctx, metadataKey, taskData, 0).Err(); err != nil {
		return fmt.Errorf("failed to store task metadata: %w", err)
	}

	// Store a trigger key that will expire and trigger our handler
	triggerKey := r.triggerKey(task.ID)
	if err := r.client.Set(r.ctx, triggerKey, "trigger", delay).Err(); err != nil {
		// Clean up metadata if trigger creation fails
		r.client.Del(r.ctx, metadataKey)
		return fmt.Errorf("failed to schedule task trigger: %w", err)
	}

	slog.Info("Task scheduled",
		slog.String("task_id", task.ID),
		slog.String("type", string(task.Type)),
		slog.Duration("delay", delay))

	return nil
}

// CancelTask cancels a scheduled task
func (r *RedisBackend) CancelTask(taskID string) error {
	// Delete both the trigger and metadata
	triggerKey := r.triggerKey(taskID)
	metadataKey := r.metadataKey(taskID)

	pipe := r.client.Pipeline()
	pipe.Del(r.ctx, triggerKey)
	pipe.Del(r.ctx, metadataKey)
	_, err := pipe.Exec(r.ctx)

	if err != nil {
		return fmt.Errorf("failed to cancel task %s: %w", taskID, err)
	}

	slog.Info("Task cancelled", slog.String("task_id", taskID))
	return nil
}

// StartListening starts listening for expired Redis keys
func (r *RedisBackend) StartListening(onExpire func(task Task)) error {
	r.onExpire = onExpire

	// Subscribe to expired key events
	r.pubsub = r.client.PSubscribe(r.ctx, "__keyevent@*__:expired")

	go func() {
		defer r.pubsub.Close()

		ch := r.pubsub.Channel()
		for {
			select {
			case msg := <-ch:
				if msg != nil {
					r.handleExpiredKey(msg.Payload)
				}
			case <-r.done:
				slog.Info("Redis backend listener stopping")
				return
			}
		}
	}()

	slog.Info("Redis backend started listening for expired keys")
	return nil
}

// Stop stops the Redis backend
func (r *RedisBackend) Stop() error {
	close(r.done)

	if r.pubsub != nil {
		if err := r.pubsub.Close(); err != nil {
			slog.Error("Failed to close Redis pubsub", slog.Any("error", err))
		}
	}

	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	slog.Info("Redis backend stopped")
	return nil
}

// Ping checks the health of the Redis connection
func (r *RedisBackend) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// handleExpiredKey processes expired Redis keys
func (r *RedisBackend) handleExpiredKey(key string) {
	// Only process our scheduled task trigger keys
	if !strings.HasPrefix(key, "scheduler:trigger:") {
		return
	}

	// Extract task ID from trigger key
	taskID := strings.TrimPrefix(key, "scheduler:trigger:")

	// Get task metadata
	metadataKey := r.metadataKey(taskID)
	taskDataBytes, err := r.client.Get(r.ctx, metadataKey).Bytes()
	if err != nil {
		slog.Error("Failed to get task metadata",
			slog.String("task_id", taskID),
			slog.Any("error", err))
		return
	}

	// Parse task data
	task, err := TaskFromJSON(taskDataBytes)
	if err != nil {
		slog.Error("Failed to unmarshal task data",
			slog.String("task_id", taskID),
			slog.Any("error", err))
		// Clean up corrupted metadata
		r.client.Del(r.ctx, metadataKey)
		return
	}

	// Clean up metadata
	r.client.Del(r.ctx, metadataKey)

	// Execute the task
	if r.onExpire != nil {
		r.onExpire(*task)
	}
}

// Key generation helpers
func (r *RedisBackend) triggerKey(taskID string) string {
	return fmt.Sprintf("scheduler:trigger:%s", taskID)
}

func (r *RedisBackend) metadataKey(taskID string) string {
	return fmt.Sprintf("scheduler:metadata:%s", taskID)
}

// GetStats returns Redis backend statistics
func (r *RedisBackend) GetStats() (map[string]interface{}, error) {
	info, err := r.client.Info(r.ctx, "memory", "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Count scheduled tasks
	triggerKeys, err := r.client.Keys(r.ctx, "scheduler:trigger:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count trigger keys: %w", err)
	}

	metadataKeys, err := r.client.Keys(r.ctx, "scheduler:metadata:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count metadata keys: %w", err)
	}

	return map[string]interface{}{
		"scheduled_tasks": len(triggerKeys),
		"metadata_keys":   len(metadataKeys),
		"redis_info":      info,
	}, nil
}
