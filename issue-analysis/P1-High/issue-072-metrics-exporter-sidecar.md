# Issue #72: Operator Doesn't Manage the Metric Exporter Sidecar

**Priority**: P1 - High  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/72  
**Status**: Open  
**Labels**: Bug, Duplicate  
**Comments**: 11  
**Created**: 2022-12-07  
**Last Updated**: 2024-10-13

## Overview

After a locust test completes, the worker job/pod transitions to `Completed` state, but the master job continues running indefinitely. The root cause is that the `locust-metrics-exporter` sidecar container never stops, preventing the master pod from reaching `Completed` state.

## Problem Statement

Kubernetes doesn't provide native sidecar lifecycle management. When the main locust container exits gracefully, the metrics exporter sidecar continues running, which:
1. Prevents the master job from showing as `Completed`
2. Blocks programmatic detection of test completion
3. Causes issues for automation and CI/CD pipelines
4. Wastes cluster resources

## Technical Implications

### Root Cause Analysis
**@AbdelrhmanHamouda** identified:
> "The exporter is a sidecar container that the locust native image is not aware of its existence. Also Kubernetes doesn't provide native support for sidecars container behavior definition e.g. shutdown after container x exits."

### Current Workarounds

#### Option 1: Custom Locust Image
Call `/quitquitquit` endpoint in custom entrypoint:
```bash
curl -fsI -XPOST http://localhost:9646/quitquitquit
```

#### Option 2: In-Test Shutdown
Call endpoint from locust test's `@events.quitting.add_listener` function.

### Investigated Solutions

**PreStop Hook** (Rejected):
> "@AbdelrhmanHamouda: After some investigation, PreStop hook won't fit this use case. According to the documentation, it only gets invoked if the container termination is triggered from outside and not in cases where containers exit gracefully because of their internal process."

**Liveness Probe** (Under Investigation):
Assess if liveness probe can achieve desired effect without marking job as "error".

### Proposed Long-term Solution

**@AbdelrhmanHamouda**:
> "The proposed idea is to extend the Operator operation to include container management as a secondary resource to the custom resource. Meaning that after the Operator creates the main cluster resources, it needs to register a reconciler to manage the secondary resources created by k8s itself."

#### Implementation Plan
1. Register created pods/containers as secondary resources
2. Subscribe to container-level events
3. When locust-master container terminates, operator calls `/quitquitquit` endpoint
4. Ensures graceful sidecar shutdown via Kubernetes API

### Security Considerations
Need to investigate additional RBAC privileges required for container-level event watching and pod exec permissions.

## Why Implement This

### Production Impact
- **Blocks CI/CD automation**: Cannot programmatically determine test success/failure
- **Resource waste**: Pods continue running unnecessarily
- **User confusion**: Jobs appear "failed" when they're actually complete

### Scope of Impact
- Affects anyone running tests programmatically
- Critical for GitOps/automation workflows
- Issues also arise with other sidecar injectors (Istio, etc.)

### Current Limitations
- Workarounds require custom images or test modifications
- Not a clean, out-of-box experience
- Doesn't solve the general sidecar problem

## Author's Comments Summary

**@AbdelrhmanHamouda** (2022-12-12):
> "Your assessment of the root cause is on point. Indeed this behavior is because of the metrics exporter. This is a known issue that I am aware and intend on solving."

**@AbdelrhmanHamouda** (2024-10-04):
> "Thank you for your patience regarding this request. The fact is, that I never had the time to get around to this since it needed some prerequisite code changes (which were done thankfully)... That being said, I am starting to get some time back and I believe I will be able to give this request the attention it deserves."

**Key Insight**: You acknowledged the issue early and have a clear technical solution in mind. The operator needs secondary resource reconciliation to manage container lifecycles.

## Community Suggestions

**@eduardchernomaz** (2022-12-12):
> "I wonder if we can also just add a PreStop container hook to the master deployment which would call the /quitquitquit endpoint."

**@S-mishina** (2024-09-26):
Suggested adding a flag to disable metrics exporter as interim solution.

**Your Response**: Rejected the flag approach to avoid breaking changes and maintain observability as a core feature.

## Technical Complexity: **High**

### Required Changes
1. **Operator Logic**: Secondary resource reconciliation
2. **RBAC**: Additional cluster permissions
3. **Event Handling**: Container-level event subscription
4. **API Calls**: Programmatic endpoint invocation
5. **Testing**: Edge cases, race conditions

### Risks
- Increased operator complexity
- Additional cluster permissions
- Potential race conditions
- Backward compatibility concerns

## Recommended Solution Approach

### Phase 1: Operator-based Management
1. Extend operator to watch created pods as secondary resources
2. Subscribe to container exit events
3. Call `/quitquitquit` when locust-master container exits
4. Add configurable timeout for metrics scraping

### Phase 2: General Sidecar Support
1. Make solution generic for any sidecar (Istio, etc.)
2. Add configuration for custom sidecar shutdown logic
3. Document sidecar management capabilities

### Alternative: Native Kubernetes Sidecar Support
Monitor Kubernetes enhancement proposals for native sidecar lifecycle support (KEP-753).

## Related Issues
- #50 - Allow automatic resolution of completed tests (direct duplicate)
- General sidecar lifecycle problem in Kubernetes

## Estimated Effort
- **Research & Design**: Medium (2-3 days)
- **Implementation**: High (5-7 days)
- **RBAC & Security Review**: Medium (2 days)
- **Testing**: High (3-4 days)
- **Documentation**: Medium (1 day)
- **Total**: 13-17 days

## Success Criteria
1. Master pods transition to `Completed` state after test ends
2. Metrics are still scraped before sidecar shutdown
3. Works with other sidecar injectors (Istio, Linkerd)
4. No breaking changes to existing deployments
5. Configurable timeout for metrics collection
6. Comprehensive documentation
