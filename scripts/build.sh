#!/usr/bin/env bash
set -euo pipefail

VERSION=${VERSION:-dev}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS="-s -w \
  -X github.com/sanketpadhyal/towie/internal/buildinfo.Version=${VERSION} \
  -X github.com/sanketpadhyal/towie/internal/buildinfo.Commit=${COMMIT} \
  -X github.com/sanketpadhyal/towie/internal/buildinfo.Date=${DATE}"

mkdir -p bin
go build -ldflags "${LDFLAGS}" -o bin/towie ./cmd/towie
echo "built bin/towie (${VERSION})"
