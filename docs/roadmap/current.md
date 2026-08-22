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

Start the apply workflow for `investigate-single-service-errors` with task 1.1, confirming the capability-oriented modular-monolith layout, vertical use-case slices and consumer-owned ports without restoring the intentionally removed `go.mod` bootstrap.

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
