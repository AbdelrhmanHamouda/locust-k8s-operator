---
title: License  
description: Contribution information.
---

# How does it work

To run a performance test, basic configuration is provided through a simple and intuitive kubernetes custom resource. Once deployed the
_Operator_ does all the heavy work of creating and scheduling the resources while making sure that all created load generation pods can
effectively communicate with each other.

To handle the challenge of delivering test script/s from local environment to the cluster and in turn to the deployed _locust_ pods,
the _Operator_ support dynamic volume mounting from a configMaps source. This is indicated by a simple optional configuration. Meaning, if
the configuration is present, the volume is mounted, and if it is not, no volume is mounted.

Since a "_Picture Is Worth a Thousand Words_", here is a gif!
![Short demo for how the operator works](assets/images/operatorDemo.gif "Operator Demo")

## Steps performed in demo

- Test configmap created in cluster.
- LocustTest CR deployed into the cluster.
- The _Operator_ creating, configuring and scheduling test resources on CR creation event.
- The _Operator_ cleaning up test resources after test CR has been removed event. contribution section.