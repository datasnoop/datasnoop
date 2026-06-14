# ADR 0003: Single-Service Investigation First

## Status

Accepted

## Context

The product is intended to evolve across multiple observability scenarios, but its first proof must close a useful diagnosis with controlled scope.

## Decision

The first experience investigates one service: identify error-prone endpoints, open an occurrence, correlate operations and logs, and present basic host context. Trace-compatible identifiers are preserved, but distributed visualization and analysis remain outside the current outcome.

## Consequences

**Positive:** smaller surface, demonstrable success criteria, and room to validate the model before distributed relationships.

**Negative:** users with distributed architectures receive incomplete context, and some OTLP structures arrive before the UI can fully explore them.

## Alternatives considered

- Distributed tracing from the start: more complete visibility but high modeling and UX cost before the basic journey is validated.
- Logs only: smaller implementation but no automatic identification of problematic endpoints.

## Follow-ups

Promote distributed tracing after validating the single-service journey and observing real diagnoses blocked by missing service relationships.
