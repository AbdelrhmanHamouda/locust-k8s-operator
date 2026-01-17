# Phase 2: Configuration System

**Status:** Pending  
**Effort:** 0.5 day  
**Priority:** P0 - Critical Path  
**Dependency:** Phase 0 (Complete)

---

## Overview

Implement environment-based configuration matching Java `SysConfig.java`. This phase creates the configuration system that other phases (particularly Phase 3: Resource Builders) depend on for operator-wide settings like resource limits, metrics exporter configuration, and feature flags.

## Documents

| Document | Purpose |
|----------|---------|
| [IMPLEMENTATION.md](./IMPLEMENTATION.md) | Detailed step-by-step implementation guide |
| [CHECKLIST.md](./CHECKLIST.md) | Quick reference task checklist |

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Config Source** | Environment variables | Match K8s operator conventions, simpler than file-based config |
| **No DI Framework** | Explicit struct wiring | Go idiom; avoids unnecessary complexity |
| **Nullable TTL** | `*int32` pointer type | Distinguishes "not set" from "set to 0" |
| **Default Values** | In `LoadConfig()` function | Centralized, testable defaults |

## Source of Truth

- **Java Config:** `/src/main/java/com/locust/operator/controller/config/SysConfig.java`
- **Java Defaults:** `/src/main/resources/application.yml`

## Configuration Categories

1. **Job Configuration** - TTL after finished
2. **Pod Resources** - CPU/memory requests and limits for Locust containers
3. **Metrics Exporter** - Image, port, resources for sidecar container
4. **Kafka** - Bootstrap servers and security (optional integration)
5. **Feature Flags** - Affinity/tolerations CR injection toggles

## Dependencies

- **Upstream:** Phase 0 (project scaffolding)
- **Downstream:** Phase 3 (Resource Builders), Phase 4 (Core Reconciler)

## Acceptance Criteria

1. Config loads with defaults when env vars not set
2. Config respects env var overrides
3. All Java `application.yml` properties have Go equivalents
4. Unit tests cover default values and env var overrides
5. `*int32` pointer correctly handles nullable TTL

## References

- [ROADMAP.md](../../ROADMAP.md) - Phase 2 definition (lines 136-166)
- [REQUIREMENTS.md](../../REQUIREMENTS.md) - §3.1 Configuration
- [analysis/TECHNICAL.md](../../analysis/TECHNICAL.md) - §5.2 Configuration
- [research/JAVA_TO_GO_MAPPING.md](../../research/JAVA_TO_GO_MAPPING.md) - §3 Configuration Binding
