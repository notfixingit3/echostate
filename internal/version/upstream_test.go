package version

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetInfo_UpdateAvailableForDevBranch(t *testing.T) {
	ResetUpstreamCache()
	oldVersion := Version
	Version = "0.0.1-beta.11"
	t.Cleanup(func() {
		Version = oldVersion
		ResetUpstreamCache()
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/repos/notfixingit3/echostate/releases", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"tag_name":"v0.0.1-beta.12","prerelease":true,"draft":false},
			{"tag_name":"v0.0.1-beta.10","prerelease":true,"draft":false}
		]`))
	}))
	t.Cleanup(server.Close)

	info := GetInfo(context.Background(), Options{
		Branch:     "dev",
		GitHubRepo: "notfixingit3/echostate",
		Enabled:    true,
		APIBaseURL: server.URL,
		HTTPClient: server.Client(),
	})

	require.Equal(t, "0.0.1-beta.11", info.Current)
	require.Equal(t, "dev", info.Branch)
	require.Equal(t, "0.0.1-beta.12", info.Latest)
	require.True(t, info.UpdateAvailable)
	require.NotEmpty(t, info.CheckedAt)
}

func TestGetInfo_IgnoresStableReleasesOnDevBranch(t *testing.T) {
	ResetUpstreamCache()
	oldVersion := Version
	Version = "0.0.1-beta.12"
	t.Cleanup(func() {
		Version = oldVersion
		ResetUpstreamCache()
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"tag_name":"v1.0.0","prerelease":false,"draft":false},
			{"tag_name":"v0.0.1-beta.12","prerelease":true,"draft":false}
		]`))
	}))
	t.Cleanup(server.Close)

	info := GetInfo(context.Background(), Options{
		Branch:     "dev",
		GitHubRepo: "notfixingit3/echostate",
		Enabled:    true,
		APIBaseURL: server.URL,
		HTTPClient: server.Client(),
	})

	require.Equal(t, "0.0.1-beta.12", info.Latest)
	require.False(t, info.UpdateAvailable)
}

func TestGetInfo_MainBranchUsesStableOnly(t *testing.T) {
	ResetUpstreamCache()
	oldVersion := Version
	Version = "0.0.1"
	t.Cleanup(func() {
		Version = oldVersion
		ResetUpstreamCache()
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"tag_name":"v0.0.2","prerelease":false,"draft":false},
			{"tag_name":"v0.0.1-beta.99","prerelease":true,"draft":false}
		]`))
	}))
	t.Cleanup(server.Close)

	info := GetInfo(context.Background(), Options{
		Branch:     "main",
		GitHubRepo: "notfixingit3/echostate",
		Enabled:    true,
		APIBaseURL: server.URL,
		HTTPClient: server.Client(),
	})

	require.Equal(t, "0.0.2", info.Latest)
	require.True(t, info.UpdateAvailable)
}
