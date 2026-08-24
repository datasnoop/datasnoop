#!/usr/bin/env bash
set -euo pipefail

go test ./apps/api/internal/ingestion -run 'TestProcessorReturnsRetryableCapacityErrorsAtIndependentBounds|TestProcessorAppliesPersistenceTimeout'
go test ./apps/api/internal/ingestion/receiver -run 'TestOTLPExportRejectsMissingOrInvalidCredentialsWithoutPassingRequestsToSink|TestOTLPPartialSuccessAndWhollyInvalidMappings'
go test ./apps/api/internal/live
go test ./apps/api/internal/platform/retention -run 'TestRunnerRecordsFailureAndScheduleHasBound|TestConcurrentRetentionDoesNotConsumeIngestionBudget'
go test ./sdk/go -run 'TestBuffer|Test.*Shutdown'
