# Local Development Guide

This guide describes the setup and workflow for local development on the Locust K8s Operator project. It's intended for developers who want to contribute code changes.

## Development Setup

### Prerequisites

- **Go 1.24+**: Required for building the operator
- **Docker**: Running Docker daemon for building images
- **kubectl**: Kubernetes CLI for cluster interaction
- **Kind** or **Minikube**: Local Kubernetes cluster for testing
- **Helm 3.x**: For chart packaging and installation

### Initial Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/AbdelrhmanHamouda/locust-k8s-operator.git
   cd locust-k8s-operator
   ```

2. Install dependencies and tools:
   ```bash
   # Download Go dependencies
   make tidy
   
   # Install development tools (controller-gen, envtest, etc.)
   make controller-gen
   make envtest
   make kustomize
   ```

## Development Guidelines

- This project follows the [Conventional Commits](https://www.conventionalcommits.org/) standard to automate [Semantic Versioning](https://semver.org/) and [Keep A Changelog](https://keepachangelog.com/) with [Commitizen](https://github.com/commitizen-tools/commitizen).

- All code should include appropriate tests. See the [integration testing guide](integration-testing.md) for details on the test setup.

## Common Development Commands

The project uses a `Makefile` for common development tasks. Run `make help` to see all available targets.

### Build & Test

```bash
# Build the operator binary
make build

# Run all tests (unit + integration via envtest)
make test

# Run linter
make lint

# Run linter with auto-fix
make lint-fix

# Run all CI checks locally
make ci
```

### Code Generation

```bash
# Generate CRDs, RBAC, and webhook manifests
make manifests

# Generate DeepCopy implementations
make generate

# Format code
make fmt

# Run go vet
make vet
```

### Running Locally

```bash
# Run the operator locally against your current kubeconfig cluster
make run

# Install CRDs into the cluster
make install

# Uninstall CRDs from the cluster
make uninstall
```

## Local Testing with Kind

For local development and testing, Kind (Kubernetes in Docker) is the recommended approach.

### Steps

1. **Create a Kind Cluster**

   ```bash
   kind create cluster --name locust-dev
   ```

2. **Build and Load the Docker Image**

   ```bash
   # Build the Docker image
   make docker-build IMG=locust-k8s-operator:dev
   
   # Load the image into Kind
   kind load docker-image locust-k8s-operator:dev --name locust-dev
   ```

3. **Deploy the Operator**

   Option A: Using kustomize (for development):
   ```bash
   # Deploy CRDs and operator
   make deploy IMG=locust-k8s-operator:dev
   ```

   Option B: Using Helm (for production-like testing):
   ```bash
   # Package the Helm chart
   helm package ../charts/locust-k8s-operator
   
   # Install with local image
   helm install locust-operator locust-k8s-operator-*.tgz \
     --set image.repository=locust-k8s-operator \
     --set image.tag=dev \
     --set image.pullPolicy=IfNotPresent
   ```

4. **Verify the Deployment**

   !!! note "Development vs Production Namespaces"
       The `make deploy` command generates a namespace based on your project name. For production deployments, use the `locust-system` namespace as documented in the [Helm Deployment Guide](helm_deploy.md).

   ```bash
   # Check pods in the generated namespace
   kubectl get pods -A | grep locust

   # Follow operator logs
   kubectl logs -f -n <namespace> deployment/<deployment-name>
   ```

5. **Test with a Sample CR**

   ```bash
   # Create a test ConfigMap with a simple Locust script
   kubectl create configmap locust-test --from-literal=locustfile.py='
   from locust import HttpUser, task
   class TestUser(HttpUser):
       @task
       def hello(self):
           self.client.get("/")
   '
   
   # Apply a sample LocustTest CR
   kubectl apply -f config/samples/locust_v2_locusttest.yaml
   
   # Watch the resources
   kubectl get locusttests,jobs,pods -w
   ```

6. **Cleanup**

   ```bash
   # Remove the operator
   make undeploy
   
   # Delete the Kind cluster
   kind delete cluster --name locust-dev
   ```

## Writing Documentation

All documentation is located under the `docs/` directory. The documentation is hosted on [GitHub Pages](https://abdelrhmanhamouda.github.io/locust-k8s-operator/) and updated automatically with each release. To manage and build the documentation, the project uses [MkDocs](https://www.mkdocs.org/) & [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) framework.

### Preview Documentation Locally

```bash
# Install MkDocs (if not installed)
pip install mkdocs mkdocs-material

# Serve documentation locally
mkdocs serve

# Build documentation
mkdocs build --strict
```

During development, the **_CI_** workflow will build the documentation as part of the validation.
