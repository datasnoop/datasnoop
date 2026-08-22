# ADR 0007: Risk-Based Test Strategy

## Status

Accepted

## Context

The initial DataSnoop outcome crosses an SDK, OTLP/gRPC, concurrent bounded ingestion, PostgreSQL/TimescaleDB, historical queries, SSE, and a browser investigation journey. A passing collection of isolated package tests would not prove that these boundaries preserve correlation, degrade safely, or remain usable together on modest hardware.

The repository is still pre-implementation, so exact coverage thresholds, latency budgets, throughput limits, and browser matrices cannot yet be calibrated from executable evidence. The strategy must establish what each test level proves without turning estimates into guarantees.

## Decision

Organize tests by observable behavior and risk rather than mirroring code layers. Every normative OpenSpec scenario maps to at least one automated contract, integration, system, or acceptance test; unit tests provide cheaper evidence for the rules beneath those scenarios.

Use the following levels:

- static and architecture checks enforce formatting, analysis, capability dependency direction, and forbidden infrastructure leakage;
- unit tests cover deterministic domain rules, validation, normalization policies, state classification, retry decisions, and bounded lifecycle behavior;
- component and contract tests exercise real OTLP/gRPC, GraphQL, SSE, HTTP instrumentation, and logging boundaries while isolating only downstream ports;
- integration tests run migrations, persistence, idempotency, retention, and investigation queries against the supported PostgreSQL/TimescaleDB version rather than SQLite or mocked SQL behavior;
- end-to-end tests exercise a small number of complete user journeys through public boundaries;
- resilience and system workload tests run the complete deployment under controlled failure and load profiles.

Keep test inputs deterministic. Use injected clocks, explicit synchronization, controllable ports, and known datasets instead of timing correctness with arbitrary sleeps. Run concurrent Go paths with the race detector. Fuzz untrusted decoding, attribute limits, identifiers, and normalization boundaries with reproducible seed corpora. Treat coverage as diagnostic evidence; do not impose a repository-wide percentage until implementation provides a meaningful baseline.

The system workload acceptance test combines deterministic OTLP producers, a complete single-node DataSnoop deployment, real PostgreSQL/TimescaleDB, protocol-level read clients, and one or a few browser sentinel users. Load comes from protocol clients rather than large browser fleets. A versioned workload oracle declares expected request counts, error impact, correlations, retransmissions, and host measurements so the test validates correctness and performance together.

Version workload profiles for smoke, steady, burst, saturation, concurrent retention, recovery, and soak behavior. Record sent, accepted, rejected and throttled volume; time to historical and live visibility; investigation latency; CPU, memory, connections and queue occupancy; SSE gaps and recovery; and semantic correctness. A profile exceeding capacity succeeds only when resource bounds hold, degradation is explicit, and recovery is demonstrated.

Use tiered gates: pull requests run deterministic unit, contract, integration, race and short system-smoke coverage appropriate to the change; scheduled pipelines run heavier fuzz, browser, resilience, saturation, retention and soak profiles; release evidence runs the declared reference profile on recorded hardware. Absolute budgets and regression tolerances are published only after they are measured reproducibly.

## Consequences

**Positive:** tests follow product promises, protocol and storage semantics receive real evidence, concurrency failures are reproducible, and load validation measures whether a user can investigate while telemetry arrives rather than ingestion throughput alone.

**Negative:** real dependency and system tests cost more time and infrastructure; tiered pipelines require disciplined ownership; exact gates remain intentionally unset until a runnable baseline exists.

## Alternatives considered

- A conventional test pyramid defined only by quantity: encourages many cheap tests without proving the risky boundaries.
- Mocking PostgreSQL or replacing it with SQLite: runs quickly but cannot prove TimescaleDB migrations, constraints, query plans, retention, or concurrency.
- Load testing ingestion alone: finds receiver throughput but can miss an unusable Lounge or starved queries.
- Browser fleets as load generators: measures browser automation overhead instead of controlled DataSnoop protocol load.
- A global coverage percentage from the first commit: is easy to game and does not express compatibility, failure, concurrency, or user-journey risk.

## Follow-ups

- Calibrate coverage expectations, latency budgets, workload rates, hardware profiles, browser coverage, and regression tolerances from the first executable vertical slice.
- Promote failures found by fuzz, resilience, or system tests into deterministic regression cases at the cheapest level that preserves the behavior.
