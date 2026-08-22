# Project Layout

DataSnoop uses a multi-module Go workspace so the core API and reference Go SDK
can evolve with explicit dependency boundaries. The repository root deliberately
does not contain a `go.mod`; an earlier root-module bootstrap was intentionally
removed and must not be restored.

```text
apps/
├── api/       Go API, ingestion, persistence, and platform services
└── lounge/    Frontend application
sdk/
└── go/        Reference Go SDK
```

The Go workspace at the repository root includes `apps/api` and `sdk/go`.
Both modules use the repository's canonical import-path prefix:
`github.com/datasnoop/datasnoop`.

The Lounge directory is intentionally an empty layout placeholder. Its frontend
package manager, framework, scripts, and test tooling are introduced by task
1.4, which establishes automated quality commands.

## Baseline commands

Run Go checks across the workspace modules from the repository root:

```bash
go build ./apps/api/... ./sdk/go/...
go test ./apps/api/... ./sdk/go/...
```

At this foundation stage, both modules contain documentation-only package stubs.
The commands therefore succeed with no test files. That is a scaffold check,
not a product build.
