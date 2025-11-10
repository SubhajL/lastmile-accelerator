> Global Rules (must‑read)
> - Follow the root guidelines: ../../AGENTS.md#ai-assisted-programming-guidelines-by-sabrina-ramonov
> - Also see: ../../WARP.md
> - Use shortcuts: QNEW, QPLAN, QCODE, QCHECK/QCHECKF/QCHECKT, QUX, QGIT (see root AGENTS.md)

## Docs Links
- Active Spec: ../../documentation/features/active/ai-debugger-service-spec.md
- Progress:    ../../documentation/features/active/ai-debugger-service-progress.md
- Planned:     ../../documentation/features/planned/ai-debugger-service-spec.md

## Package Identity

- Python service for AI debugging (REST 7102).

## Setup & Run

```bash
python3 -m venv .venv && . .venv/bin/activate && pip install -r requirements.txt
make test   # pytest
make run
```

## Patterns & Conventions

- Code under `app/**`.
- Tests: pytest style under `tests/**` if present.

## Touch Points

- `app/`, `requirements.txt`, `Dockerfile`, `Makefile`, Helm templates

## JIT Index Hints

```bash
grep -R -n "def test_" tests || true
```

## Pre-PR Checks

```bash
make test
```
