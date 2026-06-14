# Product and Engineering Principles

## Vertical value before horizontal completeness

A delivery must close a user-observable outcome. Technically complete components disconnected from the diagnostic journey do not prove the product.

## Simplicity before technical exposure

The default user does not need to understand OTLP, exporters, collectors, spans, metric temporality, or storage details. The interface and SDKs translate these concepts into services, requests, logs, measurements, and errors.

## Open standards at the boundary

Interoperable protocols reduce lock-in and accelerate ecosystem growth. OTLP is the preferred external telemetry contract, while DataSnoop remains free to normalize and persist signals according to its queries.

## SDKs as experience, not barriers

Official SDKs provide safe defaults, idiomatic instrumentation, and installation diagnostics. Applications capable of emitting OTLP do not depend on a proprietary SDK to use the backend.

## Extensibility without premature implementation

We version contracts, preserve correlation, and isolate real responsibilities. We do not create generic abstractions or future capabilities without a current use case and a resumption criterion.

## Safe failure for the monitored application

Instrumentation must never crash or block the application indefinitely. Queues are bounded; timeouts, controlled dropping, retries, and shutdown have explicit and observable behavior.

## Predictable operation on modest hardware

Memory, concurrency, buffers, retention, and connections have limits. Product resilience is demonstrated through tests and controlled degradation, not estimates presented as guarantees.

## One canonical source per knowledge type

Vision, decisions, behavior, execution, and follow-ups belong in distinct documents. We reference canonical sources instead of duplicating divergent narratives.

## English repository output

All persisted repository content is written in English so documentation, specifications, code, tests, and agent state remain consistent and broadly accessible. Source code follows the same rule for identifiers, packages, modules, APIs, database schemas, comments, errors, logs, and test descriptions. Agents may converse in any language without changing this output requirement.
