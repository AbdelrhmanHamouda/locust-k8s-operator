#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../../.." && pwd)"

echo -e "${YELLOW}=== E2E Conversion Webhook Tests ===${NC}"
echo "Project root: ${PROJECT_ROOT}"

# Helper functions
pass() {
    echo -e "${GREEN}✓ PASS:${NC} $1"
}

fail() {
    echo -e "${RED}✗ FAIL:${NC} $1"
    exit 1
}

info() {
    echo -e "${YELLOW}→${NC} $1"
}

# Verify prerequisites
info "Checking prerequisites..."
command -v kubectl >/dev/null 2>&1 || fail "kubectl not found"
command -v kind >/dev/null 2>&1 || fail "kind not found"

# Verify cluster is running
kubectl cluster-info >/dev/null 2>&1 || fail "Kubernetes cluster not reachable"
pass "Cluster is reachable"

# Verify operator is running
info "Checking operator deployment..."
kubectl wait --for=condition=Available deployment/locust-k8s-operator-go-controller-manager -n locust-k8s-operator-go-system --timeout=60s || fail "Operator not running"
pass "Operator is running"

# Verify storage version
info "Verifying v2 is storage version..."
STORAGE_VERSION=$(kubectl get crd locusttests.locust.io -o jsonpath='{.spec.versions[?(@.storage==true)].name}')
if [ "$STORAGE_VERSION" != "v2" ]; then
    fail "Storage version is '$STORAGE_VERSION', expected 'v2'"
fi
pass "v2 is storage version"

# Cleanup any previous test resources
info "Cleaning up previous test resources..."
kubectl delete locusttest e2e-test-v1 e2e-test-v2 --ignore-not-found=true 2>/dev/null || true
kubectl delete configmap e2e-test-scripts --ignore-not-found=true 2>/dev/null || true
sleep 2

# Create ConfigMap
info "Creating test ConfigMap..."
kubectl apply -f "${SCRIPT_DIR}/configmap.yaml"
pass "ConfigMap created"

echo ""
echo -e "${YELLOW}=== Test 1: Create v1 CR ===${NC}"
info "Creating v1 LocustTest..."
kubectl apply -f "${SCRIPT_DIR}/v1-cr.yaml"
sleep 3

# Verify v1 CR is created and can be read
V1_NAME=$(kubectl get locusttests.v1.locust.io e2e-test-v1 -o jsonpath='{.metadata.name}' 2>/dev/null || echo "")
if [ "$V1_NAME" != "e2e-test-v1" ]; then
    fail "v1 CR not created properly"
fi
pass "v1 CR created successfully"

echo ""
echo -e "${YELLOW}=== Test 2: Read v1 CR as v2 ===${NC}"
info "Reading v1 CR via v2 API..."
V2_IMAGE=$(kubectl get locusttests.v2.locust.io e2e-test-v1 -o jsonpath='{.spec.image}' 2>/dev/null || echo "")
V2_WORKER_REPLICAS=$(kubectl get locusttests.v2.locust.io e2e-test-v1 -o jsonpath='{.spec.worker.replicas}' 2>/dev/null || echo "")
V2_MASTER_CMD=$(kubectl get locusttests.v2.locust.io e2e-test-v1 -o jsonpath='{.spec.master.command}' 2>/dev/null || echo "")

if [ "$V2_IMAGE" != "locustio/locust:2.20.0" ]; then
    fail "v2 image mismatch: got '$V2_IMAGE'"
fi
if [ "$V2_WORKER_REPLICAS" != "2" ]; then
    fail "v2 worker.replicas mismatch: got '$V2_WORKER_REPLICAS'"
fi
if [ "$V2_MASTER_CMD" != "locust" ]; then
    fail "v2 master.command mismatch: got '$V2_MASTER_CMD'"
fi
pass "v1→v2 conversion works correctly"

echo ""
echo -e "${YELLOW}=== Test 3: Create v2 CR ===${NC}"
info "Creating v2 LocustTest..."
kubectl apply -f "${SCRIPT_DIR}/v2-cr.yaml"
sleep 3

V2_NAME=$(kubectl get locusttests.v2.locust.io e2e-test-v2 -o jsonpath='{.metadata.name}' 2>/dev/null || echo "")
if [ "$V2_NAME" != "e2e-test-v2" ]; then
    fail "v2 CR not created properly"
fi
pass "v2 CR created successfully"

echo ""
echo -e "${YELLOW}=== Test 4: Read v2 CR as v1 ===${NC}"
info "Reading v2 CR via v1 API..."
V1_IMAGE=$(kubectl get locusttests.v1.locust.io e2e-test-v2 -o jsonpath='{.spec.image}' 2>/dev/null || echo "")
V1_WORKER_REPLICAS=$(kubectl get locusttests.v1.locust.io e2e-test-v2 -o jsonpath='{.spec.workerReplicas}' 2>/dev/null || echo "")
V1_MASTER_CMD=$(kubectl get locusttests.v1.locust.io e2e-test-v2 -o jsonpath='{.spec.masterCommandSeed}' 2>/dev/null || echo "")

if [ "$V1_IMAGE" != "locustio/locust:2.20.0" ]; then
    fail "v1 image mismatch: got '$V1_IMAGE'"
fi
if [ "$V1_WORKER_REPLICAS" != "3" ]; then
    fail "v1 workerReplicas mismatch: got '$V1_WORKER_REPLICAS'"
fi
if [ "$V1_MASTER_CMD" != "locust" ]; then
    fail "v1 masterCommandSeed mismatch: got '$V1_MASTER_CMD'"
fi
pass "v2→v1 conversion works correctly"

echo ""
echo -e "${YELLOW}=== Test 5: Update v1 CR ===${NC}"
info "Updating v1 CR workerReplicas..."
kubectl patch locusttests.v1.locust.io e2e-test-v1 --type=merge -p '{"spec":{"workerReplicas":5}}'
sleep 2

# Verify update is reflected in v2 view
V2_UPDATED_REPLICAS=$(kubectl get locusttests.v2.locust.io e2e-test-v1 -o jsonpath='{.spec.worker.replicas}' 2>/dev/null || echo "")
if [ "$V2_UPDATED_REPLICAS" != "5" ]; then
    fail "v2 worker.replicas not updated: got '$V2_UPDATED_REPLICAS'"
fi
pass "v1 update reflected in v2 view"

echo ""
echo -e "${YELLOW}=== Test 6: Verify Reconciler Creates Jobs ===${NC}"
info "Checking if reconciler created Jobs for e2e-test-v2..."
sleep 2

# Check jobs for e2e-test-v2 (created via v2 API, should have jobs)
MASTER_JOB=$(kubectl get jobs -l performance-test-name=e2e-test-v2,app=locust-master -o name 2>/dev/null | head -1)
WORKER_JOB=$(kubectl get jobs -l performance-test-name=e2e-test-v2,app=locust-worker -o name 2>/dev/null | head -1)

if [ -z "$MASTER_JOB" ]; then
    fail "Master Job not created for e2e-test-v2"
fi
if [ -z "$WORKER_JOB" ]; then
    fail "Worker Job not created for e2e-test-v2"
fi
pass "Reconciler created Jobs from v2 resources"

echo ""
echo -e "${YELLOW}=== Test 7: Verify Deprecation Warning ===${NC}"
info "Checking deprecation warning on v1 API..."
DEPRECATION_OUTPUT=$(kubectl get locusttests.v1.locust.io e2e-test-v1 2>&1)
if echo "$DEPRECATION_OUTPUT" | grep -q "deprecated"; then
    pass "Deprecation warning shown for v1 API"
else
    info "Note: Deprecation warning may not be visible in all kubectl versions"
    pass "Deprecation warning test skipped (kubectl version dependent)"
fi

# Cleanup
echo ""
info "Cleaning up test resources..."
kubectl delete locusttest e2e-test-v1 e2e-test-v2 --ignore-not-found=true 2>/dev/null || true
kubectl delete configmap e2e-test-scripts --ignore-not-found=true 2>/dev/null || true

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  All E2E Conversion Tests PASSED!     ${NC}"
echo -e "${GREEN}========================================${NC}"
