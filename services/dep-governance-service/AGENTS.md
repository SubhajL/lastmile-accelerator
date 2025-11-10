> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/dep-governance-service-spec.md
- Progress:    ../../documentation/features/active/dep-governance-service-progress.md
- Planned:     ../../documentation/features/planned/dep-governance-service-spec.md

## Package Identity

- Rust service for dependency governance (REST 7106).

## Setup & Run

```bash
cargo build --release
cargo test --all --locked
make run
```

## Patterns & Conventions

- Code in `src/**` (e.g., `src/main.rs`, `src/handlers/**`, `src/services/**`).
- Integration tests in `tests/**` (e.g., `tests/http_dependencies_test.rs`).
- DB migrations under `migrations/**`.

## Touch Points

- `Cargo.toml`, `src/main.rs`, `src/services/*`, `tests/**`, `migrations/**`
- Deploy with Helm in `helm/templates/`

## JIT Index Hints

```bash
grep -R -nE '^\s*#\[test\]' src tests
```

## Pre-PR Checks

```bash
cargo test --all --locked && cargo build --release
```
