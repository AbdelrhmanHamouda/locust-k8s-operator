---
title: How does it work
description: A high-level overview of the operator's architecture and workflow.
---

# How does it work

To run a performance test, basic configuration is provided through a simple and intuitive Kubernetes custom resource. Once deployed, the _Operator_ does all the heavy work of creating and scheduling the resources while making sure that all created load generation pods can effectively communicate with each other.

## Architecture Overview

The Locust K8s Operator is built using **Go** with the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) framework, following the standard Kubernetes operator pattern.

```
┌─────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                       │
│  ┌───────────────┐                                          │
│  │   LocustTest  │ ◄── User creates Custom Resource         │
│  │      CR       │                                          │
│  └───────┬───────┘                                          │
│          │ watches                                          │
│          ▼                                                  │
│  ┌───────────────┐     ┌─────────────────────────────────┐  │
│  │   Operator    │────►│  Creates owned resources:       │  │
│  │  Controller   │     │  • Master Service               │  │
│  │               │     │  • Master Job (1 pod)           │  │
│  └───────────────┘     │  • Worker Job (N pods)          │  │
│          │             └─────────────────────────────────┘  │
│          │ updates                                          │
│          ▼                                                  │
│  ┌───────────────┐                                          │
│  │    Status     │ ◄── Phase, Conditions, Worker Count      │
│  │  Subresource  │                                          │
│  └───────────────┘                                          │
└─────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

### Immutable Tests

Tests are **immutable by design**. Updates to a LocustTest CR after creation are ignored. To change test parameters, delete and recreate the CR. This ensures:

- Predictable behavior
- Clean test isolation

### Owner References

All created resources (Jobs, Services) have owner references pointing to the LocustTest CR. This enables:

- Automatic garbage collection on CR deletion
- Clear resource ownership in `kubectl get`
- No orphaned resources

### Status Tracking

The operator maintains rich status information:

```yaml
status:
  phase: Running
  expectedWorkers: 5
  connectedWorkers: 5
  startTime: "2026-01-15T10:00:00Z"
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2026-01-15T10:00:05Z"
      reason: AllWorkersConnected
      message: "All 5 workers connected to master"
```

## Demo

Since a "_Picture Is Worth a Thousand Words_", here is a gif!
![Short demo for how the operator works](assets/images/operatorDemo.gif "Operator Demo")

### Steps performed in demo

-   :material-file-code-outline: Test ConfigMap created in cluster
-   :material-file-document-edit-outline: LocustTest CR deployed into the cluster
-   :material-robot-outline: The _Operator_ creating, configuring and scheduling test resources on CR creation event
-   :material-delete-sweep-outline: The _Operator_ cleaning up test resources after test CR has been removed
