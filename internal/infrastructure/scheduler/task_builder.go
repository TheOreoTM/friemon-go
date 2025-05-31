package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/theoreotm/friemon/internal/types"
)

// TaskBuilder provides a fluent API for building and scheduling tasks
type TaskBuilder struct {
	scheduler *Scheduler
	delay     time.Duration
	data      types.TaskData
	taskType  string
	id        string
	queue     string
	maxRetry  int
	retention time.Duration
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

// With adds a key-value pair to the task data
func (tb *TaskBuilder) With(key string, value interface{}) *TaskBuilder {
	if tb.data == nil {
		tb.data = make(types.TaskData)
	}
	tb.data[key] = value
	return tb
}

// WithData sets the entire task data
func (tb *TaskBuilder) WithData(data types.TaskData) *TaskBuilder {
	tb.data = data
	return tb
}

// Queue sets the queue name for the task
func (tb *TaskBuilder) Queue(queue string) *TaskBuilder {
	tb.queue = queue
	return tb
}

// MaxRetry sets the maximum number of retries
func (tb *TaskBuilder) MaxRetry(retries int) *TaskBuilder {
	tb.maxRetry = retries
	return tb
}

// Retention sets how long to keep the task in Redis after completion
func (tb *TaskBuilder) Retention(duration time.Duration) *TaskBuilder {
	tb.retention = duration
	return tb
}

// Emit schedules the task with the specified type
func (tb *TaskBuilder) Emit(taskType string) (string, error) {
	tb.taskType = taskType
	return tb.execute()
}

// Execute schedules the task (alias for Emit)
func (tb *TaskBuilder) Execute(taskType string) (string, error) {
	return tb.Emit(taskType)
}

// Do schedules the task and registers the handler inline
func (tb *TaskBuilder) Do(handler types.TaskHandler) (string, error) {
	if tb.taskType == "" {
		return "", fmt.Errorf("task type is required")
	}

	// Register the handler
	tb.scheduler.On(tb.taskType, handler)

	return tb.execute()
}

// execute creates and enqueues the task
func (tb *TaskBuilder) execute() (string, error) {
	if tb.taskType == "" {
		return "", fmt.Errorf("task type is required")
	}

	// Marshal task data
	payload, err := json.Marshal(tb.data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task data: %w", err)
	}

	// Create Asynq task
	task := asynq.NewTask(tb.taskType, payload)

	// Prepare task options
	opts := []asynq.Option{
		asynq.MaxRetry(tb.maxRetry),
		asynq.Queue(tb.queue),
	}

	// Add delay if specified
	if tb.delay > 0 {
		opts = append(opts, asynq.ProcessIn(tb.delay))
	}

	// Add custom ID if specified
	if tb.id != "" {
		opts = append(opts, asynq.TaskID(tb.id))
	}

	// Add retention if specified
	if tb.retention > 0 {
		opts = append(opts, asynq.Retention(tb.retention))
	}

	// Enqueue the task
	info, err := tb.scheduler.client.Enqueue(task, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to enqueue task: %w", err)
	}

	tb.scheduler.logger.Info("Task scheduled",
		"task_type", tb.taskType,
		"task_id", info.ID,
		"queue", info.Queue,
		"delay", tb.delay,
	)

	return info.ID, nil
}
