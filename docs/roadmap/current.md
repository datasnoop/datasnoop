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

Continue the apply workflow for `investigate-single-service-errors` with task 3.1, implementing authenticated OTLP/gRPC export services and separate liveness and readiness endpoints.

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

### 2026-08-22 — Architecture pattern consolidation

- Selected a capability-oriented modular monolith with vertical use-case slices and consumer-owned Ports and Adapters.
- Defined OTLP ingestion as at-least-once, with signal-specific identity and idempotent retransmission of identifiable HTTP operations.
- Kept historical queries authoritative and live SSE publication best-effort after commit; deferred a durable outbox until a downstream consumer requires guaranteed delivery.
- Reconciled the ingestion spec, design and implementation tasks with the accepted structural and delivery decisions.
- Evidence: `openspec status --change investigate-single-service-errors` reports 4/4 artifacts complete, and `openspec validate investigate-single-service-errors --strict` passed.
- Next unit remains task 1.1 under the refined architecture constraints.

### 2026-08-22 — Test strategy consolidation

- Selected risk-based test levels that map normative scenarios to automated evidence without mirroring package layers.
- Required real PostgreSQL/TimescaleDB integration, independent OTLP fixtures, deterministic concurrency, race detection and reproducible fuzz seeds at the relevant boundaries.
- Defined a system workload acceptance test that combines deterministic telemetry producers, a complete deployment and browser sentinel users repeating the investigation journey under load.
- Split workload validation into versioned profiles with a known semantic oracle and recorded ingest, query, live-delivery and resource evidence.
- Evidence: `openspec status --change investigate-single-service-errors` reports 4/4 artifacts complete, and `openspec validate investigate-single-service-errors --strict` passed.
- Next unit remains task 1.1; exact budgets and regression tolerances will be calibrated from the first executable vertical slice.

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

### 2026-08-22 — Persistence foundation reconciliation

- Merged the existing repository, Lounge, contract, fixture, normalized-domain, migration, index, and transactional-persistence foundation onto the current architecture and testability baseline.
- Reorganized backend code into stable telemetry concepts, the ingestion capability with a consumer-owned batch-store port, and a PostgreSQL/TimescaleDB adapter; added an automated capability dependency check.
- Documented at-least-once delivery and signal-specific duplicate behavior, added an exporter-independent retransmission fixture, and preserved independent acceptance for logs and measurements without reliable identity.
- Added a regular operation identity table keyed by service/environment identity, trace ID, and span ID so TimescaleDB hypertable partitioning does not weaken semantic idempotency.
- Extended persistence outcomes with `repeated` and verified that retransmitting an HTTP operation retains one operation while reporting the second delivery explicitly.
- Added explicit unit, contract, integration, race, fuzz-seed, frontend, system-smoke, formatting, architecture, and OpenSpec commands. Race execution reports its missing local C-compiler prerequisite; CI runs it on the supported Ubuntu runner. Fuzz and system commands report their expected pre-implementation state.
- Evidence: `make quality`, `make build`, `make integration-test`, and `openspec validate investigate-single-service-errors --strict` passed, with the documented local race prerequisite result.
- Next unit: task 3.1, authenticated OTLP/gRPC export services and independent health endpoints.

### 2026-08-22 — Authenticated OTLP receiver boundary

- Added standard OTLP/gRPC logs, traces, and metrics export services guarded by exactly one `x-datasnoop-token` metadata value.
- Added separate `GET /health/live` and `GET /health/ready` endpoints; liveness reports process vitality while readiness reports dependency state.
- Kept the raw authenticated request sink isolated so task 3.2 can add record normalization without widening the transport contract.
- Evidence: `go test ./...` in `apps/api` passed, including real gRPC boundary coverage that valid metadata reaches all three services and missing or invalid metadata never reaches the sink.
- Next unit: task 3.2, normalize and validate the supported OTLP records.

### 2026-08-22 — OTLP record normalization

- Added record-granular normalization of supported OTLP server spans, log records, and monitored-service host gauges into the normalized telemetry domain.
- Preserved service, environment, host, source timestamps, severity, and trace/span correlation; unsupported or invalid records receive contract reason categories.
- Corrected the independent OTLP fixtures to use Protobuf JSON's base64 representation for byte identifiers, making their trace and span IDs valid at the protocol boundary.
- Evidence: `go test ./...` and `go test -run=^$ -fuzz=FuzzNormalizeLogs -fuzztime=2s ./internal/ingestion/otlp` passed in `apps/api`.
- Next unit: task 3.3, bound admission, queueing, batch writing, and persistence timeouts.

### 2026-08-22 — Bounded ingestion processing

- Added independent admission and queue limits, worker execution, and a per-write persistence deadline around the consumer-owned batch store.
- Made the bounded ingest queue observable without sharing its resource budget with future query, retention, or live-delivery work.
- Evidence: `go test ./internal/ingestion -race` passed in `apps/api`, covering deterministic admission/queue saturation and persistence timeout behavior.
- Next unit: task 3.4, map normalized outcomes to OTLP partial-success and retryable responses.

### 2026-08-22 — OTLP outcomes and retry mapping

- Connected normalized records to the bounded processor and mapped record rejections to OTLP partial-success fields for logs, traces, and metrics.
- Authenticated wholly invalid requests now return `InvalidArgument`; recoverable admission, queue, timeout, and persistence failures return `Unavailable` without partial-success masking.
- Evidence: `go test ./internal/ingestion/receiver -race` and `go test ./...` passed in `apps/api`, including mixed-validity and wholly-invalid OTLP/gRPC fixture coverage.
- Next unit: task 3.5, expose bounded ingestion operational metrics and restart identity.

### 2026-08-22 — Ingestion operational diagnostics

- Added restart-scoped atomic counters for accepted, rejected, throttled, and dropped telemetry outcomes.
- Attached the counters to the processing sink so partial rejections and capacity throttling are observable alongside a stable restart identity.
- Evidence: `go test ./...` passed in `apps/api` (29 tests); `openspec validate investigate-single-service-errors --strict` passed.
- Next unit: task 4.1, implement the Go reference SDK bootstrap.

### 2026-08-22 — Go SDK bootstrap

- Added the reference SDK's one-call bootstrap with service name, endpoint, and credential configuration.
- Bootstrap configures the official OpenTelemetry OTLP/gRPC trace exporter and resource provider while attaching the DataSnoop authentication metadata; malformed configuration returns user-actionable errors.
- Evidence: `go test ./...` passed in `sdk/go`.
- Next unit: task 4.2, add `net/http` server instrumentation.

### 2026-08-22 — Go HTTP instrumentation

- Added `net/http` middleware that creates server spans with normalized route, method, response status, duration, and the active OpenTelemetry context.
- The route resolver allows application routers to provide a low-cardinality route template instead of recording raw paths.
- Evidence: `go test ./...` passed in `sdk/go`, covering successful and failing HTTP handlers using a high-cardinality request path with a normalized route.
- Next unit: task 4.3, add correlated `slog` integration.

### 2026-08-22 — Go structured log correlation

- Added an idiomatic `slog.Handler` adapter that preserves downstream logging and emits the original message, severity, attributes, and active trace/span identifiers to a configured exporter.
- Background logs export without fabricated correlation identifiers.
- Evidence: `go test ./...` passed in `sdk/go`.
- Next unit: task 4.4, collect supported application-host measurements.

### 2026-08-22 — Application-host collection

- Added stable-host `monitored-service` measurement collection for CPU, runtime memory, and filesystem utilization.
- CPU is emitted only after two valid `/proc/stat` samples; unavailable CPU or filesystem measurements are omitted.
- Evidence: `go test ./...` passed in `sdk/go`.
- Next unit: task 4.5, add SDK buffering, retry, discard metrics, and deadline-aware shutdown.

### 2026-08-22 — Bounded SDK delivery

- Added a bounded asynchronous SDK delivery buffer with retry and discard counters.
- Submission is non-blocking under backend outage or queue saturation; shutdown drains only until the caller's deadline.
- Evidence: `go test ./... -race` passed in `sdk/go`.
- Next unit: task 5.1, add endpoint aggregation queries.

### 2026-08-22 — Endpoint aggregation query

- Added a capability-local PostgreSQL endpoint summary reader with service, environment, and time-window isolation.
- It groups by normalized route and returns request volume, server-error impact, and average operation duration.
- Evidence: `make integration-test` passed with the TimescaleDB integration assertion confirming retransmission does not inflate summaries.
- Next unit: task 5.2, retrieve filtered error occurrences.

### 2026-08-22 — Isolated error-occurrence query

- Added error-occurrence retrieval constrained by service, environment, normalized route, HTTP status, and time window.
- Evidence: `make integration-test` passed with a real-database test proving results do not cross service or environment boundaries.
- Next unit: task 5.3, retrieve exact-correlated logs.

### 2026-08-22 — Correlated evidence and host context

- Added exact trace/span correlated-log retrieval with severity filtering; time-adjacent logs remain excluded.
- Added time-aligned monitored-host context with explicit `available`, `stale`, and `missing` states, without causal claims or fabricated healthy measurements.
- Evidence: `make integration-test` passed with real TimescaleDB coverage for both behaviors.
- Next unit: task 5.5, expose historical queries through GraphQL.

### 2026-08-22 — GraphQL investigation boundary

- Added product-oriented GraphQL endpoint summaries and platform diagnostics, with bounded cursor pagination and a 1 MiB request budget.
- Evidence: GraphQL contract coverage plus `make integration-test`, `make quality`, `make build`, and strict OpenSpec validation passed.
- Next unit: task 6.1, add bounded post-commit live delivery.

### 2026-08-22 — Bounded live delivery

- Added post-persistence best-effort publication through bounded SSE client buffers.
- Slow clients receive explicit `gap` events; publication never waits for a client and no durable outbox was introduced.
- Evidence: SSE boundary and race coverage plus `make quality`, `make build`, and strict OpenSpec validation passed.
- Next unit: task 6.2, build the Lounge service overview.

### 2026-08-24 — Lounge service overview

- Built the service and time-window overview with endpoint request volume, errors, and duration ordered by error impact.
- Added rendered UI coverage for loading, empty, populated, and degraded states.
- Evidence: `make quality`, `make build`, and strict OpenSpec validation passed.
- Next unit: task 6.3, build occurrence and operation detail views.

### 2026-08-24 — Lounge incident detail

- Added an endpoint occurrence detail with status and severity filters, correlated logs, and explicit non-causal host context.
- Evidence: rendered journey UI coverage plus `make quality`, `make build`, and strict OpenSpec validation passed.
- Next unit: task 6.4, build bounded Live View recovery states.

### 2026-08-24 — Lounge Live View

- Added connected, interrupted, and persisted-history recovery states for Live View.
- Evidence: rendered UI recovery coverage plus `make quality`, `make build`, and strict OpenSpec validation passed.
- Next unit: task 6.5, build onboarding and platform diagnostics.

### 2026-08-24 — Lounge onboarding diagnostics

- Added product-oriented connected-signal and platform-health diagnostics.
- Missing logs and host measurements now have actionable integration guidance without requiring OpenTelemetry terminology.
- Evidence: rendered diagnostics coverage plus `make quality`, `make build`, and strict OpenSpec validation passed.
- Next unit: task 7.1, add authorized retention configuration.

### 2026-08-24 — Authorized retention configuration

- Added a platform retention policy with a safe 30-day default and supported bounds of one to 365 days.
- Rejected unauthorized and unsupported updates, and exposed the active policy through a read-only platform HTTP boundary.
- Evidence: retention unit and HTTP-boundary coverage plus `make quality`, `make build`, and strict OpenSpec validation passed.
- Next unit: task 7.2, add scheduled TimescaleDB chunk expiration with recorded outcomes.
