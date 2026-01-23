---
title: Advanced topics
description: Advanced configuration and integration topics for experienced users
tags:
  - advanced
  - configuration
  - kafka
  - aws msk
  - technical
---

# Advanced topics

Basic configuration is not always enough to satisfy the performance test needs, for example when needing to work with Kafka and MSK. Below is a collection of topics of an advanced nature. This list will be keep growing as the tool matures more and more.

## :material-apache-kafka: Kafka & AWS MSK configuration

Generally speaking, the usage of Kafka in a _locustfile_ is identical to how it would be used anywhere else within the cloud context. Thus, no special setup is needed for the purposes of performance testing with the _Operator_.  
That being said, if an organization is using kafka in production, chances are that authenticated kafka is being used. One of the main providers of such managed service is _AWS_ in the form of _MSK_. For that end, the _Operator_ have an _out-of-the-box_ support for MSK.

To enable performance testing with _MSK_, a central/global Kafka user can be created by the "cloud admin" or "the team" responsible for the _Operator_ deployment within the organization. The _Operator_ can then be easily configured to inject the configuration of that user as environment variables in all generated resources. Those variables can be used by the test to establish authentication with the kafka broker.

| Variable Name                    | Description                                                                      |
|:---------------------------------|:---------------------------------------------------------------------------------|
| `KAFKA_BOOTSTRAP_SERVERS`        | Kafka bootstrap servers                                                          |
| `KAFKA_SECURITY_ENABLED`         | -                                                                                |
| `KAFKA_SECURITY_PROTOCOL_CONFIG` | Security protocol. Options: `PLAINTEXT`, `SASL_PLAINTEXT`, `SASL_SSL`, `SSL`     |
| `KAFKA_SASL_MECHANISM`           | Authentication mechanism. Options: `PLAINTEXT`, `SCRAM-SHA-256`, `SCRAM-SHA-512` |
| `KAFKA_USERNAME`                 | The username used to authenticate Kafka clients with the Kafka server            |
| `KAFKA_PASSWORD`                 | The password used to authenticate Kafka clients with the Kafka server            |

---

## :material-server-network: Dedicated Kubernetes Nodes

To run test resources on dedicated _Kubernetes_ node(s), the _Operator_ support deploying resources with **_Affinity_** and **_Taint Tolerations_**.

### :material-vector-link: Affinity

This allows generated resources to have specific _Affinity_ options.

!!! Note

    The _Custom Resource Definition Spec_ is designed with modularity and expandability in mind. This means that although a specific set of _Kubernetes Affinity_ options are supported today, extending this support based on need is a streamlined and easy processes. If additonal support is needed, don't hesitate to open a [feature request](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues).

#### Affinity Options

The specification for affinity is defined as follows

=== "v2 API"

    ```yaml
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: my-test
    spec:
      image: locustio/locust:2.20.0
      master:
        command: "--locustfile /lotest/src/test.py --host https://example.com"
      worker:
        command: "--locustfile /lotest/src/test.py"
        replicas: 3
      scheduling:
        affinity:
          nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              nodeSelectorTerms:
                - matchExpressions:
                    - key: <label-key>
                      operator: In
                      values:
                        - <label-value>
    ```

=== "v1 API (Deprecated)"

    ```yaml
    apiVersion: locust.io/v1
    ...
    spec:
    ...
    affinity:
        nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution
                <label-key>: <label-value>
                ...
    ...
    ```

##### Node Affinity

This optional section causes generated pods to declare specific _Node Affinity_ so _Kubernetes scheduler_ becomes aware of this requirement.

The implementation from the _Custom Resource_ perspective is strongly influenced by Kubernetes native definition
of [node affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity). However, the implementation is
on purpose slightly simplified in order to allow users to have easier time working with affinity.

The `nodeAffinity` section supports declaring node affinity under `requiredDuringSchedulingIgnoredDuringExecution`. Meaning that any
declared affinity labels **must** be present on nodes in order for resources to be deployed on them.

**Example**:

In the below example, generated pods will declare 3 **required** labels (keys and values) to be present on nodes before they are scheduled.

=== "v2 API"

    ```yaml
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: affinity-example
    spec:
      image: locustio/locust:2.20.0
      master:
        command: "--locustfile /lotest/src/test.py --host https://example.com"
      worker:
        command: "--locustfile /lotest/src/test.py"
        replicas: 5
      scheduling:
        affinity:
          nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              nodeSelectorTerms:
                - matchExpressions:
                    - key: nodeAffinityLabel1
                      operator: In
                      values:
                        - locust-cloud-tests
                    - key: nodeAffinityLabel2
                      operator: In
                      values:
                        - performance-nodes
                    - key: nodeAffinityLabel3
                      operator: In
                      values:
                        - high-memory
    ```

=== "v1 API (Deprecated)"

    ```yaml
    apiVersion: locust.io/v1
    ...
    spec:
        ...
        affinity:
            nodeAffinity:
              requiredDuringSchedulingIgnoredDuringExecution:
                nodeAffinityLabel1: locust-cloud-tests
                nodeAffinityLabel2: performance-nodes
                nodeAffinityLabel3: high-memory
                ...
        ...
    ```

### :material-vector-difference: Taint Tolerations

This allows generated resources to have specific _Taint Tolerations_ options.

#### Toleration Options

The specification for tolerations is defined as follows

=== "v2 API"

    ```yaml
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: my-test
    spec:
      image: locustio/locust:2.20.0
      master:
        command: "--locustfile /lotest/src/test.py --host https://example.com"
      worker:
        command: "--locustfile /lotest/src/test.py"
        replicas: 3
      scheduling:
        tolerations:
          - key: <string value>
            operator: <"Exists", "Equal">
            value: <string value> #(1)!
            effect: <"NoSchedule", "PreferNoSchedule", "NoExecute">
    ```

    1. Optional when `operator` value is set to `Exists`.

=== "v2 API Example"

    ```yaml
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: toleration-example
    spec:
      image: locustio/locust:2.20.0
      master:
        command: "--locustfile /lotest/src/test.py --host https://example.com"
      worker:
        command: "--locustfile /lotest/src/test.py"
        replicas: 3
      scheduling:
        tolerations:
          - key: taint-A
            operator: Equal
            value: ssd
            effect: NoSchedule
          - key: taint-B
            operator: Exists
            effect: NoExecute
    ```

=== "v1 API (Deprecated)"

    ```yaml
    apiVersion: locust.io/v1
    ...
    spec:
      ...
      tolerations:
        - key: taint-A
          operator: Equal
          value: ssd
          effect: NoSchedule
        - key: taint-B
          operator: Exists
          effect: NoExecute
    ```


## Resource Management

The operator allows for fine-grained control over the resource requests and limits for the Locust master and worker pods. This is useful for ensuring that your load tests have the resources they need, and for preventing them from consuming too many resources on your cluster.

### Global Defaults via Helm

Configuration is done through Helm values. The following properties are available:

- `locustPods.resources.requests.cpu`
- `locustPods.resources.requests.memory`
- `locustPods.resources.limits.cpu`
- `locustPods.resources.limits.memory`

These defaults apply to all Locust pods unless overridden in individual CRs.

### Per-CR Resource Configuration (v2 API)

!!! info "New in v2.0"
    The v2 API allows you to configure resources independently for master and worker pods.

You can specify resources directly in your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: resource-example
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
    resources:
      requests:
        memory: "256Mi"
        cpu: "100m"
      limits:
        memory: "512Mi"
        cpu: "500m"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 10
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "1Gi"
        cpu: "1000m"
```

[:octicons-arrow-right-24: Learn more about separate resource specs](advanced_topics.md#separate-resource-specs)

### :material-cpu-64-bit: Disabling CPU Limits

In some scenarios, particularly during performance-sensitive tests, you may want to disable CPU limits to prevent throttling.

=== "Global (Helm Values)"

    ```yaml
    locustPods:
      resources:
        limits:
          cpu: "" # (1)!
    ```
    
    1.  Setting the CPU limit to an empty string disables it globally.

=== "Per-CR (v2 API)"

    ```yaml
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: no-cpu-limit-test
    spec:
      image: locustio/locust:2.20.0
      master:
        command: "--locustfile /lotest/src/test.py --host https://example.com"
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            # No CPU limit specified
      worker:
        command: "--locustfile /lotest/src/test.py"
        replicas: 10
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            # No CPU limit specified
    ```

!!! Note
    When the CPU limit is disabled, the pod is allowed to use as much CPU as is available on the node. This can be useful for maximizing performance, but it can also lead to resource contention if not managed carefully.

---

## :material-docker: Usage of a private image registry

Images from a private image registry can be used through various methods as described in the [kubernetes documentation](https://kubernetes.io/docs/concepts/containers/images/#using-a-private-registry), one of those methods depends on setting `imagePullSecrets` for pods. This is supported in the operator by simply setting the `imagePullSecrets` option in the deployed custom resource. For example:

=== "v2 API"

    ```yaml title="locusttest-pull-secret-cr.yaml"
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: private-registry-test
    spec:
      image: ghcr.io/mycompany/locust:latest # (1)!
      imagePullSecrets: # (2)!
        - gcr-secret
      master:
        command: "--locustfile /lotest/src/test.py --host https://example.com"
      worker:
        command: "--locustfile /lotest/src/test.py"
        replicas: 3
    ```

    1.  Specify which Locust image to use for both master and worker containers.
    2.  [Optional] Specify an existing pull secret to use for master and worker pods.

=== "v1 API (Deprecated)"

    ```yaml
    apiVersion: locust.io/v1
    ...
    spec:
      image: ghcr.io/mycompany/locust:latest
      imagePullSecrets:
        - gcr-secret
      ...
    ```

### :material-sync: Image pull policy

Kubernetes uses the image tag and pull policy to control when kubelet attempts to download (pull) a container image. The image pull policy can be defined through the `imagePullPolicy` option, as explained in the [kubernetes documentation](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy). When using the operator, the `imagePullPolicy` option can be directly configured in the custom resource. For example:

=== "v2 API"

    ```yaml title="locusttest-pull-policy-cr.yaml"
    apiVersion: locust.io/v2
    kind: LocustTest
    metadata:
      name: pull-policy-test
    spec:
      image: ghcr.io/mycompany/locust:latest # (1)!
      imagePullPolicy: Always # (2)!
      master:
        command: "--locustfile /lotest/src/test.py --host https://example.com"
      worker:
        command: "--locustfile /lotest/src/test.py"
        replicas: 3
    ```

    1.  Specify which Locust image to use for both master and worker containers.
    2.  [Optional] Specify the pull policy to use for containers. Supported options: `Always`, `IfNotPresent`, `Never`.

=== "v1 API (Deprecated)"

    ```yaml
    apiVersion: locust.io/v1
    ...
    spec:
      image: ghcr.io/mycompany/locust:latest
      imagePullPolicy: Always
      ...
    ```

## :material-auto-fix: Automatic Cleanup for Finished Master and Worker Jobs

Once load tests finish, master and worker jobs remain available in Kubernetes.
You can set up a time-to-live (TTL) value in the operator's Helm chart, so that
kubernetes jobs are eligible for cascading removal once the TTL expires. This means
that Master and Worker jobs and their dependent objects (e.g., pods) will be deleted.

Note that setting up a TTL will not delete `LocustTest` or `ConfigMap` resources.

To set a TTL value, override the key `ttlSecondsAfterFinished` in `values.yaml`:

=== ":octicons-file-code-16: `values.yaml`"
    ```yaml
    locustPods:
      # Either leave empty or use an empty string to avoid setting this option
      ttlSecondsAfterFinished: 3600 # (1)!
    ```
    
    1.  Time in seconds to keep the job after it finishes.

You can also use Helm's CLI arguments: `helm install ... --set locustPods.ttlSecondsAfterFinished=0`.

!!! note "Backward Compatibility"
    The old path `config.loadGenerationJobs.ttlSecondsAfterFinished` is still supported via helper functions in the Helm chart.

Read more about the `ttlSecondsAfterFinished` parameter in Kubernetes's [official documentation](https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/).

### Kubernetes Support for `ttlSecondsAfterFinished`

Support for parameter `ttlSecondsAfterFinished` was added in Kubernetes v1.12.
In case you're deploying the locust operator to a Kubernetes cluster that does not
support `ttlSecondsAfterFinished`, you may leave the Helm key empty or use an empty
string. In this case, job definitions will not include the parameter.

---

## :material-chart-timeline: OpenTelemetry Integration {: #opentelemetry }

!!! info "New in v2.0"
    This feature is only available in the v2 API.

Locust 2.x supports native OpenTelemetry for exporting traces and metrics. The operator can configure this automatically, eliminating the need for the metrics exporter sidecar.

### Enabling OpenTelemetry

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: otel-enabled-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: "grpc"  # or "http"
      insecure: false
      extraEnvVars:
        - name: OTEL_SERVICE_NAME
          value: "my-load-test"
        - name: OTEL_RESOURCE_ATTRIBUTES
          value: "environment=staging,team=platform"
```

### OTel Environment Variables

When OpenTelemetry is enabled, the operator injects the following environment variables:

| Variable | Description |
|----------|-------------|
| `OTEL_TRACES_EXPORTER` | Set to `otlp` |
| `OTEL_METRICS_EXPORTER` | Set to `otlp` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Your configured endpoint |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `grpc` or `http` |
| `OTEL_EXPORTER_OTLP_INSECURE` | Only set if `insecure: true` |

### OTel vs Metrics Sidecar

| Aspect | OpenTelemetry | Metrics Sidecar |
|--------|---------------|-----------------|
| Setup complexity | Low | Low |
| Traces | :white_check_mark: Yes | :x: No |
| Metrics | :white_check_mark: Yes | :white_check_mark: Yes |
| Additional containers | None | 1 sidecar |
| Recommended for | New deployments | Legacy compatibility |

When OpenTelemetry is enabled:

- The `--otel` flag is added to Locust commands
- The metrics exporter sidecar is **not** deployed
- The metrics port is excluded from the Service

---

## :material-key: Environment & Secret Injection {: #environment-injection }

!!! info "New in v2.0"
    This feature is only available in the v2 API.

Inject configuration and credentials into Locust pods without hardcoding them in test files.

### ConfigMap Environment Variables

Inject all keys from a ConfigMap as environment variables:

```yaml
spec:
  env:
    configMapRefs:
      - name: app-config
        prefix: "APP_"  # Results in APP_KEY1, APP_KEY2, etc.
```

### Secret Environment Variables

Inject all keys from a Secret as environment variables:

```yaml
spec:
  env:
    secretRefs:
      - name: api-credentials
        prefix: ""  # No prefix, use key names directly
```

### Individual Variables

Define individual environment variables with values or references:

```yaml
spec:
  env:
    variables:
      - name: TARGET_HOST
        value: "https://api.example.com"
      - name: API_TOKEN
        valueFrom:
          secretKeyRef:
            name: api-secret
            key: token
      - name: CONFIG_VALUE
        valueFrom:
          configMapKeyRef:
            name: app-config
            key: some-key
```

### Secret File Mounts

Mount secrets as files in the container:

```yaml
spec:
  env:
    secretMounts:
      - name: tls-certs
        mountPath: /etc/locust/certs
        readOnly: true
```

### Reserved Paths

The following paths are reserved and cannot be used for secret mounts:

| Path | Purpose |
|------|---------|
| `/lotest/src/` | Test script mount point (default) |
| `/opt/locust/lib` | Library mount point (default) |

!!! note
    If you customize `testFiles.srcMountPath` or `testFiles.libMountPath`, those custom paths become reserved instead.

---

## :material-folder-multiple: Volume Mounting {: #volume-mounting }

!!! info "New in v2.0"
    This feature is only available in the v2 API.

Mount arbitrary volumes to Locust pods for test data, certificates, or configuration files.

### Basic Volume Mount

```yaml
spec:
  volumes:
    - name: test-data
      persistentVolumeClaim:
        claimName: test-data-pvc
  volumeMounts:
    - name: test-data
      mountPath: /data
      target: both  # master, worker, or both
```

### Target Filtering

Control which pods receive the volume mount:

| Target | Master | Worker |
|--------|--------|--------|
| `master` | :white_check_mark: | :x: |
| `worker` | :x: | :white_check_mark: |
| `both` (default) | :white_check_mark: | :white_check_mark: |

### Supported Volume Types

You can use any Kubernetes volume type:

=== "PersistentVolumeClaim"

    ```yaml
    volumes:
      - name: test-data
        persistentVolumeClaim:
          claimName: my-pvc
    ```

=== "ConfigMap"

    ```yaml
    volumes:
      - name: config-files
        configMap:
          name: my-configmap
    ```

=== "Secret"

    ```yaml
    volumes:
      - name: certs
        secret:
          secretName: tls-secret
    ```

=== "EmptyDir"

    ```yaml
    volumes:
      - name: cache
        emptyDir: {}
    ```

### Reserved Volume Names

The following volume names are reserved:

| Pattern | Purpose |
|---------|---------|
| `<crName>-master` | Master ConfigMap volume |
| `<crName>-worker` | Worker ConfigMap volume |
| `locust-lib` | Library ConfigMap volume |
| `secret-*` | Secret volumes from `env.secretMounts` |

---

## :material-tune-vertical: Separate Resource Specs {: #separate-resource-specs }

!!! info "New in v2.0"
    This feature is only available in the v2 API.

Configure resources independently for master and worker pods, allowing you to optimize each component based on its specific needs.

### Independent Resource Configuration

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: optimized-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
    resources:
      requests:
        memory: "256Mi"
        cpu: "100m"
      limits:
        memory: "512Mi"
        cpu: "500m"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 10
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
      limits:
        memory: "1Gi"
        cpu: "1000m"
```

### Use Cases

- **Master pod**: Lower resources since it primarily coordinates workers
- **Worker pods**: Higher resources for actual load generation
- **Memory-intensive tests**: Increase memory limits for workers
- **CPU-intensive tests**: Increase CPU limits or remove limits entirely

### Fallback to Operator Defaults

If `resources` is not specified in the CR, the operator uses default values from its configuration (set via Helm values).
