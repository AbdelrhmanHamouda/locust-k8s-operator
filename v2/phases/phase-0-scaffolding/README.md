# Phase 0: Project Scaffolding

**Status:** Complete  
**Effort:** 0.5 day  
**Priority:** P0 - Critical Path

---

## Overview

Initialize the Go operator project using Operator SDK. This is the foundation phase that all other phases depend on.

## Documents

| Document | Purpose |
|----------|---------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Detailed step-by-step implementation guide |
| [CHECKLIST.md](./CHECKLIST.md) | Quick reference task checklist |

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Project Location** | `locust-k8s-operator-go/` | Keep separate from Java during migration |
| **SDK Plugin** | `go/v4` | Latest stable Go plugin |
| **Domain** | `locust.io` | Match existing CRD API group |
| **Initial API** | v1 only | v2 added in Phase 7 |

## Dependencies

- **Upstream:** None (Phase 0 has no dependencies)
- **Downstream:** All other phases depend on Phase 0

## References

- [ROADMAP.md](../../ROADMAP.md) - Phase 0 definition
- [REQUIREMENTS.md](../../REQUIREMENTS.md) - §3.1 Technology Stack, §3.4 Project Structure
- [analysis/TECHNICAL.md](../../analysis/TECHNICAL.md) - §2.2 Go Target Structure
- [research/OPERATOR_SDK_PATTERNS.md](../../research/OPERATOR_SDK_PATTERNS.md) - §2 Project Scaffolding
