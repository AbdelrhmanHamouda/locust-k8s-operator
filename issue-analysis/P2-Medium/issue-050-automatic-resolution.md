# Issue #50: Allow Automatic Resolution of Completed Tests

**Priority**: P2 - Medium  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/50  
**Status**: Open  
**Labels**: Enhancement, Feature request  
**Comments**: 6  
**Created**: 2022-11-02  
**Last Updated**: 2023-02-16

## Overview

Users launching tests programmatically need a way to determine whether a test was successful. Currently, the locust-exporter container never stops running, preventing pods from transitioning to `Completed` state, making it impossible to automatically detect test completion.

## Problem Statement

When running tests via automation/CI/CD:
- Admission webhooks monitor pod state changes
- Master pod never reaches `Completed` state
- Cannot programmatically determine test success/failure
- Must manually call `/quitquitquit` endpoint on exporter

**Note**: This is a duplicate of #72 from a different user perspective (programmatic test execution vs general job completion).

## Technical Implications

### Current Workaround
Manually calling `/quitquitquit` endpoint makes pod transition to `Completed`:
```bash
curl -X POST http://localhost:9646/quitquitquit
```

### Proposed Solutions

#### User's Initial Proposals

**Option 1: TTL After Finished**
- Set time-to-live after test completion
- Wait for Prometheus scraping interval
- Then terminate exporter
- **Challenge**: Ensuring metrics are scraped at least once

**Option 2: Prometheus Pushgateway**
- Push metrics instead of pull model
- Designed for ephemeral batch jobs
- **Challenge**: More architectural change, requires pushgateway deployment

#### Sidecar Container Approach
User suggested monitoring sidecar:
- Queries metrics endpoint
- Looks for `locust_up = 0`
- Calls `/quitquitquit` after TTL period

### Your Proposed Solution

**@AbdelrhmanHamouda**:
> "My current plan to address this is to use the recently added endpoint by having the controller register the created pod/container as a secondary resource and subscribe to events related to that specific pod/container. When the locust-master container terminates, the controller should be notified and will call the quit endpoint or execute a graceful shutdown through k8s api."

This is the same solution proposed for #72.

## Why Implement This

### CI/CD Integration
- Programmatic test execution requires completion detection
- Webhook-based automation blocked
- Integration with GitOps workflows

### User Experience
- Should work out-of-box without custom images
- Clean solution vs workarounds
- Industry standard behavior

### Metrics Reliability
- Need to ensure metrics are scraped before shutdown
- TTL approach allows configured scraping window
- Prometheus scraping frequency consideration

## Author's Comments Summary

**@AbdelrhmanHamouda** (2022-11-06):
Provided detailed analysis of each proposed solution:

**On Sidecar Approach**:
> "this will definitely work and get the desired result, however, it does not (at least to my eyes) look like a 'clean' solution. One of the things I am very keen on maintaining with the Operator, is to keep it clean and free of hacks / workarounds as much as possible."

**On TTL**:
> "do you have an idea on the mechanism to set the k8s/job TTL **after** its creation?"

**On Pushgateway**:
> "Performance tests are usually long enough to get meaningful metrics out of them and the Prometheus server for that is usually configured with high enough frequency to guarantee such result. Taking all this into account, a test that is so fast it does not get scraped would sound like an invalid test or miss-configured test."

**Preferred Solution**:
Controller-based secondary resource management for container lifecycle.

**Workaround Provided**:
Custom locust image with entrypoint script calling `/quitquitquit`.

**Key Insight**: You want a clean, operator-native solution rather than workarounds, and have a clear technical vision using secondary resource reconciliation.

## Technical Complexity: **High**

Same as #72 - requires operator enhancement for secondary resource management.

### Required Changes
1. **Operator Logic**: Secondary resource reconciliation
2. **Event Handling**: Container-level events
3. **Metrics Protection**: Ensure scraping before shutdown
4. **Configuration**: TTL/timeout settings
5. **Testing**: Race conditions, timing issues

## Recommended Solution Approach

### Align with #72 Solution
This issue should be resolved as part of #72 implementation.

### Additional Considerations
1. **Configurable TTL**: Allow users to set post-completion wait time
2. **Metrics Confirmation**: Optional validation that metrics were scraped
3. **Failure Detection**: Distinguish between test failure and completion

### Configuration Example
```yaml
# Helm values
locust:
  metricsExporter:
    shutdownTTL: 30  # seconds to wait after test completion
```

## Related Issues
- **#72** - Primary issue for this problem
- Should be fixed together

## Estimated Effort
**Included in #72 effort** - no additional work needed beyond #72 implementation.

## Success Criteria
1. Master pods transition to `Completed` after test completion
2. Metrics are scraped before exporter shutdown
3. Programmatic test success/failure detection works
4. Webhook integration supported
5. Configurable TTL for metrics scraping window
6. Works without custom images or test modifications

## Recommendation
**Close as duplicate of #72** and ensure #72 solution addresses the programmatic test execution use case.
