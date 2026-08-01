package security

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sanketpadhyal/towie/internal/config"
)

func New(cfg config.Security) func(http.Handler) http.Handler {
	headers := precompute(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func precompute(cfg config.Security) map[string]string {
	if !cfg.Enabled {
		return map[string]string{}
	}

	headers := make(map[string]string)

	if cfg.HSTS.Enabled {
		var sb strings.Builder
		fmt.Fprintf(&sb, "max-age=%d", cfg.HSTS.MaxAge)
		if cfg.HSTS.IncludeSubdomains {
			sb.WriteString("; includeSubDomains")
		}
		if cfg.HSTS.Preload {
			sb.WriteString("; preload")
		}
		headers["Strict-Transport-Security"] = sb.String()
	}

	if cfg.FrameOptions != "" && cfg.FrameOptions != "disabled" {
		headers["X-Frame-Options"] = cfg.FrameOptions
	}

	if cfg.ContentTypeNoSniff {
		headers["X-Content-Type-Options"] = "nosniff"
	}

	if cfg.XSSProtection {
		headers["X-XSS-Protection"] = "1; mode=block"
	}

	if cfg.ReferrerPolicy != "" {
		headers["Referrer-Policy"] = cfg.ReferrerPolicy
	}

	if cfg.PermissionsPolicy != "" {
		headers["Permissions-Policy"] = cfg.PermissionsPolicy
	}

	return headers
}
