# Migration from Java to Go

## Overview

The Locust Kubernetes Operator has been completely rewritten from Java to Go. This represents a full architectural transition, not a simple port. While the core functionality remains the same, the implementation is fundamentally different.

## What Changed

### Language and Framework
- **Before**: Java with Micronaut framework and Java Operator SDK
- **After**: Go with Operator SDK / controller-runtime (Kubernetes standard operator framework)

### Project Structure
- **Before**: Java source in `src/`, Gradle build files, Maven dependencies
- **After**: Go source at repository root (`cmd/`, `api/`, `internal/`), Go modules

### Performance Characteristics
- **Memory footprint**: ~256MB (Java) → ~64MB (Go)
- **Startup time**: ~60s (Java) → <1s (Go)
- **Binary size**: ~100MB (Java + JVM) → ~30MB (Go static binary)

## What Stayed the Same

### API Compatibility
The `LocustTest` Custom Resource Definition (CRD) maintains full backward compatibility:
- Both v1 and v2 API versions are supported
- Existing manifests continue to work without modification
- Helm chart values remain compatible (with new optional features added)

### Behavior
The operator provides the same functionality:
- Creates master and worker Pods with Locust
- Manages Services for master UI and headless communication
- Handles ConfigMap-based Locust script injection
- Supports Secret and environment variable configuration

### Deployment
The Helm chart remains at `charts/locust-k8s-operator/` with the same installation process.

## Finding the Old Java Code

The Java operator source code has been preserved in the `archive/java-operator-v1` branch for reference.

To access it:

```bash
git fetch origin archive/java-operator-v1
git checkout archive/java-operator-v1
```

This branch contains the complete Java codebase as it existed before the Go rewrite. It is maintained for historical reference only and will not receive further updates.

## Key Differences for Developers

### Testing
- **Before**: JUnit 5, Mockito, Testcontainers
- **After**: Go's testing package, envtest for controller tests, Kind for E2E tests

### Build System
- **Before**: Gradle with multi-stage Dockerfile
- **After**: Make with standard Go build commands, multi-arch builds via BuildKit

### Dependencies
- **Before**: Maven Central packages
- **After**: Go modules from Go package ecosystem

### CI/CD
The CI/CD pipelines have been updated to use Go tooling:
- `go build`, `go test`, `go vet`, `golangci-lint` instead of Gradle tasks
- Multi-platform Docker builds (amd64 + arm64)
- Helm chart testing remains unchanged

## Migration for Users

Most users will not need to make any changes. The Go operator is a drop-in replacement for the Java operator:

1. Update the operator deployment via Helm (same chart, new appVersion)
2. Existing `LocustTest` resources continue to function
3. Review new features in v2.0 (OpenTelemetry, enhanced volumes, separate resources per role)

For detailed migration guidance, see the [Migration Guide](https://abdelrhmanhamouda.github.io/locust-k8s-operator/migration/) in the documentation.

## Rationale

The rewrite to Go was motivated by:

1. **Performance**: Lower memory usage and faster startup align with Kubernetes ecosystem expectations
2. **Ecosystem alignment**: controller-runtime is the de facto standard for Kubernetes operators
3. **Maintainability**: Simpler deployment (static binary), broader contributor pool familiar with Go
4. **Cloud-native fit**: Go is the lingua franca of cloud-native tooling
