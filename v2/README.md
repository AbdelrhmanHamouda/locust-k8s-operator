# Locust K8s Operator v2.0 - Go Migration

This directory contains planning and analysis documentation for migrating the Locust Kubernetes Operator from Java to Go.

## Overview

**Goal:** Migrate from Java (Micronaut + JOSDK) to Go (Operator SDK) while adding new features.  
**Estimated Effort:** 13-18 days  
**Status:** Planning

## Documents

| Document | Description |
|----------|-------------|
| [ROADMAP.md](./ROADMAP.md) | Phased implementation plan with tasks and acceptance criteria |
| [REQUIREMENTS.md](./REQUIREMENTS.md) | Feature requirements specification |

### Analysis

| Document | Description |
|----------|-------------|
| [analysis/ASSESSMENT.md](./analysis/ASSESSMENT.md) | Migration viability assessment |
| [analysis/TECHNICAL.md](./analysis/TECHNICAL.md) | Detailed technical analysis and component mapping |
| [analysis/LOCUST_FEATURES.md](./analysis/LOCUST_FEATURES.md) | New Locust features to support |

### Research

Deep-dive research documents on specific technical areas:

- [research/CONTROLLER_RUNTIME_DEEP_DIVE.md](./research/CONTROLLER_RUNTIME_DEEP_DIVE.md)
- [research/CRD_API_DESIGN.md](./research/CRD_API_DESIGN.md)
- [research/JAVA_TO_GO_MAPPING.md](./research/JAVA_TO_GO_MAPPING.md)
- [research/OPERATOR_SDK_PATTERNS.md](./research/OPERATOR_SDK_PATTERNS.md)
- [research/RESOURCE_MANAGEMENT.md](./research/RESOURCE_MANAGEMENT.md)

### Phases

Detailed planning for individual roadmap phases will be added to [phases/](./phases/) as work progresses.

## Quick Links

- **Critical Path:** Phases 0→1→2→3→4→5→6→13→14
- **New Features:** Phases 7-12 (can be parallelized after Phase 6)
