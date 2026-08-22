# Current State

This file is the persistent checkpoint for the Ralph loop. Update it at the end of every relevant iteration.

## Current outcome

Plan a vertical slice that lets an instrumented application send telemetry and lets a user identify an endpoint with repeated errors, open an occurrence, and inspect correlated logs with basic machine context.

## Planning completion criteria

- [x] Initial documentation harness created
- [x] Harness validated for links, consistency, and repository state
- [x] Vertical OpenSpec change scaffolded
- [x] Proposal created with explicit scope and non-goals
- [x] Delta specs created for durable capabilities
- [x] Design records flows, decisions, risks, and alternatives
- [x] Tasks organize verifiable increments for future application
- [x] OpenSpec change validated
- [x] Existing persisted repository content translated to English
- [x] Repository README created with honest pre-implementation status and navigation

## Current constraints

- The first experience is single-service.
- OTLP is the preferred external ingestion contract.
- OpenTelemetry remains hidden in default onboarding.
- The Go SDK is a reference and dogfooding implementation, not a market boundary.
- No spec name uses D0, D1, D2, MVP, or temporal phases.
- Follow-ups remain explicit in `opportunities.md`.
- All persisted repository content and source code are written in English, including identifiers, APIs, schemas, comments, errors, logs, and tests; agent-user interaction may use any language.

## Next item

Continue the apply workflow for `investigate-single-service-errors` with task 3.1, implementing authenticated OTLP/gRPC export services and health endpoints.

## Iteration log

### 2026-08-22 — Documentation harness bootstrap

- Consolidated the product vision, principles, capability map, and glossary.
- Created an opportunity register with resumption triggers.
- Recorded decisions about OTLP, the Go SDK, single-service scope, and simplified UX.
- Established the reading order and iteration protocol in `AGENTS.md`.
- Validated OpenSpec root resolution and the presence of canonical documents.
- Removed the previous `go.mod`-only bootstrap from active history.
- No application code was changed.

### 2026-08-22 — Vertical slice planning

- Scaffolded `investigate-single-service-errors` through the OpenSpec CLI.
- Created five delta specs: ingestion, Go instrumentation, investigation, host monitoring, and platform operation.
- Recorded that OTLP uses standard unary `Export` services instead of proprietary client streaming.
- Created incremental tasks with explicit verification and end-to-end acceptance.
- Evidence: `openspec status` reports 4/4 planning artifacts complete.
- Evidence: `openspec validate investigate-single-service-errors --strict` passed.
- Next unit: change task 1.1, only after explicit authorization for apply.

### 2026-08-22 — Repository language convention

- Established English as the mandatory language for persisted repository content.
- Kept agent-user conversation language independent from repository output language.
- Translated the documentation harness, ADRs, proposal, and design without changing their semantics.
- Clarified that source code, identifiers, APIs, database schemas, comments, errors, logs, and tests must also use English.

### 2026-08-22 — Repository entry point

- Added `README.md` with the product problem, initial investigation journey, architecture principles, scope boundaries, repository map, and development workflow.
- Marked the project explicitly as planning-stage software with no runnable product yet.

### 2026-08-22 — Project foundation layout

- Preserved the intentional absence of a root `go.mod`.
- Added a Go workspace with separate `apps/api` and `sdk/go` modules using the canonical `github.com/datasnoop/datasnoop` import-path prefix.
- Reserved `apps/lounge` for the frontend without selecting tooling ahead of the quality-command task.
- Documented the layout and baseline Go commands in `docs/development/layout.md`.
- Evidence: `go build ./apps/api/... ./sdk/go/...` and `go test ./apps/api/... ./sdk/go/...` succeeded with the expected empty-package warnings.
- Next unit: change task 1.2, document the supported OTLP contract and map every ingestion scenario to it.

### 2026-08-22 — OTLP ingestion contract

- Defined the supported unary OTLP/gRPC logs, HTTP server spans, and host metric subset in `docs/contracts/otlp-grpc.md`.
- Documented authentication metadata, transport and record limits, normalization mappings, partial-success behavior, and retryable versus non-retryable failures.
- Mapped all nine telemetry-ingestion specification scenarios to contract behavior.
- Evidence: the contract traceability table contains all nine scenarios and the documented signal and partial-success sections are present.
- Next unit: change task 1.3, create exporter-independent OTLP fixtures.

### 2026-08-22 — Exporter-independent OTLP fixtures

- Added OTLP JSON fixtures for valid traces and metrics, mixed-validity logs, oversized attribute sets, and correlated trace/log records.
- Added decoder tests that use official OpenTelemetry Protobuf definitions without importing the DataSnoop SDK.
- Evidence: `go test ./...` in `apps/api` passed.
- Next unit: change task 1.4, add repeatable quality commands.

### 2026-08-22 — Scaffold quality commands

- Added the Lounge React/TypeScript scaffold with Vitest and Prettier checks.
- Added root `Makefile` targets for Go tests, frontend tests, formatting checks, strict OpenSpec validation, and builds.
- Evidence: `make quality` and `make build` passed.
- Next unit: change task 2.1, define and test the normalized telemetry domain.

### 2026-08-22 — Normalized telemetry domain

- Added API-internal resource, service, host, operation, log, metric, correlation, and source-role types.
- Added validation limits aligned with the OTLP contract, including identifiers, timestamps, attributes, host identity, and finite measurements.
- Evidence: `go test ./...` in `apps/api` passed, covering valid correlation, invalid timestamps and identifiers, invalid attributes, source roles, and host requirements.
- Next unit: change task 2.2, create database migrations with rollback verification.

### 2026-08-22 — Temporal persistence migrations

- Added idempotent TimescaleDB migrations for service, host, resource, operation, log, and metric storage.
- Added a disposable containerized migration verifier that applies the migration twice and rolls it back before any retention behavior exists.
- Evidence: `make migration-verify` passed against `timescale/timescaledb:2.29.1-pg17`.
- Next unit: change task 2.3, add indexes and verify representative query plans.

### 2026-08-22 — Investigation indexes

- Added indexes for service/time, route/status/time, trace/span correlation, and resource/time metric filters.
- Added a seeded TimescaleDB query-plan verifier for endpoint, correlated-log, and host-metric queries.
- Evidence: `make index-verify` passed, with each representative plan using its intended chunk index.
- Next unit: change task 2.4, implement transactional batched persistence.

### 2026-08-22 — Transactional batch persistence

- Added transactional persistence for normalized operations, logs, and metrics with explicit committed, rejected, and failed outcomes.
- Added an isolated database integration verifier that checks mixed valid/invalid batches and a persistence failure rollback.
- Evidence: `make persistence-verify` passed against the disposable TimescaleDB container.
- Next unit: change task 3.1, expose authenticated OTLP/gRPC and health endpoints.
