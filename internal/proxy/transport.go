package proxy

import (
	"net"
	"net/http"
	"time"

	"github.com/sanketpadhyal/towie/internal/config"
)

func NewTransport(cfg config.Backend) *http.Transport {
	dialer := &net.Dialer{
		Timeout:   cfg.DialTimeout,
		KeepAlive: 30 * time.Second,
	}

	return &http.Transport{
		DialContext:           dialer.DialContext,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		DisableKeepAlives:     !cfg.KeepAlive,
		ForceAttemptHTTP2:     true,
	}
}
