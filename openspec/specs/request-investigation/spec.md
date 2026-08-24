# request-investigation Specification

## Purpose
Enable a user to move from an overview of failing endpoints to the operation and logs that explain a single-service incident without writing a query language.

## Requirements

### Requirement: Endpoint health overview
The Lounge SHALL show request rate, error rate and duration for normalized endpoints of a selected service and time window, and SHALL allow endpoints to be ordered by error impact.

#### Scenario: Endpoint accumulates server errors
- **WHEN** a normalized endpoint has operations with server error status in the selected window
- **THEN** the overview shows its request volume, error count or rate and duration summary and allows the user to select it

#### Scenario: User changes the time window
- **WHEN** the user selects a different supported time window
- **THEN** endpoint summaries are recalculated for that window and the chosen service

### Requirement: Error occurrence exploration
The Lounge SHALL list error operations for a selected endpoint and SHALL display each occurrence using product terminology, timestamp, status, duration and available correlation context.

#### Scenario: User opens a failing endpoint
- **WHEN** the user selects an endpoint with errors
- **THEN** the Lounge lists its error operations for the active service and time window

#### Scenario: User opens an occurrence
- **WHEN** the user selects one error operation
- **THEN** the Lounge displays its request details and attributes without requiring knowledge of spans or OTLP

### Requirement: Correlated log retrieval
The system SHALL retrieve logs correlated to a selected operation by valid trace and span context and SHALL distinguish them from nearby uncorrelated logs.

#### Scenario: Correlated logs exist
- **WHEN** a selected error operation has logs carrying matching correlation identifiers
- **THEN** those logs are presented in chronological order with severity, message and supported attributes

#### Scenario: No correlated logs exist
- **WHEN** a selected error operation has no matching logs
- **THEN** the Lounge states that no correlated logs were received and offers available operation context without implying that no logs occurred

### Requirement: Basic investigation filters
The investigation experience SHALL support filtering by service, time range, normalized route, HTTP status and log severity without requiring a query language.

#### Scenario: User filters an incident
- **WHEN** the user combines supported filters
- **THEN** endpoint, occurrence and log results consistently reflect all active filters

### Requirement: Live telemetry view
The Lounge SHALL provide a bounded live view of newly accepted logs and error operations and SHALL communicate interruption or overload to the viewer.

#### Scenario: New correlated error arrives
- **WHEN** the live view is connected and a new error operation with logs is accepted
- **THEN** the relevant items appear without a manual refresh and retain navigation to the investigation detail

#### Scenario: Live delivery is interrupted
- **WHEN** the live connection is lost or the viewer cannot keep up
- **THEN** the Lounge shows that live data may be incomplete and provides a recovery action

### Requirement: Investigation during supported ingest load
The system SHALL keep the historical incident journey available and semantically correct while receiving telemetry within a published workload profile for a declared hardware budget, and SHALL expose degradation rather than silently presenting incomplete data when that supported envelope is exceeded.

#### Scenario: User investigates during supported load
- **WHEN** telemetry matching a published supported profile is being ingested and a user follows the endpoint-to-occurrence-to-logs-to-host journey
- **THEN** endpoint summaries match the profile's known dataset and the occurrence, correlated logs and host context remain available within the published response budget

#### Scenario: Workload exceeds the supported envelope
- **WHEN** ingest or query demand exceeds the declared capacity and the system cannot keep the investigation current within its published budget
- **THEN** the Lounge indicates delayed, incomplete or degraded data and preserves a recovery path instead of presenting stale results as complete
