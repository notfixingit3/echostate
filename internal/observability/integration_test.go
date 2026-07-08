package observability_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/notfixingit3/echostate/internal/middleware"
	"github.com/notfixingit3/echostate/internal/observability/health"
	obslog "github.com/notfixingit3/echostate/internal/observability/log"
	"github.com/notfixingit3/echostate/internal/observability/metrics"
	obstrace "github.com/notfixingit3/echostate/internal/observability/trace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestFullMiddlewareStackObservabilityEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_TRACES_EXPORTER", "")
	t.Setenv("OTEL_TRACES_SAMPLER", "always_on")
	t.Setenv("OTEL_TRACES_SAMPLER_ARG", "")

	tp, err := obstrace.SetupOTel(context.Background(), nil)
	if err != nil {
		t.Fatalf("setup OTel: %v", err)
	}
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
		otel.SetTracerProvider(noop.NewTracerProvider())
	})

	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	metricsRegistry := metrics.New()
	readyChecker := &integrationChecker{name: "postgres", required: true}

	router := gin.New()
	router.Use(obslog.RequestIDMiddleware(logger))
	router.Use(obslog.RecoveryMiddleware(logger))
	router.Use(otelgin.Middleware("echostate-api"))
	router.Use(obslog.TraceEnricherMiddleware())
	router.Use(obslog.AccessLogMiddleware(logger))
	router.Use(metrics.HTTPMiddleware(metricsRegistry))
	router.Use(middleware.NewCORS("http://localhost:3000", "development"))
	router.GET("/health", health.HealthHandler("test", "test-version"))
	router.GET("/ready", health.ReadyHandler(health.NewRegistry(readyChecker)))

	apiServer := httptest.NewServer(router)
	defer apiServer.Close()
	metricsServer := httptest.NewServer(metrics.MetricsHandler(metricsRegistry))
	defer metricsServer.Close()

	client := apiServer.Client()

	healthResp, healthBody := getJSON(t, client, apiServer.URL+"/health", map[string]string{
		"Origin":       "http://localhost:3000",
		"X-Request-ID": "test-123",
	})
	defer healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("/health status = %d, body=%v", healthResp.StatusCode, healthBody)
	}
	if healthBody["status"] != "ok" {
		t.Fatalf("/health status field = %v", healthBody["status"])
	}
	if got := healthResp.Header.Get("X-Request-ID"); got != "test-123" {
		t.Fatalf("X-Request-ID response header = %q", got)
	}
	if got := healthResp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("CORS allow-origin header = %q", got)
	}

	readyResp, readyBody := getJSON(t, client, apiServer.URL+"/ready", nil)
	defer readyResp.Body.Close()
	if readyResp.StatusCode != http.StatusOK {
		t.Fatalf("healthy /ready status = %d, body=%v", readyResp.StatusCode, readyBody)
	}
	if readyBody["status"] != "ready" {
		t.Fatalf("healthy /ready status field = %v", readyBody["status"])
	}

	readyChecker.setErr(errors.New("database unavailable"))
	unhealthyResp, unhealthyBody := getJSON(t, client, apiServer.URL+"/ready", nil)
	defer unhealthyResp.Body.Close()
	if unhealthyResp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("unhealthy /ready status = %d, body=%v", unhealthyResp.StatusCode, unhealthyBody)
	}
	if unhealthyBody["status"] != "not_ready" {
		t.Fatalf("unhealthy /ready status field = %v", unhealthyBody["status"])
	}

	metricsResp, err := metricsServer.Client().Get(metricsServer.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer metricsResp.Body.Close()
	metricsBody, err := io.ReadAll(metricsResp.Body)
	if err != nil {
		t.Fatalf("read /metrics body: %v", err)
	}
	if metricsResp.StatusCode != http.StatusOK {
		t.Fatalf("/metrics status = %d, body=%s", metricsResp.StatusCode, metricsBody)
	}
	metricsText := string(metricsBody)
	if !strings.Contains(metricsText, "echostate_http_requests_total") {
		t.Fatalf("/metrics missing echostate_http_requests_total:\n%s", metricsText)
	}
	if !strings.Contains(metricsText, "echostate_http_request_duration_seconds") {
		t.Fatalf("/metrics missing echostate_http_request_duration_seconds:\n%s", metricsText)
	}

	entries := decodeAccessLogs(t, logs.String())
	requestLog := findLogEntry(t, entries, "request_id", "test-123")
	assertLogField(t, requestLog, "request_id", "test-123")
	traceID, ok := requestLog["trace_id"].(string)
	if !ok || traceID == "" {
		t.Fatalf("trace_id missing from access log: %#v", requestLog)
	}
}

type integrationChecker struct {
	name     string
	required bool

	mu  sync.Mutex
	err error
}

func (c *integrationChecker) Name() string { return c.name }

func (c *integrationChecker) Required() bool { return c.required }

func (c *integrationChecker) Check(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.err
}

func (c *integrationChecker) setErr(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
}

func getJSON(t *testing.T, client *http.Client, url string, headers map[string]string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create request %s: %v", url, err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		t.Fatalf("read %s body: %v", url, err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close %s body: %v", url, err)
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("decode %s body %q: %v", url, body, err)
	}
	return resp, decoded
}

func decodeAccessLogs(t *testing.T, output string) []map[string]any {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("decode log line %q: %v", line, err)
		}
		if entry["msg"] == "http request" {
			entries = append(entries, entry)
		}
	}
	if len(entries) == 0 {
		t.Fatalf("no access log entries found in %q", output)
	}
	return entries
}

func findLogEntry(t *testing.T, entries []map[string]any, key, value string) map[string]any {
	t.Helper()
	for _, entry := range entries {
		if entry[key] == value {
			return entry
		}
	}
	t.Fatalf("no access log entry with %s=%q in %#v", key, value, entries)
	return nil
}

func assertLogField(t *testing.T, entry map[string]any, key, want string) {
	t.Helper()
	if got, ok := entry[key].(string); !ok || got != want {
		t.Fatalf("log field %s = %#v, want %q", key, entry[key], want)
	}
}
