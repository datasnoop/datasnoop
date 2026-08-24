# Workload Profiles

DataSnoop validates a declared single-node envelope with versioned deterministic profiles. The `v1` profiles use a fixed operation/error/log/measurement oracle and run on the repository's TimescaleDB container.

| Profile | Operations | Errors | Retries | Correlated logs | Host measurements |
| --- | ---: | ---: | ---: | ---: | ---: |
| `smoke-v1` | 3 | 2 | 1 | 2 | 3 |
| `steady-v1` | 60 | 12 | 3 | 12 | 30 |
| `burst-v1` | 240 | 48 | 12 | 48 | 30 |
| `saturation-v1` | 400 | 80 | 20 | 80 | 30 |
| `concurrent-retention-v1` | 120 | 24 | 6 | 24 | 30 |
| `recovery-v1` | 80 | 16 | 8 | 16 | 30 |
| `soak-v1` | 600 | 120 | 30 | 120 | 300 |

The oracle requires errors not to exceed sent operations, correlated logs to cover every error, and at least CPU, memory, and filesystem measurements. These profiles record accepted, rejected, throttled, and dropped outcomes; queue occupancy; connections; historical and live visibility; investigation latency; SSE gaps; and semantic correctness. The declared baseline is one local PostgreSQL/TimescaleDB container and one API process on a modest single-node host; profile results always record the actual host before comparison.
