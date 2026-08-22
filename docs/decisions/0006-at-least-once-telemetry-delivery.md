# ADR 0006: At-Least-Once Telemetry Delivery

## Status

Accepted

## Context

An OTLP exporter can lose an acknowledgement after the receiver has accepted a request and can resend the same telemetry. The receiver therefore cannot promise exactly-once delivery at its external boundary. Retransmission must not corrupt persisted data or inflate the endpoint error summaries that drive the initial investigation.

Live SSE delivery has a different reliability role. It accelerates the interface but historical persisted data remains authoritative, so losing an in-memory notification must be recoverable without making every live event durable.

## Decision

Treat OTLP ingestion as at-least-once delivery. Document identity and duplicate behavior per supported signal instead of applying one synthetic payload hash to every record.

An HTTP operation with the same service, environment, trace identifier, and span identifier represents the same logical operation. Reingesting it is idempotent: it does not create another occurrence or increase endpoint Rate, Errors, and Duration summaries. When a supported signal has another reliable identity, its persistence rule may also be idempotent and must be documented in the support matrix. The telemetry-ingestion spec is the normative behavior contract.

Records without a reliable identity are accepted independently rather than being silently collapsed by content, timestamp, or payload hash. Diagnostics and compatibility documentation make this limitation visible.

Batch persistence uses short transactions and reports committed, repeated, rejected, and failed outcomes. OTLP partial-success and error responses remain governed by the protocol: partially accepted requests are not converted into retryable failures, while recoverable capacity or persistence failures return an explicitly retryable outcome.

Live references are published best-effort only after persistence commits. Per-client buffers are bounded, and a missed reference produces a gap that the Lounge reconciles from historical queries. The initial slice does not use a transactional outbox.

## Consequences

**Positive:** endpoint summaries remain stable across common OTLP retries, delivery guarantees match the external protocol, and live delivery does not add durable messaging infrastructure to the single-node path.

**Negative:** duplicate behavior differs by signal; records without reliable identity can remain duplicated; a process failure between commit and live publication can create a visible gap that requires historical reconciliation.

## Alternatives considered

- Exactly-once delivery: cannot be guaranteed across the OTLP acknowledgement boundary without a protocol-level idempotency contract.
- Payload-hash deduplication: can collapse legitimate repeated logs or measurements and hides semantic differences among signals.
- Transactional outbox for all live updates: closes the commit-to-publication gap but adds durable scheduling, cleanup, and replay before live delivery requires that guarantee.
- Publish before commit: lowers latency but can show telemetry that never becomes authoritative history.

## Follow-ups

- Add a durable outbox when alerts, webhooks, integrations, or another process require guaranteed downstream delivery.
- Revisit signal-specific identities when the supported OTLP subset expands.
