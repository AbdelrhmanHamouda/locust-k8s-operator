# Issue #252: Provide Ability to Mount Volumes to Locust Master/Worker Pods

**Priority**: P1 - High  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/252  
**Status**: Open  
**Labels**: Stale  
**Comments**: 1  
**Created**: 2025-09-22  
**Last Updated**: 2025-12-28

## Overview

Users need the ability to mount volumes (PVCs, emptyDir, hostPath, etc.) to locust master and worker pods to retrieve application-specific results and support advanced use cases.

## Problem Statement

Currently, the CRD doesn't expose any mechanism to add volumes or volume mounts to the pod specifications created by the operator. This limitation prevents users from:
- Collecting application-specific output from worker pods
- Sharing data between workers
- Mounting configuration files beyond ConfigMaps
- Using persistent storage for test results

## Technical Implications

### Requested Capability
**User's request**:
> "This is a request for operator developers so that they can embed the containerSpec and give users full mutability over the specification of the Pods the operator creates. The operator developers should add that functionality to the CRD and users should be adding overridden fields to the CR."

### Implementation Requirements

#### CRD Extension
Add volume and volumeMount specifications to the CRD:
```yaml
apiVersion: locust.io/v1
kind: LocustTest
spec:
  volumes:
    - name: test-results
      persistentVolumeClaim:
        claimName: locust-results-pvc
    - name: shared-data
      emptyDir: {}
  volumeMounts:
    - name: test-results
      mountPath: /results
    - name: shared-data
      mountPath: /shared
```

#### Pod Template Modifications
- Merge user-defined volumes with operator-managed volumes
- Apply volumeMounts to both master and worker containers
- Support separate configurations for master vs worker pods

### Use Cases
1. **Result Collection**: Mount PVC to collect detailed test results
2. **Shared State**: EmptyDir for worker coordination
3. **Large Config Files**: Mount volumes for large test data files
4. **Custom Certificates**: Mount cert volumes for HTTPS testing
5. **Application Logging**: Separate volume for application-specific logs

## Why Implement This

### Flexibility & Extensibility
- Enables advanced use cases not covered by basic ConfigMap mounting
- Provides escape hatch for custom requirements
- Reduces need for custom operator modifications

### Enterprise Requirements
- Large organizations often need persistent result storage
- Compliance requirements may mandate result retention
- Integration with existing storage solutions

### Community Pattern
This is part of the broader request for pod spec customization (#149 mentions similar needs).

## Technical Complexity: **Medium**

### Required Changes
1. **CRD Schema**: Add `volumes` and `volumeMounts` fields
2. **Controller Logic**: Merge user volumes with operator volumes
3. **Validation**: Ensure no conflicts with operator-managed volumes
4. **Helm Chart**: Add default values for common patterns
5. **Documentation**: Examples for common use cases

### Risks
- Volume name conflicts with operator-managed volumes
- Path conflicts with operator-managed mounts
- Permission issues in restricted environments

## Recommended Solution Approach

### Phase 1: Basic Volume Support
```yaml
apiVersion: locust.io/v1
kind: LocustTest
spec:
  workerVolumes:
    - name: results
      persistentVolumeClaim:
        claimName: test-results
  workerVolumeMounts:
    - name: results
      mountPath: /app/results
```

### Phase 2: Advanced Volume Types
- Support for all Kubernetes volume types
- Projected volumes for combining sources
- CSI driver support

### Phase 3: General Pod Spec Override
Consider implementing the broader suggestion from #149 to allow arbitrary pod spec overrides via strategic merge patch.

## Related Issues
- #149 - Secret injection (complementary volume use case)
- Part of broader pod customization theme

## Estimated Effort
- **CRD Design**: Low (1 day)
- **Implementation**: Medium (2-3 days)
- **Validation Logic**: Low (1 day)
- **Testing**: Medium (2 days)
- **Documentation**: Low (1 day)
- **Total**: 7-8 days

## Success Criteria
1. Users can define volumes in LocustTest CR
2. Volumes are properly mounted in master/worker pods
3. No conflicts with operator-managed volumes
4. Support for common volume types (PVC, emptyDir, configMap, secret)
5. Separate configuration for master vs worker if needed
6. Examples for common patterns in documentation
7. Validation prevents operator-managed volume conflicts
