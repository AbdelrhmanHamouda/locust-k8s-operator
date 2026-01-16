# Phase 0 Completion Notes

**Completed:** 2026-01-16

---

## Key Decisions Made

### Domain Configuration
- **Domain**: `io` (not `locust.io`)
- **Group**: `locust`
- **Result**: CRD group is `locust.io`, matching the existing Java CRD

This was corrected during implementation. The original plan used `--domain locust.io` which resulted in `locust.locust.io`. The fix was to use `--domain io --group locust`.

---

## Versions Installed

| Tool | Version |
|------|---------|
| Operator SDK | 1.42.0 |
| Go | 1.24.0 |
| controller-runtime | 0.21.0 |
| controller-gen | 0.18.0 |
| k8s.io/api | 0.33.0 |
| k8s.io/apimachinery | 0.33.0 |
| k8s.io/client-go | 0.33.0 |

---

## Notes for Phase 1 (v1 API Types)

1. **API Types Location**: `api/v1/locusttest_types.go` is the skeleton to populate with fields from the Java CRD.

2. **Existing Java CRD Reference**: Use `/kube/crd/locust-test-crd.yaml` as the source of truth for field definitions.

3. **Required Fields from Java CRD**:
   - `masterCommandSeed` (string, required)
   - `workerCommandSeed` (string, required)
   - `workerReplicas` (integer, required, min: 1, max: 500, default: 1)
   - `image` (string, required)
   - `imagePullPolicy` (enum: Always, IfNotPresent, Never)
   - `imagePullSecrets` (array of strings)
   - `configMap` (string)
   - `libConfigMap` (string)
   - `labels` (object with master/worker sub-objects)
   - `annotations` (object with master/worker sub-objects)
   - `affinity` (object with nodeAffinity)
   - `tolerations` (array of toleration objects)

4. **Kubebuilder Markers**: Add validation markers for:
   - Required fields: `// +kubebuilder:validation:Required`
   - Min/Max: `// +kubebuilder:validation:Minimum=1`, `// +kubebuilder:validation:Maximum=500`
   - Enums: `// +kubebuilder:validation:Enum=Always;IfNotPresent;Never`
   - Defaults: `// +kubebuilder:default=1`

5. **Printer Columns**: The Java CRD defines additional printer columns for `kubectl get lotest`:
   - `master_cmd`, `worker_replica_count`, `Image`, `Age`

6. **Short Name**: Add `lotest` as a short name via kubebuilder marker.

7. **Regenerate After Changes**: Run `make manifests` after updating types to regenerate the CRD.
