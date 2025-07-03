# Pull Request Process

This document outlines the process for submitting pull requests to the Locust K8s Operator project. Following these guidelines helps maintain code quality and ensures a smooth review process.

## Before Creating a Pull Request

1. **Discuss Changes First**: Before making significant changes, please discuss the proposed changes via an issue in the GitHub repository.

2. **Follow Coding Conventions**: Ensure your code follows the project's coding standards and conventions.

3. **Write Tests**: All new features or bug fixes should be covered by appropriate tests. See the [integration testing guide](integration-testing.md) for details on integration testing.

## Pull Request Workflow

1. **Fork and Clone**: Fork the repository and clone it locally.

2. **Create a Branch**: Create a branch for your changes using a descriptive name:
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make Your Changes**: Implement your changes, following the project's coding standards.

4. **Commit Your Changes**: Use the [Conventional Commits](https://www.conventionalcommits.org/) standard for commit messages. This is important as the commit messages directly influence the content of the CHANGELOG.md and version increments.
   
   Examples of good commit messages:
   ```
   feat: add support for Locust worker autoscaling
   fix: correct container resource allocation
   docs: update installation instructions
   ```

5. **Run Tests Locally**: Run both unit and integration tests to ensure your changes don't break existing functionality:
   ```bash
   # Run unit tests
   ./gradlew test
   
   # Run integration tests
   ./scripts/run-integration-test.sh
   ```

6. **Submit Your Pull Request**: Push your branch to your fork and submit a pull request to the main repository.

## Pull Request Requirements

1. **Clean Build Dependencies**: Ensure any install or build dependencies are removed before the final build.

2. **Documentation**: Update the documentation with details of changes to interfaces, configuration options, or other important aspects.

3. **Commit Messages**: Ensure commit messages follow the Conventional Commits standard. This is critical for automated changelog generation and semantic versioning.

4. **Tests**:
   - Write clean and well-structured tests.
   - Ensure your changes don't cause regressions.
   - All changes (within reason) should be covered by tests.
   - Update existing tests if your changes represent breaking changes.

## Review Process

1. **Initial Review**: Maintainers will review your PR to ensure it meets the project's requirements.

2. **CI Checks**: The CI system will run tests and other checks against your PR. Make sure these pass.

3. **Feedback**: Maintainers may request changes or improvements to your PR.

4. **Merge**: Once approved, a maintainer will merge your PR.

## After Your PR is Merged

1. **Update Your Fork**: Keep your fork up to date with the main repository.

2. **Celebrate**: Thank you for contributing to the Locust K8s Operator project! Your efforts help make the project better for everyone.

---

Remember that this is a collaborative open-source project. Constructive feedback and discussions are welcome, and all interactions should adhere to the project's Code of Conduct.
