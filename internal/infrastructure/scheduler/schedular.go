package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// TaskData represents flexible task data
type TaskData map[string]interface{}

// Task represents a scheduled task
type Task struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Data      TaskData  `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// TaskHandler is a function that processes tasks
type TaskHandler func(ctx context.Context, data TaskData) error

// TaskOptions for configuring scheduled tasks
type TaskOptions struct {
	ID    string
	Data  TaskData
	Delay time.Duration
}

// Backend interface remains the same but simplified
type Backend interface {
	Store(task Task, delay time.Duration) error
	Cancel(taskID string) error
	StartListening(onExpire func(task Task)) error
	Stop() error
	Ping() error
}

// Scheduler provides a flexible task scheduling system
type Scheduler struct {
	backend   Backend
	handlers  map[string]TaskHandler
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	idCounter int64
	idMu      sync.Mutex
}

// New creates a new scheduler
func New(backend Backend) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		backend:  backend,
		handlers: make(map[string]TaskHandler),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start initializes the scheduler
func (s *Scheduler) Start() error {
	return s.backend.StartListening(s.executeTask)
}

// Stop shuts down the scheduler
func (s *Scheduler) Stop() error {
	s.cancel()
	return s.backend.Stop()
}

// On registers a handler for a task type (like addEventListener)
func (s *Scheduler) On(taskType string, handler TaskHandler) *Scheduler {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[taskType] = handler
	return s
}

// Off removes a handler for a task type
func (s *Scheduler) Off(taskType string) *Scheduler {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.handlers, taskType)
	return s
}

// After schedules a task to run after a delay
func (s *Scheduler) After(delay time.Duration) *TaskBuilder {
	return &TaskBuilder{
		scheduler: s,
		delay:     delay,
		data:      make(TaskData),
	}
}

// In is an alias for After (more natural for some use cases)
func (s *Scheduler) In(delay time.Duration) *TaskBuilder {
	return s.After(delay)
}

// At schedules a task to run at a specific time
func (s *Scheduler) At(when time.Time) *TaskBuilder {
	delay := time.Until(when)
	if delay < 0 {
		delay = 0
	}
	return s.After(delay)
}

// Cancel cancels a scheduled task
func (s *Scheduler) Cancel(taskID string) error {
	return s.backend.Cancel(taskID)
}

// Health checks scheduler health
func (s *Scheduler) Health() error {
	return s.backend.Ping()
}

// TaskBuilder provides a fluent interface for building tasks
type TaskBuilder struct {
	scheduler *Scheduler
	delay     time.Duration
	data      TaskData
	taskType  string
	id        string
}

// Type sets the task type
func (tb *TaskBuilder) Type(taskType string) *TaskBuilder {
	tb.taskType = taskType
	return tb
}

// ID sets a custom task ID
func (tb *TaskBuilder) ID(id string) *TaskBuilder {
	tb.id = id
	return tb
}

// With adds data to the task
func (tb *TaskBuilder) With(key string, value interface{}) *TaskBuilder {
	tb.data[key] = value
	return tb
}

// WithData sets multiple data fields at once
func (tb *TaskBuilder) WithData(data TaskData) *TaskBuilder {
	for k, v := range data {
		tb.data[k] = v
	}
	return tb
}

// Do executes a task with a specific handler (JS-like callback)
func (tb *TaskBuilder) Do(handler TaskHandler) (string, error) {
	taskType := tb.taskType
	if taskType == "" {
		taskType = fmt.Sprintf("anonymous_%d", tb.scheduler.generateID())
	}

	// Register the handler for this specific task type
	tb.scheduler.On(taskType, handler)

	return tb.scheduler.scheduleTask(taskType, tb.data, tb.delay, tb.id)
}

// Emit schedules a task of the specified type (assumes handler is already registered)
func (tb *TaskBuilder) Emit(taskType string) (string, error) {
	return tb.scheduler.scheduleTask(taskType, tb.data, tb.delay, tb.id)
}

// Execute is like Do but for when you want to schedule with pre-registered handlers
func (tb *TaskBuilder) Execute(taskType string) (string, error) {
	return tb.Emit(taskType)
}

// scheduleTask internal method to schedule a task
func (s *Scheduler) scheduleTask(taskType string, data TaskData, delay time.Duration, customID string) (string, error) {
	id := customID
	if id == "" {
		id = fmt.Sprintf("%s_%d_%d", taskType, time.Now().UnixNano(), s.generateID())
	}

	task := Task{
		ID:        id,
		Type:      taskType,
		Data:      data,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(delay),
	}

	err := s.backend.Store(task, delay)
	if err != nil {
		return "", err
	}

	return id, nil
}

// generateID generates a unique ID
func (s *Scheduler) generateID() int64 {
	s.idMu.Lock()
	defer s.idMu.Unlock()
	s.idCounter++
	return s.idCounter
}

// executeTask executes a task by calling the appropriate handler
func (s *Scheduler) executeTask(task Task) {
	s.mu.RLock()
	handler, exists := s.handlers[task.Type]
	s.mu.RUnlock()

	if !exists {
		return // No handler registered
	}

	// Execute handler with context and error handling
	go func() {
		ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
		defer cancel()

		if err := handler(ctx, task.Data); err != nil {
			// Log error - task execution failed
			fmt.Printf("Task execution failed: %v\n", err)
		}
	}()
}

// Utility methods for TaskData
func (td TaskData) String(key string) (string, bool) {
	val, exists := td[key]
	if !exists {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

func (td TaskData) Int(key string) (int, bool) {
	val, exists := td[key]
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}

func (td TaskData) Bool(key string) (bool, bool) {
	val, exists := td[key]
	if !exists {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

func (td TaskData) Get(key string) (interface{}, bool) {
	val, exists := td[key]
	return val, exists
}

// Must variants (panic on missing/wrong type)
func (td TaskData) MustString(key string) string {
	val, ok := td.String(key)
	if !ok {
		panic(fmt.Sprintf("key %s not found or not string", key))
	}
	return val
}

func (td TaskData) MustInt(key string) int {
	val, ok := td.Int(key)
	if !ok {
		panic(fmt.Sprintf("key %s not found or not int", key))
	}
	return val
}

func (td TaskData) MustBool(key string) bool {
	val, ok := td.Bool(key)
	if !ok {
		panic(fmt.Sprintf("key %s not found or not bool", key))
	}
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

// Convenience functions for common patterns

// Every creates a recurring task (you'd need to re-schedule in the handler)
func (s *Scheduler) Every(interval time.Duration) *TaskBuilder {
	return s.After(interval)
}

// Once schedules a one-time task with inline handler
func (s *Scheduler) Once(delay time.Duration, handler TaskHandler) (string, error) {
	return s.After(delay).Do(handler)
}

// Debounce creates a debounced task (cancels previous if exists)
func (s *Scheduler) Debounce(taskID string, delay time.Duration) *TaskBuilder {
	// Cancel existing task if it exists
	s.Cancel(taskID)
	return s.After(delay).ID(taskID)
}
