package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sanketpadhyal/towie/internal/buildinfo"
	"github.com/sanketpadhyal/towie/internal/config"
	"github.com/sanketpadhyal/towie/internal/health"
)

func TestHealth_OK(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	backendURL, _ := url.Parse(backend.URL)
	cfg := config.Health{Enabled: true, ProbeBackend: true, Path: "/health"}
	info := buildinfo.Info{Version: "v0.1.0"}

	h := health.New(cfg, backendURL, info, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status ok, got %v", resp["status"])
	}
}

func TestHealth_BackendDown(t *testing.T) {
	backendURL, _ := url.Parse("http://127.0.0.1:19998")
	cfg := config.Health{Enabled: true, ProbeBackend: true, Path: "/health"}
	info := buildinfo.Info{Version: "v0.1.0"}

	h := health.New(cfg, backendURL, info, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["status"] != "degraded" {
		t.Errorf("expected status degraded, got %v", resp["status"])
	}
}

func TestHealth_NoProbe(t *testing.T) {
	backendURL, _ := url.Parse("http://127.0.0.1:19997")
	cfg := config.Health{Enabled: true, ProbeBackend: false, Path: "/health"}
	info := buildinfo.Info{Version: "v0.1.0"}

	h := health.New(cfg, backendURL, info, nil)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 without probe, got %d", rr.Code)
	}
}
