# Issue #246: Allow Master CPU and Memory Requests

**Priority**: P2 - Medium  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/246  
**Status**: Open  
**Labels**: Enhancement, Pending Feedback  
**Comments**: 2  
**Created**: 2025-08-08  
**Last Updated**: 2025-09-01

## Overview

Users need the ability to specify different CPU and memory resource requests/limits for master pods versus worker pods. Currently, the operator only supports unified resource settings that apply to both master and worker identically.

## Problem Statement

When scaling beyond 100 workers per master, the master node requires significantly more resources than individual workers. Users want to run many small workers with a high-resource master, but the current unified resource model forces the same constraints on both.

**User's requirement**:
> "I need to scale more than 100 workers per master, and want to keep worker resources requests and limits different from master, I need to run tiny workers, but high resources for master."

## Technical Implications

### Current State
The operator supports unified resource management via:
- https://abdelrhmanhamouda.github.io/locust-k8s-operator/advanced_topics/#resource-management

All pods (master and workers) get the same resource allocation.

### Requested Functionality

#### Helm Values Approach
```yaml
# Current (unified)
resources:
  limits:
    cpu: "1000m"
    memory: "1Gi"
  requests:
    cpu: "500m"
    memory: "512Mi"

# Proposed (separate)
masterResources:
  limits:
    cpu: "4000m"
    memory: "8Gi"
  requests:
    cpu: "2000m"
    memory: "4Gi"

workerResources:
  limits:
    cpu: "500m"
    memory: "512Mi"
  requests:
    cpu: "250m"
    memory: "256Mi"
```

### Implementation Considerations

#### Question from @AbdelrhmanHamouda:
> "How do you see it happening? is this something that changes with every run or more like a stable thing? The implication of this would be if this to be done in a CR or helm values"

#### User Response (@iahmad-khan):
> "It should be configurable in helm values, but currently the values we configure in helm apply to both worker and master the same. When deploying the operator we should be able to specify the resource requests and limits for master separately and then when create a load test CRD, the master should have those resources."

**Key Decision**: This should be helm-level configuration (operator-wide), not CR-level (per-test).

### Architecture Impact

1. **Helm Values Schema**: Add separate master/worker resource sections
2. **Maintain Backward Compatibility**: Keep unified `resources` field as default
3. **Precedence Logic**: 
   - If `masterResources`/`workerResources` specified → use them
   - Else → fall back to unified `resources`
4. **Validation**: Ensure resource values are valid Kubernetes quantities

## Why Implement This

### Resource Optimization
- **Cost Savings**: Right-size worker pods for efficiency
- **Master Scaling**: Enable 100+ workers per master
- **Cluster Efficiency**: Better bin-packing in multi-tenant clusters

### Real-World Use Case
At scale:
- Master coordinates 100+ workers: needs CPU for connection management
- Workers execute simple HTTP requests: minimal resources needed
- Current model: 100 workers × 4GB = 400GB wasted memory

### Industry Standard
Most Kubernetes operators with master/worker topology support differentiated resource allocation.

## Author's Comments Summary

**@AbdelrhmanHamouda** (2025-08-21):
> "at the moment, the operator support _unified_ resource request and limit mechanic. I understand this is not what you asked for but it is something that you can use for now if you are not aware of it."

**Key Insight**: You acknowledged the limitation and requested clarification on implementation approach (helm vs CR level).

## Technical Complexity: **Low-Medium**

### Required Changes
1. **Helm Values Schema**: Add `masterResources` and `workerResources`
2. **Operator Configuration**: Read separate resource configs
3. **Pod Creation Logic**: Apply correct resources based on pod type
4. **Backward Compatibility**: Maintain `resources` field functionality
5. **Validation**: Resource format validation
6. **Documentation**: Update helm values documentation

### Risks
- Minimal - backward compatible change
- Clear user intent
- Straightforward implementation

## Recommended Solution Approach

### Phase 1: Helm Values Extension
```yaml
# values.yaml
# Unified resources (backward compatible)
resources:
  limits:
    cpu: "1000m"
    memory: "1Gi"
  requests:
    cpu: "500m"
    memory: "512Mi"

# Override for master (optional)
masterResources: {}
  # limits:
  #   cpu: "4000m"
  #   memory: "8Gi"

# Override for worker (optional)
workerResources: {}
  # limits:
  #   cpu: "500m"
  #   memory: "512Mi"
```

### Phase 2: CR-Level Override (Optional Future Enhancement)
Allow per-test resource customization:
```yaml
apiVersion: locust.io/v1
kind: LocustTest
spec:
  masterResources:
    limits:
      cpu: "8000m"
      memory: "16Gi"
```

### Precedence Order
1. CR-level resources (if specified in future)
2. Helm masterResources/workerResources
3. Helm unified resources (fallback)

## Related Issues
None directly, but part of broader resource management improvements.

## Estimated Effort
- **Helm Schema Updates**: Low (0.5 day)
- **Operator Logic**: Low (1 day)
- **Testing**: Low (1 day)
- **Documentation**: Low (0.5 day)
- **Total**: 3 days

## Success Criteria
1. Users can specify `masterResources` in helm values
2. Users can specify `workerResources` in helm values
3. Backward compatible with unified `resources`
4. Master pods get master-specific resources
5. Worker pods get worker-specific resources
6. Clear documentation with examples
7. Validation prevents invalid resource specifications
