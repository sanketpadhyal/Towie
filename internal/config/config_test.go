package config_test

import (
	"os"
	"testing"

	"github.com/sanketpadhyal/towie/internal/config"
)

func TestLoad_ValidConfig(t *testing.T) {
	content := `
backend:
  target: http://localhost:3000
`
	f, err := os.CreateTemp("", "towie-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := config.Load(f.Name())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Backend.Target != "http://localhost:3000" {
		t.Errorf("expected backend target http://localhost:3000, got %s", cfg.Backend.Target)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoad_MissingBackend(t *testing.T) {
	content := `
server:
  port: 9090
`
	f, err := os.CreateTemp("", "towie-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = config.Load(f.Name())
	if err == nil {
		t.Fatal("expected error for missing backend.target, got nil")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	content := `
backend:
  target: http://localhost:3000
server:
  port: 99999
`
	f, err := os.CreateTemp("", "towie-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = config.Load(f.Name())
	if err == nil {
		t.Fatal("expected error for invalid port, got nil")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/tmp/towie-does-not-exist-xyz.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidScheme(t *testing.T) {
	content := `
backend:
  target: ftp://localhost:3000
`
	f, err := os.CreateTemp("", "towie-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	_, err = config.Load(f.Name())
	if err == nil {
		t.Fatal("expected error for invalid scheme ftp://, got nil")
	}
}

func TestLoad_InvalidCompressionLevel(t *testing.T) {
	content := `
backend:
  target: http://localhost:3000
compression:
  enabled: true
  level: ultra
`
	f, err := os.CreateTemp("", "towie-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(content)
	f.Close()

	_, err = config.Load(f.Name())
	if err == nil {
		t.Fatal("expected error for invalid compression level 'ultra', got nil")
	}
}
