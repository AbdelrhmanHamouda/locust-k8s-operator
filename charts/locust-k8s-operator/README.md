# Locust Kubernetes Operator

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/locust-k8s-operator)](https://artifacthub.io/packages/helm/locust-k8s-operator/locust-k8s-operator)

Production-ready Kubernetes operator for distributed Locust load testing with native OpenTelemetry support, high availability, and automatic v1→v2 migration.

## ✨ Features

- **🚀 Complete Go Rewrite**: 4x memory reduction, 60x faster startup
- **📊 Native OpenTelemetry**: Built-in metrics, traces, and optional collector deployment
- **🔒 High Availability**: Leader election support for multi-replica deployments
- **🔄 Zero-Downtime Migration**: Automatic conversion webhooks for v1→v2 upgrades
- **💚 Pod Health Monitoring**: Automatic pod recovery and health checks
- **🛡️ Production-Ready**: Comprehensive RBAC, security contexts, and resource management

## 📦 Installation

### Prerequisites

- Kubernetes 1.25+
- Helm 3.x

### Quick Start

```bash
# Add the Helm repository
helm repo add locust-k8s-operator https://abdelrhmanhamouda.github.io/locust-k8s-operator

# Update repository
helm repo update

# Install the operator
helm install my-locust-operator locust-k8s-operator/locust-k8s-operator

# Install with OpenTelemetry collector
helm install my-locust-operator locust-k8s-operator/locust-k8s-operator \
  --set otelCollector.enabled=true
```

## ⚙️ Configuration

### Key Values

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicas` | Number of operator replicas | `1` |
| `resources.limits.memory` | Memory limit | `128Mi` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `webhook.enabled` | Enable conversion webhooks | `false` |
| `webhook.certManager.enabled` | Use cert-manager for webhook certs | `false` |
| `otelCollector.enabled` | Deploy OTel collector sidecar | `false` |
| `leaderElection.enabled` | Enable leader election for HA | `true` |

### Example: Production Deployment with HA

```bash
helm install locust-operator locust-k8s-operator/locust-k8s-operator \
  --set replicas=3 \
  --set leaderElection.enabled=true \
  --set webhook.enabled=true \
  --set webhook.certManager.enabled=true \
  --set otelCollector.enabled=true \
  --set resources.limits.memory=256Mi
```

## 🔄 Migrating from v1.x

The operator includes automatic conversion webhooks for zero-downtime migration:

```bash
# Install v2 with webhooks enabled
helm upgrade locust-operator locust-k8s-operator/locust-k8s-operator \
  --set webhook.enabled=true \
  --set webhook.certManager.enabled=true

# Your existing v1 LocustTest resources will be automatically converted
```

For detailed migration guide, see: https://abdelrhmanhamouda.github.io/locust-k8s-operator/migration/

## 📚 Documentation

- **Full Documentation**: https://abdelrhmanhamouda.github.io/locust-k8s-operator/
- **Getting Started**: https://abdelrhmanhamouda.github.io/locust-k8s-operator/getting_started/
- **Migration Guide**: https://abdelrhmanhamouda.github.io/locust-k8s-operator/migration/
- **API Reference**: https://abdelrhmanhamouda.github.io/locust-k8s-operator/api_reference/

## 🐛 Support

- **GitHub Issues**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues

## 📝 License

Apache-2.0

## 🙏 Contributing

Contributions are welcome! Please see our [Contributing Guide](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/CONTRIBUTING.md).
