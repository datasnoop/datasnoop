.PHONY: build test unit-test contract-test integration-test race-test fuzz-seed-test frontend-test system-smoke architecture-check format-check openspec-validate migration-verify index-verify persistence-verify demo-verify failure-injection-verify quality

build:
	go build ./apps/api/... ./sdk/go/ && cd apps/lounge && npm run build

test: unit-test contract-test frontend-test

unit-test:
	go test ./apps/api/... ./sdk/go/...

contract-test:
	cd apps/api && go test . -run '^TestOTLPFixtures|^TestRetransmissionFixture'

integration-test: migration-verify index-verify persistence-verify

race-test:
	bash scripts/verify-race.sh

fuzz-seed-test:
	bash scripts/verify-fuzz-seeds.sh

frontend-test:
	cd apps/lounge && npm test

format-check:
	test -z "$$(gofmt -l apps/api sdk/go)"
	cd apps/lounge && npm run format:check

openspec-validate:
	openspec validate investigate-single-service-errors --strict

migration-verify:
	bash scripts/verify-migrations.sh

index-verify:
	bash scripts/verify-indexes.sh

persistence-verify:
	bash scripts/verify-persistence.sh

demo-verify:
	go test ./sdk/go -run TestDemonstrationProducesCanonicalIncidentDataset

workload-verify:
	go test ./apps/api/internal/workload

workload-report-verify:
	go test ./apps/api/internal/workload -run TestReportsMakeOverloadAndRecoveryExplicit

failure-injection-verify:
	bash scripts/verify-failure-injection.sh

system-smoke:
	bash scripts/verify-system-smoke.sh

architecture-check:
	bash scripts/verify-architecture.sh

quality: format-check architecture-check test race-test fuzz-seed-test system-smoke openspec-validate
