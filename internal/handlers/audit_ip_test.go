package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/notfixingit3/echostate/internal/pwhois"
)

func TestShouldEnrichIP(t *testing.T) {
	require.False(t, shouldEnrichIP(""))
	require.False(t, shouldEnrichIP("not-an-ip"))
	require.False(t, shouldEnrichIP("127.0.0.1"))
	require.False(t, shouldEnrichIP("10.0.0.4"))
	require.False(t, shouldEnrichIP("192.168.1.10"))
	require.True(t, shouldEnrichIP("203.0.113.10"))
}

func TestIPInfoFromRecord(t *testing.T) {
	info := ipInfoFromRecord(pwhois.PWHOISRecord{
		OriginAS:    "AS15169",
		OrgName:     "Google LLC",
		CountryCode: "US",
		City:        "Mountain View",
		Prefix:      "203.0.113.0/24",
	})
	require.Equal(t, "AS15169", info["origin_as"])
	require.Equal(t, "Google LLC", info["org_name"])
	require.Equal(t, "US", info["country_code"])
	require.Equal(t, "Mountain View", info["city"])
	require.Equal(t, "203.0.113.0/24", info["prefix"])
}

func TestMergeIPInfo(t *testing.T) {
	merged := mergeIPInfo(map[string]any{"host": "example.com"}, map[string]any{
		"org_name": "Example ISP",
	})
	require.Equal(t, "example.com", merged["host"])
	require.Equal(t, "Example ISP", merged["ip_info"].(map[string]any)["org_name"])
}

func TestResolveClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("uses client ip", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.RemoteAddr = "203.0.113.10:1234"
		require.Equal(t, "203.0.113.10", resolveClientIP(c))
	})

	t.Run("falls back to x-forwarded-for", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-Forwarded-For", "198.51.100.20, 10.0.0.1")
		require.Equal(t, "198.51.100.20", resolveClientIP(c))
	})
}
