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
2. Install  [pre-commit](https://pre-commit.com/) and run the below commands to add and register needed git hooks
    1. Run `pre-commit install --install-hooks`
    2. Run `pre-commit install --hook-type commit-msg`

</details>

<details>
<summary>Developing</summary>

- This project follows the [Conventional Commits](https://www.conventionalcommits.org/) standard to
  automate [Semantic Versioning](https://semver.org/) and [Keep A Changelog](https://keepachangelog.com/)
  with [Commitizen](https://github.com/commitizen-tools/commitizen).

</details>

### Local Testing with Minikube and Helm

For local development and testing, you can use Minikube to create a local Kubernetes cluster. This allows you to test the operator and your changes in an environment that closely resembles a production setup.

#### Prerequisites

- [Minikube](https://minikube.sigs.k8s.io/docs/start/)
- [Helm](https://helm.sh/docs/intro/install/)

#### Steps

1.  **Start Minikube**

    Start a local Kubernetes cluster using Minikube:

    ```bash
    minikube start
    ```

2.  **Build and Load the Docker Image**

    If you've made changes to the operator's source code, you'll need to build a new Docker image and load it into your Minikube cluster. This project uses the Jib Gradle plugin to build images directly, so you don't need a `Dockerfile`.

    First, build the image to your local Docker daemon:
    ```bash
    ./gradlew jibDockerBuild
    ```

    Next, load the image into Minikube's internal registry:
    ```bash
    minikube image load locust-k8s-operator:latest
    ```

3.  **Package the Helm Chart**

    Package the Helm chart to create a distributable `.tgz` file.

    ```bash
    helm package ./charts/locust-k8s-operator
    ```

4.  **Install the Operator with Helm**

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