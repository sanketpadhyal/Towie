package benchmark_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/sanketpadhyal/towie/internal/buildinfo"
	"github.com/sanketpadhyal/towie/internal/config"
	"github.com/sanketpadhyal/towie/internal/health"
	"github.com/sanketpadhyal/towie/internal/router"
)

func TestLoad_Smoke(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte(`{"status":"ok","message":"hello from backend"}`))
	}))

	cfg := &config.Config{
		Server:      config.Server{Port: 18090, ReadTimeout: 10e9, WriteTimeout: 10e9, IdleTimeout: 30e9, ShutdownTimeout: 5e9},
		Backend:     config.Backend{Target: backend.URL, DialTimeout: 5e9, ResponseHeaderTimeout: 5e9, KeepAlive: true, MaxIdleConns: 100, MaxIdleConnsPerHost: 20},
		Logging:     config.Logging{Level: "error", Format: "json", Output: "stdout"},
		Health:      config.Health{Enabled: true, Path: "/health", ProbeBackend: false},
		Compression: config.Compression{Enabled: true, Level: "default", MinSize: 10},
		CORS:        config.CORS{Enabled: true, AllowedOrigins: []string{"*"}, AllowedMethods: []string{"GET", "POST", "OPTIONS"}},
		Security:    config.Security{Enabled: true, FrameOptions: "SAMEORIGIN", ContentTypeNoSniff: true},
	}

	backendURL, _ := url.Parse(backend.URL)
	info := buildinfo.Info{Version: "v0.1.0"}
	h := router.New(cfg, health.New(cfg.Health, backendURL, info, slog.Default()), slog.Default())

	ts := httptest.NewServer(h)

	client := ts.Client()

	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)
	initialGoroutines := runtime.NumGoroutine()

	concurrency := 50
	totalRequests := 5000
	var wg sync.WaitGroup
	errCount := 0
	var mu sync.Mutex

	start := time.Now()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < totalRequests/concurrency; j++ {
				req, _ := http.NewRequest("GET", ts.URL+"/api/data", nil)
				req.Header.Set("Accept-Encoding", "gzip")
				resp, err := client.Do(req)
				if err != nil {
					mu.Lock()
					errCount++
					mu.Unlock()
					continue
				}
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	client.CloseIdleConnections()
	ts.Close()
	backend.Close()
	time.Sleep(50 * time.Millisecond)

	runtime.GC()
	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)
	finalGoroutines := runtime.NumGoroutine()

	rps := float64(totalRequests) / elapsed.Seconds()

	t.Logf("=== LOAD TEST RESULTS ===")
	t.Logf("Total Requests: %d", totalRequests)
	t.Logf("Concurrency:    %d", concurrency)
	t.Logf("Duration:       %v", elapsed.Round(time.Millisecond))
	t.Logf("Throughput:     %.2f req/sec", rps)
	t.Logf("Errors:         %d", errCount)
	t.Logf("Goroutines:     %d -> %d", initialGoroutines, finalGoroutines)
	t.Logf("Allocated Heap: %d KB -> %d KB", initialMem.HeapAlloc/1024, finalMem.HeapAlloc/1024)

	if errCount > 0 {
		t.Fatalf("Load test encountered %d errors", errCount)
	}
	if finalGoroutines > initialGoroutines+5 {
		t.Fatalf("Goroutine leak detected: start=%d, end=%d", initialGoroutines, finalGoroutines)
	}
}
