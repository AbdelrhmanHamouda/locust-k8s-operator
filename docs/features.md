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

Everything the Locust Kubernetes Operator can do. Click any feature to learn how.

<div class="grid cards" markdown>

-   <a id="cloud-native"></a> :material-cloud-check: **Cloud Native & Kubernetes Integration**

    ---

    Leverage the full power of Kubernetes for distributed performance testing. The operator is designed to be cloud-native, enabling seamless deployment and scaling on any Kubernetes cluster.

    [:octicons-arrow-right-24: How it works](how_does_it_work.md)

-   <a id="automation"></a> :material-robot-happy: **Automation & CI/CD**

    ---

    Integrate performance testing directly into your CI/CD pipelines. Automate the deployment, execution, and teardown of your Locust tests for continuous performance validation.

    [:octicons-arrow-right-24: CI/CD integration tutorial](tutorials/ci-cd-integration.md)

-   <a id="governance"></a> :material-shield-check: **Governance & Resource Management**

    ---

    Maintain control over how resources are deployed and used. Configure resource requests and limits for Locust master and worker pods, and even disable CPU limits for performance-sensitive tests.

    [:octicons-arrow-right-24: Configure resources](how-to-guides/configuration/configure-resources.md)

-   <a id="observability"></a> :material-chart-bar: **Observability & Monitoring**

    ---

    Gain insights into test results and infrastructure usage. The operator supports Prometheus metrics out-of-the-box and native OpenTelemetry integration.

    [:octicons-arrow-right-24: Configure OpenTelemetry](how-to-guides/observability/configure-opentelemetry.md) · [:octicons-arrow-right-24: Metrics reference](metrics_and_dashboards.md)

-   :material-scale-balance: **Cost Optimization**

    ---

    Optimize cloud costs by deploying resources only when needed and for as long as needed. The operator's automatic cleanup feature ensures that resources are terminated after a test run.

    [:octicons-arrow-right-24: Configure TTL](how-to-guides/configuration/configure-ttl.md)

-   :material-layers-triple: **Test Isolation & Parallelism**

    ---

    Run multiple tests in parallel with guaranteed isolation. Each test runs in its own set of resources, preventing any cross-test interference.

    [:octicons-arrow-right-24: How it works](how_does_it_work.md)

-   :material-docker: **Private Image Registry Support**

    ---

    Use images from private registries for your Locust tests. The operator supports `imagePullSecrets` and configurable `imagePullPolicy`.

    [:octicons-arrow-right-24: Use private registry](how-to-guides/configuration/use-private-registry.md)

-   :material-folder-multiple: **Lib ConfigMap Support**

    ---

    Mount lib directories via ConfigMap for your Locust tests. This feature allows you to include shared libraries and modules without modifying test files or patching images, similar to the helm chart's `locust_lib_configmap` functionality.

-   :material-kubernetes: **Advanced Scheduling**

    ---

    Control where your Locust pods are scheduled using Kubernetes affinity and taint tolerations. This allows you to run tests on dedicated nodes or in specific availability zones.

    [:octicons-arrow-right-24: Use node affinity](how-to-guides/scaling/use-node-affinity.md) · [:octicons-arrow-right-24: Configure tolerations](how-to-guides/scaling/configure-tolerations.md) · [:octicons-arrow-right-24: Use node selector](how-to-guides/scaling/use-node-selector.md)

-   :material-apache-kafka: **Kafka & AWS MSK Integration**

    ---

    Seamlessly integrate with Kafka and AWS MSK for performance testing of event-driven architectures. The operator provides out-of-the-box support for authenticated Kafka.

    [:octicons-arrow-right-24: Configure Kafka](how-to-guides/configuration/configure-kafka.md)

-   :material-chart-timeline: **Native OpenTelemetry Support**

    ---

    Export traces and metrics directly from Locust using native OpenTelemetry integration. No sidecar required—configure endpoints, protocols, and custom attributes directly in your CR.

    [:octicons-arrow-right-24: Configure OpenTelemetry](how-to-guides/observability/configure-opentelemetry.md)

-   :material-key-variant: **Secret & ConfigMap Injection**

    ---

    Securely inject credentials, API keys, and configuration from Kubernetes Secrets and ConfigMaps. Supports environment variables and file mounts with automatic prefix handling.

    [:octicons-arrow-right-24: Inject secrets](how-to-guides/security/inject-secrets.md)

-   :material-harddisk: **Flexible Volume Mounting**

    ---

    Mount test data, certificates, and configuration files from PersistentVolumes, ConfigMaps, or Secrets. Target specific components (master, worker, or both) with fine-grained control.

    [:octicons-arrow-right-24: Mount volumes](how-to-guides/configuration/mount-volumes.md)

-   :material-tune-vertical: **Separate Resource Specs**

    ---

    Configure resources, labels, and annotations independently for master and worker pods. Optimize each component based on its specific needs.

    [:octicons-arrow-right-24: API Reference](api_reference.md)

-   :material-list-status: **Enhanced Status Tracking**

    ---

    Monitor test progress with rich status information including phase (Pending, Running, Succeeded, Failed), Kubernetes conditions, and worker connection status.

    [:octicons-arrow-right-24: Monitor test status](how-to-guides/observability/monitor-test-status.md) · [:octicons-arrow-right-24: API Reference](api_reference.md#status-fields)

</div>
