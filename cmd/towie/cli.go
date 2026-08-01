package main

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/sanketpadhyal/towie/internal/buildinfo"
	"github.com/sanketpadhyal/towie/internal/config"
	"github.com/sanketpadhyal/towie/internal/health"
	"github.com/sanketpadhyal/towie/internal/logger"
	"github.com/sanketpadhyal/towie/internal/router"
	"github.com/sanketpadhyal/towie/internal/server"
)

//go:embed towie.yaml
var defaultConfig []byte

func run(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	cmd := args[0]
	rest := args[1:]

	switch cmd {
	case "init":
		return cmdInit(rest)
	case "start":
		return cmdStart(rest)
	case "validate":
		return cmdValidate(rest)
	case "doctor":
		return cmdDoctor(rest)
	case "reload":
		return cmdReload(rest)
	case "version":
		return cmdVersion()
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command %q — run 'towie help' for usage", cmd)
	}
}

func cmdInit(_ []string) error {
	const target = "towie.yaml"
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("%s already exists — remove it first if you want to reinitialize", target)
	}
	if err := os.WriteFile(target, defaultConfig, 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", target, err)
	}
	fmt.Printf("created %s\n", target)
	fmt.Println("edit backend.target to point to your application, then run: towie start")
	return nil
}

func cmdStart(args []string) error {
	cfgPath := configPath(args)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	log := logger.New(cfg.Logging)

	backendURL, err := url.Parse(cfg.Backend.Target)
	if err != nil {
		return fmt.Errorf("parsing backend target: %w", err)
	}

	info := buildinfo.Get()
	healthHandler := health.New(cfg.Health, backendURL, info, log)
	handler := router.New(cfg, healthHandler, log)
	srv := server.New(cfg, handler)

	_ = os.WriteFile("/tmp/towie.pid", []byte(fmt.Sprintf("%d\n", os.Getpid())), 0o644)
	defer os.Remove("/tmp/towie.pid")

	return server.Run(srv, log)
}

func cmdValidate(args []string) error {
	cfgPath := configPath(args)

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "✓ configuration is valid")
	fmt.Fprintf(w, "  server\t%s:%d (max_body: %dMB)\n", cfg.Server.Host, cfg.Server.Port, cfg.Server.MaxBodySize/(1024*1024))
	fmt.Fprintf(w, "  backend\t%s\n", cfg.Backend.Target)
	fmt.Fprintf(w, "  logging\t%s / %s\n", cfg.Logging.Level, cfg.Logging.Format)
	fmt.Fprintf(w, "  health\t%s\n", cfg.Health.Path)
	fmt.Fprintf(w, "  compression\t%v\n", cfg.Compression.Enabled)
	fmt.Fprintf(w, "  cors\t%v\n", cfg.CORS.Enabled)
	fmt.Fprintf(w, "  security\t%v\n", cfg.Security.Enabled)
	fmt.Fprintf(w, "  routes\t%d configured\n", len(cfg.Routes))
	w.Flush()
	return nil
}

func cmdDoctor(args []string) error {
	cfgPath := configPath(args)
	passed := true

	check := func(name string, fn func() error) {
		err := fn()
		if err != nil {
			fmt.Printf("  ✗ %s: %s\n", name, err)
			passed = false
		} else {
			fmt.Printf("  ✓ %s\n", name)
		}
	}

	fmt.Println("towie doctor")
	fmt.Println()

	var cfg *config.Config
	check("config file readable", func() error {
		var err error
		cfg, err = config.Load(cfgPath)
		return err
	})

	if cfg == nil {
		fmt.Println()
		fmt.Println("fix config errors above before continuing")
		return errors.New("doctor found issues")
	}

	check("backend URL parseable", func() error {
		_, err := url.ParseRequestURI(cfg.Backend.Target)
		return err
	})

	check("backend reachable (TCP)", func() error {
		u, err := url.Parse(cfg.Backend.Target)
		if err != nil {
			return err
		}
		host := u.Hostname()
		port := u.Port()
		if port == "" {
			if u.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		}
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 3*time.Second)
		if err != nil {
			return fmt.Errorf("cannot reach %s:%s — is the backend running?", host, port)
		}
		conn.Close()
		return nil
	})

	check("server port available", func() error {
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("port %d is already in use", cfg.Server.Port)
		}
		ln.Close()
		return nil
	})

	fmt.Println()
	if !passed {
		return errors.New("doctor found issues — resolve them before starting towie")
	}
	fmt.Println("all checks passed — run 'towie start' to begin")
	return nil
}

func cmdReload(args []string) error {
	cfgPath := configPath(args)
	if _, err := config.Load(cfgPath); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	pidBytes, err := os.ReadFile("/tmp/towie.pid")
	if err != nil {
		fmt.Println("✓ configuration is valid")
		fmt.Println("note: no running towie PID file found at /tmp/towie.pid")
		fmt.Println("restart towie to apply changes: towie start")
		return nil
	}

	var pid int
	if _, err := fmt.Sscanf(string(pidBytes), "%d", &pid); err != nil {
		return fmt.Errorf("invalid PID file content: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process %d not found: %w", pid, err)
	}

	if err := proc.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("failed to send SIGHUP to process %d: %w", pid, err)
	}

	fmt.Printf("✓ configuration reloaded (SIGHUP sent to PID %d)\n", pid)
	return nil
}

func cmdVersion() error {
	info := buildinfo.Get()
	fmt.Printf("towie %s\n", info.Version)
	fmt.Printf("  commit: %s\n", info.Commit)
	fmt.Printf("  built:  %s\n", info.Date)
	fmt.Printf("  go:     %s\n", info.Go)
	fmt.Printf("  os:     %s/%s\n", info.OS, info.Arch)
	return nil
}

func configPath(args []string) string {
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return "towie.yaml"
}

func printUsage() {
	fmt.Print(`towie — the easiest way to make any backend production-ready

Usage:
  towie <command> [flags]

Commands:
  init      Create a towie.yaml in the current directory
  start     Start the Towie server
  validate  Validate the configuration file
  doctor    Run preflight diagnostics
  reload    Reload configuration in a running instance
  version   Print version information
  help      Show this message

Flags:
  --config <path>   Path to config file (default: towie.yaml)

Examples:
  towie init
  towie start --config /etc/towie/towie.yaml
  towie validate
  towie doctor
  towie version
`)
}
