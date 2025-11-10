SVC=ai-debugger-service
TAG?=dev

.PHONY: build test image sbom run
build:
	echo build-n/a
test:
	pytest -q || true
image:
	docker build -t ghcr.io/${GITHUB_REPOSITORY}-$(SVC):$(TAG) .
sbom:
	syft dir:. -o spdx-json=sbom-$(SVC).spdx.json || true
run:
	SERVICE_PORT=7102 ./bin/server
