---
title: Features
description: List of available features
tags:
  - features
  - capabilities
  - cloud native
  - kubernetes
  - automation
---

# Features

<div class="grid cards" markdown>

-   <a id="cloud-native"></a> :material-cloud-check: **Cloud Native & Kubernetes Integration**

    ---

    Leverage the full power of Kubernetes for distributed performance testing. The operator is designed to be cloud-native, enabling seamless deployment and scaling on any Kubernetes cluster.

-   <a id="automation"></a> :material-robot-happy: **Automation & CI/CD**

    ---

    Integrate performance testing directly into your CI/CD pipelines. Automate the deployment, execution, and teardown of your Locust tests for continuous performance validation.

-   <a id="governance"></a> :material-shield-check: **Governance & Resource Management**

    ---

    Maintain control over how resources are deployed and used. Configure resource requests and limits for Locust master and worker pods, and even disable CPU limits for performance-sensitive tests.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#resource-management)

-   <a id="observability"></a> :material-chart-bar: **Observability & Monitoring**

    ---

    Gain insights into test results and infrastructure usage. The operator supports Prometheus metrics out-of-the-box, allowing you to build rich monitoring dashboards.

    [:octicons-arrow-right-24: Learn more](metrics_and_dashboards.md)

-   :material-scale-balance: **Cost Optimization**

    ---

    Optimize cloud costs by deploying resources only when needed and for as long as needed. The operator's automatic cleanup feature ensures that resources are terminated after a test run.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#automatic-cleanup)

-   :material-layers-triple: **Test Isolation & Parallelism**

    ---

    Run multiple tests in parallel with guaranteed isolation. Each test runs in its own set of resources, preventing any cross-test interference.

-   :material-docker: **Private Image Registry Support**

    ---

    Use images from private registries for your Locust tests. The operator supports `imagePullSecrets` and configurable `imagePullPolicy`.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#private-image-registry)

-   :material-folder-multiple: **Lib ConfigMap Support**

    ---

    Mount lib directories via ConfigMap for your Locust tests. This feature allows you to include shared libraries and modules without modifying test files or patching images, similar to the helm chart's `locust_lib_configmap` functionality.

-   :material-kubernetes: **Advanced Scheduling**

    ---

    Control where your Locust pods are scheduled using Kubernetes affinity and taint tolerations. This allows you to run tests on dedicated nodes or in specific availability zones.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#dedicated-kubernetes-nodes)

-   :material-apache-kafka: **Kafka & AWS MSK Integration**

    ---

    Seamlessly integrate with Kafka and AWS MSK for performance testing of event-driven architectures. The operator provides out-of-the-box support for authenticated Kafka.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#kafka-integration)

-   :material-chart-timeline: **Native OpenTelemetry Support**

    ---

    Export traces and metrics directly from Locust using native OpenTelemetry integration. No sidecar required—configure endpoints, protocols, and custom attributes directly in your CR.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#opentelemetry-integration)

-   :material-key-variant: **Secret & ConfigMap Injection**

    ---

    Securely inject credentials, API keys, and configuration from Kubernetes Secrets and ConfigMaps. Supports environment variables and file mounts with automatic prefix handling.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#environment-secret-injection)

-   :material-harddisk: **Flexible Volume Mounting**

    ---

    Mount test data, certificates, and configuration files from PersistentVolumes, ConfigMaps, or Secrets. Target specific components (master, worker, or both) with fine-grained control.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#volume-mounting)

-   :material-tune-vertical: **Separate Resource Specs**

    ---

    Configure resources, labels, and annotations independently for master and worker pods. Optimize each component based on its specific needs.

    [:octicons-arrow-right-24: Learn more](advanced_topics.md#separate-resource-specs)

-   :material-list-status: **Enhanced Status Tracking**

    ---

    Monitor test progress with rich status information including phase (Pending, Running, Succeeded, Failed), Kubernetes conditions, and worker connection status.

    [:octicons-arrow-right-24: Learn more](api_reference.md#status-fields)

</div>

