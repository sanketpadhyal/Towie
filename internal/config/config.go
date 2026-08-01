package config

import "time"

type Config struct {
	Server      Server      `yaml:"server"`
	Backend     Backend     `yaml:"backend"`
	Logging     Logging     `yaml:"logging"`
	Health      Health      `yaml:"health"`
	Compression Compression `yaml:"compression"`
	CORS        CORS        `yaml:"cors"`
	Security    Security    `yaml:"security"`
	Routes      []Route     `yaml:"routes"`
}

type Server struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	MaxBodySize     int64         `yaml:"max_body_size"`
}

type Backend struct {
	Target                string        `yaml:"target"`
	DialTimeout           time.Duration `yaml:"dial_timeout"`
	ResponseHeaderTimeout time.Duration `yaml:"response_header_timeout"`
	KeepAlive             bool          `yaml:"keep_alive"`
	MaxIdleConns          int           `yaml:"max_idle_conns"`
	MaxIdleConnsPerHost   int           `yaml:"max_idle_conns_per_host"`
}

type Logging struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

type Health struct {
	Enabled      bool   `yaml:"enabled"`
	Path         string `yaml:"path"`
	ProbeBackend bool   `yaml:"probe_backend"`
}

type Compression struct {
	Enabled bool   `yaml:"enabled"`
	Level   string `yaml:"level"`
	MinSize int    `yaml:"min_size"`
}

type CORS struct {
	Enabled          bool     `yaml:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type Security struct {
	Enabled            bool   `yaml:"enabled"`
	HSTS               HSTS   `yaml:"hsts"`
	FrameOptions       string `yaml:"frame_options"`
	ContentTypeNoSniff bool   `yaml:"content_type_nosniff"`
	XSSProtection      bool   `yaml:"xss_protection"`
	ReferrerPolicy     string `yaml:"referrer_policy"`
	PermissionsPolicy  string `yaml:"permissions_policy"`
}

type HSTS struct {
	Enabled           bool  `yaml:"enabled"`
	MaxAge            int64 `yaml:"max_age"`
	IncludeSubdomains bool  `yaml:"include_subdomains"`
	Preload           bool  `yaml:"preload"`
}

type Route struct {
	Path    string        `yaml:"path"`
	Methods []string      `yaml:"methods"`
	Backend RouteBackend  `yaml:"backend"`
}

type RouteBackend struct {
	Target string `yaml:"target"`
}
