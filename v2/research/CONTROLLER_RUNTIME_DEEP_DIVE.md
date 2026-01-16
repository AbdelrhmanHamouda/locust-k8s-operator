# Controller Runtime Deep Dive

**Research Date:** January 2026  
**Version:** controller-runtime v0.19+

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Manager Component](#2-manager-component)
3. [Controller & Reconciler](#3-controller--reconciler)
4. [Client Abstraction](#4-client-abstraction)
5. [Cache & Informers](#5-cache--informers)
6. [Event Handling](#6-event-handling)
7. [Predicates & Filtering](#7-predicates--filtering)
8. [Finalizers](#8-finalizers)

---

## 1. Architecture Overview

### 1.1 Component Hierarchy

```
┌─────────────────────────────────────────────────────────────────┐
│                           Manager                                │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  - Shared Cache (read-optimized)                          │  │
│  │  - Client (API Server communication)                      │  │
│  │  - Scheme (type registration)                             │  │
│  │  - Leader Election                                        │  │
│  │  - Metrics Server                                         │  │
│  │  - Health Probes                                          │  │
│  └───────────────────────────────────────────────────────────┘  │
│                              │                                   │
│  ┌───────────────────────────┼───────────────────────────────┐  │
│  │          Controller 1     │           Controller 2         │  │
│  │  ┌─────────────────────┐  │  ┌─────────────────────────┐  │  │
│  │  │   Work Queue        │  │  │   Work Queue            │  │  │
│  │  │   Event Handlers    │  │  │   Event Handlers        │  │  │
│  │  │   Predicates        │  │  │   Predicates            │  │  │
│  │  └─────────┬───────────┘  │  └─────────┬───────────────┘  │  │
│  │            │              │            │                   │  │
│  │  ┌─────────▼───────────┐  │  ┌─────────▼───────────────┐  │  │
│  │  │   Reconciler        │  │  │   Reconciler            │  │  │
│  │  │   (Your Logic)      │  │  │   (Your Logic)          │  │  │
│  │  └─────────────────────┘  │  └─────────────────────────┘  │  │
│  └───────────────────────────┴───────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Request Flow

```
API Server
    │
    ▼ (Watch events)
┌────────────────┐
│  Informer/     │
│  Cache         │
└───────┬────────┘
        │ (Event)
        ▼
┌────────────────┐
│  Event Handler │──────┐
└───────┬────────┘      │
        │               │ (Filter)
        ▼               │
┌────────────────┐      │
│  Predicate     │◄─────┘
└───────┬────────┘
        │ (Enqueue)
        ▼
┌────────────────┐
│  Work Queue    │
└───────┬────────┘
        │ (Dequeue)
        ▼
┌────────────────┐
│  Reconciler    │
│  (Your Code)   │
└────────────────┘
```

---

## 2. Manager Component

### 2.1 Manager Configuration

```go
import (
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/healthz"
    metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

func main() {
    mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
        // Scheme contains all registered types
        Scheme: scheme,
        
        // Metrics server configuration
        Metrics: metricsserver.Options{
            BindAddress:   ":8080",
            SecureServing: false,
        },
        
        // Health probe configuration
        HealthProbeBindAddress: ":8081",
        
        // Leader election for HA deployments
        LeaderElection:   true,
        LeaderElectionID: "locust-operator-leader",
        
        // Namespace restriction (empty = all namespaces)
        // Cache: cache.Options{
        //     DefaultNamespaces: map[string]cache.Config{
        //         "my-namespace": {},
        //     },
        // },
        
        // Controller configuration
        Controller: config.Controller{
            MaxConcurrentReconciles: 1,
        },
    })
    if err != nil {
        setupLog.Error(err, "unable to create manager")
        os.Exit(1)
    }
    
    // Add health checks
    if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
        setupLog.Error(err, "unable to set up health check")
        os.Exit(1)
    }
    if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
        setupLog.Error(err, "unable to set up ready check")
        os.Exit(1)
    }
    
    // Start manager (blocks)
    if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
        setupLog.Error(err, "unable to start manager")
        os.Exit(1)
    }
}
```

### 2.2 Manager Responsibilities

| Responsibility | Description |
|---------------|-------------|
| **Lifecycle** | Start/stop all controllers, webhooks |
| **Leader Election** | Ensure single active instance in HA |
| **Shared Cache** | Single cache for all controllers |
| **Client** | Provides read (cache) and write (API) client |
| **Metrics** | Exposes Prometheus metrics endpoint |
| **Health** | Liveness and readiness probes |

---

## 3. Controller & Reconciler

### 3.1 Controller Builder Pattern

```go
func (r *LocustTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        // Primary resource to watch
        For(&locustv1.LocustTest{}).
        
        // Secondary resources (owned by primary)
        Owns(&batchv1.Job{}).
        Owns(&corev1.Service{}).
        
        // Watch external resources with custom handler
        Watches(
            &corev1.ConfigMap{},
            handler.EnqueueRequestsFromMapFunc(r.findLocustTestsForConfigMap),
        ).
        
        // Event filtering
        WithEventFilter(predicate.GenerationChangedPredicate{}).
        
        // Controller options
        WithOptions(controller.Options{
            MaxConcurrentReconciles: 1,
            RateLimiter:             workqueue.DefaultControllerRateLimiter(),
        }).
        
        Complete(r)
}

// Custom mapping function for Watches
func (r *LocustTestReconciler) findLocustTestsForConfigMap(ctx context.Context, cm client.Object) []reconcile.Request {
    locustTests := &locustv1.LocustTestList{}
    if err := r.List(ctx, locustTests, client.InNamespace(cm.GetNamespace())); err != nil {
        return nil
    }
    
    var requests []reconcile.Request
    for _, lt := range locustTests.Items {
        if lt.Spec.ConfigMap == cm.GetName() {
            requests = append(requests, reconcile.Request{
                NamespacedName: types.NamespacedName{
                    Name:      lt.Name,
                    Namespace: lt.Namespace,
                },
            })
        }
    }
    return requests
}
```

### 3.2 Reconcile Result Semantics

```go
// Success - no requeue needed
return ctrl.Result{}, nil

// Requeue immediately
return ctrl.Result{Requeue: true}, nil

// Requeue after specific duration
return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil

// Error - triggers exponential backoff requeue
return ctrl.Result{}, err

// Note: Requeue: true with RequeueAfter set uses the duration
return ctrl.Result{Requeue: true, RequeueAfter: 30 * time.Second}, nil
```

### 3.3 Reconciliation Best Practices

```go
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)
    
    // PATTERN 1: Always fetch fresh resource
    lt := &locustv1.LocustTest{}
    if err := r.Get(ctx, req.NamespacedName, lt); err != nil {
        // Not found is normal (deleted before reconcile)
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    // PATTERN 2: Deep copy before mutation
    original := lt.DeepCopy()
    
    // PATTERN 3: Defer status update
    defer func() {
        if !equality.Semantic.DeepEqual(original.Status, lt.Status) {
            if err := r.Status().Update(ctx, lt); err != nil {
                log.Error(err, "Failed to update status")
            }
        }
    }()
    
    // PATTERN 4: Use helper methods for clarity
    if result, err := r.reconcileMasterService(ctx, lt); err != nil || result.Requeue {
        return result, err
    }
    
    if result, err := r.reconcileMasterJob(ctx, lt); err != nil || result.Requeue {
        return result, err
    }
    
    if result, err := r.reconcileWorkerJob(ctx, lt); err != nil || result.Requeue {
        return result, err
    }
    
    return ctrl.Result{}, nil
}
```

---

## 4. Client Abstraction

### 4.1 Client Types

| Client | Source | Use Case |
|--------|--------|----------|
| **Read (Get/List)** | Cache | Fast, eventually consistent reads |
| **Write (Create/Update/Patch/Delete)** | API Server | Mutations |
| **Status** | API Server | Status subresource updates |

### 4.2 Client Operations

```go
// Get a single resource
lt := &locustv1.LocustTest{}
err := r.Get(ctx, types.NamespacedName{Name: "test", Namespace: "default"}, lt)

// List resources with options
jobs := &batchv1.JobList{}
err := r.List(ctx, jobs,
    client.InNamespace("default"),
    client.MatchingLabels{"app": "locust"},
    client.Limit(100),
)

// Create
err := r.Create(ctx, job)

// Update (full resource)
err := r.Update(ctx, job)

// Patch (partial update)
patch := client.MergeFrom(original)
err := r.Patch(ctx, modified, patch)

// Server-side apply
err := r.Patch(ctx, job, client.Apply, client.FieldOwner("locust-operator"))

// Delete
err := r.Delete(ctx, job)

// Delete with options
err := r.Delete(ctx, job, client.PropagationPolicy(metav1.DeletePropagationForeground))

// Status update
err := r.Status().Update(ctx, lt)
```

### 4.3 Direct API Client (Bypass Cache)

```go
// For cases requiring fresh data from API server
directClient, err := client.New(mgr.GetConfig(), client.Options{Scheme: scheme})

// Or use APIReader from manager
err := mgr.GetAPIReader().Get(ctx, key, obj)
```

---

## 5. Cache & Informers

### 5.1 How Cache Works

```
                              ┌─────────────────────┐
                              │   API Server        │
                              └──────────┬──────────┘
                                         │
                                         │ Watch
                                         ▼
┌────────────────────────────────────────────────────────────────┐
│                          Cache                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Informer   │  │   Informer   │  │   Informer   │   ...   │
│  │  (LocustTest)│  │    (Job)     │  │  (Service)   │         │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘         │
│         │                 │                 │                  │
│         ▼                 ▼                 ▼                  │
│  ┌──────────────────────────────────────────────────────┐     │
│  │                    Indexer (Store)                    │     │
│  │   - By Namespace                                      │     │
│  │   - By Labels                                         │     │
│  │   - Custom Indexes                                    │     │
│  └──────────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────────┘
                                         │
                                         │ Read
                                         ▼
                              ┌──────────────────────┐
                              │    Reconciler        │
                              └──────────────────────┘
```

### 5.2 Custom Cache Indexes

```go
// Add custom index in main.go
if err := mgr.GetFieldIndexer().IndexField(
    context.Background(),
    &locustv1.LocustTest{},
    ".spec.configMap",
    func(obj client.Object) []string {
        lt := obj.(*locustv1.LocustTest)
        if lt.Spec.ConfigMap == "" {
            return nil
        }
        return []string{lt.Spec.ConfigMap}
    },
); err != nil {
    return err
}

// Use index in reconciler
locustTests := &locustv1.LocustTestList{}
err := r.List(ctx, locustTests,
    client.InNamespace(namespace),
    client.MatchingFields{".spec.configMap": configMapName},
)
```

---

## 6. Event Handling

### 6.1 Event Handlers

```go
// EnqueueRequestForObject - enqueue the object itself
handler.EnqueueRequestForObject{}

// EnqueueRequestForOwner - enqueue the owner of the object
handler.EnqueueRequestForOwner(
    mgr.GetScheme(),
    mgr.GetRESTMapper(),
    &locustv1.LocustTest{},
)

// EnqueueRequestsFromMapFunc - custom mapping
handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
    // Return list of LocustTests to reconcile based on this object
    return []reconcile.Request{{NamespacedName: types.NamespacedName{...}}}
})
```

### 6.2 Recording Events

```go
type LocustTestReconciler struct {
    client.Client
    Scheme   *runtime.Scheme
    Recorder record.EventRecorder  // Event recorder
}

// Setup in main.go
reconciler := &controller.LocustTestReconciler{
    Client:   mgr.GetClient(),
    Scheme:   mgr.GetScheme(),
    Recorder: mgr.GetEventRecorderFor("locust-controller"),
}

// Use in reconciler
func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    // ...
    
    // Normal event
    r.Recorder.Event(lt, corev1.EventTypeNormal, "Created", "Master job created successfully")
    
    // Warning event
    r.Recorder.Eventf(lt, corev1.EventTypeWarning, "Failed", "Failed to create worker job: %v", err)
    
    // With annotations
    r.Recorder.AnnotatedEventf(lt, map[string]string{
        "workerCount": strconv.Itoa(int(lt.Spec.WorkerReplicas)),
    }, corev1.EventTypeNormal, "Scaled", "Scaled workers to %d", lt.Spec.WorkerReplicas)
}
```

---

## 7. Predicates & Filtering

### 7.1 Built-in Predicates

```go
import "sigs.k8s.io/controller-runtime/pkg/predicate"

// Only trigger on generation changes (spec changes, not status)
predicate.GenerationChangedPredicate{}

// Only trigger on label changes
predicate.LabelChangedPredicate{}

// Only trigger on annotation changes
predicate.AnnotationChangedPredicate{}

// Combine predicates
predicate.And(
    predicate.GenerationChangedPredicate{},
    predicate.LabelChangedPredicate{},
)

predicate.Or(
    predicate.GenerationChangedPredicate{},
    customPredicate,
)

// Negate predicate
predicate.Not(predicate.GenerationChangedPredicate{})
```

### 7.2 Custom Predicates

```go
// Custom predicate implementation
type NamespacePredicate struct {
    Namespaces []string
}

func (p NamespacePredicate) Create(e event.CreateEvent) bool {
    return p.matches(e.Object.GetNamespace())
}

func (p NamespacePredicate) Update(e event.UpdateEvent) bool {
    return p.matches(e.ObjectNew.GetNamespace())
}

func (p NamespacePredicate) Delete(e event.DeleteEvent) bool {
    return p.matches(e.Object.GetNamespace())
}

func (p NamespacePredicate) Generic(e event.GenericEvent) bool {
    return p.matches(e.Object.GetNamespace())
}

func (p NamespacePredicate) matches(namespace string) bool {
    for _, ns := range p.Namespaces {
        if ns == namespace {
            return true
        }
    }
    return false
}

// Using funcs for simpler predicates
predicate.Funcs{
    CreateFunc: func(e event.CreateEvent) bool {
        return e.Object.GetLabels()["managed-by"] == "locust-operator"
    },
    UpdateFunc: func(e event.UpdateEvent) bool {
        return e.ObjectNew.GetLabels()["managed-by"] == "locust-operator"
    },
    DeleteFunc: func(e event.DeleteEvent) bool {
        return true // Always process deletes
    },
}
```

---

## 8. Finalizers

### 8.1 Finalizer Pattern

```go
const finalizerName = "locust.io/finalizer"

func (r *LocustTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    lt := &locustv1.LocustTest{}
    if err := r.Get(ctx, req.NamespacedName, lt); err != nil {
        return ctrl.Result{}, client.IgnoreNotFound(err)
    }
    
    // Check if being deleted
    if !lt.DeletionTimestamp.IsZero() {
        // Resource is being deleted
        if controllerutil.ContainsFinalizer(lt, finalizerName) {
            // Run cleanup logic
            if err := r.cleanupExternalResources(ctx, lt); err != nil {
                return ctrl.Result{}, err
            }
            
            // Remove finalizer to allow deletion
            controllerutil.RemoveFinalizer(lt, finalizerName)
            if err := r.Update(ctx, lt); err != nil {
                return ctrl.Result{}, err
            }
        }
        return ctrl.Result{}, nil
    }
    
    // Add finalizer if not present
    if !controllerutil.ContainsFinalizer(lt, finalizerName) {
        controllerutil.AddFinalizer(lt, finalizerName)
        if err := r.Update(ctx, lt); err != nil {
            return ctrl.Result{}, err
        }
        // Requeue to continue with reconciliation
        return ctrl.Result{Requeue: true}, nil
    }
    
    // Normal reconciliation
    return r.reconcileNormal(ctx, lt)
}

func (r *LocustTestReconciler) cleanupExternalResources(ctx context.Context, lt *locustv1.LocustTest) error {
    // Clean up external resources not managed by owner references
    // e.g., external monitoring, cloud resources, etc.
    return nil
}
```

### 8.2 When to Use Finalizers

| Use Case | Finalizer Needed? |
|----------|-------------------|
| Child resources with owner references | **No** - automatic GC |
| External cloud resources | **Yes** |
| Custom cleanup logic | **Yes** |
| Metric/monitoring deregistration | **Yes** |
| Cross-namespace resources | **Yes** |

---

## References

- [controller-runtime pkg docs](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Kubebuilder Book](https://book.kubebuilder.io/)
- [controller-runtime source](https://github.com/kubernetes-sigs/controller-runtime)
