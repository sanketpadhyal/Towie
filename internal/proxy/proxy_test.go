package proxy_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/sanketpadhyal/towie/internal/proxy"
)

func TestProxy_ForwardsRequest(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-For") == "" {
			t.Error("expected X-Forwarded-For header")
		}
		if r.Header.Get("X-Real-IP") == "" {
			t.Error("expected X-Real-IP header")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("proxied"))
	}))
	defer backend.Close()

	target, _ := url.Parse(backend.URL)
	p := proxy.New(target, http.DefaultTransport, slog.Default())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if rr.Body.String() != "proxied" {
		t.Errorf("expected body 'proxied', got %q", rr.Body.String())
	}
}

func TestProxy_BackendDown_Returns502(t *testing.T) {
	target, _ := url.Parse("http://127.0.0.1:19999")
	p := proxy.New(target, http.DefaultTransport, slog.Default())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	p.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rr.Code)
	}
}

func TestProxy_HopHeadersStripped(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Connection") != "" {
			t.Error("Connection header should be stripped")
		}
		if r.Header.Get("Keep-Alive") != "" {
			t.Error("Keep-Alive header should be stripped")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	target, _ := url.Parse(backend.URL)
	p := proxy.New(target, http.DefaultTransport, slog.Default())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Keep-Alive", "timeout=5")
	p.ServeHTTP(rr, req)
}
