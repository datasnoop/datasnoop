# ADR 0002: Go SDK as the Reference Implementation

## Status

Accepted

## Context

Go is the core stack and provides proximity for dogfooding, tests, debugging, and learning across the complete pipeline. However, Go developers do not necessarily represent the largest segment of the target audience.

## Decision

The initial Go SDK is the reference implementation for validating onboarding, HTTP instrumentation, log integration, batching, retries, shutdown, and safe failure. It exports OTLP and does not introduce a Go-specific protocol.

The backend must be testable with OTLP exporters that do not use the DataSnoop SDK.

## Consequences

**Positive:** fast learning cycles, shared stack, end-to-end demonstration, and contracts exercised through real use.

**Negative:** risk of confusing implementation convenience with market priority and leaking idiomatic Go choices into the external model.

## Alternatives considered

- Start with TypeScript: potentially broader reach but higher initial cost across two stacks.
- Provide no official SDK: lower maintenance but weaker one-shot onboarding.

## Follow-ups

Select adoption-oriented SDKs through discovery and evidence. TypeScript, Python, and PHP remain candidates, not a committed order.
