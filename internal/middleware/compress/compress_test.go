package compress_test

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sanketpadhyal/towie/internal/config"
	"github.com/sanketpadhyal/towie/internal/middleware/compress"
)

func TestCompress_CompressesLargeResponse(t *testing.T) {
	cfg := config.Compression{Enabled: true, MinSize: 10}

	body := strings.Repeat("hello world ", 100)
	handler := compress.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected gzip Content-Encoding, got %q", rr.Header().Get("Content-Encoding"))
	}

	gr, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("failed to create gzip reader: %v", err)
	}
	defer gr.Close()

	decoded, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed to decompress: %v", err)
	}

	if string(decoded) != body {
		t.Error("decompressed body does not match original")
	}
}

func TestCompress_SkipsSmallResponse(t *testing.T) {
	cfg := config.Compression{Enabled: true, MinSize: 1024}

	handler := compress.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("small"))
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no gzip encoding for small response")
	}
	if rr.Body.String() != "small" {
		t.Errorf("expected body 'small', got %q", rr.Body.String())
	}
}

func TestCompress_NoAcceptEncoding(t *testing.T) {
	cfg := config.Compression{Enabled: true, MinSize: 1}

	handler := compress.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("uncompressed"))
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no gzip when Accept-Encoding is absent")
	}
}

func TestCompress_Disabled(t *testing.T) {
	cfg := config.Compression{Enabled: false}

	handler := compress.New(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("plain"))
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") == "gzip" {
		t.Error("expected no compression when disabled")
	}
}
