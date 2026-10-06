---
title: Install with OLM (OperatorHub)
description: Install the Locust Kubernetes Operator through Operator Lifecycle Manager, from OperatorHub.io or the OpenShift OperatorHub.
tags:
  - deployment
  - olm
  - operatorhub
  - openshift
  - installation
---

# Install with OLM (OperatorHub)

The operator is packaged as an [Operator Lifecycle Manager](https://olm.operatorframework.io/) (OLM) bundle under the package name `locust-k8s-operator`. The bundle is published to [OperatorHub.io](https://operatorhub.io/) for any Kubernetes cluster running OLM, and to the community catalog that OpenShift's built-in OperatorHub reads. If you don't run OLM, use the [Helm chart](helm_deploy.md) instead; both install the same operator image.

## How an OLM install differs from Helm

| | OLM | Helm (defaults) |
|---|---|---|
| Admission and conversion webhooks | On. OLM issues and rotates the certificates. | Off unless `webhook.enabled=true`, which also needs cert-manager. |
| `locust.io/v1` resources | Converted to v2 by the webhook. | Only converted when webhooks are on. |
| Install mode | All namespaces only. The operator watches every namespace and the conversion webhook rules out the other modes. | Cluster-wide. |
| Configuration | Environment variables through the Subscription (see [Configure the operator](#configure-the-operator)). | `values.yaml`. |
| Upgrades | OLM follows the `stable` channel. Set `installPlanApproval: Manual` if you want to approve each one. | `helm upgrade`. |

The bundle doesn't include the optional OpenTelemetry collector or the PodDisruptionBudget the chart can create.

## Install on Kubernetes from OperatorHub.io

These steps need the package to be live on [operatorhub.io/operator/locust-k8s-operator](https://operatorhub.io/operator/locust-k8s-operator). If that page doesn't exist, use [Helm](helm_deploy.md) or [install the bundle from this repository](#install-the-bundle-from-this-repository).

1. Install OLM, if the cluster doesn't have it yet. With the [`operator-sdk` CLI](https://sdk.operatorframework.io/docs/installation/):

    ```bash
    operator-sdk olm install
    ```

    OperatorHub.io's own instructions offer an install script too; either works.

2. Install the operator. This creates a Subscription in the `operators` namespace, which OLM sets up for all-namespace operators:

    ```bash
    kubectl create -f https://operatorhub.io/install/locust-k8s-operator.yaml
    ```

3. Wait until the ClusterServiceVersion reports `Succeeded`:

    ```bash
    kubectl get csv -n operators -w
    ```

## Install on OpenShift

These steps need the package to be in the **Community** catalog of your OpenShift version (4.16 or later). If searching for "Locust" in OperatorHub finds nothing, use [Helm](helm_deploy.md) instead.

=== "Web console"

    1. Open **Operators → OperatorHub** (or **Ecosystem → Software Catalog** on newer releases) and search for **Locust**.
    2. Pick **Locust Kubernetes Operator**, then **Install**.
    3. Keep **All namespaces on the cluster** as the installation mode. Use the suggested `locust-system` namespace or pick another.
    4. Wait for the status to show **Succeeded** under **Installed Operators**.

=== "CLI"

    ```yaml
    apiVersion: v1
    kind: Namespace
    metadata:
      name: locust-system
    ---
    apiVersion: operators.coreos.com/v1
    kind: OperatorGroup
    metadata:
      name: locust-system
      namespace: locust-system
    # No spec.targetNamespaces: the operator watches all namespaces.
    ---
    apiVersion: operators.coreos.com/v1alpha1
    kind: Subscription
    metadata:
      name: locust-k8s-operator
      namespace: locust-system
    spec:
      channel: stable
      name: locust-k8s-operator
      source: community-operators
      sourceNamespace: openshift-marketplace
      installPlanApproval: Automatic
    ```

    ```bash
    oc apply -f locust-operator-subscription.yaml
    oc get csv -n locust-system -w
    ```

    You can also install into the existing `openshift-operators` namespace, which already has an all-namespaces OperatorGroup. Then you only need the Subscription, with `namespace: openshift-operators`.

The operator doesn't pin a UID, so it runs under the default `restricted-v2` SCC. For the Locust pods it creates, see [OpenShift compatibility](how-to-guides/security/configure-pod-security.md#openshift-compatibility).

## Run a first test

Put a locustfile in a ConfigMap in the namespace where the test runs, then create the sample `LocustTest` from the listing:

```bash
kubectl create configmap locust-scripts --from-file=locustfile.py
kubectl apply -f https://raw.githubusercontent.com/AbdelrhmanHamouda/locust-k8s-operator/master/config/samples/locust_v2_locusttest_quickstart.yaml
kubectl get locusttest locusttest-sample -w
```

The [first load test tutorial](tutorials/first-load-test.md) walks through the rest.

## Configure the operator

The settings that Helm exposes under `locustPods` and `config` are environment variables on the operator Deployment. With OLM you set them on the Subscription; OLM merges them into the Deployment and a variable with the same name replaces the bundle's value. For example:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: locust-k8s-operator
  namespace: locust-system
spec:
  channel: stable
  name: locust-k8s-operator
  source: community-operators            # operatorhubio-catalog on OperatorHub.io installs
  sourceNamespace: openshift-marketplace # olm on OperatorHub.io installs
  config:
    env:
      - name: POD_CPU_LIMIT
        value: "2000m"
      - name: JOB_TTL_SECONDS_AFTER_FINISHED
        value: "3600"
      - name: DEFAULT_RUNTIME_CLASS_NAME
        value: gvisor
```

Common variables:

| Variable | Default in the bundle | Helm value |
|---|---|---|
| `POD_CPU_REQUEST` / `POD_MEM_REQUEST` / `POD_EPHEMERAL_REQUEST` | `250m` / `128Mi` / `30M` | `locustPods.resources.requests` |
| `POD_CPU_LIMIT` / `POD_MEM_LIMIT` / `POD_EPHEMERAL_LIMIT` | `1000m` / `1024Mi` / `50M` | `locustPods.resources.limits` |
| `MASTER_POD_*`, `WORKER_POD_*` (same suffixes) | unset (fall back to `POD_*`) | `locustPods.masterResources`, `locustPods.workerResources` |
| `JOB_TTL_SECONDS_AFTER_FINISHED` | unset | `locustPods.ttlSecondsAfterFinished` |
| `DEFAULT_RUNTIME_CLASS_NAME` | unset | `locustPods.runtimeClassName` |
| `ENABLE_AFFINITY_CR_INJECTION` / `ENABLE_TAINT_TOLERATIONS_CR_INJECTION` | `true` | `locustPods.affinityInjection` / `locustPods.tolerationsInjection` |
| `METRICS_EXPORTER_IMAGE` | `docker.io/containersol/locust_exporter:v0.5.2` | `locustPods.metricsExporter.image` |

Resource requests and limits for the operator pod itself go in `spec.config.resources`.

## Uninstall

```bash
kubectl delete subscription locust-k8s-operator -n <namespace>
kubectl delete csv -n <namespace> -l operators.coreos.com/locust-k8s-operator.<namespace>
```

OLM leaves the `locusttests.locust.io` CRD and your `LocustTest` resources in place. Delete the CRD only when you want every test gone:

```bash
kubectl delete crd locusttests.locust.io
```

## Install the bundle from this repository

The bundle lives in [`bundle/`](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/tree/master/bundle) and is regenerated with `make bundle`. To try it on a cluster with OLM before (or without) a catalog listing, build and push the bundle image to a registry the cluster can pull from, then let `operator-sdk` create a throwaway catalog for it:

```bash
make bundle-build bundle-push BUNDLE_IMG=<registry>/locust-k8s-operator-bundle:v2.3.1
operator-sdk run bundle <registry>/locust-k8s-operator-bundle:v2.3.1 --install-mode AllNamespaces
```

`operator-sdk cleanup locust-k8s-operator` removes it again.
