# Phase 11: Volume Mounting

**Effort:** 0.5 day  
**Priority:** P1 - Must Have  
**Status:** ✅ Complete  
**Dependencies:** Phase 7 (v2 API Types), Phase 10 (Environment & Secret Injection)

---

## Objective

Enable arbitrary volume mounting to Locust master and/or worker pods, allowing users to mount PVCs, emptyDir, hostPath, and other volume types. This addresses Issue #252, enabling advanced use cases like persistent result storage, shared state between workers, and custom certificate mounting.

---

## Requirements Reference

- **REQUIREMENTS.md §5.1.3:** Volume Mounting
- **Issue #252:** Provide Ability to Mount Volumes to Locust Master/Worker Pods
- **issue-analysis/P1-High/issue-252-volume-mounting.md:** Detailed analysis

---

## Scope

### In Scope

- Process `spec.volumes` → add to Pod's `volumes` list
- Process `spec.volumeMounts` → add to Locust container's `volumeMounts`
- Support target filtering (`master`, `worker`, `both`) via `TargetedVolumeMount`
- Validation webhook to prevent conflicts with operator-managed paths and volumes
- Unit and integration tests for volume mounting

### Out of Scope

- Init containers for volume preparation (future enhancement)
- Volume claim templates for StatefulSet-like behavior
- CSI driver-specific configuration
- Mounting volumes to the metrics exporter sidecar

---

## Key Deliverables

| File | Description |
|------|-------------|
| `internal/resources/volumes.go` | User volume building functions with target filtering |
| `internal/resources/volumes_test.go` | Unit tests for volume builders |
| `internal/resources/job.go` | Updated to merge user volumes with operator volumes |
| `api/v2/locusttest_webhook.go` | Extended validation for volume path/name conflicts |
| `api/v2/locusttest_webhook_test.go` | Additional webhook tests for volume validation |

---

## Success Criteria

1. User-defined volumes added to both master and worker pods
2. `TargetedVolumeMount.Target` correctly filters mounts to master, worker, or both
3. Validation rejects volume names that conflict with operator-managed names
4. Validation rejects mount paths that conflict with reserved paths (`/lotest/src/`, `/opt/locust/lib`)
5. All tests pass with ≥80% coverage for new code
6. Backward compatible with existing deployments (empty volumes = no change)

---

## API Types (Already Defined)

The v2 API types are already defined in `api/v2/locusttest_types.go`:

```go
// TargetedVolumeMount extends VolumeMount with target pod selection.
type TargetedVolumeMount struct {
    corev1.VolumeMount `json:",inline"`

    // Target specifies which pods receive this mount.
    // +kubebuilder:validation:Enum=master;worker;both
    // +kubebuilder:default=both
    Target string `json:"target,omitempty"`
}

// In LocustTestSpec:
// Volumes to add to pods.
Volumes []corev1.Volume `json:"volumes,omitempty"`

// VolumeMounts for the locust container with target selection.
VolumeMounts []TargetedVolumeMount `json:"volumeMounts,omitempty"`
```

---

## Example Usage

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: load-test-with-volumes
spec:
  image: locustio/locust:2.20.0
  master:
    command: "locust -f /lotest/src/locustfile.py"
  worker:
    command: "locust -f /lotest/src/locustfile.py"
    replicas: 3
  testFiles:
    configMapRef: locust-scripts
  
  # User-defined volumes (all Kubernetes volume types supported)
  volumes:
    - name: test-results
      persistentVolumeClaim:
        claimName: locust-results-pvc
    - name: shared-data
      emptyDir: {}
    - name: custom-certs
      secret:
        secretName: tls-certificates
  
  # Volume mounts with target selection
  volumeMounts:
    - name: test-results
      mountPath: /results
      target: master  # Only master writes results
    - name: shared-data
      mountPath: /shared
      target: both    # All pods share this volume
    - name: custom-certs
      mountPath: /etc/ssl/custom
      readOnly: true
      target: worker  # Only workers need certs for HTTPS testing
```

---

## Quick Start

```bash
# After implementation, verify with:
make generate
make manifests
make build
make test

# Test volume mounting manually
kubectl apply -f config/samples/locust_v2_locusttest_with_volumes.yaml
kubectl get pods -l performance-test-name=load-test-with-volumes

# Verify volumes on master
kubectl exec -it <master-pod> -- ls /results /shared

# Verify volumes on worker (should not have /results)
kubectl exec -it <worker-pod> -- ls /shared /etc/ssl/custom
```

---

## Related Documents

- [CHECKLIST.md](./CHECKLIST.md) - Detailed implementation checklist
- [DESIGN.md](./DESIGN.md) - Technical design and code patterns
