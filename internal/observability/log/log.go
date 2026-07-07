// Package log provides slog setup and request-scoped Gin logging middleware.
package log

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

type contextKey struct{}

var loggerKey contextKey

var (
	levelMu  sync.RWMutex
	levelVar = new(slog.LevelVar)
)

// Initialize configures the process-wide slog default logger.
func Initialize(level, format string) error {
	parsedLevel, err := parseLevel(level)
	if err != nil {
		return err
	}

	nextLevel := new(slog.LevelVar)
	nextLevel.Set(parsedLevel)
	opts := &slog.HandlerOptions{Level: nextLevel}

	var handler slog.Handler
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json", "":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		return fmt.Errorf("unsupported log format %q", format)
	}

	levelMu.Lock()
	levelVar = nextLevel
	levelMu.Unlock()

	slog.SetDefault(slog.New(handler))
	return nil
}

// SetLevel updates the active slog handler level configured by Initialize.
func SetLevel(level string) error {
	parsedLevel, err := parseLevel(level)
	if err != nil {
		return err
	}

	levelMu.RLock()
	current := levelVar
	levelMu.RUnlock()
	current.Set(parsedLevel)
	return nil
}

// Fatal logs msg at error level and terminates the process with status code 1.
func Fatal(msg string, attrs ...slog.Attr) {
	slog.Default().LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
	os.Exit(1)
}

// WithLogger stores logger in ctx for request-scoped logging.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	if logger == nil {
		return ctx
	}
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns a request-scoped logger from ctx or slog.Default().
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	if !ok || logger == nil {
		return slog.Default()
	}
	return logger
}

// RequestIDMiddleware installs gin-contrib/requestid and stores request_id on
// the request-scoped context logger via requestid.WithHandler.
func RequestIDMiddleware(logger *slog.Logger) gin.HandlerFunc {
	base := loggerOrDefault(logger)
	return requestid.New(requestid.WithHandler(func(c *gin.Context, requestID string) {
		ctx := c.Request.Context()
		requestLogger, ok := loggerFromContext(ctx)
		if !ok {
			requestLogger = base
		}
		ctx = WithLogger(ctx, requestLogger.With(slog.String("request_id", requestID)))
		c.Request = c.Request.WithContext(ctx)
	}))
}

// TraceEnricherMiddleware adds the active OpenTelemetry trace_id to the
// request-scoped logger. It should run after otelgin has attached a span.
func TraceEnricherMiddleware() gin.HandlerFunc {
	return slogTraceEnricherMiddleware()
}

// RecoveryMiddleware recovers panics, logs a structured stack trace, and
// returns HTTP 500 without leaking request bodies or sensitive headers.
func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	base := loggerOrDefault(logger)
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				base.LogAttrs(c.Request.Context(), slog.LevelError, "panic recovered",
					slog.Any("panic", recovered),
					slog.String("stack_trace", string(debug.Stack())),
					slog.String("request_id", requestid.Get(c)),
					slog.String("trace_id", traceIDFromContext(c.Request.Context())),
					slog.String("method", c.Request.Method),
					slog.String("path", requestPath(c)),
					slog.String("client_ip", c.ClientIP()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()

		c.Next()
	}
}

// AccessLogMiddleware logs one structured record after downstream handlers run.
func AccessLogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	base := loggerOrDefault(logger)
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		base.LogAttrs(c.Request.Context(), slog.LevelInfo, "http request",
			slog.String("request_id", requestid.Get(c)),
			slog.String("trace_id", traceIDFromContext(c.Request.Context())),
			slog.String("method", c.Request.Method),
			slog.String("path", requestPath(c)),
			slog.Int("status", c.Writer.Status()),
			slog.Float64("latency_ms", float64(time.Since(start).Microseconds())/1000.0),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}

func slogTraceEnricherMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		traceID := traceIDFromContext(ctx)
		if traceID != "" {
			ctx = WithLogger(ctx, FromContext(ctx).With(slog.String("trace_id", traceID)))
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}

func loggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	if ctx == nil {
		return nil, false
	}
	logger, ok := ctx.Value(loggerKey).(*slog.Logger)
	return logger, ok && logger != nil
}

func loggerOrDefault(logger *slog.Logger) *slog.Logger {
	if logger != nil {
		return logger
	}
	return slog.Default()
}

func parseLevel(level string) (slog.Level, error) {
	value := strings.ToUpper(strings.TrimSpace(level))
	if value == "" {
		value = "INFO"
	}
	if value == "WARNING" {
		value = "WARN"
	}

	var parsed slog.Level
	if err := parsed.UnmarshalText([]byte(value)); err != nil {
		return 0, fmt.Errorf("unsupported log level %q", level)
	}
	return parsed, nil
}

func requestPath(c *gin.Context) string {
	if path := c.FullPath(); path != "" {
		return path
	}
	if c.Request != nil && c.Request.URL != nil {
		return c.Request.URL.Path
	}
	return ""
}

func traceIDFromContext(ctx context.Context) string {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() {
		return ""
	}
	return spanContext.TraceID().String()
}
