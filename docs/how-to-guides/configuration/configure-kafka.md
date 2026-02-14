---
title: Configure Kafka and AWS MSK integration
description: Set up authenticated Kafka access for event-driven testing
tags:
  - configuration
  - kafka
  - aws msk
  - integration
---

# Configure Kafka and AWS MSK integration

Configure Locust pods to connect to authenticated Kafka clusters, including AWS MSK, for performance testing of event-driven architectures.

## Prerequisites

- Kafka cluster or AWS MSK cluster accessible from Kubernetes
- Kafka credentials (username/password for SASL authentication)
- Basic understanding of Kafka security protocols

!!! warning "Deprecated Feature"
    The operator-level Kafka configuration via Helm values is deprecated and will be removed in a future release. For new deployments, configure Kafka directly in your locustfile using [locust-plugins](https://github.com/SvenskaSpworker/locust-plugins) or a similar library. This approach gives you full control over Kafka client settings and is independent of the operator.

## Two-level configuration model

The operator supports two approaches to Kafka configuration:

**1. Operator-level (centralized):** Configure Kafka credentials once during operator installation. The operator automatically injects these as environment variables into all Locust pods. Test creators don't need to manage credentials.

**2. Per-test (override):** Specify Kafka configuration in individual LocustTest CRs using `spec.env.variables`. This overrides operator-level configuration for specific tests.

**Priority:** Per-test configuration overrides operator-level defaults.

## Configure at operator level (Helm)

Set Kafka credentials during operator installation:

```yaml
# values.yaml
kafka:
  enabled: true
  bootstrapServers: "kafka-broker1:9092,kafka-broker2:9092"
  security:
    enabled: true
    protocol: "SASL_SSL"        # Default: SASL_PLAINTEXT. Options: PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, or SSL
    saslMechanism: "SCRAM-SHA-512"  # PLAINTEXT, SCRAM-SHA-256, or SCRAM-SHA-512
    jaasConfig: ""                  # Optional: raw JAAS config string for advanced auth
  credentials:
    secretName: "kafka-credentials"    # Name of K8s Secret containing credentials
    usernameKey: "username"            # Key in Secret for username (default: "username")
    passwordKey: "password"            # Key in Secret for password (default: "password")
```

Install or upgrade the operator:

```bash
helm upgrade --install locust-operator locust-k8s-operator/locust-k8s-operator \
  --namespace locust-system \
  -f values.yaml
```

All Locust pods will automatically receive Kafka environment variables.

**For AWS MSK:**

```yaml
# values.yaml for AWS MSK with IAM authentication
kafka:
  enabled: true
  bootstrapServers: "b-1.mycluster.kafka.us-east-1.amazonaws.com:9096"
  security:
    enabled: true
    protocol: "SASL_SSL"             # Default: SASL_PLAINTEXT
    saslMechanism: "SCRAM-SHA-512"  # Or AWS_MSK_IAM for IAM auth
  credentials:
    secretName: "msk-credentials"      # Name of K8s Secret containing MSK credentials
    usernameKey: "username"            # Key in Secret for username (default: "username")
    passwordKey: "password"            # Key in Secret for password (default: "password")
```

## Configure per-test (override)

Override operator-level Kafka configuration for specific tests:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: kafka-test
spec:
  image: locustio/locust:2.43.3
  testFiles:
    configMapRef: kafka-test-script
  master:
    command: "--locustfile /lotest/src/kafka_test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/kafka_test.py"
    replicas: 5
  env:
    variables:
      - name: KAFKA_BOOTSTRAP_SERVERS
        value: "different-kafka:9092"  # Override operator setting
      - name: KAFKA_SECURITY_ENABLED
        value: "true"
      - name: KAFKA_SECURITY_PROTOCOL_CONFIG
        value: "SASL_SSL"
      - name: KAFKA_SASL_MECHANISM
        value: "SCRAM-SHA-256"
      - name: KAFKA_USERNAME
        value: "test-specific-user"
      - name: KAFKA_PASSWORD
        valueFrom:  # Reference secret for password
          secretKeyRef:
            name: kafka-test-creds
            key: password
```

Create the secret:

```bash
kubectl create secret generic kafka-test-creds \
  --from-literal=password='my-kafka-password'
```

Apply the test:

```bash
kubectl apply -f locusttest-kafka.yaml
```

## Available environment variables

When Kafka configuration is enabled, these environment variables are available in Locust pods:

| Variable | Description | Example values |
|----------|-------------|----------------|
| `KAFKA_BOOTSTRAP_SERVERS` | Kafka broker addresses | `broker1:9092,broker2:9092` |
| `KAFKA_SECURITY_ENABLED` | Whether security is enabled | `true`, `false` |
| `KAFKA_SECURITY_PROTOCOL_CONFIG` | Security protocol | `PLAINTEXT`, `SASL_PLAINTEXT`, `SASL_SSL`, `SSL` |
| `KAFKA_SASL_MECHANISM` | Authentication mechanism | `PLAINTEXT`, `SCRAM-SHA-256`, `SCRAM-SHA-512` |
| `KAFKA_USERNAME` | Kafka username | `kafka-user` |
| `KAFKA_PASSWORD` | Kafka password | `kafka-password` |
| `KAFKA_SASL_JAAS_CONFIG` | Raw JAAS configuration string for advanced authentication scenarios | `org.apache.kafka.common.security.scram.ScramLoginModule required username="user" password="pass";` |

## Use in Locust test script

Access Kafka environment variables in your test:

```python
# kafka_test.py
import os
from locust import User, task, between
from kafka import KafkaProducer, KafkaConsumer
import json

class KafkaUser(User):
    wait_time = between(1, 3)

    def on_start(self):
        """Initialize Kafka producer using operator-injected config."""
        bootstrap_servers = os.getenv('KAFKA_BOOTSTRAP_SERVERS').split(',')

        security_enabled = os.getenv('KAFKA_SECURITY_ENABLED', 'false').lower() == 'true'

        if security_enabled:
            # Use authenticated connection
            # Note: kafka-python uses sasl_plain_username/sasl_plain_password
            # for all SASL mechanisms (PLAIN, SCRAM-SHA-256, SCRAM-SHA-512),
            # not just PLAIN. The parameter names are misleading.
            self.producer = KafkaProducer(
                bootstrap_servers=bootstrap_servers,
                security_protocol=os.getenv('KAFKA_SECURITY_PROTOCOL_CONFIG', 'SASL_SSL'),
                sasl_mechanism=os.getenv('KAFKA_SASL_MECHANISM', 'SCRAM-SHA-512'),
                sasl_plain_username=os.getenv('KAFKA_USERNAME'),
                sasl_plain_password=os.getenv('KAFKA_PASSWORD'),
                value_serializer=lambda v: json.dumps(v).encode('utf-8')
            )
        else:
            # Use unauthenticated connection
            self.producer = KafkaProducer(
                bootstrap_servers=bootstrap_servers,
                value_serializer=lambda v: json.dumps(v).encode('utf-8')
            )

    @task
    def produce_message(self):
        """Send a message to Kafka."""
        message = {
            'user_id': 12345,
            'action': 'view_product',
            'timestamp': '2026-02-12T10:30:00Z'
        }

        future = self.producer.send('user-events', value=message)
        try:
            record_metadata = future.get(timeout=10)
            # Track success
            self.environment.events.request.fire(
                request_type="KAFKA",
                name="produce_message",
                response_time=future._elapsed_ms,
                response_length=len(str(message)),
                exception=None,
                context={}
            )
        except Exception as e:
            # Track failure
            self.environment.events.request.fire(
                request_type="KAFKA",
                name="produce_message",
                response_time=0,
                response_length=0,
                exception=e,
                context={}
            )
```

## Verify Kafka configuration

Check that environment variables are injected:

```bash
# Get a worker pod name
WORKER_POD=$(kubectl get pod -l performance-test-pod-name=kafka-test-worker -o jsonpath='{.items[0].metadata.name}')

# Verify Kafka environment variables
kubectl exec $WORKER_POD -- env | grep KAFKA_
```

Expected output:

```
KAFKA_BOOTSTRAP_SERVERS=kafka-broker1:9092,kafka-broker2:9092
KAFKA_SECURITY_ENABLED=true
KAFKA_SECURITY_PROTOCOL_CONFIG=SASL_SSL
KAFKA_SASL_MECHANISM=SCRAM-SHA-512
KAFKA_USERNAME=kafka-user
KAFKA_PASSWORD=kafka-password
```

## Troubleshoot connection issues

**Authentication failed:**

```python
kafka.errors.NoBrokersAvailable: NoBrokersAvailable
```

Check credentials:

```bash
kubectl exec $WORKER_POD -- env | grep KAFKA_USERNAME
kubectl exec $WORKER_POD -- env | grep KAFKA_PASSWORD
```

Verify credentials are correct in your Kafka cluster.

**Connection timeout:**

```python
kafka.errors.KafkaTimeoutError: KafkaTimeoutError
```

Check network connectivity:

```bash
# Test connection from pod
kubectl exec $WORKER_POD -- nc -zv kafka-broker1 9092
```

Verify:

- Bootstrap servers address is correct
- Network policies allow egress to Kafka
- Kafka cluster is reachable from Kubernetes

**Wrong protocol:**

```python
kafka.errors.BrokerResponseError: SASL_AUTHENTICATION_FAILED
```

Verify `KAFKA_SECURITY_PROTOCOL_CONFIG` matches your Kafka cluster setup:

```bash
kubectl exec $WORKER_POD -- env | grep KAFKA_SECURITY_PROTOCOL_CONFIG
```

## What's next

- **[Inject secrets](../security/inject-secrets.md)** — Manage Kafka credentials using Kubernetes secrets
- **[Scale worker replicas](../scaling/scale-workers.md)** — Size workers for high Kafka throughput
- **[Configure resources](configure-resources.md)** — Ensure pods have enough resources for Kafka connections
