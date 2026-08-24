## Purpose

Provide time-aligned host resource context so users can distinguish application failures from CPU, memory or disk pressure during an investigation.

## ADDED Requirements

### Requirement: Host resource measurements
The system SHALL accept and retain timestamped CPU, memory and disk measurements with stable host and service association.

#### Scenario: Host metrics arrive for a service
- **WHEN** supported host measurements arrive with service and host identity
- **THEN** they are associated with that service and host at their source timestamps

#### Scenario: Host identity changes
- **WHEN** telemetry for the same service arrives from a different host identity
- **THEN** the system keeps the hosts distinguishable rather than merging their resource measurements

### Requirement: Time-aligned incident context
The Lounge SHALL present available host measurements covering the selected operation's surrounding time window and SHALL identify gaps or stale data.

#### Scenario: Resource pressure overlaps an error
- **WHEN** host measurements show elevated resource usage around a selected error operation
- **THEN** the investigation displays the relevant measurements aligned to the occurrence timestamp without claiming causation

#### Scenario: Metrics are unavailable
- **WHEN** no sufficiently recent host measurement exists for the selected operation
- **THEN** the Lounge labels host context as unavailable or stale instead of displaying a healthy default

### Requirement: Source health distinction
The system SHALL distinguish host metrics of the monitored application from operational metrics of the DataSnoop installation.

#### Scenario: Application and platform share a machine
- **WHEN** application-host and DataSnoop-platform measurements originate from the same physical machine
- **THEN** the Lounge labels their source roles so users can distinguish monitored workload health from platform health
