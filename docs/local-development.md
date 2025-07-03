# Local Development Guide

This guide describes the setup and workflow for local development on the Locust K8s Operator project. It's intended for developers who want to contribute code changes.

## Development Setup

<details open>
<summary>Initial Setup</summary>

1. Clone the repository:
   ```bash
   git clone https://github.com/AbdelrhmanHamouda/locust-k8s-operator.git
   cd locust-k8s-operator
   ```

2. Install [pre-commit](https://pre-commit.com/) and set up the git hooks:
   ```bash
   pre-commit install --install-hooks
   pre-commit install --hook-type commit-msg
   ```
</details>

## Development Guidelines

- This project follows the [Conventional Commits](https://www.conventionalcommits.org/) standard to automate [Semantic Versioning](https://semver.org/) and [Keep A Changelog](https://keepachangelog.com/) with [Commitizen](https://github.com/commitizen-tools/commitizen).

- All code should include appropriate tests. See the [integration testing guide](integration-testing.md) for details on the integration test setup.

## Local Testing with Minikube and Helm

For local development and testing, you can use Minikube to create a local Kubernetes cluster. This allows you to test the operator and your changes in an environment that closely resembles a production setup.

### Prerequisites

- [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Helm](https://helm.sh/docs/intro/install/)

### Steps

1. **Start Minikube**

   Start a local Kubernetes cluster using Minikube:

   ```bash
   minikube start
   ```

2. **Build and Load the Docker Image**

   If you've made changes to the operator's source code, you'll need to build a new Docker image and load it into your Minikube cluster. This project uses the Jib Gradle plugin to build images directly, so you don't need a `Dockerfile`.

   First, build the image to your local Docker daemon:
   ```bash
   ./gradlew jibDockerBuild
   ```

   Next, load the image into Minikube's internal registry:
   ```bash
   minikube image load locust-k8s-operator:latest
   ```

3. **Package the Helm Chart**

   Package the Helm chart to create a distributable `.tgz` file.

   ```bash
   helm package ./charts/locust-k8s-operator
   ```

4. **Install the Operator with Helm**

   Install the Helm chart on your Minikube cluster. The command below overrides the default image settings to use the one you just built and loaded.

   You can use a `values.yaml` file to override other settings.

   ```yaml
   # values.yaml (optional)
   # Example: Set resource requests and limits for the operator pod
   config:
     loadGenerationPods:
       resource:
         cpuRequest: 250m
         memRequest: 128Mi
         ephemeralRequest: 300M
         cpuLimit: 1000m
         memLimit: 1024Mi
         ephemeralLimit: 50M
   
   # To leave a resource unbound, Leave the limit empty
   # This is useful when you don't want to set a specific limit.
   # example:
   # config:
   #   loadGenerationPods:
   #     resource:
   #       cpuLimit: ""
   #       memLimit: ""
   #       ephemeralLimit: ""
   ```

   Install the chart using the following command. The `-f values.yaml` flag is optional.

   ```bash
   helm install locust-operator locust-k8s-operator-*.tgz -f values.yaml \
     --set image.repository=locust-k8s-operator \
     --set image.tag=latest \
     --set image.pullPolicy=IfNotPresent
   ```

   This will deploy the operator to your Minikube cluster using the settings defined in your `values.yaml` file.

## Writing Documentation

All documentation is located under the `docs/` directory. The documentation is hosted on [GitHub Pages](https://abdelrhmanhamouda.github.io/locust-k8s-operator/) and updated automatically with each release. To manage and build the documentation, the project uses [MkDocs](https://www.mkdocs.org/) & [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/) framework.

During development, the **_CI_** workflow will build the documentation as part of the validation.
