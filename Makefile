.PHONY: build test go-test frontend-test format-check openspec-validate migration-verify index-verify persistence-verify quality

build:
	go build ./apps/api/... ./sdk/go/ && cd apps/lounge && npm run build

test: go-test frontend-test

go-test:
	go test ./apps/api/... ./sdk/go/...

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

quality: format-check test openspec-validate
