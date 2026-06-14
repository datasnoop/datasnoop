# ADR 0001: OTLP as the External Ingestion Contract

## Status

Accepted

## Context

DataSnoop must receive logs, metrics, and operations from different languages without requiring a complete proprietary implementation for every ecosystem. Correlation, versioning, and interoperability are difficult to recover after public SDKs adopt a contract.

## Decision

OTLP is the preferred external telemetry ingestion contract. The initial official path uses OTLP/gRPC. The receiver normalizes accepted signals into an internal DataSnoop model; OTLP does not directly define the persistence schema or Lounge UX.

The supported subset and behavior for partially or entirely incompatible data must be explicit. Compatibility must never mean silent dropping.

## Consequences

**Positive:** reuse of existing instrumentation, multilingual expansion, standardized correlation identifiers, reduced lock-in, and future Collector compatibility.

**Negative:** broader input surface, complex metric semantics, cardinality controls, and a required partial-acceptance policy.

## Alternatives considered

- Proprietary Protobuf: smaller initial surface but requires custom exporters and contracts.
- Proprietary HTTP/JSON: conceptually simple onboarding but less efficient and interoperable.
- Mandatory Collector: delegates operational capabilities but violates the initial one-shot experience.

## Follow-ups

- Add OTLP/HTTP when compatibility needs are demonstrated.
- Document Collector compatibility.
- Build a conformance harness independent from official SDKs.
