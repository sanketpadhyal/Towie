package recovery_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanketpadhyal/towie/internal/middleware/recovery"
)

func TestRecovery_Panic(t *testing.T) {
	log := slog.Default()
	handler := recovery.New(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestRecovery_NoPanic(t *testing.T) {
	log := slog.Default()
	handler := recovery.New(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}
