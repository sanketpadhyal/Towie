package requestid_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sanketpadhyal/towie/internal/middleware/requestid"
)

func TestRequestID_Generated(t *testing.T) {
	handler := requestid.New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestid.FromContext(r.Context())
		if id == "" {
			t.Error("expected request ID in context, got empty string")
		}
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID in response header")
	}
}

func TestRequestID_Propagated(t *testing.T) {
	const existingID = "test-correlation-id-123"

	handler := requestid.New()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestid.FromContext(r.Context())
		if id != existingID {
			t.Errorf("expected %q, got %q", existingID, id)
		}
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Request-ID", existingID)
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("X-Request-ID"); got != existingID {
		t.Errorf("expected response header %q, got %q", existingID, got)
	}
}
