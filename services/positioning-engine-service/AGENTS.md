> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/positioning-engine-service-spec.md
- Progress:    ../../documentation/features/active/positioning-engine-service-progress.md
- Planned:     ../../documentation/features/planned/positioning-engine-service-spec.md

## Package Identity

- Python positioning engine (REST 7501).

## Setup & Run

```bash
python3 -m venv .venv && . .venv/bin/activate && pip install -r requirements.txt
make test || true
make run
```

## Patterns & Conventions

- Code under `app/**`; add pytest tests in `tests/**` as needed.

## Touch Points

- `app/`, `requirements.txt`, `Makefile`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build | true && make build
```
