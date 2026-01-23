# Pull Request Process

This document outlines the process for submitting pull requests to the Locust K8s Operator project. Following these guidelines helps maintain code quality and ensures a smooth review process.

## Before Creating a Pull Request

1. **Discuss Changes First**: Before making significant changes, please discuss the proposed changes via an issue in the GitHub repository.

2. **Follow Coding Conventions**: Ensure your code follows the project's coding standards and conventions.

3. **Write Tests**: All new features or bug fixes should be covered by appropriate tests. See the [testing guide](integration-testing.md) for details on the testing setup.

## Pull Request Workflow

1. **Fork and Clone**: Fork the repository and clone it locally.

2. **Create a Branch**: Create a branch for your changes using a descriptive name:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Your Changes**: Implement your changes in the `locust-k8s-operator-go/` directory, following Go coding standards.

4. **Commit Your Changes**: Use the [Conventional Commits](https://www.conventionalcommits.org/) standard for commit messages. This is important as the commit messages directly influence the content of the CHANGELOG.md and version increments.
   
   Examples of good commit messages:
   ```
   feat: add support for OpenTelemetry metrics export
   fix: correct volume mount path validation
   docs: update API reference for v2 fields
   refactor: simplify resource builder functions
   test: add integration tests for env injection
   ```

5. **Run Tests Locally**: Run tests and linting to ensure your changes don't break existing functionality:
   ```bash
   cd locust-k8s-operator
   
   # Run all CI checks (lint + tests)
   make ci
   
   # Or run individually:
   make lint        # Run linter
   make test        # Run unit + integration tests
   make test-e2e    # Run E2E tests (requires Docker)
   ```

6. **Generate Manifests**: If you modified API types, regenerate manifests:
   ```bash
   make generate    # Generate DeepCopy methods
   make manifests   # Generate CRDs, RBAC, webhooks
   ```

7. **Submit Your Pull Request**: Push your branch to your fork and submit a pull request to the main repository.

## Pull Request Requirements

### Code Quality

- [ ] Code follows Go conventions and project style
- [ ] No linting errors (`make lint` passes)
- [ ] All tests pass (`make test` passes)
- [ ] New code has appropriate test coverage (≥80% for new packages)

### Documentation

- [ ] API changes are reflected in `docs/api_reference.md`
- [ ] New features are documented in `docs/features.md` or `docs/advanced_topics.md`
- [ ] Breaking changes are noted in the PR description
- [ ] Helm chart updates include `docs/helm_deploy.md` changes

### Commit Messages

- [ ] Follow Conventional Commits standard
- [ ] Each commit represents a logical unit of change
- [ ] Commit messages are clear and descriptive

### Tests

- [ ] Unit tests for new/modified functions
- [ ] Integration tests for controller behavior changes
- [ ] Existing tests updated if behavior changes
- [ ] No test regressions

## CI Pipeline Checks

The following checks run automatically on each PR:

| Check         | Description                   | Command          |
|---------------|-------------------------------|------------------|
| **Lint**      | golangci-lint static analysis | `make lint`      |
| **Test**      | Unit + integration tests      | `make test`      |
| **Build**     | Binary compilation            | `make build`     |
| **Manifests** | CRD/RBAC generation           | `make manifests` |

All checks must pass before a PR can be merged.

## Review Process

1. **Initial Review**: Maintainers will review your PR to ensure it meets the project's requirements.

2. **CI Checks**: The CI system will run tests and other checks against your PR. Make sure these pass.

3. **Feedback**: Maintainers may request changes or improvements to your PR.

4. **Merge**: Once approved and CI passes, a maintainer will merge your PR.

## After Your PR is Merged

1. **Update Your Fork**: Keep your fork up to date with the main repository:
   ```bash
   git checkout master
   git pull upstream master
   git push origin master
   ```

2. **Celebrate**: Thank you for contributing to the Locust K8s Operator project! Your efforts help make the project better for everyone.

---

Remember that this is a collaborative open-source project. Constructive feedback and discussions are welcome, and all interactions should adhere to the project's Code of Conduct.
