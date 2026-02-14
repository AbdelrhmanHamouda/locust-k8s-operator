---
title: Mount volumes to test pods
description: Attach data, certificates, or configuration files from various sources
tags:
  - configuration
  - volumes
  - storage
---

# Mount volumes to test pods

Mount test data, TLS certificates, or configuration files into Locust pods from PersistentVolumes, ConfigMaps, Secrets, or EmptyDir.

!!! info "v2 API only"
    Volume mounting is only available in the v2 API.

## Prerequisites

- Locust Kubernetes Operator v2.0+ installed
- Volume source created (PVC, ConfigMap, or Secret)

## Mount a PersistentVolumeClaim

Use a PVC to share large test data files across pods:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: pvc-volume-test
spec:
  image: locustio/locust:2.43.3
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
  volumes:  # Define the volume
    - name: test-data
      persistentVolumeClaim:
        claimName: test-data-pvc  # Must exist in same namespace
  volumeMounts:  # Mount into pods
    - name: test-data
      mountPath: /data  # Access files at /data in containers
      target: both      # Mount to both master and worker pods
```

Create the PVC first:

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-data-pvc
spec:
  accessModes:
    - ReadOnlyMany  # Multiple pods can read
  resources:
    requests:
      storage: 10Gi
```

!!! warning "StorageClass compatibility"
    Not all StorageClasses support `ReadOnlyMany` (ROX) access mode. Check your cluster's StorageClass documentation to confirm ROX support before using this access mode.

Apply both:

```bash
kubectl apply -f pvc.yaml
kubectl apply -f locusttest-pvc.yaml
```

## Mount a ConfigMap

Mount configuration files from a ConfigMap:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: configmap-volume-test
spec:
  image: locustio/locust:2.43.3
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
  volumes:
    - name: config-files
      configMap:
        name: app-config  # ConfigMap containing config files
  volumeMounts:
    - name: config-files
      mountPath: /config  # Files appear at /config/key1, /config/key2, etc.
      target: both
```

Create the ConfigMap:

```bash
kubectl create configmap app-config \
  --from-file=config.json \
  --from-file=settings.yaml
```

Your test script can read files:

```python
import json

# Read config from mounted volume
with open('/config/config.json') as f:
    config = json.load(f)
```

## Mount a Secret

Mount TLS certificates or API keys as files:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: secret-volume-test
spec:
  image: locustio/locust:2.43.3
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
  volumes:
    - name: tls-certs
      secret:
        secretName: tls-secret
  volumeMounts:
    - name: tls-certs
      mountPath: /certs
      readOnly: true  # Best practice for secrets
      target: both
```

Create the secret:

```bash
kubectl create secret generic tls-secret \
  --from-file=tls.crt=cert.pem \
  --from-file=tls.key=key.pem
```

Use in test:

```python
import requests

# Use client certificates
response = requests.get(
    'https://api.example.com',
    cert=('/certs/tls.crt', '/certs/tls.key')
)
```

## Use EmptyDir for temporary storage

Create temporary storage available within a pod (shared between containers in the same pod, but not across pods):

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: emptydir-volume-test
spec:
  image: locustio/locust:2.43.3
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
  volumes:
    - name: cache
      emptyDir: {}  # Created when pod starts, deleted when pod stops
  volumeMounts:
    - name: cache
      mountPath: /tmp/cache
      target: worker  # Only workers need cache
```

**Use cases for EmptyDir:**

- Temporary file processing
- Download cache
- Scratch space for generated data

**Note:** EmptyDir is pod-specific. Each worker pod has its own EmptyDir, not shared across pods.

## Target specific pod types

Control which pods receive the volume mount:

```yaml
volumeMounts:
  - name: test-data
    mountPath: /data
    target: master  # Options: master, worker, both (default)
```

| Target | Master | Worker | Use case |
|--------|--------|--------|----------|
| `master` | ✓ | ✗ | Master-only processing or UI data |
| `worker` | ✗ | ✓ | Worker-specific data or libraries |
| `both` (default) | ✓ | ✓ | Shared test data or configuration |

**Example with different targets:**

```yaml
volumes:
  - name: shared-data
    persistentVolumeClaim:
      claimName: shared-pvc
  - name: worker-cache
    emptyDir: {}

volumeMounts:
  - name: shared-data
    mountPath: /data
    target: both         # Both master and workers read test data
  - name: worker-cache
    mountPath: /cache
    target: worker       # Only workers need cache space
```

## Reserved mount paths

The following paths are reserved and cannot be used for volume mounts:

| Path | Purpose | Customizable |
|------|---------|--------------|
| `/lotest/src` | Test script mount point | Yes, via `testFiles.srcMountPath` |
| `/opt/locust/lib` | Library mount point | Yes, via `testFiles.libMountPath` |

If you customize these paths, the custom paths become reserved instead.

## Reserved volume names

The following volume name patterns are reserved:

| Pattern | Purpose |
|---------|---------|
| `<crName>-master` | Master ConfigMap volume |
| `<crName>-worker` | Worker ConfigMap volume |
| `locust-lib` | Library ConfigMap volume |
| `secret-*` | Secret volumes from `env.secretMounts` |

Choose different names for your volumes to avoid conflicts.

## Verify volume mount

Check that volumes are mounted correctly:

```bash
# Get a worker pod name
WORKER_POD=$(kubectl get pod -l performance-test-pod-name=<cr-name>-worker -o jsonpath='{.items[0].metadata.name}')

# Check mount exists
kubectl exec $WORKER_POD -- ls -la /data

# Verify file contents
kubectl exec $WORKER_POD -- cat /data/test-file.json
```

## Full example with multiple volumes

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: multi-volume-test
spec:
  image: locustio/locust:2.43.3
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  volumes:
    # Large test data on PVC
    - name: test-data
      persistentVolumeClaim:
        claimName: test-data-pvc
    # TLS certificates from secret
    - name: tls-certs
      secret:
        secretName: api-certs
    # Configuration from ConfigMap
    - name: app-config
      configMap:
        name: test-config
    # Temporary cache per worker
    - name: cache
      emptyDir: {}
  volumeMounts:
    - name: test-data
      mountPath: /data
      target: both         # All pods read test data
    - name: tls-certs
      mountPath: /certs
      readOnly: true
      target: both         # All pods use same certs
    - name: app-config
      mountPath: /config
      target: both         # All pods read config
    - name: cache
      mountPath: /tmp/cache
      target: worker       # Only workers use cache
```

## What's next

- **[Inject secrets](../security/inject-secrets.md)** — Pass credentials as environment variables instead of files
- **[Use private registry](use-private-registry.md)** — Pull custom images with volume-specific tools
- **[Configure resources](configure-resources.md)** — Ensure pods have enough resources for I/O operations
