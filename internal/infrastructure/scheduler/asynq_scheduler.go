package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	"github.com/theoreotm/friemon/internal/types"
)

// AsynqConfig holds the configuration for the Asynq scheduler
type AsynqConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Concurrency   int
	Queues        map[string]int
}

// Scheduler wraps Asynq client and server for task scheduling
type Scheduler struct {
	client    *asynq.Client
	server    *asynq.Server
	inspector *asynq.Inspector
	mux       *asynq.ServeMux
	handlers  map[string]types.TaskHandler
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	logger    *slog.Logger
}

// NewAsynqScheduler creates a new scheduler with Asynq backend
func NewAsynqScheduler(config AsynqConfig, logger *slog.Logger) *Scheduler {
	redisOpt := asynq.RedisClientOpt{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	}

	client := asynq.NewClient(redisOpt)
	inspector := asynq.NewInspector(redisOpt)

	serverConfig := asynq.Config{
		Concurrency: config.Concurrency,
		Queues:      config.Queues,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			logger.Error("Task execution failed",
				"task_type", task.Type(),
				"error", err.Error(),
				"payload", string(task.Payload()),
			)
		}),
	}

	if len(config.Queues) == 0 {
		serverConfig.Queues = map[string]int{"default": 1}
	}

	server := asynq.NewServer(redisOpt, serverConfig)
	mux := asynq.NewServeMux()

	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		client:    client,
		server:    server,
		inspector: inspector,
		mux:       mux,
		handlers:  make(map[string]types.TaskHandler),
		ctx:       ctx,
		cancel:    cancel,
		logger:    logger,
	}
}

// Start begins processing tasks
func (s *Scheduler) Start() error {
	go func() {
		if err := s.server.Run(s.mux); err != nil {
			s.logger.Error("Failed to start task server", "error", err)
		}
	}()

	s.logger.Info("Task scheduler started")
	return nil
}

// Stop gracefully shuts down the scheduler
func (s *Scheduler) Stop() error {
	s.cancel()
	s.server.Shutdown()
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %w", err)
	}
	if err := s.inspector.Close(); err != nil {
		return fmt.Errorf("failed to close inspector: %w", err)
	}
	s.logger.Info("Task scheduler stopped")
	return nil
}

// On registers a task handler for a specific task type
func (s *Scheduler) On(taskType string, handler types.TaskHandler) *Scheduler {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[taskType] = handler
	s.mux.HandleFunc(taskType, s.createAsynqHandler(taskType))

	return s
}

// Off removes a task handler
func (s *Scheduler) Off(taskType string) *Scheduler {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.handlers, taskType)
	return s
}

// After creates a task builder with a delay
func (s *Scheduler) After(delay time.Duration) *TaskBuilder {
	return &TaskBuilder{
		scheduler: s,
		delay:     delay,
		data:      make(types.TaskData),
		queue:     "default",
		maxRetry:  3,
	}
}

// In is an alias for After
func (s *Scheduler) In(delay time.Duration) *TaskBuilder {
	return s.After(delay)
}

// At schedules a task at a specific time
func (s *Scheduler) At(when time.Time) *TaskBuilder {
	delay := time.Until(when)
	if delay < 0 {
		delay = 0
	}
	return s.After(delay)
}

// Now creates a task builder for immediate execution
func (s *Scheduler) Now() *TaskBuilder {
	return s.After(0)
}

// Cancel cancels a scheduled task by ID
func (s *Scheduler) Cancel(queue, taskID string) error {
	return s.inspector.DeleteTask(queue, taskID)
}

// CancelByType cancels all scheduled tasks of a specific type
func (s *Scheduler) CancelByType(taskType string) error {
	// Get all scheduled tasks and filter by type
	queues, err := s.inspector.Queues()
	if err != nil {
		return fmt.Errorf("failed to get queues: %w", err)
	}

	for _, queue := range queues {
		tasks, err := s.inspector.ListScheduledTasks(queue)
		if err != nil {
			continue
		}

		for _, task := range tasks {
			if task.Type == taskType {
				if err := s.inspector.DeleteTask(queue, task.ID); err != nil {
					s.logger.Warn("Failed to cancel task",
						"task_id", task.ID,
						"task_type", taskType,
						"error", err,
					)
				}
			}
		}
	}

	return nil
}

// createAsynqHandler creates an Asynq-compatible handler from our TaskHandler
func (s *Scheduler) createAsynqHandler(taskType string) func(context.Context, *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		s.mu.RLock()
		handler, exists := s.handlers[taskType]
		s.mu.RUnlock()

		if !exists {
			return fmt.Errorf("no handler registered for task type: %s", taskType)
		}

		var data types.TaskData
		if err := json.Unmarshal(task.Payload(), &data); err != nil {
			return fmt.Errorf("failed to unmarshal task data: %w", err)
		}

		s.logger.Debug("Executing task",
			"task_type", taskType,
			"task_id", task.ResultWriter().TaskID(),
		)

		return handler(ctx, data)
	}
}

// Health checks if the scheduler is healthy
func (s *Scheduler) Health() error {
	_, err := s.client.Enqueue(asynq.NewTask("health_check", nil))
	return err
}
