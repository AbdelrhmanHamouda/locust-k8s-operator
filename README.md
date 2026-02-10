<div align="center" style="margin: 2rem 0;">
  <img src="docs/assets/images/logo.gif"
       alt="Locust Kubernetes Operator Logo"
       width="120"
       height="120"
       style="border-radius: 30px; box-shadow: 0 4px 12px rgba(0,0,0,0.15); margin-bottom: 0.1rem;">
</div>

<h1 align="center" style="margin-top: 0.01rem;">Locust Kubernetes Operator</h1>

<p align="center">
Enable performance testing for the modern era!
</p>

<p align="center">
Utilize the full power of <em><a href="https://github.com/locustio/locust">Locust</a></em> in the cloud.
</p>

Docs: [github.io/locust-k8s-operator/](https://abdelrhmanhamouda.github.io/locust-k8s-operator/)

-----------------------------

## v2.0 - Complete Go Rewrite

The operator has been completely rewritten in Go, bringing significant improvements:

| Improvement   | Before (Java)     | After (Go)          |
|---------------|-------------------|---------------------|
| **Memory**    | ~256MB            | ~64MB               |
| **Startup**   | ~60s              | <1s                 |
| **Framework** | Java Operator SDK | Operator SDK / controller-runtime |

### New Features in v2.0

- **Native OpenTelemetry** - Export traces and metrics directly with `--otel` flag
- **Secret & ConfigMap Injection** - Securely inject credentials as env vars or file mounts
- **Volume Mounting** - Mount PVCs, ConfigMaps, Secrets with target filtering (master/worker/both)
- **Separate Resource Specs** - Independent resource configuration for master and worker pods
- **Enhanced Status** - Phase tracking, conditions, and worker connection status

**[Migration Guide](https://abdelrhmanhamouda.github.io/locust-k8s-operator/migration/)** for existing v1 users

-----------------------------

[//]: # (Badges)
[![CI Pipeline][pipeline-status]][pipeline-status-url]
[![Codacy Badge][code-coverage]][code-coverage-url]
[![Codacy Badge][code-quality]][code-quality-url]
[![Documentation][docs]][docs-url]
[![Docker Pulls][docker-pulls]][docker-url]

![](docs/assets/images/run-anywhere.png)

## At a Glance

The Operator is designed to unlock seamless and effortless distributed performance testing in the cloud and enable continuous integration for CI/CD. By design, the entire system is cloud native and focuses on automation and CI practices. One strong feature about the system is its ability to horizontally scale to meet any required performance demands.

## Documentation

All documentation for this project is available at [github.io/locust-k8s-operator/](https://abdelrhmanhamouda.github.io/locust-k8s-operator/).

## Quick Start

### Prerequisites

- **Go 1.24+** for local development
- **Docker** for building container images
- **kubectl** configured for your cluster
- **Helm 3.x** for chart installation
- **Kind** (optional, for local E2E testing)

### Installation

Install the operator using Helm:

```bash
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator
helm install locust-operator locust-k8s-operator/locust-k8s-operator
```

Or from the repository:

```bash
helm install locust-operator charts/locust-k8s-operator/
```

### Development

```bash
# Install CRDs
make install

# Run operator locally (against configured cluster)
make run

# Run tests
make test

# Run E2E tests (requires Kind)
make test-e2e

# Build and push operator image
make docker-build docker-push IMG=<your-registry>/locust-operator:tag
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed development setup.

## Project Status

The project is actively maintained and under continuous development and improvement. If you have any request or want to chat, kindly open a ticket. If you wish to contribute code and/or ideas, kindly check the contribution section.

## Contribute

There's plenty to do, come say hi in [the issues](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues)!

Also check out [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Open source licensed under Apache-2.0 license (see LICENSE file for details).

[//]: # (Pipeline status badge)
[pipeline-status]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/actions/workflows/ci.yaml/badge.svg?branch=master
[pipeline-status-url]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/actions/workflows/ci.yaml

[//]: # (Code coverage badge)
[code-coverage]: https://app.codacy.com/project/badge/Grade/70b76e69dbde4a9ebfd36ad5ccf6de78
[code-coverage-url]: https://app.codacy.com/gh/AbdelrhmanHamouda/locust-k8s-operator/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade

[//]: # (Code quality badge)
[code-quality]: https://app.codacy.com/project/badge/Coverage/70b76e69dbde4a9ebfd36ad5ccf6de78
[code-quality-url]: https://app.codacy.com/gh/AbdelrhmanHamouda/locust-k8s-operator/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage
[//]: # (Documentation badge)
[docs]: https://img.shields.io/badge/Documentation-gh--pages-green
[docs-url]:https://abdelrhmanhamouda.github.io/locust-k8s-operator/
[//]: # (Docker badge)
[docker-url]: https://hub.docker.com/r/lotest/locust-k8s-operator
[docker-pulls]:https://img.shields.io/docker/pulls/lotest/locust-k8s-operator?style=flat&logo=docker&logoColor=green&label=Image%20Pulls&color=green&link=https%3A%2F%2Fhub.docker.com%2Fr%2Flotest%2Flocust-k8s-operator
