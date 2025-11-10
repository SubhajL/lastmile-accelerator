package logger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	serviceName string
	environment string
	writer      io.Writer
	fields      []Field
}

type Field struct {
	Key   string
	Value interface{}
}

func New(serviceName, env string, writer io.Writer) *Logger {
	if writer == nil {
		writer = os.Stdout
	}
	return &Logger{
		serviceName: serviceName,
		environment: env,
		writer:      writer,
		fields:      []Field{},
	}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	newLogger := &Logger{
		serviceName: l.serviceName,
		environment: l.environment,
		writer:      l.writer,
		fields:      make([]Field, len(l.fields)),
	}
	copy(newLogger.fields, l.fields)

	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		newLogger.fields = append(newLogger.fields, Field{
			Key:   "trace_id",
			Value: spanCtx.TraceID().String(),
		})
		newLogger.fields = append(newLogger.fields, Field{
			Key:   "span_id",
			Value: spanCtx.SpanID().String(),
		})
	}

	return newLogger
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.log("info", msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.log("error", msg, fields...)
}

func (l *Logger) Debug(msg string, fields ...Field) {
	l.log("debug", msg, fields...)
}

func (l *Logger) log(level, msg string, fields ...Field) {
	entry := map[string]interface{}{
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
		"level":        level,
		"message":      msg,
		"service.name": l.serviceName,
		"environment":  l.environment,
	}

	// Add logger-level fields
	for _, f := range l.fields {
		entry[f.Key] = f.Value
	}

	// Add call-specific fields
	for _, f := range fields {
		entry[f.Key] = f.Value
	}

	data, _ := json.Marshal(entry)
	l.writer.Write(data)
	l.writer.Write([]byte("\n"))
}
