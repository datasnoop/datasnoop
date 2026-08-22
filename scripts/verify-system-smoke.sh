#!/usr/bin/env bash
set -euo pipefail

if [[ -d tests/system ]]; then
  echo "system test harness detected but no runner is configured" >&2
  exit 1
fi

echo "expected empty-project result: system smoke starts with the complete deployment harness in task 8.5"
