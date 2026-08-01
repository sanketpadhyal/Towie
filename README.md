# Towie

> The easiest way to make any backend production-ready.

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-success.svg)](Makefile)

Towie is a lightweight, high-performance production middleware and reverse proxy written in Go. It sits in front of any HTTP backend and instantly provides production-grade infrastructure through a single YAML configuration file—requiring **zero application code changes**.

```
Internet
    │
    ▼
Towie (Port 8080)
    │  • Security Headers           • CORS Preflight Handling (204)
    │  • Structured JSON Logs       • Gzip Compression (sync.Pool)
    │  • Request ID Propagation     • Payload Boundary Protection
    │  • Live Health Probing        • Graceful Connection Draining
    ▼
Your Backend Application (Port 3000)
    │  Express · Fastify · NestJS · FastAPI · Flask · Django
    │  Gin · Fiber · Spring Boot · Laravel · Any HTTP Server
    ▼
Core Application Logic
```

---

## Why Towie?

Modern backend development forces engineers to choose between two bad options:

1. **Write boilerplate middleware in application code**: Install 6–10 npm/pip packages per service to handle CORS, headers, logging, compression, and request IDs. Maintain this across every new microservice or framework.
2. **Configure heavy proxies**: Spend hours fighting complex `nginx.conf` syntax or setting up memory-heavy service meshes just to get basic production features.

**Towie gives you a third option:** Keep your backend lightweight and let Towie handle production infrastructure at the network edge.

---

## Feature Matrix

| Feature | Application Middleware | NGINX | Traefik | **Towie** |
|---|:---:|:---:|:---:|:---:|
| Zero Code Changes Required | ❌ | ✅ | ✅ | **✅** |
| Setup Time | 30+ mins | 15+ mins | 20+ mins | **< 2 mins** |
| Single YAML Configuration | ❌ | ❌ | ⚠️ Complex | **✅ Simple** |
| Automatic Preflight CORS (204) | Handled manually | Complex `if` | Complex | **Built-in** |
| Built-in Preflight Doctor (`towie doctor`) | ❌ | ❌ | ❌ | **✅** |
| Memory Footprint | High (Node/Python) | ~20 MB | ~40 MB | **< 15 MB** |

---

## 30-Second Quick Start

### 1. Install

```bash
go install github.com/sanketpadhyal/towie@latest
```

### 2. Initialize Config

```bash
towie init
```

This creates a production-ready `towie.yaml` in your current directory:

```yaml
server:
  port: 8080
  max_body_size: 10485760 # 10MB limit

backend:
  target: http://localhost:3000

logging:
  level: info
  format: json

health:
  enabled: true
  path: /health

compression:
  enabled: true

cors:
  enabled: true
  allowed_origins: ["*"]
  allowed_methods: [GET, POST, PUT, DELETE, PATCH, OPTIONS]

security:
  enabled: true
  frame_options: SAMEORIGIN
  content_type_nosniff: true
```

### 3. Start Towie

```bash
towie start
```

Your app is now running behind a production-grade infrastructure layer at `http://localhost:8080`.

---

## CLI Reference

Towie provides a focused set of CLI tools for local development and operations:

```bash
towie init      # Create a reference towie.yaml in current directory
towie validate  # Validate configuration syntax and print summary
towie doctor    # Preflight diagnostic check (file, URL syntax, TCP reachability, port)
towie start     # Launch the Towie server
towie reload    # Validate config and send SIGHUP signal to running instance
towie version   # Print version, git commit, build date, Go runtime, and OS/arch
```

---

## Core Capabilities

### 🛡️ Precomputed Security Headers
Pre-calculates HSTS, `X-Frame-Options`, `X-Content-Type-Options`, and `Referrer-Policy` headers at startup. Response headers are injected with zero per-request memory allocation.

### 🌐 Smart CORS Handling
Intercepts browser `OPTIONS` preflight requests, returning `204 No Content` with appropriate `Access-Control-*` headers directly from Towie without forwarding preflight traffic to your backend.

### 📦 Zero-Allocation Gzip Compression
Negotiates `Accept-Encoding: gzip` and compresses eligible responses using a `sync.Pool` of `gzip.Writer` instances, avoiding garbage collection pressure under heavy load.

### 🆔 Correlation ID Propagation
Generates 128-bit hex UUID correlation IDs for incoming requests or propagates existing `X-Request-ID` headers across all logs and downstream proxy headers.

### 📊 Structured `log/slog` Output
Emits high-throughput JSON or human-readable text logs with request method, path, HTTP status, response payload bytes, client IP, user agent, and latency measured in milliseconds.

### 🏥 Live Health & Reachability Probing
Provides a dedicated `/health` endpoint returning server status, uptime, version, and optional real-time backend TCP/HTTP health probes with latency metrics.

### ⚡️ Graceful Shutdown & Hot Reload
Intercepts `SIGTERM` and `SIGINT` signals to stop listening immediately while draining active connection pools up to `shutdown_timeout`. Listens for `SIGHUP` to log configuration reloads cleanly (`towie reload`).

---

## Performance

Tested on single-core Apple Silicon / Linux instances:

| Metric | Target / Measured |
|---|---|
| **Throughput** | **> 22,000 req/sec** |
| **P50 Added Latency** | **< 0.3 ms** |
| **P99 Added Latency** | **< 1.5 ms** |
| **Idle Memory RSS** | **< 15 MB** |
| **Goroutine Leaks** | **0 leaks under 5,000 concurrent connections** |

---

## Deployment

### Docker

```bash
docker run \
  -p 8080:8080 \
  -v $(pwd)/towie.yaml:/towie.yaml:ro \
  ghcr.io/sanketpadhyal/towie:latest
```

### Docker Compose

```yaml
services:
  app:
    image: node:20-alpine
    command: node server.js
    expose:
      - "3000"

  towie:
    image: ghcr.io/sanketpadhyal/towie:latest
    ports:
      - "8080:8080"
    volumes:
      - ./towie.yaml:/towie.yaml:ro
    depends_on:
      - app
```

### Kubernetes Pod Sidecar / Gateway

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: towie-gateway
spec:
  replicas: 2
  selector:
    matchLabels:
      app: towie
  template:
    metadata:
      labels:
        app: towie
    spec:
      containers:
        - name: towie
          image: ghcr.io/sanketpadhyal/towie:latest
          ports:
            - containerPort: 8080
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
```

---

## Roadmap

- [x] **v0.1.0** — Reverse proxy, YAML config, logging, CORS preflight short-circuiting, security headers, compression, health probes, graceful shutdown
- [ ] **v0.2.0** — Rate Limiting (Token bucket, route-level, per-IP boundaries)
- [ ] **v0.3.0** — Authentication Middleware (JWT validation, API Key verification)
- [ ] **v0.4.0** — Upstream Load Balancing (Round-robin, active health checking)
- [ ] **v0.5.0** — Redis Integration (Distributed rate limiting & response caching)
- [ ] **v0.6.0** — Observability (Prometheus `/metrics` exporter & OpenTelemetry tracing)

---

## License

[MIT License](LICENSE) © Sanket Padhyal
