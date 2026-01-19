# Phase 5: Unit Tests - Implementation Plan

**Effort:** 2 days  
**Priority:** P0 - Critical Path  
**Prerequisites:** Phase 4 (Core Reconciler)  
**Requirements:** §7.1 Testing Requirements (80% coverage)

---

## Objective

Port and enhance unit tests from Java to Go, achieving **80% code coverage** for resource builders and **70% coverage** for the controller. This phase validates behavioral parity between the Go and Java implementations through comprehensive testing.

---

## Day 1: Resource Builder Tests

### Task 5.1: Enhance Job Builder Tests

**File:** `internal/resources/job_test.go`

The existing tests cover basic functionality. Add these additional test cases:

#### Kafka Environment Variable Tests

```go
func TestBuildMasterJob_KafkaEnvVars(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()
	cfg.KafkaSecurityEnabled = true
	cfg.KafkaBootstrapServers = "kafka.example.com:9092"
	cfg.KafkaSecurityProtocol = "SASL_SSL"
	cfg.KafkaUsername = "user"
	cfg.KafkaPassword = "secret"
	cfg.KafkaSaslMechanism = "PLAIN"

	job := BuildMasterJob(lt, cfg)

	container := findContainer(job.Spec.Template.Spec.Containers, "my-test-master")
	require.NotNil(t, container)

	// Verify Kafka env vars are set
	envMap := envVarsToMap(container.Env)
	assert.Equal(t, "kafka.example.com:9092", envMap["KAFKA_BOOTSTRAP_SERVERS"])
	assert.Equal(t, "SASL_SSL", envMap["KAFKA_SECURITY_PROTOCOL"])
}

func TestBuildMasterJob_KafkaDisabled(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()
	cfg.KafkaSecurityEnabled = false

	job := BuildMasterJob(lt, cfg)

	container := findContainer(job.Spec.Template.Spec.Containers, "my-test-master")
	require.NotNil(t, container)

	// Kafka security env vars should not be set when disabled
	envMap := envVarsToMap(container.Env)
	assert.Empty(t, envMap["KAFKA_USERNAME"])
	assert.Empty(t, envMap["KAFKA_PASSWORD"])
}
```

#### Job Spec Verification Tests

```go
func TestBuildMasterJob_JobSpecDefaults(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	// Verify Job-level settings
	require.NotNil(t, job.Spec.Parallelism)
	assert.Equal(t, int32(1), *job.Spec.Parallelism)
	
	require.NotNil(t, job.Spec.Completions)
	assert.Equal(t, int32(1), *job.Spec.Completions)
	
	require.NotNil(t, job.Spec.BackoffLimit)
	assert.Equal(t, int32(0), *job.Spec.BackoffLimit)
	
	// Verify RestartPolicy
	assert.Equal(t, corev1.RestartPolicyNever, job.Spec.Template.Spec.RestartPolicy)
}

func TestBuildWorkerJob_Parallelism(t *testing.T) {
	tests := []struct {
		name           string
		workerReplicas int32
		expected       int32
	}{
		{"single worker", 1, 1},
		{"multiple workers", 5, 5},
		{"max workers", 500, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := newTestLocustTest()
			lt.Spec.WorkerReplicas = tt.workerReplicas
			cfg := newTestConfig()

			job := BuildWorkerJob(lt, cfg)

			require.NotNil(t, job.Spec.Parallelism)
			assert.Equal(t, tt.expected, *job.Spec.Parallelism)
		})
	}
}
```

#### Edge Case Tests

```go
func TestBuildMasterJob_LongCRName(t *testing.T) {
	lt := newTestLocustTest()
	lt.Name = "very-long-locust-test-name-that-might-exceed-limits"
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	// Should not panic, name should be constructed
	assert.Contains(t, job.Name, "master")
}

func TestBuildMasterJob_NilConfigMap(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.ConfigMap = ""
	cfg := newTestConfig()

	// Should handle gracefully - may panic or produce invalid Job
	// This test documents the behavior
	job := BuildMasterJob(lt, cfg)
	assert.NotNil(t, job)
}

func TestBuildMasterJob_SpecialCharactersInName(t *testing.T) {
	lt := newTestLocustTest()
	lt.Name = "test.with.dots"
	cfg := newTestConfig()

	job := BuildMasterJob(lt, cfg)

	// Dots should be replaced with dashes in pod name
	assert.Equal(t, "test-with-dots-master", job.Name)
}
```

#### Helper Functions for Tests

```go
// findContainer finds a container by name in a slice
func findContainer(containers []corev1.Container, name string) *corev1.Container {
	for i := range containers {
		if containers[i].Name == name {
			return &containers[i]
		}
	}
	return nil
}

// envVarsToMap converts a slice of EnvVar to a map for easier assertions
func envVarsToMap(envVars []corev1.EnvVar) map[string]string {
	result := make(map[string]string)
	for _, env := range envVars {
		result[env.Name] = env.Value
	}
	return result
}
```

---

### Task 5.2: Enhance Service Builder Tests

**File:** `internal/resources/service_test.go`

```go
func TestBuildMasterService_AllPorts(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()
	cfg.MetricsExporterPort = 9646

	svc := BuildMasterService(lt, cfg)

	// Expected ports: 5557, 5558, 8089, 8080, 9646
	expectedPorts := map[string]int32{
		"master":       5557,
		"master-p2p":   5558,
		"web":          8089,
		"web-internal": 8080,
		"metrics":      9646,
	}

	assert.Len(t, svc.Spec.Ports, len(expectedPorts))

	for _, port := range svc.Spec.Ports {
		expected, exists := expectedPorts[port.Name]
		assert.True(t, exists, "Unexpected port: %s", port.Name)
		assert.Equal(t, expected, port.Port, "Port %s has wrong value", port.Name)
	}
}

func TestBuildMasterService_Selector(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	svc := BuildMasterService(lt, cfg)

	// Selector should match master pod labels
	assert.Equal(t, "my-test-master", svc.Spec.Selector[LabelPodName])
}

func TestBuildMasterService_Type(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()

	svc := BuildMasterService(lt, cfg)

	assert.Equal(t, corev1.ServiceTypeClusterIP, svc.Spec.Type)
}

func TestBuildMasterService_CustomMetricsPort(t *testing.T) {
	lt := newTestLocustTest()
	cfg := newTestConfig()
	cfg.MetricsExporterPort = 8080 // Custom port

	svc := BuildMasterService(lt, cfg)

	var metricsPort *corev1.ServicePort
	for i := range svc.Spec.Ports {
		if svc.Spec.Ports[i].Name == "metrics" {
			metricsPort = &svc.Spec.Ports[i]
			break
		}
	}

	require.NotNil(t, metricsPort)
	assert.Equal(t, int32(8080), metricsPort.Port)
}
```

---

### Task 5.3: Enhance Command Builder Tests

**File:** `internal/resources/command_test.go`

```go
func TestBuildMasterCommand(t *testing.T) {
	tests := []struct {
		name           string
		commandSeed    string
		workerReplicas int32
		wantContains   []string
	}{
		{
			name:           "basic command",
			commandSeed:    "locust -f /lotest/src/test.py",
			workerReplicas: 3,
			wantContains: []string{
				"locust", "-f", "/lotest/src/test.py",
				"--master",
				"--master-port=5557",
				"--expect-workers=3",
				"--autostart",
				"--autoquit", "60",
				"--enable-rebalancing",
				"--only-summary",
			},
		},
		{
			name:           "single worker",
			commandSeed:    "locust -f test.py",
			workerReplicas: 1,
			wantContains: []string{
				"--expect-workers=1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := newTestLocustTest()
			lt.Spec.MasterCommandSeed = tt.commandSeed
			lt.Spec.WorkerReplicas = tt.workerReplicas

			cmd := BuildMasterCommand(lt)

			cmdStr := strings.Join(cmd, " ")
			for _, want := range tt.wantContains {
				assert.Contains(t, cmdStr, want)
			}
		})
	}
}

func TestBuildWorkerCommand(t *testing.T) {
	tests := []struct {
		name         string
		crName       string
		commandSeed  string
		wantContains []string
	}{
		{
			name:        "basic command",
			crName:      "my-test",
			commandSeed: "locust -f /lotest/src/test.py",
			wantContains: []string{
				"locust", "-f", "/lotest/src/test.py",
				"--worker",
				"--master-port=5557",
				"--master-host=my-test-master",
			},
		},
		{
			name:        "name with dots",
			crName:      "team.a.test",
			commandSeed: "locust -f test.py",
			wantContains: []string{
				"--master-host=team-a-test-master",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := newTestLocustTest()
			lt.Name = tt.crName
			lt.Spec.WorkerCommandSeed = tt.commandSeed

			cmd := BuildWorkerCommand(lt)

			cmdStr := strings.Join(cmd, " ")
			for _, want := range tt.wantContains {
				assert.Contains(t, cmdStr, want)
			}
		})
	}
}

func TestBuildMasterCommand_ArgumentOrder(t *testing.T) {
	lt := newTestLocustTest()
	lt.Spec.MasterCommandSeed = "locust -f /lotest/src/test.py --host https://example.com"
	lt.Spec.WorkerReplicas = 5

	cmd := BuildMasterCommand(lt)

	// Seed arguments should come first
	seedIndex := -1
	masterFlagIndex := -1
	for i, arg := range cmd {
		if arg == "-f" {
			seedIndex = i
		}
		if arg == "--master" {
			masterFlagIndex = i
		}
	}

	assert.Less(t, seedIndex, masterFlagIndex, "Seed args should come before --master flag")
}
```

---

### Task 5.4: Create Test Fixtures

**Directory:** `internal/testdata/`

Create reusable test fixtures for consistent test data.

**File:** `internal/testdata/locusttest_minimal.json`

```json
{
  "apiVersion": "locust.io/v1",
  "kind": "LocustTest",
  "metadata": {
    "name": "minimal-test",
    "namespace": "default"
  },
  "spec": {
    "masterCommandSeed": "locust -f /lotest/src/test.py",
    "workerCommandSeed": "locust -f /lotest/src/test.py",
    "workerReplicas": 1,
    "image": "locustio/locust:latest"
  }
}
```

**File:** `internal/testdata/locusttest_full.json`

```json
{
  "apiVersion": "locust.io/v1",
  "kind": "LocustTest",
  "metadata": {
    "name": "full-featured-test",
    "namespace": "load-testing"
  },
  "spec": {
    "masterCommandSeed": "locust -f /lotest/src/test.py --host https://api.example.com",
    "workerCommandSeed": "locust -f /lotest/src/test.py",
    "workerReplicas": 10,
    "image": "locustio/locust:2.20.0",
    "imagePullPolicy": "IfNotPresent",
    "imagePullSecrets": ["registry-secret"],
    "configMap": "locust-scripts",
    "libConfigMap": "locust-lib",
    "labels": {
      "master": {
        "team": "platform",
        "environment": "staging"
      },
      "worker": {
        "team": "platform"
      }
    },
    "annotations": {
      "master": {
        "description": "Load test master"
      },
      "worker": {
        "description": "Load test worker"
      }
    },
    "affinity": {
      "nodeAffinity": {
        "requiredDuringSchedulingIgnoredDuringExecution": {
          "node-type": "performance"
        }
      }
    },
    "tolerations": [
      {
        "key": "dedicated",
        "operator": "Equal",
        "value": "performance",
        "effect": "NoSchedule"
      }
    ]
  }
}
```

**File:** `internal/testdata/fixtures.go`

```go
package testdata

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
)

// LoadLocustTest loads a LocustTest from a JSON fixture file.
func LoadLocustTest(filename string) (*locustv1.LocustTest, error) {
	_, currentFile, _, _ := runtime.Caller(0)
	testdataDir := filepath.Dir(currentFile)
	
	data, err := os.ReadFile(filepath.Join(testdataDir, filename))
	if err != nil {
		return nil, err
	}

	var lt locustv1.LocustTest
	if err := json.Unmarshal(data, &lt); err != nil {
		return nil, err
	}

	return &lt, nil
}
```

---

## Day 2: Controller & Config Tests

### Task 5.5: Rewrite Controller Tests

**File:** `internal/controller/locusttest_controller_test.go`

Replace the minimal Ginkgo scaffold with comprehensive unit tests:

```go
package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
	"github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
)

func newTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = locustv1.AddToScheme(scheme)
	_ = batchv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	return scheme
}

func newTestReconciler(objs ...client.Object) (*LocustTestReconciler, *record.FakeRecorder) {
	scheme := newTestScheme()
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objs...).
		Build()
	recorder := record.NewFakeRecorder(10)

	return &LocustTestReconciler{
		Client:   fakeClient,
		Scheme:   scheme,
		Config:   config.LoadConfig(),
		Recorder: recorder,
	}, recorder
}

func newTestLocustTest(name, namespace string) *locustv1.LocustTest {
	return &locustv1.LocustTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			Namespace:  namespace,
			Generation: 1,
		},
		Spec: locustv1.LocustTestSpec{
			MasterCommandSeed: "locust -f /lotest/src/test.py",
			WorkerCommandSeed: "locust -f /lotest/src/test.py",
			WorkerReplicas:    3,
			Image:             "locustio/locust:latest",
			ConfigMap:         "test-configmap",
		},
	}
}

func TestReconcile_NotFound(t *testing.T) {
	reconciler, _ := newTestReconciler()

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "nonexistent",
			Namespace: "default",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestReconcile_CreateResources(t *testing.T) {
	lt := newTestLocustTest("my-test", "default")
	reconciler, recorder := newTestReconciler(lt)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Verify Service created
	svc := &corev1.Service{}
	err = reconciler.Client.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, svc)
	assert.NoError(t, err)
	assert.Equal(t, "my-test-master", svc.Name)

	// Verify master Job created
	masterJob := &batchv1.Job{}
	err = reconciler.Client.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, masterJob)
	assert.NoError(t, err)

	// Verify worker Job created
	workerJob := &batchv1.Job{}
	err = reconciler.Client.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-worker",
		Namespace: "default",
	}, workerJob)
	assert.NoError(t, err)

	// Verify events recorded
	select {
	case event := <-recorder.Events:
		assert.Contains(t, event, "Created")
	default:
		t.Error("Expected event to be recorded")
	}
}

func TestReconcile_NoOpOnUpdate(t *testing.T) {
	lt := newTestLocustTest("my-test", "default")
	lt.Generation = 2 // Simulates an update
	reconciler, _ := newTestReconciler(lt)

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	// Verify no resources created
	svc := &corev1.Service{}
	err = reconciler.Client.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, svc)
	assert.True(t, apierrors.IsNotFound(err), "Service should not be created on update")
}

func TestReconcile_IdempotentCreate(t *testing.T) {
	lt := newTestLocustTest("my-test", "default")
	reconciler, _ := newTestReconciler(lt)

	// First reconcile - creates resources
	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	assert.NoError(t, err)

	// Second reconcile - should be idempotent (resources exist)
	// Reset generation to 1 to simulate a re-reconcile
	lt.Generation = 1
	_ = reconciler.Client.Update(context.Background(), lt)

	_, err = reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	assert.NoError(t, err) // Should not error even if resources exist
}

func TestReconcile_OwnerReferences(t *testing.T) {
	lt := newTestLocustTest("my-test", "default")
	lt.UID = "test-uid-12345"
	reconciler, _ := newTestReconciler(lt)

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "my-test",
			Namespace: "default",
		},
	})
	require.NoError(t, err)

	// Verify owner reference on Service
	svc := &corev1.Service{}
	err = reconciler.Client.Get(context.Background(), types.NamespacedName{
		Name:      "my-test-master",
		Namespace: "default",
	}, svc)
	require.NoError(t, err)
	
	require.Len(t, svc.OwnerReferences, 1)
	assert.Equal(t, "my-test", svc.OwnerReferences[0].Name)
	assert.Equal(t, "LocustTest", svc.OwnerReferences[0].Kind)
}
```

---

### Task 5.6: Verify Config Tests

**File:** `internal/config/config_test.go`

The existing config tests are comprehensive. Add these edge cases:

```go
func TestLoadConfig_PartialOverrides(t *testing.T) {
	// Only set some env vars, verify others use defaults
	t.Setenv("POD_CPU_REQUEST", "500m")
	// POD_MEM_REQUEST not set - should use default

	cfg := LoadConfig()

	assert.Equal(t, "500m", cfg.PodCPURequest)
	assert.Equal(t, "128Mi", cfg.PodMemRequest) // Default
}

func TestLoadConfig_VeryLargeTTL(t *testing.T) {
	t.Setenv("JOB_TTL_SECONDS_AFTER_FINISHED", "2147483647") // Max int32

	cfg := LoadConfig()

	require.NotNil(t, cfg.TTLSecondsAfterFinished)
	assert.Equal(t, int32(2147483647), *cfg.TTLSecondsAfterFinished)
}

func TestLoadConfig_OverflowTTL(t *testing.T) {
	t.Setenv("JOB_TTL_SECONDS_AFTER_FINISHED", "9999999999999") // Overflow

	cfg := LoadConfig()

	// Should be nil or handle gracefully
	assert.Nil(t, cfg.TTLSecondsAfterFinished)
}
```

---

## Verification

### Coverage Commands

```bash
# Run all tests with coverage
go test -coverprofile=coverage.out ./internal/...

# View coverage summary
go tool cover -func=coverage.out

# Expected output format:
# github.com/.../internal/resources/job.go:XX      BuildMasterJob     100.0%
# github.com/.../internal/resources/service.go:XX  BuildMasterService 100.0%
# github.com/.../internal/controller/...           Reconcile          85.0%
# total:                                           (statements)       82.5%

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Targets

| Package | Target | Measured |
|---------|--------|----------|
| `internal/resources/` | ≥ 80% | TBD |
| `internal/controller/` | ≥ 70% | TBD |
| `internal/config/` | ≥ 80% | TBD |

---

## Java Test Equivalents

| Java Test | Go Equivalent | Status |
|-----------|---------------|--------|
| `ResourceCreationHelpersTest` | `job_test.go`, `service_test.go` | Existing |
| `LoadGenHelpersTest` | `command_test.go`, `labels_test.go` | Existing |
| `SysConfigTest` | `config_test.go` | Existing |
| `LocustTestReconcilerTest` | `locusttest_controller_test.go` | Needs expansion |

---

## References

- [Go Testing Documentation](https://go.dev/doc/tutorial/add-a-test)
- [Testify Assertions](https://github.com/stretchr/testify)
- [controller-runtime Testing](https://book.kubebuilder.io/reference/envtest.html)
- [Table-Driven Tests](https://go.dev/blog/subtests)
