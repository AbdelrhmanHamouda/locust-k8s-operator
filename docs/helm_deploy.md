---
title: HELM deployment
description: Instructions on how to deploy Locust Kubernetes Operator with HELM
---

# HELM deployment

In order to deploy using helm, follow the below steps:

1.  Add the _Operator´s_ HELM repo

    - `helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/`

    !!! note

        If the repo has been added before, run `helm repo update` in order to pull the latest available release!

2.  Install the _Operator_

    - `#!bash helm install locust-operator locust-k8s-operator/locust-k8s-operator` - The _Operator_ will ready up in around 40-60 seconds
    - This will cause the bellow resources to be deployed in the currently active _k8s_ context & namespace.
      - [crd-locusttest.yaml]
        - This _CRD_ is the first part of the _Operator_ pattern. It is needed in order to enable _Kubernetes_ to understand the _LocustTest_
          custom resource and allow its deployment.
      - [serviceaccount-and-roles.yaml]
        - ServiceAccount and Role bindings that enable the _Controller_ to have the needed privilege inside the cluster to watch and
          manage the related resources.
      - [deployment.yaml]
        - The _Controller_ responsible for managing and reacting to the cluster resources.

---------

## Defaults configuration values

The Locust-K8s-Operator comes with a set of sensible default configuration values that are designed to work seamlessly on a wide range of Kubernetes clusters. These defaults are carefully chosen to provide a smooth out-of-the-box experience.

### How to override

While the default configuration covers most scenarios, there might be cases where you need to customize the values to align with your specific requirements. The Locust-K8s-Operator is flexible and allows you to easily override these default values.

The most straightforward and organized approach to customize configuration values is by using a values file. This involves creating a separate YAML file that contains your desired configuration overrides. You can then apply this file when deploying the Helm chart.

To learn more about how to use a values file and override configuration settings, refer to [HELM's official documentation] on Values Files.

By following this approach, you can tailor the _Locust-K8s-Operator_'s behavior to match your specific needs while maintaining a structured and maintainable configuration setup.


### Default values overview

This section serves as a comprehensive guide to the configurable options present within the Helm chart. Each subsection details the precise value names applicable for overrides, and when necessary, highlight the underlying rationale behind each configuration.

#### General Configuration
- `appPort`: Specifies the port that the _Operator_ will listen on.

#### Deployment Settings
- `replicaCount`: Number of replicas for the _Operator_ deployment.
- `image.repository`: The repository of the Docker image.
- `image.pullPolicy`: The image pull policy.
- `image.tag`: Overrides the default image tag (useful for specifying a version).

##### Liveness and Readiness Probes

Configure liveness and readiness probes for the operator.

- `[livenessProbe || readinessProbe].httpGet.scheme`:  The scheme for the probe (HTTP).
- `[livenessProbe || readinessProbe].httpGet.path`: The path to probe.
- `[livenessProbe || readinessProbe].httpGet.port`: The port to probe.
- `[livenessProbe || readinessProbe].initialDelaySeconds`: Initial delay before starting probes.
- `[livenessProbe || readinessProbe].periodSeconds`: Time between consecutive probes.
- `[livenessProbe || readinessProbe].timeoutSeconds`: Probe timeout.
- `[livenessProbe || readinessProbe].failureThreshold`: Number of consecutive failures before considering the probe failed.

#### Kubernetes Configuration
  
- `k8s.customResourceDefinition.deploy`: Specifies whether to deploy the _LocustTest_ custom resource definition.

##### Service Account Configuration

- `serviceAccount.create`: Specifies whether a service account should be created.
- `serviceAccount.annotations`: Annotations to add to the service account.
- `serviceAccount.name`: The name of the service account to use.

##### Pod Annotations

- `podAnnotations`: Additional annotations to apply to the _Operator_.

##### Resources Configuration

Configure resource requests and limits for the _Operator_.

- `resources`: Resource requests and limits for the _Operator_.

##### Node Selector

- `nodeSelector`: Node selector settings for the _Operator_ pod scheduling.

##### Tolerations

- `tolerations`: Tolerations for the _Operator_ pod scheduling.

##### Affinity Rules

- `affinity`: Affinity rules for the _Operator_ pod scheduling.

#### Operator Configuration

- `config.loadGenerationJobs.ttlSecondsAfterFinished`: Time-to-live in seconds for finished load generation jobs.

##### Load Generation Pods Configuration

- `config.loadGenerationPods.resource.cpuRequest`: CPU resource _request_ for load generation pods.
- `config.loadGenerationPods.resource.memRequest`: Memory resource _request_ for load generation pods.
- `config.loadGenerationPods.resource.ephemeralRequest`: Ephemeral Storage resource _request_ for load generation pods.
- `config.loadGenerationPods.resource.cpuLimit`: CPU resource _limit_ for load generation pods.
- `config.loadGenerationPods.resource.memLimit`: Memory resource _limit_ for load generation pods.
- `config.loadGenerationPods.resource.ephemeralLimit`: Ephemeral Storage resource _limit_ for load generation pods.
- `config.loadGenerationPods.affinity.enableCrInjection`: Enable Custom Resource injection for affinity settings.
- `config.loadGenerationPods.taintTolerations.enableCrInjection`: Enable Custom Resource injection for taint tolerations settings.

##### Metrics Exporter Configuration

- `config.loadGenerationPods.metricsExporter.image`: Metrics Exporter Docker image.
- `config.loadGenerationPods.metricsExporter.port`: Metrics Exporter port.
- `config.loadGenerationPods.metricsExporter.pullPolicy`: Image pull policy.
- `config.loadGenerationPods.metricsExporter.resource.cpuRequest`: CPU resource _request_ for Metrics Exporter container.
- `config.loadGenerationPods.metricsExporter.resource.memRequest`: Memory resource _request_ for Metrics Exporter container.
- `config.loadGenerationPods.metricsExporter.resource.ephemeralRequest`: Ephemeral Storage resource _request_ for Metrics Exporter container.
- `config.loadGenerationPods.metricsExporter.resource.cpuLimit`: CPU resource _limit_ for Metrics Exporter container.
- `config.loadGenerationPods.metricsExporter.resource.memLimit`: Memory resource _limit_ for Metrics Exporter container.
- `config.loadGenerationPods.metricsExporter.resource.ephemeralLimit`: Ephemeral Storage resource _limit_ for Metrics Exporter container.

##### Kafka Configuration

- `config.loadGenerationPods.kafka.bootstrapServers`: Kafka bootstrap servers.
- `config.loadGenerationPods.kafka.locustK8sKafkaUser`: Kafka user for Locust-K8s communication.
- `config.loadGenerationPods.kafka.acl.enabled`: Enable ACL settings.
- `config.loadGenerationPods.kafka.acl.protocol`: ACL protocol.
- `config.loadGenerationPods.kafka.acl.secret.userKey`: Key for the Kafka username secret.
- `config.loadGenerationPods.kafka.acl.secret.passwordKey`: Key for the Kafka password secret.
- `config.loadGenerationPods.kafka.sasl.mechanism`: SASL mechanism.
- `config.loadGenerationPods.kafka.sasl.jaas.config`: JAAS configuration for SASL.

##### Micronaut Application Configuration

Configure Micronaut application metrics.

- `micronaut.metrics.enabled`: Enable/disable metrics.
- `micronaut.metrics.web.enabled`: Enable/disable web metrics.
- `micronaut.metrics.jvm.enabled: Enable/disable JVM metrics.
- ... (Continues for other Micronaut metrics options)


[//]: # "Resources urls"
[crd-locusttest.yaml]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/kube/crd/locust-test-crd.yaml
[serviceaccount-and-roles.yaml]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/charts/locust-k8s-operator/templates/serviceaccount-and-roles.yaml
[deployment.yaml]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/charts/locust-k8s-operator/templates/deployment.yaml
[HELM's official documentation]: https://helm.sh/docs/chart_template_guide/values_files/
