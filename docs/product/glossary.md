# Glossary

## DataSnoop

The complete ingestion, persistence, investigation, and operation platform.

## Lounge

The user interface for overview, investigation, live monitoring, alerts, and configuration. It is not the name of the GraphQL API or a specific technical layer.

## Service

A monitored application identified by name, environment, and resource attributes.

## Operation

A correlatable unit of work, such as an HTTP request or database query. It may be represented externally by a span, but the default UX uses product-domain language.

## Event

A discrete occurrence relevant to diagnosis. When associated with an operation, it may be transported as a span event; when independent, it may be transported as a structured record.

## Log

A textual or structured record with timestamp, severity, body, and attributes, optionally correlated with an operation.

## Metric

A numeric measurement over time, such as CPU, memory, request count, or aggregated duration.

## OTLP

The external telemetry transport contract. It is a technical foundation and an advanced integration option, not a conceptual prerequisite for default onboarding.

## Reference SDK

An implementation used for dogfooding, learning, contract tests, and demonstrating the desired experience. The initial Go SDK does not define the target market by itself.

## Documentation harness

The set of canonical sources and persistent state that lets agents and people continue work without depending on conversation history.
