# Locust Kubernetes Operator (Go)

> **Note:** This is the Go-based rewrite of the [Locust Kubernetes Operator](https://github.com/AbdelrhmanHamouda/locust-k8s-operator).

Enable performance testing for the modern era! Utilize the full power of [Locust](https://github.com/locustio/locust) in the cloud.

## Overview

The Operator is designed to unlock seamless & effortless distributed performance testing in the **cloud** and enable **continuous integration for CI/CD**. By design, the entire system is cloud native and focuses on automation and CI practices. One strong feature about the system is its ability to **horizontally scale** to meet any required performance demands.

## Documentation

For comprehensive documentation, visit [github.io/locust-k8s-operator/](https://abdelrhmanhamouda.github.io/locust-k8s-operator/).

## Development

### Prerequisites

- Go 1.24+
- Docker 17.03+
- kubectl v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster

### Quick Start

```sh
# Install CRDs
make install

# Run the controller locally
make run

# Apply sample CR
kubectl apply -k config/samples/
```

### Build & Deploy

```sh
# Build and push image
make docker-build docker-push IMG=<registry>/locust-k8s-operator:tag

# Deploy to cluster
make deploy IMG=<registry>/locust-k8s-operator:tag
```

### Cleanup

```sh
kubectl delete -k config/samples/  # Delete CRs
make uninstall                      # Delete CRDs
make undeploy                       # Remove controller
```

## Contributing

There's plenty to do, come say hi in [the issues](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues)! 👋

Also check out the [CONTRIBUTING.MD](../CONTRIBUTING.md)

Run `make help` for all available make targets.

## License

Open source licensed under Apache-2.0 license (see [LICENSE](../LICENSE) file for details).

