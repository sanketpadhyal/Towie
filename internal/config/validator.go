package config

import (
	"errors"
	"fmt"
	"net/url"
)

func Validate(cfg *Config) error {
	if cfg.Backend.Target == "" {
		return errors.New("backend.target is required (e.g. http://localhost:3000)")
	}

	u, err := url.ParseRequestURI(cfg.Backend.Target)
	if err != nil {
		return fmt.Errorf("backend.target is not a valid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("backend.target scheme must be http or https, got %q", u.Scheme)
	}

	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", cfg.Server.Port)
	}

	if cfg.Server.MaxBodySize < 0 {
		return fmt.Errorf("server.max_body_size cannot be negative, got %d", cfg.Server.MaxBodySize)
	}

	switch cfg.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("logging.level must be one of debug, info, warn, error; got %q", cfg.Logging.Level)
	}

	switch cfg.Logging.Format {
	case "json", "text":
	default:
		return fmt.Errorf("logging.format must be json or text; got %q", cfg.Logging.Format)
	}

	if cfg.Compression.Enabled {
		switch cfg.Compression.Level {
		case "default", "speed", "size":
		default:
			return fmt.Errorf("compression.level must be default, speed, or size; got %q", cfg.Compression.Level)
		}
	}

	switch cfg.Security.FrameOptions {
	case "DENY", "SAMEORIGIN", "disabled", "":
	default:
		return fmt.Errorf("security.frame_options must be DENY, SAMEORIGIN, or disabled; got %q", cfg.Security.FrameOptions)
	}

	for i, route := range cfg.Routes {
		if route.Path == "" {
			return fmt.Errorf("routes[%d].path is required", i)
		}
		if route.Backend.Target != "" {
			ru, err := url.ParseRequestURI(route.Backend.Target)
			if err != nil {
				return fmt.Errorf("routes[%d].backend.target is not a valid URL: %w", i, err)
			}
			if ru.Scheme != "http" && ru.Scheme != "https" {
				return fmt.Errorf("routes[%d].backend.target scheme must be http or https, got %q", i, ru.Scheme)
			}
		}
	}

	return nil
}
