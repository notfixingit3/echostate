package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/prometheus/common/expfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRegistry(t *testing.T) {
	r := New()
	require.NotNil(t, r)
	require.NotNil(t, r.reg)
}

func TestHTTPMetricsIncrement(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := New()
	router := gin.New()
	router.Use(HTTPMiddleware(r))
	router.GET("/test/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test/42", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Counter should be 1 for method=GET, path=/test/:id, status=200
	cnt, err := testutil.GatherAndCount(r.reg, "echostate_http_requests_total")
	require.NoError(t, err)
	assert.Equal(t, 1, cnt, "expected 1 http request metric family")

	// Verify the label values
	expected := `
# HELP echostate_http_requests_total Total number of HTTP requests by method, path, and status code.
# TYPE echostate_http_requests_total counter
echostate_http_requests_total{method="GET",path="/test/:id",status="200"} 1
`
	err = testutil.GatherAndCompare(r.reg, strings.NewReader(expected), "echostate_http_requests_total")
	assert.NoError(t, err, "counter metric should match expected labels and value")
}

func TestHTTPMetricsMultipleRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := New()
	router := gin.New()
	router.Use(HTTPMiddleware(r))
	router.GET("/a", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/a", func(c *gin.Context) { c.Status(http.StatusCreated) })
	router.GET("/b", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })

	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/a", nil))
		assert.Equal(t, http.StatusOK, w.Code)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/a", nil))
	assert.Equal(t, http.StatusCreated, w.Code)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/b", nil))
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	expected := `
# HELP echostate_http_requests_total Total number of HTTP requests by method, path, and status code.
# TYPE echostate_http_requests_total counter
echostate_http_requests_total{method="GET",path="/a",status="200"} 3
echostate_http_requests_total{method="GET",path="/b",status="500"} 1
echostate_http_requests_total{method="POST",path="/a",status="201"} 1
`
	err := testutil.GatherAndCompare(r.reg, strings.NewReader(expected), "echostate_http_requests_total")
	assert.NoError(t, err, "counter should reflect multiple requests with different labels")
}

func TestHTTPRequestDurationHistogram(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := New()
	router := gin.New()
	router.Use(HTTPMiddleware(r))
	router.GET("/slow", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/slow", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the histogram exists and has at least one observation
	cnt, err := testutil.GatherAndCount(r.reg, "echostate_http_request_duration_seconds")
	require.NoError(t, err)
	assert.Equal(t, 1, cnt, "expected 1 histogram metric family")

	// Check that the histogram has a non-zero sample count by scraping the metrics handler
	handler := MetricsHandler(r)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	body, err := io.ReadAll(w2.Result().Body)
	require.NoError(t, err)
	w2.Result().Body.Close()
	output := string(body)
	assert.Contains(t, output, "echostate_http_request_duration_seconds_count")
	assert.Contains(t, output, "echostate_http_request_duration_seconds_sum")
	assert.Contains(t, output, `method="GET"`)
	assert.Contains(t, output, `path="/slow"`)
	assert.Contains(t, output, `status="200"`)
}

func TestWorkerMetrics(t *testing.T) {
	r := New()

	// Record job outcomes
	RecordJobOutcome(r, "scanner", "success")
	RecordJobOutcome(r, "scanner", "success")
	RecordJobOutcome(r, "scanner", "error")
	RecordJobOutcome(r, "reporter", "success")

	expected := `
# HELP echostate_worker_jobs_total Total number of worker jobs processed by worker name and outcome status.
# TYPE echostate_worker_jobs_total counter
echostate_worker_jobs_total{status="error",worker="scanner"} 1
echostate_worker_jobs_total{status="success",worker="reporter"} 1
echostate_worker_jobs_total{status="success",worker="scanner"} 2
`
	err := testutil.GatherAndCompare(r.reg, strings.NewReader(expected), "echostate_worker_jobs_total")
	assert.NoError(t, err, "worker job counter should match expected values")

	// Set queue depth
	SetQueueDepth(r, "scanner", 5)
	SetQueueDepth(r, "reporter", 0)

	expectedDepth := `
# HELP echostate_worker_queue_depth Current number of items queued for a worker.
# TYPE echostate_worker_queue_depth gauge
echostate_worker_queue_depth{worker="reporter"} 0
echostate_worker_queue_depth{worker="scanner"} 5
`
	err = testutil.GatherAndCompare(r.reg, strings.NewReader(expectedDepth), "echostate_worker_queue_depth")
	assert.NoError(t, err, "queue depth gauge should match expected values")

	// Observe queue wait
	ObserveQueueWait(r, "scanner", 0.5)
	ObserveQueueWait(r, "scanner", 1.5)
	ObserveQueueWait(r, "reporter", 0.1)

	handler := MetricsHandler(r)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	w.Result().Body.Close()
	output := string(body)
	assert.Contains(t, output, "echostate_worker_queue_wait_seconds_count")
	assert.Contains(t, output, "echostate_worker_queue_wait_seconds_sum")
	assert.Contains(t, output, `worker="scanner"`)
	assert.Contains(t, output, `worker="reporter"`)
}

func TestMetricsHandlerReturnsPrometheusFormat(t *testing.T) {
	r := New()

	// Record some metrics so the output is non-trivial
	RecordJobOutcome(r, "scanner", "success")
	SetQueueDepth(r, "scanner", 3)
	ObserveQueueWait(r, "scanner", 0.2)

	handler := MetricsHandler(r)
	require.NotNil(t, handler)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	contentType := resp.Header.Get("Content-Type")
	assert.Contains(t, contentType, "text/plain; version=0.0.4")

	output := string(body)

	// Must include Go runtime metrics (from GoCollector)
	assert.Contains(t, output, "go_info")
	assert.Contains(t, output, "go_goroutines")
	assert.Contains(t, output, "go_memstats_alloc_bytes")

	// Must include process metrics (from ProcessCollector)
	assert.Contains(t, output, "process_cpu_seconds_total")
	assert.Contains(t, output, "process_resident_memory_bytes")

	// Must include our application metrics
	assert.Contains(t, output, "echostate_worker_jobs_total")
	assert.Contains(t, output, "echostate_worker_queue_depth")
	assert.Contains(t, output, "echostate_worker_queue_wait_seconds_count")
}

func TestMetricsHandlerEmptyRegistry(t *testing.T) {
	r := New()

	handler := MetricsHandler(r)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body, err := io.ReadAll(w.Result().Body)
	require.NoError(t, err)
	w.Result().Body.Close()

	// Even with no application metrics recorded, Go/process collectors should be present
	output := string(body)
	assert.Contains(t, output, "go_info")
	assert.Contains(t, output, "process_cpu_seconds_total")
}

func TestPromRegistryAccessor(t *testing.T) {
	r := New()
	pr := r.PromRegistry()
	require.NotNil(t, pr)
	assert.Equal(t, r.reg, pr)
}

func TestMustRegister(t *testing.T) {
	r := New()
	// Registering a duplicate should panic
	assert.Panics(t, func() {
		r.MustRegister(r.httpRequestsTotal)
	})
}

func TestHTTPMiddlewarePathLabelUsesFullPath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := New()
	router := gin.New()
	router.Use(HTTPMiddleware(r))
	router.GET("/api/v1/users/:id", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Request with a specific ID — the path label should be the template, not the raw value
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/users/abc-123", nil))
	assert.Equal(t, http.StatusOK, w.Code)

	output, err := testutil.CollectAndFormat(r.httpRequestsTotal, expfmt.TypeTextPlain, "echostate_http_requests_total")
	require.NoError(t, err)

	// Must use the route template, not the raw path
	assert.Contains(t, string(output), `path="/api/v1/users/:id"`)
	assert.NotContains(t, string(output), `path="/api/v1/users/abc-123"`)
}
