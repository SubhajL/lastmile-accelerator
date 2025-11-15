SVC=db-guardian-service
TAG?=dev

.PHONY: build test image sbom run migrate-up migrate-down migrate-status migrate-create bench perf
build:
	go build -o bin/app ./cmd/$(SVC)
test:
	go test ./... -cover || true
verify:
	@echo "Running go vet..." && go vet ./...
	@echo "Running unit tests..." && go test ./... -cover
	@echo "Verify complete"
bench:
	go test -run '^$$' -bench . -benchmem ./internal/analyzer
perf:
	@echo "Running analyzer benchmarks..." && \
	go test -run '^$$' -bench . -benchmem ./internal/analyzer | tee tmp/bench.txt || true
image:
	docker build -t ghcr.io/${GITHUB_REPOSITORY}-$(SVC):$(TAG) .
sbom:
	syft dir:. -o spdx-json=sbom-$(SVC).spdx.json || true
run:
	SERVICE_PORT=7105 ./bin/app

grpc-codegen:
	protoc -I api/proto \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/dbguardian/v1/dbguardian.proto

# Migration targets
migrate-up:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL not set"; \
		exit 1; \
	fi
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL not set"; \
		exit 1; \
	fi
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION not set (e.g., VERSION=0)"; \
		exit 1; \
	fi
	goose -dir migrations postgres "$(DATABASE_URL)" down-to $(VERSION)

migrate-status:
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "ERROR: DATABASE_URL not set"; \
		exit 1; \
	fi
	goose -dir migrations postgres "$(DATABASE_URL)" status

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "ERROR: NAME not set (e.g., NAME=add_users_table)"; \
		exit 1; \
	fi
	goose -dir migrations create $(NAME) sql
