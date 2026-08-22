# DataSnoop

DataSnoop is an open-source, self-hosted observability platform designed for independent developers and small teams running cost-conscious infrastructure.

It aims to shorten the path from noticing a production incident to finding its likely cause without requiring users to assemble a fragmented monitoring stack, write a query language, or understand OpenTelemetry internals.

> [!IMPORTANT]
> DataSnoop is currently in the planning and architecture stage. The repository contains validated product specifications and an implementation plan, but no runnable product yet.

## The first investigation journey

The initial product slice focuses on one service and one concrete diagnostic flow:

```text
An endpoint is failing
        │
        ▼
Find the affected endpoint
        │
        ▼
Open a failing operation
        │
        ▼
Inspect correlated logs
        │
        ▼
Compare with CPU, memory, and disk context
```

The default experience should let a user complete this journey without manually configuring OpenTelemetry providers, deploying a Collector, creating dashboards, or learning a telemetry query language.

## Product direction

DataSnoop is planned around five durable capabilities:

- **Telemetry ingestion:** authenticated OTLP/gRPC reception with explicit limits, partial acceptance, and backpressure.
- **Reference Go instrumentation:** opinionated `net/http` and `slog` instrumentation built on OpenTelemetry.
- **Request investigation:** endpoint error overview, failing operation details, and correlated logs.
- **Host monitoring:** time-aligned CPU, memory, and disk context for the monitored service.
- **Platform operations:** health diagnostics, ingestion visibility, and bounded telemetry retention.

The Go SDK is a reference and dogfooding implementation, not a boundary around the target audience. OTLP compatibility is intended to keep the backend accessible to other language ecosystems, while future official SDKs provide simpler, idiomatic onboarding.

## Architecture principles

- Deliver a complete user outcome before broad component coverage.
- Keep OpenTelemetry behind product-oriented language in the default UX.
- Use open standards at the external boundary.
- Keep instrumentation failures isolated from the monitored application.
- Bound memory, concurrency, queues, connections, and retention explicitly.
- Preserve future extensibility without implementing speculative abstractions.
- Treat historical persisted data as authoritative when live delivery is interrupted.

The planned ingestion path separates the external protocol from DataSnoop's internal domain:

```text
OTLP exporters
      │
      ▼
Authenticated OTLP/gRPC receiver
      │
      ▼
Validation and normalization
      │
      ▼
Bounded ingestion pipeline
      │
      ▼
PostgreSQL + TimescaleDB
      │
      ├── Historical investigation
      └── Post-persistence live updates
```

## Current scope boundaries

The active plan intentionally does not include:

- distributed tracing across services;
- market-oriented TypeScript, Python, or PHP SDKs;
- advanced Boolean query syntax;
- multiple alert integrations;
- OTLP/HTTP ingestion;
- alternative analytical storage;
- multi-node coordination or high availability.

These are tracked as opportunities with explicit dependencies and resumption triggers in [the opportunity register](docs/roadmap/opportunities.md).

## Repository map

```text
.
├── AGENTS.md                  Agent workflow and repository rules
├── docs/
│   ├── product/              Vision, principles, capabilities, and glossary
│   ├── decisions/            Architectural decision records
│   └── roadmap/              Current state and deferred opportunities
└── openspec/
    ├── config.yaml           OpenSpec context and artifact rules
    ├── specs/                Archived durable behavior specifications
    └── changes/              Active change proposals and implementation plans
```

Start with these documents:

1. [Product vision](docs/product/vision.md)
2. [Product and engineering principles](docs/product/principles.md)
3. [Capability map](docs/product/capability-map.md)
4. [Current state](docs/roadmap/current.md)
5. [Active OpenSpec proposal](openspec/changes/investigate-single-service-errors/proposal.md)
6. [Technical design](openspec/changes/investigate-single-service-errors/design.md)
7. [Implementation tasks](openspec/changes/investigate-single-service-errors/tasks.md)

## Development workflow

DataSnoop uses OpenSpec for spec-driven development and a Ralph-inspired loop for persistent, incremental agent work.

Each implementation iteration must:

1. read the canonical product and roadmap context;
2. select the first unblocked task;
3. complete a small, verifiable increment;
4. run the relevant validation;
5. record evidence and the next step in persistent repository state.

Validate the active plan with:

```bash
openspec status --change investigate-single-service-errors
openspec validate investigate-single-service-errors --strict
```

All repository content and source code must be written in English. Agent-user interaction may use the user's preferred language.

## Project status

The first OpenSpec change, [`investigate-single-service-errors`](openspec/changes/investigate-single-service-errors/), has complete and strictly validated planning artifacts. Its project-layout foundation is in place, but no runnable product exists yet.
