package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/sanketpadhyal/towie/internal/buildinfo"
	"github.com/sanketpadhyal/towie/internal/config"
)

type response struct {
	Status  string          `json:"status"`
	Version string          `json:"version"`
	Uptime  string          `json:"uptime"`
	Backend backendStatus   `json:"backend"`
}

type backendStatus struct {
	Target    string  `json:"target"`
	Reachable bool    `json:"reachable"`
	LatencyMS float64 `json:"latency_ms,omitempty"`
}

type Handler struct {
	cfg     config.Health
	backend *url.URL
	info    buildinfo.Info
	log     *slog.Logger
	started time.Time
	healthy atomic.Bool
}

func New(cfg config.Health, backend *url.URL, info buildinfo.Info, log *slog.Logger) *Handler {
	h := &Handler{
		cfg:     cfg,
		backend: backend,
		info:    info,
		log:     log,
		started: time.Now(),
	}
	h.healthy.Store(true)
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	reachable := true
	var latency float64
	if h.cfg.ProbeBackend {
		reachable, latency = h.probeBackend()
	}

	status := "ok"
	code := http.StatusOK
	if !reachable {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}

	resp := response{
		Status:  status,
		Version: h.info.Version,
		Uptime:  time.Since(h.started).Round(time.Second).String(),
		Backend: backendStatus{
			Target:    h.backend.String(),
			Reachable: reachable,
			LatencyMS: latency,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) probeBackend() (bool, float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, h.backend.String(), nil)
	if err != nil {
		return false, 0
	}

	resp, err := http.DefaultClient.Do(req)
	latency := float64(time.Since(start).Microseconds()) / 1000.0
	if err != nil {
		return false, latency
	}
	resp.Body.Close()
	return true, latency
}
