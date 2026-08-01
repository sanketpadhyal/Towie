#!/usr/bin/env bash
set -euo pipefail

echo "running unit and integration tests..."
CGO_ENABLED=0 go test -count=1 -timeout=120s ./...
echo "all tests passed"
