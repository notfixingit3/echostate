package scanner

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/notfixingit3/echostate/internal/models"
)

// ---------- Hash tests ----------

func TestHash_Deterministic(t *testing.T) {
	t.Parallel()

	r := &models.ScanResult{
		Host:  "example.com",
		WHOIS: map[string]any{"domain": "example.com"},
		ASN:   map[string]any{"asn": "15169"},
		Web:   map[string]any{"title": "Example"},
	}

	h1, err := Hash(r)
	if err != nil {
		t.Fatalf("Hash() first call: %v", err)
	}
	h2, err := Hash(r)
	if err != nil {
		t.Fatalf("Hash() second call: %v", err)
	}
	if h1 != h2 {
		t.Errorf("Hash() not deterministic: %q != %q", h1, h2)
	}
}

func TestHash_DifferentInputs(t *testing.T) {
	t.Parallel()

	r1 := &models.ScanResult{Host: "example.com", WHOIS: map[string]any{"domain": "example.com"}}
	r2 := &models.ScanResult{Host: "example.org", WHOIS: map[string]any{"domain": "example.org"}}

	h1, err := Hash(r1)
	if err != nil {
		t.Fatalf("Hash(r1): %v", err)
	}
	h2, err := Hash(r2)
	if err != nil {
		t.Fatalf("Hash(r2): %v", err)
	}
	if h1 == h2 {
		t.Error("Hash() produced same output for different inputs")
	}
}

func TestHash_EmptyResult(t *testing.T) {
	t.Parallel()

	r := &models.ScanResult{}
	h, err := Hash(r)
	if err != nil {
		t.Fatalf("Hash(empty): %v", err)
	}
	if h == "" {
		t.Error("Hash(empty) returned empty string")
	}
}

// ---------- mergeMap tests ----------

func TestMergeMap_MergesIntoDestination(t *testing.T) {
	t.Parallel()

	dst := map[string]any{"a": 1, "b": 2}
	src := map[string]any{"c": 3}

	mergeMap(dst, src)

	if dst["a"] != 1 {
		t.Errorf("dst[a] = %v, want 1", dst["a"])
	}
	if dst["b"] != 2 {
		t.Errorf("dst[b] = %v, want 2", dst["b"])
	}
	if dst["c"] != 3 {
		t.Errorf("dst[c] = %v, want 3", dst["c"])
	}
}

func TestMergeMap_OverwritesExistingKeys(t *testing.T) {
	t.Parallel()

	dst := map[string]any{"key": "old"}
	src := map[string]any{"key": "new"}

	mergeMap(dst, src)

	if dst["key"] != "new" {
		t.Errorf("dst[key] = %v, want 'new'", dst["key"])
	}
}

func TestMergeMap_NilDestination(t *testing.T) {
	t.Parallel()

	src := map[string]any{"a": 1}

	// mergeMap does not guard against nil destination; verify it panics.
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil destination, got none")
		}
	}()

	mergeMap(nil, src)
}

func TestMergeMap_NilSource(t *testing.T) {
	t.Parallel()

	dst := map[string]any{"a": 1}

	// Should not panic.
	mergeMap(dst, nil)

	if dst["a"] != 1 {
		t.Errorf("dst[a] = %v, want 1", dst["a"])
	}
}

func TestMergeMap_DoesNotMutateSource(t *testing.T) {
	t.Parallel()

	dst := map[string]any{"a": 1}
	src := map[string]any{"b": 2}

	mergeMap(dst, src)

	if len(src) != 1 || src["b"] != 2 {
		t.Error("mergeMap mutated the source map")
	}
}

// ---------- Run tests (with mock gatherers) ----------

// mockGatherer returns a Gatherer that produces a fixed key/value or error.
func mockGatherer(key string, value map[string]any, err error) Gatherer {
	return func(_ context.Context, host string) (string, map[string]any, error) {
		return key, value, err
	}
}

func TestRun_SuccessfulGatherer(t *testing.T) {
	t.Parallel()

	s := &Scanner{
		gatherers: []Gatherer{
			mockGatherer("whois", map[string]any{"domain": "test.com"}, nil),
		},
	}

	ctx := context.Background()
	result, err := s.Run(ctx, "test.com")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if result.Host != "test.com" {
		t.Errorf("result.Host = %q, want %q", result.Host, "test.com")
	}
	if result.WHOIS["domain"] != "test.com" {
		t.Errorf("result.WHOIS[domain] = %v, want 'test.com'", result.WHOIS["domain"])
	}
}

func TestRun_FailingGatherer(t *testing.T) {
	t.Parallel()

	s := &Scanner{
		gatherers: []Gatherer{
			mockGatherer("whois", nil, errors.New("connection refused")),
		},
	}

	ctx := context.Background()
	result, err := s.Run(ctx, "test.com")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0] != "whois: connection refused" {
		t.Errorf("error message = %q, want %q", result.Errors[0], "whois: connection refused")
	}
}

func TestRun_MultipleGatherersAggregate(t *testing.T) {
	t.Parallel()

	s := &Scanner{
		gatherers: []Gatherer{
			mockGatherer("whois", map[string]any{"domain": "test.com"}, nil),
			mockGatherer("asn", map[string]any{"asn": "15169"}, nil),
			mockGatherer("web", map[string]any{"title": "Test"}, nil),
		},
	}

	ctx := context.Background()
	result, err := s.Run(ctx, "test.com")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if result.WHOIS["domain"] != "test.com" {
		t.Errorf("WHOIS[domain] = %v, want 'test.com'", result.WHOIS["domain"])
	}
	if result.ASN["asn"] != "15169" {
		t.Errorf("ASN[asn] = %v, want '15169'", result.ASN["asn"])
	}
	if result.Web["title"] != "Test" {
		t.Errorf("Web[title] = %v, want 'Test'", result.Web["title"])
	}
}

func TestRun_MixedSuccessAndFailure(t *testing.T) {
	t.Parallel()

	s := &Scanner{
		gatherers: []Gatherer{
			mockGatherer("whois", map[string]any{"domain": "test.com"}, nil),
			mockGatherer("asn", nil, errors.New("dns timeout")),
			mockGatherer("web", map[string]any{"title": "Test"}, nil),
		},
	}

	ctx := context.Background()
	result, err := s.Run(ctx, "test.com")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if result.WHOIS["domain"] != "test.com" {
		t.Errorf("WHOIS[domain] = %v, want 'test.com'", result.WHOIS["domain"])
	}
	if result.Web["title"] != "Test" {
		t.Errorf("Web[title] = %v, want 'Test'", result.Web["title"])
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0] != "asn: dns timeout" {
		t.Errorf("error message = %q, want %q", result.Errors[0], "asn: dns timeout")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	t.Parallel()

	// A gatherer that blocks until the context is cancelled.
	blocking := func(ctx context.Context, host string) (string, map[string]any, error) {
		<-ctx.Done()
		return "whois", nil, ctx.Err()
	}

	s := &Scanner{
		gatherers: []Gatherer{blocking},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediately cancel

	result, err := s.Run(ctx, "test.com")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0] != "whois: context canceled" {
		t.Errorf("error message = %q, want %q", result.Errors[0], "whois: context canceled")
	}
}

func TestRun_ContextTimeout(t *testing.T) {
	t.Parallel()

	// A gatherer that sleeps longer than the per-gatherer timeout.
	sleepy := func(ctx context.Context, host string) (string, map[string]any, error) {
		select {
		case <-time.After(5 * time.Second):
			return "whois", map[string]any{"domain": "slow"}, nil
		case <-ctx.Done():
			return "whois", nil, ctx.Err()
		}
	}

	s := &Scanner{
		gatherers: []Gatherer{sleepy},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := s.Run(ctx, "test.com")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0] != "whois: context deadline exceeded" {
		t.Errorf("error message = %q, want %q", result.Errors[0], "whois: context deadline exceeded")
	}
}

func TestRun_FailingGathererMergesPartialData(t *testing.T) {
	t.Parallel()

	s := &Scanner{
		gatherers: []Gatherer{
			mockGatherer("whois", map[string]any{"domain": "partial.com"}, errors.New("partial failure")),
		},
	}

	ctx := context.Background()
	result, err := s.Run(ctx, "test.com")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if result.WHOIS["domain"] != "partial.com" {
		t.Errorf("result.WHOIS[domain] = %v, want 'partial.com'", result.WHOIS["domain"])
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0] != "whois: partial failure" {
		t.Errorf("error message = %q, want %q", result.Errors[0], "whois: partial failure")
	}
}

func TestRun_EmptyGatherers(t *testing.T) {
	t.Parallel()

	s := &Scanner{
		gatherers: []Gatherer{},
	}

	ctx := context.Background()
	result, err := s.Run(ctx, "test.com")
	if err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	if result.Host != "test.com" {
		t.Errorf("result.Host = %q, want %q", result.Host, "test.com")
	}
	if len(result.Errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(result.Errors))
	}
}
