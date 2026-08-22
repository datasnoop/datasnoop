# Project Layout

DataSnoop uses a multi-module Go workspace so the core API and reference Go SDK
can evolve with explicit dependency boundaries. The repository root deliberately
does not contain a `go.mod`; an earlier root-module bootstrap was intentionally
removed and must not be restored.

```text
apps/
├── api/                         Go modular monolith
│   └── internal/
│       ├── telemetry/           Stable normalized telemetry concepts
│       └── ingestion/           Ingestion capability and consumer-owned port
│           └── postgres/        PostgreSQL/TimescaleDB adapter
└── lounge/                      Frontend application
sdk/
└── go/        Reference Go SDK
```

The Go workspace at the repository root includes `apps/api` and `sdk/go`.
Both modules use the repository's canonical import-path prefix:
`github.com/datasnoop/datasnoop`.

Backend packages follow durable capabilities and vertical use cases rather than
global controller, service, repository, or model layers. A capability owns its
application ports; adapters import those contracts, while application behavior
does not import adapters. Shared packages are limited to stable concepts needed
across capabilities. `make architecture-check` enforces the current dependency
rules and rejects global horizontal package names.

The Lounge is a React and TypeScript application with its own package and test
tooling. The Go SDK remains a separate reference and dogfooding module.

## Baseline commands

Run Go checks across the workspace modules from the repository root:

```bash
go build ./apps/api/... ./sdk/go/...
make quality
```

Database integration suites run separately through `make integration-test`
because they start the supported real PostgreSQL/TimescaleDB container.
