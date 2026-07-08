package version

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const upstreamCacheTTL = time.Hour

// Info describes the running build and optional upstream release metadata.
type Info struct {
	Current         string `json:"current"`
	Branch          string `json:"branch"`
	Latest          string `json:"latest,omitempty"`
	UpdateAvailable bool   `json:"update_available"`
	CheckedAt       string `json:"checked_at,omitempty"`
	CheckError      string `json:"check_error,omitempty"`
}

// Options configures upstream release lookups.
type Options struct {
	Branch     string
	GitHubRepo string
	Enabled    bool
	APIBaseURL string
	HTTPClient *http.Client
}

type releaseCache struct {
	latest    string
	checkedAt time.Time
	err       string
}

var (
	upstreamMu    sync.Mutex
	upstreamCache = make(map[string]releaseCache)
)

type githubRelease struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

// GetInfo returns version metadata, optionally checking GitHub releases for the
// configured branch.
func GetInfo(ctx context.Context, cfg Options) Info {
	current := Normalize(Version)
	info := Info{
		Current: current,
		Branch:  cfg.Branch,
	}

	if !cfg.Enabled || cfg.GitHubRepo == "" {
		return info
	}
	if current == "" || current == "dev" {
		info.CheckError = "local build"
		return info
	}

	latest, checkedAt, checkErr := latestForBranch(ctx, cfg)
	info.CheckedAt = checkedAt.UTC().Format(time.RFC3339)
	if checkErr != "" {
		info.CheckError = checkErr
		return info
	}
	if latest == "" {
		return info
	}

	info.Latest = latest
	info.UpdateAvailable = Compare(current, latest) < 0
	return info
}

func latestForBranch(ctx context.Context, cfg Options) (string, time.Time, string) {
	cacheKey := cfg.GitHubRepo + ":" + cfg.Branch

	upstreamMu.Lock()
	if cached, ok := upstreamCache[cacheKey]; ok && time.Since(cached.checkedAt) < upstreamCacheTTL {
		upstreamMu.Unlock()
		return cached.latest, cached.checkedAt, cached.err
	}
	upstreamMu.Unlock()

	latest, err := fetchLatestRelease(ctx, cfg)
	checkedAt := time.Now()
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	upstreamMu.Lock()
	upstreamCache[cacheKey] = releaseCache{
		latest:    latest,
		checkedAt: checkedAt,
		err:       errMsg,
	}
	upstreamMu.Unlock()

	return latest, checkedAt, errMsg
}

func fetchLatestRelease(ctx context.Context, cfg Options) (string, error) {
	apiBase := cfg.APIBaseURL
	if apiBase == "" {
		apiBase = "https://api.github.com"
	}

	client := cfg.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	url := fmt.Sprintf("%s/repos/%s/releases?per_page=30", strings.TrimRight(apiBase, "/"), cfg.GitHubRepo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "echostate-version-check")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github releases: %s", strings.TrimSpace(string(body)))
	}

	var releases []githubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		return "", err
	}

	var latest string
	for _, release := range releases {
		if release.Draft || release.TagName == "" {
			continue
		}
		tag := Normalize(release.TagName)
		if !matchesBranch(cfg.Branch, tag, release.Prerelease) {
			continue
		}
		if latest == "" || Compare(tag, latest) > 0 {
			latest = tag
		}
	}
	return latest, nil
}

func matchesBranch(branch, tag string, prerelease bool) bool {
	switch strings.ToLower(strings.TrimSpace(branch)) {
	case "main", "master", "stable", "release":
		return !prerelease && !IsPrerelease(tag)
	default:
		return prerelease || IsPrerelease(tag)
	}
}

// ResetUpstreamCache clears cached upstream lookups (for tests).
func ResetUpstreamCache() {
	upstreamMu.Lock()
	upstreamCache = make(map[string]releaseCache)
	upstreamMu.Unlock()
}
