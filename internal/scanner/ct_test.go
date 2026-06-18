package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCTSearchDomain(t *testing.T) {
	tests := []struct {
		host string
		want string
	}{
		{host: "api.staging.example.com", want: "example.com"},
		{host: "www.example.com", want: "example.com"},
		{host: "example.com", want: "example.com"},
		{host: "localhost", want: ""},
	}

	for _, tt := range tests {
		if got := ctSearchDomain(tt.host); got != tt.want {
			t.Fatalf("ctSearchDomain(%q) = %q, want %q", tt.host, got, tt.want)
		}
	}
}

func TestParseCTResponse(t *testing.T) {
	payload := []ctCertRecord{
		{NameValue: "*.example.com\nexample.com"},
		{NameValue: "vpn.example.com"},
		{NameValue: "dev.example.com\n*.ignored.other.com"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}

	got, err := parseCTResponse(body, "example.com")
	if err != nil {
		t.Fatal(err)
	}

	want := []string{"dev.example.com", "example.com", "vpn.example.com"}
	if len(got) != len(want) {
		t.Fatalf("parseCTResponse() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parseCTResponse() = %#v, want %#v", got, want)
		}
	}
}

func TestGatherCT_SkipsIP(t *testing.T) {
	_, data, err := gatherCT(context.Background(), "93.184.216.34")
	if err != nil {
		t.Fatalf("gatherCT() error = %v", err)
	}
	if data["skipped"] == nil {
		t.Fatalf("expected skipped message, got %#v", data)
	}
}

func TestIsFastRetryableCTError(t *testing.T) {
	tests := []struct {
		err  error
		want bool
	}{
		{err: fmt.Errorf("HTTP 502"), want: true},
		{err: context.DeadlineExceeded, want: false},
		{err: fmt.Errorf(`Get "https://crt.sh": context deadline exceeded (Client.Timeout exceeded while awaiting headers)`), want: false},
	}

	for _, tt := range tests {
		if got := isFastRetryableCTError(tt.err); got != tt.want {
			t.Fatalf("isFastRetryableCTError(%v) = %v, want %v", tt.err, got, tt.want)
		}
	}
}

func TestFetchCRTSh_RetriesOnFastFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			time.Sleep(200 * time.Millisecond)
			http.Error(w, "slow", http.StatusGatewayTimeout)
			return
		}
		_ = json.NewEncoder(w).Encode([]ctCertRecord{{NameValue: "api.example.com"}})
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := fetchCRTSh(ctx, server.URL+"/?q=%25.example.com&output=json")
	if err != nil {
		t.Fatalf("fetchCRTSh() error = %v", err)
	}
	if attempts < 2 {
		t.Fatalf("attempts = %d, want retry", attempts)
	}
	subdomains, err := parseCTResponse(body, "example.com")
	if err != nil || len(subdomains) != 1 {
		t.Fatalf("parseCTResponse() = %#v, err = %v", subdomains, err)
	}
}

func TestGatherCT_FromMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("output") != "json" {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode([]ctCertRecord{
			{NameValue: "mail.example.com"},
			{NameValue: "portal.example.com"},
		})
	}))
	defer server.Close()

	orig := ctAPIBaseURL
	ctAPIBaseURL = server.URL + "/"
	t.Cleanup(func() { ctAPIBaseURL = orig })

	key, data, err := gatherCT(context.Background(), "www.example.com")
	if err != nil {
		t.Fatalf("gatherCT() error = %v", err)
	}
	if key != "ct" {
		t.Fatalf("key = %q, want ct", key)
	}

	subdomains, ok := data["subdomains"].([]string)
	if !ok || len(subdomains) != 2 {
		t.Fatalf("subdomains = %#v, want 2 entries", data["subdomains"])
	}
}