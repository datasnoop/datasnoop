#!/usr/bin/env bash
set -euo pipefail

if ! command -v gcc >/dev/null 2>&1 && ! command -v clang >/dev/null 2>&1; then
  echo "expected environment result: race tests require a C compiler; CI installs one on the supported Ubuntu runner"
  exit 0
fi

CGO_ENABLED=1 go test -race ./apps/api/... ./sdk/go/...
