---
title: Getting Started
description: How to get started using Locust Kubernetes operator
tags:
  - tutorial
  - getting started
  - setup
  - guide
  - quickstart
---

# Getting started

Only few simple steps are needed to get a test up and running in the cluster. The following is a step-by-step guide on how to achieve this.

### :material-language-python: Step 1: Write a valid Locust test script

For this example, we will be using the following script

```python title="demo_test.py"
from locust import HttpUser, task

class User(HttpUser): # (1)!
    @task #(2)!
    def get_employees(self) -> None:
        """Get a list of employees."""
        self.client.get("/api/v1/employees") #(3)!
```

1.  Class representing `users` that will be simulated by Locust.
2.  One or more `task` that each simulated `user` will be performing.
3.  HTTP call to a specific endpoint.

!!! note

    To be able to run performance tests effectivly, an understanding of _Locust_ which is the underline load generation tool is required. All tests must be valid _locust_ tests.

    _Locust_ provide a very good and detail rich documentation that can be [found here](https://docs.locust.io/en/stable/quickstart.html).

### :material-file-document-outline: Step 2: Write a valid custom resource for _LocustTest_ CRD

A simple _custom resource_ for the previous test can be something like the following example;

> To streamline this step, [_intensive-brew_](https://abdelrhmanhamouda.github.io/intensive-brew/) should be used. It is a simple cli tool that converts a declarative yaml into a compatible LocustTest kubernetes custom resource.

```yaml title="locusttest-cr.yaml"
apiVersion: locust.io/v1 # (1)!
kind: LocustTest # (2)!
metadata:
  name: demo.test # (3)!
spec:
  image: locustio/locust:latest # (4)!
  masterCommandSeed: # (5)!
    --locustfile /lotest/src/demo_test.py
    --host https://dummy.restapiexample.com
    --users 100
    --spawn-rate 3
    --run-time 3m
  workerCommandSeed: --locustfile /lotest/src/demo_test.py # (6)!
  workerReplicas: 3 # (7)!
  configMap: demo-test-map # (8)!
  libConfigMap: demo-lib-map # (9)!
```

1.  API version based on the deployed _LocustTest_ CRD.
2.  Resource kind.
3.  The name field used by the operator to infer the names of test generated resources. While this value is insignificant to the Operator
    itself, it is important to keep a good convention here since it helps in tracking resources across the cluster when needed.
4.  Image to use for the load generation pods
5.  Seed command for the _master_ node. The _Operator_ will append to this seed command/s all operational parameters needed for the _master_
    to perform its job e.g. ports, rebalancing settings, timeouts, etc...
6.  Seed command for the _worker_ node. The _Operator_ will append to this seed command/s all operational parameters needed for the _worker_
    to perform its job e.g. ports, master node url, master node ports, etc...
7.  The amount of _worker_ nodes to spawn in the cluster.
8.  [Optional] Name of _configMap_ to mount into the pod
9.  [Optional] Name of _configMap_ containing lib directory files to mount at `/opt/locust/lib`

#### Other options

##### Labels and annotations

You can add labels and annotations to generated Pods. For example:

```yaml title="locusttest-cr.yaml"
apiVersion: locust.io/v1
...
spec:
  image: locustio/locust:latest
  labels: # (1)!
    master:
      locust.io/role: "master"
      myapp.com/testId: "abc-123"
      myapp.com/tenantId: "xyz-789"
    worker:
      locust.io/role: "worker"
  annotations: # (2)!
    master:
      myapp.com/threads: "1000"
      myapp.com/version: "2.1.0"
    worker:
      myapp.com/version: "2.1.0"
  ...
```

1.  [Optional] Labels are attached to both master and worker pods. They can later be used to identify pods belonging to a particular execution context. This is useful, for example, when tests are deployed programmatically. A launcher application can query the Kubernetes API for specific resources.
2.  [Optional] Annotations too are attached to master and worker pods. They can be used to include additional context about a test. For example, configuration parameters of the software system being tested.

Both labels and annotations can be added to the Prometheus configuration, so that metrics are associated with the appropriate information, such as the test and tenant ids. You can read more about this in the [Prometheus documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config) site.

### :material-rocket-launch-outline: Step 3: Deploy _Locust k8s Operator_ in the cluster.

The recommended way to install the _Operator_ is by using the official HELM chart. Documentation on how to perform that
is [available here](helm_deploy.md).

### :material-cogs: Step 4: Deploy test as a configMap

For the purposes of this example, the `demo_test.py` test previously demonstrated will be deployed into the cluster as a _configMap_ that
the _Operator_ will mount to the load generation pods.  
To deploy the test as a configMap, run the bellow command following this
template `kubectl create configmap <configMap-name> --from-file <your_test.py>`:

- `kubectl create configmap demo-test-map --from-file demo_test.py`

#### :material-folder-multiple: Step 4.1: Deploy lib files as a configMap (Optional)

!!! info "What are lib files and why use this feature?"
    **Lib files** are Python modules and libraries that your Locust tests depend on. When your tests require custom helper functions, utilities, or shared code that should be available across multiple test files, this feature allows you to package and deploy them alongside your tests.

##### How it works

The Locust Kubernetes Operator provides a mechanism to deploy your custom Python libraries as a ConfigMap, which will then be mounted to the `/opt/locust/lib` directory inside all Locust pods (both master and worker). This allows your test scripts to import and use these libraries.

For example, if you have the following structure:

```
project/
├── my_test.py          # Your main Locust test file
└── lib/
    ├── helpers.py      # Helper functions
    ├── utils.py        # Utility functions
    └── models.py       # Data models
```

Your test might import these libraries like this:

```python
# in my_test.py
from lib.helpers import some_helper_function
from lib.utils import format_data
```

To make these imports work when your test runs in Kubernetes, you need to:

1. Deploy your lib files as a ConfigMap
2. Reference this ConfigMap in your LocustTest custom resource

##### Step-by-step instructions

**1. Create a ConfigMap from your lib directory:**

You can deploy all library files from a directory:

```bash
# Deploy all files from the lib/ directory as a ConfigMap
kubectl create configmap demo-lib-map --from-file=lib/
```

Alternatively, you can create it from individual files:

```bash
# Deploy specific files as a ConfigMap
kubectl create configmap demo-lib-map --from-file=lib/helpers.py --from-file=lib/utils.py
```

**2. Reference the lib ConfigMap in your LocustTest custom resource:**

```yaml
apiVersion: locust.io/v1
kind: LocustTest
metadata:
  name: example-locusttest
spec:
  masterConfig:
    replicas: 1
  workerConfig:
    replicas: 2
  configMap: demo-test-map    # Your test script ConfigMap
  libConfigMap: demo-lib-map   # Your lib files ConfigMap
```

!!! tip "Organizing your code"
    This feature is especially useful when:
    
    1. You have complex test scenarios that benefit from modular code
    2. You want to share code between multiple Locust tests
    3. You need to keep your test scripts clean by separating implementation details

!!! note "Fresh cluster resources"

    Fresh cluster resources are allocated for each running test, meaning that tests **DO NOT** have any cross impact on each other.

### :material-play-circle-outline: Step 5: Start the test by deploying the _LocustTest_ custom resource.

Deploying a _custom resource_, signals to the _Operator_ the desire to start a test and thus the _Operator_ starts creating and scheduling
all needed resources.  
To do that, deploy the custom resource following this template `kubectl apply -f <valid_cr>.yaml`:

- `kubectl apply -f locusttest-cr.yaml`

#### Step 5.1: Check cluster for running resources

At this point, it is possible to check the cluster and all required resources will be running based on the passed configuration in the
custom resource.

The Operator will create the following resources in the cluster for each valid custom resource:

- A kubernetes _service_ for the _master_ node so it is reachable by other _worker_ nodes.
- A kubernetes _Job_ to manage the _master_ node.
- A kubernetes _Job_ to manage the _worker_ node.

### :material-delete-outline: Step 6: Clear resources after test run

In order to remove the cluster resources after a test run, simply remove the custom resource and the _Operator_ will react to this event by
cleaning the cluster of all **related** resources.  
To delete a resource, run the below command following this template `kubectl delete -f <valid_cr>.yaml`:

- `kubectl delete -f locusttest-cr.yaml`
