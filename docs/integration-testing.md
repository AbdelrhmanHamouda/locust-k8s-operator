# Testing Guide

This document describes the comprehensive testing setup for the Locust K8s Operator, covering unit tests, integration tests (envtest), and end-to-end tests.

## Overview

The operator uses a multi-layered testing strategy:

| Test Type | Framework | Scope | Speed |
|-----------|-----------|-------|-------|
| **Unit Tests** | Go testing | Individual functions | Fast (~seconds) |
| **Integration Tests** | envtest | Controller + API Server | Medium (~30s) |
| **E2E Tests** | Ginkgo + Kind | Full cluster deployment | Slow (~5-10min) |

## Test Structure

```
locust-k8s-operator/
├── api/
│   ├── v1/
│   │   └── *_test.go              # v1 API tests
│   └── v2/
│       ├── *_test.go              # v2 API tests
│       └── locusttest_webhook_test.go  # Webhook validation tests
├── internal/
│   ├── config/
│   │   └── config_test.go         # Configuration tests
│   ├── controller/
│   │   ├── suite_test.go          # envtest setup
│   │   ├── locusttest_controller_test.go  # Unit tests
│   │   └── integration_test.go    # Integration tests
│   └── resources/
│       ├── job_test.go            # Job builder tests
│       ├── service_test.go        # Service builder tests
│       ├── labels_test.go         # Label builder tests
│       ├── env_test.go            # Environment builder tests
│       └── command_test.go        # Command builder tests
└── test/
    └── e2e/
        ├── e2e_suite_test.go      # E2E test setup
        └── e2e_test.go            # E2E test scenarios
```

## Prerequisites

- **Go 1.23+**: Required for running tests
- **Docker**: Required for E2E tests (Kind)
- **Kind**: Required for E2E tests

## Running Tests

### Unit & Integration Tests (envtest)

The primary test command runs both unit tests and integration tests using envtest:

```bash
# Run all tests with coverage
make test

# Run tests with verbose output
go test ./... -v

# Run specific package tests
go test ./internal/resources/... -v
go test ./internal/controller/... -v
go test ./api/v2/... -v

# Run specific test by name
go test ./internal/controller/... -v -run TestReconcile

# Generate coverage report
make test
go tool cover -html=cover.out -o coverage.html
```

### E2E Tests (Kind)

End-to-end tests run against a real Kubernetes cluster using Kind:

```bash
# Run E2E tests (creates Kind cluster automatically)
make test-e2e

# Run E2E tests with verbose output
KIND_CLUSTER=locust-test go test ./test/e2e/ -v -ginkgo.v

# Cleanup E2E test cluster
make cleanup-test-e2e
```

### CI Pipeline

All tests run automatically in GitHub Actions:

```bash
# Run the same checks as CI locally
make ci

# This runs:
# - make lint (golangci-lint)
# - make test (unit + integration tests)
```

## Test Fixtures

Test fixtures and sample data are located in:

- `internal/testdata/` - Test fixtures for unit tests
- `config/samples/` - Sample CRs for integration/E2E tests

## Troubleshooting

### Common Issues

#### envtest Binary Issues
```bash
# Re-download envtest binaries
make setup-envtest

# Verify binaries are installed
ls bin/k8s/
```

#### Test Timeouts
```bash
# Increase timeout for slow systems
go test ./... -v -timeout 10m
```

#### Kind Cluster Issues
```bash
# Check if cluster exists
kind get clusters

# Delete and recreate
kind delete cluster --name locust-k8s-operator-test-e2e
make test-e2e
```

### Debug Mode

Run tests with verbose logging:
```bash
# Verbose test output
go test ./internal/controller/... -v -ginkgo.v

# With debug logs from controller
go test ./internal/controller/... -v -args -zap-log-level=debug
```

## Writing New Tests

### Guidelines

1. **Unit tests**: Test pure functions in isolation
2. **Integration tests**: Test controller behavior with envtest
3. **E2E tests**: Test user-facing scenarios in real cluster

### Test Naming Conventions

```go
// Unit tests: Test<FunctionName>_<Scenario>
func TestBuildMasterJob_WithEnvConfig(t *testing.T) {}

// Integration tests: Describe/Context/It
Describe("LocustTest Controller", func() {
    Context("When creating a LocustTest", func() {
        It("Should create master Job", func() {})
    })
})
```

### Adding Integration Tests

1. Add test to `internal/controller/integration_test.go`
2. Use `k8sClient` for Kubernetes operations
3. Use `Eventually` for async assertions
4. Clean up resources in `AfterEach`

## Related Documentation

- [Local Development](local-development.md) - Development setup
- [Contributing](contribute.md) - Contribution guidelines
- [Pull Request Process](pull-request-process.md) - PR workflow
