package maxbody_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanketpadhyal/towie/internal/middleware/maxbody"
)

func TestMaxBody_Exceeded(t *testing.T) {
	handler := maxbody.New(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Error("expected error when body exceeds limit, got nil")
		}
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("this payload exceeds 10 bytes"))
	handler.ServeHTTP(rr, req)
}

func TestMaxBody_WithinLimit(t *testing.T) {
	handler := maxbody.New(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(b) != "short" {
			t.Errorf("expected 'short', got %q", string(b))
		}
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("short"))
	handler.ServeHTTP(rr, req)
}
