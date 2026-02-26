SVC=fix-engine-service
TAG?=dev

.PHONY: build test image sbom run
build:
	bun run build
test:
	bun run test
image:
	docker build -t ghcr.io/${GITHUB_REPOSITORY}-$(SVC):$(TAG) .
sbom:
	syft dir:. -o spdx-json=sbom-$(SVC).spdx.json || true
run:
	SERVICE_PORT=7151 ./bin/server
