---
title: Quick Start
description: Get your first distributed load test running on Kubernetes in 5 minutes
tags:
  - quickstart
  - tutorial
  - getting started
---

# Quick Start (5 minutes)

Get your first distributed load test running on Kubernetes.

## Prerequisites

- Kubernetes cluster (any: Minikube, Kind, GKE, EKS, AKS)
- kubectl configured
- Helm 3 installed

## 1. Install the operator

```bash
# Add the Helm repository
helm repo add locust-k8s-operator https://locust-k8s-operator.github.io/helm-charts/
helm repo update

# Install the operator into a dedicated namespace
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system \
  --create-namespace
```

*Installs the operator into a dedicated namespace. Takes ~30 seconds.*

## 2. Create a test script

```bash
cat > demo_test.py << 'EOF'
from locust import HttpUser, task

class DemoUser(HttpUser):
    @task  # Define a task that users will execute
    def get_homepage(self):
        # Simple test that requests the homepage repeatedly
        self.client.get("/")
EOF
```

*Simple test that requests the homepage repeatedly.*

## 3. Deploy the test as ConfigMap

```bash
# Make your test script available to Kubernetes pods
kubectl create configmap demo-test --from-file=demo_test.py
```

*Makes your test script available to Kubernetes pods.*

## 4. Run the load test

```bash
kubectl apply -f - <<EOF
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: demo
spec:
  image: locustio/locust:2.20.0  # Use a specific version
  testFiles:
    configMapRef: demo-test  # Reference the test script ConfigMap
  master:
    # Run with 10 users, spawning 2 per second, for 1 minute
    command: "--locustfile /lotest/src/demo_test.py --host https://httpbin.org --users 10 --spawn-rate 2 --run-time 1m"
  worker:
    command: "--locustfile /lotest/src/demo_test.py"
    replicas: 2  # Distribute load across 2 worker pods
EOF
```

*Creates a distributed test with 10 simulated users across 2 worker pods.*

## 5. Watch the results

```bash
# Monitor test progress
kubectl get locusttest demo -w
```

You'll see output like this:

```
NAME   PHASE     EXPECTED   CONNECTED   AGE
demo   Pending   2          0           2s
demo   Running   2          2           15s
demo   Succeeded 2          2           75s
```

To access the Locust web UI and view real-time statistics:

```bash
# Forward the master pod's port to your local machine
kubectl port-forward job/demo-master 8089:8089
```

Then open [http://localhost:8089](http://localhost:8089) in your browser to see request statistics, response times, and charts.

## Cleanup

```bash
kubectl delete locusttest demo
kubectl delete configmap demo-test
```

## What's Next?

- **[Your First Load Test](../tutorials/first-load-test.md)** - Build a realistic test with multiple scenarios (10 minutes)
- **[CI/CD Integration](../tutorials/ci-cd-integration.md)** - Automate tests in your pipeline (15 minutes)
- **[API Reference](../api_reference.md)** - Complete LocustTest CR specification
