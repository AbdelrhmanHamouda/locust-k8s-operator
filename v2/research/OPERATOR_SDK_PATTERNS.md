# Operator SDK Patterns & Best Practices

**Research Date:** January 2026  
**Operator SDK Version:** v1.37+ (latest stable)  
**controller-runtime Version:** v0.19+

---

## Table of Contents

1. [Framework Architecture](#1-framework-architecture)
2. [Project Scaffolding](#2-project-scaffolding)
3. [Reconciliation Patterns](#3-reconciliation-patterns)
4. [Resource Management](#4-resource-management)
5. [Status Management](#5-status-management)
6. [Error Handling & Retry](#6-error-handling--retry)
7. [Configuration Management](#7-configuration-management)
8. [Metrics & Observability](#8-metrics--observability)

---

## 1. Framework Architecture

### 1.1 Layer Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Operator SDK CLI                         │
│  (Scaffolding, Bundle Generation, OLM Integration)          │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    controller-runtime                        │
│  (Manager, Controller, Client, Cache, Webhook)              │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                       client-go                              │
│  (REST Client, Informers, Work Queues, Discovery)           │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Kubernetes API Server                       │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 Key Components

| Component | Responsibility | Java Equivalent |
|-----------|---------------|-----------------|
| **Manager** | Lifecycle, shared caches, leader election | Micronaut + JOSDK Operator |
| **Controller** | Watch resources, queue events | JOSDK Reconciler |
| **Reconciler** | Business logic implementation | JOSDK Reconciler.reconcile() |
| **Client** | CRUD operations on resources | Fabric8 KubernetesClient |
| **Scheme** | Type registration | JOSDK CRD registration |

---

## 2. Project Scaffolding

### 2.1 Initialization Commands

```bash
# Initialize project
operator-sdk init \
  --domain locust.io \
  --repo github.com/AbdelrhmanHamouda/locust-k8s-operator \
  --plugins go/v4

# Create API (CRD + Controller)
operator-sdk create api \
  --group locust \
  --version v1 \
  --kind LocustTest \
  --resource \
  --controller

# Create additional version for API migration
operator-sdk create api \
  --group locust \
  --version v2 \
  --kind LocustTest \
  --resource \
  --controller=false

# Create webhook for conversion
operator-sdk create webhook \
  --group locust \
  --version v1 \
  --kind LocustTest \
  --conversion
```

### 2.2 Generated Project Structure

```
/
├── cmd/
│   └── main.go                    # Entry point
├── api/
│   ├── v1/
│   │   ├── locusttest_types.go    # CRD types
│   │   ├── groupversion_info.go   # API group registration
│   │   ├── locusttest_webhook.go  # Webhooks (if created)
│   │   └── zz_generated.deepcopy.go
│   └── v2/
│       └── ...                    # v2 API types
├── internal/
│   └── controller/
│       ├── locusttest_controller.go
│       └── locusttest_controller_test.go
├── config/
│   ├── crd/
│   │   └── bases/                 # Generated CRDs
│   ├── rbac/
│   │   ├── role.yaml
│   │   └── role_binding.yaml
│   ├── manager/
│   │   └── manager.yaml
│   └── samples/                   # Example CRs
├── hack/
│   └── boilerplate.go.txt
├── Dockerfile
├── Makefile
├── go.mod
└── PROJECT                        # Operator SDK metadata
```

---

## 3. Reconciliation Patterns

### 3.1 Standard Reconciler Interface

```go
// Reconciler interface from controller-runtime
type Reconciler interface {
    Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
}
```

### 3.2 Reconciler Implementation Pattern

```go
type LocustTestReconciler struct {
    client.Client                          // Embedded client
    Scheme   *runtime.Scheme               // Type scheme
    Recorder record.EventRecorder          // Event recording
    Config   *config.OperatorConfig        // Operator configuration
}

func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    
    // 1. Fetch the resource
    locustTest := &locustv1.LocustTest{}
    if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
        if apierrors.IsNotFound(err) {
            // Resource deleted - nothing to do (cleanup via finalizers/ownership)
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }
    
    // 2. Check if being deleted (finalizer pattern)
    if !locustTest.DeletionTimestamp.IsZero() {
        return r.reconcileDelete(ctx, locustTest)
    }
    
    // 3. Add finalizer if not present
    if !controllerutil.ContainsFinalizer(locustTest, finalizerName) {
        controllerutil.AddFinalizer(locustTest, finalizerName)
        if err := r.Update(ctx, locustTest); err != nil {
            return ctrl.Result{}, err
        }
    }
    
    // 4. Main reconciliation logic
    return r.reconcileNormal(ctx, locustTest)
}
```

### 3.3 NO-OP on Updates Pattern (Matching Java Behavior)

```go
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    
    locustTest := &locustv1.LocustTest{}
    if err := r.Get(ctx, req.NamespacedName, locustTest); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    // NO-OP on updates - matching Java behavior
    // Generation > 1 means the spec has been modified after creation
    if locustTest.Generation > 1 {
        log.Info("Update detected - NO-OP by design",
            "name", locustTest.Name,
            "generation", locustTest.Generation)
        return ctrl.Result{}, nil
    }
    
    // Only process on initial creation
    log.Info("LocustTest created", "name", locustTest.Name)
    
    // Create resources...
    return r.createResources(ctx, locustTest)
}
```

### 3.4 Controller Setup

```go
func (r *LocustTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&locustv1.LocustTest{}).
        Owns(&batchv1.Job{}).           // Watch owned Jobs
        Owns(&corev1.Service{}).        // Watch owned Services
        WithEventFilter(predicate.GenerationChangedPredicate{}).
        Complete(r)
}
```

---

## 4. Resource Management

### 4.1 Owner References (Automatic Cleanup)

```go
func (r *LocustTestReconciler) createJob(ctx context.Context, lt *locustv1.LocustTest, job *batchv1.Job) error {
    // Set owner reference for automatic garbage collection
    if err := controllerutil.SetControllerReference(lt, job, r.Scheme); err != nil {
        return err
    }
    
    return r.Create(ctx, job)
}
```

### 4.2 CreateOrUpdate Pattern

```go
func (r *LocustTestReconciler) reconcileService(ctx context.Context, lt *locustv1.LocustTest) error {
    service := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{
            Name:      lt.Name + "-master",
            Namespace: lt.Namespace,
        },
    }
    
    op, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
        // Mutate function - called on both create and update
        service.Spec.Selector = map[string]string{
            "app": lt.Name + "-master",
        }
        service.Spec.Ports = []corev1.ServicePort{
            {Name: "master", Port: 5557, TargetPort: intstr.FromInt(5557)},
            {Name: "master-bind", Port: 5558, TargetPort: intstr.FromInt(5558)},
            {Name: "web-ui", Port: 8089, TargetPort: intstr.FromInt(8089)},
        }
        return controllerutil.SetControllerReference(lt, service, r.Scheme)
    })
    
    if err != nil {
        return err
    }
    
    log.Info("Service reconciled", "operation", op)
    return nil
}
```

### 4.3 Server-Side Apply Pattern

```go
func (r *LocustTestReconciler) applyJob(ctx context.Context, job *batchv1.Job) error {
    // Using Patch with Apply strategy for idempotent updates
    return r.Patch(ctx, job, client.Apply, client.FieldOwner("locust-operator"))
}
```

---

## 5. Status Management

### 5.1 Status Subresource Pattern

```go
// In types file
type LocustTestStatus struct {
    // Phase represents the current lifecycle phase
    // +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
    Phase string `json:"phase,omitempty"`
    
    // Expected number of workers
    ExpectedWorkers int32 `json:"expectedWorkers,omitempty"`
    
    // Connected workers count
    ConnectedWorkers int32 `json:"connectedWorkers,omitempty"`
    
    // Start time of the test
    StartTime *metav1.Time `json:"startTime,omitempty"`
    
    // Completion time of the test
    CompletionTime *metav1.Time `json:"completionTime,omitempty"`
    
    // Standard Kubernetes conditions
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Workers",type=string,JSONPath=`.status.connectedWorkers`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type LocustTest struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    
    Spec   LocustTestSpec   `json:"spec,omitempty"`
    Status LocustTestStatus `json:"status,omitempty"`
}
```

### 5.2 Updating Status

```go
func (r *LocustTestReconciler) updateStatus(ctx context.Context, lt *locustv1.LocustTest, phase string) error {
    lt.Status.Phase = phase
    
    // Use StatusClient to update only status subresource
    return r.Status().Update(ctx, lt)
}

// Using conditions
func (r *LocustTestReconciler) setCondition(ctx context.Context, lt *locustv1.LocustTest) error {
    meta.SetStatusCondition(&lt.Status.Conditions, metav1.Condition{
        Type:               "Ready",
        Status:             metav1.ConditionTrue,
        Reason:             "AllResourcesCreated",
        Message:            "Master and worker jobs are running",
        LastTransitionTime: metav1.Now(),
    })
    
    return r.Status().Update(ctx, lt)
}
```

---

## 6. Error Handling & Retry

### 6.1 Return Patterns

```go
// Immediate requeue (typically on transient errors)
return ctrl.Result{Requeue: true}, nil

// Requeue after duration
return ctrl.Result{RequeueAfter: 30 * time.Second}, nil

// Don't requeue (success or permanent failure)
return ctrl.Result{}, nil

// Error - will trigger exponential backoff requeue
return ctrl.Result{}, err
```

### 6.2 Error Classification

```go
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ...
    
    if err := r.createMasterJob(ctx, lt); err != nil {
        if apierrors.IsAlreadyExists(err) {
            // Not an error - resource already exists
            log.V(1).Info("Master job already exists")
        } else if apierrors.IsConflict(err) {
            // Conflict - requeue immediately
            return ctrl.Result{Requeue: true}, nil
        } else {
            // Actual error - trigger backoff
            return ctrl.Result{}, err
        }
    }
    
    return ctrl.Result{}, nil
}
```

### 6.3 Rate Limiting Configuration

```go
// In main.go
mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
    // ...
    Controller: config.Controller{
        // Maximum concurrent reconciles
        MaxConcurrentReconciles: 1,
        // Rate limiter configuration
        RateLimiter: workqueue.NewMaxOfRateLimiter(
            workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 1000*time.Second),
            &workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
        ),
    },
})
```

---

## 7. Configuration Management

### 7.1 Environment-Based Configuration

```go
// internal/config/config.go
package config

import (
    "os"
    "strconv"
    "time"
)

type OperatorConfig struct {
    // Pod resource defaults
    PodCPURequest    string
    PodMemRequest    string
    PodCPULimit      string
    PodMemLimit      string
    
    // Job configuration
    TTLSecondsAfterFinished *int32
    
    // Metrics exporter
    MetricsExporterImage string
    MetricsExporterPort  int32
    
    // Feature flags
    EnableAffinityCRInjection    bool
    EnableTolerationsCRInjection bool
}

func LoadConfig() *OperatorConfig {
    return &OperatorConfig{
        PodCPURequest:    getEnv("POD_CPU_REQUEST", "250m"),
        PodMemRequest:    getEnv("POD_MEM_REQUEST", "128Mi"),
        PodCPULimit:      getEnv("POD_CPU_LIMIT", ""),
        PodMemLimit:      getEnv("POD_MEM_LIMIT", ""),
        
        TTLSecondsAfterFinished: getEnvInt32Ptr("JOB_TTL_SECONDS_AFTER_FINISHED"),
        
        MetricsExporterImage: getEnv("METRICS_EXPORTER_IMAGE", "containersol/locust_exporter:v0.5.0"),
        MetricsExporterPort:  getEnvInt32("METRICS_EXPORTER_PORT", 9646),
        
        EnableAffinityCRInjection:    getEnvBool("ENABLE_AFFINITY_CR_INJECTION", false),
        EnableTolerationsCRInjection: getEnvBool("ENABLE_TOLERATIONS_CR_INJECTION", false),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        b, _ := strconv.ParseBool(value)
        return b
    }
    return defaultValue
}

func getEnvInt32(key string, defaultValue int32) int32 {
    if value := os.Getenv(key); value != "" {
        i, _ := strconv.ParseInt(value, 10, 32)
        return int32(i)
    }
    return defaultValue
}

func getEnvInt32Ptr(key string) *int32 {
    if value := os.Getenv(key); value != "" {
        i, _ := strconv.ParseInt(value, 10, 32)
        v := int32(i)
        return &v
    }
    return nil
}
```

---

## 8. Metrics & Observability

### 8.1 Built-in Metrics

controller-runtime exposes metrics automatically:

| Metric | Description |
|--------|-------------|
| `controller_runtime_reconcile_total` | Total reconciliations |
| `controller_runtime_reconcile_errors_total` | Reconciliation errors |
| `controller_runtime_reconcile_time_seconds` | Reconciliation duration |
| `workqueue_depth` | Current queue depth |
| `workqueue_adds_total` | Total items added to queue |

### 8.2 Custom Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
    locustTestsActive = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "locust_tests_active",
            Help: "Number of active LocustTest resources",
        },
        []string{"namespace"},
    )
    
    locustWorkersTotal = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "locust_workers_total",
            Help: "Total workers across all active tests",
        },
        []string{"namespace", "test"},
    )
)

func init() {
    metrics.Registry.MustRegister(locustTestsActive, locustWorkersTotal)
}
```

### 8.3 Structured Logging

```go
import (
    "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    
    // Info level
    log.Info("Reconciling LocustTest",
        "name", req.Name,
        "namespace", req.Namespace)
    
    // Debug level (V(1))
    log.V(1).Info("Detailed reconciliation info",
        "generation", lt.Generation,
        "workerReplicas", lt.Spec.WorkerReplicas)
    
    // Error level
    log.Error(err, "Failed to create Job",
        "job", jobName)
    
    return ctrl.Result{}, nil
}
```

---

## References

- [Operator SDK Documentation](https://sdk.operatorframework.io/docs/)
- [controller-runtime Documentation](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [client-go Examples](https://github.com/kubernetes/client-go/tree/master/examples)
