package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/mafzaidi/stackforge/internal/domain/service"
)

// Logger provides application logging with structured output
// Implements domain/service.Logger interface
type Logger struct {
	*log.Logger
}

// Fields represents structured log fields
type Fields = service.Fields

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Fields    Fields `json:"fields,omitempty"`
}

func New() *Logger {
	return &Logger{Logger: log.New(os.Stdout, "", 0)}
}

// Info logs an informational message with structured fields
func (l *Logger) Info(message string, fields Fields) {
	l.log("INFO", message, fields)
}

// Warn logs a warning message with structured fields
func (l *Logger) Warn(message string, fields Fields) {
	l.log("WARN", message, fields)
}

// Error logs an error message with structured fields
func (l *Logger) Error(message string, fields Fields) {
	l.log("ERROR", message, fields)
}

// log formats and outputs a structured log entry
func (l *Logger) log(level, message string, fields Fields) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Fields:    fields,
	}

	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.Logger.Printf("[%s] %s %v (marshal error: %v)", level, message, fields, err)
		return
	}

	l.Logger.Println(string(jsonBytes))
}

// TruncateToken returns the last 4 characters of a token for safe logging
// Returns empty string if token is too short
func TruncateToken(token string) string {
	if len(token) < 4 {
		return ""
	}
	return "..." + token[len(token)-4:]
}
