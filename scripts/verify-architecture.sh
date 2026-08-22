#!/usr/bin/env bash
set -euo pipefail

forbidden="$(find apps/api/internal -mindepth 1 -maxdepth 1 -type d \( -name controller -o -name service -o -name repository -o -name model -o -name domain -o -name persistence \) -print)"
if [[ -n "$forbidden" ]]; then
  echo "forbidden horizontal package directories:" >&2
  echo "$forbidden" >&2
  exit 1
fi

if grep -R --include='*.go' -n 'internal/ingestion/postgres' apps/api/internal/ingestion --exclude-dir=postgres; then
  echo "ingestion application packages must not import their PostgreSQL adapter" >&2
  exit 1
fi

go list ./apps/api/... >/dev/null
echo "capability dependency check passed"
