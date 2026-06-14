# ADR 0004: Hide OpenTelemetry Behind the Default Product UX

## Status

Accepted

## Context

OTLP provides interoperability, but OpenTelemetry terminology and configuration would increase the learning curve for an audience seeking one-shot monitoring.

## Decision

The default integration is presented in terms of services, requests, operations, logs, errors, and measurements. Official SDKs configure exporters, resources, propagation, batching, and safe defaults. Direct OTLP remains an advanced path for users with existing instrumentation.

The Lounge may expose technical details in advanced areas, but its primary navigation does not require knowledge of spans, scopes, temporality, or Collector.

## Consequences

**Positive:** simple adoption without losing interoperability and the option for advanced configuration.

**Negative:** an additional conceptual translation layer, a need for clear diagnostics when instrumentation is partial, and a risk of hiding details required for troubleshooting.

## Alternatives considered

- Expose OpenTelemetry directly: less translation but transfers complexity to users.
- Fully proprietary protocol and SDK: controls the UX but increases lock-in and maintenance.

## Follow-ups

Build onboarding diagnostics that report connected sources and missing integrations without requiring manual inspection of OpenTelemetry configuration.
