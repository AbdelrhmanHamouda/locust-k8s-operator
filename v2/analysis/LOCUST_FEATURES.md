# Locust New Features - Go Migration Considerations

**Research Date:** January 2026  
**Locust Version Range:** 2.32.0 → 2.43.1 (Latest)  
**Purpose:** Document new Locust features that should be considered for the Go-based operator

---

## Executive Summary

This document catalogs significant Locust features introduced since the Java operator was developed. These features present opportunities to enhance the Go-based operator with native support for modern observability, new user types, and improved distributed testing capabilities.

**Key Highlights:**
- 🔭 **OpenTelemetry Integration** (v2.42.4) - Native OTel support for traces and metrics
- 🧩 **New Protocol Users** - DNSUser, MqttUser, SocketIOUser, MilvusUser, OpenAIUser
- 🧪 **Pytest Support** (v2.40.0) - Use pytest syntax for locustfiles
- 📊 **Enhanced Observability** - JSON output, improved HTML reports, profile support
- ⚡ **Performance Improvements** - Process-based distribution, faster startup

---

## Table of Contents

1. [Observability Features](#1-observability-features)
2. [New User Types & Protocols](#2-new-user-types--protocols)
3. [Distributed Testing Enhancements](#3-distributed-testing-enhancements)
4. [Configuration & CLI Improvements](#4-configuration--cli-improvements)
5. [Web UI Enhancements](#5-web-ui-enhancements)
6. [Operator Integration Recommendations](#6-operator-integration-recommendations)

---

## 1. Observability Features

### 1.1 OpenTelemetry Integration ⭐ HIGH PRIORITY

**Introduced:** v2.42.4 (November 2025)  
**Status:** Production-ready

Locust now natively supports OpenTelemetry for exporting traces and metrics to any OTel-compatible backend.

**Setup:**
```bash
pip install locust[otel]
locust --otel
```

**Configuration via Environment Variables:**
```bash
OTEL_TRACES_EXPORTER=otlp
OTEL_EXPORTER_OTLP_ENDPOINT=https://your-collector:4317
OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf  # or grpc
```

**Supported Exporters:**
- OTLP (gRPC and HTTP) - default
- Console (debugging)

**Auto-Instrumentation:**
| User Class | Instrumented Library |
|------------|---------------------|
| `HttpUser` | `requests`          |

**Operator Implications:**
- Consider adding `--otel` flag support to CR spec
- Add environment variable injection for OTel configuration
- Can potentially **replace** or complement the `locust_exporter` sidecar
- Native spans/traces provide richer data than Prometheus metrics alone

**Recommended CR Changes:**
```yaml
spec:
  observability:
    openTelemetry:
      enabled: true
      endpoint: "otel-collector.monitoring:4317"
      protocol: "grpc"  # or "http/protobuf"
      exporter: "otlp"  # or "console"
```

### 1.2 JSON Output & JSON File Export

**Introduced:** v2.37.0  
**Flag:** `--json-file <path>`

Export test results to JSON file format for programmatic processing.

**Operator Implications:**
- Add volume mount for JSON output
- Enable integration with test result aggregation systems

### 1.3 HTML Report Improvements

**Introduced:** v2.33.0, v2.35.0  
**Features:**
- Dynamic filename parsing (`--html` now supports templates)
- Profile argument display in reports
- Improved formatting

**New Flag:** `--profile <name>`
- Displays profile name in HTML reports
- Useful for identifying test configurations

### 1.4 Current Metrics Exporter Analysis

**Current Implementation:** `containersol/locust_exporter:v0.5.0` sidecar

**Considerations for Migration:**
| Approach | Pros | Cons |
|----------|------|------|
| **Keep Exporter Sidecar** | Proven, backward compatible | Additional container, limited to Prometheus |
| **Native OTel** | Richer data, traces, vendor-agnostic | Requires OTel infrastructure |
| **Both Options** | Maximum flexibility | More complex configuration |

**Recommendation:** Support both approaches, with OTel as the modern default.

---

## 2. New User Types & Protocols

### 2.1 DNSUser

**Introduced:** v2.42.0

Load test DNS servers directly.

```python
from locust import DNSUser

class MyDNSTest(DNSUser):
    @task
    def query_domain(self):
        self.client.query("example.com", "A")
```

### 2.2 MqttUser

**Introduced:** v2.41.0

Native MQTT protocol support using `paho-mqtt`.

```python
from locust import MqttUser

class MyMqttTest(MqttUser):
    @task
    def publish_message(self):
        self.client.publish("topic/test", "payload")
```

### 2.3 SocketIOUser

**Introduced:** v2.39.0  
**Status:** Experimental

WebSocket/SocketIO testing support.

```python
from locust import SocketIOUser

class MySocketTest(SocketIOUser):
    @task
    def send_event(self):
        self.client.emit("message", {"data": "test"})
```

### 2.4 MilvusUser

**Introduced:** v2.39.0

Load test Milvus vector databases.

### 2.5 OpenAIUser

**Introduced:** v2.34.0  
**Status:** Experimental

Specialized user for testing OpenAI/LLM APIs with proper token handling.

```python
from locust.contrib.openai import OpenAIUser

class MyAITest(OpenAIUser):
    @task
    def chat_completion(self):
        self.client.chat.completions.create(
            model="gpt-4",
            messages=[{"role": "user", "content": "Hello"}]
        )
```

### 2.6 MarkovTaskSet

**Introduced:** v2.38.0

State-machine based task execution for realistic user journey modeling.

**Operator Implications:**
- No direct operator changes needed
- Document supported user types in operator docs
- Consider validation of user type compatibility

---

## 3. Distributed Testing Enhancements

### 3.1 Process-Based Distribution

**Current:** `--processes` flag

```bash
# Auto-detect CPU cores
locust --processes -1

# Specific number of processes
locust --processes 4
```

**Key Features:**
- Simplifies single-node multi-process setup
- Uses `fork()` (Linux/macOS only)
- Can combine with `--worker` for worker-only processes

**Operator Implications:**
- Consider adding `processes` field to CR spec for worker pods
- Enables better CPU utilization per pod

**Recommended CR Addition:**
```yaml
spec:
  worker:
    replicas: 10
    processesPerPod: 4  # Run 4 worker processes per pod
```

### 3.2 Locustfile from Master

**Feature:** `-f -` argument

Workers can retrieve locustfile from master instead of local filesystem.

```bash
# On master
locust -f my_locustfile.py --master

# On worker (gets locustfile from master)
locust -f - --worker --master-host <master>
```

**Limitations:** Single locustfiles only

**Operator Implications:**
- Simplifies ConfigMap management for single-file tests
- Consider supporting this mode as an option

### 3.3 Custom Messaging Between Nodes

Enhanced `register_message()` API with:
- Concurrent message handling (non-blocking)
- Better payload support

```python
environment.runner.register_message('my_event', handler, concurrent=True)
```

### 3.4 --expect-workers Improvements

Better handling of worker connection synchronization.

**Current Operator:** Already uses `--expect-workers=<N>`

---

## 4. Configuration & CLI Improvements

### 4.1 TOML Configuration Support

**Introduced:** v2.42.x improvements

Locust now properly parses `.toml` configuration files alongside `.conf`.

### 4.2 List-Based Custom Arguments

**Introduced:** v2.38.0

```python
@events.init_command_line_parser.add_listener
def add_args(parser):
    parser.add_argument(
        "--regions",
        choices=["us-east", "eu-west", "ap-south"],
        is_multiple=True  # Multi-select dropdown in UI
    )
```

### 4.3 Web Base Path Configuration

**Introduced:** v2.32.2

`--web-base-path` allows serving UI under a subpath (useful for ingress routing).

```bash
locust --web-base-path /locust
```

**Operator Implications:**
- Useful for shared ingress configurations
- Add as optional CR field

### 4.4 StopTest Exception

**Introduced:** v2.43.0

Users can now raise `StopTest` to programmatically stop a test run.

```python
from locust import StopTest

class MyUser(HttpUser):
    @task
    def check_and_stop(self):
        if some_condition:
            raise StopTest("Stopping due to condition")
```

---

## 5. Web UI Enhancements

### 5.1 Responsive Design

**Introduced:** v2.34.0

Web UI now works on smaller screens/mobile devices.

### 5.2 Authentication Support

**Enhanced in:** v2.31.x - v2.32.x

- `--web-login` flag for authentication
- Custom auth providers (SSO, OAuth)
- Custom form fields
- Markdown support on auth page

**Operator Implications:**
- Consider adding auth configuration to CR
- Useful for production/shared environments

### 5.3 Host Validation

**Introduced:** v2.37.x

- Validates host URL format
- Warns on missing host
- Blocks test start in distributed mode until workers connect

### 5.4 Chart Improvements

- Zoom slider for charts
- Better time axis localization
- Line chart fixes

---

## 6. Operator Integration Recommendations

### 6.1 High Priority Features for v2.0

| Feature | Effort | Value | Recommendation |
|---------|--------|-------|----------------|
| **OpenTelemetry Support** | Medium | High | Native support in CR spec |
| **JSON Output** | Low | Medium | Add optional volume mount |
| **Processes per Pod** | Low | High | New field in worker spec |
| **Web Base Path** | Low | Medium | Optional field for ingress |

### 6.2 Proposed CR Spec Extensions

```yaml
apiVersion: locust.io/v1
kind: LocustTest
metadata:
  name: example-test
spec:
  # Existing fields...
  
  # NEW: Observability configuration
  observability:
    # Option 1: Legacy Prometheus exporter (current behavior)
    prometheusExporter:
      enabled: true
      image: "containersol/locust_exporter:v0.5.0"
      port: 9646
    
    # Option 2: Native OpenTelemetry (new)
    openTelemetry:
      enabled: false
      endpoint: ""
      protocol: "grpc"  # grpc | http/protobuf
      tracesExporter: "otlp"  # otlp | console
      metricsExporter: "otlp"
      # Additional env vars passed to pods
      envVars:
        OTEL_RESOURCE_ATTRIBUTES: "service.name=locust-test"
  
  # NEW: JSON/HTML output configuration
  output:
    jsonFile:
      enabled: false
      path: "/output/results.json"
    htmlReport:
      enabled: false
      path: "/output/report.html"
      profile: ""  # Optional profile name
  
  # NEW: Worker process configuration
  worker:
    replicas: 10
    processesPerPod: 1  # Number of locust processes per pod
  
  # NEW: Web UI configuration
  webUI:
    basePath: ""  # e.g., "/locust" for ingress routing
    auth:
      enabled: false
      # Additional auth config...
```

### 6.3 Command Line Mapping

| New Feature | CLI Flag | CR Field |
|-------------|----------|----------|
| OpenTelemetry | `--otel` | `observability.openTelemetry.enabled` |
| JSON Output | `--json-file` | `output.jsonFile.path` |
| HTML Report | `--html` | `output.htmlReport.path` |
| Profile | `--profile` | `output.htmlReport.profile` |
| Multi-process | `--processes` | `worker.processesPerPod` |
| Web Base Path | `--web-base-path` | `webUI.basePath` |

### 6.4 Environment Variable Requirements

For OpenTelemetry support, the operator should inject these env vars when enabled:

```yaml
env:
  - name: OTEL_TRACES_EXPORTER
    value: "otlp"
  - name: OTEL_METRICS_EXPORTER
    value: "otlp"
  - name: OTEL_EXPORTER_OTLP_ENDPOINT
    value: "{{ .observability.openTelemetry.endpoint }}"
  - name: OTEL_EXPORTER_OTLP_PROTOCOL
    value: "{{ .observability.openTelemetry.protocol }}"
```

### 6.5 Image Considerations

When using OpenTelemetry, users need Locust installed with OTel extras:
```dockerfile
pip install locust[otel]
```

The operator should document this requirement or provide a default image with OTel support.

---

## Deprecations & Breaking Changes

### Removed Features
- `--hatch-rate` parameter (v2.32.6) - use `--spawn-rate`
- `locust.cloud` references (v2.43.0) - service shut down
- Python 3.9 support deprecated (v2.34.1)

### Minimum Python Version
- Python 3.10+ recommended
- Python 3.14 officially supported (v2.41.6)

---

## References

- [Locust Changelog](https://github.com/locustio/locust/blob/master/CHANGELOG.md)
- [Locust Releases](https://github.com/locustio/locust/releases)
- [OpenTelemetry Documentation](https://docs.locust.io/en/stable/telemetry.html)
- [Distributed Testing](https://docs.locust.io/en/stable/running-distributed.html)

---

**Document Version:** 1.0  
**Companion Documents:**
- `ASSESSMENT.md` - Migration viability and framework recommendation
- `TECHNICAL.md` - Detailed technical analysis and migration guide

**Next Steps:**
1. Review proposed CR spec extensions with team
2. Prioritize features for v2.0 release
3. Create GitHub issues for each enhancement
4. Update Helm chart templates for new features
