---
title: Validate with Kind Cluster
description: Complete guide to validating the Locust K8s Operator deployment on a local Kind cluster
tags:
  - validation
  - testing
  - kind
  - local-development
---

# Validate with Kind Cluster

This guide provides quick commands to validate the Locust K8s operator deployment on a local Kind cluster. It combines the official [Helm deployment guide](../helm_deploy.md) with local development best practices to help you verify the operator works correctly.

## Overview

You'll learn how to:

1. Create a local Kind cluster
2. Deploy the operator via Helm (from the published Helm repository)
3. Run a simple distributed load test
4. Verify the operator works correctly

This validation process is useful for:

- **New users**: Quickly try the operator before production deployment
- **Contributors**: Validate changes during local development
- **CI/CD**: Automated testing in ephemeral environments

## Prerequisites

Ensure you have installed:

- **Docker**: Running Docker daemon
- **kubectl**: Kubernetes CLI
- **Helm 3.x**: Package manager for Kubernetes
- **Kind**: Kubernetes in Docker

??? tip "Install Prerequisites"
    ```bash
    # macOS (using Homebrew)
    brew install kubectl helm kind

    # Linux
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind

    # Verify installations
    kubectl version --client
    helm version
    kind version
    ```

## Quick Start

For experienced users, here's the complete validation flow:

```bash
# 1. Create cluster
kind create cluster --name locust-test

# 2. Install operator
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/
helm repo update
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system --create-namespace

# 3. Create test
kubectl create configmap demo-test --from-literal=demo_test.py='
from locust import HttpUser, task
class DemoUser(HttpUser):
    @task
    def get_homepage(self):
        self.client.get("/")
'

# 4. Run test
kubectl apply -f - <<EOF
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: demo
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: demo-test
  master:
    command: "--locustfile /lotest/src/demo_test.py --host https://httpbin.org --users 10 --spawn-rate 2 --run-time 1m"
  worker:
    command: "--locustfile /lotest/src/demo_test.py"
    replicas: 2
EOF

# 5. Watch progress
kubectl get locusttest demo -w
```

## Step-by-Step Guide

### Step 1: Create Kind Cluster

Create a dedicated Kind cluster for testing:

```bash
kind create cluster --name locust-test
```

**Validate the cluster is ready:**

```bash
# Check cluster info
kubectl cluster-info --context kind-locust-test

# Verify nodes are ready
kubectl get nodes
```

Expected output:
```
NAME                        STATUS   ROLES           AGE   VERSION
locust-test-control-plane   Ready    control-plane   30s   v1.27.3
```

### Step 2: Install Operator via Helm

Add the Helm repository and install the operator:

```bash
# Add the Locust K8s Operator Helm repository
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/
helm repo update

# Install the operator into locust-system namespace
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system \
  --create-namespace
```

**Validate the operator is running:**

```bash
# Check pods status (should see operator pods running)
kubectl get pods -n locust-system

# View operator logs
kubectl logs -f -n locust-system -l app.kubernetes.io/name=locust-k8s-operator
```

Expected output:
```
NAME                                      READY   STATUS    RESTARTS   AGE
locust-operator-controller-manager-xxx    2/2     Running   0          30s
```

!!! success "Verify CRD Registration"
    ```bash
    kubectl get crd locusttests.locust.io
    ```
    You should see the `LocustTest` custom resource definition registered.

### Step 3: Create Test Script

Create a simple Locust test script as a ConfigMap:

```bash
# Create the test script
cat > demo_test.py << 'EOF'
from locust import HttpUser, task

class DemoUser(HttpUser):
    @task
    def get_homepage(self):
        # Simple test that requests the homepage
        self.client.get("/")
EOF

# Deploy the test script as a ConfigMap
kubectl create configmap demo-test --from-file=demo_test.py
```

**Validate ConfigMap creation:**

```bash
kubectl get configmap demo-test
kubectl describe configmap demo-test
```

??? tip "Alternative: Inline ConfigMap"
    You can also create the ConfigMap inline without a separate file:
    ```bash
    kubectl create configmap demo-test --from-literal=demo_test.py='
    from locust import HttpUser, task
    class DemoUser(HttpUser):
        @task
        def get_homepage(self):
            self.client.get("/")
    '
    ```

### Step 4: Deploy LocustTest CR

Create a `LocustTest` custom resource to run the load test:

```bash
kubectl apply -f - <<EOF
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: demo
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: demo-test
  master:
    command: "--locustfile /lotest/src/demo_test.py --host https://httpbin.org --users 10 --spawn-rate 2 --run-time 1m"
  worker:
    command: "--locustfile /lotest/src/demo_test.py"
    replicas: 2
EOF
```

This creates a distributed load test with:

- **Target**: https://httpbin.org (public test API)
- **Users**: 10 concurrent users
- **Spawn rate**: 2 users per second
- **Duration**: 1 minute
- **Workers**: 2 worker replicas

### Step 5: Watch Test Execution

Monitor the test as it progresses through its phases:

```bash
# Watch the LocustTest status
kubectl get locusttest demo -w
```

**Expected progression:**

```
NAME   PHASE       WORKERS   CONNECTED   AGE
demo   Pending     2         0           2s
demo   Running     2         2           15s
demo   Succeeded   2         2           75s
```

**View detailed status:**

```bash
# View all resources created by the operator
kubectl get locusttests,jobs,pods

# Check master job logs
kubectl logs job/demo-master

# Check worker deployment logs
kubectl logs -l app=locust,role=worker --prefix=true
```

??? info "Understanding Test Phases"
    The `LocustTest` CR transitions through these phases:

    - **Pending**: Operator is creating resources (Job, Deployment, Service)
    - **Running**: Test is actively executing, workers are connected
    - **Succeeded**: Test completed successfully (master job finished)
    - **Failed**: Test encountered errors (check logs for details)

### Step 6: Access Locust Web UI (Optional)

While the test is running, you can access the Locust web UI:

```bash
# Port-forward to access the web UI
kubectl port-forward job/demo-master 8089:8089
```

Then open http://localhost:8089 in your browser to see:

- Request statistics (RPS, response times, failures)
- Response time charts
- Real-time test progress
- Worker status

!!! note "Web UI Availability"
    The web UI is only available while the master pod is running. After the test completes, the Job pod may terminate (depending on your TTL settings).

### Step 7: Cleanup

Remove test resources and optionally the cluster:

```bash
# Delete the test (also removes Job and Deployment)
kubectl delete locusttest demo

# Delete the ConfigMap
kubectl delete configmap demo-test

# Uninstall the operator (optional)
helm uninstall locust-operator -n locust-system

# Delete the Kind cluster when done
kind delete cluster --name locust-test
```

## Verification Checklist

Use this checklist to ensure everything is working correctly:

### ✅ Operator Installation

- [ ] Operator pods are running in `locust-system` namespace
- [ ] Operator logs show successful startup (no errors)
- [ ] CRD `locusttests.locust.io` is registered

```bash
kubectl get crd locusttests.locust.io
```

### ✅ Test Execution

- [ ] LocustTest CR transitions from `Pending` → `Running` → `Succeeded`
- [ ] Master job is created and completes successfully
- [ ] Worker deployment scales to 2 replicas
- [ ] Workers connect to master (CONNECTED count matches WORKERS count)

### ✅ Validation Commands

```bash
# Check LocustTest status
kubectl get locusttest demo -o jsonpath='{.status.phase}'

# Verify workers connected
kubectl get locusttest demo -o jsonpath='{.status.connectedWorkers}'

# Check master job succeeded
kubectl get job demo-master -o jsonpath='{.status.succeeded}'
```

## Troubleshooting

### Operator Pods Not Starting

**Symptoms**: Operator pods stuck in `Pending`, `CrashLoopBackOff`, or `ImagePullBackOff`

```bash
# Check pod details
kubectl describe pods -n locust-system

# View previous logs if pod restarted
kubectl logs -n locust-system -l app.kubernetes.io/name=locust-k8s-operator --previous
```

**Common causes**:

- Insufficient cluster resources (CPU/memory)
- Image pull issues (check Docker Hub rate limits)
- RBAC permissions (check ServiceAccount and Roles)

### LocustTest Stays in Pending

**Symptoms**: LocustTest CR remains in `Pending` phase, no resources created

```bash
# Check LocustTest details and events
kubectl describe locusttest demo

# View recent cluster events
kubectl get events --sort-by='.lastTimestamp'
```

**Common causes**:

- Invalid test configuration (check `spec` fields)
- Missing ConfigMap reference
- Operator not reconciling (check operator logs)

### Workers Don't Connect

**Symptoms**: Workers remain disconnected, `CONNECTED` count is 0

```bash
# Check worker pod logs
kubectl logs -l app=locust,role=worker

# Verify service exists
kubectl get svc demo-master

# Check service endpoints
kubectl get endpoints demo-master
```

**Common causes**:

- Service not created or misconfigured
- Network policy blocking traffic
- Workers using wrong master address
- Locust version mismatch between master and workers

### Test Fails or Times Out

**Symptoms**: Test transitions to `Failed` phase or hangs indefinitely

```bash
# Check master logs for errors
kubectl logs job/demo-master

# Check worker logs for errors
kubectl logs -l app=locust,role=worker --tail=50
```

**Common causes**:

- Target host unreachable (DNS, firewall, internet access)
- Locust script errors (Python syntax, import errors)
- Insufficient resources (CPU/memory limits too low)
- Timeout too short for test workload

## Advanced Testing

### Test with Local Build

To test local changes to the operator code:

```bash
# Build and load local image
make docker-build IMG=locust-k8s-operator:dev
kind load docker-image locust-k8s-operator:dev --name locust-test

# Install with local image
helm install locust-operator ./charts/locust-k8s-operator \
  --namespace locust-system \
  --create-namespace \
  --set image.repository=locust-k8s-operator \
  --set image.tag=dev \
  --set image.pullPolicy=IfNotPresent
```

!!! tip "Use Local Chart"
    When testing Helm chart changes, install from your local chart directory instead of the published repository.

### Test Production Features

Try the production-ready example with resource limits, affinity, and tolerations:

```bash
kubectl apply -f config/samples/locusttest_v2_production.yaml
```

Or test OpenTelemetry integration:

```bash
kubectl apply -f config/samples/locust_v2_locusttest_with_otel.yaml
```

??? example "Custom Load Profile"
    Create a more realistic load test with custom stages:
    ```yaml
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: staged-test
    spec:
      image: locustio/locust:2.20.0
      testFiles:
        configMapRef: demo-test
      master:
        command: |
          --locustfile /lotest/src/demo_test.py
          --host https://httpbin.org
          --headless
          --users 100
          --spawn-rate 10
          --run-time 5m
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
      worker:
        command: "--locustfile /lotest/src/demo_test.py"
        replicas: 5
        resources:
          requests:
            cpu: 500m
            memory: 512Mi
          limits:
            cpu: 1000m
            memory: 1Gi
    ```

### Automated Validation Script

Create a validation script for CI/CD pipelines:

```bash
#!/bin/bash
set -e

# Validation script for Kind cluster testing
CLUSTER_NAME="locust-ci-test"
NAMESPACE="locust-system"
TEST_NAME="ci-validation"

echo "Creating Kind cluster..."
kind create cluster --name $CLUSTER_NAME

echo "Installing operator..."
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/
helm repo update
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace $NAMESPACE \
  --create-namespace \
  --wait \
  --timeout 5m

echo "Waiting for operator to be ready..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=locust-k8s-operator \
  -n $NAMESPACE --timeout=5m

echo "Creating test resources..."
kubectl create configmap demo-test --from-literal=demo_test.py='
from locust import HttpUser, task
class DemoUser(HttpUser):
    @task
    def get_homepage(self):
        self.client.get("/")
'

kubectl apply -f - <<EOF
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: $TEST_NAME
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: demo-test
  master:
    command: "--locustfile /lotest/src/demo_test.py --host https://httpbin.org --users 10 --spawn-rate 2 --run-time 1m"
  worker:
    command: "--locustfile /lotest/src/demo_test.py"
    replicas: 2
EOF

echo "Waiting for test to complete..."
kubectl wait --for=jsonpath='{.status.phase}'=Succeeded locusttest/$TEST_NAME \
  --timeout=5m || {
  echo "Test failed or timed out!"
  kubectl describe locusttest/$TEST_NAME
  kubectl logs job/${TEST_NAME}-master
  exit 1
}

echo "Validation successful!"
kubectl get locusttest/$TEST_NAME -o yaml

echo "Cleaning up..."
kind delete cluster --name $CLUSTER_NAME
```

## Success Criteria

The validation is successful if:

1. ✅ Operator installs without errors
2. ✅ LocustTest CR is accepted and reconciled
3. ✅ Master job and worker deployment are created
4. ✅ Workers successfully connect to master
5. ✅ Load test runs and completes (transitions to `Succeeded` phase)
6. ✅ No errors in operator logs
7. ✅ Resources are cleaned up properly when CR is deleted

## Next Steps

After validating the operator with Kind:

- **Production Deployment**: Follow the [Production Deployment tutorial](../tutorials/production-deployment.md)
- **Configure Resources**: Set up [resource limits and requests](configuration/configure-resources.md)
- **Set up Monitoring**: Configure [OpenTelemetry](observability/configure-opentelemetry.md) or [Prometheus](../metrics_and_dashboards.md)
- **CI/CD Integration**: Integrate with your [CI/CD pipeline](../tutorials/ci-cd-integration.md)

## Related Documentation

- [Helm Deployment Guide](../helm_deploy.md) — Official Helm installation instructions
- [Local Development Guide](../local-development.md) — Development workflow for contributors
- [Integration Testing](../integration-testing.md) — Automated testing with envtest
- [First Load Test Tutorial](../tutorials/first-load-test.md) — Complete beginner walkthrough
