# Issue #254: Key Metrics Clarification

**Priority**: P3 - Low  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/254  
**Status**: Open  
**Labels**: Documentation, question  
**Comments**: 4  
**Created**: 2025-09-27  
**Last Updated**: 2025-10-04

## Overview

User requesting clarification on the metrics mentioned in documentation, including their type (histogram, time series), metadata availability, and whether additional metrics beyond those documented are available.

## Problem Statement

The documentation lists these metrics:
- `locust_requests_total`: The total number of requests made
- `locust_requests_failed_total`: The total number of failed requests
- `locust_response_time_seconds`: The response time of requests
- `locust_users`: The number of simulated users

**User's questions**:
1. What kind of metrics are they? Histogram? Time series?
2. Is there any metadata? Method/resource?
3. Are there other metrics available besides these?
4. Interested in stats history: https://docs.locust.io/en/stable/retrieving-stats.html

## Technical Implications

### Source of Truth

**@AbdelrhmanHamouda's response**:
> "The operator bundles in a stable version of this open source metrics exporter project https://github.com/ContainerSolutions/locust_exporter that itself does the scraping of the metrics from locust own endpoint. Regarding the meaning of the metrics + the metadata, I'm afraid I would be basically repeating documentation."

**Reference Documentation**:
https://github.com/ContainerSolutions/locust_exporter/blob/f44dac5bc76834b8c830a3f12f170c16ef976773/locust_exporter.md

### Potential Documentation Issue

**@AbdelrhmanHamouda**:
> "And I do realize that there is a misalignment between what's there and some of the metrics I listed which could be a possible bug on the operator side. I want to spin up the operator in my own setup to confirm but I have a problem in my setup (unrelated) that I need to fix before I can check the actual list and possibly update the docs to remove any unintended listings."

**Action Item**: Verify actual exposed metrics vs. documentation and update docs accordingly.

## Why Address This

### Documentation Accuracy
- Ensure documentation matches reality
- Prevent user confusion
- Build trust in documentation

### User Enablement
- Users need to know available metrics for dashboard building
- Understanding metric types helps with proper Prometheus queries
- Metadata knowledge enables better filtering/grouping

### Low-Hanging Fruit
- Quick documentation fix
- No code changes required
- Improves user experience

## Author's Comments Summary

**@AbdelrhmanHamouda** (2025-10-03):
Pointed to locust_exporter documentation and mentioned custom exporter configuration option:
> "I want to highlight that the operator also supports the ability to use your own exporter by configuring these values https://abdelrhmanhamouda.github.io/locust-k8s-operator/helm_deploy/?h=metrics+expo#metrics-exporter"

**@AbdelrhmanHamouda** (2025-10-04):
Acknowledged potential documentation bug.

**Key Insight**: The metrics come from locust_exporter project, so documentation should reference that project's metrics documentation rather than maintaining a separate list.

## Technical Complexity: **Very Low**

This is a documentation-only issue.

### Required Actions
1. Verify actual metrics exposed by locust_exporter
2. Compare with operator documentation
3. Update documentation to match reality
4. Add link to locust_exporter metrics documentation
5. Document metric types and metadata

## Recommended Solution Approach

### Phase 1: Documentation Audit
1. Deploy test with operator
2. Query Prometheus/check metrics endpoint
3. List actual available metrics
4. Compare with documentation

### Phase 2: Documentation Update

#### Link to Source
Instead of maintaining separate metrics list, reference the source:
```markdown
## Available Metrics

The locust-k8s-operator automatically configures a 
[locust_exporter](https://github.com/ContainerSolutions/locust_exporter) 
sidecar container with each test.

For a complete list of available metrics, their types, and metadata, see:
https://github.com/ContainerSolutions/locust_exporter/blob/main/locust_exporter.md

### Key Metrics Include:
- `locust_users` - Current number of simulated users (Gauge)
- `locust_requests_total` - Total requests by method/name/status (Counter)
- `locust_requests_fail_total` - Failed requests (Counter)
- `locust_response_time_seconds` - Response time histogram (Histogram)
- And more...
```

#### Document Metric Types
For each metric mentioned, specify:
- **Type**: Counter, Gauge, Histogram
- **Labels**: Available metadata (method, name, status, etc.)
- **Usage**: Example PromQL queries

### Phase 3: Additional Information
1. Link to Grafana dashboard examples
2. Document how to access raw stats endpoint if needed
3. Explain difference between exporter metrics and locust native stats

## Related Issues
- #118 - Metrics & Dashboard Documentation (broader documentation issue)

## Estimated Effort
- **Metrics Verification**: Very Low (0.25 day)
- **Documentation Update**: Very Low (0.5 day)
- **Examples & Review**: Very Low (0.25 day)
- **Total**: 1 day

## Success Criteria
1. Documentation accurately lists available metrics
2. Metric types (Counter/Gauge/Histogram) documented
3. Available labels/metadata documented
4. Link to locust_exporter documentation provided
5. Example Prometheus queries included
6. No discrepancies between docs and reality

## Quick Fix

### Immediate Action
Add to documentation:
```markdown
## Metrics Reference

The operator uses [locust_exporter](https://github.com/ContainerSolutions/locust_exporter)
to expose Prometheus metrics. For the complete and authoritative list of metrics, see:
https://github.com/ContainerSolutions/locust_exporter/blob/main/locust_exporter.md

You can also use a custom metrics exporter by configuring the operator's metrics
exporter settings. See the [deployment documentation](link) for details.
```

This immediately resolves the question by pointing to the authoritative source.
