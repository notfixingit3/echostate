package main

import (
	"net/http"
	"testing"

	"github.com/notfixingit3/echostate/internal/config"
)

func TestNewServer_ReadHeaderTimeout(t *testing.T) {
	cfg := &config.Config{Port: "8080"}
	srv := newServer(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	if srv.ReadHeaderTimeout == 0 {
		t.Error("ReadHeaderTimeout must be non-zero to mitigate Slowloris (G112)")
	}

	if srv.Addr != ":8080" {
		t.Errorf("expected Addr ':8080', got %q", srv.Addr)
	}
}
