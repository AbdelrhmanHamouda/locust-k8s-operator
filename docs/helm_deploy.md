---
title: HELM Deployment
description: Instructions on how to deploy the Locust Kubernetes Operator with HELM.
---

# HELM Deployment Guide

This guide provides comprehensive instructions for deploying the Locust Kubernetes Operator using its official Helm chart.

## Quick Start

For experienced users, here are the essential commands to get the operator running in the `default` namespace:

```bash
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/
helm repo update
helm install locust-operator locust-k8s-operator/locust-k8s-operator
```

## Installation

### Prerequisites

*   A running Kubernetes cluster (e.g., Minikube, GKE, EKS, AKS).
*   [Helm 3](https://helm.sh/docs/intro/install/) installed on your local machine.

### Step 1: Add the Helm Repository

First, add the Locust Kubernetes Operator Helm repository to your local Helm client:

```bash
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/
```

Next, update your local chart repository cache to ensure you have the latest version:

```bash
helm repo update
```

### Step 2: Install the Chart

You can install the chart with a release name of your choice (e.g., `locust-operator`).

**Default Installation:**

To install the chart with the default configuration into the currently active namespace, run:

```bash
helm install locust-operator locust-k8s-operator/locust-k8s-operator
```

**Installation with a Custom Values File:**

For more advanced configurations, it's best to use a custom `values.yaml` file. Create a file named `my-values.yaml` and add your overrides:

```yaml
# my-values.yaml
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

## Verifying the Installation

After installation, you can verify that the operator is running correctly by checking the pods in the target namespace:

```bash
kubectl get pods -n locust-system
```

You should see a pod with a name similar to `locust-operator-b5c9f4f7-xxxxx` in the `Running` state.

To view the operator's logs, run:

```bash
kubectl logs -f -n locust-system -l app.kubernetes.io/name=locust-k8s-operator
```

## Configuration

The following tables list the configurable parameters of the Locust Operator Helm chart and their default values.

### General Configuration

| Parameter | Description | Default |
|---|---|---|
| `appPort` | The port that the operator will listen on. | `8080` |

### Deployment Settings

| Parameter | Description | Default |
|---|---|---|
| `replicaCount` | Number of replicas for the operator deployment. | `1` |
| `image.repository` | The repository of the Docker image. | `lotest/locust-k8s-operator` |
| `image.pullPolicy` | The image pull policy. | `IfNotPresent` |
| `image.tag` | Overrides the default image tag (defaults to the chart's `appVersion`). | `""` |

### Kubernetes Resources

| Parameter | Description | Default |
|---|---|---|
| `k8s.customResourceDefinition.deploy` | Specifies whether to deploy the `LocustTest` CRD. | `true` |
| `k8s.clusterRole.enabled` | Deploy with a cluster-wide role (`true`) or a namespaced role (`false`). | `true` |
| `serviceAccount.create` | Specifies whether a service account should be created. | `true` |
| `serviceAccount.name` | The name of the service account to use. If not set, a name is generated. | `""` |
| `resources` | Resource requests and limits for the operator pod. | `{}` |

### Operator Configuration

| Parameter | Description | Default |
|---|---|---|
| `config.loadGenerationJobs.ttlSecondsAfterFinished` | Time-to-live in seconds for finished load generation jobs. Set to `""` to disable. | `""` |

#### Load Generation Pods

| Parameter | Description | Default |
|---|---|---|
| `config.loadGenerationPods.resource.cpuRequest` | CPU resource request for load generation pods. | `250m` |
| `config.loadGenerationPods.resource.memRequest` | Memory resource request for load generation pods. | `128Mi` |
| `config.loadGenerationPods.resource.ephemeralRequest` | Ephemeral storage request for load generation pods. | `30M` |
| `config.loadGenerationPods.resource.cpuLimit` | CPU resource limit for load generation pods. Set to `""` to unbind. | `1000m` |
| `config.loadGenerationPods.resource.memLimit` | Memory resource limit for load generation pods. Set to `""` to unbind. | `1024Mi` |
| `config.loadGenerationPods.resource.ephemeralLimit` | Ephemeral storage limit for load generation pods. Set to `""` to unbind. | `50M` |
| `config.loadGenerationPods.affinity.enableCrInjection` | Enable Custom Resource injection for affinity settings. | `true` |
| `config.loadGenerationPods.taintTolerations.enableCrInjection` | Enable Custom Resource injection for taint tolerations. | `true` |

#### Metrics Exporter

| Parameter | Description | Default |
|---|---|---|
| `config.loadGenerationPods.metricsExporter.image` | Metrics Exporter Docker image. | `containersol/locust_exporter:v0.5.0` |
| `config.loadGenerationPods.metricsExporter.port` | Metrics Exporter port. | `9646` |
| `config.loadGenerationPods.metricsExporter.pullPolicy` | Image pull policy for the metrics exporter. | `IfNotPresent` |

### Pod Scheduling

| Parameter | Description | Default |
|---|---|---|
| `nodeSelector` | Node selector for scheduling the operator pod. | `{}` |
| `tolerations` | Tolerations for scheduling the operator pod. | `[]` |
| `affinity` | Affinity rules for scheduling the operator pod. | `{}` |

### Liveness and Readiness Probes

| Parameter | Description | Default |
|---|---|---|
| `livenessProbe.initialDelaySeconds` | Initial delay for the liveness probe. | `10` |
| `livenessProbe.periodSeconds` | How often to perform the liveness probe. | `20` |
| `livenessProbe.timeoutSeconds` | When the liveness probe times out. | `10` |
| `livenessProbe.failureThreshold` | When to give up on the liveness probe. | `1` |
| `readinessProbe.initialDelaySeconds` | Initial delay for the readiness probe. | `30` |
| `readinessProbe.periodSeconds` | How often to perform the readiness probe. | `20` |
| `readinessProbe.timeoutSeconds` | When the readiness probe times out. | `10` |
| `readinessProbe.failureThreshold` | When to give up on the readiness probe. | `1` |

### Advanced Configuration

The following sections cover advanced configuration options. For a complete list of parameters, refer to the `values.yaml` file in the chart.

#### Kafka Configuration

| Parameter | Description | Default |
|---|---|---|
| `config.loadGenerationPods.kafka.bootstrapServers` | Kafka bootstrap servers. | `localhost:9092` |
| `config.loadGenerationPods.kafka.acl.enabled` | Enable ACL settings for Kafka. | `false` |
| `config.loadGenerationPods.kafka.sasl.mechanism` | SASL mechanism for authentication. | `SCRAM-SHA-512` |

#### Micronaut Metrics

| Parameter | Description | Default |
|---|---|---|
| `micronaut.metrics.enabled` | Enable/disable all Micronaut metrics. | `true` |
| `micronaut.metrics.export.prometheus.step` | The step size (duration) for Prometheus metrics export. | `PT30S` |

## Upgrading the Chart

To upgrade an existing release to a new version, use the `helm upgrade` command:

```bash
helm upgrade locust-operator locust-k8s-operator/locust-k8s-operator -f my-values.yaml
```

## Uninstalling the Chart

To uninstall and delete the `locust-operator` deployment, run:

```bash
helm uninstall locust-operator
```

This command will remove all the Kubernetes components associated with the chart and delete the release.

## Next Steps

Once the operator is installed, you're ready to start running performance tests! Head over to the [Getting Started](./getting_started.md) guide to learn how to deploy your first `LocustTest`.
