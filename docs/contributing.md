# Contributing to Towie

Thank you for your interest in contributing.

## Before You Start

Read the [architecture documentation](architecture.md) to understand the package structure and design decisions.

## Development Setup

```bash
git clone https://github.com/sanketpadhyal/towie
cd towie
go mod tidy
make test
```

## Standards

### Code

- Follow idiomatic Go. Read `net/http` and `log/slog` source for the bar.
- Every function has one purpose and stays under 40 lines.
- Errors are wrapped with context: `fmt.Errorf("doing X: %w", err)`.
- No global mutable state.
- No `interface{}` in the request path.

### Tests

- Every new package requires a `_test.go` file.
- Tests use only the standard library — no test frameworks.
- All tests must pass with `-race`.

### Pull Requests

- Keep scope small. One concern per PR.
- Write a clear description of what changed and why.
- Update documentation when adding user-facing features.
- All CI checks must pass before merge.

## Adding a New Middleware

1. Create `internal/middleware/<name>/<name>.go`
2. Export one function: `New(cfg config.X, log *slog.Logger) func(http.Handler) http.Handler`
3. Add a `_test.go` beside it
4. Add the config field to `internal/config/config.go`
5. Add the default in `internal/config/defaults.go`
6. Wire it into the chain in `internal/router/router.go`
7. Document it in `docs/configuration.md`

The existing packages are not modified.
