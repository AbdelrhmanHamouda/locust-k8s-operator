# Phase 11: Volume Mounting - Technical Design

**Version:** 1.0  
**Status:** Draft

---

## Overview

This document details the technical design for implementing arbitrary volume mounting to Locust pods. The implementation extends the existing resource builders to process user-defined volumes from the v2 spec, with support for target filtering (master/worker/both).

---

## 1. Architecture

### 1.1 Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                      LocustTest CR (v2)                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ spec.volumes:                                             │   │
│  │   - name: test-results                                    │   │
│  │     persistentVolumeClaim:                                │   │
│  │       claimName: locust-results-pvc                       │   │
│  │   - name: shared-data                                     │   │
│  │     emptyDir: {}                                          │   │
│  │ spec.volumeMounts:                                        │   │
│  │   - name: test-results                                    │   │
│  │     mountPath: /results                                   │   │
│  │     target: master                                        │   │
│  │   - name: shared-data                                     │   │
│  │     mountPath: /shared                                    │   │
│  │     target: both                                          │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Resource Builders                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ BuildUserVolumes(lt)          → []corev1.Volume          │   │
│  │ BuildUserVolumeMounts(lt, m)  → []corev1.VolumeMount     │   │
│  │   where m = Master or Worker                             │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Master Job Pod Spec                          │
│  volumes:                                                        │
│    - name: <crName>-master          # Operator: ConfigMap        │
│    - name: locust-lib               # Operator: LibConfigMap     │
│    - name: test-results             # User: PVC                  │
│    - name: shared-data              # User: emptyDir             │
│  containers:                                                     │
│    - name: <crName>-master                                       │
│      volumeMounts:                                               │
│        - name: <crName>-master, mountPath: /lotest/src           │
│        - name: locust-lib, mountPath: /opt/locust/lib            │
│        - name: test-results, mountPath: /results  # target=master│
│        - name: shared-data, mountPath: /shared    # target=both  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Worker Job Pod Spec                          │
│  volumes:                                                        │
│    - name: <crName>-worker          # Operator: ConfigMap        │
│    - name: locust-lib               # Operator: LibConfigMap     │
│    - name: shared-data              # User: emptyDir             │
│  containers:                                                     │
│    - name: <crName>-worker                                       │
│      volumeMounts:                                               │
│        - name: <crName>-worker, mountPath: /lotest/src           │
│        - name: locust-lib, mountPath: /opt/locust/lib            │
│        - name: shared-data, mountPath: /shared    # target=both  │
│        # Note: test-results NOT included (target=master only)    │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Component Dependencies

```
api/v2/locusttest_types.go         ← Already defines Volumes, TargetedVolumeMount
         │
         ▼
internal/resources/volumes.go      ← NEW: User volume building functions
         │
         ▼
internal/resources/job.go          ← MODIFY: Merge user volumes with operator volumes
         │
         ▼
api/v2/locusttest_webhook.go       ← MODIFY: Add volume name/path validation
```

---

## 2. Target Filtering Logic

### 2.1 Target Values

| Target | Applies To |
|--------|------------|
| `master` | Master pod only |
| `worker` | Worker pods only |
| `both` (default) | Both master and worker pods |

### 2.2 Filtering Algorithm

```go
func shouldApplyMount(mount TargetedVolumeMount, mode OperationalMode) bool {
    target := mount.Target
    if target == "" {
        target = "both"  // Default
    }
    
    switch target {
    case "both":
        return true
    case "master":
        return mode == Master
    case "worker":
        return mode == Worker
    default:
        return false
    }
}
```

### 2.3 Volume Inclusion Logic

Volumes must be included if **any** of their mounts apply to the current mode:

```go
func shouldIncludeVolume(volumeName string, mounts []TargetedVolumeMount, mode OperationalMode) bool {
    for _, mount := range mounts {
        if mount.Name == volumeName && shouldApplyMount(mount, mode) {
            return true
        }
    }
    return false
}
```

---

## 3. Implementation Details

### 3.1 Volume Builder Functions

**File:** `internal/resources/volumes.go`

```go
package resources

import (
    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
    corev1 "k8s.io/api/core/v1"
)

// Target constants for volume mount filtering.
const (
    TargetMaster = "master"
    TargetWorker = "worker"
    TargetBoth   = "both"
)

// BuildUserVolumes returns user-defined volumes filtered for the given mode.
// Only volumes that have at least one mount targeting this mode are included.
func BuildUserVolumes(lt *locustv2.LocustTest, mode OperationalMode) []corev1.Volume {
    if len(lt.Spec.Volumes) == 0 {
        return nil
    }

    var volumes []corev1.Volume
    for _, vol := range lt.Spec.Volumes {
        if shouldIncludeVolume(vol.Name, lt.Spec.VolumeMounts, mode) {
            volumes = append(volumes, vol)
        }
    }
    return volumes
}

// BuildUserVolumeMounts returns user-defined volume mounts filtered for the given mode.
func BuildUserVolumeMounts(lt *locustv2.LocustTest, mode OperationalMode) []corev1.VolumeMount {
    if len(lt.Spec.VolumeMounts) == 0 {
        return nil
    }

    var mounts []corev1.VolumeMount
    for _, tvm := range lt.Spec.VolumeMounts {
        if shouldApplyMount(tvm, mode) {
            // Convert TargetedVolumeMount to VolumeMount (strip Target field)
            mounts = append(mounts, tvm.VolumeMount)
        }
    }
    return mounts
}

// shouldApplyMount checks if a mount applies to the given operational mode.
func shouldApplyMount(mount locustv2.TargetedVolumeMount, mode OperationalMode) bool {
    target := mount.Target
    if target == "" {
        target = TargetBoth
    }

    switch target {
    case TargetBoth:
        return true
    case TargetMaster:
        return mode == Master
    case TargetWorker:
        return mode == Worker
    default:
        return false
    }
}

// shouldIncludeVolume checks if a volume has any mounts for the given mode.
func shouldIncludeVolume(volumeName string, mounts []locustv2.TargetedVolumeMount, mode OperationalMode) bool {
    for _, mount := range mounts {
        if mount.Name == volumeName && shouldApplyMount(mount, mode) {
            return true
        }
    }
    return false
}
```

### 3.2 Job Builder Updates

**File:** `internal/resources/job.go`

Update `buildVolumes` to include user volumes:

```go
// buildVolumes creates the volumes for ConfigMap, LibConfigMap, Secrets, and user volumes.
func buildVolumes(lt *locustv2.LocustTest, nodeName string, mode OperationalMode) []corev1.Volume {
    var volumes []corev1.Volume

    // [Existing ConfigMap volume logic...]
    
    // [Existing secret volumes from env.secretMounts...]

    // Add user-defined volumes (filtered by target)
    userVolumes := BuildUserVolumes(lt, mode)
    if len(userVolumes) > 0 {
        volumes = append(volumes, userVolumes...)
    }

    return volumes
}
```

Update `buildVolumeMounts` to include user mounts:

```go
// buildVolumeMounts creates volume mounts for ConfigMap, LibConfigMap, Secrets, and user mounts.
func buildVolumeMounts(lt *locustv2.LocustTest, nodeName string, mode OperationalMode) []corev1.VolumeMount {
    var mounts []corev1.VolumeMount

    // [Existing ConfigMap mount logic...]
    
    // [Existing secret mounts from env.secretMounts...]

    // Add user-defined volume mounts (filtered by target)
    userMounts := BuildUserVolumeMounts(lt, mode)
    if len(userMounts) > 0 {
        mounts = append(mounts, userMounts...)
    }

    return mounts
}
```

**Important:** The `buildVolumes` and `buildVolumeMounts` functions need to be updated to accept `mode OperationalMode` as a parameter. This requires updating the call sites in `buildJob` and `buildLocustContainer`.

### 3.3 Signature Changes

Current signatures:
```go
func buildVolumes(lt *locustv2.LocustTest, nodeName string) []corev1.Volume
func buildVolumeMounts(lt *locustv2.LocustTest, nodeName string) []corev1.VolumeMount
```

New signatures:
```go
func buildVolumes(lt *locustv2.LocustTest, nodeName string, mode OperationalMode) []corev1.Volume
func buildVolumeMounts(lt *locustv2.LocustTest, nodeName string, mode OperationalMode) []corev1.VolumeMount
```

Update in `buildJob`:
```go
Volumes: buildVolumes(lt, nodeName, mode),
```

Update in `buildLocustContainer`:
```go
VolumeMounts: buildVolumeMounts(lt, name, mode),
```

And update `buildLocustContainer` signature to include mode:
```go
func buildLocustContainer(lt *locustv2.LocustTest, name string, command []string, 
    ports []corev1.ContainerPort, cfg *config.OperatorConfig, mode OperationalMode) corev1.Container
```

---

## 4. Validation Webhook Updates

**File:** `api/v2/locusttest_webhook.go`

### 4.1 Reserved Volume Names

The operator manages these volume names internally:

| Name Pattern | Purpose |
|--------------|---------|
| `<crName>-master` | ConfigMap for master |
| `<crName>-worker` | ConfigMap for worker |
| `locust-lib` | Library ConfigMap |
| `secret-<name>` | Secret mounts from `env.secretMounts` |

### 4.2 Validation Functions

```go
// Reserved volume name patterns
var reservedVolumeNamePrefix = "secret-"
const libVolumeName = "locust-lib"

// validateVolumes checks for volume name and mount path conflicts.
func (r *LocustTest) validateVolumes() (admission.Warnings, error) {
    // Check volume names
    for _, vol := range r.Spec.Volumes {
        if err := r.validateVolumeName(vol.Name); err != nil {
            return nil, err
        }
    }

    // Check mount paths
    for _, mount := range r.Spec.VolumeMounts {
        if err := r.validateMountPath(mount.MountPath); err != nil {
            return nil, err
        }
    }

    // Validate that all mounts reference defined volumes
    if err := r.validateMountReferences(); err != nil {
        return nil, err
    }

    return nil, nil
}

// validateVolumeName checks if a volume name conflicts with operator-managed names.
func (r *LocustTest) validateVolumeName(name string) error {
    // Check for reserved prefix
    if strings.HasPrefix(name, reservedVolumeNamePrefix) {
        return fmt.Errorf("volume name %q uses reserved prefix %q", name, reservedVolumeNamePrefix)
    }

    // Check for lib volume name
    if name == libVolumeName {
        return fmt.Errorf("volume name %q is reserved by the operator", name)
    }

    // Check for CR-based names
    masterName := r.Name + "-master"
    workerName := r.Name + "-worker"
    if name == masterName || name == workerName {
        return fmt.Errorf("volume name %q conflicts with operator-generated name", name)
    }

    return nil
}

// validateMountPath checks if a mount path conflicts with reserved paths.
func (r *LocustTest) validateMountPath(path string) error {
    reservedPaths := r.getReservedPaths()
    for _, reserved := range reservedPaths {
        if pathConflicts(path, reserved) {
            return fmt.Errorf("volumeMount path %q conflicts with reserved path %q", path, reserved)
        }
    }
    return nil
}

// validateMountReferences ensures all mounts reference defined volumes.
func (r *LocustTest) validateMountReferences() error {
    volumeNames := make(map[string]bool)
    for _, vol := range r.Spec.Volumes {
        volumeNames[vol.Name] = true
    }

    for _, mount := range r.Spec.VolumeMounts {
        if !volumeNames[mount.Name] {
            return fmt.Errorf("volumeMount %q references undefined volume", mount.Name)
        }
    }

    return nil
}
```

### 4.3 Updated Validate Functions

```go
// ValidateCreate implements webhook.Validator
func (r *LocustTest) ValidateCreate() (admission.Warnings, error) {
    var allWarnings admission.Warnings
    
    // Existing secret mount validation
    warnings, err := r.validateSecretMounts()
    if err != nil {
        return warnings, err
    }
    allWarnings = append(allWarnings, warnings...)
    
    // New volume validation
    warnings, err = r.validateVolumes()
    if err != nil {
        return allWarnings, err
    }
    allWarnings = append(allWarnings, warnings...)
    
    return allWarnings, nil
}

// ValidateUpdate implements webhook.Validator
func (r *LocustTest) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
    return r.ValidateCreate()  // Same validation for updates
}
```

---

## 5. Testing Strategy

### 5.1 Unit Tests

**File:** `internal/resources/volumes_test.go`

| Test | Description |
|------|-------------|
| `TestBuildUserVolumes_Empty` | Returns nil for no volumes |
| `TestBuildUserVolumes_AllTargetBoth` | All volumes included when all targets are "both" |
| `TestBuildUserVolumes_MasterOnly` | Only master-targeted volumes for master mode |
| `TestBuildUserVolumes_WorkerOnly` | Only worker-targeted volumes for worker mode |
| `TestBuildUserVolumes_Mixed` | Correct filtering with mixed targets |
| `TestBuildUserVolumeMounts_Empty` | Returns nil for no mounts |
| `TestBuildUserVolumeMounts_MasterMode` | Filters correctly for master |
| `TestBuildUserVolumeMounts_WorkerMode` | Filters correctly for worker |
| `TestBuildUserVolumeMounts_DefaultTarget` | Empty target treated as "both" |
| `TestShouldApplyMount_AllCases` | Test all target/mode combinations |
| `TestShouldIncludeVolume` | Volume included only if mount matches mode |

### 5.2 Webhook Tests

**File:** `api/v2/locusttest_webhook_test.go`

| Test | Description |
|------|-------------|
| `TestValidateVolumes_Empty` | Passes for no volumes |
| `TestValidateVolumes_ValidNames` | Passes for valid volume names |
| `TestValidateVolumes_ReservedPrefix` | Fails for `secret-*` prefix |
| `TestValidateVolumes_LibVolumeName` | Fails for `locust-lib` |
| `TestValidateVolumes_CRNameConflict` | Fails for `<crName>-master/worker` |
| `TestValidateVolumes_PathConflict` | Fails for reserved path conflicts |
| `TestValidateVolumes_UndefinedMount` | Fails for mount referencing undefined volume |

### 5.3 Integration Tests

**File:** `internal/controller/integration_test.go`

| Test | Description |
|------|-------------|
| `TestReconcile_WithUserVolumes` | Volumes added to both master and worker |
| `TestReconcile_WithTargetedMounts_MasterOnly` | Mount only in master |
| `TestReconcile_WithTargetedMounts_WorkerOnly` | Mount only in workers |
| `TestReconcile_WithTargetedMounts_Both` | Mount in both master and workers |
| `TestReconcile_VolumeNotDuplicatedWithSecrets` | User volumes don't conflict with secret volumes |

---

## 6. Reserved Resources Summary

### 6.1 Reserved Paths

| Path | Purpose | Condition |
|------|---------|-----------|
| `/lotest/src` (or custom) | Test files ConfigMap | When `testFiles.configMapRef` set |
| `/opt/locust/lib` (or custom) | Library ConfigMap | When `testFiles.libConfigMapRef` set |

### 6.2 Reserved Volume Names

| Pattern | Purpose |
|---------|---------|
| `<crName>-master` | Master ConfigMap volume |
| `<crName>-worker` | Worker ConfigMap volume |
| `locust-lib` | Library ConfigMap volume |
| `secret-*` | Secret volumes from `env.secretMounts` |

---

## 7. Error Messages

Clear error messages for validation failures:

```
volume name "secret-custom" uses reserved prefix "secret-"
volume name "locust-lib" is reserved by the operator
volume name "my-test-master" conflicts with operator-generated name
volumeMount path "/lotest/src/data" conflicts with reserved path "/lotest/src"
volumeMount "results" references undefined volume
```

---

## 8. Backward Compatibility

- **No breaking changes:** Empty `volumes` and `volumeMounts` produce same behavior as before
- **Existing tests unaffected:** All current integration tests should pass
- **Function signature change:** `buildVolumes` and `buildVolumeMounts` gain `mode` parameter
  - Internal change only, no public API impact

---

## 9. References

- [Kubernetes Volumes](https://kubernetes.io/docs/concepts/storage/volumes/)
- [PersistentVolumeClaim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)
- [Issue #252 Analysis](../../../issue-analysis/P1-High/issue-252-volume-mounting.md)
- [Phase 10 Implementation](../phase-10-env-secret-injection/) - Similar patterns for secret mounts
