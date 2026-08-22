# Opportunities and Follow-ups

Items in this document are valid but outside the current outcome. Their presence does not imply an implementation commitment.

## Adoption-oriented TypeScript/Node.js SDK

**State:** `deferred`

**Motivation:** reach APIs and small SaaS products in a broad ecosystem with idiomatic web framework and logging integrations.

**Why not now:** the Go SDK reduces learning cost and validates the contract during the first vertical slice.

**Dependencies:** validated OTLP contract; stable correlation semantics; SDK conformance harness; tested Go SDK experience.

**Resumption trigger:** validate the complete investigation flow and select the first market-oriented ecosystem through discovery.

## Python and PHP SDKs

**State:** `exploring`

**Motivation:** support Python APIs, data and AI workloads, and Laravel or economically hosted applications.

**Why not now:** there is not enough evidence to rank TypeScript, Python, and PHP by impact on the initial audience.

**Dependencies:** stable OTLP receiver; compatibility matrix; interviews or adoption signals.

**Resumption trigger:** demand evidence or an onboarding gap that generic OpenTelemetry exporters cannot address acceptably.

## Distributed tracing

**State:** `deferred`

**Motivation:** correlate operations across services and locate dependencies responsible for failures or latency.

**Why not now:** the initial single-service scenario can prove value through HTTP operations and correlated logs.

**Dependencies:** preserved trace identifiers; stable operation model; validated single-service investigation.

**Resumption trigger:** users running multiple services cannot locate causes through the single-service view.

## OTLP/HTTP ingestion

**State:** `deferred`

**Motivation:** compatibility with environments where gRPC or HTTP/2 is inconvenient, including some proxies, serverless platforms, and short-lived runtimes.

**Why not now:** one official transport path limits the initial operational surface.

**Dependencies:** transport-independent normalization pipeline; shared authentication policy.

**Resumption trigger:** adoption of a language or environment whose recommended exporter uses OTLP/HTTP, or recurring evidence of gRPC failures.

## Alerts and integrations

**State:** `planned`

**Motivation:** react to increased errors or resource pressure without continuously watching the Lounge.

**Why not now:** alerts depend on trustworthy signals and aggregations. A generic webhook should precede specific integrations.

**Dependencies:** validated derived metrics; persisted rules; deduplication and cooldown control.

**Resumption trigger:** the investigation flow receives trustworthy data and exposes a stable error condition for end-to-end validation.

## Durable downstream delivery

**State:** `deferred`

**Motivation:** guarantee delivery and replay for alerts, webhooks, integrations, or consumers running outside the DataSnoop process.

**Why not now:** the initial Live View is recoverable from authoritative history and does not require a transactional outbox or durable event bus.

**Dependencies:** stable post-commit event identities; alert or integration contracts; retry, retention, and dead-letter policies.

**Resumption trigger:** a downstream consumer requires delivery to survive a process failure between database commit and publication.

## Advanced Boolean queries

**State:** `deferred`

**Motivation:** allow grouped AND/OR filters over events and custom attributes.

**Why not now:** basic time, service, route, severity, status, and identifier filters cover the initial scenario.

**Dependencies:** attribute model; cardinality policy; filter language or AST; indexes driven by real measurements.

**Resumption trigger:** basic queries block investigations observed in tests or user feedback.

## Alternative analytical storage

**State:** `deferred`

**Motivation:** support volumes beyond the efficient PostgreSQL/TimescaleDB profile.

**Why not now:** a premature generic abstraction does not remove differences in querying, consistency, aggregation, and retention.

**Dependencies:** production benchmarks; stable query model; load profile that justifies additional operations.

**Resumption trigger:** measured cost or performance limits rather than theoretical projections.
