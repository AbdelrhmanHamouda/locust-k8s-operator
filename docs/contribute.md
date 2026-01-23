---
title: Contributing & Development  
description: Information for contributors and developers.
tags:
  - contributing
  - development
  - community
  - open source
  - collaboration
---

# Contributing & Development

## Ways to Contribute

There are several ways you can contribute to the Locust K8s Operator project:

### For Everyone

- **Reporting Issues**: Found a bug or have a feature request? Open an [issue][issues-url] 👋
- **Documentation**: Help improve the documentation by suggesting clarifications or additions
- **Community Support**: Answer questions and help others in the issue tracker

### For Developers

> **Note**: The following sections are intended for developers who want to contribute code to the project. If you're just using the operator, you can skip these sections.

- **Code Contributions**: Implement new features or fix bugs
- **Testing**: Improve test coverage and test in different environments
- **Review**: Review pull requests from other contributors

## Project Status

The project is **_actively_** maintained and is under continuous development and improvement. If you have any request or want to chat, kindly open a ticket. If you wish to contribute code and/or ideas, please review the development documentation below.

## Technology Stack

The operator is built with **Go** using the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) framework. Key technologies:

- **Language**: Go 1.23+
- **Framework**: controller-runtime / Operator SDK
- **Testing**: envtest, Ginkgo, Kind
- **Build**: Make, Docker
- **Deployment**: Helm, Kustomize

## Development Documentation

For developers contributing to the Locust K8s Operator project, we provide detailed documentation on various development aspects:

- [Local Development Guide](local-development.md): Setting up your development environment
- [Testing Guide](integration-testing.md): Running unit, integration, and E2E tests
- [Pull Request Process](pull-request-process.md): Guidelines for submitting code changes
- [How It Works](how_does_it_work.md): Architecture overview

You can also refer to the comprehensive [CONTRIBUTING.MD][contributing-url] file in the GitHub repository for more information.

[//]: # (common urls)
[contributing-url]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/CONTRIBUTING.md
[issues-url]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues