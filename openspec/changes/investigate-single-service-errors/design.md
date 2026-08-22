## Context

The repository has no product implementation or main specs yet. This change introduces a vertical path across the SDK, external contract, ingestion, persistence, queries, and Lounge. See `proposal.md` for motivation and `docs/decisions/` for accepted architectural decisions.

The determining constraints are single-node operation on modest hardware, onboarding that does not require OpenTelemetry knowledge, OTLP interoperability, and safe instrumentation failure. Although the future vision includes multiple languages and distributed services, the first investigation closes the context of one service.

## Goals / Non-Goals

**Goals:**

- keep a standard OTLP contract at the boundary and an explicit DataSnoop domain internally;
- close the `problematic endpoint → failing operation → logs → host` journey;
- make queues, concurrency, payloads, retention, and connections observably bounded;
- allow an independent OTLP exporter to replace the Go SDK;
- distinguish platform health from monitored-application health.

**Non-Goals:**

- implement the entire OTLP model or promise full visualization of every received field;
- build distributed tracing, an arbitrary query language, or market-oriented SDKs in this change;
- make PostgreSQL interchangeable through a generic storage interface;
- require or bundle an OpenTelemetry Collector;
- provide high availability or multi-node coordination.

## Decisions

### 1. Use standard OTLP/gRPC services without a proprietary streaming protocol

The receiver implements the official log, trace, and metric export services. Each `Export` call is unary and carries OTLP-defined batches; batching, concurrency, and retries belong to the exporter. DataSnoop does not add a proprietary client-streaming RPC because that would break compatibility with existing exporters.

The initial official path is gRPC. Credentials travel through metadata and never through query strings. Message limits, deadlines, keepalive, and concurrency are configured explicitly.

**Alternatives:** Proprietary Protobuf would provide a smaller contract but require custom exporters. OTLP/HTTP would broaden immediate compatibility but duplicate the transport surface before demonstrated demand.

### 2. Separate the external contract, normalized domain, and persistence

The pipeline has explicit boundaries:

```text
OTLP receiver
    │ validates authentication and transport limits
    ▼
signal validator
    │ accepts or rejects records and produces partial success
    ▼
normalizer
    │ translates OTLP into resources, operations, logs, and measurements
    ▼
bounded ingest queue
    │ controls memory and backpressure
    ▼
batch writer ──▶ PostgreSQL/TimescaleDB
                      │
                      ├──▶ query services
                      └──▶ live publisher after persistence
```

Generated OTLP types do not cross into queries or Lounge components. The normalized domain preserves payload information required for compatibility and investigation without blindly mirroring the entire OTLP tree.

**Alternative:** Persisting raw Protobuf would simplify initial ingestion but move normalization into every query and make predictable indexing harder.

### 3. Model resources, operations, logs, and metrics separately

The relational model has stable resource, service, and host identities plus distinct temporal tables for operations, logs, and metrics. Universal fields and high-value filters use typed columns; additional attributes use JSONB with documented limits.

HTTP operations preserve trace ID, span ID, parent span ID, normalized route, method, status, start time, and duration. Logs preserve trace and span IDs when available. Host metrics include host identity, source role, timestamp, name, unit, and value.

PostgreSQL with TimescaleDB is the initial persistence layer. Isolation occurs through packages and domain services oriented around real queries, not through a `Repository<T>` that promises transparent replacement with an analytical database.

**Alternatives:** A universal event table would reduce DDL but weaken semantics, indexes, and validation. Splitting `logs_ingest` and `logs_search` from the start will be considered only if benchmarks demonstrate index contention; it is not part of this change's contract.

### 4. Derive RED from operations before duplicating metrics

Rate, Errors, and Duration per route are calculated from HTTP operations in the requested window. Continuous or materialized aggregates are added only when measurements show that direct queries violate the response budget.

OTLP metrics remain accepted for numeric signals such as host CPU, memory, and disk, but the SDK does not send a second copy of RED by default.

**Alternative:** Writing operations and RED series simultaneously would speed up queries but create dual consistency and higher volume before need is measured.

### 5. Publish live data only after acceptance

The live publisher receives normalized references after persistence confirmation. Each viewer has a bounded buffer; slow clients lose their connection or receive a gap marker instead of accumulating memory or delaying ingestion.

The Lounge uses GraphQL for historical queries and SSE for unidirectional updates. The UI treats live streaming as a convenience: after interruption, recovery uses historical queries, which remain authoritative.

**Alternatives:** Publishing before persistence reduces latency but can display data that disappears after failure. WebSocket provides bidirectional communication that the initial view does not require.

### 6. Build the Go SDK as an opinionated OpenTelemetry distribution

The SDK configures resources, propagation, providers, processors, and official OTLP exporters. `net/http` and `slog` integrations use OpenTelemetry context, while the public DataSnoop API uses simple concepts and safe defaults.

The queue and exporter have configurable size, batch, timeout, retry, and shutdown deadline with conservative defaults. The default policy never blocks a request indefinitely; after capacity and retry are exhausted, it drops according to the documented policy and increments local metrics.

A fixture suite sends OTLP directly to the receiver without importing the SDK, preventing the backend from coupling to Go decisions.

**Alternatives:** A fully proprietary SDK would maximize control but duplicate propagation, batching, and contracts. Requiring manual OpenTelemetry configuration would transfer the learning curve to the primary audience.

### 7. Collect host context with explicit identity and role

The reference SDK may collect basic application-host metrics at a bounded interval. The core separately collects its operational telemetry. Both use host identity and a source role (`monitored-service` or `datasnoop-platform`) to avoid ambiguity when they share a machine.

The Lounge aligns nearby points with an occurrence and marks missing or stale data; temporal correlation is presented as context, never automatic causation.

**Alternative:** Collecting only the DataSnoop machine simplifies operation but does not help distinguish application-host pressure during an incident.

### 8. Include health, diagnostics, and retention in the vertical slice

Liveness measures process vitality only. Readiness depends on the receiver being able to accept work and persistence being able to perform required operations. Onboarding diagnostics combine the latest accepted telemetry, observed signal categories, and rejection categories without exposing internal OpenTelemetry configuration.

Temporal data uses configurable retention with safe limits. The scheduler records the start, finish, duration, and error of each cycle; with TimescaleDB, removal prefers expired chunks. Purging does not block the ingest queue, and load tests cover its impact.

**Alternative:** Manual retention would reduce scope but contradict predictable VPS operation and prevent the slice from being usable as a product.

### 9. Build the Lounge with custom visual primitives and scoped CSS

The Lounge uses React and TypeScript with CSS Modules and CSS custom properties
for design tokens. Product components remain custom so the visual identity can
support DataSnoop's distinctive style while keeping critical investigation and
operational states clear.

Accessible low-level primitives such as dialogs, menus, selects, and tooltips
may use Radix UI when a current interaction requires them. The Lounge does not
adopt Tailwind CSS or a complete visual component suite in this change. This
avoids distributing visual decisions across utility markup or inheriting a
generic dashboard appearance before the product design language is established.

**Alternatives:** Tailwind CSS would accelerate utility-first composition but
would couple visual decisions more tightly to component markup. A complete UI
suite would provide ready-made components but would constrain the product's
visual identity and add components beyond the current investigation journey.

## Risks / Trade-offs

- **[The supported OTLP subset may surprise exporters]** → publish a support matrix, use partial success, and test official exporters; never drop silently.
- **[High-cardinality attributes may degrade storage]** → limit sizes and counts, index only proven filters, and expose rejection or truncation.
- **[PostgreSQL and purge compete with ingestion]** → use batches, hypertables, chunk removal, a configurable window, and concurrent benchmarks.
- **[The SDK may hide details required for troubleshooting]** → provide opinionated diagnostics and an advanced mode with exporter status without polluting onboarding.
- **[Live View may diverge from history]** → publish after persistence and signal gaps; history remains authoritative.
- **[Host metrics vary across operating systems]** → start with a documented matrix and represent absence without fabricated values.
- **[A broad vertical change increases coordination]** → establish contracts and fixtures first, use verifiable tasks, and keep the end-to-end demonstration executable throughout the work.

## Migration Plan

The project is greenfield, so there is no existing data migration. Application proceeds in this operational order:

1. create contracts, fixtures, and conformance tests;
2. introduce a versioned schema and idempotent migrations;
3. make the receiver and health checks available without unauthenticated public exposure;
4. connect the Go SDK and demonstration application;
5. enable queries and the Lounge over the same persisted data;
6. enable retention and validate under load before considering the change available.

Each database migration has a rollback path while it has not removed data. After a retention cycle removes expired data, rollback preserves the previous schema but cannot restore deleted data; this consequence must be explicit in configuration.
