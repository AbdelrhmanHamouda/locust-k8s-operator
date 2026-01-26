---
title: HELM Deployment
description: Instructions on how to deploy the Locust Kubernetes Operator with HELM.
tags:
  - deployment
  - helm
  - installation
  - kubernetes
  - setup
---

# HELM Deployment Guide

This guide provides comprehensive instructions for deploying the Locust Kubernetes Operator using its official Helm chart.

## :material-run-fast: Quick Start

For experienced users, here are the essential commands to get the operator running in the `default` namespace:

```bash
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/
helm repo update
helm install locust-operator locust-k8s-operator/locust-k8s-operator
```

## :material-download-box-outline: Installation

### :material-check-circle-outline: Prerequisites

*   A running Kubernetes cluster (e.g., Minikube, GKE, EKS, AKS).
*   [Helm 3](https://helm.sh/docs/intro/install/) installed on your local machine.

### :material-source-repository: Step 1: Add the Helm Repository

First, add the Locust Kubernetes Operator Helm repository to your local Helm client:

```bash
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/
```

Next, update your local chart repository cache to ensure you have the latest version:

```bash
helm repo update
```

### :material-package-variant-closed: Step 2: Install the Chart

You can install the chart with a release name of your choice (e.g., `locust-operator`).

**Default Installation:**

To install the chart with the default configuration into the currently active namespace, run:

```bash
helm install locust-operator locust-k8s-operator/locust-k8s-operator
```

**Installation with a Custom Values File:**

For more advanced configurations, it's best to use a custom `values.yaml` file. Create a file named `my-values.yaml` and add your overrides:

=== "v2 Helm Values (Recommended)"

    ```yaml
    # my-values.yaml
    replicaCount: 2
    
    locustPods:
      resources:
        limits:
          cpu: "2000m"
          memory: "2048Mi"
        requests:
          cpu: "500m"
          memory: "512Mi"
    ```

=== "v1 Helm Values (Deprecated)"

    ```yaml
    # my-values.yaml (old format - still works via compatibility shims)
    replicaCount: 2
    
    config:
      loadGenerationPods:
        resource:
          cpuLimit: "2000m"
          memLimit: "2048Mi"
    ```

Then, install the chart, specifying your custom values file and a target namespace:

```bash
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system \
  --create-namespace \
  -f my-values.yaml
```

## :material-check-decagram-outline: Verifying the Installation

After installation, you can verify that the operator is running correctly by checking the pods in the target namespace:

```bash
kubectl get pods -n locust-system
```

You should see a pod with a name similar to `locust-operator-b5c9f4f7-xxxxx` in the `Running` state.

To view the operator's logs, run:

```bash
kubectl logs -f -n locust-system -l app.kubernetes.io/name=locust-k8s-operator
```

## :material-tune: Configuration

The following tables list the configurable parameters of the Locust Operator Helm chart and their default values.

!!! info "v2.0 Changes"
    The v2 Helm chart has been updated for the Go operator. Java-specific settings (Micronaut, JVM) have been removed. Backward compatibility shims are provided for common settings.

### Deployment Settings

| Parameter | Description | Default |
|---|---|---|
| `replicaCount` | Number of replicas for the operator deployment. | `2` |
| `image.repository` | The repository of the Docker image. | `lotest/locust-k8s-operator` |
| `image.pullPolicy` | The image pull policy. | `IfNotPresent` |
| `image.tag` | Overrides the default image tag (defaults to the chart's `appVersion`). | `""` |
| `image.pullSecrets` | List of image pull secrets. | `[]` |

### Kubernetes Resources

| Parameter | Description | Default |
|---|---|---|
| `crd.install` | Specifies whether to deploy the `LocustTest` CRD. | `true` |
| `k8s.clusterRole.enabled` | Deploy with a cluster-wide role (`true`) or a namespaced role (`false`). | `true` |
| `serviceAccount.create` | Specifies whether a service account should be created. | `true` |
| `serviceAccount.name` | The name of the service account to use. If empty and `serviceAccount.create` is `true`, a name is generated using the release name. If `serviceAccount.create` is `false`, defaults to `default`. | `""` |
| `serviceAccount.annotations` | Annotations to add to the service account. | `{}` |

### Operator Resources

The Go operator requires significantly fewer resources than the Java version:

| Parameter | Description | Default |
|---|---|---|
| `resources.limits.memory` | Operator memory limit. | `128Mi` |
| `resources.limits.cpu` | Operator CPU limit. | `500m` |
| `resources.requests.memory` | Operator memory request. | `64Mi` |
| `resources.requests.cpu` | Operator CPU request. | `10m` |

### Feature Toggles

| Parameter | Description | Default |
|---|---|---|
| `leaderElection.enabled` | Enable leader election for HA deployments. | `true` |
| `metrics.enabled` | Enable Prometheus metrics endpoint. | `false` |
| `metrics.port` | Metrics server port. | `8080` |
| `metrics.secure` | Use HTTPS for metrics endpoint. | `false` |
| `webhook.enabled` | Enable conversion webhook (requires cert-manager). | `false` |

### Webhook Configuration

Required when `webhook.enabled: true`:

| Parameter | Description | Default |
|---|---|---|
| `webhook.port` | Webhook server port. | `9443` |
| `webhook.certManager.enabled` | Use cert-manager for TLS certificate management. | `true` |

!!! note
    The conversion webhook requires [cert-manager](https://cert-manager.io/) to be installed in your cluster for automatic TLS certificate management.

### Locust Pod Configuration

| Parameter | Description | Default |
|---|---|---|
| `locustPods.resources.requests.cpu` | CPU request for Locust pods. | `250m` |
| `locustPods.resources.requests.memory` | Memory request for Locust pods. | `128Mi` |
| `locustPods.resources.requests.ephemeralStorage` | Ephemeral storage request for Locust pods. | `30M` |
| `locustPods.resources.limits.cpu` | CPU limit for Locust pods. Set to `""` to unbind. | `1000m` |
| `locustPods.resources.limits.memory` | Memory limit for Locust pods. Set to `""` to unbind. | `1024Mi` |
| `locustPods.resources.limits.ephemeralStorage` | Ephemeral storage limit for Locust pods. | `50M` |
| `locustPods.affinityInjection` | Enable affinity injection from CRs. | `true` |
| `locustPods.tolerationsInjection` | Enable tolerations injection from CRs. | `true` |

### Metrics Exporter

| Parameter | Description | Default |
|---|---|---|
| `locustPods.metricsExporter.image` | Metrics Exporter Docker image. | `containersol/locust_exporter:v0.5.0` |
| `locustPods.metricsExporter.port` | Metrics Exporter port. | `9646` |
| `locustPods.metricsExporter.pullPolicy` | Image pull policy for the metrics exporter. | `IfNotPresent` |
| `locustPods.metricsExporter.resources.requests.cpu` | CPU request for metrics exporter. | `100m` |
| `locustPods.metricsExporter.resources.requests.memory` | Memory request for metrics exporter. | `64Mi` |
| `locustPods.metricsExporter.resources.requests.ephemeralStorage` | Ephemeral storage request for metrics exporter. | `30M` |
| `locustPods.metricsExporter.resources.limits.cpu` | CPU limit for metrics exporter. | `250m` |
| `locustPods.metricsExporter.resources.limits.memory` | Memory limit for metrics exporter. | `128Mi` |
| `locustPods.metricsExporter.resources.limits.ephemeralStorage` | Ephemeral storage limit for metrics exporter. | `50M` |

!!! tip
    When using OpenTelemetry (`spec.observability.openTelemetry.enabled: true`), the metrics exporter sidecar is not deployed.

### Job Configuration

| Parameter | Description | Default |
|---|---|---|
| `locustPods.ttlSecondsAfterFinished` | TTL for finished jobs. Set to `""` to disable. | `""` |

### Kafka Configuration

| Parameter | Description | Default |
|---|---|---|
| `kafka.enabled` | Enable Kafka configuration injection. | `false` |
| `kafka.bootstrapServers` | Kafka bootstrap servers. | `localhost:9092` |
| `kafka.security.enabled` | Enable Kafka security. | `false` |
| `kafka.security.protocol` | Security protocol (`SASL_SSL`, `SASL_PLAINTEXT`, etc.). | `SASL_PLAINTEXT` |
| `kafka.security.saslMechanism` | SASL mechanism. | `SCRAM-SHA-512` |
| `kafka.security.jaasConfig` | JAAS configuration string. | `""` |
| `kafka.credentials.secretName` | Name of secret containing Kafka credentials. | `""` |
| `kafka.credentials.usernameKey` | Key in secret for username. | `username` |
| `kafka.credentials.passwordKey` | Key in secret for password. | `password` |

### OpenTelemetry Collector (Optional)

Deploy an OTel Collector alongside the operator:

| Parameter | Description | Default |
|---|---|---|
| `otelCollector.enabled` | Deploy OTel Collector. | `false` |
| `otelCollector.image` | Collector image. | `otel/opentelemetry-collector-contrib:0.92.0` |
| `otelCollector.replicas` | Number of collector replicas. | `1` |
| `otelCollector.resources.requests.cpu` | CPU request for collector. | `50m` |
| `otelCollector.resources.requests.memory` | Memory request for collector. | `64Mi` |
| `otelCollector.resources.limits.cpu` | CPU limit for collector. | `200m` |
| `otelCollector.resources.limits.memory` | Memory limit for collector. | `256Mi` |
| `otelCollector.config` | OTel Collector configuration (YAML string). | See values.yaml |

### Pod Scheduling

| Parameter | Description | Default |
|---|---|---|
| `nodeSelector` | Node selector for scheduling the operator pod. | `{}` |
| `tolerations` | Tolerations for scheduling the operator pod. | `[]` |
| `affinity` | Affinity rules for scheduling the operator pod. | `{}` |
| `podAnnotations` | Annotations to add to the operator pod. | `{}` |

### Backward Compatibility

The following v1 paths are still supported via helper functions:

| Old Path (v1) | New Path (v2) |
|---------------|---------------|
| `config.loadGenerationPods.resource.cpuRequest` | `locustPods.resources.requests.cpu` |
| `config.loadGenerationPods.resource.memLimit` | `locustPods.resources.limits.memory` |
| `config.loadGenerationPods.affinity.enableCrInjection` | `locustPods.affinityInjection` |
| `config.loadGenerationPods.kafka.*` | `kafka.*` |

!!! warning "Removed Settings"
    The following Java-specific settings have been removed and have no effect in v2:
    
    - `appPort` - Fixed at 8081
    - `micronaut.*` - No Micronaut in Go operator
    - `livenessProbe.*` / `readinessProbe.*` - Fixed probes on `/healthz` and `/readyz`

## :material-arrow-up-bold-box-outline: Upgrading the Chart

To upgrade an existing release to a new version, use the `helm upgrade` command:

```bash
helm upgrade locust-operator locust-k8s-operator/locust-k8s-operator -f my-values.yaml
```

## :material-trash-can-outline: Uninstalling the Chart

To uninstall and delete the `locust-operator` deployment, run:

```bash
helm uninstall locust-operator
```

This command will remove all the Kubernetes components associated with the chart and delete the release.

## :material-arrow-right-bold-box-outline: Next Steps

Once the operator is installed, you're ready to start running performance tests! Head over to the [Getting Started](./getting_started.md) guide to learn how to deploy your first `LocustTest`.
