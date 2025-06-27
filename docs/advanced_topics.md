---
title: Advanced topics
---

# Advanced topics

Basic configuration is not always enough to satisfy the performance test needs, for example when needing to work with Kafka and MSK. Below is a collection of topics of an advanced nature. This list will be keep growing as the tool matures more and more.

## Kafka & AWS MSK configuration

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

## Dedicated Kubernetes Nodes

To run test resources on dedicated _Kubernetes_ node(s), the _Operator_ support deploying resources with **_Affinity_** and **_Taint Tolerations_**.

### Affinity

This allows generated resources to have specific _Affinity_ options.

!!! Note

    The _Custom Resource Definition Spec_ is designed with modularity and expandability in mind. This means that although a specific set of _Kubernetes Affinity_ options are supported today, extending this support based on need is a streamlined and easy processes. If additonal support is needed, don't hesitate to open a [feature request](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues).

#### Affinity Options

The specification for affinity is defined as follows

=== ":octicons-file-code-16: `affinity-spec.yaml`"

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

=== ":octicons-file-code-16: `node-affinity-example.yaml`"

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

### Taint Tolerations

This allows generated resources to have specific _Taint Tolerations_ options.

#### Toleration Options

The specification for tolerations is defined as follows

=== ":octicons-file-code-16: `taint-tolerations-spec.yaml`"

    ```yaml
    apiVersion: locust.io/v1
    ...
    spec:
      ...
      tolerations:
        - key: <string value>
          operator: <"Exists", "Equal">
          value: <string value> #(1)!
          effect: <"NoSchedule", "PreferNoSchedule", "NoExecute">
        ...
    ```

    1. Optional when `operator` value is set to `Exists`.

=== ":octicons-file-code-16: **Example**"
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

Configuration is done via the `application.yml` file or through Helm values. The following properties are available:

- `locust.operator.resource.pod-mem-request`
- `locust.operator.resource.pod-cpu-request`
- `locust.operator.resource.pod-ephemeral-storage-request`
- `locust.operator.resource.pod-mem-limit`
- `locust.operator.resource.pod-cpu-limit`
- `locust.operator.resource.pod-ephemeral-storage-limit`

### Disabling CPU Limits

In some scenarios, particularly during performance-sensitive tests, you may want to disable CPU limits to prevent throttling. This can be achieved by setting the `pod-cpu-limit` property to a blank string.

=== ":octicons-file-code-16: **Example**"

    ```yaml
    locust:
      operator:
        resource:
          pod-cpu-limit: "" # (1)!
    ```
    
    1.  Setting the CPU limit to an empty string disables it.

!!! Note
    When the CPU limit is disabled, the pod is allowed to use as much CPU as is available on the node. This can be useful for maximizing performance, but it can also lead to resource contention if not managed carefully.

---

## Usage of a private image registry

Images from a private image registry can be used through various methods as described in the [kubernetes documentation](https://kubernetes.io/docs/concepts/containers/images/#using-a-private-registry), one of those methods depends on setting `imagePullSecrets` for pods. This is supported in the operator by simply setting the `imagePullSecrets` option in the deployed custom resource. For example:

```yaml title="locusttest-pull-secret-cr.yaml"
apiVersion: locust.io/v1
...
spec:
  image: ghcr.io/mycompany/locust:latest # (1)!
  imagePullSecrets: # (2)!
    - gcr-secret
  ...
```

1.  Specify which Locust image to use for both master and worker containers.
2.  [Optional] Specify an existing pull secret to use for master and worker pods.

### Image pull policy

Kubernetes uses the image tag and pull policy to control when kubelet attempts to download (pull) a container image. The image pull policy can be defined through the `imagePullPolicy` option, as explained in the [kubernetes documentation](https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy). When using the operator, the `imagePullPolicy` option can be directly configured in the custom resource. For example:

```yaml title="locusttest-pull-policy-cr.yaml"
apiVersion: locust.io/v1
...
spec:
  image: ghcr.io/mycompany/locust:latest # (1)!
  imagePullPolicy: Always # (2)!
  ...
```

1.  Specify which Locust image to use for both master and worker containers.
2.  [Optional] Specify the pull policy to use for containers defined within master and worker containers. Supported options include `Always`, `IfNotPresent` and `Never`.

## Automatic Cleanup for Finished Master and Worker Jobs

Once load tests finish, master and worker jobs remain available in Kubernetes.
You can set up a time-to-live (TTL) value in the operator's Helm chart, so that
kubernetes jobs are eligible for cascading removal once the TTL expires. This means
that Master and Worker jobs and their dependent objects (e.g., pods) will be deleted.

Note that setting up a TTL will not delete `LocustTest` or `ConfigMap` resources.

To set a TTL value, override the key `ttlSecondsAfterFinished` in `values.yaml`:

=== ":octicons-file-code-16: `values.yaml`"
    ```yaml
    ...
    config:
      loadGenerationJobs:
        # Either leave empty or use an empty string to avoid setting this option
        ttlSecondsAfterFinished: 3600 # (1)!
    ...
    ```
    
    1.  Time in seconds to keep the job after it finishes.

You can also use Helm's CLI arguments: `helm install ... --set config.loadGenerationJobs.ttlSecondsAfterFinished=0`.

Read more about the `ttlSecondsAfterFinished` parameter in Kubernetes's [official documentation](https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/).

### Kubernetes Support for `ttlSecondsAfterFinished`

Support for parameter `ttlSecondsAfterFinished` was added in Kubernetes v1.12.
In case you're deploying the locust operator to a Kubernetes cluster that does not
support `ttlSecondsAfterFinished`, you may leave the Helm key empty or use an empty
string. In this case, job definitions will not include the parameter.
