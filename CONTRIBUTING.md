# Contributing

So you'd like to contribute? Awesome! Here are some things worth knowing.

## Reporting a bug / requesting a feature / asking a question

Go [open an issue](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues) and I'll probably reply soon.

# Contributing code

### Preface

If you're willing to contribute your ideas and effort to making the _locust-k8s-operator_ better, then that's awesome and I'm truly
grateful. I don't have all the answers and it is important for this project to benefit from diverse perspective and technical expertise.

That being said, please be aware that a lot of thought has gone into the architecture of the project, and whilst I know it's not perfect,
and I am very interested in alternative perspectives, I do have opinions (reasonable I hope) about how certain things should. This
particularly applies to naming and internal APIs. There is a lot to consider in terms of making sure the tool stays simple, flexible, and
performant. So please don't be offended if there is some push back. It is ultimately for the benefit of the tool and nothing more.

### General

When contributing to this repository, please first discuss the change you wish to make via an issue.

Please note that we have a code of conduct and thus you are kindly asked to follow it in all your interactions with the project.

### Developing code

<details open>
<summary>Setup: once per project</summary>

1. Clone this repository.
2. Install prerequisites:
    - [Go 1.24+](https://go.dev/dl/)
    - [Docker](https://docs.docker.com/get-docker/)
    - [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (Kubernetes in Docker)
    - [Helm 3](https://helm.sh/docs/intro/install/)
3. Install development tools:
    ```bash
    make controller-gen envtest kustomize
    ```
4. Install [pre-commit](https://pre-commit.com/) and run:
    1. `pre-commit install --install-hooks`
    2. `pre-commit install --hook-type commit-msg`

</details>

<details>
<summary>Common development commands</summary>

- This project follows the [Conventional Commits](https://www.conventionalcommits.org/) standard to
  automate [Semantic Versioning](https://semver.org/) and [Keep A Changelog](https://keepachangelog.com/)
  with [Commitizen](https://github.com/commitizen-tools/commitizen).

```bash
# Build the operator binary
make build

# Run all tests (unit + integration via envtest)
make test

# Run linter
make lint

# Generate CRDs, RBAC, and webhook manifests
make manifests

# Run all CI checks locally
make ci
```

</details>

### Local Testing with Kind

For local development and testing, [Kind](https://kind.sigs.k8s.io/) (Kubernetes in Docker) is the recommended approach.

#### Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [Helm 3](https://helm.sh/docs/intro/install/)

#### Steps

1.  **Create a Kind Cluster**

    ```bash
    kind create cluster --name locust-dev
    ```

2.  **Build and Load the Docker Image**

    Build a Docker image and load it into Kind:

    ```bash
    # Build the Docker image
    make docker-build IMG=locust-k8s-operator:dev

    # Load the image into Kind
    kind load docker-image locust-k8s-operator:dev --name locust-dev
    ```

3.  **Install the Operator with Helm**

    ```bash
    # Package the Helm chart
    helm package charts/locust-k8s-operator

    # Install with local image
    helm install locust-operator locust-k8s-operator-*.tgz \
      --set image.repository=locust-k8s-operator \
      --set image.tag=dev \
      --set image.pullPolicy=IfNotPresent
    ```

4.  **Verify the Deployment**

    ```bash
    kubectl get pods -A | grep locust
    ```

5.  **Cleanup**

    ```bash
    # Uninstall the operator
    helm uninstall locust-operator

    # Delete the Kind cluster
    kind delete cluster --name locust-dev
    ```

### Writing documentation

All documentation is located under the `docs/` directory. The documentation is hosted
on [gh-pages](https://abdelrhmanhamouda.github.io/locust-k8s-operator/) and updated automatically with each release. To manage and build the
documentation, the project uses [MkDocs] & [Material for MkDocs] framework to manage and build the documentation.

During development, the **_CI_** workflow will build the documentation as part of the validation.

### Pull Request Process

1. Ensure any install or build dependencies are removed before the end layer when doing a build.
2. Update the documentation when needed with the details of the proposed change, this includes any useful information for the end user of
   the tool maintainer / contributor(s).
3. Make sure that the commit messages are aligned with the used standard. This is very important since the commit message directly influence
   the content of the CHANGELOG.md and the version increase.
4. **Tests**
    1. Clean and well written tests are very important.
    2. Changes must not cause any type of regression.
    3. All changes / additions (within reason) must be covered by tests.
    4. If the additions represent a breaking change, existing tests must be updated.

[//]: # (Documentation framework urls)

[MkDocs]: https://www.mkdocs.org/

[Material for MkDocs]: https://squidfunk.github.io/mkdocs-material/