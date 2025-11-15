SVC=observability-service
TAG?=dev

.PHONY: build test lint coverage image sbom run clean budget

build:
	go build -o bin/app ./cmd/$(SVC)

test:
	go test ./... -v -cover

lint:
	golangci-lint run ./...

lint.ci:
	@golangci-lint version || true
	golangci-lint run ./...

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
