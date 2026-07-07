package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock checker helpers ----------------------------------------------------

type mockChecker struct {
	name     string
	required bool
	delay    time.Duration
	err      error
	mu       sync.Mutex
	called   bool
}

func (m *mockChecker) Name() string { return m.name }

func (m *mockChecker) Required() bool { return m.required }

func (m *mockChecker) Check(ctx context.Context) error {
	m.mu.Lock()
	m.called = true
	m.mu.Unlock()

	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return m.err
}

func (m *mockChecker) Called() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.called
}

// --- HealthHandler tests -----------------------------------------------------

func TestHealthHandler_Returns200WithStatusAndMeta(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	HealthHandler("test-env", "v1.0.0")(c)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "test-env", body["env"])
	assert.Equal(t, "v1.0.0", body["version"])
}

func TestHealthHandler_DoesNotCheckDependencies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/health", nil)

	// Even with no registry, /health should work.
	HealthHandler("dev", "dev")(c)
	assert.Equal(t, http.StatusOK, rr.Code)
}

// --- ReadyHandler tests ------------------------------------------------------

func TestReadyHandler_AllRequiredPass_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reg := NewRegistry(
		&mockChecker{name: "postgres", required: true},
		&mockChecker{name: "browserless", required: false},
	)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)

	ReadyHandler(reg)(c)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body struct {
		Status string          `json:"status"`
		Checks []checkerResult `json:"checks"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ready", body.Status)
	require.Len(t, body.Checks, 2)
	assert.Equal(t, "postgres", body.Checks[0].Name)
	assert.Equal(t, "ok", body.Checks[0].Status)
	assert.Equal(t, "browserless", body.Checks[1].Name)
	assert.Equal(t, "ok", body.Checks[1].Status)
}

func TestReadyHandler_RequiredCheckerFails_Returns503(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reg := NewRegistry(
		&mockChecker{name: "postgres", required: true, err: assert.AnError},
		&mockChecker{name: "browserless", required: false},
	)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)

	ReadyHandler(reg)(c)

	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var body struct {
		Status string          `json:"status"`
		Checks []checkerResult `json:"checks"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "not_ready", body.Status)
	require.Len(t, body.Checks, 2)
	assert.Equal(t, "postgres", body.Checks[0].Name)
	assert.Equal(t, "unhealthy", body.Checks[0].Status)
	assert.Equal(t, "browserless", body.Checks[1].Name)
	assert.Equal(t, "ok", body.Checks[1].Status)
}

func TestReadyHandler_OnlyOptionalCheckersFail_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reg := NewRegistry(
		&mockChecker{name: "postgres", required: true},
		&mockChecker{name: "browserless", required: false, err: assert.AnError},
		&mockChecker{name: "neo4j", required: false, err: assert.AnError},
	)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)

	ReadyHandler(reg)(c)

	// Required checker passed, so overall is 200 even though optional ones failed.
	assert.Equal(t, http.StatusOK, rr.Code)

	var body struct {
		Status string          `json:"status"`
		Checks []checkerResult `json:"checks"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ready", body.Status)
	require.Len(t, body.Checks, 3)
	assert.Equal(t, "postgres", body.Checks[0].Name)
	assert.Equal(t, "ok", body.Checks[0].Status)
	assert.Equal(t, "browserless", body.Checks[1].Name)
	assert.Equal(t, "unhealthy", body.Checks[1].Status)
	assert.Equal(t, "neo4j", body.Checks[2].Name)
	assert.Equal(t, "unhealthy", body.Checks[2].Status)
}

func TestReadyHandler_EmptyRegistry_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reg := NewRegistry()

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)

	ReadyHandler(reg)(c)

	assert.Equal(t, http.StatusOK, rr.Code)

	var body struct {
		Status string          `json:"status"`
		Checks []checkerResult `json:"checks"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ready", body.Status)
	assert.Empty(t, body.Checks)
}

// --- Timeout test ------------------------------------------------------------

func TestReadyHandler_CheckerTimeout_ReportsUnhealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// A required checker that sleeps longer than the 5s timeout.
	slow := &mockChecker{name: "slow-db", required: true, delay: 10 * time.Second}

	reg := NewRegistry(slow)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)

	start := time.Now()
	ReadyHandler(reg)(c)
	elapsed := time.Since(start)

	// Must complete well before the 10s sleep (timeout is 5s).
	assert.Less(t, elapsed, 8*time.Second, "should timeout before the checker sleep completes")
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

	var body struct {
		Status string          `json:"status"`
		Checks []checkerResult `json:"checks"`
	}
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "not_ready", body.Status)
	require.Len(t, body.Checks, 1)
	assert.Equal(t, "slow-db", body.Checks[0].Name)
	assert.Equal(t, "unhealthy", body.Checks[0].Status)
}

// --- Registry tests ----------------------------------------------------------

func TestRegistry_RegisterAndCheckers(t *testing.T) {
	reg := NewRegistry()
	assert.Empty(t, reg.Checkers())

	c1 := &mockChecker{name: "a", required: true}
	c2 := &mockChecker{name: "b", required: false}
	reg.Register(c1)
	reg.Register(c2)

	checkers := reg.Checkers()
	assert.Len(t, checkers, 2)
	assert.Equal(t, "a", checkers[0].Name())
	assert.Equal(t, "b", checkers[1].Name())
}

func TestNewRegistry_WithDefaults(t *testing.T) {
	c1 := &mockChecker{name: "a", required: true}
	c2 := &mockChecker{name: "b", required: false}
	reg := NewRegistry(c1, c2)

	checkers := reg.Checkers()
	assert.Len(t, checkers, 2)
}

func TestRegistry_CheckersReturnsCopy(t *testing.T) {
	c1 := &mockChecker{name: "a", required: true}
	reg := NewRegistry(c1)

	// Mutate the returned slice — should not affect the registry.
	checkers := reg.Checkers()
	checkers[0] = &mockChecker{name: "mutated", required: false}

	assert.Equal(t, "a", reg.Checkers()[0].Name())
}

// --- Concrete checker unit tests ---------------------------------------------

func TestPostgresChecker_Metadata(t *testing.T) {
	// We cannot easily test Check() without a real pool, but we can verify
	// the metadata methods and that the constructor does not panic.
	c := NewPostgresChecker(nil)
	assert.Equal(t, "postgres", c.Name())
	assert.True(t, c.Required())
}

func TestBrowserlessChecker_Metadata(t *testing.T) {
	c := NewBrowserlessChecker("http://localhost:3000")
	assert.Equal(t, "browserless", c.Name())
	assert.False(t, c.Required())
}

func TestBrowserlessChecker_Check_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/pressure", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewBrowserlessChecker(srv.URL)
	err := c.Check(context.Background())
	assert.NoError(t, err)
}

func TestBrowserlessChecker_Check_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("overloaded"))
	}))
	defer srv.Close()

	c := NewBrowserlessChecker(srv.URL)
	err := c.Check(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "503")
}

func TestBrowserlessChecker_Check_ConnectionRefused(t *testing.T) {
	// Point at a port that nothing is listening on.
	c := NewBrowserlessChecker("http://127.0.0.1:1")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Check(ctx)
	assert.Error(t, err)
}

func TestNeo4jChecker_Metadata(t *testing.T) {
	c := NewNeo4jChecker(nil)
	assert.Equal(t, "neo4j", c.Name())
	assert.False(t, c.Required())
}

// --- Context cancellation / timeout propagation -----------------------------

func TestChecker_RespectsContextCancellation(t *testing.T) {
	// Verify that a checker that blocks forever gets cancelled by the
	// ReadyHandler's 5s timeout.
	gin.SetMode(gin.TestMode)

	blocker := &mockChecker{name: "blocker", required: true, delay: 1 * time.Hour}
	reg := NewRegistry(blocker)

	rr := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rr)
	c.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)

	start := time.Now()
	ReadyHandler(reg)(c)
	elapsed := time.Since(start)

	assert.Less(t, elapsed, 10*time.Second, "should not wait for the full blocker sleep")
	assert.Equal(t, http.StatusServiceUnavailable, rr.Code)
}
