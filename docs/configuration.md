# Configuration Reference

Every configuration field supported by Towie.

---

## server

Controls the HTTP listener.

| Field | Type | Default | Description |
|---|---|---|---|
| `port` | int | `8080` | Port to listen on |
| `host` | string | `""` | Bind address. Empty = all interfaces |
| `read_timeout` | duration | `30s` | Max time to read full request |
| `write_timeout` | duration | `30s` | Max time to write response |
| `idle_timeout` | duration | `120s` | Max idle keep-alive connection lifetime |
| `shutdown_timeout` | duration | `10s` | Connection drain time on SIGTERM |
| `max_body_size` | int64 | `10485760` | Max request body size in bytes (default 10MB) |

---

## backend

Controls the upstream connection.

| Field | Type | Default | Description |
|---|---|---|---|
| `target` | string | **required** | Upstream URL |
| `dial_timeout` | duration | `10s` | TCP connection timeout |
| `response_header_timeout` | duration | `30s` | Time waiting for first backend response header |
| `keep_alive` | bool | `true` | Reuse TCP connections |
| `max_idle_conns` | int | `100` | Max idle connections across all hosts |
| `max_idle_conns_per_host` | int | `10` | Max idle connections per backend host |

---

## logging

| Field | Type | Default | Description |
|---|---|---|---|
| `level` | string | `info` | `debug` \| `info` \| `warn` \| `error` |
| `format` | string | `json` | `json` \| `text` |
| `output` | string | `stdout` | `stdout` \| `stderr` \| `/path/to/file` |

Log fields emitted per request:

```json
{
  "time": "2025-01-15T10:30:00.000Z",
  "level": "INFO",
  "msg": "request",
  "request_id": "a1b2c3d4e5f6...",
  "method": "GET",
  "path": "/api/users",
  "status": 200,
  "bytes": 1024,
  "latency_ms": 12.5,
  "remote_addr": "203.0.113.1:42000",
  "user_agent": "curl/7.88.1"
}
```

---

## health

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `true` | Enable the health endpoint |
| `path` | string | `/health` | URL path |
| `probe_backend` | bool | `true` | Probe backend on each health check |

Response when healthy:

```json
{
  "status": "ok",
  "version": "v0.1.0",
  "uptime": "2h15m30s",
  "backend": {
    "target": "http://localhost:3000",
    "reachable": true
  }
}
```

Response when backend is unreachable: HTTP 503, `"status": "degraded"`.

---

## compression

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `true` | Enable gzip compression |
| `level` | string | `default` | `default` \| `speed` \| `size` |
| `min_size` | int | `1024` | Minimum response size (bytes) to compress |

Compression is negotiated via `Accept-Encoding: gzip`. Responses smaller than `min_size` are sent uncompressed.

---

## cors

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `false` | Enable CORS |
| `allowed_origins` | []string | `[]` | Allowed origins. `["*"]` = any |
| `allowed_methods` | []string | common verbs | Allowed HTTP methods |
| `allowed_headers` | []string | `[Content-Type, Authorization]` | Allowed request headers |
| `expose_headers` | []string | `[]` | Response headers the browser may read |
| `allow_credentials` | bool | `false` | Allow cookies in cross-origin requests |
| `max_age` | int | `3600` | Preflight cache duration (seconds) |

---

## security

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `true` | Enable security header injection |
| `frame_options` | string | `SAMEORIGIN` | `DENY` \| `SAMEORIGIN` \| `disabled` |
| `content_type_nosniff` | bool | `true` | `X-Content-Type-Options: nosniff` |
| `xss_protection` | bool | `false` | `X-XSS-Protection` (deprecated) |
| `referrer_policy` | string | `strict-origin-when-cross-origin` | `Referrer-Policy` value |
| `permissions_policy` | string | `""` | `Permissions-Policy` raw value |

### security.hsts

| Field | Type | Default | Description |
|---|---|---|---|
| `enabled` | bool | `false` | Enable HSTS (HTTPS only) |
| `max_age` | int | `31536000` | Max-age in seconds (1 year) |
| `include_subdomains` | bool | `false` | Apply to subdomains |
| `preload` | bool | `false` | Include in HSTS preload list |

---

## routes

Route-specific configuration. Unmatched requests fall through to the default backend.

```yaml
routes:
  - path: /api/v2        # Prefix match
    methods: []          # Empty = all methods
    backend:
      target: http://localhost:3001
```

| Field | Type | Description |
|---|---|---|
| `path` | string | URL path prefix |
| `methods` | []string | Allowed methods. Empty = all |
| `backend.target` | string | Override upstream for this route |
