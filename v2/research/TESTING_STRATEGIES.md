# Testing Strategies for Go Kubernetes Operators

**Research Date:** January 2026  
**Focus:** Testing approaches for the Locust K8s Operator Go migration

---

## Table of Contents

1. [Testing Pyramid](#1-testing-pyramid)
2. [Unit Testing](#2-unit-testing)
3. [Integration Testing with envtest](#3-integration-testing-with-envtest)
4. [End-to-End Testing](#4-end-to-end-testing)
5. [Test Fixtures & Utilities](#5-test-fixtures--utilities)
6. [Mocking Strategies](#6-mocking-strategies)
7. [CI/CD Integration](#7-cicd-integration)

---

## 1. Testing Pyramid

### 1.1 Recommended Distribution

```
                    ┌─────────┐
                    │  E2E    │  ~10%
                    │  Tests  │  (Kind cluster)
                   ─┴─────────┴─
                  ┌─────────────┐
                  │ Integration │  ~30%
                  │   (envtest) │
                 ─┴─────────────┴─
                ┌─────────────────┐
                │   Unit Tests    │  ~60%
                │   (Pure Go)     │
               ─┴─────────────────┴─
```

### 1.2 Test Categories for Locust Operator

| Category | Scope | Tools | Coverage Target |
|----------|-------|-------|-----------------|
| **Unit** | Business logic, helpers | `testing`, `testify` | 80%+ |
| **Integration** | Reconciler, API types | `envtest` | All paths |
| **E2E** | Full operator lifecycle | Kind, Ginkgo | Critical journeys |

---

## 2. Unit Testing

### 2.1 Testing Resource Builders

```go
// internal/resources/job_test.go
package resources

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    
    locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildMasterJob(t *testing.T) {
    tests := []struct {
        name       string
        locustTest *locustv1.LocustTest
        config     *Config
        wantName   string
        wantCmd    []string
    }{
        {
            name: "basic master job",
            locustTest: &locustv1.LocustTest{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "demo-test",
                    Namespace: "default",
                },
                Spec: locustv1.LocustTestSpec{
                    MasterCommandSeed: "--locustfile /lotest/src/test.py --host https://example.com",
                    WorkerReplicas:    5,
                    Image:             "locustio/locust:2.43.1",
                },
            },
            config:   DefaultConfig(),
            wantName: "demo-test-master",
            wantCmd: []string{
                "--locustfile", "/lotest/src/test.py",
                "--host", "https://example.com",
                "--master", "--master-port=5557",
                "--expect-workers=5",
                "--autostart", "--autoquit", "60",
                "--enable-rebalancing", "--only-summary",
            },
        },
        {
            name: "with custom image pull policy",
            locustTest: &locustv1.LocustTest{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "custom-test",
                    Namespace: "test-ns",
                },
                Spec: locustv1.LocustTestSpec{
                    MasterCommandSeed: "--locustfile /lotest/src/test.py",
                    WorkerReplicas:    3,
                    Image:             "my-registry/locust:v1",
                    ImagePullPolicy:   "Always",
                },
            },
            config:   DefaultConfig(),
            wantName: "custom-test-master",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            job := BuildMasterJob(tt.locustTest, tt.config)
            
            assert.Equal(t, tt.wantName, job.Name)
            assert.Equal(t, tt.locustTest.Namespace, job.Namespace)
            
            require.Len(t, job.Spec.Template.Spec.Containers, 2) // locust + metrics exporter
            
            locustContainer := job.Spec.Template.Spec.Containers[0]
            assert.Equal(t, tt.locustTest.Spec.Image, locustContainer.Image)
            
            if tt.wantCmd != nil {
                assert.Equal(t, tt.wantCmd, locustContainer.Args)
            }
        })
    }
}
```

### 2.2 Testing Helper Functions

```go
// internal/resources/helpers_test.go
package resources

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    
    locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
)

func TestConstructNodeName(t *testing.T) {
    tests := []struct {
        name     string
        crName   string
        mode     OperationalMode
        expected string
    }{
        {
            name:     "master with dot notation",
            crName:   "team-a.load-test",
            mode:     Master,
            expected: "team-a-load-test-master",
        },
        {
            name:     "worker with dot notation",
            crName:   "team-a.load-test",
            mode:     Worker,
            expected: "team-a-load-test-worker",
        },
        {
            name:     "simple name master",
            crName:   "simple-test",
            mode:     Master,
            expected: "simple-test-master",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ConstructNodeName(tt.crName, tt.mode)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestConstructMasterCommand(t *testing.T) {
    tests := []struct {
        name           string
        commandSeed    string
        workerReplicas int32
        expectedParts  []string
    }{
        {
            name:           "basic command",
            commandSeed:    "--locustfile /test.py",
            workerReplicas: 10,
            expectedParts:  []string{"--master", "--expect-workers=10", "--autostart"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ConstructMasterCommand(tt.commandSeed, tt.workerReplicas)
            
            for _, part := range tt.expectedParts {
                assert.Contains(t, result, part)
            }
        })
    }
}

func TestMergeLabels(t *testing.T) {
    tests := []struct {
        name     string
        base     map[string]string
        override map[string]string
        expected map[string]string
    }{
        {
            name:     "merge without conflict",
            base:     map[string]string{"app": "locust", "managed-by": "operator"},
            override: map[string]string{"team": "platform"},
            expected: map[string]string{"app": "locust", "managed-by": "operator", "team": "platform"},
        },
        {
            name:     "override takes precedence",
            base:     map[string]string{"app": "locust"},
            override: map[string]string{"app": "custom"},
            expected: map[string]string{"app": "custom"},
        },
        {
            name:     "nil override",
            base:     map[string]string{"app": "locust"},
            override: nil,
            expected: map[string]string{"app": "locust"},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MergeLabels(tt.base, tt.override)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2.3 Table-Driven Tests Pattern

```go
func TestValidateLocustTestSpec(t *testing.T) {
    tests := map[string]struct {
        spec      locustv1.LocustTestSpec
        wantErr   bool
        errMsg    string
    }{
        "valid spec": {
            spec: locustv1.LocustTestSpec{
                MasterCommandSeed: "--locustfile /test.py",
                WorkerCommandSeed: "--locustfile /test.py",
                WorkerReplicas:    5,
                Image:             "locustio/locust:latest",
            },
            wantErr: false,
        },
        "missing image": {
            spec: locustv1.LocustTestSpec{
                MasterCommandSeed: "--locustfile /test.py",
                WorkerCommandSeed: "--locustfile /test.py",
                WorkerReplicas:    5,
            },
            wantErr: true,
            errMsg:  "image is required",
        },
        "replicas too high": {
            spec: locustv1.LocustTestSpec{
                MasterCommandSeed: "--locustfile /test.py",
                WorkerCommandSeed: "--locustfile /test.py",
                WorkerReplicas:    501,
                Image:             "locustio/locust:latest",
            },
            wantErr: true,
            errMsg:  "workerReplicas must be <= 500",
        },
    }
    
    for name, tc := range tests {
        t.Run(name, func(t *testing.T) {
            err := ValidateSpec(&tc.spec)
            
            if tc.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tc.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

---

## 3. Integration Testing with envtest

### 3.1 Suite Setup

```go
// internal/controller/suite_test.go
package controller

import (
    "context"
    "path/filepath"
    "testing"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    "k8s.io/client-go/kubernetes/scheme"
    "k8s.io/client-go/rest"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/envtest"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
    "sigs.k8s.io/controller-runtime/pkg/log/zap"
    
    locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
)

var (
    cfg       *rest.Config
    k8sClient client.Client
    testEnv   *envtest.Environment
    ctx       context.Context
    cancel    context.CancelFunc
)

func TestControllers(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
    logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
    
    ctx, cancel = context.WithCancel(context.TODO())
    
    By("bootstrapping test environment")
    testEnv = &envtest.Environment{
        CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
        ErrorIfCRDPathMissing: true,
        
        // Use existing cluster for debugging (optional)
        // UseExistingCluster: ptr.To(true),
    }
    
    var err error
    cfg, err = testEnv.Start()
    Expect(err).NotTo(HaveOccurred())
    Expect(cfg).NotTo(BeNil())
    
    err = locustv1.AddToScheme(scheme.Scheme)
    Expect(err).NotTo(HaveOccurred())
    
    // Create manager
    mgr, err := ctrl.NewManager(cfg, ctrl.Options{
        Scheme: scheme.Scheme,
    })
    Expect(err).NotTo(HaveOccurred())
    
    // Setup controller
    err = (&LocustTestReconciler{
        Client: mgr.GetClient(),
        Scheme: mgr.GetScheme(),
    }).SetupWithManager(mgr)
    Expect(err).NotTo(HaveOccurred())
    
    // Start manager in background
    go func() {
        defer GinkgoRecover()
        err = mgr.Start(ctx)
        Expect(err).NotTo(HaveOccurred())
    }()
    
    k8sClient = mgr.GetClient()
    Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
    cancel()
    By("tearing down the test environment")
    err := testEnv.Stop()
    Expect(err).NotTo(HaveOccurred())
})
```

### 3.2 Controller Tests with Ginkgo

```go
// internal/controller/locusttest_controller_test.go
package controller

import (
    "context"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
    
    locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
)

var _ = Describe("LocustTest Controller", func() {
    const (
        timeout  = time.Second * 30
        interval = time.Millisecond * 250
    )
    
    Context("When creating a LocustTest", func() {
        It("Should create master and worker Jobs", func() {
            ctx := context.Background()
            
            // Create namespace
            namespace := &corev1.Namespace{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "test-ns-" + randString(5),
                },
            }
            Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
            
            // Create LocustTest
            locustTest := &locustv1.LocustTest{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "demo.test",
                    Namespace: namespace.Name,
                },
                Spec: locustv1.LocustTestSpec{
                    MasterCommandSeed: "--locustfile /lotest/src/test.py --host https://example.com",
                    WorkerCommandSeed: "--locustfile /lotest/src/test.py",
                    WorkerReplicas:    3,
                    Image:             "locustio/locust:2.43.1",
                },
            }
            Expect(k8sClient.Create(ctx, locustTest)).To(Succeed())
            
            // Verify master Job is created
            masterJobKey := types.NamespacedName{
                Name:      "demo-test-master",
                Namespace: namespace.Name,
            }
            Eventually(func() error {
                return k8sClient.Get(ctx, masterJobKey, &batchv1.Job{})
            }, timeout, interval).Should(Succeed())
            
            // Verify worker Job is created
            workerJobKey := types.NamespacedName{
                Name:      "demo-test-worker",
                Namespace: namespace.Name,
            }
            Eventually(func() error {
                return k8sClient.Get(ctx, workerJobKey, &batchv1.Job{})
            }, timeout, interval).Should(Succeed())
            
            // Verify master Service is created
            serviceKey := types.NamespacedName{
                Name:      "demo-test-master",
                Namespace: namespace.Name,
            }
            Eventually(func() error {
                return k8sClient.Get(ctx, serviceKey, &corev1.Service{})
            }, timeout, interval).Should(Succeed())
            
            // Verify Job properties
            var masterJob batchv1.Job
            Expect(k8sClient.Get(ctx, masterJobKey, &masterJob)).To(Succeed())
            Expect(*masterJob.Spec.Parallelism).To(Equal(int32(1)))
            
            var workerJob batchv1.Job
            Expect(k8sClient.Get(ctx, workerJobKey, &workerJob)).To(Succeed())
            Expect(*workerJob.Spec.Parallelism).To(Equal(int32(3)))
        })
        
        It("Should be NO-OP on updates", func() {
            ctx := context.Background()
            
            namespace := &corev1.Namespace{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "test-noop-" + randString(5),
                },
            }
            Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
            
            locustTest := &locustv1.LocustTest{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "update.test",
                    Namespace: namespace.Name,
                },
                Spec: locustv1.LocustTestSpec{
                    MasterCommandSeed: "--locustfile /test.py",
                    WorkerCommandSeed: "--locustfile /test.py",
                    WorkerReplicas:    2,
                    Image:             "locustio/locust:2.43.1",
                },
            }
            Expect(k8sClient.Create(ctx, locustTest)).To(Succeed())
            
            // Wait for initial resources
            Eventually(func() error {
                return k8sClient.Get(ctx, types.NamespacedName{
                    Name:      "update-test-master",
                    Namespace: namespace.Name,
                }, &batchv1.Job{})
            }, timeout, interval).Should(Succeed())
            
            // Update the CR
            var lt locustv1.LocustTest
            Expect(k8sClient.Get(ctx, types.NamespacedName{
                Name:      "update.test",
                Namespace: namespace.Name,
            }, &lt)).To(Succeed())
            
            lt.Spec.WorkerReplicas = 10  // Try to change replicas
            Expect(k8sClient.Update(ctx, &lt)).To(Succeed())
            
            // Verify worker Job still has original replicas (NO-OP)
            Consistently(func() int32 {
                var workerJob batchv1.Job
                if err := k8sClient.Get(ctx, types.NamespacedName{
                    Name:      "update-test-worker",
                    Namespace: namespace.Name,
                }, &workerJob); err != nil {
                    return -1
                }
                return *workerJob.Spec.Parallelism
            }, time.Second*5, interval).Should(Equal(int32(2)))
        })
    })
    
    Context("When deleting a LocustTest", func() {
        It("Should delete all owned resources", func() {
            ctx := context.Background()
            
            namespace := &corev1.Namespace{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "test-delete-" + randString(5),
                },
            }
            Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
            
            locustTest := &locustv1.LocustTest{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "delete.test",
                    Namespace: namespace.Name,
                },
                Spec: locustv1.LocustTestSpec{
                    MasterCommandSeed: "--locustfile /test.py",
                    WorkerCommandSeed: "--locustfile /test.py",
                    WorkerReplicas:    2,
                    Image:             "locustio/locust:2.43.1",
                },
            }
            Expect(k8sClient.Create(ctx, locustTest)).To(Succeed())
            
            // Wait for resources
            Eventually(func() error {
                return k8sClient.Get(ctx, types.NamespacedName{
                    Name:      "delete-test-master",
                    Namespace: namespace.Name,
                }, &batchv1.Job{})
            }, timeout, interval).Should(Succeed())
            
            // Delete LocustTest
            Expect(k8sClient.Delete(ctx, locustTest)).To(Succeed())
            
            // Verify owned resources are deleted (owner reference cascade)
            Eventually(func() error {
                return k8sClient.Get(ctx, types.NamespacedName{
                    Name:      "delete-test-master",
                    Namespace: namespace.Name,
                }, &batchv1.Job{})
            }, timeout, interval).ShouldNot(Succeed())
        })
    })
})

func randString(n int) string {
    const letters = "abcdefghijklmnopqrstuvwxyz"
    b := make([]byte, n)
    for i := range b {
        b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
    }
    return string(b)
}
```

---

## 4. End-to-End Testing

### 4.1 Kind Cluster Setup

```go
// test/e2e/e2e_test.go
package e2e

import (
    "context"
    "os"
    "os/exec"
    "testing"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
)

var (
    clientset *kubernetes.Clientset
)

func TestE2E(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "E2E Suite")
}

var _ = BeforeSuite(func() {
    // Create Kind cluster if not exists
    cmd := exec.Command("kind", "create", "cluster", "--name", "locust-e2e")
    cmd.Stdout = GinkgoWriter
    cmd.Stderr = GinkgoWriter
    _ = cmd.Run() // Ignore error if cluster exists
    
    // Get kubeconfig
    kubeconfig := os.Getenv("HOME") + "/.kube/config"
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    Expect(err).NotTo(HaveOccurred())
    
    clientset, err = kubernetes.NewForConfig(config)
    Expect(err).NotTo(HaveOccurred())
    
    // Install CRDs
    cmd = exec.Command("kubectl", "apply", "-f", "../../config/crd/bases/")
    cmd.Stdout = GinkgoWriter
    cmd.Stderr = GinkgoWriter
    Expect(cmd.Run()).To(Succeed())
    
    // Deploy operator
    cmd = exec.Command("kubectl", "apply", "-k", "../../config/default/")
    cmd.Stdout = GinkgoWriter
    cmd.Stderr = GinkgoWriter
    Expect(cmd.Run()).To(Succeed())
    
    // Wait for operator to be ready
    Eventually(func() bool {
        pods, err := clientset.CoreV1().Pods("locust-system").List(
            context.TODO(),
            metav1.ListOptions{LabelSelector: "control-plane=controller-manager"},
        )
        if err != nil || len(pods.Items) == 0 {
            return false
        }
        return pods.Items[0].Status.Phase == corev1.PodRunning
    }, 2*time.Minute, 5*time.Second).Should(BeTrue())
})

var _ = AfterSuite(func() {
    // Cleanup (optional - keep cluster for debugging)
    // exec.Command("kind", "delete", "cluster", "--name", "locust-e2e").Run()
})
```

### 4.2 E2E Test Cases

```go
var _ = Describe("Locust Operator E2E", func() {
    Context("Full test lifecycle", func() {
        It("Should run a complete load test", func() {
            ctx := context.Background()
            
            // Create ConfigMap with test file
            configMap := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: locust-test-files
  namespace: default
data:
  locustfile.py: |
    from locust import HttpUser, task
    
    class QuickTest(HttpUser):
        @task
        def hello(self):
            self.client.get("/")
`
            cmd := exec.Command("kubectl", "apply", "-f", "-")
            cmd.Stdin = strings.NewReader(configMap)
            Expect(cmd.Run()).To(Succeed())
            
            // Create LocustTest CR
            locustTest := `
apiVersion: locust.io/v1
kind: LocustTest
metadata:
  name: e2e.demo-test
  namespace: default
spec:
  image: locustio/locust:2.43.1
  masterCommandSeed: --locustfile /lotest/src/locustfile.py --host https://httpbin.org --users 5 --spawn-rate 1 --run-time 30s
  workerCommandSeed: --locustfile /lotest/src/locustfile.py
  workerReplicas: 2
  configMap: locust-test-files
`
            cmd = exec.Command("kubectl", "apply", "-f", "-")
            cmd.Stdin = strings.NewReader(locustTest)
            Expect(cmd.Run()).To(Succeed())
            
            // Wait for Jobs to be created
            Eventually(func() int {
                jobs, _ := clientset.BatchV1().Jobs("default").List(ctx, metav1.ListOptions{
                    LabelSelector: "performance-test-name=e2e-demo-test",
                })
                return len(jobs.Items)
            }, 2*time.Minute, 5*time.Second).Should(Equal(2))
            
            // Wait for test completion
            Eventually(func() bool {
                job, err := clientset.BatchV1().Jobs("default").Get(ctx, "e2e-demo-test-master", metav1.GetOptions{})
                if err != nil {
                    return false
                }
                return job.Status.Succeeded > 0
            }, 5*time.Minute, 10*time.Second).Should(BeTrue())
            
            // Cleanup
            cmd = exec.Command("kubectl", "delete", "locusttest", "e2e.demo-test", "-n", "default")
            Expect(cmd.Run()).To(Succeed())
        })
    })
})
```

---

## 5. Test Fixtures & Utilities

### 5.1 Test Fixtures

```go
// internal/controller/testdata/fixtures.go
package testdata

import (
    locustv1 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewLocustTest creates a basic LocustTest for testing
func NewLocustTest(name, namespace string) *locustv1.LocustTest {
    return &locustv1.LocustTest{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
        },
        Spec: locustv1.LocustTestSpec{
            MasterCommandSeed: "--locustfile /lotest/src/test.py --host https://example.com",
            WorkerCommandSeed: "--locustfile /lotest/src/test.py",
            WorkerReplicas:    3,
            Image:             "locustio/locust:2.43.1",
        },
    }
}

// WithWorkerReplicas sets worker replicas
func WithWorkerReplicas(lt *locustv1.LocustTest, replicas int32) *locustv1.LocustTest {
    lt.Spec.WorkerReplicas = replicas
    return lt
}

// WithLabels sets labels
func WithLabels(lt *locustv1.LocustTest, master, worker map[string]string) *locustv1.LocustTest {
    lt.Spec.Labels = &locustv1.PodLabels{
        Master: master,
        Worker: worker,
    }
    return lt
}

// WithConfigMap sets configMap
func WithConfigMap(lt *locustv1.LocustTest, configMap string) *locustv1.LocustTest {
    lt.Spec.ConfigMap = configMap
    return lt
}
```

### 5.2 Test Utilities

```go
// internal/controller/test_helpers.go
package controller

import (
    "context"
    "time"
    
    "sigs.k8s.io/controller-runtime/pkg/client"
)

// WaitForResource waits for a resource to exist
func WaitForResource(ctx context.Context, c client.Client, obj client.Object, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); err == nil {
                return nil
            }
        }
    }
}

// WaitForResourceDeletion waits for a resource to be deleted
func WaitForResourceDeletion(ctx context.Context, c client.Client, obj client.Object, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := c.Get(ctx, client.ObjectKeyFromObject(obj), obj); client.IgnoreNotFound(err) == nil {
                return nil
            }
        }
    }
}
```

---

## 6. Mocking Strategies

### 6.1 Interface-Based Mocking

```go
// internal/resources/interfaces.go
package resources

import (
    "context"
    
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
)

// ResourceManager interface for testability
type ResourceManager interface {
    CreateJob(ctx context.Context, job *batchv1.Job) error
    CreateService(ctx context.Context, svc *corev1.Service) error
    DeleteJob(ctx context.Context, namespace, name string) error
    DeleteService(ctx context.Context, namespace, name string) error
}

// Mock implementation
type MockResourceManager struct {
    CreateJobFunc    func(ctx context.Context, job *batchv1.Job) error
    CreateServiceFunc func(ctx context.Context, svc *corev1.Service) error
    DeleteJobFunc    func(ctx context.Context, namespace, name string) error
    DeleteServiceFunc func(ctx context.Context, namespace, name string) error
    
    CreatedJobs     []*batchv1.Job
    CreatedServices []*corev1.Service
}

func (m *MockResourceManager) CreateJob(ctx context.Context, job *batchv1.Job) error {
    m.CreatedJobs = append(m.CreatedJobs, job)
    if m.CreateJobFunc != nil {
        return m.CreateJobFunc(ctx, job)
    }
    return nil
}

func (m *MockResourceManager) CreateService(ctx context.Context, svc *corev1.Service) error {
    m.CreatedServices = append(m.CreatedServices, svc)
    if m.CreateServiceFunc != nil {
        return m.CreateServiceFunc(ctx, svc)
    }
    return nil
}
```

### 6.2 Using testify/mock

```go
// internal/resources/mocks/resource_manager_mock.go
package mocks

import (
    "context"
    
    "github.com/stretchr/testify/mock"
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
)

type MockResourceManager struct {
    mock.Mock
}

func (m *MockResourceManager) CreateJob(ctx context.Context, job *batchv1.Job) error {
    args := m.Called(ctx, job)
    return args.Error(0)
}

func (m *MockResourceManager) CreateService(ctx context.Context, svc *corev1.Service) error {
    args := m.Called(ctx, svc)
    return args.Error(0)
}

// Usage in tests
func TestReconciler_CreateResources(t *testing.T) {
    mockRM := new(mocks.MockResourceManager)
    
    mockRM.On("CreateJob", mock.Anything, mock.AnythingOfType("*v1.Job")).Return(nil)
    mockRM.On("CreateService", mock.Anything, mock.AnythingOfType("*v1.Service")).Return(nil)
    
    reconciler := &LocustTestReconciler{
        ResourceManager: mockRM,
    }
    
    // Run test...
    
    mockRM.AssertExpectations(t)
    mockRM.AssertNumberOfCalls(t, "CreateJob", 2)  // master + worker
}
```

---

## 7. CI/CD Integration

### 7.1 GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./internal/...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: coverage.out
  
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Install envtest binaries
        run: make envtest
      
      - name: Run integration tests
        run: |
          KUBEBUILDER_ASSETS=$(./bin/setup-envtest use -p path) \
          go test -v ./internal/controller/... -coverprofile=coverage.out
  
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Create Kind cluster
        uses: helm/kind-action@v1
        with:
          cluster_name: e2e-test
      
      - name: Build and load operator image
        run: |
          make docker-build IMG=locust-operator:test
          kind load docker-image locust-operator:test --name e2e-test
      
      - name: Install CRDs
        run: make install
      
      - name: Deploy operator
        run: make deploy IMG=locust-operator:test
      
      - name: Run E2E tests
        run: go test -v ./test/e2e/...
```

---

## References

- [Ginkgo Documentation](https://onsi.github.io/ginkgo/)
- [Gomega Documentation](https://onsi.github.io/gomega/)
- [envtest Guide](https://book.kubebuilder.io/reference/envtest.html)
- [testify Library](https://github.com/stretchr/testify)
