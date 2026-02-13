# Documentation Action Plan - What to Do in docs/

**Goal:** Improve documentation from 8.5/10 to 9.5/10 with OSS-focused, high-impact additions

**Timeline:** Priority 1 (pre-launch), Priority 2 (post-launch), Priority 3 (ongoing)

---

### 2. Add Status Field Documentation to `docs/api_reference.md` (2-3 hours)

**Why:** Users need to understand test lifecycle for CI/CD integration

**Where:** Add new section "Understanding Status Fields" after main API reference

**Content:**

## Understanding Status Fields

### Lifecycle Phases

```yaml
status:
  phase: Running  # Pending | Running | Succeeded | Failed
```

**Phase Meanings:**
- **Pending:** Resources are being created (master pod starting)
- **Running:** Test is actively running (workers connected, load test executing)
- **Succeeded:** Test completed successfully (autoQuit triggered or manual stop)
- **Failed:** Test failed (pod crashes, script errors, validation failures)

**Phase Transitions:**
- Pending → Running: When master pod reaches Ready state
- Running → Succeeded: When test completes normally (autoQuit or manual stop)
- Running → Failed: When critical error occurs (master crash, too many worker failures)

### Conditions

```yaml
status:
  conditions:
  - type: Ready
    status: "True"
    reason: TestRunning
    message: "Master pod is ready and workers are connecting"
```

**Condition Types:**
- **Ready:** Test infrastructure is healthy and operational
- **WorkersConnected:** Workers have successfully connected to master
- **TestCompleted:** Test has finished execution

### Worker Connection Status

```yaml
status:
  workerConnectionCount: 10  # Number of workers currently connected
```

**What It Means:**
- Should match `spec.worker.replicas` when test is healthy
- Lower count indicates workers crashing/not connecting
- Zero count means no workers connected (check network/service)

### Checking Status from CLI

**Quick Status:**
```bash
kubectl get locusttest -o wide
```

**Detailed Status:**
```bash
kubectl get locusttest <name> -o jsonpath='{.status}' | jq
```

**Watch Status Changes:**
```bash
kubectl get locusttest <name> -o yaml -w
```

### Status in CI/CD Pipelines

**Wait for Success:**
```bash
kubectl wait --for=condition=Ready locusttest/<name> --timeout=300s
kubectl wait --for=jsonpath='{.status.phase}'=Succeeded locusttest/<name> --timeout=3600s
```

**Check Exit Code:**
```bash
phase=$(kubectl get locusttest <name> -o jsonpath='{.status.phase}')
if [ "$phase" = "Succeeded" ]; then
  echo "Test passed"
  exit 0
else
  echo "Test failed with phase: $phase"
  exit 1
fi
```


---

### 3. Create `docs/faq.md` (2-3 hours)

**Why:** Reduces repetitive support questions

**Content:**
```markdown
# Frequently Asked Questions (FAQ)

## General Questions

### Why can't I update a running test?

LocustTest resources are **immutable by design**. Once a test starts, you cannot modify its configuration.

**Reason:** Ensures test reproducibility and prevents mid-test configuration drift that could invalidate results.

**Solution:** Delete the old test and create a new one with updated configuration:
```bash
kubectl delete locusttest my-test
kubectl apply -f updated-test.yaml
```

---

### How do I scale workers during a test?

You **cannot** scale workers mid-test due to immutability (see above).

**Workaround:**
1. Design tests with sufficient workers from the start
2. Use multiple test runs with different worker counts to find optimal scale

---

### What's the maximum worker count?

**500 workers** (enforced by validation webhook)

**Why This Limit:**
- Prevents resource exhaustion
- Ensures reasonable load on Kubernetes API server
- Based on real-world production testing limits

**If You Need More:**
- Run multiple LocustTest resources in parallel (each with up to 500 workers)
- Use node affinity to spread across nodes
- Consider multiple clusters for extreme scale

---

### Can I run multiple tests in parallel?

**Yes!** Each LocustTest resource is fully isolated.

**Example:**
```yaml
# test-1.yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: api-test
spec:
  master:
    # ...
---
# test-2.yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: database-test
spec:
  master:
    # ...
```

**Isolation Guarantees:**
- Separate master/worker pods
- Independent resource quotas
- No shared state

---

### How do I check if my test succeeded?

Check the `status.phase` field:

```bash
kubectl get locusttest my-test -o jsonpath='{.status.phase}'
# Output: Succeeded | Failed | Running | Pending
```

**In CI/CD:**
```bash
kubectl wait --for=jsonpath='{.status.phase}'=Succeeded locusttest/my-test --timeout=3600s
```

---

### Why is my test stuck in "Pending"?

**Common Causes:**
1. **Image pull errors:** Check `kubectl describe pod` for ImagePullBackOff
2. **Resource quotas:** Namespace doesn't have enough CPU/memory
3. **Node constraints:** affinity/tolerations prevent scheduling
4. **PVC issues:** PersistentVolumeClaim not bound

**Debug:**
```bash
kubectl describe locusttest my-test
kubectl describe pod <master-pod-name>
```

See [Troubleshooting](troubleshooting.md) for detailed solutions.

---

### Do I need to manually clean up resources after tests?

**No!** The operator automatically cleans up all resources when you delete the LocustTest:

```bash
kubectl delete locusttest my-test
```

**What Gets Cleaned Up:**
- Master pod
- Worker pods
- Services
- ConfigMaps (if auto-generated)

**What Persists:**
- Secrets (you created them)
- PVCs (you created them)
- ConfigMaps (you created them)

---

## Migration Questions

### Should I upgrade from v1 to v2?

**Yes, for new deployments.** v2 offers:
- 60x faster startup (<1s vs ~60s)
- 4x lower memory (64MB vs 256MB)
- Native OpenTelemetry, pod health monitoring, HA support

**For existing v1 users:**
- v1 CRs still work (conversion webhook provides compatibility)
- Migrate at your own pace
- See [Migration Guide](migration.md)

---

### Will my v1 LocustTest resources work with v2 operator?

**Yes!** The conversion webhook automatically converts v1 ↔ v2 API versions.

**Important:** Some v1 fields are deprecated. See migration guide for details.

---

## Configuration Questions

### What's the difference between `masterResourceSpec` and `workerResourceSpec`?

**masterResourceSpec:** Resource requests/limits for the master pod (runs web UI, coordinates test)

**workerResourceSpec:** Resource requests/limits for each worker pod (generates load)

**Why Separate:**
- Master needs more memory (UI, metrics aggregation)
- Workers need more CPU (load generation)
- Optimize each independently

**Example:**
```yaml
spec:
  masterResourceSpec:
    requests:
      cpu: "500m"
      memory: "512Mi"
    limits:
      cpu: "1"
      memory: "1Gi"
  workerResourceSpec:
    requests:
      cpu: "250m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"
```

---

### How do I inject secrets into my test?

Use `envSecrets` or mount as volumes:

**As Environment Variables:**
```yaml
spec:
  envSecrets:
  - secretName: api-credentials
    prefix: "API_"  # Optional: keys become API_USERNAME, API_PASSWORD
```

**As Volume Mounts:**
```yaml
spec:
  volumes:
  - name: certs
    secret:
      secretName: tls-certs
    volumeMounts:
    - name: certs
      mountPath: /etc/ssl/certs
```

See [Advanced Topics](advanced_topics.md#secret-injection) for details.

---

## Observability Questions

### How do I monitor my tests with Prometheus?

The operator exposes metrics on port 8080:

```yaml
# ServiceMonitor for Prometheus Operator
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: locust-operator
spec:
  selector:
    matchLabels:
      app: locust-operator
  endpoints:
  - port: metrics
```

See [Metrics & Dashboards](metrics_and_dashboards.md) for Grafana dashboard.

---

### How do I use OpenTelemetry with my tests?

Configure OTEL in the LocustTest CR:

```yaml
spec:
  openTelemetry:
    enabled: true
    endpoint: "http://otel-collector:4317"
    protocol: grpc  # or http
```

Traces and metrics flow directly to your OTEL collector. No sidecar required!

See [Advanced Topics](advanced_topics.md#opentelemetry) for full configuration.

---

## Still Have Questions?

- Check [Troubleshooting Guide](troubleshooting.md)
- Read [API Reference](api_reference.md)
- Search [GitHub Issues](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues)
- Ask in [GitHub Discussions](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/discussions)
```

---

### 4. Add Production-Ready Example CR in `config/samples/` (1-2 hours)

**File:** `config/samples/locusttest_v2_production.yaml`

**Why:** Users need a reference for best practices

**Content:**
```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: production-example
  labels:
    app: my-app
    environment: production
    team: performance-engineering
  annotations:
    # Prometheus scraping
    prometheus.io/scrape: "true"
    prometheus.io/port: "8089"
    # Ownership
    owner: "performance-team@example.com"
    runbook: "https://wiki.example.com/runbooks/locust-tests"
spec:
  # Test configuration
  image: locustio/locust:2.31.8
  configMap: production-test-scripts  # Your test scripts
  autostart: true
  autoQuit:
    enabled: true
    secondsAfterFinished: 300  # 5 minutes

  # Master pod configuration
  masterResourceSpec:
    requests:
      cpu: "500m"
      memory: "512Mi"
    limits:
      cpu: "1"
      memory: "1Gi"
    extraArgs:
      - "--expect-workers=10"
      - "--csv=/results/stats"
    labels:
      component: master
      monitoring: enabled
    annotations:
      sidecar.istio.io/inject: "false"  # Disable Istio if present

  # Worker configuration
  worker:
    replicas: 10  # Adjust based on load requirements
    affinity:
      # Spread workers across nodes for better load distribution
      podAntiAffinity:
        preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchLabels:
                locust_cr: production-example
                locust_role: worker
            topologyKey: kubernetes.io/hostname
    tolerations:
      # Allow scheduling on performance testing nodes
      - key: "workload-type"
        operator: "Equal"
        value: "performance-testing"
        effect: "NoSchedule"

  workerResourceSpec:
    requests:
      cpu: "250m"
      memory: "256Mi"
    limits:
      cpu: "500m"
      memory: "512Mi"
    extraArgs:
      - "--logfile=/var/log/locust/worker.log"
    labels:
      component: worker
      monitoring: enabled

  # Secret injection for authentication
  envSecrets:
  - secretName: api-credentials
    prefix: "API_"  # Keys become API_USERNAME, API_PASSWORD, etc.

  # Volume mounts for certificates or large test data
  volumes:
  - name: tls-certs
    secret:
      secretName: api-tls-certs
    volumeMounts:
    - name: tls-certs
      mountPath: /etc/ssl/certs/api
      readOnly: true
    target: both  # Mount on both master and workers

  # OpenTelemetry for distributed tracing
  openTelemetry:
    enabled: true
    endpoint: "http://otel-collector.observability:4317"
    protocol: grpc
    customAttributes:
      environment: production
      test_suite: api_load_test
      version: v1.0.0

  # Kafka/MSK configuration (if testing Kafka producers/consumers)
  # kafka:
  #   endpoint: "kafka-broker.kafka:9092"
  #   config:
  #     security.protocol: SASL_SSL
  #     sasl.mechanism: SCRAM-SHA-512

---
# Example ConfigMap with test scripts
apiVersion: v1
kind: ConfigMap
metadata:
  name: production-test-scripts
data:
  locustfile.py: |
    from locust import HttpUser, task, between
    import os

    class APIUser(HttpUser):
        wait_time = between(1, 3)

        def on_start(self):
            # Authenticate using injected secrets
            username = os.getenv("API_USERNAME")
            password = os.getenv("API_PASSWORD")
            self.client.post("/auth/login", json={
                "username": username,
                "password": password
            })

        @task(3)
        def get_users(self):
            self.client.get("/api/users")

        @task(2)
        def create_user(self):
            self.client.post("/api/users", json={
                "name": "Test User",
                "email": "test@example.com"
            })

        @task(1)
        def delete_user(self):
            self.client.delete("/api/users/123")

---
# Example Secret (create this separately, don't commit!)
# kubectl create secret generic api-credentials \
#   --from-literal=USERNAME=test-user \
#   --from-literal=PASSWORD=super-secret
```

**Also Add Comment Header:**
```yaml
# Production-Ready LocustTest Example
#
# This example demonstrates best practices for running Locust tests in production:
# - Resource requests/limits for predictable scheduling
# - Pod anti-affinity for worker distribution
# - Secret injection for credentials
# - Volume mounts for certificates
# - OpenTelemetry for observability
# - Labels and annotations for monitoring/ownership
# - AutoQuit for automatic cleanup
#
# Adjust values based on your specific requirements:
# - worker.replicas: Scale based on target load
# - resource requests/limits: Based on profiling your test scripts
# - affinity/tolerations: Based on your cluster topology
#
# See docs/advanced_topics.md for detailed configuration options.
```


---

## PRIORITY 2: Important Enhancements (Post-Launch, 15-20 hours)

### 6. Update Architecture Diagram in `docs/how_does_it_work.md` (3-4 hours)

**What:** Replace/update existing diagram to show v2 Go-specific details

**Show:**
- controller-runtime reconciliation loop
- Webhook validation flow (ValidatingWebhookConfiguration)
- Conversion webhook (v1 ↔ v2)
- Status update flow
- Pod health monitoring
- Leader election (HA mode)

**Tools:** Use Excalidraw, draw.io, or similar (export as SVG/PNG)

**Embed:**
```markdown
![Locust K8s Operator Architecture](../images/architecture-v2.svg)
```

---



### 8. Create `docs/security.md` (3-4 hours)

**Content Outline:**
```markdown
# Security Best Practices

## RBAC Configuration

### Minimal Permissions

The operator requires these permissions:
```yaml
# See charts/locust-k8s-operator/templates/serviceaccount-and-roles.yaml
# for complete RBAC configuration
```

### User RBAC (Namespace-scoped)

Grant users permission to create LocustTest resources:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: locust-test-user
  namespace: performance-testing
rules:
- apiGroups: ["locust.io"]
  resources: ["locusttests"]
  verbs: ["get", "list", "create", "delete"]
- apiGroups: ["locust.io"]
  resources: ["locusttests/status"]
  verbs: ["get", "list"]
```

## Secret Management

### DO NOT Hardcode Credentials

❌ **Bad:**
```yaml
spec:
  env:
  - name: API_KEY
    value: "hardcoded-secret-123"  # NEVER DO THIS
```

✅ **Good:**
```yaml
spec:
  envSecrets:
  - secretName: api-credentials  # Reference K8s Secret
```

### Secret Rotation

Secrets are mounted at pod creation. To rotate:
1. Update the Secret
2. Delete and recreate the LocustTest (triggers new pods with new secrets)

### External Secrets Integration

Use External Secrets Operator to sync from AWS Secrets Manager, HashiCorp Vault, etc:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: api-credentials
spec:
  secretStoreRef:
    name: aws-secrets-manager
  target:
    name: api-credentials
  data:
  - secretKey: USERNAME
    remoteRef:
      key: prod/api/username
  - secretKey: PASSWORD
    remoteRef:
      key: prod/api/password
```



---

### 9. Expand `docs/metrics_and_dashboards.md` (3-4 hours)

**Add These Sections:**

**Grafana Dashboard JSON:**
```json
{
  "dashboard": {
    "title": "Locust K8s Operator - Test Overview",
    "panels": [
      {
        "title": "Active Tests",
        "targets": [
          {
            "expr": "count(kube_pod_info{pod=~\".*-master.*\",namespace=\"performance-testing\"})"
          }
        ]
      },
      {
        "title": "Worker Connection Count",
        "targets": [
          {
            "expr": "locust_worker_connections{test=\"my-test\"}"
          }
        ]
      }
    ]
  }
}
```

**Sample Prometheus Alerts:**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: locust-operator-alerts
spec:
  groups:
  - name: locust-operator
    rules:
    - alert: LocustTestFailed
      expr: locust_test_phase{phase="Failed"} > 0
      for: 5m
      annotations:
        summary: "Locust test {{ $labels.test_name }} failed"
    - alert: LocustWorkerDisconnected
      expr: locust_worker_connections < locust_worker_replicas
      for: 10m
      annotations:
        summary: "Workers disconnected from test {{ $labels.test_name }}"
```

**Log Aggregation Patterns:**
```yaml
# Fluentd/Fluent-bit configuration for Locust logs
<filter kubernetes.**locust**>
  @type parser
  key_name log
  <parse>
    @type json
    time_key timestamp
    time_format %Y-%m-%dT%H:%M:%S.%LZ
  </parse>
</filter>
```

---

## PRIORITY 3: Polish (Ongoing, 5-10 hours)

### 10. Improve Navigation in `mkdocs.yml` (1 hour)

**Update nav structure:**
```yaml
nav:
  - Introduction: index.md
  - How does it work: how_does_it_work.md
  - Features: features.md

  - Getting Started:
      - Quick Start: getting_started.md
      - Installation: helm_deploy.md

  - User Guide:
      - API Reference: api_reference.md
      - Advanced Topics: advanced_topics.md
      - Troubleshooting: troubleshooting.md  # NEW
      - FAQ: faq.md  # NEW

  - Operations:
      - Performance Tuning: performance-tuning.md  # NEW
      - Security: security.md  # NEW
      - Metrics & Dashboards: metrics_and_dashboards.md

  - Migration & Upgrade:
      - Migration Guide: migration.md

  - Contributing: contribute.md
```

---

### 11. Add "Next Steps" Links at Bottom of Pages (2-3 hours)

**Pattern to add to each page:**

**At bottom of `getting_started.md`:**
```markdown
## Next Steps

- 📖 [Explore Advanced Features](advanced_topics.md) - OpenTelemetry, volumes, secrets
- 🔧 [Troubleshooting Guide](troubleshooting.md) - Common issues and solutions
- ❓ [FAQ](faq.md) - Frequently asked questions
- 📊 [Set Up Monitoring](metrics_and_dashboards.md) - Grafana dashboards and alerts
```

**At bottom of `api_reference.md`:**
```markdown
## Next Steps

- 🚀 [Advanced Topics](advanced_topics.md) - Deep dive into complex configurations
- ⚡ [Performance Tuning](performance-tuning.md) - Optimize for scale
- 📖 [Migration Guide](migration.md) - Upgrading from v1 to v2
```

---

### 12. Add Production-Ready Examples for Common Scenarios (2-4 hours)

**Create `config/samples/recipes/` directory with:**

**`recipes/ci-cd-pipeline.yaml`:**
```yaml
# CI/CD Pipeline Example
# Run automated load test in GitHub Actions / Jenkins / GitLab CI
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: ci-pipeline-test
spec:
  image: locustio/locust:2.31.8
  configMap: ci-test-scripts
  autostart: true
  autoQuit:
    enabled: true
    secondsAfterFinished: 60  # Quick cleanup for CI
  worker:
    replicas: 5  # Moderate scale for CI
  # ... rest of config
```

**`recipes/large-scale-test.yaml`:**
```yaml
# Large Scale Test Example
# Simulate 100,000+ concurrent users
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: large-scale-test
spec:
  worker:
    replicas: 200  # 200 workers × 500 users = 100K users
  # ... affinity, resources, etc.
```

**`recipes/multi-environment.yaml`:**
```yaml
# Multi-Environment Example
# Test dev, staging, prod with same scripts
# Use Kustomize or Helm for environment-specific values
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: api-test
  namespace: staging  # or dev, prod
spec:
  env:
  - name: TARGET_HOST
    value: "https://staging-api.example.com"  # Different per env
  # ... rest of config
```
