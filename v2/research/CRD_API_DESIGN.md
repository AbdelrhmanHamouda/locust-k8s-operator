# CRD API Design & Versioning

**Research Date:** January 2026  
**Focus:** API design patterns for the Locust K8s Operator v2

---

## Table of Contents

1. [API Versioning Strategy](#1-api-versioning-strategy)
2. [Type Definitions with Kubebuilder Markers](#2-type-definitions-with-kubebuilder-markers)
3. [Conversion Webhooks](#3-conversion-webhooks)
4. [Validation](#4-validation)
5. [Status Subresource Design](#5-status-subresource-design)
6. [Printer Columns](#6-printer-columns)

---

## 1. API Versioning Strategy

### 1.1 Version Progression

```
v1 (Current Java)  →  v1 (Go, deprecated)  →  v2 (Go, storage)
      │                       │                      │
      │                       │                      └── New features
      │                       └── Conversion webhook support
      └── Direct field mapping
```

### 1.2 Storage Version Strategy

```go
// api/v1/groupversion_info.go
// +kubebuilder:object:generate=true
// +groupName=locust.io
package v1

import (
    "k8s.io/apimachinery/pkg/runtime/schema"
    "sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
    GroupVersion = schema.GroupVersion{Group: "locust.io", Version: "v1"}
    SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
    AddToScheme = SchemeBuilder.AddToScheme
)

// api/v2/groupversion_info.go
// +kubebuilder:object:generate=true
// +groupName=locust.io
package v2

var (
    GroupVersion = schema.GroupVersion{Group: "locust.io", Version: "v2"}
    // v2 is the storage version
)
```

### 1.3 Hub and Spoke Pattern

```
                    ┌─────────────────┐
                    │   v2 (Hub)      │
                    │ Storage Version │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
              ▼              ▼              ▼
        ┌──────────┐  ┌──────────┐  ┌──────────┐
        │ v1       │  │ v1beta1  │  │ v3       │
        │ (Spoke)  │  │ (Future) │  │ (Future) │
        └──────────┘  └──────────┘  └──────────┘
```

---

## 2. Type Definitions with Kubebuilder Markers

### 2.1 Complete v2 Type Definition

```go
// api/v2/locusttest_types.go
package v2

import (
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LocustTestSpec defines the desired state of LocustTest
type LocustTestSpec struct {
    // ============================================
    // IMAGE CONFIGURATION
    // ============================================

    // Image is the container image for Locust pods
    // +kubebuilder:validation:Required
    Image string `json:"image"`

    // ImagePullPolicy for the Locust container
    // +optional
    // +kubebuilder:validation:Enum=Always;IfNotPresent;Never
    // +kubebuilder:default=IfNotPresent
    ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

    // ImagePullSecrets for pulling from private registries
    // +optional
    ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

    // ============================================
    // MASTER CONFIGURATION (grouped)
    // ============================================

    // Master configuration for the master node
    // +kubebuilder:validation:Required
    Master MasterSpec `json:"master"`

    // ============================================
    // WORKER CONFIGURATION (grouped)
    // ============================================

    // Worker configuration for worker nodes
    // +kubebuilder:validation:Required
    Worker WorkerSpec `json:"worker"`

    // ============================================
    // TEST FILES
    // ============================================

    // TestFiles configuration for locustfile and library mounting
    // +optional
    TestFiles *TestFilesConfig `json:"testFiles,omitempty"`

    // ============================================
    // SCHEDULING
    // ============================================

    // Scheduling configuration for pod placement
    // +optional
    Scheduling *SchedulingConfig `json:"scheduling,omitempty"`

    // ============================================
    // ENVIRONMENT INJECTION
    // ============================================

    // Env configuration for environment variable injection
    // +optional
    Env *EnvConfig `json:"env,omitempty"`

    // ============================================
    // VOLUME MOUNTING
    // ============================================

    // Volumes to mount to pods
    // +optional
    Volumes []corev1.Volume `json:"volumes,omitempty"`

    // VolumeMounts for the locust container
    // +optional
    VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`

    // ============================================
    // OBSERVABILITY
    // ============================================

    // Observability configuration
    // +optional
    Observability *ObservabilityConfig `json:"observability,omitempty"`
}

// MasterSpec defines master node configuration
type MasterSpec struct {
    // Command is the base command for the master node
    // Operator appends: --master --master-port=5557 --expect-workers=N
    // +kubebuilder:validation:Required
    Command string `json:"command"`

    // Resources for the master pod
    // +optional
    Resources corev1.ResourceRequirements `json:"resources,omitempty"`

    // Labels for the master pod
    // +optional
    Labels map[string]string `json:"labels,omitempty"`

    // Annotations for the master pod
    // +optional
    Annotations map[string]string `json:"annotations,omitempty"`

    // Autostart enables --autostart flag
    // +optional
    // +kubebuilder:default=true
    Autostart *bool `json:"autostart,omitempty"`

    // Autoquit configuration
    // +optional
    Autoquit *AutoquitConfig `json:"autoquit,omitempty"`

    // ExtraArgs are additional CLI arguments
    // +optional
    ExtraArgs []string `json:"extraArgs,omitempty"`
}

// WorkerSpec defines worker node configuration
type WorkerSpec struct {
    // Command is the base command for worker nodes
    // Operator appends: --worker --master-host=<service> --master-port=5557
    // +kubebuilder:validation:Required
    Command string `json:"command"`

    // Replicas is the number of worker pods
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=500
    Replicas int32 `json:"replicas"`

    // Resources for worker pods
    // +optional
    Resources corev1.ResourceRequirements `json:"resources,omitempty"`

    // Labels for worker pods
    // +optional
    Labels map[string]string `json:"labels,omitempty"`

    // Annotations for worker pods
    // +optional
    Annotations map[string]string `json:"annotations,omitempty"`

    // ExtraArgs are additional CLI arguments
    // +optional
    ExtraArgs []string `json:"extraArgs,omitempty"`
}

// TestFilesConfig defines test file mounting
type TestFilesConfig struct {
    // ConfigMapRef is the name of the ConfigMap containing locustfile(s)
    // +optional
    ConfigMapRef string `json:"configMapRef,omitempty"`

    // LibConfigMapRef is the name of the ConfigMap containing library files
    // +optional
    LibConfigMapRef string `json:"libConfigMapRef,omitempty"`

    // SrcMountPath is the mount path for test files
    // +optional
    // +kubebuilder:default="/lotest/src"
    SrcMountPath string `json:"srcMountPath,omitempty"`

    // LibMountPath is the mount path for library files
    // +optional
    // +kubebuilder:default="/opt/locust/lib"
    LibMountPath string `json:"libMountPath,omitempty"`
}

// SchedulingConfig defines pod scheduling configuration
type SchedulingConfig struct {
    // Affinity rules for pod scheduling
    // +optional
    Affinity *corev1.Affinity `json:"affinity,omitempty"`

    // Tolerations for pod scheduling
    // +optional
    Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

    // NodeSelector for pod scheduling
    // +optional
    NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// EnvConfig defines environment variable injection
type EnvConfig struct {
    // ConfigMapRefs injects all keys from ConfigMaps as env vars
    // +optional
    ConfigMapRefs []ConfigMapEnvSource `json:"configMapRefs,omitempty"`

    // SecretRefs injects all keys from Secrets as env vars
    // +optional
    SecretRefs []SecretEnvSource `json:"secretRefs,omitempty"`

    // Variables for specific value injection
    // +optional
    Variables []corev1.EnvVar `json:"variables,omitempty"`

    // SecretMounts mounts secrets as files
    // +optional
    SecretMounts []SecretMount `json:"secretMounts,omitempty"`
}

// ConfigMapEnvSource defines a ConfigMap env source with optional prefix
type ConfigMapEnvSource struct {
    // Name of the ConfigMap
    // +kubebuilder:validation:Required
    Name string `json:"name"`

    // Prefix to add to all keys
    // +optional
    Prefix string `json:"prefix,omitempty"`
}

// SecretEnvSource defines a Secret env source with optional prefix
type SecretEnvSource struct {
    // Name of the Secret
    // +kubebuilder:validation:Required
    Name string `json:"name"`

    // Prefix to add to all keys
    // +optional
    Prefix string `json:"prefix,omitempty"`
}

// SecretMount defines a secret file mount
type SecretMount struct {
    // Name of the secret
    // +kubebuilder:validation:Required
    Name string `json:"name"`

    // MountPath is the path to mount the secret
    // +kubebuilder:validation:Required
    MountPath string `json:"mountPath"`

    // ReadOnly mounts the secret as read-only
    // +optional
    // +kubebuilder:default=true
    ReadOnly bool `json:"readOnly,omitempty"`
}

// VolumeMount extends core VolumeMount with target selection
type VolumeMount struct {
    corev1.VolumeMount `json:",inline"`

    // Target specifies which pods get this mount
    // +optional
    // +kubebuilder:validation:Enum=master;worker;both
    // +kubebuilder:default=both
    Target string `json:"target,omitempty"`
}

// AutoquitConfig defines autoquit behavior
type AutoquitConfig struct {
    // Enabled enables --autoquit flag
    // +kubebuilder:default=true
    Enabled bool `json:"enabled"`

    // Timeout in seconds after test completion
    // +optional
    // +kubebuilder:validation:Minimum=0
    // +kubebuilder:default=60
    Timeout int32 `json:"timeout,omitempty"`
}

// ObservabilityConfig defines observability settings
type ObservabilityConfig struct {
    // OpenTelemetry configuration
    // +optional
    OpenTelemetry *OpenTelemetryConfig `json:"openTelemetry,omitempty"`
}

// OpenTelemetryConfig defines OpenTelemetry settings
type OpenTelemetryConfig struct {
    // Enabled enables OpenTelemetry integration
    // +kubebuilder:default=false
    Enabled bool `json:"enabled"`

    // Endpoint is the OTel collector endpoint
    // +optional
    Endpoint string `json:"endpoint,omitempty"`

    // Protocol for OTel export (grpc or http/protobuf)
    // +optional
    // +kubebuilder:validation:Enum=grpc;http/protobuf
    // +kubebuilder:default=grpc
    Protocol string `json:"protocol,omitempty"`

    // Insecure skips TLS verification
    // +optional
    // +kubebuilder:default=false
    Insecure bool `json:"insecure,omitempty"`

    // ExtraEnvVars for additional OTel configuration
    // +optional
    ExtraEnvVars map[string]string `json:"extraEnvVars,omitempty"`
}

// LocustTestStatus defines the observed state of LocustTest
type LocustTestStatus struct {
    // Phase is the current lifecycle phase
    // +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
    Phase string `json:"phase,omitempty"`

    // ExpectedWorkers is the number of expected worker connections
    ExpectedWorkers int32 `json:"expectedWorkers,omitempty"`

    // ConnectedWorkers is the number of connected workers
    ConnectedWorkers int32 `json:"connectedWorkers,omitempty"`

    // StartTime is when the test started
    // +optional
    StartTime *metav1.Time `json:"startTime,omitempty"`

    // CompletionTime is when the test completed
    // +optional
    CompletionTime *metav1.Time `json:"completionTime,omitempty"`

    // Conditions represent the latest observations
    // +optional
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=lotest
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Workers",type=integer,JSONPath=`.spec.workerReplicas`
// +kubebuilder:printcolumn:name="Connected",type=integer,JSONPath=`.status.connectedWorkers`
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:storageversion

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

### 2.2 v1 Type Definition (Deprecated)

```go
// api/v1/locusttest_types.go
package v1

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LocustTestSpec defines the desired state of LocustTest (v1 - deprecated)
type LocustTestSpec struct {
    // +kubebuilder:validation:Required
    MasterCommandSeed string `json:"masterCommandSeed"`

    // +kubebuilder:validation:Required
    WorkerCommandSeed string `json:"workerCommandSeed"`

    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=500
    WorkerReplicas int32 `json:"workerReplicas"`

    // +kubebuilder:validation:Required
    Image string `json:"image"`

    // +optional
    ConfigMap string `json:"configMap,omitempty"`

    // +optional
    LibConfigMap string `json:"libConfigMap,omitempty"`

    // +optional
    Labels map[string]map[string]string `json:"labels,omitempty"`

    // +optional
    Annotations map[string]map[string]string `json:"annotations,omitempty"`

    // +optional
    Affinity *LocustTestAffinity `json:"affinity,omitempty"`

    // +optional
    Tolerations []LocustTestToleration `json:"tolerations,omitempty"`

    // +optional
    ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

    // +optional
    ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
}

// LocustTestAffinity matches Java implementation
type LocustTestAffinity struct {
    NodeAffinity *LocustTestNodeAffinity `json:"nodeAffinity,omitempty"`
}

type LocustTestNodeAffinity struct {
    RequiredDuringSchedulingIgnoredDuringExecution map[string]string `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

type LocustTestToleration struct {
    Key      string `json:"key"`
    Operator string `json:"operator"`
    Value    string `json:"value,omitempty"`
    Effect   string `json:"effect"`
}

// +kubebuilder:object:root=true
// +kubebuilder:deprecatedversion:warning="locust.io/v1 LocustTest is deprecated, use locust.io/v2 instead"

type LocustTest struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    Spec LocustTestSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

type LocustTestList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []LocustTest `json:"items"`
}
```

---

## 3. Conversion Webhooks

### 3.1 Hub Interface (v2)

```go
// api/v2/locusttest_conversion.go
package v2

// Hub marks v2 as the hub version for conversions
func (*LocustTest) Hub() {}
```

### 3.2 Spoke Conversion (v1 → v2)

```go
// api/v1/locusttest_conversion.go
package v1

import (
    corev1 "k8s.io/api/core/v1"
    "sigs.k8s.io/controller-runtime/pkg/conversion"
    
    v2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

// ConvertTo converts v1 to v2 (hub)
func (src *LocustTest) ConvertTo(dstRaw conversion.Hub) error {
    dst := dstRaw.(*v2.LocustTest)
    
    // Metadata
    dst.ObjectMeta = src.ObjectMeta
    
    // Image configuration
    dst.Spec.Image = src.Spec.Image
    if src.Spec.ImagePullPolicy != "" {
        dst.Spec.ImagePullPolicy = corev1.PullPolicy(src.Spec.ImagePullPolicy)
    }
    if src.Spec.ImagePullSecrets != nil {
        dst.Spec.ImagePullSecrets = make([]corev1.LocalObjectReference, len(src.Spec.ImagePullSecrets))
        for i, s := range src.Spec.ImagePullSecrets {
            dst.Spec.ImagePullSecrets[i] = corev1.LocalObjectReference{Name: s}
        }
    }
    
    // Master configuration (grouped)
    dst.Spec.Master = v2.MasterSpec{
        Command:   src.Spec.MasterCommandSeed,
        Autostart: ptr.To(true),
        Autoquit:  &v2.AutoquitConfig{Enabled: true, Timeout: 60},
    }
    if src.Spec.Labels != nil {
        dst.Spec.Master.Labels = src.Spec.Labels["master"]
    }
    if src.Spec.Annotations != nil {
        dst.Spec.Master.Annotations = src.Spec.Annotations["master"]
    }
    
    // Worker configuration (grouped)
    dst.Spec.Worker = v2.WorkerSpec{
        Command:  src.Spec.WorkerCommandSeed,
        Replicas: src.Spec.WorkerReplicas,
    }
    if src.Spec.Labels != nil {
        dst.Spec.Worker.Labels = src.Spec.Labels["worker"]
    }
    if src.Spec.Annotations != nil {
        dst.Spec.Worker.Annotations = src.Spec.Annotations["worker"]
    }
    
    // Test files configuration
    if src.Spec.ConfigMap != "" || src.Spec.LibConfigMap != "" {
        dst.Spec.TestFiles = &v2.TestFilesConfig{
            ConfigMapRef:    src.Spec.ConfigMap,
            LibConfigMapRef: src.Spec.LibConfigMap,
        }
    }
    
    // Scheduling configuration
    if src.Spec.Affinity != nil || src.Spec.Tolerations != nil {
        dst.Spec.Scheduling = &v2.SchedulingConfig{}
        if src.Spec.Affinity != nil && src.Spec.Affinity.NodeAffinity != nil {
            dst.Spec.Scheduling.Affinity = convertAffinityToV2(src.Spec.Affinity)
        }
        if src.Spec.Tolerations != nil {
            dst.Spec.Scheduling.Tolerations = make([]corev1.Toleration, len(src.Spec.Tolerations))
            for i, t := range src.Spec.Tolerations {
                dst.Spec.Scheduling.Tolerations[i] = corev1.Toleration{
                    Key:      t.Key,
                    Operator: corev1.TolerationOperator(t.Operator),
                    Value:    t.Value,
                    Effect:   corev1.TaintEffect(t.Effect),
                }
            }
        }
    }
    
    return nil
}

// ConvertFrom converts v2 (hub) to v1
func (dst *LocustTest) ConvertFrom(srcRaw conversion.Hub) error {
    src := srcRaw.(*v2.LocustTest)
    
    // Metadata
    dst.ObjectMeta = src.ObjectMeta
    
    // Image configuration
    dst.Spec.Image = src.Spec.Image
    dst.Spec.ImagePullPolicy = string(src.Spec.ImagePullPolicy)
    if src.Spec.ImagePullSecrets != nil {
        dst.Spec.ImagePullSecrets = make([]string, len(src.Spec.ImagePullSecrets))
        for i, s := range src.Spec.ImagePullSecrets {
            dst.Spec.ImagePullSecrets[i] = s.Name
        }
    }
    
    // Master configuration → flat fields
    dst.Spec.MasterCommandSeed = src.Spec.Master.Command
    
    // Worker configuration → flat fields
    dst.Spec.WorkerCommandSeed = src.Spec.Worker.Command
    dst.Spec.WorkerReplicas = src.Spec.Worker.Replicas
    
    // Labels from grouped structure
    if src.Spec.Master.Labels != nil || src.Spec.Worker.Labels != nil {
        dst.Spec.Labels = map[string]map[string]string{
            "master": src.Spec.Master.Labels,
            "worker": src.Spec.Worker.Labels,
        }
    }
    
    // Annotations from grouped structure
    if src.Spec.Master.Annotations != nil || src.Spec.Worker.Annotations != nil {
        dst.Spec.Annotations = map[string]map[string]string{
            "master": src.Spec.Master.Annotations,
            "worker": src.Spec.Worker.Annotations,
        }
    }
    
    // Test files configuration → flat fields
    if src.Spec.TestFiles != nil {
        dst.Spec.ConfigMap = src.Spec.TestFiles.ConfigMapRef
        dst.Spec.LibConfigMap = src.Spec.TestFiles.LibConfigMapRef
    }
    
    // Scheduling → flat fields (lossy - v1 has simpler affinity structure)
    if src.Spec.Scheduling != nil {
        if src.Spec.Scheduling.Affinity != nil {
            dst.Spec.Affinity = convertAffinityToV1(src.Spec.Scheduling.Affinity)
        }
        if src.Spec.Scheduling.Tolerations != nil {
            dst.Spec.Tolerations = make([]LocustTestToleration, len(src.Spec.Scheduling.Tolerations))
            for i, t := range src.Spec.Scheduling.Tolerations {
                dst.Spec.Tolerations[i] = LocustTestToleration{
                    Key:      t.Key,
                    Operator: string(t.Operator),
                    Value:    t.Value,
                    Effect:   string(t.Effect),
                }
            }
        }
        // nodeSelector is lost in v1 (v2-only)
    }
    
    // v2-only fields are lost in conversion to v1:
    // - master.resources, worker.resources
    // - master.autostart, master.autoquit, master.extraArgs
    // - worker.extraArgs
    // - testFiles.srcMountPath, testFiles.libMountPath
    // - scheduling.nodeSelector
    // - env (configMapRefs, secretRefs, variables, secretMounts)
    // - volumes, volumeMounts
    // - observability
    
    return nil
}

func convertAffinityToV2(src *LocustTestAffinity) *corev1.Affinity {
    if src.NodeAffinity == nil || src.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
        return nil
    }
    
    var terms []corev1.NodeSelectorRequirement
    for key, value := range src.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution {
        terms = append(terms, corev1.NodeSelectorRequirement{
            Key:      key,
            Operator: corev1.NodeSelectorOpIn,
            Values:   []string{value},
        })
    }
    
    return &corev1.Affinity{
        NodeAffinity: &corev1.NodeAffinity{
            RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
                NodeSelectorTerms: []corev1.NodeSelectorTerm{
                    {MatchExpressions: terms},
                },
            },
        },
    }
}
```

### 3.3 Webhook Setup

```go
// api/v1/locusttest_webhook.go
package v1

import (
    ctrl "sigs.k8s.io/controller-runtime"
    logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var locusttestlog = logf.Log.WithName("locusttest-resource")

func (r *LocustTest) SetupWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManagedBy(mgr).
        For(r).
        Complete()
}
```

---

## 4. Validation

### 4.1 Validation Webhook

```go
// api/v2/locusttest_webhook.go
package v2

import (
    "context"
    "fmt"
    
    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/webhook"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-locust-io-v2-locusttest,mutating=false,failurePolicy=fail,sideEffects=None,groups=locust.io,resources=locusttests,verbs=create;update,versions=v2,name=vlocusttest.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &LocustTest{}

func (r *LocustTest) SetupWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManagedBy(mgr).
        For(r).
        WithValidator(r).
        Complete()
}

func (r *LocustTest) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
    lt := obj.(*LocustTest)
    return r.validate(lt)
}

func (r *LocustTest) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
    lt := newObj.(*LocustTest)
    return r.validate(lt)
}

func (r *LocustTest) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
    return nil, nil
}

func (r *LocustTest) validate(lt *LocustTest) (admission.Warnings, error) {
    var warnings admission.Warnings
    
    // Validate volume mount paths don't conflict with operator-managed paths
    reservedPaths := []string{"/lotest/src/", "/opt/locust/lib"}
    for _, vm := range lt.Spec.VolumeMounts {
        for _, reserved := range reservedPaths {
            if vm.MountPath == reserved {
                return warnings, fmt.Errorf("volumeMount path %q conflicts with operator-managed path", vm.MountPath)
            }
        }
    }
    
    // Warn about deprecated v1-style patterns
    if lt.Spec.Labels != nil && len(lt.Spec.Labels.Master) == 0 && len(lt.Spec.Labels.Worker) == 0 {
        warnings = append(warnings, "empty labels configuration has no effect")
    }
    
    // Validate OTel configuration
    if lt.Spec.Observability != nil && lt.Spec.Observability.OpenTelemetry != nil {
        otel := lt.Spec.Observability.OpenTelemetry
        if otel.Enabled && otel.Endpoint == "" {
            return warnings, fmt.Errorf("openTelemetry.endpoint is required when openTelemetry is enabled")
        }
    }
    
    return warnings, nil
}
```

### 4.2 Defaulting Webhook

```go
// api/v2/locusttest_webhook.go (continued)

// +kubebuilder:webhook:path=/mutate-locust-io-v2-locusttest,mutating=true,failurePolicy=fail,sideEffects=None,groups=locust.io,resources=locusttests,verbs=create;update,versions=v2,name=mlocusttest.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &LocustTest{}

func (r *LocustTest) Default(ctx context.Context, obj runtime.Object) error {
    lt := obj.(*LocustTest)
    
    // Default image pull policy
    if lt.Spec.ImagePullPolicy == "" {
        lt.Spec.ImagePullPolicy = corev1.PullIfNotPresent
    }
    
    // Default master config
    if lt.Spec.MasterConfig == nil {
        lt.Spec.MasterConfig = &MasterConfig{}
    }
    if lt.Spec.MasterConfig.Autostart == nil {
        lt.Spec.MasterConfig.Autostart = &AutostartConfig{Enabled: true}
    }
    if lt.Spec.MasterConfig.Autoquit == nil {
        lt.Spec.MasterConfig.Autoquit = &AutoquitConfig{Enabled: true, Timeout: 60}
    }
    
    // Default OTel protocol
    if lt.Spec.Observability != nil && lt.Spec.Observability.OpenTelemetry != nil {
        otel := lt.Spec.Observability.OpenTelemetry
        if otel.Enabled && otel.Protocol == "" {
            otel.Protocol = "grpc"
        }
    }
    
    return nil
}
```

---

## 5. Status Subresource Design

### 5.1 Condition Types

```go
// api/v2/conditions.go
package v2

const (
    // ConditionTypeReady indicates resources are created
    ConditionTypeReady = "Ready"
    
    // ConditionTypeWorkersConnected indicates workers connected to master
    ConditionTypeWorkersConnected = "WorkersConnected"
    
    // ConditionTypeTestCompleted indicates test has finished
    ConditionTypeTestCompleted = "TestCompleted"
)

const (
    // Reasons for Ready condition
    ReasonResourcesCreating  = "ResourcesCreating"
    ReasonResourcesCreated   = "ResourcesCreated"
    ReasonResourcesFailed    = "ResourcesFailed"
    
    // Reasons for WorkersConnected condition
    ReasonWaitingForWorkers = "WaitingForWorkers"
    ReasonAllWorkersConnected = "AllWorkersConnected"
    ReasonWorkersMissing    = "WorkersMissing"
    
    // Reasons for TestCompleted condition
    ReasonTestInProgress = "TestInProgress"
    ReasonTestSucceeded  = "TestSucceeded"
    ReasonTestFailed     = "TestFailed"
)

// Phases
const (
    PhasePending   = "Pending"
    PhaseRunning   = "Running"
    PhaseSucceeded = "Succeeded"
    PhaseFailed    = "Failed"
)
```

### 5.2 Status Update Helpers

```go
// internal/controller/status.go
package controller

import (
    "context"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/api/meta"
    
    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
)

func (r *LocustTestReconciler) updatePhase(ctx context.Context, lt *locustv2.LocustTest, phase string) error {
    lt.Status.Phase = phase
    return r.Status().Update(ctx, lt)
}

func (r *LocustTestReconciler) setCondition(lt *locustv2.LocustTest, condType, reason, message string, status metav1.ConditionStatus) {
    meta.SetStatusCondition(&lt.Status.Conditions, metav1.Condition{
        Type:               condType,
        Status:             status,
        Reason:             reason,
        Message:            message,
        LastTransitionTime: metav1.Now(),
        ObservedGeneration: lt.Generation,
    })
}

func (r *LocustTestReconciler) setReady(lt *locustv2.LocustTest, ready bool, reason, message string) {
    status := metav1.ConditionFalse
    if ready {
        status = metav1.ConditionTrue
    }
    r.setCondition(lt, locustv2.ConditionTypeReady, reason, message, status)
}
```

---

## 6. Printer Columns

### 6.1 Column Configuration

```go
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="Current phase"
// +kubebuilder:printcolumn:name="Workers",type=integer,JSONPath=`.spec.workerReplicas`,description="Requested workers"
// +kubebuilder:printcolumn:name="Connected",type=integer,JSONPath=`.status.connectedWorkers`,description="Connected workers"
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
```

### 6.2 Expected Output

```bash
$ kubectl get locusttests
NAME          PHASE     WORKERS   CONNECTED   AGE
demo.test     Running   10        10          5m
stress.test   Pending   50        0           30s
old.test      Succeeded 5         5           2h

$ kubectl get locusttests -o wide
NAME          PHASE     WORKERS   CONNECTED   IMAGE                      AGE
demo.test     Running   10        10          locustio/locust:2.43.1     5m
```

---

## References

- [Kubebuilder Markers](https://book.kubebuilder.io/reference/markers.html)
- [API Conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- [Conversion Webhooks](https://book.kubebuilder.io/multiversion-tutorial/conversion.html)
