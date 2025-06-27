---
title: Features
description: List of available features 
---

# Features

<div class="grid cards" markdown>

-   :material-cloud-check: **Cloud Native & Kubernetes Integration**

    ---

    Leverage the full power of Kubernetes for distributed performance testing. The operator is designed to be cloud-native, enabling seamless deployment and scaling on any Kubernetes cluster.

-   :material-robot-happy: **Automation & CI/CD**

    ---

    Integrate performance testing directly into your CI/CD pipelines. Automate the deployment, execution, and teardown of your Locust tests for continuous performance validation.

-   :material-shield-check: **Governance & Resource Management**

    ---

    Maintain control over how resources are deployed and used. Configure resource requests and limits for Locust master and worker pods, and even disable CPU limits for performance-sensitive tests.

-   :material-chart-bar: **Observability & Monitoring**

    ---

    Gain insights into test results and infrastructure usage. The operator supports Prometheus metrics out-of-the-box, allowing you to build rich monitoring dashboards.

-   :material-scale-balance: **Cost Optimization**

    ---

    Optimize cloud costs by deploying resources only when needed and for as long as needed. The operator's automatic cleanup feature ensures that resources are terminated after a test run.

-   :material-layers-triple: **Test Isolation & Parallelism**

    ---

    Run multiple tests in parallel with guaranteed isolation. Each test runs in its own set of resources, preventing any cross-test interference.

-   :material-docker: **Private Image Registry Support**

    ---

    Use images from private registries for your Locust tests. The operator supports `imagePullSecrets` and configurable `imagePullPolicy`.

-   :material-kubernetes: **Advanced Scheduling**

    ---

    Control where your Locust pods are scheduled using Kubernetes affinity and taint tolerations. This allows you to run tests on dedicated nodes or in specific availability zones.

-   :material-apache-kafka: **Kafka & AWS MSK Integration**

    ---

    Seamlessly integrate with Kafka and AWS MSK for performance testing of event-driven architectures. The operator provides out-of-the-box support for authenticated Kafka.

</div>

