package proxy

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/sanketpadhyal/towie/internal/middleware/requestid"
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func New(target *url.URL, transport http.RoundTripper, log *slog.Logger) http.Handler {
	rp := &httputil.ReverseProxy{
		Director:  director(target),
		Transport: transport,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Error("proxy error",
				"error", err,
				"target", target.String(),
				"request_id", requestid.FromContext(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadGateway)
			json.NewEncoder(w).Encode(map[string]string{
				"error":      "bad gateway",
				"request_id": requestid.FromContext(r.Context()),
			})
		},
	}
	return rp
}

func director(target *url.URL) func(*http.Request) {
	return func(r *http.Request) {
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host

		if target.Path != "" && target.Path != "/" {
			r.URL.Path = singleJoiningSlash(target.Path, r.URL.Path)
		}

		if target.RawQuery == "" || r.URL.RawQuery == "" {
			r.URL.RawQuery = target.RawQuery + r.URL.RawQuery
		} else {
			r.URL.RawQuery = target.RawQuery + "&" + r.URL.RawQuery
		}

		r.Header.Del("X-Forwarded-For")

		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			clientIP = r.RemoteAddr
		}

		if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
			r.Header.Set("X-Forwarded-For", prior+", "+clientIP)
		} else {
			r.Header.Set("X-Forwarded-For", clientIP)
		}

		r.Header.Set("X-Real-IP", clientIP)
		r.Header.Set("X-Forwarded-Host", r.Host)

		if r.TLS != nil {
			r.Header.Set("X-Forwarded-Proto", "https")
		} else {
			r.Header.Set("X-Forwarded-Proto", "http")
		}

		for _, h := range hopHeaders {
			r.Header.Del(h)
		}

		if r.Header.Get("User-Agent") == "" {
			r.Header.Del("User-Agent")
		}
	}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
