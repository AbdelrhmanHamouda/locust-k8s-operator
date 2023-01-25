---
title: Advanced topics
---

# Advanced topics

Basic configuration is not always enough to satisfy the performance test needs, for example when needing to work with Kafka and MSK. Below is a  collection of topics of an advanced nature. This list will be keep growing as the tool matures more and more. 

## Kafka & AWS MSK configuration

Generally speaking, the usage of Kafka in a _locustfile_ is identical to how it would be used anywhere else within the cloud context. Thus, no special setup is needed for the purposes of performance testing with the _Operator_.  
That being said, if an organization is using kafka in production, chances are that authenticated kafka is being used. One of the main providers of such managed service is _AWS_ in the form of _MSK_.  For that end, the _Operator_ have an _out-of-the-box_ support for MSK. 

To enable performance testing with _MSK_, a central/global Kafka user can be created by the "cloud admin" or "the team" responsible for the _Operator_ deployment within the organization. The _Operator_ can then be easily configured to inject the configuration of that user as environment variables in all generated resources. Those variables can be used by the test to establish authentication with the kafka broker.

| Variable Name                    | Description                                                                      |
|:---------------------------------|:---------------------------------------------------------------------------------|
| `KAFKA_BOOTSTRAP_SERVERS`        | Kafka bootstrap servers                                                          |
| `KAFKA_SECURITY_ENABLED`         | -                                                                                |
| `KAFKA_SECURITY_PROTOCOL_CONFIG` | Security protocol. Options: `PLAINTEXT`, `SASL_PLAINTEXT`, `SASL_SSL`, `SSL`     |
| `KAFKA_SASL_MECHANISM`           | Authentication mechanism. Options: `PLAINTEXT`, `SCRAM-SHA-256`, `SCRAM-SHA-512` |
| `KAFKA_USERNAME`                 | The username used to authenticate Kafka clients with the Kafka server            |
| `KAFKA_PASSWORD`                 | The password used to authenticate Kafka clients with the Kafka server            |

--------

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
declared affinity labels **must** be present in nodes in order for resources to be deployed on them.

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

This optional sections allows deployed pods to have specific taint(s) tolerations. The features is also modeled to follow
closely [Kubernetes native definition](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/).

#### Spec breakdown & example

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

=== ":octicons-file-code-16: `taint-tolerations-example.yaml`"

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
            ...
        ...
    ```




