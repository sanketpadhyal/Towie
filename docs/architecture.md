go build -o bin/towie ./cmd/towie
# Architecture

## Request Lifecycle

Every request follows a deterministic path through Towie:

```
TCP Listener        — OS accepts connection
HTTP Server         — net/http assigns goroutine, applies timeouts
Router              — matches path and method, selects handler chain
Recovery            — defers panic → 500 to prevent goroutine crash
Request ID          — generates/propagates X-Request-ID, stores in context
Logging             — records request metadata, defers response fields
Security Headers    — writes unconditional security headers
CORS                — handles preflight, injects Access-Control-* headers
Compression         — wraps ResponseWriter with gzip writer
Reverse Proxy       — constructs outbound request, dials backend, streams response
Client              — receives response
```

## Package Dependency Graph

```
cmd/towie
    │
    ├── internal/buildinfo      (no deps from this project)
    ├── internal/config         (no deps from this project)
    ├── internal/logger         (depends on: config)
    ├── internal/server
    │       ├── internal/router
    │       │       ├── internal/proxy
    │       │       │       └── internal/proxy/transport
    │       │       └── internal/middleware/*
    │       └── internal/health
    └── (all of the above)
```

No cyclic dependencies. Enforced by the Go compiler.

## Middleware Order Rationale

**Recovery must be outermost.** It catches panics from every subsequent handler. A panic above recovery crashes the goroutine.

**Request ID before logging.** The log entry needs the correlation ID. IDs must exist before the logger reads them.

**Logging uses defer.** The middleware captures request fields on entry, defers response fields (status, bytes, latency) until the chain returns. This is the only correct way to log complete HTTP transactions.

**Security headers before CORS.** Security headers are unconditional — they apply to preflight responses and error responses too.

**CORS before compression.** Preflight responses have no body. Wrapping their ResponseWriter with a gzip writer is wasteful.

**Proxy is terminal.** The proxy consumes the request and returns. Nothing runs after it.

## Transport Design

The `http.Transport` is configured once at startup and shared across all requests. Key decisions:

- `IdleConnTimeout: 90s` — 30s margin above AWS ALB's 60s idle timeout, preventing silent broken connections
- `DisableKeepAlives: false` — mandatory; otherwise every request incurs a TCP handshake
- `MaxIdleConnsPerHost: 10` — prevents connection pool monopolization when load balancing is added

## Graceful Shutdown

```
SIGTERM/SIGINT received
    │
signal.NotifyContext cancels root context
    │
http.Server.Shutdown(ctx) called with shutdown_timeout deadline
    │
No new connections accepted
    │
In-flight requests drain
    │
After deadline: remaining connections forcibly closed
    │
Process exits 0
```
