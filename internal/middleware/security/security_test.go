package security_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanketpadhyal/towie/internal/config"
	"github.com/sanketpadhyal/towie/internal/middleware/security"
)

func TestSecurity_HeadersPresent(t *testing.T) {
	cfg := config.Security{
		Enabled:            true,
		FrameOptions:       "SAMEORIGIN",
		ContentTypeNoSniff: true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}

	handler := security.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	tests := []struct {
		header string
		want   string
	}{
		{"X-Frame-Options", "SAMEORIGIN"},
		{"X-Content-Type-Options", "nosniff"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		if got := rr.Header().Get(tt.header); got != tt.want {
			t.Errorf("header %s: expected %q, got %q", tt.header, tt.want, got)
		}
	}
}

func TestSecurity_Disabled(t *testing.T) {
	cfg := config.Security{Enabled: false}

	handler := security.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Frame-Options"); got != "" {
		t.Errorf("expected no X-Frame-Options when disabled, got %q", got)
	}
}

func TestSecurity_HSTS(t *testing.T) {
	cfg := config.Security{
		Enabled: true,
		HSTS: config.HSTS{
			Enabled:           true,
			MaxAge:            31536000,
			IncludeSubdomains: true,
			Preload:           true,
		},
	}

	handler := security.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	want := "max-age=31536000; includeSubDomains; preload"
	if got := rr.Header().Get("Strict-Transport-Security"); got != want {
		t.Errorf("expected HSTS %q, got %q", want, got)
	}
}
