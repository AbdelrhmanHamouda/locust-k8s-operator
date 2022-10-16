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




