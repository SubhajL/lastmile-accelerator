SVC=observability-service
TAG?=dev

.PHONY: build test lint coverage image sbom run clean budget proto proto.clean proto.verify

build:
	go build -o bin/app ./cmd/$(SVC)

test:
	go test ./... -v -cover

lint:
	golangci-lint run -c .golangci.yaml -D errcheck -D staticcheck -D ineffassign -D unused ./...

lint.ci:
	@golangci-lint version || true
	golangci-lint run -c .golangci.yaml -D errcheck -D staticcheck -D ineffassign -D unused ./...

# Budget checks: build size budget (<= 150MB)
budget:
	@set -e; \
	bin="bin/budget-app"; \
	GOOS=linux GOARCH=amd64 go build -o $$bin ./cmd/$(SVC); \
	size_kb=$$(du -k "$$bin" | cut -f1); \
	echo "budget: binary size=$$size_kb KB (limit 150000 KB)"; \
	[ $$size_kb -lt 150000 ]

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

image:
	docker build -t ghcr.io/${GITHUB_REPOSITORY}-$(SVC):$(TAG) .

sbom:
	syft dir:. -o spdx-json=sbom-$(SVC).spdx.json || true

run:
	SERVICE_PORT=7301 ./bin/app

clean:
	rm -rf bin/ coverage.out coverage.html

# ---- Protobuf codegen ----
# Generates Go stubs for packages/contracts/observability/observability.proto
# Requires: protoc v29+, Go toolchain; auto-installs protoc plugins locally.
PROTOC_INC=$(shell dirname $(shell dirname $(shell which protoc)))/include
PROTO_SRC=../../packages/contracts/observability/observability.proto
PROTO_OUT_DIR=internal/gen/observability/v1

proto:
	@set -e; \
	GOBIN=$$(go env GOPATH)/bin; \
	GOFLAGS=-mod=mod go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.2; \
	GOFLAGS=-mod=mod go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.4.0; \
	mkdir -p $(PROTO_OUT_DIR); \
	PATH="$$GOBIN:$$PATH" protoc \
		-I ../../packages/contracts \
		-I "$(PROTOC_INC)" \
		--go_out=$(PROTO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_SRC)

proto.clean:
	rm -rf internal/gen

proto.verify: proto
	@echo "proto: generation succeeded; please ensure generated files are committed"
