# Issue #149: Inject Kubernetes Secrets into Locust Worker Pods

**Priority**: P1 - High  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/149  
**Status**: Open  
**Labels**: Good First Issue, Feature request, Pending Feedback  
**Comments**: 6  
**Created**: 2023-11-08  
**Last Updated**: 2025-07-07

## Overview

Users need the ability to inject Kubernetes secrets into locust worker pods to avoid hard-coding sensitive credentials (tokens, API keys) directly into locust test files. This is a highly requested feature with strong community support.

## Problem Statement

Currently, users must hard-code credentials into ConfigMaps containing locust files, which violates security best practices and makes GitOps workflows difficult. Users need secrets available as environment variables or mounted files that locust test scripts can access at runtime.

## Technical Implications

### Architecture Changes
1. **CRD Extension**: Extend the `LocustTest` CRD to accept secret references
2. **Environment Variable Injection**: Support mounting secrets as environment variables in worker/master pods
3. **File Mounting**: Support mounting secrets as files in the pod filesystem
4. **Scope**: Affects both worker and master pod specifications

### Implementation Considerations
- **Multiple Sources**: Support for various secret sources (Kubernetes Secrets, External Secrets Operator, Vault)
- **Flexibility**: Similar to cloudnative-pg approach with `passwordSecret` references
- **Backward Compatibility**: Must not break existing deployments

### Security Benefits
- Removes credentials from version control
- Enables proper secret rotation
- Aligns with GitOps best practices
- Integrates with existing secret management tools (External Secrets Operator, Vault)

### Technical Complexity: **Medium**
- CRD schema changes
- Pod spec template modifications
- Helm values updates
- Documentation updates

## Why Implement This

### High User Demand
- **6+ positive reactions** on the request
- **Multiple users** actively requesting this feature
- Community consensus on the need

### Production Readiness
- Critical for enterprise adoption
- Required for compliance with security policies
- Blocks GitOps workflows without it

### Competitive Parity
- Standard feature in similar operators
- cloudnative-pg provides good reference implementation

## Author's Comments Summary

**@AbdelrhmanHamouda** (2023-11-13):
> "Thank you for proposing this feature. I for one think that it makes perfect sense that we include something like that... This question help me size the request properly since I also have been thinking about fully supporting environment variables injection and that can have multiple sources, secrets are one of them."

**@AbdelrhmanHamouda** (2025-06-27):
> "thank you everyone for sharing this feedback, i'll take a run at it and nock it down one by one."

**Key Insight**: You acknowledged this as a valid, high-priority feature and committed to implementing it. You're considering a broader environment variable injection mechanism with secrets as one source.

## Community Feedback

**@PaulRudin** suggestion (2024-01-31):
> "I think more generally it would be useful to allow arbitrary additions/override of the pod spec for the workers and master. That's much more flexible and would address this and all sorts of other things people might reasonably want."

This suggests considering a more flexible pod spec override mechanism that would solve this and other issues (volume mounts, resource limits, etc.).

**@treenerd** use case (2024-03-14):
- Uses External Secrets Operator + Vault
- Prefers environment variables
- References locust's `web_ui_auth` example pattern

**@namiyousef** (2025-01-27):
- Requests ability to inject env vars into both locust AND metrics exporter pods

## Recommended Solution Approach

### Phase 1: Environment Variables from Secrets
```yaml
apiVersion: locust.io/v1
kind: LocustTest
spec:
  secretRefs:
    - name: api-credentials
      keys:
        - API_TOKEN
        - API_SECRET
```

### Phase 2: File Mounting
```yaml
apiVersion: locust.io/v1
kind: LocustTest
spec:
  secretRefs:
    - name: tls-certs
      mountPath: /etc/locust/secrets
```

### Phase 3: General Environment Variables
Support for ConfigMaps, field refs, etc., beyond just secrets.

## Related Issues
- #252 - Volume mounting capability (complementary feature)
- Both address pod customization needs

## Estimated Effort
- **Development**: Medium (2-3 days)
- **Testing**: Medium (1-2 days)
- **Documentation**: Low (0.5 day)
- **Total**: 4-6 days

## Success Criteria
1. Users can reference Kubernetes secrets in LocustTest CR
2. Secrets are available as environment variables in worker/master pods
3. Works with External Secrets Operator and similar tools
4. Backward compatible with existing deployments
5. Comprehensive documentation with examples
