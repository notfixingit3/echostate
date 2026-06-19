package middleware

import (
	"net/http"
	"net/netip"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notfixingit3/echostate/internal/config"
)

const (
	maxTokens        = 30.0
	enrollMaxTokens  = 10.0
	enrollRefillRate = 10.0 / 15.0 // 10 attempts per 15 minutes
	cleanupAge       = 10 * time.Minute
	cleanupTick      = 1 * time.Minute
)

type bucket struct {
	tokens    float64
	lastSeen  time.Time
	updatedAt time.Time
}

// RateLimiter is an in-memory token-bucket rate limiter keyed by client IP.
type RateLimiter struct {
	mu            sync.Mutex
	buckets       map[netip.Addr]*bucket
	enrollBuckets map[netip.Addr]*bucket
	stopCh        chan struct{}
}

// NewRateLimiter creates a started rate limiter. Call Stop to release resources.
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets:       make(map[netip.Addr]*bucket),
		enrollBuckets: make(map[netip.Addr]*bucket),
		stopCh:        make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

// Stop terminates the background cleanup goroutine.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// EnrollVerifyMiddleware returns a stricter per-IP limiter for enrollment verification.
func (rl *RateLimiter) EnrollVerifyMiddleware() gin.HandlerFunc {
	return rl.middlewareFor(rl.enrollBuckets, enrollMaxTokens, enrollRefillRate)
}

// Middleware returns a Gin handler that rate-limits requests per client IP.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return rl.middlewareFor(rl.buckets, maxTokens, 0)
}

func (rl *RateLimiter) middlewareFor(store map[netip.Addr]*bucket, cap float64, fixedRate float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		ipStr := c.ClientIP()
		addr, err := netip.ParseAddr(ipStr)
		if err != nil {
			c.Next()
			return
		}

		now := time.Now()
		rl.mu.Lock()
		b, exists := store[addr]
		if !exists {
			b = &bucket{tokens: cap, lastSeen: now, updatedAt: now}
			store[addr] = b
		}

		rate := fixedRate
		if rate <= 0 {
			settings := config.GetSettings()
			rate = settings.RateLimit
			if rate <= 0 {
				rate = 30.0
			}
			rate = rate / 60.0
		} else {
			rate = rate / 60.0
		}

		elapsed := now.Sub(b.updatedAt)
		b.tokens += elapsed.Seconds() * rate
		if b.tokens > cap {
			b.tokens = cap
		}
		b.updatedAt = now
		b.lastSeen = now

		if b.tokens < 1 {
			rl.mu.Unlock()
			retryAfter := int((1 - b.tokens) / rate)
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
			for ip, b := range rl.enrollBuckets {
				if now.Sub(b.lastSeen) > cleanupAge {
					delete(rl.enrollBuckets, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}
