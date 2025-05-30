package scheduler

import (
	"encoding/json"
	"time"
)

// TaskType represents different types of scheduled tasks
type TaskType string

const (
	TaskTypeDisableSpawnButton TaskType = "disable_spawn_button"
	TaskTypeCleanupChannel     TaskType = "cleanup_channel"
	TaskTypeRemindUser         TaskType = "remind_user"
	TaskTypeResetCooldown      TaskType = "reset_cooldown"
	TaskTypeExpireItem         TaskType = "expire_item"
)

// Task represents a scheduled task with custom data
type Task struct {
	ID        string                 `json:"id"`
	Type      TaskType               `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
	ExpiresAt time.Time              `json:"expires_at"`
}

// TaskHandler is a function that processes expired tasks
type TaskHandler func(task Task) error

// Backend defines the interface for task storage backends
type Backend interface {
	// Store a task with expiration
	StoreTask(task Task, delay time.Duration) error
	
	// Cancel a scheduled task
	CancelTask(taskID string) error
	
	// Start listening for expired tasks
	StartListening(onExpire func(task Task)) error
	
	// Stop the backend and cleanup resources
	Stop() error
	
	// Health check
	Ping() error
}

// Scheduler manages scheduled tasks
type Scheduler struct {
	backend  Backend
	handlers map[TaskType]TaskHandler
}

// New creates a new scheduler with the given backend
func New(backend Backend) *Scheduler {
	return &Scheduler{
		backend:  backend,
		handlers: make(map[TaskType]TaskHandler),
	}
}

// Start initializes the scheduler and begins listening for expired tasks
func (s *Scheduler) Start() error {
	return s.backend.StartListening(s.executeTask)
}

// Stop shuts down the scheduler
func (s *Scheduler) Stop() error {
	return s.backend.Stop()
}

// RegisterHandler registers a handler for a specific task type
func (s *Scheduler) RegisterHandler(taskType TaskType, handler TaskHandler) {
	s.handlers[taskType] = handler
}

// Schedule schedules a task to be executed after the specified delay
func (s *Scheduler) Schedule(taskType TaskType, data map[string]interface{}, delay time.Duration) (string, error) {
	taskID := generateTaskID(taskType)
	
	task := Task{
		ID:        taskID,
		Type:      taskType,
		Data:      data,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(delay),
	}

	err := s.backend.StoreTask(task, delay)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// Cancel cancels a scheduled task
func (s *Scheduler) Cancel(taskID string) error {
	return s.backend.CancelTask(taskID)
}

// Health returns the health status of the scheduler
func (s *Scheduler) Health() error {
	return s.backend.Ping()
}

// executeTask executes a task by calling the appropriate handler
func (s *Scheduler) executeTask(task Task) {
	handler, exists := s.handlers[task.Type]
	if !exists {
		// Log error - no handler registered
		return
	}

	if err := handler(task); err != nil {
		// Log error - task execution failed
	}
}

// generateTaskID generates a unique task ID
func generateTaskID(taskType TaskType) string {
	return string(taskType) + ":" + generateUniqueID()
}

// generateUniqueID generates a unique identifier
func generateUniqueID() string {
	return time.Now().Format("20060102150405.000000000")
}

// Utility functions for common data operations
func (t *Task) GetString(key string) (string, bool) {
	val, exists := t.Data[key]
	if !exists {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func (t *Task) GetInt(key string) (int, bool) {
	val, exists := t.Data[key]
	if !exists {
		return 0, false
	}
	
	// Handle both int and float64 (JSON unmarshaling)
	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func (t *Task) GetBool(key string) (bool, bool) {
	val, exists := t.Data[key]
	if !exists {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

func (t *Task) MustGetString(key string) string {
	val, _ := t.GetString(key)
	return val
}

func (t *Task) MustGetInt(key string) int {
	val, _ := t.GetInt(key)
	return val
}

func (t *Task) MustGetBool(key string) bool {
	val, _ := t.GetBool(key)
	return val
}

// JSON marshaling helpers
func (t *Task) ToJSON() ([]byte, error) {
	return json.Marshal(t)
}

func TaskFromJSON(data []byte) (*Task, error) {
	var task Task
	err := json.Unmarshal(data, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}