package logger

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	logger := New()
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.Logger)
}

func TestInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: log.New(&buf, "", 0)}

	logger.Info("test message", Fields{
		"key1": "value1",
		"key2": 123,
	})

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "key1")
	assert.Contains(t, output, "value1")

	// Verify it's valid JSON
	var entry LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "INFO", entry.Level)
	assert.Equal(t, "test message", entry.Message)
}

func TestWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: log.New(&buf, "", 0)}

	logger.Warn("warning message", Fields{
		"error": "something went wrong",
	})

	output := buf.String()
	assert.Contains(t, output, "warning message")
	assert.Contains(t, output, "WARN")
	assert.Contains(t, output, "error")

	// Verify it's valid JSON
	var entry LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "WARN", entry.Level)
}

func TestError(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: log.New(&buf, "", 0)}

	logger.Error("error message", Fields{
		"error": "critical failure",
	})

	output := buf.String()
	assert.Contains(t, output, "error message")
	assert.Contains(t, output, "ERROR")

	// Verify it's valid JSON
	var entry LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	assert.NoError(t, err)
	assert.Equal(t, "ERROR", entry.Level)
}

func TestTruncateToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "normal token",
			token:    "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.signature",
			expected: "...ture",
		},
		{
			name:     "short token",
			token:    "abc",
			expected: "",
		},
		{
			name:     "exactly 4 chars",
			token:    "abcd",
			expected: "...abcd",
		},
		{
			name:     "empty token",
			token:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateToken(tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLogWithEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: log.New(&buf, "", 0)}

	logger.Info("message without fields", Fields{})

	output := buf.String()
	assert.Contains(t, output, "message without fields")
	assert.Contains(t, output, "INFO")

	// Verify it's valid JSON
	var entry LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	assert.NoError(t, err)
}

func TestLogWithNilFields(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: log.New(&buf, "", 0)}

	logger.Info("message with nil fields", nil)

	output := buf.String()
	assert.Contains(t, output, "message with nil fields")
	assert.Contains(t, output, "INFO")

	// Verify it's valid JSON
	var entry LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	assert.NoError(t, err)
}

func TestLogTimestampFormat(t *testing.T) {
	var buf bytes.Buffer
	logger := &Logger{Logger: log.New(&buf, "", 0)}

	logger.Info("test timestamp", Fields{})

	output := buf.String()
	var entry LogEntry
	err := json.Unmarshal([]byte(output), &entry)
	assert.NoError(t, err)

	// Verify timestamp is in RFC3339 format
	assert.True(t, strings.Contains(entry.Timestamp, "T"))
	assert.True(t, strings.Contains(entry.Timestamp, "Z"))
}
