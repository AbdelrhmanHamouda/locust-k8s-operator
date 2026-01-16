# Issue #118: Metrics & Dashboard Documentation

**Priority**: P2 - Medium  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/118  
**Status**: Open  
**Labels**: Pending Feedback  
**Comments**: 6  
**Created**: 2023-07-06  
**Last Updated**: 2024-05-23

## Overview

Users need comprehensive documentation on the Metrics & Dashboard functionality. Current documentation is sparse, leaving users with questions about what exists, how to configure it, and how to integrate with their observability stack.

## Problem Statement

Users have basic questions about the metrics functionality:
- What metrics are available?
- How to integrate with Prometheus/Grafana?
- What configuration is required?
- Is there a GUI?
- How to set up ServiceMonitors/PodMonitors?

## Technical Implications

### Current Implementation

**@AbdelrhmanHamouda's explanation**:
> "Every running test gets an automatically configured locust exporter container. This is a very stable and robust Go based project [locust_exporter](https://github.com/ContainerSolutions/locust_exporter). In our production environment where we are using the operator with thousands of tests, no issues were ever encountered with this exporter."

**How it works**:
1. Each test gets a locust-exporter sidecar automatically
2. Exporter exposes Prometheus-compatible metrics
3. Metrics can be scraped by any Prometheus server
4. Visualized in Grafana, NewRelic, DataDog, etc.
5. Plug-and-play with existing observability solutions

### Your Production Setup

**NewRelic Setup**:
> "We used a standard Prometheus agent that was configured to send its scraped metrics to NewRelic. The configuration was exactly as mentioned in the 'General idea' section."

**DataDog Setup**:
> "For DataDog, they have a proprietary scraper 'part of their DataDog agent' that basically accepts 'Prometheus like' config which we configured to exactly the same thing => target pods with `prometheus.io/scrape: true` and use `prometheus.io/path:` & `prometheus.io/port` to know which port and endpoint to hit on that pod."

**Standard Approach**:
- Set Prometheus to scrape pods with `prometheus.io/scrape: true`
- Prometheus looks for `prometheus.io/path` and `prometheus.io/port` annotations
- Operator already sets these annotations on master pods
- No service-level configuration needed

### ServiceMonitor Challenge

**@dennrich's question** (kube-prometheus-stack user):
> "We are using `kube-prometheus-stack` and it is by default configured to use `servicemonitors` or `podmonitors` to discover additional targets, it doesn't discover based on annotations."

**Request**:
- Service labels for ServiceMonitor selector
- Or operator-generated ServiceMonitor as an option

## Why Implement This

### Documentation Gap
- Users cannot find information easily
- Repeated questions indicate documentation insufficiency
- Production-tested knowledge not captured in docs

### Integration Barriers
- Different Prometheus setups (annotation-based vs ServiceMonitor-based)
- Users don't know how to configure their setup
- Missing examples for common stacks

### Feature Awareness
- Users may not realize metrics are available
- Under-utilization of built-in observability
- Missed value proposition

## Author's Comments Summary

**@AbdelrhmanHamouda** (2023-07-10):
> "Thank you for showing interest in the project and actively participating. It has been a while I want to write a nice documentation on this part of the operator but time is always eluding me on that front. I may use this issue as a reason to do so."

**@AbdelrhmanHamouda** (2024-05-23):
Provided detailed explanation of production setup with NewRelic and DataDog.

**Key Insight**: You have extensive production experience with metrics integration across multiple observability platforms but haven't had time to document it. The knowledge exists, just needs to be written down.

## Technical Complexity: **Low**

This is primarily a documentation effort, not a code change.

### Optional Enhancement: ServiceMonitor Support
Adding ServiceMonitor generation would be a **Medium** complexity feature:
1. Add helm option to create ServiceMonitor
2. Add proper service labels
3. Document usage

## Recommended Solution Approach

### Phase 1: Comprehensive Documentation

#### Document Structure
```
docs/metrics-and-observability.md
├── Overview
├── Architecture
│   ├── Locust Exporter Container
│   └── Available Metrics
├── Integration Guides
│   ├── Prometheus (Annotation-based)
│   ├── Prometheus (ServiceMonitor)
│   ├── Grafana
│   ├── NewRelic
│   ├── DataDog
│   └── Other Observability Tools
├── Configuration
│   ├── Prometheus Annotations
│   ├── Custom Exporter Images
│   └── Exporter Configuration
├── Grafana Dashboards
│   ├── Example Dashboards
│   └── Dashboard JSON
└── Troubleshooting
```

#### Key Content to Include
1. **Metrics List**: Document all available metrics from locust_exporter
2. **Annotations Reference**: Pre-configured annotations on pods
3. **Integration Examples**: Real configurations for popular tools
4. **Dashboard Examples**: Sample Grafana dashboards or links
5. **ServiceMonitor Template**: For kube-prometheus-stack users
6. **Your Production Wisdom**: Lessons from thousands of tests

### Phase 2: ServiceMonitor Support (Optional)

```yaml
# values.yaml
metrics:
  serviceMonitor:
    enabled: false
    labels: {}
    interval: 30s
```

### Phase 3: Helm Chart Improvements
- Add service labels for easier ServiceMonitor creation
- Document label usage
- Provide ServiceMonitor YAML template

## Community Needs

### Annotation-Based Setup (Your Current Implementation)
```yaml
# Already set by operator on master pods
apiVersion: v1
kind: Pod
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9646"
    prometheus.io/path: "/metrics"
```

### ServiceMonitor Setup (Requested)
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: locust-metrics
spec:
  selector:
    matchLabels:
      app: locust  # Needs these labels on service
  endpoints:
    - port: metrics
      interval: 30s
```

## Related Issues
- #254 - Key metrics clarification (related documentation need)
- Part of broader documentation improvements

## Estimated Effort

### Documentation Only
- **Research & Content**: Medium (2-3 days)
- **Writing**: Medium (2-3 days)
- **Examples & Screenshots**: Low (1 day)
- **Review & Polish**: Low (1 day)
- **Total**: 6-8 days

### With ServiceMonitor Support
- **Add to above**: +3 days for implementation
- **Total**: 9-11 days

## Success Criteria

### Documentation
1. Comprehensive metrics documentation page
2. Integration guides for 3+ observability platforms
3. Example Grafana dashboard(s)
4. Complete metrics reference
5. Troubleshooting guide
6. ServiceMonitor template for kube-prometheus-stack

### Optional Feature
1. Helm option to create ServiceMonitor
2. Service labels configurable
3. Documentation for ServiceMonitor usage

## Quick Wins

### Immediate Documentation Additions
1. Add link to https://github.com/ContainerSolutions/locust_exporter/blob/f44dac5bc76834b8c830a3f12f170c16ef976773/locust_exporter.md
2. Document pre-configured annotations
3. Add your DataDog/NewRelic examples to docs
4. Create "Quick Start: Prometheus Integration" guide

### Service Labels (Code Change)
Add labels to master service for ServiceMonitor selectors:
```yaml
labels:
  app: locust
  component: master
  test: {{ .Values.testName }}
```

This allows users to create ServiceMonitors targeting the service.
