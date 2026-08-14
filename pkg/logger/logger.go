package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates the application logger.
func New() *zap.Logger {
	config := zap.NewProductionConfig()

	config.Encoding = "json"
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Keep the existing production-style JSON format.
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	log, err := config.Build()
	if err != nil {
		panic(err)
	}

	return log
}

// WithContext returns a logger enriched with the OpenTelemetry
// trace ID and span ID stored in ctx.
//
// If the context does not contain a valid span, the original
// logger is returned unchanged.
func WithContext(
	ctx context.Context,
	log *zap.Logger,
) *zap.Logger {

	if log == nil {
		return nil
	}

	if ctx == nil {
		return log
	}

	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	if !spanContext.IsValid() {
		return log
	}

	return log.With(
		zap.String(
			"trace_id",
			spanContext.TraceID().String(),
		),
		zap.String(
			"span_id",
			spanContext.SpanID().String(),
		),
	)
}

// InfoContext logs an informational message with trace correlation.
func InfoContext(
	ctx context.Context,
	log *zap.Logger,
	msg string,
	fields ...zap.Field,
) {
	WithContext(ctx, log).Info(msg, fields...)
}

// WarnContext logs a warning message with trace correlation.
func WarnContext(
	ctx context.Context,
	log *zap.Logger,
	msg string,
	fields ...zap.Field,
) {
	WithContext(ctx, log).Warn(msg, fields...)
}

// ErrorContext logs an error message with trace correlation.
func ErrorContext(
	ctx context.Context,
	log *zap.Logger,
	msg string,
	fields ...zap.Field,
) {
	WithContext(ctx, log).Error(msg, fields...)
}

// DebugContext logs a debug message with trace correlation.
func DebugContext(
	ctx context.Context,
	log *zap.Logger,
	msg string,
	fields ...zap.Field,
) {
	WithContext(ctx, log).Debug(msg, fields...)
}
