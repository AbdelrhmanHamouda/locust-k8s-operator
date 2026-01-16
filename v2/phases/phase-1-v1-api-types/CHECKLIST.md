# Phase 1: v1 API Types - Checklist

**Estimated Effort:** 1 day  
**Status:** Complete

---

## Pre-Implementation

- [x] Phase 0 complete (project scaffolds exist)
- [x] `api/v1/locusttest_types.go` skeleton exists
- [x] `api/v1/groupversion_info.go` exists with `locust.io` group
- [x] Java CRD available at `/kube/crd/locust-test-crd.yaml`
- [x] Sample CR available at `/kube/sample-cr/locust-test-cr.yaml`

---

## Task 1.1: Define LocustTestSpec Struct

- [x] Add `MasterCommandSeed` field (string, required)
- [x] Add `WorkerCommandSeed` field (string, required)
- [x] Add `WorkerReplicas` field (int32, required, min:1, max:500, default:1)
- [x] Add `Image` field (string, required)
- [x] Add `ImagePullPolicy` field (string, enum, optional)
- [x] Add `ImagePullSecrets` field ([]string, optional)
- [x] Add `ConfigMap` field (string, optional)
- [x] Add `LibConfigMap` field (string, optional)
- [x] Add `Labels` field (*PodLabels, optional)
- [x] Add `Annotations` field (*PodAnnotations, optional)
- [x] Add `Affinity` field (*LocustTestAffinity, optional)
- [x] Add `Tolerations` field ([]LocustTestToleration, optional)

---

## Task 1.2: Define Supporting Types

- [x] Define `PodLabels` struct (Master, Worker maps)
- [x] Define `PodAnnotations` struct (Master, Worker maps)
- [x] Define `LocustTestAffinity` struct (NodeAffinity pointer)
- [x] Define `LocustTestNodeAffinity` struct (RequiredDuringSchedulingIgnoredDuringExecution map)
- [x] Define `LocustTestToleration` struct (Key, Operator, Value, Effect)

---

## Task 1.3: Add Kubebuilder Validation Markers

Required field markers:
- [x] `+kubebuilder:validation:Required` on MasterCommandSeed
- [x] `+kubebuilder:validation:Required` on WorkerCommandSeed
- [x] `+kubebuilder:validation:Required` on WorkerReplicas
- [x] `+kubebuilder:validation:Required` on Image

Constraint markers:
- [x] `+kubebuilder:validation:Minimum=1` on WorkerReplicas
- [x] `+kubebuilder:validation:Maximum=500` on WorkerReplicas
- [x] `+kubebuilder:default=1` on WorkerReplicas

Enum markers:
- [x] `+kubebuilder:validation:Enum=Always;IfNotPresent;Never` on ImagePullPolicy
- [x] `+kubebuilder:validation:Enum=Exists;Equal` on Toleration.Operator
- [x] `+kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute` on Toleration.Effect

Optional markers:
- [x] `+optional` on all optional fields
- [x] `omitempty` in JSON tags for optional fields

---

## Task 1.4: Add CRD Metadata Markers

- [x] `+kubebuilder:resource:shortName=lotest` on LocustTest type
- [x] `+kubebuilder:printcolumn` for master_cmd
- [x] `+kubebuilder:printcolumn` for worker_replica_count
- [x] `+kubebuilder:printcolumn` for Image
- [x] `+kubebuilder:printcolumn` for Age
- [x] `+kubebuilder:subresource:status` on LocustTest type

---

## Task 1.5: Verify groupversion_info.go

- [x] Package name is `v1`
- [x] Group is `locust.io`
- [x] Version is `v1`
- [x] SchemeBuilder registers types correctly

---

## Task 1.6: Generate Code

- [x] Run `make generate`
- [x] Verify `api/v1/zz_generated.deepcopy.go` exists and updated
- [x] No errors in generation

---

## Task 1.7: Generate CRD Manifests

- [x] Run `make manifests`
- [x] Verify `config/crd/bases/locust.io_locusttests.yaml` generated
- [x] CRD has correct apiVersion: `apiextensions.k8s.io/v1`
- [x] CRD has correct group: `locust.io`
- [x] CRD has correct kind: `LocustTest`

---

## Task 1.8: Validate Against Java CRD

Schema comparison:
- [x] All v1 spec fields present in generated CRD
- [x] Field types match Java CRD
- [x] Required fields list matches: `[masterCommandSeed, workerCommandSeed, workerReplicas, image]`
- [x] Enum values match for imagePullPolicy
- [x] Enum values match for toleration operator
- [x] Enum values match for toleration effect
- [x] Min/Max constraints on workerReplicas match
- [x] Default value for workerReplicas is 1

Printer columns:
- [x] master_cmd column present
- [x] worker_replica_count column present
- [x] Image column present
- [x] Age column present

Resource names:
- [x] plural: `locusttests`
- [x] singular: `locusttest`
- [x] shortName: `lotest`

---

## Task 1.9: Test Sample CR Validation

- [ ] Apply CRD to test cluster: `kubectl apply -f config/crd/bases/locust.io_locusttests.yaml`
- [ ] Sample CR validates: `kubectl apply --dry-run=server -f /kube/sample-cr/locust-test-cr.yaml`
- [ ] No validation errors

---

## Task 1.10: Write Type Tests (Optional)

- [x] Create `api/v1/locusttest_types_test.go`
- [x] Test JSON roundtrip marshaling
- [x] Test JSON field names are camelCase
- [x] Tests pass: `go test ./api/v1/... -v`

---

## Post-Implementation Verification

- [x] `make build` succeeds
- [x] `make test` succeeds
- [ ] `kubectl get lotest` works with short name
- [ ] Printer columns display correctly

---

## Completion

- [ ] Update `phases/NOTES.md` with Phase 1 completion notes
- [ ] Document any deviations from plan
- [ ] Note any issues discovered for future phases
