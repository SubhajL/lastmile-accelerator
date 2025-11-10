# Documentation

Lightweight, token-efficient docs for humans. Per-service work lives under `features/active` as two files: `*-spec.md` and `*-progress.md`.

## Structure

```
documentation/
  README.md
  features/
    active/            # live features by service
    planned/           # future work (optional)
    completed/         # archived results
  architecture/        # system diagrams, decisions (optional)
  fixes/               # postmortems and fixes (optional)
```

## Conventions

- One spec/progress pair per service (feature-name = service folder name under `services/*`).
- Keep content short; link to code and `services/<name>/AGENTS.md` and `service_catalog.yaml` for details.
- Prefer checklists and links over prose.

## Editing

- Update the `*-progress.md` as milestones complete.
- Move finished items to `features/completed/` when done.
