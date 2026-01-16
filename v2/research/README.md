# Go Migration Research

This directory contains domain research findings for migrating the Locust Kubernetes Operator from Java to Go using the Operator SDK framework.

## Contents

| Document | Description |
|----------|-------------|
| [OPERATOR_SDK_PATTERNS.md](./OPERATOR_SDK_PATTERNS.md) | Core patterns and best practices for Go-based operators |
| [CONTROLLER_RUNTIME_DEEP_DIVE.md](./CONTROLLER_RUNTIME_DEEP_DIVE.md) | controller-runtime internals and reconciliation patterns |
| [TESTING_STRATEGIES.md](./TESTING_STRATEGIES.md) | Testing approaches for Kubernetes operators |
| [CRD_API_DESIGN.md](./CRD_API_DESIGN.md) | API versioning, conversion webhooks, and schema design |
| [RESOURCE_MANAGEMENT.md](./RESOURCE_MANAGEMENT.md) | Kubernetes resource creation patterns in Go |
| [JAVA_TO_GO_MAPPING.md](./JAVA_TO_GO_MAPPING.md) | Detailed mapping of Java patterns to Go equivalents |

## Research Methodology

1. **Codebase Analysis** - Reviewed existing Java operator implementation
2. **Framework Research** - Studied Operator SDK and controller-runtime documentation
3. **Pattern Extraction** - Identified common patterns in production Go operators
4. **Gap Analysis** - Compared Java JOSDK patterns with Go equivalents

## Key Findings Summary

### Framework Selection: Operator SDK ✅

The Operator SDK is recommended over bare Kubebuilder for the following reasons:
- Built on Kubebuilder foundation with additional enterprise features
- OLM integration for OperatorHub distribution
- Comprehensive E2E testing framework
- Strong alignment with Java JOSDK philosophy

### Migration Complexity: LOW to MEDIUM

| Component | Complexity | Notes |
|-----------|------------|-------|
| CRD Types | Low | Direct Go struct mapping |
| Reconciler | Low | Simple stateless logic |
| Resource Builders | Medium | Different API style |
| Testing | Medium | Different mocking approach |
| Configuration | Low | Environment-based config |

### Critical Differences from Java

1. **No Dependency Injection** - Explicit struct composition
2. **No Builder Pattern** - Struct literals preferred
3. **Error Handling** - Explicit error returns vs exceptions
4. **Generics** - Limited vs extensive use in Java
5. **Logging** - Structured logging with logr/slog

## Research Date
January 2026

## Related Documents
- `../analysis/ASSESSMENT.md` - Viability assessment
- `../analysis/TECHNICAL.md` - Technical migration guide
- `../REQUIREMENTS.md` - Feature requirements
- `../analysis/LOCUST_FEATURES.md` - Locust feature analysis
