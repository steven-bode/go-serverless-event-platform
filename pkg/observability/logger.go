package observability

import (
	"encoding/json"
	"os"
	"time"
)

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelError LogLevel = "ERROR"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelDebug LogLevel = "DEBUG"
)

type Logger struct {
	correlationID string
	eventID       string
	minLevel      LogLevel
	sampleRate    float64
}

type LogEntry struct {
	Timestamp     string                 `json:"timestamp"`
	Level         LogLevel               `json:"level"`
	Message       string                 `json:"message"`
	CorrelationID string                 `json:"correlation_id,omitempty"`
	EventID       string                 `json:"event_id,omitempty"`
	Fields        map[string]interface{} `json:"fields,omitempty"`
}

func NewLogger(correlationID, eventID string) *Logger {
	return &Logger{
		correlationID: correlationID,
		eventID:       eventID,
		minLevel:      LogLevelInfo,
		sampleRate:    1.0,
	}
}

func NewLoggerWithLevel(correlationID, eventID string, minLevel LogLevel) *Logger {
	return &Logger{
		correlationID: correlationID,
		eventID:       eventID,
		minLevel:      minLevel,
		sampleRate:    1.0,
	}
}

func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}
	return levels[level] >= levels[l.minLevel]
}

func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	if level == LogLevelDebug && l.sampleRate < 1.0 {
		if time.Now().UnixNano()%100 > int64(l.sampleRate*100) {
			return
		}
	}

	entry := LogEntry{
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Level:         level,
		Message:       message,
		CorrelationID: l.correlationID,
		EventID:       l.eventID,
		Fields:        sanitizeFields(fields),
	}

	jsonBytes, _ := json.Marshal(entry)
	os.Stdout.Write(jsonBytes)
	os.Stdout.WriteString("\n")
}

func sanitizeFields(fields map[string]interface{}) map[string]interface{} {
	if fields == nil {
		return nil
	}
	sanitized := make(map[string]interface{})
	for k, v := range fields {
		if k == "payload" || k == "body" || k == "data" {
			sanitized[k] = "[REDACTED]"
			continue
		}
		if str, ok := v.(string); ok && len(str) > 500 {
			sanitized[k] = str[:500] + "..."
			continue
		}
		sanitized[k] = v
	}
	return sanitized
}

func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	mergedFields := mergeFields(fields...)
	l.log(LogLevelInfo, message, mergedFields)
}

func (l *Logger) Error(message string, err error, fields ...map[string]interface{}) {
	mergedFields := mergeFields(fields...)
	if err != nil {
		if mergedFields == nil {
			mergedFields = make(map[string]interface{})
		}
		mergedFields["error"] = err.Error()
	}
	l.log(LogLevelError, message, mergedFields)
}

func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	mergedFields := mergeFields(fields...)
	l.log(LogLevelWarn, message, mergedFields)
}

func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	mergedFields := mergeFields(fields...)
	l.log(LogLevelDebug, message, mergedFields)
}

func mergeFields(fieldsList ...map[string]interface{}) map[string]interface{} {
	if len(fieldsList) == 0 {
		return nil
	}
	result := make(map[string]interface{})
	for _, fields := range fieldsList {
		if fields != nil {
			for k, v := range fields {
				result[k] = v
			}
		}
	}
	return result
}
