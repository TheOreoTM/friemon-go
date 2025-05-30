package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisBackend implements the Backend interface using Redis
type RedisBackend struct {
	client    *redis.Client
	pubsub    *redis.PubSub
	ctx       context.Context
	onExpire  func(task Task)
	done      chan struct{}
	keyPrefix string
	ownClient bool // Track if we own the client for cleanup
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr      string
	Password  string
	DB        int
	KeyPrefix string // Optional prefix for all keys (useful for namespacing)
}

// NewRedisBackend creates a new Redis backend with its own connection
func NewRedisBackend(config RedisConfig) (*RedisBackend, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	keyPrefix := config.KeyPrefix
	if keyPrefix == "" {
		keyPrefix = "scheduler"
	}

	backend := &RedisBackend{
		client:    client,
		ctx:       ctx,
		done:      make(chan struct{}),
		keyPrefix: keyPrefix,
		ownClient: true, // We created this client, so we own it
	}

	if err := backend.enableKeyspaceNotifications(); err != nil {
		client.Close()
		return nil, err
	}

	return backend, nil
}

// NewRedisBackendWithClient creates a new Redis backend using an existing Redis client
func NewRedisBackendWithClient(client *redis.Client, keyPrefix string) (*RedisBackend, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client cannot be nil")
	}

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis client is not connected: %w", err)
	}

	if keyPrefix == "" {
		keyPrefix = "scheduler"
	}

	backend := &RedisBackend{
		client:    client,
		ctx:       ctx,
		done:      make(chan struct{}),
		keyPrefix: keyPrefix,
		ownClient: false, // We don't own this client
	}

	if err := backend.enableKeyspaceNotifications(); err != nil {
		return nil, err
	}

	return backend, nil
}

// enableKeyspaceNotifications enables Redis keyspace notifications
func (r *RedisBackend) enableKeyspaceNotifications() error {
	if err := r.client.ConfigSet(r.ctx, "notify-keyspace-events", "Ex").Err(); err != nil {
		return fmt.Errorf("failed to enable keyspace notifications: %w", err)
	}
	return nil
}

// Store stores a task in Redis with expiration
func (r *RedisBackend) Store(task Task, delay time.Duration) error {
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

	fmt.Printf("Task scheduled in Redis: %s (type: %s) in %v\n",
		task.ID, task.Type, delay)

	return nil
}

// Cancel cancels a scheduled task
func (r *RedisBackend) Cancel(taskID string) error {
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

	fmt.Printf("Task cancelled in Redis: %s\n", taskID)
	return nil
}

// StartListening starts listening for expired Redis keys
func (r *RedisBackend) StartListening(onExpire func(task Task)) error {
	r.onExpire = onExpire

	// Subscribe to expired key events for all databases
	// Pattern matches expired events from any database
	r.pubsub = r.client.PSubscribe(r.ctx, "__keyevent@*__:expired")

	go func() {
		defer func() {
			if r.pubsub != nil {
				r.pubsub.Close()
			}
		}()

		ch := r.pubsub.Channel()
		for {
			select {
			case msg := <-ch:
				if msg != nil {
					r.handleExpiredKey(msg.Payload)
				}
			case <-r.done:
				fmt.Println("Redis backend listener stopping")
				return
			}
		}
	}()

	fmt.Println("Redis backend started listening for expired keys")
	return nil
}

// Stop stops the Redis backend
func (r *RedisBackend) Stop() error {
	close(r.done)

	if r.pubsub != nil {
		if err := r.pubsub.Close(); err != nil {
			fmt.Printf("Failed to close Redis pubsub: %v\n", err)
		}
	}

	// Only close the client if we own it
	if r.ownClient && r.client != nil {
		if err := r.client.Close(); err != nil {
			return fmt.Errorf("failed to close Redis client: %w", err)
		}
	}

	fmt.Println("Redis backend stopped")
	return nil
}

// Ping checks the health of the Redis connection
func (r *RedisBackend) Ping() error {
	return r.client.Ping(r.ctx).Err()
}

// handleExpiredKey processes expired Redis keys
func (r *RedisBackend) handleExpiredKey(key string) {
	// Only process our scheduled task trigger keys
	triggerPrefix := r.keyPrefix + ":trigger:"
	if !strings.HasPrefix(key, triggerPrefix) {
		return
	}

	// Extract task ID from trigger key
	taskID := strings.TrimPrefix(key, triggerPrefix)

	// Get task metadata
	metadataKey := r.metadataKey(taskID)
	taskDataBytes, err := r.client.Get(r.ctx, metadataKey).Bytes()
	if err != nil {
		fmt.Printf("Failed to get task metadata for %s: %v\n", taskID, err)
		return
	}

	// Parse task data
	task, err := TaskFromJSON(taskDataBytes)
	if err != nil {
		fmt.Printf("Failed to unmarshal task data for %s: %v\n", taskID, err)
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
	return fmt.Sprintf("%s:trigger:%s", r.keyPrefix, taskID)
}

func (r *RedisBackend) metadataKey(taskID string) string {
	return fmt.Sprintf("%s:metadata:%s", r.keyPrefix, taskID)
}

// GetStats returns Redis backend statistics
func (r *RedisBackend) GetStats() (map[string]interface{}, error) {
	// Count scheduled tasks
	triggerPattern := r.keyPrefix + ":trigger:*"
	metadataPattern := r.keyPrefix + ":metadata:*"

	triggerKeys, err := r.client.Keys(r.ctx, triggerPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count trigger keys: %w", err)
	}

	metadataKeys, err := r.client.Keys(r.ctx, metadataPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count metadata keys: %w", err)
	}

	// Get basic Redis info
	info, err := r.client.Info(r.ctx, "memory", "stats").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Count tasks by type (optional, requires scanning metadata)
	tasksByType := make(map[string]int)
	if len(metadataKeys) < 100 { // Only do this for reasonable number of tasks
		pipe := r.client.Pipeline()
		cmds := make([]*redis.StringCmd, len(metadataKeys))

		for i, key := range metadataKeys {
			cmds[i] = pipe.Get(r.ctx, key)
		}

		_, err := pipe.Exec(r.ctx)
		if err == nil {
			for _, cmd := range cmds {
				if cmd.Err() == nil {
					if task, err := TaskFromJSON([]byte(cmd.Val())); err == nil {
						tasksByType[task.Type]++
					}
				}
			}
		}
	}

	return map[string]interface{}{
		"scheduled_tasks": len(triggerKeys),
		"metadata_keys":   len(metadataKeys),
		"tasks_by_type":   tasksByType,
		"redis_info":      info,
		"key_prefix":      r.keyPrefix,
		"owns_client":     r.ownClient,
	}, nil
}

// GetTask retrieves a specific task by ID (useful for debugging)
func (r *RedisBackend) GetTask(taskID string) (*Task, error) {
	metadataKey := r.metadataKey(taskID)
	taskDataBytes, err := r.client.Get(r.ctx, metadataKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("task %s not found", taskID)
		}
		return nil, fmt.Errorf("failed to get task %s: %w", taskID, err)
	}

	return TaskFromJSON(taskDataBytes)
}

// ListTasks returns all scheduled tasks (useful for debugging/monitoring)
func (r *RedisBackend) ListTasks() ([]Task, error) {
	metadataPattern := r.keyPrefix + ":metadata:*"
	metadataKeys, err := r.client.Keys(r.ctx, metadataPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata keys: %w", err)
	}

	if len(metadataKeys) == 0 {
		return []Task{}, nil
	}

	// Get all task data in a pipeline
	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(metadataKeys))

	for i, key := range metadataKeys {
		cmds[i] = pipe.Get(r.ctx, key)
	}

	_, err = pipe.Exec(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	var tasks []Task
	for _, cmd := range cmds {
		if cmd.Err() == nil {
			if task, err := TaskFromJSON([]byte(cmd.Val())); err == nil {
				tasks = append(tasks, *task)
			}
		}
	}

	return tasks, nil
}

// CleanupOrphanedData removes any orphaned metadata without corresponding triggers
func (r *RedisBackend) CleanupOrphanedData() error {
	metadataPattern := r.keyPrefix + ":metadata:*"
	metadataKeys, err := r.client.Keys(r.ctx, metadataPattern).Result()
	if err != nil {
		return fmt.Errorf("failed to list metadata keys: %w", err)
	}

	var orphanedKeys []string
	for _, metadataKey := range metadataKeys {
		// Extract task ID from metadata key
		taskID := strings.TrimPrefix(metadataKey, r.keyPrefix+":metadata:")
		triggerKey := r.triggerKey(taskID)

		// Check if corresponding trigger exists
		exists, err := r.client.Exists(r.ctx, triggerKey).Result()
		if err != nil {
			continue // Skip on error
		}

		if exists == 0 {
			orphanedKeys = append(orphanedKeys, metadataKey)
		}
	}

	if len(orphanedKeys) > 0 {
		err := r.client.Del(r.ctx, orphanedKeys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete orphaned keys: %w", err)
		}
		fmt.Printf("Cleaned up %d orphaned metadata keys\n", len(orphanedKeys))
	}

	return nil
}

// GetClient returns the underlying Redis client (useful for advanced operations)
func (r *RedisBackend) GetClient() *redis.Client {
	return r.client
}
