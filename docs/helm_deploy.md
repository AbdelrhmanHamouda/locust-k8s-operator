---
title: HELM deployment
description: Instructions on how to deploy Locust Kubernetes Operator with HELM
---

# HELM deployment

In order to deploy using helm, follow the below steps:

1. Add the _Operator´s_ HELM repo
    - `helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator/`

    !!! note 
    
        If the repo has been added before, run `helm repo update` in order to pull the latest available release! 

2. Install the _Operator_

    - `#!bash helm install locust-operator locust-k8s-operator/locust-k8s-operator`
          - The _Operator_ will ready up in around 40-60 seconds
    - This will cause the bellow resources to be deployed in the currently active _k8s_ context & namespace.
        - [crd-locusttest.yaml]
            - This _CRD_ is the first part of the _Operator_ pattern. It is needed in order to enable _Kubernetes_ to understand the _LocustTest_
              custom resource and allow its deployment.
        - [serviceaccount-and-roles.yaml]
            - ServiceAccount and Role bindings that enable the _Controller_ to have the needed privilege inside the cluster to watch and
              manage the related resources.
        - [deployment.yaml]
            - The _Controller_ responsible for managing and reacting to the cluster resources.

[//]: # (Resources urls)

[crd-locusttest.yaml]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/charts/locust-k8s-operator/crds/locust-test-crd.yaml

[serviceaccount-and-roles.yaml]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/charts/locust-k8s-operator/templates/serviceaccount-and-roles.yaml

[deployment.yaml]: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/charts/locust-k8s-operator/templates/deployment.yaml
