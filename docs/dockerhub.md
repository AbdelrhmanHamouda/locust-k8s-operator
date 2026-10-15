# Locust Kubernetes Operator

Kubernetes operator for [Locust](https://locust.io/). It runs distributed Locust load tests on Kubernetes: you describe a test as a `LocustTest` custom resource, and the operator creates the master and worker pods, connects them, reports whether the run passed or failed, and cleans up afterwards. It works the same on a laptop cluster and from a CI pipeline.

This image is the operator itself (the controller). You don't run it directly; install it with the Helm chart below. Your load tests run in the standard `locustio/locust` image, or your own image built on it.

- **Docs:** <https://abdelrhmanhamouda.github.io/locust-k8s-operator/>
- **Source:** <https://github.com/AbdelrhmanHamouda/locust-k8s-operator>
- **Helm chart on Artifact Hub:** <https://artifacthub.io/packages/helm/locust-k8s-operator/locust-k8s-operator>
- **Questions:** <https://github.com/AbdelrhmanHamouda/locust-k8s-operator/discussions>

## Quick install

Needs Kubernetes 1.29 or newer and Helm 3.

```bash
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/
helm repo update
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system --create-namespace
```

Then put a locustfile in a ConfigMap and start a test:

```bash
kubectl create configmap demo-test --from-file=demo_test.py

kubectl apply -f - <<EOF
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: demo
spec:
  image: locustio/locust:2.43.3
  testFiles:
    configMapRef: demo-test
  master:
    command: "--locustfile /lotest/src/demo_test.py --host https://example.com --users 10 --spawn-rate 2 --run-time 1m"
  worker:
    command: "--locustfile /lotest/src/demo_test.py"
    replicas: 2
EOF

kubectl get locusttest demo -w
```

The [Quick Start](https://abdelrhmanhamouda.github.io/locust-k8s-operator/getting_started/) and the [Locust on Kubernetes guide](https://abdelrhmanhamouda.github.io/locust-k8s-operator/locust-on-kubernetes/) go through this step by step.

## What it handles

- Master and worker Jobs plus the master Service, created from one resource
- `Pending` / `Running` / `Succeeded` / `Failed` phases and status conditions, so CI can wait on a run and fail the build
- Worker scaling up to 500 replicas, with separate resource requests and limits for master and workers
- Secrets and ConfigMaps as environment variables or files, extra volumes, private registries
- Node selectors, affinity, tolerations and `runtimeClassName` for placing load generators
- Native OpenTelemetry export from Locust, or a Prometheus exporter sidecar
- Validation and conversion webhooks; `locust.io/v1` resources keep working on the v2 API

## Tags

Images are multi-arch (`linux/amd64`, `linux/arm64`). Each release is tagged with its version, and `latest` points to the newest release. Pin a version in production; the Helm chart does this for you.

Licence: Apache-2.0.
