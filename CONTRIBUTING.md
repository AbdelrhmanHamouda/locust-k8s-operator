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