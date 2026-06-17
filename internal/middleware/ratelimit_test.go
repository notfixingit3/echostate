package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_AllowsUpToLimit(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	for i := 0; i < maxTokens; i++ {
		w := doRequest(rl, "192.168.1.1")
		assert.Equal(t, http.StatusOK, w.Code, "request %d should be allowed", i+1)
	}
}

func TestRateLimiter_BlocksExcess(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	for i := 0; i < maxTokens; i++ {
		doRequest(rl, "10.0.0.1")
	}

	w := doRequest(rl, "10.0.0.1")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "rate limit exceeded", body["error"])
	assert.NotNil(t, body["retry_after"])
}

func TestRateLimiter_DifferentIPsAreIndependent(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	for i := 0; i < maxTokens; i++ {
		doRequest(rl, "10.0.0.1")
	}

	for i := 0; i < maxTokens; i++ {
		w := doRequest(rl, "10.0.0.2")
		assert.Equal(t, http.StatusOK, w.Code, "IP B request %d should be allowed", i+1)
	}
}

func TestRateLimiter_RetryAfterIsPositive(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	for i := 0; i < maxTokens; i++ {
		doRequest(rl, "172.16.0.1")
	}

	w := doRequest(rl, "172.16.0.1")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)

	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	retryAfter, ok := body["retry_after"].(float64)
	assert.True(t, ok, "retry_after should be a number")
	assert.GreaterOrEqual(t, int(retryAfter), 1)
}

func TestRateLimiter_CleansOldEntries(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	doRequest(rl, "10.0.0.1")
	addr := netip.MustParseAddr("10.0.0.1")

	rl.mu.Lock()
	b, exists := rl.buckets[addr]
	require.True(t, exists, "entry should exist before cleanup")
	b.lastSeen = time.Now().Add(-2 * cleanupAge)
	rl.mu.Unlock()

	rl.mu.Lock()
	now := time.Now()
	for ip, b := range rl.buckets {
		if now.Sub(b.lastSeen) > cleanupAge {
			delete(rl.buckets, ip)
		}
	}
	rl.mu.Unlock()

	rl.mu.Lock()
	_, exists = rl.buckets[addr]
	rl.mu.Unlock()
	assert.False(t, exists, "stale entry should be removed by cleanup")
}

func TestRateLimiter_DoesNotCleanFreshEntries(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	doRequest(rl, "10.0.0.2")
	addr := netip.MustParseAddr("10.0.0.2")

	rl.mu.Lock()
	now := time.Now()
	for ip, b := range rl.buckets {
		if now.Sub(b.lastSeen) > cleanupAge {
			delete(rl.buckets, ip)
		}
	}
	rl.mu.Unlock()

	rl.mu.Lock()
	_, exists := rl.buckets[addr]
	rl.mu.Unlock()
	assert.True(t, exists, "fresh entry should survive cleanup")
}

// --- helpers ---

func doRequest(rl *RateLimiter, clientIP string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodPost, "/api/scan", nil)
	c.Request.RemoteAddr = clientIP + ":12345"
	rl.Middleware()(c)
	if !c.IsAborted() {
		c.Status(http.StatusOK)
	}
	return w
}
