# platform-operations Specification

## Purpose
Make the self-hosted DataSnoop installation diagnosable and storage-bounded so users can trust whether telemetry is flowing and prevent uncontrolled retention.

## Requirements

### Requirement: Platform health reporting
The system SHALL expose liveness and readiness separately and SHALL report whether required persistence and ingestion dependencies are available.

#### Scenario: Process is alive but persistence is unavailable
- **WHEN** the DataSnoop process is running and its required persistence dependency cannot accept work
- **THEN** liveness remains successful while readiness reports the dependency failure

#### Scenario: Platform is ready
- **WHEN** the process can accept authenticated telemetry and satisfy required storage operations
- **THEN** readiness reports success

### Requirement: Onboarding diagnostics
The Lounge SHALL show whether a service has connected and which expected signal categories have been received, without requiring inspection of OpenTelemetry configuration.

#### Scenario: Service sends operations but no logs
- **WHEN** a service has recently sent operations but no logs
- **THEN** onboarding diagnostics show operations as connected and logs as not observed, with an actionable integration hint

#### Scenario: No telemetry is received
- **WHEN** a configured service has not produced accepted telemetry within the documented interval
- **THEN** diagnostics distinguish missing data from platform ingestion or credential errors when that evidence is available

### Requirement: Configurable bounded retention
The system SHALL apply a documented retention duration to raw telemetry and SHALL remove expired data without requiring manual database maintenance.

#### Scenario: Telemetry expires
- **WHEN** stored raw telemetry becomes older than the configured retention duration
- **THEN** the system removes it during an automated retention cycle and records the cycle outcome

#### Scenario: Retention configuration changes
- **WHEN** an authorized user changes retention to a supported value
- **THEN** subsequent cleanup uses the new value and the Lounge displays the active policy

### Requirement: Ingestion outcome visibility
The platform SHALL expose accepted, rejected, throttled and locally dropped telemetry outcomes using bounded operational metrics.

#### Scenario: Records are rejected or throttled
- **WHEN** ingestion rejects invalid records or throttles work due to capacity
- **THEN** platform diagnostics reflect the corresponding outcome counts and reason category

#### Scenario: Operational counters restart
- **WHEN** the single-node process restarts and volatile counters reset
- **THEN** the platform identifies the restart boundary and does not present reset counters as a historical decrease in failures
