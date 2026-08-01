package config

import "time"

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30 * time.Second
	}
	if cfg.Server.IdleTimeout == 0 {
		cfg.Server.IdleTimeout = 120 * time.Second
	}
	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 10 * time.Second
	}
	if cfg.Server.MaxBodySize == 0 {
		cfg.Server.MaxBodySize = 10 * 1024 * 1024 // 10MB default
	}

	if cfg.Backend.DialTimeout == 0 {
		cfg.Backend.DialTimeout = 10 * time.Second
	}
	if cfg.Backend.ResponseHeaderTimeout == 0 {
		cfg.Backend.ResponseHeaderTimeout = 30 * time.Second
	}
	if cfg.Backend.MaxIdleConns == 0 {
		cfg.Backend.MaxIdleConns = 100
	}
	if cfg.Backend.MaxIdleConnsPerHost == 0 {
		cfg.Backend.MaxIdleConnsPerHost = 10
	}
	cfg.Backend.KeepAlive = true

	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Format == "" {
		cfg.Logging.Format = "json"
	}
	if cfg.Logging.Output == "" {
		cfg.Logging.Output = "stdout"
	}

	if cfg.Health.Path == "" {
		cfg.Health.Path = "/health"
	}
	cfg.Health.Enabled = true
	cfg.Health.ProbeBackend = true

	if cfg.Compression.Level == "" {
		cfg.Compression.Level = "default"
	}
	if cfg.Compression.MinSize == 0 {
		cfg.Compression.MinSize = 1024
	}

	if len(cfg.CORS.AllowedMethods) == 0 {
		cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	}
	if len(cfg.CORS.AllowedHeaders) == 0 {
		cfg.CORS.AllowedHeaders = []string{"Content-Type", "Authorization"}
	}

	if cfg.Security.Enabled {
		if cfg.Security.FrameOptions == "" {
			cfg.Security.FrameOptions = "SAMEORIGIN"
		}
		if cfg.Security.ReferrerPolicy == "" {
			cfg.Security.ReferrerPolicy = "strict-origin-when-cross-origin"
		}
		cfg.Security.ContentTypeNoSniff = true
	}
}
