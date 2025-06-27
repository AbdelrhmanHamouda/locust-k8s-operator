---
title: Locust Kubernetes Operator
description: Enable performance testing for the modern era! Utilize the full power of Locust in the cloud.
---

# Locust Kubernetes Operator

Enable performance testing for the modern era!

Utilize the full power of _[Locust](https://github.com/locustio/locust)_ in the cloud.

-----------------------------

[//]: # (Badges)
[![CI Pipeline][pipeline-status]][pipeline-status-url]
[![Codacy Badge][code-coverage]][code-coverage-url]
[![Codacy Badge][code-quality]][code-quality-url]

<div style="text-align: center;">
<img  src="assets/images/undraw_real_time_analytics_cropped.svg" height="500" width="500" alt=""/>
</div>

## At a glance

The _Operator_ is designed to unlock seamless & effortless distributed performance testing in the **_cloud_** and enable **_continues
integration for CI / CD_**. By design, the entire system is cloud native and focuses on automation and CI practices. One strong feature
about the system is its ability to **horizontally scale** to meet any required performance demands.

### What does it offer

<div class="grid cards" markdown>

-   [:material-cloud-check: **Cloud Native**](#)

    ---

    Leverage the full power of Kubernetes and cloud-native technologies for distributed performance testing.

-   [:material-robot-happy: **Automation & CI**](#)

    ---

    Integrate performance testing directly into your CI/CD pipelines for continuous validation.

-   [:material-shield-check: **Governance**](#)

    ---

    Maintain control over how resources are deployed and used in the cloud.

-   [:material-chart-bar: **Observability**](#)

    ---

    Gain insights into test results and infrastructure usage with built-in observability features.

</div>

[Check out the full list of features!](features.md)

![Operator feature set](assets/images/operator-feature-set.png "Operator feature set")

### Whom is it for

It is built for...

![Whom is the operator built for](assets/images/built-for.png "Built for")

### Where can it run

Due to its design, the _Operator_ can be deployed on any kubernetes cluster. Meaning that it is possible to have a full cloud native
performance testing system anywhere in a matter of seconds.

### Limits

The only real limit to this approach is the amount of cluster resources a given team or an organization is willing to dedicate to
performance testing.



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