## Cross-cutting test directive

Each task SHALL add and execute the smallest test layer that proves its changed
behavior. A task that crosses a process, network, database, migration, gRPC,
GraphQL, or SSE boundary SHALL include an integration test. A task that changes
a public OTLP contract SHALL include contract or conformance coverage. A task
that changes failure, timeout, queue, retry, overload, or slow-client behavior
SHALL include resilience coverage. User-facing Lounge behavior SHALL include
component or end-to-end coverage. Load and retention-concurrency profiles remain
in task 8.4 unless an earlier task explicitly changes a declared operational
limit.

## 1. Contracts and project foundation

- [x] 1.1 Establish the approved Go module/workspace and frontend layout as a capability-oriented modular monolith with vertical use-case slices and consumer-owned ports, without overwriting the pre-existing `go.mod` deletion, and verify the baseline build plus an architecture dependency check run or report only expected empty-project results
- [x] 1.2 Document the supported OTLP logs, traces and metrics subset, semantic mappings, authentication metadata, payload limits, partial-success policy and signal-specific duplicate identity rules, and verify every ingestion spec scenario maps to a documented contract case
- [x] 1.3 Create language-independent OTLP fixtures for valid, mixed-validity, oversized, correlated and retransmitted telemetry, and verify they decode using official OpenTelemetry Protobuf definitions without importing the DataSnoop SDK
- [x] 1.4 Add automated commands for unit, contract, PostgreSQL/TimescaleDB integration, race, fuzz-seed, frontend, system-smoke, formatting, capability dependency and OpenSpec checks, and verify a clean scaffold runs each applicable suite or reports only its expected empty-project result

## 2. Persistence and normalized domain

- [x] 2.1 Implement normalized resource, service, host, operation, log and metric domain types with explicit validation limits, and verify unit tests cover correlation identifiers, timestamps, source roles and invalid attributes
- [x] 2.2 Add versioned PostgreSQL/TimescaleDB migrations for identities, signal-specific idempotency constraints and temporal signal tables, and verify against the supported real database version that migrations apply idempotently to a clean database and roll back before destructive retention occurs
- [x] 2.3 Add indexes for service/time, normalized route/status, trace/span correlation and host/time filters, and verify representative machine-readable query plans use the intended indexes on a deterministic dataset in the supported real database version
- [x] 2.4 Implement short transactional batched persistence with signal-specific idempotency and committed, repeated, rejected and failed outcome reporting, and verify integration tests cover both first delivery and retransmission

## 3. OTLP ingestion and backpressure

- [x] 3.1 Implement authenticated standard OTLP/gRPC export services for logs, traces and metrics plus separate liveness/readiness endpoints, and verify valid metadata succeeds while missing or invalid credentials persist nothing
- [x] 3.2 Implement record-level validation and normalization for the supported signal subset, and verify independent OTLP/gRPC contract fixtures plus reproducible fuzz seeds preserve resource, service, host, timestamp, severity and correlation semantics without panics or unbounded allocation
- [x] 3.3 Implement separately bounded admission, queueing, batch writing and persistence timeouts, and verify deterministic saturation plus race-enabled tests return retryable outcomes without consuming query, retention or live-delivery resource budgets
- [x] 3.4 Implement OTLP partial-success responses for mixed batches, explicit failures for wholly invalid batches and retryable mappings only for recoverable failures, and verify counts, reason categories and client retry expectations match the fixtures
- [x] 3.5 Expose bounded accepted, rejected, throttled and dropped operational metrics with restart identity, and verify diagnostics distinguish counter reset from a historical decrease

## 4. Go reference instrumentation

- [x] 4.1 Implement the opinionated Go SDK bootstrap over official OpenTelemetry providers and OTLP exporter configuration, and verify minimum valid configuration becomes ready while malformed configuration returns actionable errors
- [x] 4.2 Add `net/http` server instrumentation with normalized route, status, duration and correlation propagation, and verify tests cover success, server error and high-cardinality raw paths
- [x] 4.3 Add an idiomatic `slog` integration that preserves records and attaches active correlation context, and verify request and background logging scenarios export the expected identifiers
- [x] 4.4 Add application-host CPU, memory and disk collection with stable host identity and `monitored-service` source role, and verify unsupported or unavailable measurements are omitted rather than fabricated
- [x] 4.5 Implement bounded SDK buffering, retry, discard counters and deadline-aware shutdown, and verify backend outage and queue saturation never block request handling indefinitely

## 5. Investigation query services

- [x] 5.1 Implement time-windowed endpoint aggregation from persisted HTTP operations, and verify seeded queries return rate, error impact and duration summaries grouped by normalized route without inflation after operation retransmission
- [x] 5.2 Implement error-occurrence retrieval with service, time, route and status filters, and verify results cannot cross service or environment boundaries
- [x] 5.3 Implement trace/span-based correlated log retrieval with severity filtering, and verify correlated and merely time-adjacent logs remain distinguishable
- [x] 5.4 Implement time-aligned host context retrieval with stale and missing-data classification, and verify the query never infers causation or substitutes a healthy default
- [x] 5.5 Expose historical investigation and platform diagnostics as capability-local vertical query slices through the GraphQL boundary, and verify GraphQL contract tests plus real-database integration tests use product terminology while retaining stable pagination, response budgets and error behavior

## 6. Live delivery and Lounge

- [x] 6.1 Implement best-effort post-commit live publication with bounded per-client buffers and SSE recovery markers, and verify through the real SSE boundary that publication failure, slow clients or disconnection cannot block ingestion and produce an explicit gap without requiring a durable outbox
- [x] 6.2 Build the Lounge service overview with service/time selection and endpoint rate, errors and duration ordering, and verify UI tests cover empty, loading, populated and degraded states
- [x] 6.3 Build endpoint occurrence and operation detail views with correlated logs, supported filters and host context, and verify a user can navigate the complete incident path without entering a query language
- [x] 6.4 Build the bounded Live View with interruption and recovery states, and verify persisted history reconciles the view after reconnection
- [x] 6.5 Build onboarding and platform diagnostics showing connected signal categories, ingestion failures and application-host versus DataSnoop-platform health, and verify missing logs or metrics produce actionable non-OpenTelemetry guidance

## 7. Retention and platform operation

- [x] 7.1 Implement validated retention configuration with safe defaults and authorization, and verify unsupported values are rejected while the active policy is visible through the platform API
- [x] 7.2 Implement scheduled TimescaleDB chunk expiration with recorded start, finish, duration and failure outcomes, and verify expired telemetry is removed while in-window telemetry remains queryable
- [x] 7.3 Coordinate retention with ingestion through a separate bounded database work budget rather than global blocking, and verify a concurrent purge load test keeps ingestion within the declared degradation budget
- [x] 7.4 Collect DataSnoop's own CPU, memory and disk metrics with `datasnoop-platform` role, and verify the Lounge distinguishes them from monitored-service host measurements even on the same machine

## 8. End-to-end validation and handoff

- [x] 8.1 Create a deterministic single-service Go demonstration that emits successful requests, repeated endpoint errors, correlated `slog` records and host metrics, and verify one command produces the documented incident dataset
- [x] 8.2 Run the same supported telemetry path through an official OTLP exporter without the DataSnoop SDK, and verify investigation semantics match the SDK-generated dataset
- [x] 8.3 Add failure-injection coverage for invalid credentials, uncertain acknowledgement with retransmission, persistence outage, queue saturation, missed post-commit live publication, slow live clients, SDK shutdown timeout and interrupted purge, and verify each failure matches the corresponding spec behavior
- [x] 8.4 Define versioned deterministic smoke, steady, burst, saturation, concurrent-retention, recovery and soak workload profiles for the declared single-node hardware budget, including an oracle for sent operations, error impact, retransmissions, correlated logs and host measurements, and verify the generator reproduces the expected dataset
- [x] 8.5 Build a system workload harness that runs the complete DataSnoop deployment with real PostgreSQL/TimescaleDB, protocol-level load producers and one or a few browser sentinel users, and verify the steady profile continuously completes `endpoint errors → occurrence → correlated logs → host context` while telemetry arrives
- [x] 8.6 Run the burst, saturation, retention and recovery profiles, recording sent, accepted, rejected and throttled volume, CPU, memory, connections, queue occupancy, time to historical and live visibility, investigation latency, SSE gaps and semantic correctness, and verify overload remains bounded, explicit and recoverable
- [ ] 8.7 Run the soak profile for the declared duration, record resource trends and worker lifecycle evidence, and verify memory, goroutines and connections do not grow without a workload-explained bound
- [ ] 8.8 Document the one-shot installation and Go onboarding path plus the advanced direct-OTLP path, and verify a clean environment reaches the first endpoint overview without manual OpenTelemetry provider or Collector configuration
- [ ] 8.9 Execute the unloaded and loaded acceptance journeys, run all unit, contract, integration, race, fuzz-seed, frontend, system-smoke and strict OpenSpec checks, and update `docs/roadmap/current.md` with measured evidence and the next eligible opportunity
