---
title: Use a private image registry
description: Pull Locust images from private registries with authentication
tags:
  - configuration
  - images
  - security
---

# Use a private image registry

Pull custom Locust images from private container registries like Docker Hub, GitHub Container Registry, or AWS ECR.

## Prerequisites

- Registry credentials (username/password or access token)
- Custom Locust image pushed to private registry

## Create image pull secret

Store registry credentials in a Kubernetes secret:

```bash
kubectl create secret docker-registry my-registry-secret \
  --docker-server=ghcr.io \
  --docker-username=myusername \
  --docker-password=ghp_myPersonalAccessToken \
  --docker-email=user@example.com
```

**For specific registries:**

=== "GitHub Container Registry"

    ```bash
    kubectl create secret docker-registry ghcr-secret \
      --docker-server=ghcr.io \
      --docker-username=myusername \
      --docker-password=ghp_myPersonalAccessToken \
      --docker-email=user@example.com
    ```

=== "Docker Hub"

    ```bash
    kubectl create secret docker-registry dockerhub-secret \
      --docker-server=docker.io \
      --docker-username=myusername \
      --docker-password=myAccessToken \
      --docker-email=user@example.com
    ```

=== "AWS ECR"

    ```bash
    # Get ECR login token (expires after 12 hours)
    aws ecr get-login-password --region us-east-1 | \
      kubectl create secret docker-registry ecr-secret \
        --docker-server=123456789012.dkr.ecr.us-east-1.amazonaws.com \
        --docker-username=AWS \
        --docker-password-stdin \
        --docker-email=user@example.com
    ```

=== "Google Container Registry"

    ```bash
    # Use JSON key file
    kubectl create secret docker-registry gcr-secret \
      --docker-server=gcr.io \
      --docker-username=_json_key \
      --docker-password="$(cat key.json)" \
      --docker-email=user@example.com
    ```

Verify the secret exists:

```bash
kubectl get secret my-registry-secret
```

## Reference secret in LocustTest

Add `imagePullSecrets` to your LocustTest CR:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: private-registry-test
spec:
  image: ghcr.io/mycompany/locust-custom:v1.2.3  # Private image
  imagePullSecrets:  # Reference the secret
    - name: my-registry-secret
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
```

Apply the CR:

```bash
kubectl apply -f locusttest-private.yaml
```

## Configure image pull policy

Control when Kubernetes pulls the image:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: pull-policy-test
spec:
  image: ghcr.io/mycompany/locust-custom:latest
  imagePullPolicy: Always  # Pull image every time
  imagePullSecrets:
    - name: my-registry-secret
  testFiles:
    configMapRef: my-test
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
```

**Pull policy options:**

| Policy | Behavior | When to use |
|--------|----------|-------------|
| `Always` | Pull image on every pod creation | Development with `:latest` tag or frequently updated images |
| `IfNotPresent` | Pull only if not cached locally | Stable versioned images (default for non-`:latest` tags) |
| `Never` | Never pull, use cached image only | Pre-loaded images or air-gapped environments |

**Recommendation:** Use `Always` with `:latest` tags. Use `IfNotPresent` or omit (default) with version tags like `v1.2.3`.

## Verify image pull

Check that pods successfully pulled the image:

```bash
# Get pod status
kubectl get pods -l locust.io/test-id=private-registry-test

# Check image field
kubectl get pod -l locust.io/role=master -o jsonpath='{.items[0].spec.containers[0].image}'
```

Expected output:

```
ghcr.io/mycompany/locust-custom:v1.2.3
```

Verify pull policy:

```bash
kubectl get pod -l locust.io/role=master -o jsonpath='{.items[0].spec.containers[0].imagePullPolicy}'
```

## Troubleshoot ImagePullBackOff

If pods fail with `ImagePullBackOff`:

```bash
# Check pod events
kubectl describe pod -l locust.io/test-id=private-registry-test | grep -A 10 "Events:"
```

**Common issues:**

**Authentication failed:**

```
Failed to pull image: unauthorized: authentication required
```

Fix: Verify secret credentials are correct. Recreate the secret if needed.

**Image not found:**

```
Failed to pull image: manifest unknown: manifest unknown
```

Fix: Verify image name, tag, and registry URL. Check the image exists:

```bash
# For Docker Hub
docker pull ghcr.io/mycompany/locust-custom:v1.2.3

# For AWS ECR
aws ecr describe-images --repository-name locust-custom --region us-east-1
```

**Wrong secret referenced:**

```
Couldn't find key .dockerconfigjson in Secret
```

Fix: Verify secret name in `imagePullSecrets` matches the created secret:

```bash
kubectl get secrets | grep registry
```

**Network policy blocking registry:**

```
Failed to pull image: dial tcp: i/o timeout
```

Fix: Check network policies allow egress to the registry:

```bash
kubectl get networkpolicies
```

## What's next

- **[Mount volumes](mount-volumes.md)** — Add test data or certificates to pods
- **[Inject secrets](../security/inject-secrets.md)** — Pass API keys and credentials as environment variables
- **[Configure resources](configure-resources.md)** — Set CPU and memory limits for custom images
