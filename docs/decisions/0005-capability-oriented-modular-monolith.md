# ADR 0005: Capability-Oriented Modular Monolith

## Status

Accepted

## Context

The first DataSnoop outcome crosses telemetry ingestion, persistence, investigation, live delivery, host monitoring, platform operation, the Lounge, and the reference SDK. The product must preserve real boundaries for future growth while remaining simple to deploy and operate on a modest single node.

A repository organized as global transport, service, repository, and model layers would scatter each user journey across unrelated directories. A microservice architecture would add distributed coordination, failure modes, and deployment cost before the product has evidence that independent scaling is required.

## Decision

The DataSnoop backend is a single deployable modular monolith organized by durable product capabilities. Each capability owns its application behavior, domain concepts, ports, and infrastructure adapters.

Use vertical slices inside a capability for user-observable use cases such as listing endpoint summaries, opening an error occurrence, retrieving correlated logs, and obtaining host context. Keep a shared pipeline inside ingestion where authentication, validation, normalization, admission control, persistence, and outcome mapping only deliver value together.

Apply Ports and Adapters at actual I/O and compatibility boundaries. A port is owned and defined by the application behavior that consumes it, exposes only the required capability, and is implemented by an adapter such as OTLP/gRPC, PostgreSQL/TimescaleDB, GraphQL, SSE, host collection, or authentication. Pure domain behavior does not require an interface.

Keep shared code limited to stable cross-capability concepts such as telemetry identifiers, time ranges, and explicit application contracts. Modules communicate through those contracts instead of reaching into one another's internal packages or tables. Do not introduce global horizontal controller, service, repository, or model layers, a generic `Repository<T>`, or a dependency-injection framework.

## Consequences

**Positive:** one process and deployment unit, local transactions, capability-local changes, explicit compatibility boundaries, direct testing of use cases, and a credible path to extract a module only after evidence justifies it.

**Negative:** module boundaries require dependency discipline inside one process; some infrastructure is shared without pretending that it is portable; ingestion remains a coordinated pipeline rather than a set of independently deployable slices.

## Alternatives considered

- Global layered architecture: familiar initially but scatters vertical outcomes and encourages shared models.
- Microservices: offers independent deployment but adds network, consistency, and operational costs outside the current scope.
- A single global hexagon: separates infrastructure from a core but tends to centralize unrelated capabilities and oversized ports.
- Fully isolated vertical slices: maximizes locality but duplicates stable telemetry concepts and fragments the ingestion pipeline.

## Follow-ups

- Enforce package dependency direction through Go `internal` boundaries and automated architecture checks.
- Consider extracting a capability only after measured scaling, reliability, ownership, or release constraints require independent deployment.
