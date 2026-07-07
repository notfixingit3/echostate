// Package health provides a Checker interface and HTTP handlers for
// application health and readiness probes. /health (liveness) reports
// the process is alive without checking dependencies. /ready (readiness)
// runs registered checkers with a per-check timeout and fails closed when
// any required dependency is unhealthy.
package health

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Checker defines a single dependency check for the /ready endpoint.
type Checker interface {
	// Name returns a human-readable label for the check (e.g. "postgres").
	Name() string
	// Required indicates whether a failure of this check should cause /ready
	// to return HTTP 503. Optional checkers that fail are still reported in
	// the response body but do not flip the overall status to unhealthy.
	Required() bool
	// Check performs the dependency probe. A non-nil error is treated as
	// unhealthy. Implementations must respect context cancellation and
	// deadlines.
	Check(ctx context.Context) error
}

// Registry holds a set of Checkers and is safe for concurrent reads after
// construction. Register should be called before the first /ready request.
type Registry struct {
	checkers []Checker
}

// NewRegistry returns a Registry pre-populated with the given checkers.
func NewRegistry(defaultCheckers ...Checker) *Registry {
	cp := make([]Checker, len(defaultCheckers))
	copy(cp, defaultCheckers)
	return &Registry{checkers: cp}
}

// Register appends a checker. It is not safe for concurrent writes.
func (r *Registry) Register(c Checker) {
	r.checkers = append(r.checkers, c)
}

// Checkers returns a snapshot of the registered checkers.
func (r *Registry) Checkers() []Checker {
	cp := make([]Checker, len(r.checkers))
	copy(cp, r.checkers)
	return cp
}

// --- Concrete checkers -------------------------------------------------------

// PostgresChecker verifies PostgreSQL connectivity by running SELECT 1.
// It is a required checker.
type PostgresChecker struct {
	pool *pgxpool.Pool
}

// NewPostgresChecker creates a PostgresChecker that uses the given pool.
func NewPostgresChecker(pool *pgxpool.Pool) *PostgresChecker {
	return &PostgresChecker{pool: pool}
}

func (c *PostgresChecker) Name() string { return "postgres" }

func (c *PostgresChecker) Required() bool { return true }

func (c *PostgresChecker) Check(ctx context.Context) error {
	var result int
	return c.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
}

// BrowserlessChecker verifies that the browserless/chrome container is
// reachable by hitting its /pressure endpoint. It is an optional checker.
type BrowserlessChecker struct {
	baseURL string
	client  *http.Client
}

// NewBrowserlessChecker creates a BrowserlessChecker that probes the given
// base URL (e.g. "http://localhost:3000").
func NewBrowserlessChecker(baseURL string) *BrowserlessChecker {
	return &BrowserlessChecker{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *BrowserlessChecker) Name() string { return "browserless" }

func (c *BrowserlessChecker) Required() bool { return false }

func (c *BrowserlessChecker) Check(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/pressure", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

// Neo4jChecker verifies that the Neo4j database is reachable by opening a
// session and running a trivial query. It is an optional checker.
type Neo4jChecker struct {
	driver neo4j.DriverWithContext
}

// NewNeo4jChecker creates a Neo4jChecker that uses the given driver.
func NewNeo4jChecker(driver neo4j.DriverWithContext) *Neo4jChecker {
	return &Neo4jChecker{driver: driver}
}

func (c *Neo4jChecker) Name() string { return "neo4j" }

func (c *Neo4jChecker) Required() bool { return false }

func (c *Neo4jChecker) Check(ctx context.Context) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	_, err := session.Run(ctx, "RETURN 1", nil)
	if err != nil {
		return fmt.Errorf("neo4j query: %w", err)
	}
	return nil
}

// --- HTTP handlers -----------------------------------------------------------

// checkerResult is the JSON body for a single check in the /ready response.
type checkerResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

// HealthHandler returns a Gin handler that responds with a simple liveness
// probe. It never checks dependencies.
func HealthHandler(env, version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"env":     env,
			"version": version,
		})
	}
}

// ReadyHandler returns a Gin handler that runs every registered checker with
// a 5-second timeout. If all required checkers pass, it returns HTTP 200.
// If any required checker fails, it returns HTTP 503. Optional checkers that
// fail are listed in the response body but do not cause a 503.
func ReadyHandler(registry *Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		results := make([]checkerResult, 0, len(registry.checkers))
		overallOK := true

		for _, chk := range registry.checkers {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			err := chk.Check(ctx)
			cancel()

			status := "ok"
			if err != nil {
				status = "unhealthy"
				if chk.Required() {
					overallOK = false
				}
			}
			results = append(results, checkerResult{Name: chk.Name(), Status: status})
		}

		statusCode := http.StatusOK
		status := "ready"
		if !overallOK {
			statusCode = http.StatusServiceUnavailable
			status = "not_ready"
		}

		c.JSON(statusCode, gin.H{
			"status": status,
			"checks": results,
		})
	}
}
