SVC=dep-governance-service
TAG?=dev

.PHONY: build test image sbom run
build:
	cargo build --release
test:
	cargo test --all --locked || true
image:
	docker build -t ghcr.io/${GITHUB_REPOSITORY}-$(SVC):$(TAG) .
sbom:
	syft dir:. -o spdx-json=sbom-$(SVC).spdx.json || true
run:
	SERVICE_PORT=7106 ./target/release/dep_governance_service
