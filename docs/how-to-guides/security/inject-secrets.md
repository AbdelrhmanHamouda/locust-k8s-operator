---
title: Inject secrets and configuration into test pods
description: Inject credentials and configuration into Locust test pods using ConfigMaps, Secrets, and environment variables
tags:
  - security
  - secrets
  - environment variables
  - configuration
---

# Inject secrets and configuration into test pods

Inject credentials and configuration into Locust test pods without hardcoding them in test files. The operator provides four methods for injecting data.

## Method 1: ConfigMap environment variables

Inject all keys from a ConfigMap as environment variables with an optional prefix.

**Create a ConfigMap:**

```bash
kubectl create configmap app-config \
  --from-literal=TARGET_HOST=https://api.example.com \
  --from-literal=LOG_LEVEL=INFO \
  --from-literal=TIMEOUT=30
```

**Reference in LocustTest CR:**

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: configmap-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
  env:
    configMapRefs:
      - name: app-config          # ConfigMap name
        prefix: "APP_"            # Prefix for all keys (optional)
```

**Result:** ConfigMap keys become environment variables with the prefix:
- `TARGET_HOST` → `APP_TARGET_HOST`
- `LOG_LEVEL` → `APP_LOG_LEVEL`
- `TIMEOUT` → `APP_TIMEOUT`

**Access in your locustfile:**

```python
import os

target_host = os.getenv('APP_TARGET_HOST')
log_level = os.getenv('APP_LOG_LEVEL', 'INFO')
timeout = int(os.getenv('APP_TIMEOUT', '30'))
```

## Method 2: Secret environment variables

Inject all keys from a Secret as environment variables with an optional prefix.

**Create a Secret:**

```bash
kubectl create secret generic api-credentials \
  --from-literal=API_TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... \
  --from-literal=API_KEY=sk_live_51H8... \
  --from-literal=DB_PASSWORD=secure-password-here
```

**Reference in LocustTest CR:**

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: secret-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  env:
    secretRefs:
      - name: api-credentials     # Secret name
        prefix: ""                # No prefix (use key names directly)
```

**Result:** Secret keys become environment variables:
- `API_TOKEN` → `API_TOKEN`
- `API_KEY` → `API_KEY`
- `DB_PASSWORD` → `DB_PASSWORD`

**Access in your locustfile:**

```python
import os

api_token = os.getenv('API_TOKEN')
api_key = os.getenv('API_KEY')
db_password = os.getenv('DB_PASSWORD')
```

!!! warning "Secret values in pod specs"
    Kubernetes injects Secret values as environment variables. They're visible in pod specs. Use RBAC to restrict access to pod definitions.

## Method 3: Individual variables

Define individual environment variables with literal values or references to ConfigMap/Secret keys. This gives you fine-grained control over which keys to inject.

**Create sources:**

```bash
kubectl create configmap app-settings --from-literal=api-url=https://api.example.com
kubectl create secret generic auth --from-literal=token=secret-token-here
```

**Reference in LocustTest CR:**

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: individual-vars-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
  env:
    variables:
      # Literal value
      - name: ENVIRONMENT
        value: "staging"

      # Reference to Secret key
      - name: API_TOKEN
        valueFrom:
          secretKeyRef:
            name: auth                    # Secret name
            key: token                    # Key within Secret

      # Reference to ConfigMap key
      - name: API_URL
        valueFrom:
          configMapKeyRef:
            name: app-settings            # ConfigMap name
            key: api-url                  # Key within ConfigMap
```

**Result:** Three environment variables are injected:
- `ENVIRONMENT=staging` (literal value)
- `API_TOKEN=secret-token-here` (from Secret)
- `API_URL=https://api.example.com` (from ConfigMap)

**Use cases:**
- Mix literal values with secrets/configs
- Select specific keys from ConfigMaps/Secrets
- Set defaults with fallback to secrets for sensitive values

## Method 4: Secret file mounts

Mount secrets as files in the container filesystem. This is useful for:
- TLS certificates
- Credential files (JSON key files, kubeconfig, etc.)
- Configuration files that must be read from disk

**Create a Secret from files:**

```bash
kubectl create secret generic tls-certs \
  --from-file=ca.crt=path/to/ca.crt \
  --from-file=client.crt=path/to/client.crt \
  --from-file=client.key=path/to/client.key
```

**Reference in LocustTest CR:**

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: file-mount-test
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 3
  env:
    secretMounts:
      - name: tls-certs                   # Secret name
        mountPath: /etc/locust/certs      # Mount path in container
        readOnly: true                    # Mount as read-only (recommended)
```

**Result:** Secret keys become files at mount path:
- `/etc/locust/certs/ca.crt`
- `/etc/locust/certs/client.crt`
- `/etc/locust/certs/client.key`

**Access in your locustfile:**

```python
import ssl

# Create SSL context with mounted certificates
ssl_context = ssl.create_default_context(cafile='/etc/locust/certs/ca.crt')
ssl_context.load_cert_chain(
    certfile='/etc/locust/certs/client.crt',
    keyfile='/etc/locust/certs/client.key'
)

# Use in HTTP client
# (implementation depends on your HTTP library)
```

### Reserved paths

The following paths are reserved and cannot be used for secret mounts:

| Path | Purpose | Customizable via |
|------|---------|------------------|
| `/lotest/src/` | Test script mount point | `testFiles.srcMountPath` |
| `/opt/locust/lib` | Library mount point | `testFiles.libMountPath` |

If you customize `srcMountPath` or `libMountPath`, those custom paths become reserved instead.

## Combined example

Use multiple injection methods together:

```yaml
apiVersion: locust.io/v2
kind: LocustTest
metadata:
  name: combined-injection
spec:
  image: locustio/locust:2.20.0
  master:
    command: "--locustfile /lotest/src/test.py --host https://api.example.com"
  worker:
    command: "--locustfile /lotest/src/test.py"
    replicas: 5
  env:
    # Method 1: ConfigMap environment variables
    configMapRefs:
      - name: app-config
        prefix: "APP_"

    # Method 2: Secret environment variables
    secretRefs:
      - name: api-credentials
        prefix: ""

    # Method 3: Individual variables
    variables:
      - name: ENVIRONMENT
        value: "production"
      - name: REGION
        value: "us-west-2"
      - name: SPECIAL_TOKEN
        valueFrom:
          secretKeyRef:
            name: special-auth
            key: token

    # Method 4: Secret file mounts
    secretMounts:
      - name: tls-certs
        mountPath: /etc/locust/certs
        readOnly: true
      - name: service-account-key
        mountPath: /etc/locust/keys
        readOnly: true
```

**Result:**
- All keys from `app-config` ConfigMap with `APP_` prefix
- All keys from `api-credentials` Secret (no prefix)
- Literal values: `ENVIRONMENT`, `REGION`
- Individual secret reference: `SPECIAL_TOKEN`
- Files mounted at `/etc/locust/certs/` and `/etc/locust/keys/`

## Verification

### Check environment variables

Verify that environment variables were injected into test pods:

```bash
# Get a pod name
POD=$(kubectl get pods -l performance-test-name=combined-injection -o jsonpath='{.items[0].metadata.name}')

# Check all environment variables
kubectl exec $POD -- printenv | sort

# Check specific prefix
kubectl exec $POD -- printenv | grep "APP_"

# Check specific variable
kubectl exec $POD -- printenv API_TOKEN
```

### Check file mounts

Verify that secret files were mounted:

```bash
# List files in mount path
kubectl exec $POD -- ls -la /etc/locust/certs/

# Read file content (use with caution for sensitive data)
kubectl exec $POD -- cat /etc/locust/certs/ca.crt
```

### Troubleshooting

| Problem | Symptom | Solution |
|---------|---------|----------|
| Pod stuck in `Pending` | ConfigurationError condition | Verify ConfigMap/Secret exists: `kubectl get configmap,secret` |
| Environment variable missing | Variable not in `printenv` output | Check spelling of ConfigMap/Secret name and key |
| File mount empty | Directory exists but no files | Verify Secret exists and has data: `kubectl get secret <name> -o yaml` |
| Permission denied reading file | `cat` fails with permission error | Check `readOnly: true` and Secret file permissions |

**Check PodsHealthy condition:**

```bash
kubectl get locusttest combined-injection -o jsonpath='{.status.conditions[?(@.type=="PodsHealthy")]}'
```

If `status=False` with reason `ConfigurationError`, the error message shows which ConfigMap or Secret is missing.

## Security best practices

1. **Use Secrets for sensitive data:** Never use ConfigMaps for passwords, tokens, or keys.

2. **Use RBAC to restrict Secret access:** Limit who can read Secrets in your namespace:
   ```bash
   # Users should NOT have direct Secret access
   # Only the operator's service account needs it
   ```

3. **Rotate secrets regularly:** See [Security Best Practices - Secret Rotation](../../security.md#secret-rotation) for the rotation process.

4. **Use External Secrets Operator:** For production, sync secrets from external vaults (AWS Secrets Manager, HashiCorp Vault, etc.). See [Security Best Practices - External Secrets](../../security.md#external-secrets-integration).

5. **Prefer file mounts for certificates:** Mount TLS certificates as files instead of environment variables (harder to accidentally log).

6. **Use read-only mounts:** Always set `readOnly: true` for secret mounts to prevent accidental modification.

## Related guides

- [Mount volumes](../configuration/mount-volumes.md) — Mount non-secret volumes (PVCs, ConfigMaps, emptyDir)
- [Security Best Practices](../../security.md) — RBAC, secret rotation, external secrets integration
- [API Reference - EnvConfig](../../api_reference.md#envconfig) — Complete env configuration reference
