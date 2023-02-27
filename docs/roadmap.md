---
title: Roadmap  
description: Planned features for Locust Kubernetes Operator.
---

# Roadmap

Not in a particular order (list wis updated when features are implemented / planned):

- Add traceability labels to generated resources
- ✅ Support for deploying test resources with node affinity / node taints
- Dashboard examples (Grafana + prometheus configuration)
- Enable event driven actions
    - Integration with MSTeams: Push notification on test run completion / termination events
- _**UNDER_INVESTIGATION**_ Benchmarking and collection of non-test generated metrics
    - Investigation is on going to study the feasibility of supplying _Locust_ pods with external metrics that are collected from service system under-test. Test pods can then use this information to assess pass / fail criteria. This is especially useful in non-REST based services e.g. assess kafka(streams) microservice based on its _consumer lag_ performance coming from the kafka broker.
