package scheduler

import (
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// MemoryBackend implements the Backend interface using in-memory storage
type MemoryBackend struct {
	tasks    map[string]*scheduledTask
	mu       sync.RWMutex
	onExpire func(task Task)
	ticker   *time.Ticker
	done     chan struct{}
}

type scheduledTask struct {
	task      Task
	expiresAt time.Time
}

// NewMemoryBackend creates a new in-memory backend
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		tasks: make(map[string]*scheduledTask),
		done:  make(chan struct{}),
	}
}

// StoreTask stores a task in memory
func (m *MemoryBackend) StoreTask(task Task, delay time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks[task.ID] = &scheduledTask{
		task:      task,
		expiresAt: time.Now().Add(delay),
	}

	slog.Info("Task scheduled in memory",
		slog.String("task_id", task.ID),
		slog.String("type", string(task.Type)),
		slog.Duration("delay", delay))

	return nil
}

// CancelTask cancels a scheduled task
func (m *MemoryBackend) CancelTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	delete(m.tasks, taskID)
	slog.Info("Task cancelled", slog.String("task_id", taskID))
	return nil
}

// StartListening starts the background cleanup process
func (m *MemoryBackend) StartListening(onExpire func(task Task)) error {
	m.onExpire = onExpire

	// Start cleanup ticker (check every 30 seconds)
	m.ticker = time.NewTicker(30 * time.Second)

	go func() {
		defer m.ticker.Stop()

		for {
			select {
			case <-m.ticker.C:
				m.processExpiredTasks()
			case <-m.done:
				slog.Info("Memory backend listener stopping")
				return
			}
		}
	}()

	slog.Info("Memory backend started")
	return nil
}

// Stop stops the memory backend
func (m *MemoryBackend) Stop() error {
	close(m.done)

	if m.ticker != nil {
		m.ticker.Stop()
	}

	slog.Info("Memory backend stopped")
	return nil
}

// Ping always returns nil for memory backend (always healthy)
func (m *MemoryBackend) Ping() error {
	return nil
}

// processExpiredTasks checks for and processes expired tasks
func (m *MemoryBackend) processExpiredTasks() {
	now := time.Now()

	m.mu.Lock()
	var expiredTasks []Task
	var expiredIDs []string

	for id, scheduled := range m.tasks {
		if now.After(scheduled.expiresAt) {
			expiredTasks = append(expiredTasks, scheduled.task)
			expiredIDs = append(expiredIDs, id)
		}
	}

	// Remove expired tasks
	for _, id := range expiredIDs {
		delete(m.tasks, id)
	}

	m.mu.Unlock()

	// Execute expired tasks (outside of mutex)
	for _, task := range expiredTasks {
		if m.onExpire != nil {
			m.onExpire(task)
		}
	}

	if len(expiredTasks) > 0 {
		slog.Info("Processed expired tasks", slog.Int("count", len(expiredTasks)))
	}
}

// GetStats returns memory backend statistics
func (m *MemoryBackend) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasksByType := make(map[string]int)
	for _, scheduled := range m.tasks {
		tasksByType[string(scheduled.task.Type)]++
	}

	return map[string]interface{}{
		"total_tasks":   len(m.tasks),
		"tasks_by_type": tasksByType,
	}
}
