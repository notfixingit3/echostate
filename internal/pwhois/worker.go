package pwhois

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/notfixingit3/echostate/internal/db"
)

const (
	defaultCacheTTL = 24 * time.Hour
	pollInterval    = 5 * time.Second
	maxRetries      = 3
	lookupTimeout   = 30 * time.Second
)

// LookupFunc performs a pwhois lookup for a batch of IP addresses.
type LookupFunc func(ctx context.Context, ips []string) ([]PWHOISRecord, error)

// Worker polls snapshots and enriches them with pwhois data in the background.
type Worker struct {
	db     *db.DB
	lookup LookupFunc
	cache  *cache

	stop chan struct{}
	wg   sync.WaitGroup

	mu      sync.Mutex
	started bool
}

// NewWorker creates a Worker. If lookup is nil, the package-level Lookup is used.
// cacheTTL controls how long pwhois results are cached in memory; if zero, defaultCacheTTL is used.
func NewWorker(database *db.DB, lookup LookupFunc, cacheTTL time.Duration) *Worker {
	if lookup == nil {
		lookup = Lookup
	}
	if cacheTTL <= 0 {
		cacheTTL = defaultCacheTTL
	}
	return &Worker{
		db:     database,
		lookup: lookup,
		cache:  newCache(cacheTTL),
		stop:   make(chan struct{}),
	}
}

// Start begins the background polling loop.
func (w *Worker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.started {
		return fmt.Errorf("pwhois worker already started")
	}

	w.started = true
	w.stop = make(chan struct{})

	w.wg.Add(1)
	go w.loop(ctx)

	return nil
}

// Stop gracefully shuts down the worker.
func (w *Worker) Stop() {
	w.mu.Lock()
	if !w.started {
		w.mu.Unlock()
		return
	}
	close(w.stop)
	w.started = false
	w.mu.Unlock()

	w.wg.Wait()
}

func (w *Worker) loop(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	w.process(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			w.process(ctx)
		}
	}
}

func (w *Worker) process(ctx context.Context) {
	ips, err := w.pendingIPs(ctx)
	if err != nil {
		slog.Default().Error("pwhois worker: fetch pending IPs", slog.String("component", "pwhois_worker"), slog.String("error", err.Error()))
		return
	}
	if len(ips) == 0 {
		return
	}

	batch := make([]string, 0, maxBatchSize)
	for _, ip := range ips {
		if w.cache.get(ip) != nil {
			continue
		}
		if !isRoutable(ip) {
			w.cache.set(ip, nil)
			if err := w.updateSnapshot(ctx, ip, nil); err != nil {
				slog.Default().Error("pwhois worker: mark skipped IP", slog.String("component", "pwhois_worker"), slog.String("ip", ip), slog.String("error", err.Error()))
			}
			continue
		}
		batch = append(batch, ip)
		if len(batch) >= maxBatchSize {
			w.lookupBatch(ctx, batch)
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		w.lookupBatch(ctx, batch)
	}
}

func (w *Worker) pendingIPs(ctx context.Context) ([]string, error) {
	rows, err := w.db.Pool.Query(ctx, `
		SELECT DISTINCT client_ip
		FROM snapshots
		WHERE client_ip IS NOT NULL AND client_ip != ''
		  AND (pwhois_looked_up_at IS NULL OR pwhois_looked_up_at < NOW() - INTERVAL '24 hours')
		LIMIT $1
	`, maxBatchSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ips, nil
}

func (w *Worker) lookupBatch(ctx context.Context, ips []string) {
	var records []PWHOISRecord
	var err error

	backoff := time.Second
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lookupCtx, cancel := context.WithTimeout(ctx, lookupTimeout)
		records, err = w.lookup(lookupCtx, ips)
		cancel()
		if err == nil {
			break
		}
		slog.Default().Warn("pwhois worker: lookup attempt failed", slog.String("component", "pwhois_worker"), slog.Int("attempt", attempt+1), slog.String("error", err.Error()))
		if attempt < maxRetries {
			select {
			case <-ctx.Done():
				return
			case <-w.stop:
				return
			case <-time.After(backoff):
				backoff *= 2
			}
		}
	}
	if err != nil {
		slog.Default().Error("pwhois worker: lookup failed after all attempts", slog.String("component", "pwhois_worker"), slog.Int("attempts", maxRetries+1), slog.String("error", err.Error()))
		return
	}

	byIP := make(map[string]PWHOISRecord, len(records))
	for i := range records {
		byIP[records[i].IP] = records[i]
		w.cache.set(records[i].IP, &records[i])
	}

	for _, ip := range ips {
		rec, ok := byIP[ip]
		if !ok {
			w.cache.set(ip, nil)
			if err := w.updateSnapshot(ctx, ip, nil); err != nil {
				slog.Default().Error("pwhois worker: update snapshot for missing IP", slog.String("component", "pwhois_worker"), slog.String("ip", ip), slog.String("error", err.Error()))
			}
			continue
		}
		if err := w.updateSnapshot(ctx, ip, &rec); err != nil {
			slog.Default().Error("pwhois worker: update snapshot", slog.String("component", "pwhois_worker"), slog.String("ip", ip), slog.String("error", err.Error()))
		}
	}
}

func (w *Worker) updateSnapshot(ctx context.Context, ip string, rec *PWHOISRecord) error {
	var data json.RawMessage
	var originAS, orgName, countryCode, city, prefix *string
	if rec != nil {
		b, err := json.Marshal(rec)
		if err != nil {
			return fmt.Errorf("marshal record: %w", err)
		}
		data = b
		originAS = nullable(rec.OriginAS)
		orgName = nullable(rec.OrgName)
		countryCode = nullable(rec.CountryCode)
		city = nullable(rec.City)
		prefix = nullable(rec.Prefix)
	} else {
		data = []byte("null")
	}

	_, err := w.db.Pool.Exec(ctx, `
		UPDATE snapshots
		SET pwhois_data = $1,
		    pwhois_looked_up_at = NOW(),
		    pwhois_origin_as = $2,
		    pwhois_org_name = $3,
		    pwhois_country_code = $4,
		    pwhois_city = $5,
		    pwhois_prefix = $6
		WHERE client_ip = $7
		  AND (pwhois_looked_up_at IS NULL OR pwhois_looked_up_at < NOW() - INTERVAL '24 hours')
	`, data, originAS, orgName, countryCode, city, prefix, ip)
	return err
}

type cache struct {
	ttl     time.Duration
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	record  *PWHOISRecord
	expires time.Time
}

func newCache(ttl time.Duration) *cache {
	return &cache{
		ttl:     ttl,
		entries: make(map[string]cacheEntry),
	}
}

func (c *cache) get(ip string) *PWHOISRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()

	e, ok := c.entries[ip]
	if !ok || time.Now().After(e.expires) {
		return nil
	}
	return e.record
}

func (c *cache) set(ip string, rec *PWHOISRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[ip] = cacheEntry{
		record:  rec,
		expires: time.Now().Add(c.ttl),
	}
}

func isRoutable(ip string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// Skip loopback (127.0.0.0/8, ::1/128), private/unique-local (RFC1918, fc00::/7),
	// and link-local (169.254.0.0/16, fe80::/10) addresses.
	if parsed.IsLoopback() || parsed.IsPrivate() || parsed.IsLinkLocalUnicast() {
		return false
	}
	return true
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
