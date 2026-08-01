package router

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/sanketpadhyal/towie/internal/config"
	"github.com/sanketpadhyal/towie/internal/middleware/compress"
	"github.com/sanketpadhyal/towie/internal/middleware/cors"
	"github.com/sanketpadhyal/towie/internal/middleware/logging"
	"github.com/sanketpadhyal/towie/internal/middleware/maxbody"
	"github.com/sanketpadhyal/towie/internal/middleware/recovery"
	"github.com/sanketpadhyal/towie/internal/middleware/requestid"
	"github.com/sanketpadhyal/towie/internal/middleware/security"
	"github.com/sanketpadhyal/towie/internal/proxy"
)

func New(cfg *config.Config, healthHandler http.Handler, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	defaultProxy := buildProxy(cfg.Backend.Target, cfg, log)
	defaultChain := chain(cfg, log, defaultProxy)

	mux.Handle(cfg.Health.Path, healthHandler)

	for _, route := range cfg.Routes {
		target := cfg.Backend.Target
		if route.Backend.Target != "" {
			target = route.Backend.Target
		}
		routeProxy := buildProxy(target, cfg, log)
		routeChain := chain(cfg, log, routeProxy)
		mux.Handle(route.Path+"/", routeChain)
		mux.Handle(route.Path, routeChain)
	}

	mux.Handle("/", defaultChain)

	return mux
}

func buildProxy(target string, cfg *config.Config, log *slog.Logger) http.Handler {
	u, _ := url.Parse(target)
	t := proxy.NewTransport(cfg.Backend)
	return proxy.New(u, t, log)
}

func chain(cfg *config.Config, log *slog.Logger, handler http.Handler) http.Handler {
	handler = maxbody.New(cfg.Server.MaxBodySize)(handler)
	handler = compress.New(cfg.Compression)(handler)
	handler = cors.New(cfg.CORS)(handler)
	handler = security.New(cfg.Security)(handler)
	handler = logging.New(log)(handler)
	handler = requestid.New()(handler)
	handler = recovery.New(log)(handler)
	return handler
}
