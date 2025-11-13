# Error Mapping Polish — Implementation Progress Tracker

Last Updated: 2025-11-12
Specification: ./error-mapping-spec.md

## Overview
Consistent 4xx mapping for expected DB conflicts across write handlers.

## Phase Completion Summary

| Phase | Status | Completion | Notes |
|------:|:------:|:----------:|-------|
| Helper | ⏳ | 0% | Pending |
| Handler Integration | ⏳ | 0% | Pending |
| Tests | ⏳ | 0% | Pending |

## Current Tasks
- [ ] Create `src/errors/db.rs` with `map_db_error`
- [ ] Replace handler mappings
- [ ] Add tests for 23505/23503 scenarios

## What needs to be done next
Implement mapper and wire through handlers.

## Blockers/Issues
None yet.
