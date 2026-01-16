# Phase 1: v1 API Types - Checklist

**Estimated Effort:** 1 day  
**Status:** Pending

---

## Pre-Implementation

- [ ] Phase 0 complete (project scaffolds exist)
- [ ] `api/v1/locusttest_types.go` skeleton exists
- [ ] `api/v1/groupversion_info.go` exists with `locust.io` group
- [ ] Java CRD available at `/kube/crd/locust-test-crd.yaml`
- [ ] Sample CR available at `/kube/sample-cr/locust-test-cr.yaml`

---

## Task 1.1: Define LocustTestSpec Struct

- [ ] Add `MasterCommandSeed` field (string, required)
- [ ] Add `WorkerCommandSeed` field (string, required)
- [ ] Add `WorkerReplicas` field (int32, required, min:1, max:500, default:1)
- [ ] Add `Image` field (string, required)
- [ ] Add `ImagePullPolicy` field (string, enum, optional)
- [ ] Add `ImagePullSecrets` field ([]string, optional)
- [ ] Add `ConfigMap` field (string, optional)
- [ ] Add `LibConfigMap` field (string, optional)
- [ ] Add `Labels` field (*PodLabels, optional)
- [ ] Add `Annotations` field (*PodAnnotations, optional)
- [ ] Add `Affinity` field (*LocustTestAffinity, optional)
- [ ] Add `Tolerations` field ([]LocustTestToleration, optional)

---

## Task 1.2: Define Supporting Types

- [ ] Define `PodLabels` struct (Master, Worker maps)
- [ ] Define `PodAnnotations` struct (Master, Worker maps)
- [ ] Define `LocustTestAffinity` struct (NodeAffinity pointer)
- [ ] Define `LocustTestNodeAffinity` struct (RequiredDuringSchedulingIgnoredDuringExecution map)
- [ ] Define `LocustTestToleration` struct (Key, Operator, Value, Effect)

---

## Task 1.3: Add Kubebuilder Validation Markers

Required field markers:
- [ ] `+kubebuilder:validation:Required` on MasterCommandSeed
- [ ] `+kubebuilder:validation:Required` on WorkerCommandSeed
- [ ] `+kubebuilder:validation:Required` on WorkerReplicas
- [ ] `+kubebuilder:validation:Required` on Image

Constraint markers:
- [ ] `+kubebuilder:validation:Minimum=1` on WorkerReplicas
- [ ] `+kubebuilder:validation:Maximum=500` on WorkerReplicas
- [ ] `+kubebuilder:default=1` on WorkerReplicas

Enum markers:
- [ ] `+kubebuilder:validation:Enum=Always;IfNotPresent;Never` on ImagePullPolicy
- [ ] `+kubebuilder:validation:Enum=Exists;Equal` on Toleration.Operator
- [ ] `+kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute` on Toleration.Effect

Optional markers:
- [ ] `+optional` on all optional fields
- [ ] `omitempty` in JSON tags for optional fields

---

## Task 1.4: Add CRD Metadata Markers

- [ ] `+kubebuilder:resource:shortName=lotest` on LocustTest type
- [ ] `+kubebuilder:printcolumn` for master_cmd
- [ ] `+kubebuilder:printcolumn` for worker_replica_count
- [ ] `+kubebuilder:printcolumn` for Image
- [ ] `+kubebuilder:printcolumn` for Age
- [ ] `+kubebuilder:subresource:status` on LocustTest type

---

## Task 1.5: Verify groupversion_info.go

- [ ] Package name is `v1`
- [ ] Group is `locust.io`
- [ ] Version is `v1`
- [ ] SchemeBuilder registers types correctly

---

## Task 1.6: Generate Code

- [ ] Run `make generate`
- [ ] Verify `api/v1/zz_generated.deepcopy.go` exists and updated
- [ ] No errors in generation

---

## Task 1.7: Generate CRD Manifests

- [ ] Run `make manifests`
- [ ] Verify `config/crd/bases/locust.io_locusttests.yaml` generated
- [ ] CRD has correct apiVersion: `apiextensions.k8s.io/v1`
- [ ] CRD has correct group: `locust.io`
- [ ] CRD has correct kind: `LocustTest`

---

## Task 1.8: Validate Against Java CRD

Schema comparison:
- [ ] All v1 spec fields present in generated CRD
- [ ] Field types match Java CRD
- [ ] Required fields list matches: `[masterCommandSeed, workerCommandSeed, workerReplicas, image]`
- [ ] Enum values match for imagePullPolicy
- [ ] Enum values match for toleration operator
- [ ] Enum values match for toleration effect
- [ ] Min/Max constraints on workerReplicas match
- [ ] Default value for workerReplicas is 1

Printer columns:
- [ ] master_cmd column present
- [ ] worker_replica_count column present
- [ ] Image column present
- [ ] Age column present

Resource names:
- [ ] plural: `locusttests`
- [ ] singular: `locusttest`
- [ ] shortName: `lotest`

---

## Task 1.9: Test Sample CR Validation

- [ ] Apply CRD to test cluster: `kubectl apply -f config/crd/bases/locust.io_locusttests.yaml`
- [ ] Sample CR validates: `kubectl apply --dry-run=server -f /kube/sample-cr/locust-test-cr.yaml`
- [ ] No validation errors

---

## Task 1.10: Write Type Tests (Optional)

- [ ] Create `api/v1/locusttest_types_test.go`
- [ ] Test JSON roundtrip marshaling
- [ ] Test JSON field names are camelCase
- [ ] Tests pass: `go test ./api/v1/... -v`

---

## Post-Implementation Verification

- [ ] `make build` succeeds
- [ ] `make test` succeeds
- [ ] `kubectl get lotest` works with short name
- [ ] Printer columns display correctly

---

## Completion

- [ ] Update `phases/NOTES.md` with Phase 1 completion notes
- [ ] Document any deviations from plan
- [ ] Note any issues discovered for future phases
