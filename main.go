package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notfixingit3/echostate/internal/config"
	"github.com/notfixingit3/echostate/internal/db"
	"github.com/notfixingit3/echostate/internal/handlers"
	"github.com/notfixingit3/echostate/internal/middleware"
	"github.com/notfixingit3/echostate/internal/pwhois"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	workerCtx, cancelWorker := context.WithCancel(context.Background())
	defer cancelWorker()

	var neo4jClient *db.Neo4jClient
	if cfg.Neo4jURI != "" {
		neo4jClient, err = db.NewNeo4jClient(cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword)
		if err != nil {
			log.Printf("failed to connect to neo4j: %v", err)
		} else {
			defer neo4jClient.Close(context.Background())
		}
	}

	var pwhoisWorker *pwhois.Worker
	if cfg.PwhoisEnabled {
		pwhoisWorker = pwhois.NewWorker(database, nil, time.Duration(cfg.PwhoisCacheTTLHours)*time.Hour)
		if err := pwhoisWorker.Start(workerCtx); err != nil {
			log.Fatalf("failed to start pwhois worker: %v", err)
		}
	}

	router := gin.New()
	router.SetTrustedProxies([]string{
		"172.16.0.0/12",
		"10.0.0.0/8",
		"192.168.0.0/16",
		"127.0.0.1",
	})
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware())
	router.Use(middleware.NewCORS(cfg.FrontendURL))

	rateLimiter := middleware.NewRateLimiter()
	defer rateLimiter.Stop()

	workers := handlers.Register(router, database, neo4jClient, cfg, rateLimiter)
	if err := workers.Reports.Start(workerCtx); err != nil {
		log.Fatalf("failed to start report worker: %v", err)
	}
	if err := workers.Scans.Start(workerCtx); err != nil {
		log.Fatalf("failed to start scan worker: %v", err)
	}
	workers.Scheduler.Start(workerCtx)

	srv := newServer(cfg, router)

	go func() {
		log.Printf("EchoState API listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down EchoState API")

	workers.Scheduler.Stop()
	workers.Scans.Stop()
	workers.Reports.Stop()

	if pwhoisWorker != nil {
		pwhoisWorker.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}

	database.Close()

	log.Println("shutdown complete")
}

func newServer(cfg *config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return param.TimeStamp.Format(time.RFC3339) + " " + param.Method + " " + param.Path + " " + strconv.Itoa(param.StatusCode) + " " + param.Latency.String() + "\n"
	})
}
