# Phase 1: v1 API Types - Implementation Plan

**Effort:** 1 day  
**Priority:** P0 - Critical Path  
**Prerequisites:** Phase 0 complete  
**Requirements:** §4.4 v1 to v2 Field Mapping (v1 fields)

---

## Objective

Define Go types that exactly match the current Java v1 CRD for backward compatibility. The generated CRD must be schema-compatible with existing LocustTest resources.

---

## Reference: Java CRD Field Structure

From `/kube/crd/locust-test-crd.yaml`:

```yaml
spec:
  properties:
    masterCommandSeed: {type: string}              # REQUIRED
    workerCommandSeed: {type: string}              # REQUIRED  
    workerReplicas: {type: integer, min: 1, max: 500, default: 1}  # REQUIRED
    image: {type: string}                          # REQUIRED
    imagePullPolicy: {type: string, enum: [Always, IfNotPresent, Never]}
    imagePullSecrets: {type: array, items: string}
    configMap: {type: string}
    libConfigMap: {type: string}
    labels:
      master: {additionalProperties: string}
      worker: {additionalProperties: string}
    annotations:
      master: {additionalProperties: string}
      worker: {additionalProperties: string}
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution: {additionalProperties: string}
    tolerations: [{key, operator, value, effect}]
  required: [masterCommandSeed, workerCommandSeed, workerReplicas, image]
```

---

## Tasks

### Task 1.1: Define LocustTestSpec Struct

**File:** `api/v1/locusttest_types.go`

Replace the skeleton `LocustTestSpec` with the full v1 spec:

```go
// LocustTestSpec defines the desired state of LocustTest
type LocustTestSpec struct {
	// MasterCommandSeed is the command seed for the master pod.
	// This forms the base of the locust master command.
	// +kubebuilder:validation:Required
	MasterCommandSeed string `json:"masterCommandSeed"`

	// WorkerCommandSeed is the command seed for worker pods.
	// This forms the base of the locust worker command.
	// +kubebuilder:validation:Required
	WorkerCommandSeed string `json:"workerCommandSeed"`

	// WorkerReplicas is the number of worker pods to spawn.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=500
	// +kubebuilder:default=1
	WorkerReplicas int32 `json:"workerReplicas"`

	// Image is the Locust container image to use.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// ImagePullPolicy defines when to pull the image.
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +optional
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets is a list of secret names for pulling images from private registries.
	// +optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// ConfigMap is the name of the ConfigMap containing the test file(s).
	// +optional
	ConfigMap string `json:"configMap,omitempty"`

	// LibConfigMap is the name of the ConfigMap containing lib directory files.
	// +optional
	LibConfigMap string `json:"libConfigMap,omitempty"`

	// Labels defines labels to attach to deployed pods.
	// +optional
	Labels *PodLabels `json:"labels,omitempty"`

	// Annotations defines annotations to attach to deployed pods.
	// +optional
	Annotations *PodAnnotations `json:"annotations,omitempty"`

	// Affinity defines affinity rules for pod scheduling.
	// +optional
	Affinity *LocustTestAffinity `json:"affinity,omitempty"`

	// Tolerations defines tolerations for pod scheduling.
	// +optional
	Tolerations []LocustTestToleration `json:"tolerations,omitempty"`
}
```

**Key Points:**
- Use `+kubebuilder:validation:Required` for required fields
- Use `+optional` and `omitempty` for optional fields
- Match JSON field names exactly to Java CRD

---

### Task 1.2: Define Supporting Types

Add these types in the same file, **before** `LocustTestSpec`:

```go
// PodLabels defines labels for master and worker pods.
type PodLabels struct {
	// Master defines labels attached to the master pod.
	// +optional
	Master map[string]string `json:"master,omitempty"`

	// Worker defines labels attached to worker pods.
	// +optional
	Worker map[string]string `json:"worker,omitempty"`
}

// PodAnnotations defines annotations for master and worker pods.
type PodAnnotations struct {
	// Master defines annotations attached to the master pod.
	// +optional
	Master map[string]string `json:"master,omitempty"`

	// Worker defines annotations attached to worker pods.
	// +optional
	Worker map[string]string `json:"worker,omitempty"`
}

// LocustTestAffinity defines affinity rules for pod scheduling.
type LocustTestAffinity struct {
	// NodeAffinity defines node affinity rules.
	// +optional
	NodeAffinity *LocustTestNodeAffinity `json:"nodeAffinity,omitempty"`
}

// LocustTestNodeAffinity defines node affinity configuration.
type LocustTestNodeAffinity struct {
	// RequiredDuringSchedulingIgnoredDuringExecution defines required node affinity rules.
	// The map keys are label keys and values are label values that nodes must have.
	// +optional
	RequiredDuringSchedulingIgnoredDuringExecution map[string]string `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// LocustTestToleration defines a toleration for pod scheduling.
type LocustTestToleration struct {
	// Key is the taint key that the toleration applies to.
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// Operator represents the relationship between the key and value.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Exists;Equal
	Operator string `json:"operator"`

	// Value is the taint value the toleration matches to.
	// +optional
	Value string `json:"value,omitempty"`

	// Effect indicates the taint effect to match.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute
	Effect string `json:"effect"`
}
```

**Notes from NOTES.md:**
- Tolerations require: key, operator, effect (value is optional)
- Operator enum: "Exists", "Equal"
- Effect enum: "NoSchedule", "PreferNoSchedule", "NoExecute"

---

### Task 1.3: Add Kubebuilder Markers for CRD Metadata

Update the `LocustTest` type with printer columns and short name:

```go
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lotest
// +kubebuilder:printcolumn:name="master_cmd",type=string,JSONPath=`.spec.masterCommandSeed`,description="Master pod command seed"
// +kubebuilder:printcolumn:name="worker_replica_count",type=integer,JSONPath=`.spec.workerReplicas`,description="Number of requested worker pods"
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`,description="Locust image"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// LocustTest is the Schema for the locusttests API
type LocustTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocustTestSpec   `json:"spec,omitempty"`
	Status LocustTestStatus `json:"status,omitempty"`
}
```

**Printer Columns (from Java CRD):**
| Column | Type | JSONPath |
|--------|------|----------|
| master_cmd | string | .spec.masterCommandSeed |
| worker_replica_count | integer | .spec.workerReplicas |
| Image | string | .spec.image |
| Age | date | .metadata.creationTimestamp |

---

### Task 1.4: Keep LocustTestStatus Minimal (v1)

For v1 parity, keep status minimal (Java v1 has no status):

```go
// LocustTestStatus defines the observed state of LocustTest
type LocustTestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}
```

**Note:** Status fields will be added in Phase 9 for v2.

---

### Task 1.5: Verify groupversion_info.go

Ensure `api/v1/groupversion_info.go` has correct group:

```go
// Package v1 contains API Schema definitions for the locust v1 API group
// +kubebuilder:object:generate=true
// +groupName=locust.io
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "locust.io", Version: "v1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
```

**Key:** `Group: "locust.io"` must match the Java CRD group.

---

### Task 1.6: Generate Code and Manifests

```bash
cd locust-k8s-operator-go

# Generate DeepCopy methods
make generate

# Verify zz_generated.deepcopy.go was created/updated
ls -la api/v1/zz_generated.deepcopy.go

# Generate CRD manifests
make manifests

# View generated CRD
cat config/crd/bases/locust.io_locusttests.yaml
```

---

### Task 1.7: Compare Generated CRD with Java CRD

Create a comparison script or manually compare:

```bash
# Extract schema sections for comparison
# Java CRD
cat /path/to/kube/crd/locust-test-crd.yaml | yq '.spec.versions[0].schema.openAPIV3Schema.properties.spec'

# Go CRD
cat config/crd/bases/locust.io_locusttests.yaml | yq '.spec.versions[0].schema.openAPIV3Schema.properties.spec'
```

**Must Match:**
- Field names
- Field types
- Required fields list
- Enum values
- Min/Max constraints
- Default values

---

### Task 1.8: Test Sample CR Validation

```bash
# Apply the CRD to a test cluster
kubectl apply -f config/crd/bases/locust.io_locusttests.yaml

# Test sample CR validates
kubectl apply --dry-run=server -f /path/to/kube/sample-cr/locust-test-cr.yaml

# Expected: no validation errors
```

---

### Task 1.9: Write Type Tests (Optional but Recommended)

**File:** `api/v1/locusttest_types_test.go`

```go
package v1

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocustTestSpec_JSONRoundTrip(t *testing.T) {
	spec := LocustTestSpec{
		MasterCommandSeed: "--locustfile /lotest/src/test.py --host https://example.com",
		WorkerCommandSeed: "--locustfile /lotest/src/test.py",
		WorkerReplicas:    3,
		Image:             "locustio/locust:latest",
		ImagePullPolicy:   "Always",
		ConfigMap:         "test-config",
		Labels: &PodLabels{
			Master: map[string]string{"app": "locust-master"},
			Worker: map[string]string{"app": "locust-worker"},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(spec)
	require.NoError(t, err)

	// Unmarshal back
	var decoded LocustTestSpec
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify roundtrip
	assert.Equal(t, spec.MasterCommandSeed, decoded.MasterCommandSeed)
	assert.Equal(t, spec.WorkerReplicas, decoded.WorkerReplicas)
	assert.Equal(t, spec.Labels.Master["app"], decoded.Labels.Master["app"])
}

func TestLocustTestSpec_JSONFieldNames(t *testing.T) {
	spec := LocustTestSpec{
		MasterCommandSeed: "test",
		WorkerCommandSeed: "test",
		WorkerReplicas:    1,
		Image:             "test",
	}

	data, err := json.Marshal(spec)
	require.NoError(t, err)

	// Verify camelCase JSON field names
	jsonStr := string(data)
	assert.Contains(t, jsonStr, `"masterCommandSeed"`)
	assert.Contains(t, jsonStr, `"workerCommandSeed"`)
	assert.Contains(t, jsonStr, `"workerReplicas"`)
}
```

Run tests:
```bash
go test ./api/v1/... -v
```

---

## Complete locusttest_types.go

Here's the complete file structure:

```go
/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
...
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodLabels defines labels for master and worker pods.
type PodLabels struct {
	// Master defines labels attached to the master pod.
	// +optional
	Master map[string]string `json:"master,omitempty"`

	// Worker defines labels attached to worker pods.
	// +optional
	Worker map[string]string `json:"worker,omitempty"`
}

// PodAnnotations defines annotations for master and worker pods.
type PodAnnotations struct {
	// Master defines annotations attached to the master pod.
	// +optional
	Master map[string]string `json:"master,omitempty"`

	// Worker defines annotations attached to worker pods.
	// +optional
	Worker map[string]string `json:"worker,omitempty"`
}

// LocustTestAffinity defines affinity rules for pod scheduling.
type LocustTestAffinity struct {
	// NodeAffinity defines node affinity rules.
	// +optional
	NodeAffinity *LocustTestNodeAffinity `json:"nodeAffinity,omitempty"`
}

// LocustTestNodeAffinity defines node affinity configuration.
type LocustTestNodeAffinity struct {
	// RequiredDuringSchedulingIgnoredDuringExecution defines required node affinity rules.
	// The map keys are label keys and values are label values that nodes must have.
	// +optional
	RequiredDuringSchedulingIgnoredDuringExecution map[string]string `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// LocustTestToleration defines a toleration for pod scheduling.
type LocustTestToleration struct {
	// Key is the taint key that the toleration applies to.
	// +kubebuilder:validation:Required
	Key string `json:"key"`

	// Operator represents the relationship between the key and value.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Exists;Equal
	Operator string `json:"operator"`

	// Value is the taint value the toleration matches to.
	// +optional
	Value string `json:"value,omitempty"`

	// Effect indicates the taint effect to match.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute
	Effect string `json:"effect"`
}

// LocustTestSpec defines the desired state of LocustTest
type LocustTestSpec struct {
	// MasterCommandSeed is the command seed for the master pod.
	// This forms the base of the locust master command.
	// +kubebuilder:validation:Required
	MasterCommandSeed string `json:"masterCommandSeed"`

	// WorkerCommandSeed is the command seed for worker pods.
	// This forms the base of the locust worker command.
	// +kubebuilder:validation:Required
	WorkerCommandSeed string `json:"workerCommandSeed"`

	// WorkerReplicas is the number of worker pods to spawn.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=500
	// +kubebuilder:default=1
	WorkerReplicas int32 `json:"workerReplicas"`

	// Image is the Locust container image to use.
	// +kubebuilder:validation:Required
	Image string `json:"image"`

	// ImagePullPolicy defines when to pull the image.
	// +kubebuilder:validation:Enum=Always;IfNotPresent;Never
	// +optional
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets is a list of secret names for pulling images from private registries.
	// +optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// ConfigMap is the name of the ConfigMap containing the test file(s).
	// +optional
	ConfigMap string `json:"configMap,omitempty"`

	// LibConfigMap is the name of the ConfigMap containing lib directory files.
	// +optional
	LibConfigMap string `json:"libConfigMap,omitempty"`

	// Labels defines labels to attach to deployed pods.
	// +optional
	Labels *PodLabels `json:"labels,omitempty"`

	// Annotations defines annotations to attach to deployed pods.
	// +optional
	Annotations *PodAnnotations `json:"annotations,omitempty"`

	// Affinity defines affinity rules for pod scheduling.
	// +optional
	Affinity *LocustTestAffinity `json:"affinity,omitempty"`

	// Tolerations defines tolerations for pod scheduling.
	// +optional
	Tolerations []LocustTestToleration `json:"tolerations,omitempty"`
}

// LocustTestStatus defines the observed state of LocustTest
type LocustTestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lotest
// +kubebuilder:printcolumn:name="master_cmd",type=string,JSONPath=`.spec.masterCommandSeed`,description="Master pod command seed"
// +kubebuilder:printcolumn:name="worker_replica_count",type=integer,JSONPath=`.spec.workerReplicas`,description="Number of requested worker pods"
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`,description="Locust image"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// LocustTest is the Schema for the locusttests API
type LocustTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocustTestSpec   `json:"spec,omitempty"`
	Status LocustTestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// LocustTestList contains a list of LocustTest
type LocustTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocustTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocustTest{}, &LocustTestList{})
}
```

---

## Acceptance Criteria Verification

| Criteria | Command | Expected |
|----------|---------|----------|
| CRD schema matches Java | Compare YAML schemas | Field names, types, constraints match |
| Sample CR validates | `kubectl apply --dry-run=server` | No validation errors |
| DeepCopy generated | `ls api/v1/zz_generated.deepcopy.go` | File exists and updated |
| Short name works | `kubectl get lotest` | Lists LocustTest resources |
| Printer columns | `kubectl get lotest` | Shows master_cmd, worker_replica_count, Image, Age |

---

## Troubleshooting

### Issue: CRD validation mismatch with Java
**Symptom:** Sample CR fails validation against Go CRD
**Solution:** Compare schemas field-by-field, check JSON tags and kubebuilder markers

### Issue: `make generate` fails
**Symptom:** controller-gen errors
**Solution:** Check for syntax errors in kubebuilder markers, ensure markers are on the line directly above the field

### Issue: Missing fields in generated CRD
**Symptom:** Some Go struct fields not appearing in CRD
**Solution:** Ensure all fields have `json:"fieldName"` tags

### Issue: Short name not working
**Symptom:** `kubectl get lotest` returns error
**Solution:** Verify `+kubebuilder:resource:shortName=lotest` marker is on the LocustTest type

---

## Notes for Next Phase

Phase 2 (Configuration System) and Phase 3 (Resource Builders) will use these types to:
1. Build Kubernetes Job and Service resources from LocustTestSpec
2. Configure resource limits, metrics exporter, and other operator settings

The types defined here are the foundation for all resource building logic.
