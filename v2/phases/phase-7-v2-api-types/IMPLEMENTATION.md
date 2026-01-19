# Phase 7: v2 API Types - Implementation Guide

**Estimated Effort:** 1.5 days  
**Prerequisites:** Phase 4 complete (Core Reconciler working)

---

## Table of Contents

1. [Day 1: API Scaffold & Core Types](#day-1-api-scaffold--core-types)
2. [Day 2: New Feature Types & Validation](#day-2-new-feature-types--validation)
3. [Verification](#verification)

---

## Day 1: API Scaffold & Core Types

### Task 7.1: Create v2 API Scaffold

Use Operator SDK to scaffold the v2 API structure.

```bash
cd /path/to/locust-k8s-operator-go

# Create v2 API without controller (we'll update existing controller)
operator-sdk create api \
  --group locust \
  --version v2 \
  --kind LocustTest \
  --resource \
  --controller=false
```

This creates:
- `api/v2/groupversion_info.go`
- `api/v2/locusttest_types.go`

### Task 7.2: Define `api/v2/groupversion_info.go`

Replace the scaffolded file with proper group/version info:

```go
/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
...
*/

// Package v2 contains API Schema definitions for the locust v2 API group.
// +kubebuilder:object:generate=true
// +groupName=locust.io
package v2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects.
	GroupVersion = schema.GroupVersion{Group: "locust.io", Version: "v2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
```

### Task 7.3: Define Core Spec Types in `api/v2/locusttest_types.go`

#### 7.3.1 Package Header and Imports

```go
/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
...
*/

package v2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
```

#### 7.3.2 MasterSpec - Grouped Master Configuration

```go
// MasterSpec defines master node configuration.
type MasterSpec struct {
	// Command is the base command for the master node.
	// The operator appends: --master --master-port=5557 --expect-workers=N
	// +kubebuilder:validation:Required
	Command string `json:"command"`

	// Resources defines resource requests and limits for the master pod.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Labels for the master pod.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for the master pod.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Autostart enables the --autostart flag to start the test automatically.
	// +optional
	// +kubebuilder:default=true
	Autostart *bool `json:"autostart,omitempty"`

	// Autoquit configuration for automatic test termination.
	// +optional
	Autoquit *AutoquitConfig `json:"autoquit,omitempty"`

	// ExtraArgs are additional CLI arguments appended to the command.
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`
}

// AutoquitConfig defines autoquit behavior for the master.
type AutoquitConfig struct {
	// Enabled enables the --autoquit flag.
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`

	// Timeout in seconds after test completion before quitting.
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=60
	Timeout int32 `json:"timeout,omitempty"`
}
```

#### 7.3.3 WorkerSpec - Grouped Worker Configuration

```go
// WorkerSpec defines worker node configuration.
type WorkerSpec struct {
	// Command is the base command for worker nodes.
	// The operator appends: --worker --master-host=<service> --master-port=5557
	// +kubebuilder:validation:Required
	Command string `json:"command"`

	// Replicas is the number of worker pods to create.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=500
	Replicas int32 `json:"replicas"`

	// Resources defines resource requests and limits for worker pods.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Labels for worker pods.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations for worker pods.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// ExtraArgs are additional CLI arguments appended to the command.
	// +optional
	ExtraArgs []string `json:"extraArgs,omitempty"`
}
```

#### 7.3.4 TestFilesConfig - Renamed ConfigMap Fields

```go
// TestFilesConfig defines test file mounting configuration.
type TestFilesConfig struct {
	// ConfigMapRef is the name of the ConfigMap containing locustfile(s).
	// +optional
	ConfigMapRef string `json:"configMapRef,omitempty"`

	// LibConfigMapRef is the name of the ConfigMap containing library files.
	// +optional
	LibConfigMapRef string `json:"libConfigMapRef,omitempty"`

	// SrcMountPath is the mount path for test files.
	// +optional
	// +kubebuilder:default="/lotest/src"
	SrcMountPath string `json:"srcMountPath,omitempty"`

	// LibMountPath is the mount path for library files.
	// +optional
	// +kubebuilder:default="/opt/locust/lib"
	LibMountPath string `json:"libMountPath,omitempty"`
}
```

#### 7.3.5 SchedulingConfig - Grouped Scheduling

```go
// SchedulingConfig defines pod scheduling configuration.
type SchedulingConfig struct {
	// Affinity rules for pod scheduling.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Tolerations for pod scheduling.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// NodeSelector for pod scheduling.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}
```

---

## Day 2: New Feature Types & Validation

### Task 7.4: Define New Feature Types

#### 7.4.1 EnvConfig - Environment Injection (Issue #149)

```go
// EnvConfig defines environment variable injection configuration.
type EnvConfig struct {
	// ConfigMapRefs injects all keys from ConfigMaps as environment variables.
	// +optional
	ConfigMapRefs []ConfigMapEnvSource `json:"configMapRefs,omitempty"`

	// SecretRefs injects all keys from Secrets as environment variables.
	// +optional
	SecretRefs []SecretEnvSource `json:"secretRefs,omitempty"`

	// Variables defines specific environment variables.
	// +optional
	Variables []corev1.EnvVar `json:"variables,omitempty"`

	// SecretMounts mounts secrets as files in the container.
	// +optional
	SecretMounts []SecretMount `json:"secretMounts,omitempty"`
}

// ConfigMapEnvSource defines a ConfigMap environment source.
type ConfigMapEnvSource struct {
	// Name of the ConfigMap.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Prefix to add to all keys when injecting as env vars.
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// SecretEnvSource defines a Secret environment source.
type SecretEnvSource struct {
	// Name of the Secret.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Prefix to add to all keys when injecting as env vars.
	// +optional
	Prefix string `json:"prefix,omitempty"`
}

// SecretMount defines a secret file mount.
type SecretMount struct {
	// Name of the secret to mount.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// MountPath is the path where the secret should be mounted.
	// +kubebuilder:validation:Required
	MountPath string `json:"mountPath"`

	// ReadOnly mounts the secret as read-only.
	// +optional
	// +kubebuilder:default=true
	ReadOnly bool `json:"readOnly,omitempty"`
}
```

#### 7.4.2 VolumeMount - Extended with Target (Issue #252)

```go
// TargetedVolumeMount extends VolumeMount with target pod selection.
type TargetedVolumeMount struct {
	corev1.VolumeMount `json:",inline"`

	// Target specifies which pods receive this mount.
	// +optional
	// +kubebuilder:validation:Enum=master;worker;both
	// +kubebuilder:default=both
	Target string `json:"target,omitempty"`
}
```

#### 7.4.3 ObservabilityConfig - OpenTelemetry (Issue #72)

```go
// ObservabilityConfig defines observability settings.
type ObservabilityConfig struct {
	// OpenTelemetry configuration for native Locust OTel integration.
	// +optional
	OpenTelemetry *OpenTelemetryConfig `json:"openTelemetry,omitempty"`
}

// OpenTelemetryConfig defines OpenTelemetry integration settings.
type OpenTelemetryConfig struct {
	// Enabled enables OpenTelemetry integration.
	// When true, adds --otel flag to Locust command.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`

	// Endpoint is the OTel collector endpoint (e.g., "otel-collector:4317").
	// Required when Enabled is true.
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// Protocol for OTel export.
	// +optional
	// +kubebuilder:validation:Enum=grpc;http/protobuf
	// +kubebuilder:default=grpc
	Protocol string `json:"protocol,omitempty"`

	// Insecure skips TLS verification for the collector connection.
	// +optional
	// +kubebuilder:default=false
	Insecure bool `json:"insecure,omitempty"`

	// ExtraEnvVars for additional OTel SDK configuration.
	// +optional
	ExtraEnvVars map[string]string `json:"extraEnvVars,omitempty"`
}
```

### Task 7.5: Define LocustTestStatus

```go
// LocustTestStatus defines the observed state of LocustTest.
type LocustTestStatus struct {
	// Phase is the current lifecycle phase of the test.
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
	// +optional
	Phase string `json:"phase,omitempty"`

	// ExpectedWorkers is the number of workers expected to connect.
	// +optional
	ExpectedWorkers int32 `json:"expectedWorkers,omitempty"`

	// ConnectedWorkers is the current number of connected workers.
	// +optional
	ConnectedWorkers int32 `json:"connectedWorkers,omitempty"`

	// StartTime is when the test started.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when the test completed.
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Conditions represent the latest available observations of the test's state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}
```

### Task 7.6: Define Complete LocustTestSpec

```go
// LocustTestSpec defines the desired state of LocustTest.
type LocustTestSpec struct {
	// ============================================
	// IMAGE CONFIGURATION
	// ============================================

	// Image is the container image for Locust pods.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// ImagePullPolicy for the Locust container.
	// +optional
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +kubebuilder:default=IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets for pulling from private registries.
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// ============================================
	// MASTER CONFIGURATION
	// ============================================

	// Master configuration for the master node.
	// +kubebuilder:validation:Required
	Master MasterSpec `json:"master"`

	// ============================================
	// WORKER CONFIGURATION
	// ============================================

	// Worker configuration for worker nodes.
	// +kubebuilder:validation:Required
	Worker WorkerSpec `json:"worker"`

	// ============================================
	// TEST FILES
	// ============================================

	// TestFiles configuration for locustfile and library mounting.
	// +optional
	TestFiles *TestFilesConfig `json:"testFiles,omitempty"`

	// ============================================
	// SCHEDULING
	// ============================================

	// Scheduling configuration for pod placement.
	// +optional
	Scheduling *SchedulingConfig `json:"scheduling,omitempty"`

	// ============================================
	// ENVIRONMENT INJECTION
	// ============================================

	// Env configuration for environment variable injection.
	// +optional
	Env *EnvConfig `json:"env,omitempty"`

	// ============================================
	// VOLUME MOUNTING
	// ============================================

	// Volumes to add to pods.
	// +optional
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// VolumeMounts for the locust container with target selection.
	// +optional
	VolumeMounts []TargetedVolumeMount `json:"volumeMounts,omitempty"`

	// ============================================
	// OBSERVABILITY
	// ============================================

	// Observability configuration for metrics and tracing.
	// +optional
	Observability *ObservabilityConfig `json:"observability,omitempty"`
}
```

### Task 7.7: Define LocustTest Root Type with Markers

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lotest
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="Current test phase"
// +kubebuilder:printcolumn:name="Workers",type=integer,JSONPath=`.spec.worker.replicas`,description="Requested worker count"
// +kubebuilder:printcolumn:name="Connected",type=integer,JSONPath=`.status.connectedWorkers`,description="Connected workers"
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// LocustTest is the Schema for the locusttests API.
type LocustTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocustTestSpec   `json:"spec,omitempty"`
	Status LocustTestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LocustTestList contains a list of LocustTest.
type LocustTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocustTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocustTest{}, &LocustTestList{})
}
```

### Task 7.8: Create `api/v2/conditions.go`

Define condition type and reason constants:

```go
/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
...
*/

package v2

// Condition types for LocustTest.
const (
	// ConditionTypeReady indicates all resources are created and ready.
	ConditionTypeReady = "Ready"

	// ConditionTypeWorkersConnected indicates workers have connected to master.
	ConditionTypeWorkersConnected = "WorkersConnected"

	// ConditionTypeTestCompleted indicates the test has finished.
	ConditionTypeTestCompleted = "TestCompleted"
)

// Condition reasons for Ready condition.
const (
	ReasonResourcesCreating = "ResourcesCreating"
	ReasonResourcesCreated  = "ResourcesCreated"
	ReasonResourcesFailed   = "ResourcesFailed"
)

// Condition reasons for WorkersConnected condition.
const (
	ReasonWaitingForWorkers   = "WaitingForWorkers"
	ReasonAllWorkersConnected = "AllWorkersConnected"
	ReasonWorkersMissing      = "WorkersMissing"
)

// Condition reasons for TestCompleted condition.
const (
	ReasonTestInProgress = "TestInProgress"
	ReasonTestSucceeded  = "TestSucceeded"
	ReasonTestFailed     = "TestFailed"
)

// Phase constants for LocustTest status.
const (
	PhasePending   = "Pending"
	PhaseRunning   = "Running"
	PhaseSucceeded = "Succeeded"
	PhaseFailed    = "Failed"
)
```

### Task 7.9: Update `cmd/main.go` to Register v2

Add v2 scheme registration:

```go
import (
	// ... existing imports
	locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(locustv1.AddToScheme(scheme))
	utilruntime.Must(locustv2.AddToScheme(scheme))  // Add v2
	// +kubebuilder:scaffold:scheme
}
```

### Task 7.10: Create Sample v2 CR

Create `config/samples/locust_v2_locusttest.yaml`:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: sample-test
  namespace: default
spec:
  image: locustio/locust:2.43.1
  imagePullPolicy: IfNotPresent
  
  master:
    command: "locust -f /lotest/src/locustfile.py"
    autostart: true
    autoquit:
      enabled: true
      timeout: 60
    labels:
      role: master
    resources:
      requests:
        cpu: "500m"
        memory: "512Mi"
      limits:
        cpu: "1000m"
        memory: "1Gi"
  
  worker:
    command: "locust -f /lotest/src/locustfile.py"
    replicas: 5
    labels:
      role: worker
    resources:
      requests:
        cpu: "250m"
        memory: "256Mi"
      limits:
        cpu: "500m"
        memory: "512Mi"
  
  testFiles:
    configMapRef: my-locust-tests
  
  scheduling:
    nodeSelector:
      workload-type: testing
```

---

## Verification

### Generate and Validate

```bash
# Generate DeepCopy methods
make generate

# Generate CRD manifests
make manifests

# Verify CRD was generated
ls -la config/crd/bases/

# Check CRD contains v2 with storageversion
grep -A5 "storageVersion" config/crd/bases/locust.io_locusttests.yaml

# Verify project builds
make build

# Run tests (should still pass)
make test
```

### Validate Sample CR

```bash
# Dry-run validation
kubectl apply --dry-run=client -f config/samples/locust_v2_locusttest.yaml

# Or validate against CRD schema
kubectl apply --dry-run=server -f config/samples/locust_v2_locusttest.yaml
```

### Check Printer Columns

```bash
# After applying a v2 CR
kubectl get locusttests

# Expected output:
# NAME          PHASE     WORKERS   CONNECTED   AGE
# sample-test   Pending   5         0           5s
```

---

## Common Issues

### Issue: "v2 scheme not registered"

**Solution:** Ensure `locustv2.AddToScheme(scheme)` is called in `cmd/main.go` before manager starts.

### Issue: "CRD shows only v1"

**Solution:** 
1. Check `+kubebuilder:storageversion` marker on v2 LocustTest
2. Run `make manifests` to regenerate

### Issue: "DeepCopy missing for new types"

**Solution:** 
1. Ensure all types have `+kubebuilder:object:generate=true` in package doc
2. Run `make generate`

### Issue: "Validation fails on sample CR"

**Solution:** Check that all required fields have values and enum values match exactly.

---

## Next Steps

After Phase 7 completion:
- **Phase 8:** Implement v1↔v2 conversion webhook
- **Phase 9:** Implement status update logic in reconciler
- **Phase 10-12:** Implement new feature logic (env, volumes, OTel)
