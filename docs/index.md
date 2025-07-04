---
title: Locust Kubernetes Operator
description: Enable performance testing for the modern era! Utilize the full power of Locust in the cloud with a fully automated, cloud-native approach.
---

# Performance testing that simply works

<div class="tx-hero">
<p>Utilize the full power of <em>Locust</em> in the cloud with a fully automated, cloud-native approach, creating professional and reliable performance tests in minutes.</p>
<div class="tx-hero__image">
  <img src="assets/images/undraw_real_time_analytics_cropped.svg" alt="Locust K8s Operator" width="500" draggable="false">
</div>
<div class="tx-hero__content">
  <a href="getting_started/" class="md-button md-button--primary">
    Get started
  </a>
  <a href="features/" class="md-button">
    Learn more
  </a>
</div>
</div>

<div class="badges-section">
  <div class="badges-container">
    <a href="https://github.com/AbdelrhmanHamouda/locust-k8s-operator/actions/workflows/ci.yaml" target="_blank">
      <img src="https://github.com/AbdelrhmanHamouda/locust-k8s-operator/actions/workflows/ci.yaml/badge.svg?branch=master" alt="CI Pipeline" class="badge">
    </a>
    <a href="https://app.codacy.com/gh/AbdelrhmanHamouda/locust-k8s-operator/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade" target="_blank">
      <img src="https://app.codacy.com/project/badge/Grade/70b76e69dbde4a9ebfd36ad5ccf6de78" alt="Code Quality" class="badge">
    </a>
    <a href="https://app.codacy.com/gh/AbdelrhmanHamouda/locust-k8s-operator/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage" target="_blank">
      <img src="https://app.codacy.com/project/badge/Coverage/70b76e69dbde4a9ebfd36ad5ccf6de78" alt="Code Coverage" class="badge">
    </a>
    <a href="https://hub.docker.com/r/lotest/locust-k8s-operator" target="_blank">
      <img src="https://img.shields.io/docker/pulls/lotest/locust-k8s-operator?style=flat&logo=docker&logoColor=green&label=Image%20Pulls&color=green" alt="Docker Pulls" class="badge">
    </a>
  </div>
</div>



## Build for cloud-native performance testing { .text-center }

The _Operator_ is designed to unlock seamless & effortless distributed performance testing in the **_cloud_** and enable **_continuous integration for CI/CD pipelines_**. By design, the entire system is cloud native and focuses on automation and CI practices. One strong feature
about the system is its ability to **horizontally scale** to meet any required performance demands.

### Key capabilities { .text-center }

<div class="grid cards" markdown>

-   :material-cloud-check:{ .lg .middle } __Cloud Native__

    ---

    Leverage the full power of Kubernetes and cloud-native technologies for distributed performance testing.

    [:octicons-arrow-right-24: Learn more](features.md#cloud-native)

-   :material-robot-happy:{ .lg .middle } __Automation & CI__

    ---

    Integrate performance testing directly into your CI/CD pipelines for continuous validation.

    [:octicons-arrow-right-24: Learn more](features.md#automation)

-   :material-shield-check:{ .lg .middle } __Governance__

    ---

    Maintain control over how resources are deployed and used in the cloud.

    [:octicons-arrow-right-24: Learn more](features.md#governance)

-   :material-chart-bar:{ .lg .middle } __Observability__

    ---

    Gain insights into test results and infrastructure usage with built-in observability features.

    [:octicons-arrow-right-24: Learn more](features.md#observability)

</div>

[Check out the full list of features!](features.md)

![Operator feature set](assets/images/operator-feature-set.png "Operator feature set")

## Designed for teams and organizations { .text-center }

<div class="grid" markdown>

<div markdown>

### Who is it for { .text-center }

It is built for performance engineers, DevOps teams, and organizations looking to integrate performance testing into their CI/CD pipelines.

![Whom is the operator built for](assets/images/built-for.png "Built for")

</div>

<div markdown>

### Universal deployment { .text-center }

Due to its design, the _Operator_ can be deployed on any Kubernetes cluster. This means you can have a full cloud-native
performance testing system anywhere in a matter of seconds.

</div>

<div markdown>

### Scalable resources { .text-center }

The only real limit to this approach is the amount of cluster resources a team or organization is willing to dedicate to
performance testing. Scale up or down based on your needs.

</div>

</div>



[//]: # (Pipeline status badge)
[pipeline-status]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/actions/workflows/ci.yaml/badge.svg?branch=master
[pipeline-status-url]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/actions/workflows/ci.yaml

[//]: # (Code coverage badge)
[code-coverage]: https://app.codacy.com/project/badge/Grade/70b76e69dbde4a9ebfd36ad5ccf6de78
[code-coverage-url]: https://www.codacy.com/gh/AbdelrhmanHamouda/locust-k8s-operator/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=AbdelrhmanHamouda/locust-k8s-operator&amp;utm_campaign=Badge_Grade

[//]: # (Code quality badge)
[code-quality]: https://app.codacy.com/project/badge/Coverage/70b76e69dbde4a9ebfd36ad5ccf6de78
[code-quality-url]: https://www.codacy.com/gh/AbdelrhmanHamouda/locust-k8s-operator/dashboard?utm_source=github.com&utm_medium=referral&utm_content=AbdelrhmanHamouda/locust-k8s-operator&utm_campaign=Badge_Coverage

[//]: # (common urls)
[contributing-url]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/CONTRIBUTING.md
[issues-url]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues
[LocustTest]:https://github.com/AbdelrhmanHamouda/locust-k8s-operator/tree/master/kube/crd/locust-test-crd.yaml
[cr-example]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/tree/master/kube/sample-cr/locust-test-cr.yaml

[//]: # (Docker badge)
[docker-url]: https://hub.docker.com/r/lotest/locust-k8s-operator
[docker-pulls]:https://img.shields.io/docker/pulls/lotest/locust-k8s-operator?style=flat&logo=docker&logoColor=green&label=Image%20Pulls&color=green&link=https%3A%2F%2Fhub.docker.com%2Fr%2Flotest%2Flocust-k8s-operator