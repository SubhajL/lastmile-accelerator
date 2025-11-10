> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/finops-service-spec.md
- Progress:    ../../documentation/features/active/finops-service-progress.md
- Planned:     ../../documentation/features/planned/finops-service-spec.md

## Package Identity

- Python FinOps service (REST 7401).

## Setup & Run

```bash
python3 -m venv .venv && . .venv/bin/activate && pip install -r requirements.txt
make test || true
make run
```

## Patterns & Conventions

- Code under `app/**`; add pytest tests in `tests/**`.

## Touch Points

- `app/`, `requirements.txt`, `Makefile`, Helm in `helm/templates/`

## Pre-PR Checks

```bash
make test && make build || make build
```
