package cors_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanketpadhyal/towie/internal/config"
	"github.com/sanketpadhyal/towie/internal/middleware/cors"
)

func TestCORS_Preflight(t *testing.T) {
	cfg := config.CORS{
		Enabled:        true,
		AllowedOrigins: []string{"https://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
		MaxAge:         3600,
	}

	handler := cors.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api", nil)
	req.Header.Set("Origin", "https://example.com")
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Error("expected ACAO header")
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected ACAM header")
	}
}

func TestCORS_PreflightDoesNotReachBackend(t *testing.T) {
	cfg := config.CORS{
		Enabled:        true,
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
	}

	backendReached := false
	handler := cors.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendReached = true
		w.WriteHeader(http.StatusNotImplemented)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	req.Header.Set("Origin", "https://frontend.com")
	handler.ServeHTTP(rr, req)

	if backendReached {
		t.Error("expected CORS preflight (OPTIONS) to NOT reach the backend, but it did")
	}
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected status 204 No Content, got %d", rr.Code)
	}
}

func TestCORS_WildcardOrigin(t *testing.T) {
	cfg := config.CORS{
		Enabled:        true,
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
	}

	handler := cors.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://any.com")
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "https://any.com" {
		t.Error("expected ACAO header for wildcard origin")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	cfg := config.CORS{
		Enabled:        true,
		AllowedOrigins: []string{"https://allowed.com"},
	}

	handler := cors.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("expected no ACAO header for disallowed origin")
	}
}

func TestCORS_Disabled(t *testing.T) {
	cfg := config.CORS{Enabled: false}

	called := false
	handler := cors.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("expected next handler to be called when CORS is disabled")
	}
}
