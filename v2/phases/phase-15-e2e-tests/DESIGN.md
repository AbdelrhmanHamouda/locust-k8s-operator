# Phase 15: E2E Tests (Kind) - Technical Design

**Version:** 1.0  
**Status:** Draft  
**Approach:** Extend existing E2E framework with comprehensive LocustTest lifecycle tests

---

## Design Philosophy

**Incremental enhancement.** Build on existing E2E infrastructure. Keep tests focused, fast, and deterministic. Use real Kubernetes resources with Kind cluster.

---

## 1. Current E2E Architecture

### 1.1 Existing Structure

```
test/
├── e2e/
│   ├── e2e_suite_test.go     # Ginkgo suite setup, CertManager install
│   ├── e2e_test.go           # Basic operator deployment tests
│   ├── kind-config.yaml      # Kind cluster configuration
│   ├── conversion/           # Conversion webhook tests (placeholder)
│   └── webhook/              # Webhook tests (placeholder)
└── utils/
    └── utils.go              # Test utilities (Run, LoadImageToKind, etc.)
```

### 1.2 Current Test Flow

```
BeforeSuite:
  1. Build operator image (make docker-build)
  2. Load image into Kind cluster
  3. Install CertManager (if not present)

Tests:
  1. Create namespace
  2. Apply security labels
  3. Install CRDs (make install)
  4. Deploy operator (make deploy)
  5. Verify pod running
  6. Verify metrics accessible

AfterAll:
  1. Undeploy operator
  2. Uninstall CRDs
  3. Delete namespace
  4. Uninstall CertManager (if we installed it)
```

### 1.3 Makefile Targets

| Target | Purpose |
|--------|---------|
| `setup-test-e2e` | Create Kind cluster if not exists |
| `test-e2e` | Run E2E tests with Ginkgo |
| `cleanup-test-e2e` | Delete Kind cluster |

---

## 2. Target E2E Architecture

### 2.1 Proposed Structure

```
test/
├── e2e/
│   ├── e2e_suite_test.go         # Suite setup (existing, minor updates)
│   ├── e2e_test.go               # Manager tests (existing)
│   ├── locusttest_e2e_test.go    # NEW: LocustTest CR lifecycle tests
│   ├── v1_compatibility_test.go  # NEW: v1 API backward compatibility
│   ├── otel_e2e_test.go          # NEW: OpenTelemetry integration
│   ├── validation_e2e_test.go    # NEW: Webhook validation tests
│   ├── kind-config.yaml          # Kind configuration
│   ├── conversion/               # Conversion tests
│   └── testdata/                 # NEW: Sample CRs
│       ├── v1/
│       │   └── locusttest-basic.yaml
│       ├── v2/
│       │   ├── locusttest-basic.yaml
│       │   ├── locusttest-with-env.yaml
│       │   ├── locusttest-with-otel.yaml
│       │   ├── locusttest-with-volumes.yaml
│       │   └── locusttest-invalid.yaml
│       └── configmaps/
│           └── test-config.yaml
└── utils/
    └── utils.go                  # Utilities (extend as needed)
```

### 2.2 Test Data Strategy

Use YAML fixtures in `testdata/` rather than inline Go structs for:
- Easier maintenance and readability
- Realistic CR examples
- Reuse in documentation

---

## 3. Test Categories Design

### 3.1 LocustTest Lifecycle Tests

**File:** `locusttest_e2e_test.go`

```go
var _ = Describe("LocustTest", Ordered, func() {
    Context("v2 API lifecycle", func() {
        It("should create master Service on CR creation", func() { ... })
        It("should create master Job on CR creation", func() { ... })
        It("should create worker Job on CR creation", func() { ... })
        It("should set owner references on created resources", func() { ... })
        It("should update status phase to Running", func() { ... })
        It("should clean up resources on CR deletion", func() { ... })
    })

    Context("with environment injection", func() {
        It("should inject ConfigMap env vars", func() { ... })
        It("should inject Secret env vars", func() { ... })
        It("should inject inline env variables", func() { ... })
    })

    Context("with custom volumes", func() {
        It("should mount volumes to master pod", func() { ... })
        It("should mount volumes to worker pods", func() { ... })
        It("should respect target selector", func() { ... })
    })
})
```

### 3.2 v1 Compatibility Tests

**File:** `v1_compatibility_test.go`

```go
var _ = Describe("v1 API Compatibility", Ordered, func() {
    It("should accept v1 LocustTest CR", func() { ... })
    It("should create resources from v1 CR", func() { ... })
    It("should emit deprecation warning event", func() { ... })
    It("should allow reading v1 CR as v2", func() { ... })
})
```

### 3.3 OpenTelemetry Tests

**File:** `otel_e2e_test.go`

```go
var _ = Describe("OpenTelemetry", Ordered, func() {
    It("should add --otel flag when enabled", func() { ... })
    It("should inject OTEL_* environment variables", func() { ... })
    It("should NOT deploy metrics sidecar when OTel enabled", func() { ... })
    It("should still deploy sidecar when OTel disabled", func() { ... })
})
```

### 3.4 Validation Tests

**File:** `validation_e2e_test.go`

```go
var _ = Describe("Validation Webhook", Ordered, func() {
    It("should reject CR with missing required fields", func() { ... })
    It("should reject CR with reserved path conflict", func() { ... })
    It("should reject CR with invalid workerReplicas", func() { ... })
    It("should accept valid CR", func() { ... })
})
```

---

## 4. Helper Functions

### 4.1 CR Application Helpers

```go
// ApplyLocustTestFromFile applies a LocustTest CR from a YAML file
func ApplyLocustTestFromFile(namespace, path string) error {
    cmd := exec.Command("kubectl", "apply", "-f", path, "-n", namespace)
    _, err := utils.Run(cmd)
    return err
}

// DeleteLocustTest deletes a LocustTest CR by name
func DeleteLocustTest(namespace, name string) error {
    cmd := exec.Command("kubectl", "delete", "locusttest", name, "-n", namespace)
    _, err := utils.Run(cmd)
    return err
}
```

### 4.2 Resource Verification Helpers

```go
// WaitForService waits for a Service to exist
func WaitForService(namespace, name string, timeout time.Duration) error {
    return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
        cmd := exec.Command("kubectl", "get", "service", name, "-n", namespace)
        _, err := utils.Run(cmd)
        return err == nil, nil
    })
}

// WaitForJob waits for a Job to exist
func WaitForJob(namespace, name string, timeout time.Duration) error { ... }

// GetJobPodSpec retrieves the pod spec from a Job
func GetJobPodSpec(namespace, jobName string) (*corev1.PodSpec, error) { ... }

// VerifyOwnerReference checks if a resource has the expected owner
func VerifyOwnerReference(namespace, resourceType, name, ownerName string) error { ... }
```

### 4.3 Status Verification Helpers

```go
// GetLocustTestPhase retrieves the current phase of a LocustTest
func GetLocustTestPhase(namespace, name string) (string, error) {
    cmd := exec.Command("kubectl", "get", "locusttest", name, 
        "-n", namespace, "-o", "jsonpath={.status.phase}")
    output, err := utils.Run(cmd)
    return output, err
}

// WaitForPhase waits for a LocustTest to reach a specific phase
func WaitForPhase(namespace, name, phase string, timeout time.Duration) error { ... }
```

---

## 5. Test Data Files

### 5.1 Basic v2 LocustTest

**File:** `testdata/v2/locusttest-basic.yaml`

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: e2e-test-basic
spec:
  master:
    command: "-f /lotest/src/locustfile.py"
  worker:
    replicas: 2
    command: "-f /lotest/src/locustfile.py"
  image:
    name: locustio/locust:latest
  testFiles:
    configMapRef: e2e-test-configmap
```

### 5.2 LocustTest with Environment

**File:** `testdata/v2/locusttest-with-env.yaml`

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: e2e-test-env
spec:
  master:
    command: "-f /lotest/src/locustfile.py"
  worker:
    replicas: 1
    command: "-f /lotest/src/locustfile.py"
  image:
    name: locustio/locust:latest
  testFiles:
    configMapRef: e2e-test-configmap
  env:
    configMapRefs:
      - name: e2e-env-configmap
    variables:
      - name: E2E_TEST_VAR
        value: "test-value"
```

### 5.3 LocustTest with OpenTelemetry

**File:** `testdata/v2/locusttest-with-otel.yaml`

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: e2e-test-otel
spec:
  master:
    command: "-f /lotest/src/locustfile.py"
  worker:
    replicas: 1
    command: "-f /lotest/src/locustfile.py"
  image:
    name: locustio/locust:latest
  testFiles:
    configMapRef: e2e-test-configmap
  observability:
    openTelemetry:
      enabled: true
      endpoint: "http://otel-collector:4317"
```

### 5.4 Basic v1 LocustTest

**File:** `testdata/v1/locusttest-basic.yaml`

```yaml
apiVersion: locust.io/v1
kind: LocustTest
metadata:
  name: e2e-test-v1
spec:
  masterCommandSeed: "-f /lotest/src/locustfile.py"
  workerCommandSeed: "-f /lotest/src/locustfile.py"
  workerReplicas: 1
  image: locustio/locust:latest
  configMap: e2e-test-configmap
```

### 5.5 Test ConfigMap

**File:** `testdata/configmaps/test-config.yaml`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: e2e-test-configmap
data:
  locustfile.py: |
    from locust import HttpUser, task
    
    class TestUser(HttpUser):
        @task
        def hello(self):
            self.client.get("/")
```

---

## 6. CI Integration

### 6.1 Enhanced Workflow

**File:** `.github/workflows/go-test-e2e.yml`

```yaml
name: 🚀 Go E2E Tests

on:
  push:
    branches: [main, master]
    paths:
      - 'locust-k8s-operator-go/**'
  pull_request:
    paths:
      - 'locust-k8s-operator-go/**'

defaults:
  run:
    working-directory: locust-k8s-operator-go

jobs:
  test-e2e:
    name: 🎯 Run E2E Tests
    runs-on: ubuntu-latest
    steps:
      - name: 📥 Clone the code
        uses: actions/checkout@v4

      - name: 🔧 Setup Go
        uses: actions/setup-go@v5
        with:
          go-version-file: locust-k8s-operator-go/go.mod

      - name: ☸️ Install Kind
        run: |
          curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
          chmod +x ./kind
          sudo mv ./kind /usr/local/bin/kind

      - name: ✔️ Verify Kind installation
        run: kind version

      - name: 🎯 Run E2E tests
        run: |
          go mod tidy
          make test-e2e

      - name: 📊 Upload E2E coverage (optional)
        if: success()
        uses: codecov/codecov-action@v5
        with:
          files: locust-k8s-operator-go/cover-e2e.out
          flags: e2e-tests
          name: e2e-coverage
          fail_ci_if_error: false
```

### 6.2 Timeout Configuration

Set appropriate timeouts in tests:

```go
SetDefaultEventuallyTimeout(3 * time.Minute)
SetDefaultEventuallyPollingInterval(2 * time.Second)
```

### 6.3 Parallel Execution

Tests within different `Context` blocks can run in parallel using Ginkgo:

```go
var _ = Describe("LocustTest", Ordered, func() {
    // Ordered contexts run sequentially
    // Individual It blocks within are atomic
})
```

---

## 7. Test Execution Flow

### 7.1 Complete Test Sequence

```
1. make setup-test-e2e
   └── Create Kind cluster (if not exists)

2. make test-e2e
   ├── BeforeSuite:
   │   ├── Build operator image
   │   ├── Load image to Kind
   │   └── Install CertManager
   │
   ├── Manager Tests (existing):
   │   ├── Create namespace
   │   ├── Install CRDs
   │   ├── Deploy operator
   │   ├── Verify pod running
   │   └── Verify metrics
   │
   ├── LocustTest Lifecycle Tests:
   │   ├── Apply test ConfigMap
   │   ├── Apply LocustTest CR
   │   ├── Verify Service created
   │   ├── Verify Jobs created
   │   ├── Verify owner references
   │   ├── Verify status phase
   │   ├── Delete CR
   │   └── Verify cleanup
   │
   ├── v1 Compatibility Tests:
   │   ├── Apply v1 CR
   │   ├── Verify conversion
   │   └── Verify resources
   │
   ├── OTel Tests:
   │   ├── Apply OTel-enabled CR
   │   ├── Verify OTel config
   │   └── Verify no sidecar
   │
   ├── Validation Tests:
   │   ├── Apply invalid CR
   │   └── Verify rejection
   │
   └── AfterSuite:
       ├── Undeploy operator
       ├── Uninstall CRDs
       ├── Delete namespace
       └── Uninstall CertManager

3. make cleanup-test-e2e
   └── Delete Kind cluster
```

---

## 8. Debugging Support

### 8.1 Log Collection on Failure

Existing `AfterEach` already collects:
- Controller manager logs
- Kubernetes events
- Pod descriptions

### 8.2 Additional Debug Info

Add to failure collection:
- LocustTest CR status
- Job pod logs
- Service endpoints

```go
AfterEach(func() {
    if CurrentSpecReport().Failed() {
        // Existing debug output...
        
        By("Fetching LocustTest CR status")
        cmd := exec.Command("kubectl", "get", "locusttest", "-n", namespace, "-o", "yaml")
        output, _ := utils.Run(cmd)
        fmt.Fprintf(GinkgoWriter, "LocustTest CRs:\n%s", output)
        
        By("Fetching Jobs")
        cmd = exec.Command("kubectl", "get", "jobs", "-n", namespace, "-o", "yaml")
        output, _ = utils.Run(cmd)
        fmt.Fprintf(GinkgoWriter, "Jobs:\n%s", output)
    }
})
```

---

## 9. Performance Considerations

### 9.1 Target Execution Time

| Category | Tests | Est. Time |
|----------|-------|-----------|
| Manager tests | 2 | 30s |
| LocustTest lifecycle | 6 | 90s |
| v1 compatibility | 4 | 60s |
| OTel tests | 4 | 60s |
| Validation tests | 4 | 30s |
| Setup/teardown | - | 120s |
| **Total** | **20** | **< 7 min** |

### 9.2 Optimization Strategies

1. **Reuse namespace** where possible within ordered contexts
2. **Skip image rebuild** if unchanged (cache in CI)
3. **Parallel test files** with `ginkgo -p` (if tests are independent)
4. **Minimal wait times** - use `Eventually` with appropriate polling

---

## 10. Implementation Order

1. **Day 1:**
   - Create `testdata/` fixtures
   - Implement `locusttest_e2e_test.go` (core lifecycle)
   - Add helper functions

2. **Day 2:**
   - Implement `v1_compatibility_test.go`
   - Implement `otel_e2e_test.go`
   - Implement `validation_e2e_test.go`
   - Enhance CI workflow
   - Verify all tests pass

---

## 11. References

- [Ginkgo Testing Framework](https://onsi.github.io/ginkgo/)
- [Operator SDK E2E Testing](https://sdk.operatorframework.io/docs/building-operators/golang/testing/)
- [Kind Documentation](https://kind.sigs.k8s.io/)
- [Kubernetes E2E Framework](https://github.com/kubernetes-sigs/e2e-framework)
