# Phase 10: Environment & Secret Injection - Technical Design

**Version:** 1.0  
**Status:** Draft

---

## Overview

This document details the technical design for implementing environment variable and secret injection into Locust pods. The implementation extends the existing resource builders to process the v2 `EnvConfig` specification.

---

## 1. Architecture

### 1.1 Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                      LocustTest CR (v2)                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ spec.env:                                                 │   │
│  │   configMapRefs:                                          │   │
│  │     - name: app-config                                    │   │
│  │       prefix: "APP_"                                      │   │
│  │   secretRefs:                                             │   │
│  │     - name: api-creds                                     │   │
│  │   variables:                                              │   │
│  │     - name: TARGET_HOST                                   │   │
│  │       value: "https://example.com"                        │   │
│  │   secretMounts:                                           │   │
│  │     - name: tls-certs                                     │   │
│  │       mountPath: /etc/certs                               │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Resource Builders                             │
│  ┌─────────────────────┐    ┌─────────────────────┐            │
│  │ BuildEnvFrom()      │    │ BuildEnvVars()      │            │
│  │ → ConfigMapEnvSource│    │ → []corev1.EnvVar   │            │
│  │ → SecretEnvSource   │    │                     │            │
│  └─────────────────────┘    └─────────────────────┘            │
│  ┌─────────────────────┐    ┌─────────────────────┐            │
│  │ BuildSecretVolumes()│    │ BuildSecretMounts() │            │
│  │ → []corev1.Volume   │    │ → []corev1.VolMount │            │
│  └─────────────────────┘    └─────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         Pod Spec                                 │
│  containers:                                                     │
│    - name: locust-master                                         │
│      env:                                                        │
│        - name: TARGET_HOST                                       │
│          value: "https://example.com"                            │
│        - name: KAFKA_BOOTSTRAP_SERVERS  # Existing kafka vars    │
│          value: "..."                                            │
│      envFrom:                                                    │
│        - configMapRef:                                           │
│            name: app-config                                      │
│          prefix: "APP_"                                          │
│        - secretRef:                                              │
│            name: api-creds                                       │
│      volumeMounts:                                               │
│        - name: secret-tls-certs                                  │
│          mountPath: /etc/certs                                   │
│          readOnly: true                                          │
│  volumes:                                                        │
│    - name: secret-tls-certs                                      │
│      secret:                                                     │
│        secretName: tls-certs                                     │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Component Dependencies

```
api/v2/locusttest_types.go     ← Already defines EnvConfig, SecretMount, etc.
         │
         ▼
internal/resources/env.go      ← NEW: Environment building functions
         │
         ▼
internal/resources/job.go      ← MODIFY: Integrate env builders
         │
         ▼
api/v2/locusttest_webhook.go   ← NEW: Validation for path conflicts
```

---

## 2. API Types (Already Defined)

The v2 API types are already defined in `api/v2/locusttest_types.go`:

```go
// EnvConfig defines environment variable injection configuration.
type EnvConfig struct {
    ConfigMapRefs []ConfigMapEnvSource `json:"configMapRefs,omitempty"`
    SecretRefs    []SecretEnvSource    `json:"secretRefs,omitempty"`
    Variables     []corev1.EnvVar      `json:"variables,omitempty"`
    SecretMounts  []SecretMount        `json:"secretMounts,omitempty"`
}

type ConfigMapEnvSource struct {
    Name   string `json:"name"`
    Prefix string `json:"prefix,omitempty"`
}

type SecretEnvSource struct {
    Name   string `json:"name"`
    Prefix string `json:"prefix,omitempty"`
}

type SecretMount struct {
    Name      string `json:"name"`
    MountPath string `json:"mountPath"`
    ReadOnly  bool   `json:"readOnly,omitempty"`
}
```

---

## 3. Implementation Details

### 3.1 Environment Builder Functions

**File:** `internal/resources/env.go`

```go
package resources

import (
    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
    "github.com/AbdelrhmanHamouda/locust-k8s-operator/internal/config"
    corev1 "k8s.io/api/core/v1"
)

// BuildEnvFrom creates EnvFromSource entries from ConfigMap and Secret refs.
// Returns envFrom slice for container spec.
func BuildEnvFrom(lt *locustv2.LocustTest) []corev1.EnvFromSource {
    if lt.Spec.Env == nil {
        return nil
    }

    var envFrom []corev1.EnvFromSource

    // Process ConfigMapRefs
    for _, cmRef := range lt.Spec.Env.ConfigMapRefs {
        envFrom = append(envFrom, corev1.EnvFromSource{
            Prefix: cmRef.Prefix,
            ConfigMapRef: &corev1.ConfigMapEnvSource{
                LocalObjectReference: corev1.LocalObjectReference{
                    Name: cmRef.Name,
                },
            },
        })
    }

    // Process SecretRefs
    for _, secretRef := range lt.Spec.Env.SecretRefs {
        envFrom = append(envFrom, corev1.EnvFromSource{
            Prefix: secretRef.Prefix,
            SecretRef: &corev1.SecretEnvSource{
                LocalObjectReference: corev1.LocalObjectReference{
                    Name: secretRef.Name,
                },
            },
        })
    }

    return envFrom
}

// BuildUserEnvVars creates EnvVar entries from the variables list.
// These are appended to the existing Kafka env vars.
func BuildUserEnvVars(lt *locustv2.LocustTest) []corev1.EnvVar {
    if lt.Spec.Env == nil || len(lt.Spec.Env.Variables) == 0 {
        return nil
    }
    
    // Return a copy to avoid mutating the original
    result := make([]corev1.EnvVar, len(lt.Spec.Env.Variables))
    copy(result, lt.Spec.Env.Variables)
    return result
}

// BuildEnvVars combines Kafka env vars with user-defined env vars.
func BuildEnvVars(lt *locustv2.LocustTest, cfg *config.OperatorConfig) []corev1.EnvVar {
    // Start with Kafka env vars (existing behavior)
    envVars := buildKafkaEnvVars(cfg)
    
    // Append user-defined variables
    userVars := BuildUserEnvVars(lt)
    if len(userVars) > 0 {
        envVars = append(envVars, userVars...)
    }
    
    return envVars
}

// BuildSecretVolumes creates Volume entries for secret mounts.
func BuildSecretVolumes(lt *locustv2.LocustTest) []corev1.Volume {
    if lt.Spec.Env == nil || len(lt.Spec.Env.SecretMounts) == 0 {
        return nil
    }

    var volumes []corev1.Volume
    for _, sm := range lt.Spec.Env.SecretMounts {
        volumes = append(volumes, corev1.Volume{
            Name: secretVolumeName(sm.Name),
            VolumeSource: corev1.VolumeSource{
                Secret: &corev1.SecretVolumeSource{
                    SecretName: sm.Name,
                },
            },
        })
    }
    return volumes
}

// BuildSecretVolumeMounts creates VolumeMount entries for secret mounts.
func BuildSecretVolumeMounts(lt *locustv2.LocustTest) []corev1.VolumeMount {
    if lt.Spec.Env == nil || len(lt.Spec.Env.SecretMounts) == 0 {
        return nil
    }

    var mounts []corev1.VolumeMount
    for _, sm := range lt.Spec.Env.SecretMounts {
        mounts = append(mounts, corev1.VolumeMount{
            Name:      secretVolumeName(sm.Name),
            MountPath: sm.MountPath,
            ReadOnly:  sm.ReadOnly,
        })
    }
    return mounts
}

// secretVolumeName generates a unique volume name for a secret mount.
func secretVolumeName(secretName string) string {
    return "secret-" + secretName
}
```

### 3.2 Job Builder Updates

**File:** `internal/resources/job.go`

Update `buildLocustContainer` to include envFrom and additional env vars:

```go
// buildLocustContainer creates the main Locust container.
func buildLocustContainer(lt *locustv2.LocustTest, name string, command []string, ports []corev1.ContainerPort, cfg *config.OperatorConfig) corev1.Container {
    container := corev1.Container{
        Name:            name,
        Image:           lt.Spec.Image,
        ImagePullPolicy: corev1.PullPolicy(lt.Spec.ImagePullPolicy),
        Args:            command,
        Ports:           ports,
        Resources:       buildResourceRequirements(cfg, false),
        Env:             BuildEnvVars(lt, cfg),           // CHANGED: Was buildKafkaEnvVars(cfg)
        EnvFrom:         BuildEnvFrom(lt),                // NEW: Add envFrom
        VolumeMounts:    buildVolumeMounts(lt, name),     // Will be updated to include secret mounts
    }

    // Default to IfNotPresent if not specified
    if container.ImagePullPolicy == "" {
        container.ImagePullPolicy = corev1.PullIfNotPresent
    }

    return container
}
```

Update `buildVolumes` to include secret volumes:

```go
// buildVolumes creates the volumes for ConfigMap, LibConfigMap, and Secrets.
func buildVolumes(lt *locustv2.LocustTest, nodeName string) []corev1.Volume {
    var volumes []corev1.Volume

    // Existing ConfigMap volumes...
    // (keep existing code)

    // Add secret volumes from env.secretMounts
    secretVolumes := BuildSecretVolumes(lt)
    if len(secretVolumes) > 0 {
        volumes = append(volumes, secretVolumes...)
    }

    return volumes
}
```

Update `buildVolumeMounts` to include secret mounts:

```go
// buildVolumeMounts creates the volume mounts for ConfigMap, LibConfigMap, and Secrets.
func buildVolumeMounts(lt *locustv2.LocustTest, nodeName string) []corev1.VolumeMount {
    var mounts []corev1.VolumeMount

    // Existing ConfigMap mounts...
    // (keep existing code)

    // Add secret mounts from env.secretMounts
    secretMounts := BuildSecretVolumeMounts(lt)
    if len(secretMounts) > 0 {
        mounts = append(mounts, secretMounts...)
    }

    return mounts
}
```

### 3.3 Validation Webhook

**File:** `api/v2/locusttest_webhook.go`

```go
package v2

import (
    "fmt"
    "strings"

    "k8s.io/apimachinery/pkg/runtime"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/webhook"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Reserved paths that cannot be used for secret mounts
var reservedPaths = []string{
    "/lotest/src",
    "/opt/locust/lib",
}

func (r *LocustTest) SetupWebhookWithManager(mgr ctrl.Manager) error {
    return ctrl.NewWebhookManagedBy(mgr).
        For(r).
        Complete()
}

// +kubebuilder:webhook:path=/validate-locust-io-v2-locusttest,mutating=false,failurePolicy=fail,sideEffects=None,groups=locust.io,resources=locusttests,verbs=create;update,versions=v2,name=vlocusttest.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &LocustTest{}

// ValidateCreate implements webhook.Validator
func (r *LocustTest) ValidateCreate() (admission.Warnings, error) {
    return r.validateSecretMounts()
}

// ValidateUpdate implements webhook.Validator
func (r *LocustTest) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
    return r.validateSecretMounts()
}

// ValidateDelete implements webhook.Validator
func (r *LocustTest) ValidateDelete() (admission.Warnings, error) {
    return nil, nil
}

// validateSecretMounts checks that secret mount paths don't conflict with reserved paths.
func (r *LocustTest) validateSecretMounts() (admission.Warnings, error) {
    if r.Spec.Env == nil || len(r.Spec.Env.SecretMounts) == 0 {
        return nil, nil
    }

    for _, sm := range r.Spec.Env.SecretMounts {
        for _, reserved := range reservedPaths {
            if pathConflicts(sm.MountPath, reserved) {
                return nil, fmt.Errorf(
                    "secretMount path %q conflicts with reserved path %q; "+
                        "operator uses this path for test files", 
                    sm.MountPath, reserved)
            }
        }
    }

    return nil, nil
}

// pathConflicts checks if two paths would conflict.
// Conflict occurs if one path is a prefix of the other.
func pathConflicts(path1, path2 string) bool {
    // Normalize paths
    p1 := strings.TrimSuffix(path1, "/")
    p2 := strings.TrimSuffix(path2, "/")

    // Check if either is a prefix of the other
    return p1 == p2 || 
           strings.HasPrefix(p1, p2+"/") || 
           strings.HasPrefix(p2, p1+"/")
}
```

---

## 4. Reserved Paths

The following paths are reserved by the operator and cannot be used for secret mounts:

| Path | Purpose |
|------|---------|
| `/lotest/src/` | Default mount path for locustfile ConfigMap |
| `/opt/locust/lib` | Default mount path for library ConfigMap |

Users can customize these via `testFiles.srcMountPath` and `testFiles.libMountPath`, but the defaults must be protected.

**Note:** If users customize mount paths, the validation should use the customized paths. This requires accessing both `testFiles` and `env` during validation.

### 4.1 Dynamic Reserved Path Calculation

```go
func (r *LocustTest) getReservedPaths() []string {
    paths := []string{}
    
    srcPath := "/lotest/src"
    libPath := "/opt/locust/lib"
    
    if r.Spec.TestFiles != nil {
        if r.Spec.TestFiles.SrcMountPath != "" {
            srcPath = r.Spec.TestFiles.SrcMountPath
        }
        if r.Spec.TestFiles.LibMountPath != "" {
            libPath = r.Spec.TestFiles.LibMountPath
        }
    }
    
    // Only add paths that are in use
    if r.Spec.TestFiles != nil && r.Spec.TestFiles.ConfigMapRef != "" {
        paths = append(paths, srcPath)
    }
    if r.Spec.TestFiles != nil && r.Spec.TestFiles.LibConfigMapRef != "" {
        paths = append(paths, libPath)
    }
    
    return paths
}
```

---

## 5. Testing Strategy

### 5.1 Unit Tests

**File:** `internal/resources/env_test.go`

| Test | Description |
|------|-------------|
| `TestBuildEnvFrom_Empty` | Returns nil for nil EnvConfig |
| `TestBuildEnvFrom_ConfigMapRefs` | Correct ConfigMapEnvSource created |
| `TestBuildEnvFrom_SecretRefs` | Correct SecretEnvSource created |
| `TestBuildEnvFrom_WithPrefix` | Prefix correctly applied |
| `TestBuildEnvFrom_Multiple` | Multiple sources combined |
| `TestBuildUserEnvVars_Empty` | Returns nil for empty variables |
| `TestBuildUserEnvVars_Values` | Direct values correctly set |
| `TestBuildUserEnvVars_ValueFrom` | ValueFrom references preserved |
| `TestBuildEnvVars_CombinesKafkaAndUser` | Both sources combined |
| `TestBuildSecretVolumes_Empty` | Returns nil for no mounts |
| `TestBuildSecretVolumes_Single` | Single volume created correctly |
| `TestBuildSecretVolumes_Multiple` | Multiple volumes created |
| `TestBuildSecretVolumeMounts_Empty` | Returns nil for no mounts |
| `TestBuildSecretVolumeMounts_ReadOnly` | ReadOnly flag honored |
| `TestSecretVolumeName` | Name generation correct |

### 5.2 Webhook Tests

**File:** `api/v2/locusttest_webhook_test.go`

| Test | Description |
|------|-------------|
| `TestValidateSecretMounts_NoEnv` | Passes when env is nil |
| `TestValidateSecretMounts_NoMounts` | Passes when secretMounts empty |
| `TestValidateSecretMounts_Valid` | Passes for valid paths |
| `TestValidateSecretMounts_ConflictSrc` | Fails for /lotest/src conflict |
| `TestValidateSecretMounts_ConflictLib` | Fails for /opt/locust/lib conflict |
| `TestValidateSecretMounts_Subpath` | Fails for subpath conflicts |
| `TestPathConflicts` | Unit test for pathConflicts function |

### 5.3 Integration Tests

**File:** `internal/controller/integration_test.go` (update)

| Test | Description |
|------|-------------|
| `TestReconcile_WithEnvConfigMapRef` | ConfigMapRef appears in envFrom |
| `TestReconcile_WithEnvSecretRef` | SecretRef appears in envFrom |
| `TestReconcile_WithEnvVariables` | Variables appear in env |
| `TestReconcile_WithSecretMount` | Secret volume and mount created |
| `TestReconcile_EnvCombinedWithKafka` | User vars don't override Kafka vars |

---

## 6. Migration Considerations

### 6.1 Backward Compatibility

- **No breaking changes:** Empty `env` spec produces same behavior as before
- **Kafka env vars preserved:** User env vars are appended, not replaced
- **Existing tests unaffected:** All current integration tests should pass

### 6.2 Conversion Webhook Update

The v1↔v2 conversion already handles `env` field loss gracefully:
- v2→v1 conversion loses `env` field (documented in Phase 8)
- v1→v2 conversion sets `env` to nil

No changes needed to conversion webhook.

---

## 7. Error Handling

### 7.1 Invalid References

If a ConfigMap or Secret referenced in `env` doesn't exist:
- Pod will fail to start with `CreateContainerConfigError`
- Standard Kubernetes behavior, no special handling needed
- Users see clear error in pod events

### 7.2 Path Validation Errors

If a secretMount path conflicts with reserved paths:
- Admission webhook rejects the CR
- Clear error message indicating the conflict
- User must choose a different path

---

## 8. References

- [Kubernetes EnvFromSource](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envfromsource-v1-core)
- [Kubernetes EnvVar](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.28/#envvar-v1-core)
- [Secret as Volume](https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod)
- [Kubebuilder Webhooks](https://book.kubebuilder.io/cronjob-tutorial/webhook-implementation.html)
