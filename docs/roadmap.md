---
title: Roadmap  
description: Planned features for Locust Kubernetes Operator.
---

# Roadmap

The following is a list of planned features and improvements for the Locust Kubernetes Operator. This list is not exhaustive and may change over time.

-   :material-chart-line: **Enhanced Observability**: Provide out-of-the-box Grafana dashboard examples and more detailed Prometheus configuration guides to make monitoring even easier.

-   :material-send-clock-outline: **Event-Driven Actions**: Integrate with notification systems like Microsoft Teams or Slack to send alerts on test completion, failure, or other significant events.

-   :material-speedometer: **Advanced Benchmarking**: Investigate the feasibility of incorporating external metrics into test results. This would allow for more sophisticated pass/fail criteria, such as assessing the performance of a Kafka-based service by its consumer lag.

-   :material-update: **Dynamic Updates**: Add support for updating a `LocustTest` custom resource while a test is running. This would allow for dynamically adjusting test parameters without restarting the test.

-   :material-web: **Web UI/Dashboard**: Explore the possibility of creating a simple web UI or dashboard for managing and monitoring tests directly through the operator.
