package scheduler

import (
	"fmt"
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

// Store stores a task in memory
func (m *MemoryBackend) Store(task Task, delay time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.tasks[task.ID] = &scheduledTask{
		task:      task,
		expiresAt: time.Now().Add(delay),
	}

	fmt.Printf("Task scheduled: %s (type: %s) in %v\n",
		task.ID, task.Type, delay)
	return nil
}

// Cancel cancels a scheduled task
func (m *MemoryBackend) Cancel(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[taskID]; !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	delete(m.tasks, taskID)
	fmt.Printf("Task cancelled: %s\n", taskID)
	return nil
}

// StartListening starts the background cleanup process
func (m *MemoryBackend) StartListening(onExpire func(task Task)) error {
	m.onExpire = onExpire
	m.ticker = time.NewTicker(time.Second) // Check every second for demo

	go func() {
		defer m.ticker.Stop()

		for {
			select {
			case <-m.ticker.C:
				m.processExpiredTasks()
			case <-m.done:
				return
			}
		}
	}()

	return nil
}

// Stop stops the memory backend
func (m *MemoryBackend) Stop() error {
	close(m.done)
	if m.ticker != nil {
		m.ticker.Stop()
	}
	return nil
}

// Ping always returns nil for memory backend
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

	for _, id := range expiredIDs {
		delete(m.tasks, id)
	}

	m.mu.Unlock()

	for _, task := range expiredTasks {
		if m.onExpire != nil {
			m.onExpire(task)
		}
	}
}
