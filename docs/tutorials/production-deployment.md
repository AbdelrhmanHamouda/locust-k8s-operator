---
title: Production Deployment (20 minutes)
description: Configure production-grade load tests with resources, scheduling, and observability
tags:
  - tutorial
  - production
  - opentelemetry
  - resources
  - affinity
  - scaling
---

# Production Deployment (20 minutes)

Configure your load tests for production-grade reliability, observability, and performance.

## What you'll learn

- How to size resources for master and worker pods
- How to isolate load test workloads on dedicated nodes
- How to export metrics to OpenTelemetry
- How to scale workers for high user counts
- How to monitor test health via status conditions

## Prerequisites

- Completed the [CI/CD Integration](ci-cd-integration.md) tutorial
- Understanding of Kubernetes resource management
- OpenTelemetry Collector deployed (optional, for Step 4)

## The scenario

You're running a 1000-user production load test against your staging environment. The test needs:

- Dedicated nodes (no interference with production workloads)
- Resource limits (predictable cluster usage)
- Metrics export to your observability stack
- 30-minute sustained test with monitoring

## Step 1: Enhanced test script

Building on the `ecommerce_test.py` from Tutorial 1, here's a production-ready version with realistic complexity:

```python
# production_test.py
from locust import HttpUser, task, between
import logging
import uuid

class ProductionShopperUser(HttpUser):
    """Production-grade e-commerce load test with authentication and realistic behavior."""

    wait_time = between(1, 3)  # Realistic user pacing: 1-3 seconds between actions

    def on_start(self):
        """Called once when user starts - simulate authentication."""
        self.user_id = str(uuid.uuid4())
        # Authenticate with the API
        response = self.client.post("/api/v1/auth/login", json={
            "username": "loadtest@example.com",
            "password": "test-password"
        }, name="Login")

        if response.status_code == 200:
            # Store auth token for subsequent requests
            self.auth_token = response.json().get("token")
            self.client.headers.update({"Authorization": f"Bearer {self.auth_token}"})
            logging.info("User authenticated successfully")
        else:
            logging.error(f"Authentication failed: {response.status_code}")

    @task(5)  # 50% of requests - most common action
    def browse_products(self):
        """Browse product catalog - main landing page."""
        self.client.get("/api/v1/products", name="Browse Products")

    @task(2)  # 20% of requests
    def search_products(self):
        """Search for specific products."""
        search_terms = ["laptop", "phone", "tablet", "monitor"]
        term = search_terms[hash(str(self.user_id)) % len(search_terms)]
        self.client.get(f"/api/v1/products?q={term}", name="Search Products")

    @task(2)  # 20% of requests
    def view_product_detail(self):
        """View specific product details."""
        # Simulate viewing different products
        product_id = 1000 + (hash(str(self.user_id)) % 100)
        self.client.get(f"/api/v1/products/{product_id}", name="View Product Detail")

    @task(1)  # 10% of requests
    def add_to_cart(self):
        """Add product to shopping cart."""
        product_id = 1000 + (hash(str(self.user_id)) % 100)
        self.client.post("/api/v1/cart", json={
            "product_id": product_id,
            "quantity": 1
        }, name="Add to Cart")

    def on_stop(self):
        """Called when user stops - cleanup if needed."""
        logging.info("User session ended")
```

**Key enhancements:**

- **`on_start` hook** — Authenticates once per user (realistic behavior)
- **Weighted tasks** — 5:2:2:1 ratio matches real user behavior (browsing > searching > viewing > purchasing)
- **Dynamic data** — Product IDs vary per user (prevents cache hits)
- **Wait time** — `between(1, 3)` simulates realistic user pacing
- **Logging** — Helps debug test issues in production

## Step 2: Size resources appropriately

### Why resource sizing matters

Load tests need consistent performance to generate reliable results. Without resource limits:

- **Worker pods compete** for CPU with other workloads → inconsistent request rates
- **Memory exhaustion** can crash pods mid-test → incomplete results
- **Cluster instability** affects production services

### Sizing guidelines

**Master pod** (coordinator):
- **Memory**: 512Mi request, 1Gi limit — handles test coordination and statistics
- **CPU**: 500m request, 1000m limit — moderate processing needs

**Worker pods** (load generators):
- **Memory**: 256Mi request, 512Mi limit — per-worker estimate: ~50 users
- **CPU**: 250m request, **no limit** — maximizes request generation throughput
- **Replica count**: Total users ÷ 50 = worker count (e.g., 1000 users = 20 workers)

### Resource configuration example

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: resource-sized-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: production-test
  master:
    command: |
      --locustfile /lotest/src/production_test.py
      --host https://api.staging.example.com
      --users 1000
      --spawn-rate 50
      --run-time 30m
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "1Gi"
        cpu: "1000m"
  worker:
    command: "--locustfile /lotest/src/production_test.py"
    replicas: 20  # 1000 users ÷ 50 users/worker = 20 workers
    resources:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "512Mi"
        # CPU limit intentionally omitted for maximum performance
```

**Why omit CPU limit on workers?** CPU limits can throttle request generation, reducing test accuracy. Workers with only CPU requests get maximum available CPU while still being schedulable.

## Step 3: Isolate on dedicated nodes

### Why dedicated nodes prevent interference

Running load tests on shared nodes can:

- **Throttle production workloads** — high CPU usage from workers affects critical services
- **Skew test results** — resource contention from other pods creates inconsistent performance
- **Violate policies** — some clusters prohibit non-production workloads on production nodes

### Label nodes for load testing

```bash
# Identify nodes for load testing (e.g., separate node pool)
kubectl get nodes

# Label dedicated node(s)
kubectl label nodes worker-node-1 workload-type=load-testing
kubectl label nodes worker-node-2 workload-type=load-testing
kubectl label nodes worker-node-3 workload-type=load-testing
```

### Configure node affinity

Add node affinity to ensure pods only run on labeled nodes:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: isolated-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: production-test
  master:
    command: |
      --locustfile /lotest/src/production_test.py
      --host https://api.staging.example.com
      --users 1000
      --spawn-rate 50
      --run-time 30m
  worker:
    command: "--locustfile /lotest/src/production_test.py"
    replicas: 20
  scheduling:
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: workload-type
                  operator: In
                  values:
                    - load-testing  # Only schedule on labeled nodes
```

### Add tolerations for tainted nodes

If your dedicated nodes have taints (prevents accidental scheduling), add tolerations:

```yaml
spec:
  scheduling:
    affinity:
      # ... (node affinity from above)
    tolerations:
      - key: "workload-type"
        operator: "Equal"
        value: "load-testing"
        effect: "NoSchedule"
```

**Verification:**

```bash
# Check where master and worker pods are scheduled
kubectl get pods -l performance-test-name=isolated-test -o wide

# You should see NODE column showing only your labeled nodes
```

## Step 4: Enable OpenTelemetry

### Why native OpenTelemetry beats sidecars

The v2 operator includes native OpenTelemetry support, eliminating the need for sidecar containers:

- **Lower overhead** — no extra containers per pod
- **Simpler configuration** — environment variables injected automatically
- **Better performance** — direct export from Locust to collector

### Configure OpenTelemetry export

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: otel-enabled-test
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: production-test
  master:
    command: |
      --locustfile /lotest/src/production_test.py
      --host https://api.staging.example.com
      --users 1000
      --spawn-rate 50
      --run-time 30m
  worker:
    command: "--locustfile /lotest/src/production_test.py"
    replicas: 20
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"  # Your OTel Collector endpoint
      protocol: "grpc"  # or "http/protobuf"
      # TLS is the default; set insecure: true only for development without TLS
      extraEnvVars:
        OTEL_SERVICE_NAME: "production-load-test"
        OTEL_RESOURCE_ATTRIBUTES: "environment=staging,team=platform,test.type=load"
```

**Configuration details:**

- **`endpoint`** — OpenTelemetry Collector gRPC endpoint (format: `host:port`)
- **`protocol`** — `grpc` (default) or `http/protobuf`
- **`insecure`** — TLS is the default; set `true` only for development without TLS
- **`extraEnvVars`** — Custom attributes for trace/metric filtering

### Verify OpenTelemetry injection

```bash
# Check environment variables in master pod
kubectl get pod -l performance-test-pod-name=otel-enabled-test-master \
  -o yaml | grep OTEL_

# Expected output:
# OTEL_TRACES_EXPORTER: otlp
# OTEL_METRICS_EXPORTER: otlp
# OTEL_EXPORTER_OTLP_ENDPOINT: otel-collector.monitoring:4317
# OTEL_EXPORTER_OTLP_PROTOCOL: grpc
# OTEL_SERVICE_NAME: production-load-test
```

## Step 5: Deploy the complete production test

Combining all previous steps, here's the full production-ready LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: production-load-test
  namespace: load-testing
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: production-test
  master:
    command: |
      --locustfile /lotest/src/production_test.py
      --host https://api.staging.example.com
      --users 1000
      --spawn-rate 50
      --run-time 30m
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "1Gi"
        cpu: "1000m"
  worker:
    command: "--locustfile /lotest/src/production_test.py"
    replicas: 20  # 1000 users ÷ 50 users per worker
    resources:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "512Mi"
        # CPU limit omitted for maximum worker performance
  scheduling:
    affinity:
      nodeAffinity:
        requiredDuringSchedulingIgnoredDuringExecution:
          nodeSelectorTerms:
            - matchExpressions:
                - key: workload-type
                  operator: In
                  values:
                    - load-testing
    tolerations:
      - key: "workload-type"
        operator: "Equal"
        value: "load-testing"
        effect: "NoSchedule"
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: "grpc"
      extraEnvVars:
        OTEL_SERVICE_NAME: "production-load-test"
        OTEL_RESOURCE_ATTRIBUTES: "environment=staging,team=platform"
```

### Deploy the test

```bash
# Create ConfigMap from enhanced test script
kubectl create configmap production-test \
  --from-file=production_test.py \
  --namespace load-testing

# Apply the LocustTest CR
kubectl apply -f production-load-test.yaml
```

## Step 6: Monitor and verify

### Watch test progression

```bash
# Monitor test status (watch mode)
kubectl get locusttest production-load-test -n load-testing -w

# Expected progression:
# NAME                     PHASE       WORKERS   CONNECTED   AGE
# production-load-test     Pending     20         0           5s
# production-load-test     Running     20         20          45s
# production-load-test     Succeeded   20         20          31m
```

### Check status conditions

```bash
# View detailed status conditions
kubectl get locusttest production-load-test -n load-testing \
  -o jsonpath='{.status.conditions[*]}' | jq

# Expected conditions:
# {
#   "type": "PodsHealthy",
#   "status": "True",
#   "reason": "PodsHealthy",
#   "message": "All pods are healthy"
# }
```

### Verify worker health

```bash
# Check all worker pods are running
kubectl get pods -l performance-test-pod-name=production-load-test-worker \
  -n load-testing

# Expected: 20 pods in Running state
```

### Verify OpenTelemetry traces

If OpenTelemetry is configured, check your observability backend:

**Prometheus (metrics):**
```promql
# Query Locust request metrics (illustrative)
locust_requests_total{service_name="production-load-test"}

# Query response time metrics (illustrative)
locust_request_duration_seconds{service_name="production-load-test"}
```

!!! note "Metric names are illustrative"
    Actual metric names depend on your OpenTelemetry/Prometheus setup and exporter configuration. Check your OTel Collector and Prometheus documentation for the exact names available in your environment.

**Jaeger/Tempo (traces):**

Filter by `service.name=production-load-test` to see:

- Individual request spans
- Request duration distribution
- Error traces

### Access real-time Locust UI

```bash
# Port-forward to master pod
kubectl port-forward -n load-testing \
  job/production-load-test-master 8089:8089

# Open http://localhost:8089 in browser
```

The UI shows:

- Live request statistics (RPS, response times, failures)
- Charts showing performance trends over time
- Worker connection status
- Test phase and remaining duration

## What you learned

✓ How to size master and worker resources for production workloads
✓ How to isolate load tests on dedicated nodes using affinity and tolerations
✓ How to export traces and metrics to OpenTelemetry collectors
✓ How to scale worker replicas based on simulated user count
✓ How to monitor test health through status conditions
✓ How to deploy complete production-grade load tests

## Next steps

- [Configure resources](../how-to-guides/configuration/configure-resources.md) — Deep dive into resource optimization
- [Configure OpenTelemetry](../how-to-guides/observability/configure-opentelemetry.md) — Advanced observability setup
- [Use node affinity](../how-to-guides/scaling/use-node-affinity.md) — More scheduling strategies
- [API Reference](../api_reference.md) — Explore all configuration options
- [Security best practices](../security.md) — Secure your load testing infrastructure
