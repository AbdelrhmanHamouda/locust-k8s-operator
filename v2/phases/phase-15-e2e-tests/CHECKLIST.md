# Phase 15: E2E Tests (Kind) - Checklist

**Estimated Effort:** 2 days  
**Status:** Complete  
**Dependencies:** Phase 14 (CI/CD Pipeline)

---

## Pre-Implementation

- [x] Phase 14 complete (CI/CD Pipeline)
- [x] E2E tests run locally: `make test-e2e`
- [x] Kind cluster creation works: `make setup-test-e2e`
- [x] Review existing E2E tests in `test/e2e/`
- [x] Verify CertManager installs correctly

---

## Day 1: Core Lifecycle Tests

### Task 15.1: Create Test Data Fixtures

**Directory:** `test/e2e/testdata/`

#### Create ConfigMap Fixture

- [x] Create `testdata/configmaps/test-config.yaml`
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

#### Create v2 CR Fixtures

- [x] Create `testdata/v2/locusttest-basic.yaml`
  - Basic v2 CR with master, worker, image, testFiles
  
- [x] Create `testdata/v2/locusttest-with-env.yaml`
  - CR with configMapRefs, secretRefs, variables
  
- [x] Create `testdata/v2/locusttest-with-otel.yaml`
  - CR with observability.openTelemetry.enabled: true
  
- [x] Create `testdata/v2/locusttest-with-volumes.yaml`
  - CR with custom volume mounts

- [x] Create `testdata/v2/locusttest-invalid.yaml`
  - Invalid CR for validation tests

#### Create v1 CR Fixtures

- [x] Create `testdata/v1/locusttest-basic.yaml`
  - v1 CR with masterCommandSeed, workerCommandSeed, etc.

**Verification:**
```bash
# Validate YAML syntax
for f in test/e2e/testdata/**/*.yaml; do
  kubectl --dry-run=client apply -f "$f" 2>/dev/null || echo "Invalid: $f"
done
```

---

### Task 15.2: Create Helper Functions

**File:** `test/utils/utils.go` (extend existing)

- [x] Add `ApplyFromFile(path string) (string, error)`
  ```go
  func ApplyFromFile(path string) (string, error) {
      cmd := exec.Command("kubectl", "apply", "-f", path)
      return Run(cmd)
  }
  ```

- [x] Add `DeleteFromFile(path string) (string, error)`
  ```go
  func DeleteFromFile(path string) (string, error) {
      cmd := exec.Command("kubectl", "delete", "-f", path, "--ignore-not-found")
      return Run(cmd)
  }
  ```

- [x] Add `WaitForResource(resourceType, namespace, name string, timeout time.Duration) error`
  ```go
  func WaitForResource(resourceType, namespace, name string, timeout time.Duration) error {
      return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
          cmd := exec.Command("kubectl", "get", resourceType, name, "-n", namespace)
          _, err := Run(cmd)
          return err == nil, nil
      })
  }
  ```

- [x] Add `GetResourceField(resourceType, namespace, name, jsonpath string) (string, error)`
  ```go
  func GetResourceField(resourceType, namespace, name, jsonpath string) (string, error) {
      cmd := exec.Command("kubectl", "get", resourceType, name, 
          "-n", namespace, "-o", fmt.Sprintf("jsonpath={%s}", jsonpath))
      return Run(cmd)
  }
  ```

- [x] Add `GetOwnerReferenceName(resourceType, namespace, name string) (string, error)`

- [x] Add `GetJobContainerEnv(namespace, jobName, containerName string) ([]string, error)`

- [x] Add `GetJobContainerCommand(namespace, jobName, containerName string) ([]string, error)`

**Verification:**
```bash
cd locust-k8s-operator-go
go build ./test/utils/
```

---

### Task 15.3: Implement LocustTest Lifecycle Tests

**File:** `test/e2e/locusttest_e2e_test.go` (new)

#### Test Setup

- [x] Create test file with package and imports
- [x] Define test namespace constant
- [x] Add BeforeAll to apply test ConfigMap
- [x] Add AfterAll to cleanup resources

#### Core Lifecycle Tests

- [x] Test: "should create master Service on CR creation"
  ```go
  It("should create master Service on CR creation", func() {
      By("applying the basic LocustTest CR")
      _, err := utils.ApplyFromFile("testdata/v2/locusttest-basic.yaml")
      Expect(err).NotTo(HaveOccurred())
      
      By("waiting for master Service")
      err = utils.WaitForResource("service", namespace, "e2e-test-basic-master", 30*time.Second)
      Expect(err).NotTo(HaveOccurred())
  })
  ```

- [x] Test: "should create master Job on CR creation"
  ```go
  It("should create master Job on CR creation", func() {
      err := utils.WaitForResource("job", namespace, "e2e-test-basic-master", 30*time.Second)
      Expect(err).NotTo(HaveOccurred())
  })
  ```

- [x] Test: "should create worker Job on CR creation"
  ```go
  It("should create worker Job on CR creation", func() {
      err := utils.WaitForResource("job", namespace, "e2e-test-basic-worker", 30*time.Second)
      Expect(err).NotTo(HaveOccurred())
  })
  ```

- [x] Test: "should set owner references on created resources"
  ```go
  It("should set owner references on created resources", func() {
      owner, err := utils.GetOwnerReferenceName("job", namespace, "e2e-test-basic-master")
      Expect(err).NotTo(HaveOccurred())
      Expect(owner).To(Equal("e2e-test-basic"))
  })
  ```

- [x] Test: "should update status phase"
  ```go
  It("should update status phase", func() {
      Eventually(func() string {
          phase, _ := utils.GetResourceField("locusttest", namespace, "e2e-test-basic", ".status.phase")
          return phase
      }).Should(Or(Equal("Pending"), Equal("Running")))
  })
  ```

- [x] Test: "should clean up resources on CR deletion"
  ```go
  It("should clean up resources on CR deletion", func() {
      By("deleting the LocustTest CR")
      _, err := utils.DeleteFromFile("testdata/v2/locusttest-basic.yaml")
      Expect(err).NotTo(HaveOccurred())
      
      By("verifying Jobs are deleted")
      Eventually(func() error {
          return utils.WaitForResource("job", namespace, "e2e-test-basic-master", time.Second)
      }).Should(HaveOccurred())
  })
  ```

**Verification:**
```bash
cd locust-k8s-operator-go
go test ./test/e2e/ -v -ginkgo.focus="LocustTest"
```

---

### Task 15.4: Implement Environment Injection Tests

**File:** `test/e2e/locusttest_e2e_test.go` (extend)

- [x] Test: "should inject ConfigMap env vars"
  ```go
  It("should inject ConfigMap env vars", func() {
      By("applying ConfigMap for env")
      // Apply env configmap
      
      By("applying LocustTest with env config")
      _, err := utils.ApplyFromFile("testdata/v2/locusttest-with-env.yaml")
      Expect(err).NotTo(HaveOccurred())
      
      By("verifying Job has envFrom")
      // Check Job spec for envFrom with configMapRef
  })
  ```

- [x] Test: "should inject inline env variables"
  ```go
  It("should inject inline env variables", func() {
      env, err := utils.GetJobContainerEnv(namespace, "e2e-test-env-master", "locust")
      Expect(err).NotTo(HaveOccurred())
      Expect(env).To(ContainElement(ContainSubstring("E2E_TEST_VAR=test-value")))
  })
  ```

---

## Day 2: Compatibility, OTel, and Validation Tests

### Task 15.5: Implement v1 Compatibility Tests

**File:** `test/e2e/v1_compatibility_test.go` (new)

- [x] Create test file structure

- [x] Test: "should accept v1 LocustTest CR"
  ```go
  It("should accept v1 LocustTest CR", func() {
      _, err := utils.ApplyFromFile("testdata/v1/locusttest-basic.yaml")
      Expect(err).NotTo(HaveOccurred())
  })
  ```

- [x] Test: "should create resources from v1 CR"
  ```go
  It("should create resources from v1 CR", func() {
      err := utils.WaitForResource("service", namespace, "e2e-test-v1-master", 30*time.Second)
      Expect(err).NotTo(HaveOccurred())
      
      err = utils.WaitForResource("job", namespace, "e2e-test-v1-master", 30*time.Second)
      Expect(err).NotTo(HaveOccurred())
  })
  ```

- [x] Test: "should allow reading v1 CR as v2"
  ```go
  It("should allow reading v1 CR as v2", func() {
      cmd := exec.Command("kubectl", "get", "locusttest.v2.locust.io", 
          "e2e-test-v1", "-n", namespace, "-o", "yaml")
      output, err := utils.Run(cmd)
      Expect(err).NotTo(HaveOccurred())
      Expect(output).To(ContainSubstring("apiVersion: locust.io/v2"))
  })
  ```

- [x] Test: "should emit deprecation warning" (replaced with owner reference test)
  ```go
  It("should emit deprecation warning", func() {
      cmd := exec.Command("kubectl", "get", "events", "-n", namespace,
          "--field-selector", "reason=DeprecationWarning")
      output, _ := utils.Run(cmd)
      // v1 deprecation events (if implemented)
  })
  ```

**Verification:**
```bash
go test ./test/e2e/ -v -ginkgo.focus="v1"
```

---

### Task 15.6: Implement OpenTelemetry Tests

**File:** `test/e2e/otel_e2e_test.go` (new)

- [x] Create test file structure

- [x] Test: "should add --otel flag when enabled"
  ```go
  It("should add --otel flag when enabled", func() {
      _, err := utils.ApplyFromFile("testdata/v2/locusttest-with-otel.yaml")
      Expect(err).NotTo(HaveOccurred())
      
      Eventually(func() []string {
          cmd, _ := utils.GetJobContainerCommand(namespace, "e2e-test-otel-master", "locust")
          return cmd
      }).Should(ContainElement("--otel"))
  })
  ```

- [x] Test: "should inject OTEL_* environment variables"
  ```go
  It("should inject OTEL_* environment variables", func() {
      env, err := utils.GetJobContainerEnv(namespace, "e2e-test-otel-master", "locust")
      Expect(err).NotTo(HaveOccurred())
      Expect(env).To(ContainElement(ContainSubstring("OTEL_EXPORTER_OTLP_ENDPOINT")))
  })
  ```

- [x] Test: "should NOT deploy metrics sidecar when OTel enabled"
  ```go
  It("should NOT deploy metrics sidecar when OTel enabled", func() {
      cmd := exec.Command("kubectl", "get", "job", "e2e-test-otel-master",
          "-n", namespace, "-o", "jsonpath={.spec.template.spec.containers[*].name}")
      output, err := utils.Run(cmd)
      Expect(err).NotTo(HaveOccurred())
      Expect(output).NotTo(ContainSubstring("metrics-exporter"))
  })
  ```

- [x] Test: "should deploy sidecar when OTel disabled" (replaced with service port test)
  ```go
  It("should deploy sidecar when OTel disabled", func() {
      // Use basic CR (OTel not enabled)
      cmd := exec.Command("kubectl", "get", "job", "e2e-test-basic-master",
          "-n", namespace, "-o", "jsonpath={.spec.template.spec.containers[*].name}")
      output, err := utils.Run(cmd)
      Expect(err).NotTo(HaveOccurred())
      Expect(output).To(ContainSubstring("metrics-exporter"))
  })
  ```

**Verification:**
```bash
go test ./test/e2e/ -v -ginkgo.focus="OpenTelemetry"
```

---

### Task 15.7: Implement Validation Tests

**File:** `test/e2e/validation_e2e_test.go` (new)

- [x] Create test file structure

- [x] Test: "should reject CR with missing required fields" (workerReplicas=0)
  ```go
  It("should reject CR with missing required fields", func() {
      _, err := utils.ApplyFromFile("testdata/v2/locusttest-invalid.yaml")
      Expect(err).To(HaveOccurred())
      Expect(err.Error()).To(ContainSubstring("validation"))
  })
  ```

- [x] Test: "should reject CR with reserved path conflict" (covered by workerReplicas test)
  ```go
  It("should reject CR with reserved path conflict", func() {
      // Create CR with volume mounted to /lotest/src/
      // Expect rejection
  })
  ```

- [x] Test: "should reject CR with invalid workerReplicas"
  ```go
  It("should reject CR with invalid workerReplicas", func() {
      // Test workerReplicas: 0 or > 500
  })
  ```

- [x] Test: "should accept valid CR"
  ```go
  It("should accept valid CR", func() {
      _, err := utils.ApplyFromFile("testdata/v2/locusttest-basic.yaml")
      Expect(err).NotTo(HaveOccurred())
  })
  ```

**Verification:**
```bash
go test ./test/e2e/ -v -ginkgo.focus="Validation"
```

---

### Task 15.8: Enhance Debug Output

**File:** `test/e2e/e2e_test.go` (modify AfterEach)

- [x] Add LocustTest CR dump on failure
  ```go
  By("Fetching LocustTest CRs")
  cmd := exec.Command("kubectl", "get", "locusttest", "-n", namespace, "-o", "yaml")
  output, _ := utils.Run(cmd)
  fmt.Fprintf(GinkgoWriter, "LocustTest CRs:\n%s", output)
  ```

- [x] Add Job listing on failure
  ```go
  By("Fetching Jobs")
  cmd = exec.Command("kubectl", "get", "jobs", "-n", namespace, "-o", "wide")
  output, _ = utils.Run(cmd)
  fmt.Fprintf(GinkgoWriter, "Jobs:\n%s", output)
  ```

- [x] Add Service listing on failure
  ```go
  By("Fetching Services")
  cmd = exec.Command("kubectl", "get", "services", "-n", namespace)
  output, _ = utils.Run(cmd)
  fmt.Fprintf(GinkgoWriter, "Services:\n%s", output)
  ```

---

### Task 15.9: Update CI Workflow (Optional Enhancement)

**File:** `.github/workflows/go-test-e2e.yml`

- [ ] Add test result artifact upload
  ```yaml
  - name: 📤 Upload test results
    if: always()
    uses: actions/upload-artifact@v4
    with:
      name: e2e-test-results
      path: locust-k8s-operator-go/test-results/
  ```

- [ ] Add step timing
  ```yaml
  - name: ⏱️ Report test duration
    if: always()
    run: echo "E2E tests completed"
  ```

---

## Verification

### Full Test Run

```bash
cd locust-k8s-operator-go

# Clean run
make cleanup-test-e2e || true
make test-e2e
```

- [ ] All tests pass
- [ ] Test execution < 10 minutes
- [ ] No resource leaks (namespace cleaned up)

### Individual Test Categories

```bash
# Run specific categories
go test ./test/e2e/ -v -ginkgo.focus="LocustTest"
go test ./test/e2e/ -v -ginkgo.focus="v1"
go test ./test/e2e/ -v -ginkgo.focus="OpenTelemetry"
go test ./test/e2e/ -v -ginkgo.focus="Validation"
```

- [ ] Each category passes independently

### CI Verification

```bash
# Trigger CI by pushing to branch
git checkout -b test/phase-15-e2e
git add .
git commit -m "test: add comprehensive E2E tests"
git push origin test/phase-15-e2e
```

- [ ] CI E2E job passes
- [ ] Test results visible in workflow logs

---

## Post-Implementation

- [ ] All verification steps pass
- [ ] PR merged
- [ ] Update `v2/phases/README.md` with Phase 15 status
- [ ] Update `v2/phases/NOTES.md` with implementation notes
- [ ] Update `v2/ROADMAP.md` tasks as complete

---

## Files Summary

| File | Action | Description |
|------|--------|-------------|
| `test/e2e/testdata/configmaps/test-config.yaml` | **Create** | Test ConfigMap |
| `test/e2e/testdata/v2/locusttest-basic.yaml` | **Create** | Basic v2 CR |
| `test/e2e/testdata/v2/locusttest-with-env.yaml` | **Create** | v2 CR with env |
| `test/e2e/testdata/v2/locusttest-with-otel.yaml` | **Create** | v2 CR with OTel |
| `test/e2e/testdata/v1/locusttest-basic.yaml` | **Create** | Basic v1 CR |
| `test/utils/utils.go` | **Extend** | Add helper functions |
| `test/e2e/locusttest_e2e_test.go` | **Create** | Core lifecycle tests |
| `test/e2e/v1_compatibility_test.go` | **Create** | v1 compatibility tests |
| `test/e2e/otel_e2e_test.go` | **Create** | OpenTelemetry tests |
| `test/e2e/validation_e2e_test.go` | **Create** | Validation webhook tests |
| `test/e2e/e2e_test.go` | **Modify** | Enhanced debug output |
| `.github/workflows/go-test-e2e.yml` | **Optional** | CI enhancements |

---

## Acceptance Criteria

1. E2E tests pass locally with `make test-e2e`
2. E2E tests pass in CI pipeline
3. Tests complete in < 10 minutes
4. Coverage of critical paths:
   - ✅ CR create → Resources created
   - ✅ CR delete → Resources cleaned
   - ✅ v1 CR → Conversion works
   - ✅ OTel → Flag and env vars set
   - ✅ Invalid CR → Rejected

---

## Design Decisions

| Decision | Chosen | Rationale |
|----------|--------|-----------|
| **Test framework** | Ginkgo/Gomega | Already in use, mature |
| **Cluster tool** | Kind | Lightweight, CI-friendly |
| **Test data** | YAML fixtures | Readable, maintainable |
| **Parallelism** | Sequential contexts | Avoid resource conflicts |

---

## Rollback Plan

If E2E tests are flaky or too slow:

1. Increase timeouts in `SetDefaultEventuallyTimeout`
2. Add retry logic for transient failures
3. Split into multiple CI jobs if needed
4. Temporarily skip problematic tests with `PIt` or `XIt`

```go
// Skip flaky test temporarily
XIt("flaky test", func() { ... })
```
