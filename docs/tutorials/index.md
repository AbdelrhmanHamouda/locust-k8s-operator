---
title: Tutorials
description: Learn to use the Locust Kubernetes Operator through progressive, hands-on tutorials
tags:
  - tutorials
  - learning
  - guides
---

# Tutorials

Learn to use the Locust Kubernetes Operator through progressive, hands-on tutorials. Each tutorial builds on the previous one, taking you from basic load testing to production-ready deployments.

## Learning Path

Follow these tutorials in order to build your expertise:

### 1. [Your First Load Test](first-load-test.md) (10 minutes)

Learn distributed testing fundamentals by building a realistic e-commerce test from scratch.

**What you'll learn:**
- How Locust master and worker pods communicate
- Writing multi-task test scripts with realistic user behavior
- Configuring test parameters (users, spawn rate, duration)
- Monitoring test progress and interpreting results

### 2. [CI/CD Integration](ci-cd-integration.md) (15 minutes)

Automate load tests in your deployment pipeline.

**What you'll learn:**
- Running tests as part of CI/CD workflows
- Validating performance before production deployments
- Extracting test results for automated decisions
- Handling test failures and rollback scenarios

### 3. [Production Deployment](production-deployment.md) (20 minutes)

Configure production-grade load tests with resource limits, monitoring, and security.

**What you'll learn:**
- Setting resource requests and limits for stable tests
- Integrating with Prometheus for metrics
- Securing test workloads
- Scaling to thousands of users

## Prerequisites

Before starting the tutorials, complete the [Quick Start](../getting_started/index.md) guide to ensure you have:

- A working Kubernetes cluster
- The Locust Kubernetes Operator installed
- kubectl and Helm configured
- Basic understanding of Kubernetes and HTTP

## Need Help?

- **Troubleshooting:** See common issues and solutions in the [FAQ](../reference/faq.md)
- **API Reference:** Complete field documentation in the [API Reference](../reference/api.md)
- **Community:** Join discussions on [GitHub Discussions](https://github.com/locust-k8s-operator/locust-k8s-operator/discussions)
