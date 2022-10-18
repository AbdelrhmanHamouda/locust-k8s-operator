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

Fundamentally, the _Operator_ provide the following as part of its core offerings; **cloud native**, **automation & CI**, **governance**,
**Observability**.

**Distributed cloud performance testing**: _[Locust](https://github.com/locustio/locust)_ is a great and very powerful load testing tool. It
is capable of generating a significant amount of load specially when configured correctly. That being said, there is only so much a single
instance and vertical scaling can do. Luckily, _Locust_ has a native out of the box support for distributed mode. This _Locust Kubernetes
Operator_ project leverage this feature and adds systems and functionalities to address challenges and situations that are exclusive to the
cloud context.

**Low barrier of entry**: Utilizing the power of the _Operator_ lowers significantly the barrier of entry to run in the cloud. From an
end-user perspective, running a performance test in the cloud becomes a **single command** operation.

**Test isolation** and **Parallel tests**: By default, the _Operator_ is able to support any number of Parallel test executions with an
absolute guarantee that each test is fully protected from being polluted by the existence of any number of other tests.

**Automation & CI**: By having automation as a core focus point, teams and organizations can build performance testing directly into CI/CD
pipelines. Meaning that every new service, feature or system can be potentially tested and validated for performance SLOs / SLAs.

**Separation of concerns**: By using the _Operator_, _engineering teams_ can focus on building a robust performance test/s and SREs
DevOps teams can focus on managing the resources.

**Governance**: Enable organizations to have governance over what / how resources are deployed and run on the cloud.

**Cloud cost optimization**: Using the _Operator_ enables for a more effective control over the **_cloud cost_**. Since resources are
**only** deployed when needed and **only** for as long as needed, the cost of performance testing is kept to a minimum.

**Observability**: For both engineering teams and cloud admins, the _Operator_ unlocks the ability to build observability & monitoring
dashboards in order to analyse test results during test runtime or retroactively (interesting for teams) and infrastructure usage and
resource monitoring ( interesting for
cloud admins, SREs, etc...).

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