package types

import (
	"context"
	"fmt"
)

// TaskData represents the data passed to a task
type TaskData map[string]interface{}

// TaskHandler is the function signature for handling tasks
type TaskHandler func(ctx context.Context, data TaskData) error

// String returns a string value from TaskData
func (td TaskData) String(key string) (string, bool) {
	val, exists := td[key]
	if !exists {
		return "", false
	}

	str, ok := val.(string)
	return str, ok
}

// Int returns an int value from TaskData
func (td TaskData) Int(key string) (int, bool) {
	val, exists := td[key]
	if !exists {
		return 0, false
	}

	switch v := val.(type) {
	case int:
		return v, true
	case float64: // JSON numbers are float64
		return int(v), true
	case int64:
		return int(v), true
	default:
		return 0, false
	}
}

// Bool returns a bool value from TaskData
func (td TaskData) Bool(key string) (bool, bool) {
	val, exists := td[key]
	if !exists {
		return false, false
	}

	b, ok := val.(bool)
	return b, ok
}

// Get returns the raw value from TaskData
func (td TaskData) Get(key string) (interface{}, bool) {
	val, exists := td[key]
	return val, exists
}

// MustString returns a string value or panics
func (td TaskData) MustString(key string) string {
	val, ok := td.String(key)
	if !ok {
		panic(fmt.Sprintf("required string key '%s' not found or invalid type", key))
	}
	return val
}

// MustInt returns an int value or panics
func (td TaskData) MustInt(key string) int {
	val, ok := td.Int(key)
	if !ok {
		panic(fmt.Sprintf("required int key '%s' not found or invalid type", key))
	}
	return val
}

// MustBool returns a bool value or panics
func (td TaskData) MustBool(key string) bool {
	val, ok := td.Bool(key)
	if !ok {
		panic(fmt.Sprintf("required bool key '%s' not found or invalid type", key))
	}
	return val
}
