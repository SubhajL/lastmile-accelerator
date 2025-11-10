SVC=observability-service
TAG?=dev

.PHONY: build test lint coverage image sbom run clean

build:
	go build -o bin/app ./cmd/$(SVC)

test:
	go test ./... -v -cover

lint:
	golangci-lint run ./...

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
