---
title: Your First Load Test
description: Learn how distributed load testing works by building a realistic test from scratch
tags:
  - tutorial
  - distributed testing
  - load testing
  - beginners
---

# Your First Load Test (10 minutes)

Learn how distributed load testing works by building a realistic test from scratch. You'll create an e-commerce scenario that simulates 100 users browsing products and viewing details.

## What you'll learn

- How Locust master and worker pods communicate and distribute load
- How to write realistic test scripts with multiple tasks and weighted behavior
- How to configure test parameters for meaningful results
- How to monitor test progress and interpret statistics

## Prerequisites

- Completed the [Quick Start](../getting_started/index.md) guide
- Basic understanding of HTTP and REST APIs
- Kubernetes cluster with the operator installed

## The scenario

You're testing an e-commerce API before a big sale. You need to verify it can handle 100 simultaneous users over 5 minutes, with users primarily browsing products (75% of traffic) and occasionally viewing product details (25% of traffic).

This simulates realistic user behavior - most users browse, fewer drill into specific items.

## Step 1: Write the test script

Create a test script that simulates realistic shopping behavior:

```python
cat > ecommerce_test.py << 'EOF'
from locust import HttpUser, task, between

class ShopperUser(HttpUser):
    # Wait 1-3 seconds between tasks to simulate realistic user pacing
    wait_time = between(1, 3)

    @task(3)  # This task runs 3x more often (75% of requests)
    def browse_products(self):
        """Browse the product catalog."""
        # The name parameter helps identify requests in statistics
        self.client.get(
            "/anything/products",
            name="GET /products"
        )

    @task(1)  # This task runs 1x as often (25% of requests)
    def view_product_detail(self):
        """View details for a specific product."""
        # Simulate viewing product ID 42
        self.client.get(
            "/anything/products/42",
            name="GET /products/:id"
        )
EOF
```

### What's happening here

- **`HttpUser`**: Base class for simulating HTTP clients. Each instance represents one user.
- **`wait_time = between(1, 3)`**: Adds realistic pauses between requests. Real users don't hammer APIs continuously.
- **`@task(3)` and `@task(1)`**: Task weights control distribution. Weight 3 means "run 3x as often as weight 1", giving us 75%/25% split.
- **`name` parameter**: Groups similar URLs (like `/products/42`, `/products/99`) into one statistic row. Without this, you'd see hundreds of separate rows.

We're using `https://httpbin.org/anything` as a mock API - it accepts any request and returns 200, perfect for learning without deploying a real e-commerce backend.

## Step 2: Deploy the test

First, create the ConfigMap:

```bash
kubectl create configmap ecommerce-test --from-file=ecommerce_test.py
```

Now create the LocustTest resource:

```bash
kubectl apply -f - <<EOF
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: ecommerce-load
spec:
  image: locustio/locust:2.20.0
  testFiles:
    configMapRef: ecommerce-test
  master:
    # 100 users, spawn 10 per second, run for 5 minutes
    command: "--locustfile /lotest/src/ecommerce_test.py --host https://httpbin.org --users 100 --spawn-rate 10 --run-time 5m"
  worker:
    command: "--locustfile /lotest/src/ecommerce_test.py"
    replicas: 5  # Distribute 100 users across 5 workers (~20 users each)
EOF
```

### Understanding the configuration

- **`--users 100`**: Total simulated users across all workers
- **`--spawn-rate 10`**: Add 10 users per second until reaching 100 (takes 10 seconds to ramp up)
- **`--run-time 5m`**: Stop after 5 minutes
- **`replicas: 5`**: Five worker pods distribute the load. Each worker handles ~20 users. This is a good ratio for moderate load.

## Step 3: Monitor the test

Watch the test progress through its lifecycle:

```bash
kubectl get locusttest ecommerce-load -w
```

You'll see output like:

```
NAME             PHASE      WORKERS   CONNECTED   AGE
ecommerce-load   Pending    5          0           2s
ecommerce-load   Running    5          1           8s
ecommerce-load   Running    5          3           12s
ecommerce-load   Running    5          5           18s
ecommerce-load   Succeeded  5          5           5m22s
```

### What the phases mean

- **Pending**: Kubernetes is scheduling the pods
- **Running**: Test is active, workers are connected and generating load
- **Succeeded**: Test completed successfully (ran for full 5 minutes)

You can also check individual pods:

```bash
kubectl get pods -l performance-test-name=ecommerce-load
```

Expected output:

```
NAME                          READY   STATUS    RESTARTS   AGE
ecommerce-load-master-xxxxx   1/1     Running   0          25s
ecommerce-load-worker-xxxxx   1/1     Running   0          25s
ecommerce-load-worker-yyyyy   1/1     Running   0          25s
ecommerce-load-worker-zzzzz   1/1     Running   0          25s
ecommerce-load-worker-aaaaa   1/1     Running   0          25s
ecommerce-load-worker-bbbbb   1/1     Running   0          25s
```

You should see 6 pods total: 1 master + 5 workers.

## Step 4: Access real-time statistics

View the Locust web UI to see live statistics:

```bash
kubectl port-forward job/ecommerce-load-master 8089:8089
```

Open [http://localhost:8089](http://localhost:8089) and look for:

### Key metrics to watch

- **Request statistics table**:
    - **RPS** (requests per second): Should be steady once all users spawn
    - **Response times**: Median, 95th percentile, and 99th percentile
    - **Failures**: Should be 0 for this test (httpbin is reliable)

- **Charts tab**:
    - **Total Requests per Second**: Shows load distribution over time
    - **Response Times**: Visualizes latency trends
    - **Number of Users**: Shows the 10-second ramp-up to 100 users

- **Workers tab**:
    - Verify all 5 workers are connected
    - Check that users are distributed across workers

For a 100-user test with 2-second average wait time and 2 tasks per user cycle, expect roughly **50-75 requests per second** total.

## Step 5: Clean up

After the test completes (or if you stop it early):

```bash
kubectl delete locusttest ecommerce-load
kubectl delete configmap ecommerce-test
```

The operator automatically cleans up all related pods, services, and jobs when you delete the LocustTest resource.

## What you learned

- **Distributed architecture**: Master pod coordinates, worker pods generate load
- **Realistic user simulation**: Task weights and wait times model real behavior
- **Test lifecycle**: Pending → Running → Succeeded phases
- **Resource scaling**: 5 workers for 100 users is a good baseline (~20 users per worker)
- **Monitoring**: LocustTest status shows health, web UI shows performance metrics

## Next steps

- **[CI/CD Integration](ci-cd-integration.md)**: Run tests automatically in your pipeline (15 minutes)
- **[How-To: Configure resources](../how-to-guides/configuration/configure-resources.md)**: Set CPU/memory for stable tests
- **[How-To: Set up Prometheus monitoring](../metrics_and_dashboards.md)**: Export metrics for long-term analysis
- **[API Reference](../api_reference.md)**: Explore all LocustTest configuration options
