# Phase 12: OpenTelemetry Support - Technical Design

**Version:** 1.0  
**Status:** Draft

---

## Overview

This document details the technical design for implementing native OpenTelemetry support in the Locust operator. The implementation modifies the command builders to add the `--otel` flag when enabled, injects OTel environment variables, and conditionally skips the metrics exporter sidecar.

---

## 1. Architecture

### 1.1 Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                      LocustTest CR (v2)                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ spec.observability.openTelemetry:                         │   │
│  │   enabled: true                                           │   │
│  │   endpoint: "otel-collector.monitoring:4317"              │   │
│  │   protocol: "grpc"                                        │   │
│  │   insecure: true                                          │   │
│  │   extraEnvVars:                                           │   │
│  │     OTEL_RESOURCE_ATTRIBUTES: "service.name=locust"       │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Resource Builders                             │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ IsOTelEnabled(lt)              → bool                    │   │
│  │ BuildOTelEnvVars(lt)           → []corev1.EnvVar         │   │
│  │ BuildMasterCommand(seed, ...)  → adds --otel flag        │   │
│  │ BuildWorkerCommand(seed, ...)  → adds --otel flag        │   │
│  │ shouldDeployMetricsSidecar(lt) → bool                    │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Master Job Pod Spec                          │
│  containers:                                                     │
│    - name: <crName>-master                                       │
│      args: ["locust", ..., "--otel", ...]                       │
│      env:                                                        │
│        - name: OTEL_TRACES_EXPORTER                             │
│          value: "otlp"                                          │
│        - name: OTEL_METRICS_EXPORTER                            │
│          value: "otlp"                                          │
│        - name: OTEL_EXPORTER_OTLP_ENDPOINT                      │
│          value: "otel-collector.monitoring:4317"                │
│        - name: OTEL_EXPORTER_OTLP_PROTOCOL                      │
│          value: "grpc"                                          │
│        - name: OTEL_EXPORTER_OTLP_INSECURE                      │
│          value: "true"                                          │
│        - name: OTEL_RESOURCE_ATTRIBUTES  # from extraEnvVars    │
│          value: "service.name=locust"                           │
│    # NO locust-exporter sidecar when OTel enabled               │
└─────────────────────────────────────────────────────────────────┘
```

### 1.2 Component Dependencies

```
api/v2/locusttest_types.go         ← Already defines ObservabilityConfig, OpenTelemetryConfig
         │
         ▼
internal/resources/otel.go         ← NEW: OTel env var building functions
         │
         ▼
internal/resources/command.go      ← MODIFY: Add --otel flag support
         │
         ▼
internal/resources/job.go          ← MODIFY: Conditional sidecar, merge OTel env vars
         │
         ▼
internal/resources/service.go      ← MODIFY: Conditional metrics port exclusion
         │
         ▼
api/v2/locusttest_webhook.go       ← MODIFY: Validate OTel configuration
```

---

## 2. OTel Environment Variables

### 2.1 Standard OTel SDK Variables

| Variable | Value Source | Required |
|----------|--------------|----------|
| `OTEL_TRACES_EXPORTER` | `"otlp"` (hardcoded) | When enabled |
| `OTEL_METRICS_EXPORTER` | `"otlp"` (hardcoded) | When enabled |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `spec.observability.openTelemetry.endpoint` | When enabled |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `spec.observability.openTelemetry.protocol` | When enabled |
| `OTEL_EXPORTER_OTLP_INSECURE` | `spec.observability.openTelemetry.insecure` | Optional |

### 2.2 Extra Environment Variables

Users can specify additional OTel SDK configuration via `extraEnvVars`:

```yaml
observability:
  openTelemetry:
    enabled: true
    endpoint: "otel-collector:4317"
    extraEnvVars:
      OTEL_RESOURCE_ATTRIBUTES: "service.name=locust,environment=prod"
      OTEL_TRACES_SAMPLER: "parentbased_traceidratio"
      OTEL_TRACES_SAMPLER_ARG: "0.1"
      OTEL_LOGS_EXPORTER: "none"
```

---

## 3. Implementation Details

### 3.1 OTel Helper Functions

**File:** `internal/resources/otel.go`

```go
package resources

import (
    "strconv"

    locustv2 "github.com/AbdelrhmanHamouda/locust-k8s-operator/api/v2"
    corev1 "k8s.io/api/core/v1"
)

// OTel environment variable names
const (
    EnvOTelTracesExporter     = "OTEL_TRACES_EXPORTER"
    EnvOTelMetricsExporter    = "OTEL_METRICS_EXPORTER"
    EnvOTelExporterEndpoint   = "OTEL_EXPORTER_OTLP_ENDPOINT"
    EnvOTelExporterProtocol   = "OTEL_EXPORTER_OTLP_PROTOCOL"
    EnvOTelExporterInsecure   = "OTEL_EXPORTER_OTLP_INSECURE"
)

// Default OTel values
const (
    OTelExporterOTLP     = "otlp"
    OTelProtocolGRPC     = "grpc"
    OTelProtocolHTTP     = "http/protobuf"
)

// IsOTelEnabled returns true if OpenTelemetry is enabled in the spec.
func IsOTelEnabled(lt *locustv2.LocustTest) bool {
    if lt.Spec.Observability == nil {
        return false
    }
    if lt.Spec.Observability.OpenTelemetry == nil {
        return false
    }
    return lt.Spec.Observability.OpenTelemetry.Enabled
}

// GetOTelConfig returns the OpenTelemetry configuration, or nil if not configured.
func GetOTelConfig(lt *locustv2.LocustTest) *locustv2.OpenTelemetryConfig {
    if lt.Spec.Observability == nil {
        return nil
    }
    return lt.Spec.Observability.OpenTelemetry
}

// BuildOTelEnvVars creates environment variables for OpenTelemetry configuration.
// Returns nil if OTel is not enabled.
func BuildOTelEnvVars(lt *locustv2.LocustTest) []corev1.EnvVar {
    if !IsOTelEnabled(lt) {
        return nil
    }

    otelCfg := GetOTelConfig(lt)
    if otelCfg == nil {
        return nil
    }

    var envVars []corev1.EnvVar

    // Core OTel exporter configuration
    envVars = append(envVars,
        corev1.EnvVar{Name: EnvOTelTracesExporter, Value: OTelExporterOTLP},
        corev1.EnvVar{Name: EnvOTelMetricsExporter, Value: OTelExporterOTLP},
    )

    // Endpoint (required when enabled - validated by webhook)
    if otelCfg.Endpoint != "" {
        envVars = append(envVars, corev1.EnvVar{
            Name:  EnvOTelExporterEndpoint,
            Value: otelCfg.Endpoint,
        })
    }

    // Protocol (default: grpc)
    protocol := otelCfg.Protocol
    if protocol == "" {
        protocol = OTelProtocolGRPC
    }
    envVars = append(envVars, corev1.EnvVar{
        Name:  EnvOTelExporterProtocol,
        Value: protocol,
    })

    // Insecure flag (only set if true)
    if otelCfg.Insecure {
        envVars = append(envVars, corev1.EnvVar{
            Name:  EnvOTelExporterInsecure,
            Value: strconv.FormatBool(true),
        })
    }

    // Extra environment variables from spec
    for key, value := range otelCfg.ExtraEnvVars {
        envVars = append(envVars, corev1.EnvVar{
            Name:  key,
            Value: value,
        })
    }

    return envVars
}
```

### 3.2 Command Builder Updates

**File:** `internal/resources/command.go`

Update command builders to accept OTel flag:

```go
// BuildMasterCommand constructs the command arguments for the master node.
// Template: "{seed} [--otel] --master --master-port=5557 --expect-workers={N} --autostart --autoquit 60 --enable-rebalancing --only-summary"
func BuildMasterCommand(commandSeed string, workerReplicas int32, otelEnabled bool) []string {
    var cmdParts []string
    cmdParts = append(cmdParts, commandSeed)
    
    // Add --otel flag if enabled (must come before other flags)
    if otelEnabled {
        cmdParts = append(cmdParts, "--otel")
    }
    
    cmdParts = append(cmdParts,
        "--master",
        fmt.Sprintf("--master-port=%d", MasterPort),
        fmt.Sprintf("--expect-workers=%d", workerReplicas),
        "--autostart",
        "--autoquit", "60",
        "--enable-rebalancing",
        "--only-summary",
    )
    
    // Flatten into string and split by whitespace
    cmd := strings.Join(cmdParts, " ")
    return strings.Fields(cmd)
}

// BuildWorkerCommand constructs the command arguments for worker nodes.
// Template: "{seed} [--otel] --worker --master-port=5557 --master-host={master-name}"
func BuildWorkerCommand(commandSeed string, masterHost string, otelEnabled bool) []string {
    var cmdParts []string
    cmdParts = append(cmdParts, commandSeed)
    
    // Add --otel flag if enabled (must come before other flags)
    if otelEnabled {
        cmdParts = append(cmdParts, "--otel")
    }
    
    cmdParts = append(cmdParts,
        "--worker",
        fmt.Sprintf("--master-port=%d", MasterPort),
        fmt.Sprintf("--master-host=%s", masterHost),
    )
    
    cmd := strings.Join(cmdParts, " ")
    return strings.Fields(cmd)
}
```

### 3.3 Job Builder Updates

**File:** `internal/resources/job.go`

Update Job builders to pass OTel flag and conditionally skip sidecar:

```go
// BuildMasterJob creates a Kubernetes Job for the Locust master node.
func BuildMasterJob(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
    nodeName := NodeName(lt.Name, Master)
    otelEnabled := IsOTelEnabled(lt)
    command := BuildMasterCommand(lt.Spec.Master.Command, lt.Spec.Worker.Replicas, otelEnabled)

    return buildJob(lt, cfg, Master, nodeName, command)
}

// BuildWorkerJob creates a Kubernetes Job for the Locust worker nodes.
func BuildWorkerJob(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *batchv1.Job {
    nodeName := NodeName(lt.Name, Worker)
    masterHost := NodeName(lt.Name, Master)
    otelEnabled := IsOTelEnabled(lt)
    command := BuildWorkerCommand(lt.Spec.Worker.Command, masterHost, otelEnabled)

    return buildJob(lt, cfg, Worker, nodeName, command)
}
```

Update `buildJob` to conditionally skip metrics sidecar:

```go
// buildJob is the internal function that constructs a Job for either master or worker.
func buildJob(lt *locustv2.LocustTest, cfg *config.OperatorConfig, mode OperationalMode, nodeName string, command []string) *batchv1.Job {
    // ... existing code ...

    // Build containers
    containers := []corev1.Container{
        buildLocustContainer(lt, nodeName, command, ports, cfg, mode),
    }

    // Master gets the metrics exporter sidecar ONLY if OTel is disabled
    if mode == Master && !IsOTelEnabled(lt) {
        containers = append(containers, buildMetricsExporterContainer(cfg))
    }

    // ... rest of existing code ...
}
```

Update `BuildEnvVars` or create a merged function that includes OTel env vars:

```go
// BuildEnvVars creates environment variables for the Locust container.
// Includes OTel env vars when enabled.
func BuildEnvVars(lt *locustv2.LocustTest, cfg *config.OperatorConfig) []corev1.EnvVar {
    var envVars []corev1.EnvVar

    // Existing env var logic...
    
    // Add OTel environment variables if enabled
    otelEnvVars := BuildOTelEnvVars(lt)
    if len(otelEnvVars) > 0 {
        envVars = append(envVars, otelEnvVars...)
    }

    return envVars
}
```

---

## 4. Service Builder Updates

**File:** `internal/resources/service.go`

The master Service currently always exposes the metrics port. When OTel is enabled and the sidecar is not deployed, this port should be excluded.

### 4.1 Current Implementation

```go
// BuildMasterService creates a Kubernetes Service for the Locust master node.
func BuildMasterService(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *corev1.Service {
    // ... existing code ...
    
    // Add metrics port (ALWAYS added currently)
    servicePorts = append(servicePorts, corev1.ServicePort{
        Name:     MetricsPortName,
        Protocol: corev1.ProtocolTCP,
        Port:     cfg.MetricsExporterPort,
    })
    // ...
}
```

### 4.2 Updated Implementation

```go
// BuildMasterService creates a Kubernetes Service for the Locust master node.
func BuildMasterService(lt *locustv2.LocustTest, cfg *config.OperatorConfig) *corev1.Service {
    // ... existing code ...
    
    // Add metrics port ONLY if OTel is disabled (sidecar will be deployed)
    if !IsOTelEnabled(lt) {
        servicePorts = append(servicePorts, corev1.ServicePort{
            Name:     MetricsPortName,
            Protocol: corev1.ProtocolTCP,
            Port:     cfg.MetricsExporterPort,
        })
    }
    // ...
}
```

### 4.3 Service Ports Summary

| Port | Name | Included When |
|------|------|---------------|
| 5557 | `port-5557` | Always (master communication) |
| 5558 | `port-5558` | Always (bind port) |
| 9646 | `metrics` | Only when OTel disabled |

---

## 5. Validation Webhook Updates

**File:** `api/v2/locusttest_webhook.go`

### 5.1 Validation Functions

```go
// validateOTelConfig validates OpenTelemetry configuration.
func (r *LocustTest) validateOTelConfig() (admission.Warnings, error) {
    if r.Spec.Observability == nil {
        return nil, nil
    }
    
    otelCfg := r.Spec.Observability.OpenTelemetry
    if otelCfg == nil {
        return nil, nil
    }
    
    // If OTel is enabled, endpoint is required
    if otelCfg.Enabled && otelCfg.Endpoint == "" {
        return nil, fmt.Errorf("observability.openTelemetry.endpoint is required when OpenTelemetry is enabled")
    }
    
    // Validate protocol if specified
    if otelCfg.Protocol != "" && otelCfg.Protocol != "grpc" && otelCfg.Protocol != "http/protobuf" {
        return nil, fmt.Errorf("observability.openTelemetry.protocol must be 'grpc' or 'http/protobuf', got %q", otelCfg.Protocol)
    }
    
    return nil, nil
}
```

### 5.2 Updated Validate Functions

```go
// ValidateCreate implements webhook.Validator
func (r *LocustTest) ValidateCreate() (admission.Warnings, error) {
    var allWarnings admission.Warnings
    
    // Existing validations...
    
    // OTel configuration validation
    warnings, err := r.validateOTelConfig()
    if err != nil {
        return allWarnings, err
    }
    allWarnings = append(allWarnings, warnings...)
    
    return allWarnings, nil
}
```

---

## 6. Testing Strategy

### 6.1 Unit Tests

**File:** `internal/resources/otel_test.go`

| Test | Description |
|------|-------------|
| `TestIsOTelEnabled_Nil` | Returns false when observability is nil |
| `TestIsOTelEnabled_Disabled` | Returns false when enabled is false |
| `TestIsOTelEnabled_True` | Returns true when enabled is true |
| `TestGetOTelConfig` | Returns config or nil appropriately |
| `TestBuildOTelEnvVars_Disabled` | Returns nil when OTel disabled |
| `TestBuildOTelEnvVars_MinimalConfig` | Core env vars with endpoint only |
| `TestBuildOTelEnvVars_FullConfig` | All env vars including protocol, insecure |
| `TestBuildOTelEnvVars_ExtraEnvVars` | Extra env vars merged correctly |
| `TestBuildOTelEnvVars_DefaultProtocol` | Defaults to grpc when not specified |

**File:** `internal/resources/command_test.go`

| Test | Description |
|------|-------------|
| `TestBuildMasterCommand_OTelDisabled` | No --otel flag when disabled |
| `TestBuildMasterCommand_OTelEnabled` | --otel flag present when enabled |
| `TestBuildWorkerCommand_OTelDisabled` | No --otel flag when disabled |
| `TestBuildWorkerCommand_OTelEnabled` | --otel flag present when enabled |

**File:** `internal/resources/job_test.go`

| Test | Description |
|------|-------------|
| `TestBuildMasterJob_OTelDisabled_HasSidecar` | Sidecar present when OTel disabled |
| `TestBuildMasterJob_OTelEnabled_NoSidecar` | Sidecar absent when OTel enabled |
| `TestBuildMasterJob_OTelEnabled_HasEnvVars` | OTel env vars in Locust container |
| `TestBuildWorkerJob_OTelEnabled_HasEnvVars` | OTel env vars in worker container |
| `TestBuildJob_OTelEnabled_CommandHasFlag` | Command includes --otel flag |

**File:** `internal/resources/service_test.go`

| Test | Description |
|------|-------------|
| `TestBuildMasterService_OTelDisabled_HasMetricsPort` | Metrics port present when OTel disabled |
| `TestBuildMasterService_OTelEnabled_NoMetricsPort` | Metrics port absent when OTel enabled |
| `TestBuildMasterService_NoObservability_HasMetricsPort` | Metrics port present when no config |

### 6.2 Webhook Tests

**File:** `api/v2/locusttest_webhook_test.go`

| Test | Description |
|------|-------------|
| `TestValidateOTelConfig_Disabled` | Passes when OTel disabled |
| `TestValidateOTelConfig_EnabledWithEndpoint` | Passes when enabled with endpoint |
| `TestValidateOTelConfig_EnabledNoEndpoint` | Fails when enabled without endpoint |
| `TestValidateOTelConfig_InvalidProtocol` | Fails for invalid protocol value |

### 6.3 Integration Tests

**File:** `internal/controller/integration_test.go`

| Test | Description |
|------|-------------|
| `TestReconcile_WithOTelEnabled` | Full reconcile with OTel config |
| `TestReconcile_OTelMasterNoSidecar` | Verify no sidecar in master pod |
| `TestReconcile_OTelEnvVarsInjected` | Verify OTel env vars in containers |

---

## 7. OTel Flag Position

The `--otel` flag should be added early in the command, but after the locustfile specification. Locust's argument parsing is flexible, but for consistency:

```bash
# Recommended order
locust -f /lotest/src/locustfile.py --otel --master --master-port=5557 ...

# Also valid
locust --otel -f /lotest/src/locustfile.py --master ...
```

Since the `commandSeed` already contains the `-f` flag, we append `--otel` immediately after it.

---

## 8. Backward Compatibility

| Scenario | Behavior |
|----------|----------|
| No `observability` field | Metrics sidecar deployed (current behavior) |
| `observability.openTelemetry.enabled: false` | Metrics sidecar deployed |
| `observability.openTelemetry.enabled: true` | No sidecar, OTel env vars injected |
| Existing v1 CRs | No change (v1 doesn't have observability field) |

---

## 9. Error Messages

Clear error messages for validation failures:

```
observability.openTelemetry.endpoint is required when OpenTelemetry is enabled
observability.openTelemetry.protocol must be 'grpc' or 'http/protobuf', got "invalid"
```

---

## 10. Image Requirements Note

The operator should document that OTel support requires a Locust image with OTel extras:

```dockerfile
FROM locustio/locust:2.32.0
RUN pip install locust[otel]
```

This is NOT validated by the operator since it would require image inspection. Users are responsible for using an appropriate image.

---

## 11. Future Considerations

### 11.1 OTel Collector Deployment (Out of Scope)

Future phases could add optional OTel Collector deployment via the Helm chart, similar to how the metrics exporter is currently deployed.

### 11.2 Logs Exporter

OTel also supports logs export. This could be added as an optional feature:

```yaml
observability:
  openTelemetry:
    enabled: true
    logsExporter: "otlp"  # Future enhancement
```

---

## 12. References

- [Locust OpenTelemetry Documentation](https://docs.locust.io/en/stable/telemetry.html)
- [OpenTelemetry Environment Variables](https://opentelemetry.io/docs/concepts/sdk-configuration/general-sdk-configuration/)
- [OTel OTLP Exporter Configuration](https://opentelemetry.io/docs/concepts/sdk-configuration/otlp-exporter-configuration/)
- [analysis/LOCUST_FEATURES.md §1.1](../../../analysis/LOCUST_FEATURES.md)
- [Phase 10 Implementation](../phase-10-env-secret-injection/) - Similar env var injection patterns
