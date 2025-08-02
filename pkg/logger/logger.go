package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents log severity levels
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns string representation of log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a structured logger
type Logger struct {
	level  LogLevel
	format string
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
}

var defaultLogger *Logger

// Init initializes the default logger
func Init(level, format string) {
	defaultLogger = &Logger{
		level:  parseLogLevel(level),
		format: format,
	}
}

// parseLogLevel converts string to LogLevel
func parseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

// Debug logs a debug message
func Debug(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(DEBUG, message, fields...)
	}
}

// Info logs an info message
func Info(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(INFO, message, fields...)
	}
}

// Warn logs a warning message
func Warn(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(WARN, message, fields...)
	}
}

// Error logs an error message
func Error(message string, fields ...map[string]interface{}) {
	if defaultLogger != nil {
		defaultLogger.log(ERROR, message, fields...)
	}
}

// log performs the actual logging
func (l *Logger) log(level LogLevel, message string, fields ...map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level.String(),
		Message:   message,
	}

	// Add fields if provided
	if len(fields) > 0 && fields[0] != nil {
		entry.Fields = fields[0]
	}

	// Add caller information
	if _, file, line, ok := runtime.Caller(3); ok {
		entry.Caller = fmt.Sprintf("%s:%d", file, line)
	}

	// Output based on format
	switch l.format {
	case "json":
		l.outputJSON(entry)
	default:
		l.outputText(entry)
	}
}

// outputJSON outputs log entry as JSON
func (l *Logger) outputJSON(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}
	fmt.Fprintln(os.Stdout, string(data))
}

// outputText outputs log entry as formatted text
func (l *Logger) outputText(entry LogEntry) {
	output := fmt.Sprintf("[%s] %s: %s", entry.Timestamp, entry.Level, entry.Message)
	
	if entry.Fields != nil && len(entry.Fields) > 0 {
		fieldsStr := ""
		for k, v := range entry.Fields {
			if fieldsStr != "" {
				fieldsStr += " "
			}
			fieldsStr += fmt.Sprintf("%s=%v", k, v)
		}
		output += fmt.Sprintf(" | %s", fieldsStr)
	}
	
	if entry.Caller != "" {
		output += fmt.Sprintf(" [%s]", entry.Caller)
	}
	
	fmt.Fprintln(os.Stdout, output)
}