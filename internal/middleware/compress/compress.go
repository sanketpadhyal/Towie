package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/sanketpadhyal/towie/internal/config"
)

var gzipPool = sync.Pool{
	New: func() any {
		return gzip.NewWriter(io.Discard)
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gw      *gzip.Writer
	minSize int
	buf     []byte
	written bool
	skip    bool
}

func (grw *gzipResponseWriter) WriteHeader(status int) {
	grw.ResponseWriter.WriteHeader(status)
}

func (grw *gzipResponseWriter) Write(b []byte) (int, error) {
	if grw.skip {
		return grw.ResponseWriter.Write(b)
	}

	if !grw.written {
		grw.buf = append(grw.buf, b...)
		if len(grw.buf) < grw.minSize {
			return len(b), nil
		}
		grw.written = true
		grw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		grw.ResponseWriter.Header().Add("Vary", "Accept-Encoding")
		grw.ResponseWriter.Header().Del("Content-Length")
		grw.gw.Reset(grw.ResponseWriter)
		n, err := grw.gw.Write(grw.buf)
		grw.buf = nil
		return n, err
	}

	return grw.gw.Write(b)
}

func (grw *gzipResponseWriter) flush() error {
	if grw.skip {
		return nil
	}
	if !grw.written && len(grw.buf) > 0 {
		_, err := grw.ResponseWriter.Write(grw.buf)
		grw.buf = nil
		return err
	}
	if grw.gw != nil && grw.written {
		return grw.gw.Close()
	}
	return nil
}

func New(cfg config.Compression) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !cfg.Enabled {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gw := gzipPool.Get().(*gzip.Writer)
			defer gzipPool.Put(gw)

			grw := &gzipResponseWriter{
				ResponseWriter: w,
				gw:             gw,
				minSize:        cfg.MinSize,
			}

			next.ServeHTTP(grw, r)
			grw.flush()
		})
	}
}
