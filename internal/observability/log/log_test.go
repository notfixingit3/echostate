package log

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func TestInitializeJSONAndText(t *testing.T) {
	jsonOutput := captureStdout(t, func() {
		if err := Initialize("debug", "json"); err != nil {
			t.Fatalf("Initialize json: %v", err)
		}
		slog.Default().Debug("json initialized", slog.String("component", "test"))
	})
	jsonEntry := decodeSingleLogLine(t, jsonOutput)
	if jsonEntry["msg"] != "json initialized" {
		t.Fatalf("json msg = %v", jsonEntry["msg"])
	}
	if jsonEntry["component"] != "test" {
		t.Fatalf("json component = %v", jsonEntry["component"])
	}

	textOutput := captureStdout(t, func() {
		if err := Initialize("info", "text"); err != nil {
			t.Fatalf("Initialize text: %v", err)
		}
		slog.Default().Info("text initialized", slog.String("component", "test"))
	})
	if !strings.Contains(textOutput, "msg=\"text initialized\"") {
		t.Fatalf("text output missing message: %q", textOutput)
	}
	if !strings.Contains(textOutput, "component=test") {
		t.Fatalf("text output missing component: %q", textOutput)
	}
}

func TestSetLevelChangesActiveLevel(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Initialize("info", "json"); err != nil {
			t.Fatalf("Initialize: %v", err)
		}
		if err := SetLevel("debug"); err != nil {
			t.Fatalf("SetLevel debug: %v", err)
		}
		slog.Default().Debug("debug enabled")
		if err := SetLevel("warn"); err != nil {
			t.Fatalf("SetLevel warn: %v", err)
		}
		slog.Default().Info("info suppressed")
		slog.Default().Warn("warn enabled")
	})

	entries := decodeLogLines(t, output)
	if len(entries) != 2 {
		t.Fatalf("expected 2 log entries, got %d from %q", len(entries), output)
	}
	if entries[0]["msg"] != "debug enabled" {
		t.Fatalf("first msg = %v", entries[0]["msg"])
	}
	if entries[1]["msg"] != "warn enabled" {
		t.Fatalf("second msg = %v", entries[1]["msg"])
	}
	if strings.Contains(output, "info suppressed") {
		t.Fatalf("info log was not suppressed: %q", output)
	}
}

func TestFatalExitsWithStatusOne(t *testing.T) {
	if os.Getenv("ECHOSTATE_LOG_FATAL_SUBPROCESS") == "1" {
		if err := Initialize("debug", "json"); err != nil {
			panic(err)
		}
		Fatal("fatal test", slog.String("key", "value"))
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatalExitsWithStatusOne")
	cmd.Env = append(os.Environ(), "ECHOSTATE_LOG_FATAL_SUBPROCESS=1")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("Fatal subprocess unexpectedly succeeded; output=%s", output)
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Fatalf("exit code = %d, output=%s", exitErr.ExitCode(), output)
	}
	entry := decodeSingleLogLine(t, string(output))
	if entry["msg"] != "fatal test" || entry["key"] != "value" {
		t.Fatalf("unexpected fatal log entry: %#v", entry)
	}
}

func TestRecoveryMiddlewareLogsPanicAndReturns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	router := gin.New()
	router.Use(RecoveryMiddleware(logger))
	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body=%s", response.Code, response.Body.String())
	}
	entry := decodeSingleLogLine(t, buf.String())
	if entry["msg"] != "panic recovered" {
		t.Fatalf("msg = %v", entry["msg"])
	}
	if entry["panic"] != "boom" {
		t.Fatalf("panic field = %v", entry["panic"])
	}
	stackTrace, ok := entry["stack_trace"].(string)
	if !ok || !strings.Contains(stackTrace, "TestRecoveryMiddlewareLogsPanicAndReturns500") {
		t.Fatalf("stack_trace missing test frame: %#v", entry["stack_trace"])
	}
	if entry["method"] != http.MethodGet || entry["path"] != "/panic" {
		t.Fatalf("request fields missing: %#v", entry)
	}
}

func TestAccessLogMiddlewareLogsRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	router := gin.New()
	router.Use(RequestIDMiddleware(logger))
	router.Use(AccessLogMiddleware(logger))
	router.POST("/access", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/access", nil)
	request.Header.Set("X-Request-ID", "req-123")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	entry := decodeSingleLogLine(t, buf.String())
	assertLogString(t, entry, "request_id", "req-123")
	assertLogString(t, entry, "trace_id", "")
	assertLogString(t, entry, "method", http.MethodPost)
	assertLogString(t, entry, "path", "/access")
	assertLogNumber(t, entry, "status", http.StatusNoContent)
	assertLogString(t, entry, "client_ip", "192.0.2.1")
	if _, ok := entry["latency_ms"].(float64); !ok {
		t.Fatalf("latency_ms missing or not numeric: %#v", entry["latency_ms"])
	}
}

func TestFromContextEnrichedByRequestIDAndTraceMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	traceID := trace.TraceID{0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe, 0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe}
	spanID := trace.SpanID{0x10, 0x32, 0x54, 0x76, 0x98, 0xba, 0xdc, 0xfe}

	router := gin.New()
	router.Use(RequestIDMiddleware(logger))
	router.Use(func(c *gin.Context) {
		spanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID:    traceID,
			SpanID:     spanID,
			TraceFlags: trace.FlagsSampled,
		})
		c.Request = c.Request.WithContext(trace.ContextWithSpanContext(c.Request.Context(), spanContext))
		c.Next()
	})
	router.Use(slogTraceEnricherMiddleware())
	router.GET("/context", func(c *gin.Context) {
		FromContext(c.Request.Context()).Info("handler log")
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/context", nil)
	request.Header.Set("X-Request-ID", "req-context")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	entry := decodeSingleLogLine(t, buf.String())
	assertLogString(t, entry, "request_id", "req-context")
	assertLogString(t, entry, "trace_id", traceID.String())
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = writer

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = oldStdout
	t.Cleanup(func() { os.Stdout = oldStdout })
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	return string(output)
}

func decodeSingleLogLine(t *testing.T, output string) map[string]any {
	t.Helper()
	entries := decodeLogLines(t, output)
	if len(entries) != 1 {
		t.Fatalf("expected 1 log entry, got %d from %q", len(entries), output)
	}
	return entries[0]
}

func decodeLogLines(t *testing.T, output string) []map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("decode log line %q: %v", line, err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func assertLogString(t *testing.T, entry map[string]any, key, want string) {
	t.Helper()
	got, ok := entry[key].(string)
	if !ok || got != want {
		t.Fatalf("%s = %#v, want %q", key, entry[key], want)
	}
}

func assertLogNumber(t *testing.T, entry map[string]any, key string, want int) {
	t.Helper()
	got, ok := entry[key].(float64)
	if !ok || int(got) != want {
		t.Fatalf("%s = %#v, want %d", key, entry[key], want)
	}
}
