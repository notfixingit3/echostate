package main

import (
	"context"
	stdlog "log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notfixingit3/echostate/internal/cli"
	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/handlers"
	"github.com/notfixingit3/echostate/internal/middleware"
	obshealth "github.com/notfixingit3/echostate/internal/observability/health"
	obslog "github.com/notfixingit3/echostate/internal/observability/log"
	"github.com/notfixingit3/echostate/internal/observability/metrics"
	obstrace "github.com/notfixingit3/echostate/internal/observability/trace"
	"github.com/notfixingit3/echostate/internal/pwhois"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		if err := cli.RunAuth(os.Args[2:]); err != nil {
			stdlog.Fatalf("auth command failed: %v", err)
		}
		return
	}
	cfg, err := config.Load()
	if err != nil {
		stdlog.Fatalf("failed to load config: %v", err)
	}

	if err := obslog.Initialize(cfg.LogLevel, cfg.LogFormat); err != nil {
		stdlog.Fatalf("failed to initialize logger: %v", err)
	}
	logger := slog.Default()

	tracerProvider, err := obstrace.SetupOTel(context.Background(), cfg)
	if err != nil {
		obslog.Fatal("failed to initialize OpenTelemetry", slog.Any("error", err))
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		obslog.Fatal("failed to connect to database", slog.Any("error", err))
	}

	if err := db.Migrate(database); err != nil {
		obslog.Fatal("failed to run migrations", slog.Any("error", err))
	}

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	var neo4jClient *db.Neo4jClient
	if cfg.Neo4jURI != "" {
		neo4jClient, err = db.NewNeo4jClient(cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword)
		if err != nil {
			logger.Warn("failed to connect to neo4j", slog.Any("error", err))
		} else {
			defer neo4jClient.Close(context.Background())
		}
	}

	var pwhoisWorker *pwhois.Worker
	if cfg.PwhoisEnabled {
		pwhoisWorker = pwhois.NewWorker(database, nil, time.Duration(cfg.PwhoisCacheTTLHours)*time.Hour)
		if err := pwhoisWorker.Start(workerCtx); err != nil {
			obslog.Fatal("failed to start pwhois worker", slog.Any("error", err))
		}
	}

	metricsRegistry := metrics.New()
	metricsSrv := newMetricsServer(cfg, metrics.MetricsHandler(metricsRegistry))
	healthRegistry := obshealth.NewRegistry(
		obshealth.NewPostgresChecker(database.Pool),
		obshealth.NewBrowserlessChecker(browserlessHealthURL(cfg.BrowserWSURL)),
	)

	router := gin.New()
	router.SetTrustedProxies([]string{
		"172.16.0.0/12",
		"10.0.0.0/8",
		"192.168.0.0/16",
		"127.0.0.1",
	})
	router.Use(obslog.RequestIDMiddleware(logger))
	router.Use(obslog.RecoveryMiddleware(logger))
	router.Use(otelgin.Middleware("echostate-api"))
	router.Use(obslog.TraceEnricherMiddleware())
	router.Use(obslog.AccessLogMiddleware(logger))
	router.Use(metrics.HTTPMiddleware(metricsRegistry))
	router.Use(middleware.NewCORS(cfg.FrontendURL, cfg.Env))

	rateLimiter := middleware.NewRateLimiter()
	defer rateLimiter.Stop()

	workers := handlers.Register(router, database, neo4jClient, cfg, rateLimiter, metricsRegistry, healthRegistry)
	if err := workers.Reports.Start(workerCtx); err != nil {
		obslog.Fatal("failed to start report worker", slog.Any("error", err))
	}
	if err := workers.Scans.Start(workerCtx); err != nil {
		obslog.Fatal("failed to start scan worker", slog.Any("error", err))
	}
	workers.Scheduler.Start(workerCtx)
	workers.AuditPurger.Start(workerCtx)

	srv := newServer(cfg, router)

	go func() {
		logger.Info("EchoState API listening", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			obslog.Fatal("server error", slog.Any("error", err))
		}
	}()

	go func() {
		logger.Info("EchoState metrics listening", slog.String("addr", metricsSrv.Addr))
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			obslog.Fatal("metrics server error", slog.Any("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down EchoState")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := tracerProvider.Shutdown(ctx); err != nil {
		logger.Error("failed to shutdown OpenTelemetry", slog.Any("error", err))
	}
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced API shutdown", slog.Any("error", err))
	}
	if err := metricsSrv.Shutdown(ctx); err != nil {
		logger.Error("forced metrics shutdown", slog.Any("error", err))
	}

	workers.Scheduler.Stop()
	workers.AuditPurger.Stop()
	workers.Scans.Stop()
	workers.Reports.Stop()

	if pwhoisWorker != nil {
		pwhoisWorker.Stop()
	}

	database.Close()

	logger.Info("shutdown complete")
}

func newServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func newMetricsServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + strconv.Itoa(cfg.MetricsPort),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func browserlessHealthURL(browserWSURL string) string {
	u, err := url.Parse(strings.TrimSpace(browserWSURL))
	if err != nil || u.Host == "" {
		return strings.TrimRight(browserWSURL, "/")
	}
	switch u.Scheme {
	case "ws":
		u.Scheme = "http"
	case "wss":
		u.Scheme = "https"
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/")
}
