// Package metrics provides a custom Prometheus registry with HTTP and worker
// instrumentation for the EchoState application. It uses an injectable registry
// (not the global default) so tests do not pollute production state.
package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry wraps a custom prometheus.Registry and the metric descriptors so
// callers can record observations without importing prometheus directly.
type Registry struct {
	reg *prometheus.Registry

	httpRequestsTotal        *prometheus.CounterVec
	httpRequestDuration      *prometheus.HistogramVec
	workerJobsTotal          *prometheus.CounterVec
	workerQueueDepth         *prometheus.GaugeVec
	workerQueueWaitSeconds   *prometheus.HistogramVec
}

// New creates a new Registry, registers Go runtime and process collectors, and
// registers the application-level metric descriptors.
func New() *Registry {
	reg := prometheus.NewRegistry()

	// Standard Go runtime and process metrics.
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	httpRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "echostate_http_requests_total",
			Help: "Total number of HTTP requests by method, path, and status code.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "echostate_http_request_duration_seconds",
			Help:    "HTTP request latency in seconds by method, path, and status code.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	workerJobsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "echostate_worker_jobs_total",
			Help: "Total number of worker jobs processed by worker name and outcome status.",
		},
		[]string{"worker", "status"},
	)

	workerQueueDepth := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "echostate_worker_queue_depth",
			Help: "Current number of items queued for a worker.",
		},
		[]string{"worker"},
	)

	workerQueueWaitSeconds := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "echostate_worker_queue_wait_seconds",
			Help:    "Time items spend waiting in the worker queue before processing.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"worker"},
	)

	reg.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		workerJobsTotal,
		workerQueueDepth,
		workerQueueWaitSeconds,
	)

	return &Registry{
		reg:                  reg,
		httpRequestsTotal:    httpRequestsTotal,
		httpRequestDuration:  httpRequestDuration,
		workerJobsTotal:      workerJobsTotal,
		workerQueueDepth:     workerQueueDepth,
		workerQueueWaitSeconds: workerQueueWaitSeconds,
	}
}

// HTTPMiddleware returns a Gin middleware that records request count and
// duration. The path label uses c.FullPath() (the route template) to avoid
// high cardinality from raw paths.
func HTTPMiddleware(r *Registry) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		method := c.Request.Method
		path := c.FullPath()
		status := strconv.Itoa(c.Writer.Status())

		r.httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		r.httpRequestDuration.WithLabelValues(method, path, status).Observe(
			time.Since(start).Seconds(),
		)
	}
}

// MetricsHandler returns an http.Handler that serves Prometheus exposition
// format from the custom registry.
func MetricsHandler(r *Registry) http.Handler {
	return promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{
		EnableOpenMetrics: false,
	})
}

// RecordJobOutcome increments the worker jobs counter for the given worker and
// outcome status (e.g. "success", "error").
func RecordJobOutcome(r *Registry, worker, status string) {
	r.workerJobsTotal.WithLabelValues(worker, status).Inc()
}

// SetQueueDepth sets the current queue depth for a worker.
func SetQueueDepth(r *Registry, worker string, depth int64) {
	r.workerQueueDepth.WithLabelValues(worker).Set(float64(depth))
}

// ObserveQueueWait records the time an item spent waiting in the worker queue.
func ObserveQueueWait(r *Registry, worker string, seconds float64) {
	r.workerQueueWaitSeconds.WithLabelValues(worker).Observe(seconds)
}

// PromRegistry exposes the underlying prometheus.Registry for advanced use
// cases (e.g. registering additional collectors outside this package).
func (r *Registry) PromRegistry() *prometheus.Registry {
	return r.reg
}

// MustRegister is a convenience wrapper around the underlying registry.
func (r *Registry) MustRegister(cs ...prometheus.Collector) {
	r.reg.MustRegister(cs...)
}
