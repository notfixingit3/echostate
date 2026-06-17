package middleware

import (
	"net/http"
	"net/netip"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	maxTokens   = 30
	refillRate  = 30.0          // tokens per minute
	refillEvery = 2 * time.Second // refill 1 token every 2 seconds
	cleanupAge  = 10 * time.Minute
	cleanupTick = 1 * time.Minute
)

type bucket struct {
	tokens    float64
	lastSeen  time.Time
	updatedAt time.Time
}

// RateLimiter is an in-memory token-bucket rate limiter keyed by client IP.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[netip.Addr]*bucket
	stopCh  chan struct{}
}

// NewRateLimiter creates a started rate limiter. Call Stop to release resources.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[netip.Addr]*bucket),
		stopCh:  make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// Middleware returns a Gin handler that rate-limits requests per client IP.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipStr := c.ClientIP()
		addr, err := netip.ParseAddr(ipStr)
		if err != nil {
			c.Next()
			return
		}

		now := time.Now()
		rl.mu.Lock()
		b, exists := rl.buckets[addr]
		if !exists {
			b = &bucket{tokens: maxTokens, lastSeen: now, updatedAt: now}
			rl.buckets[addr] = b
		}

		// Refill based on elapsed time since last refill.
		elapsed := now.Sub(b.updatedAt)
		b.tokens += elapsed.Seconds() * (refillRate / 60.0)
		if b.tokens > maxTokens {
			b.tokens = maxTokens
		}
		b.updatedAt = now
		b.lastSeen = now

		if b.tokens < 1 {
			rl.mu.Unlock()
			retryAfter := int((1 - b.tokens) * 60.0 / refillRate)
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": retryAfter,
			})
			return
		}

		b.tokens--
		rl.mu.Unlock()
		c.Next()
	}
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(cleanupTick)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stopCh:
			return
		case now := <-ticker.C:
			rl.mu.Lock()
			for ip, b := range rl.buckets {
				if now.Sub(b.lastSeen) > cleanupAge {
					delete(rl.buckets, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}
