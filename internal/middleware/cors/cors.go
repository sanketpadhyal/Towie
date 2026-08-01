package cors

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/sanketpadhyal/towie/internal/config"
)

type handler struct {
	cfg  config.CORS
	next http.Handler
}

func New(cfg config.CORS) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &handler{cfg: cfg, next: next}
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.cfg.Enabled {
		h.next.ServeHTTP(w, r)
		return
	}

	if r.Method == http.MethodOptions {
		origin := r.Header.Get("Origin")
		if origin != "" && h.isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Add("Vary", "Origin")
			if h.cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
		} else if len(h.cfg.AllowedOrigins) > 0 {
			if h.cfg.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
		}

		if len(h.cfg.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(h.cfg.AllowedMethods, ", "))
		}

		reqHeaders := r.Header.Get("Access-Control-Request-Headers")
		if len(h.cfg.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(h.cfg.AllowedHeaders, ", "))
		} else if reqHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		}

		if len(h.cfg.ExposeHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(h.cfg.ExposeHeaders, ", "))
		}

		if h.cfg.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(h.cfg.MaxAge))
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	origin := r.Header.Get("Origin")
	if origin != "" && h.isAllowedOrigin(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Add("Vary", "Origin")

		if h.cfg.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(h.cfg.ExposeHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(h.cfg.ExposeHeaders, ", "))
		}
	} else if len(h.cfg.AllowedOrigins) > 0 && h.cfg.AllowedOrigins[0] == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	h.next.ServeHTTP(w, r)
}

func (h *handler) isAllowedOrigin(origin string) bool {
	for _, allowed := range h.cfg.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}
