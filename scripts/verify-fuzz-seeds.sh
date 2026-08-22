#!/usr/bin/env bash
set -euo pipefail

if find apps/api sdk/go -path '*/testdata/fuzz/*' -type f -print -quit | grep -q .; then
  go test ./apps/api/... ./sdk/go/...
else
  echo "expected empty-project result: no fuzz seed corpus exists before untrusted decoding is implemented"
fi
