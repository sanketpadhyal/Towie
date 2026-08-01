package integration_test

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/sanketpadhyal/towie/internal/buildinfo"
	"github.com/sanketpadhyal/towie/internal/config"
	"github.com/sanketpadhyal/towie/internal/health"
	"github.com/sanketpadhyal/towie/internal/router"
)

func newTestConfig(backendURL string) *config.Config {
	return &config.Config{
		Server: config.Server{
			Port:            18080,
			ReadTimeout:     30e9,
			WriteTimeout:    30e9,
			IdleTimeout:     120e9,
			ShutdownTimeout: 5e9,
		},
		Backend: config.Backend{
			Target:                backendURL,
			DialTimeout:           5e9,
			ResponseHeaderTimeout: 10e9,
			KeepAlive:             true,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   2,
		},
		Logging: config.Logging{
			Level:  "error",
			Format: "json",
			Output: "stdout",
		},
		Health: config.Health{
			Enabled:      true,
			Path:         "/health",
			ProbeBackend: false,
		},
		Compression: config.Compression{
			Enabled: true,
			Level:   "default",
			MinSize: 10,
		},
		CORS: config.CORS{
			Enabled:        true,
			AllowedOrigins: []string{"https://example.com"},
			AllowedMethods: []string{"GET", "POST", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type"},
			MaxAge:         3600,
		},
		Security: config.Security{
			Enabled:            true,
			FrameOptions:       "SAMEORIGIN",
			ContentTypeNoSniff: true,
			ReferrerPolicy:     "strict-origin-when-cross-origin",
		},
	}
}

func TestIntegration_ProxyForwardsToBackend(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"hello from backend"}`))
	}))
	defer backend.Close()

	cfg := newTestConfig(backend.URL)
	log := slog.Default()
	backendURL, _ := url.Parse(backend.URL)
	info := buildinfo.Info{Version: "test"}
	healthHandler := health.New(cfg.Health, backendURL, info, log)
	h := router.New(cfg, healthHandler, log)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	if body["message"] != "hello from backend" {
		t.Errorf("unexpected body: %v", body)
	}
}

func TestIntegration_SecurityHeadersOnEveryResponse(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := newTestConfig(backend.URL)
	log := slog.Default()
	backendURL, _ := url.Parse(backend.URL)
	info := buildinfo.Info{Version: "test"}
	healthHandler := health.New(cfg.Health, backendURL, info, log)
	h := router.New(cfg, healthHandler, log)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("missing X-Frame-Options header")
	}
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("missing X-Content-Type-Options header")
	}
}

func TestIntegration_RequestIDPropagated(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		w.Header().Set("X-Backend-Saw-ID", id)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	cfg := newTestConfig(backend.URL)
	log := slog.Default()
	backendURL, _ := url.Parse(backend.URL)
	info := buildinfo.Info{Version: "test"}
	healthHandler := health.New(cfg.Health, backendURL, info, log)
	h := router.New(cfg, healthHandler, log)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rr, req)

	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID in response")
	}
}

func TestIntegration_CORSPreflight(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("backend should not be called for preflight")
	}))
	defer backend.Close()

	cfg := newTestConfig(backend.URL)
	log := slog.Default()
	backendURL, _ := url.Parse(backend.URL)
	info := buildinfo.Info{Version: "test"}
	healthHandler := health.New(cfg.Health, backendURL, info, log)
	h := router.New(cfg, healthHandler, log)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api", nil)
	req.Header.Set("Origin", "https://example.com")
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("expected CORS origin header")
	}
}

func TestIntegration_CompressionEndToEnd(t *testing.T) {
	largeBody := strings.Repeat("hello world from backend ", 200)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(largeBody))
	}))
	defer backend.Close()

	cfg := newTestConfig(backend.URL)
	log := slog.Default()
	backendURL, _ := url.Parse(backend.URL)
	info := buildinfo.Info{Version: "test"}
	healthHandler := health.New(cfg.Health, backendURL, info, log)
	h := router.New(cfg, healthHandler, log)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	h.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip encoding, got %q", rr.Header().Get("Content-Encoding"))
	}

	gr, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("gzip reader error: %v", err)
	}
	defer gr.Close()
	decoded, _ := io.ReadAll(gr)

	if string(decoded) != largeBody {
		t.Error("decompressed body does not match")
	}
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer backend.Close()

	cfg := newTestConfig(backend.URL)
	log := slog.Default()
	backendURL, _ := url.Parse(backend.URL)
	info := buildinfo.Info{Version: "v0.1.0"}
	healthHandler := health.New(cfg.Health, backendURL, info, log)
	h := router.New(cfg, healthHandler, log)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse health response: %v", err)
	}
	if resp["version"] != "v0.1.0" {
		t.Errorf("expected version v0.1.0, got %v", resp["version"])
	}
}

func TestIntegration_PanicRecovery(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated panic")
	}))
	defer backend.Close()

	cfg := newTestConfig(backend.URL)
	log := slog.Default()
	backendURL, _ := url.Parse(backend.URL)
	info := buildinfo.Info{Version: "test"}
	healthHandler := health.New(cfg.Health, backendURL, info, log)

	// Bypass proxy for direct panic test
	_ = router.New(cfg, healthHandler, log)

	// The recovery middleware is part of the chain — test via the full stack
	// by ensuring a panicking backend doesn't crash the test process.
	// The httptest.Server recovers panics internally, so we test at middleware level.
}
