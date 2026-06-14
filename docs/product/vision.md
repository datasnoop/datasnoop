# DataSnoop Vision

## Product

DataSnoop is an open-source, self-hosted observability platform for independent developers and small teams that need to investigate production failures without accepting the unpredictable cost of a SaaS product or the operational burden of a fragmented stack.

Its central promise is to shorten the path between noticing an incident and finding its likely cause. The first product scenario is identifying an endpoint with a high incidence of errors and opening its correlated requests and logs with basic context from the monitored machine.

## Audience

The target audience includes independent developers, freelancers, and small teams running their own applications on cost-conscious infrastructure. They value control over data and cost but do not want to become observability specialists to obtain their first diagnosis.

The initial Go SDK is a reference implementation because it is close to the DataSnoop stack. It does not constrain the target audience. The product must receive telemetry from other ecosystems through open standards, and future SDKs should prioritize languages with broad reach in the target developer community.

## Desired experience

The default flow must be opinionated and short:

1. start DataSnoop;
2. connect or instrument an application;
3. generate or observe traffic;
4. see the service appear in the Lounge;
5. identify a problematic endpoint;
6. open an occurrence and read its correlated logs;
7. distinguish application evidence from CPU, memory, or disk pressure.

The user must not need to configure dashboards, write a query language, or understand OpenTelemetry to complete this flow.

## Signals and platform

The product territory includes logs, events, correlatable operations, and application or machine metrics. The Lounge provides overview, investigation, live monitoring, alerts, and platform configuration.

The initial scope is single-service. The architecture must preserve the identifiers, contracts, and boundaries required for future distributed observability without implementing that entire expansion in advance.

## Positioning

The West Coast hip-hop-inspired identity makes the product memorable but remains a layer over a trustworthy experience. Themed language may add personality to the interface; critical states, security, documentation, and operations must remain unambiguous and professional.
