SVC=billing-service
TAG?=dev

.PHONY: build test image sbom run
build:
	go build -o bin/app ./cmd/$(SVC)
test:
	go test ./... -cover || true
image:
	docker build -t ghcr.io/${GITHUB_REPOSITORY}-$(SVC):$(TAG) .
sbom:
	syft dir:. -o spdx-json=sbom-$(SVC).spdx.json || true
run:
	SERVICE_PORT=7901 ./bin/app
