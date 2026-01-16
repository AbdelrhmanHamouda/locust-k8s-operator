# Issue #245: Configurable Master CMD Template

**Priority**: P2 - Medium  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/245  
**Status**: Open  
**Labels**: Feature request  
**Comments**: 4  
**Created**: 2025-07-08  
**Last Updated**: 2025-07-22

## Overview

The master node command template is hardcoded in the operator and not configurable. Users need to customize specific flags like `--autoquit`, `--autostart`, and `--expect-workers` to control test behavior and UI availability.

## Problem Statement

The master command template is defined in constants:
- https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/c61cde3cb09a8dc514dcb931c53883465e9634b0/src/main/java/com/locust/operator/controller/utils/Constants.java#L23-L33

Users cannot override these flags, which causes issues like:
- UI shutting down too quickly after test completion
- Inflexible autostart behavior
- Duplicate flag conflicts when added to `masterCommandSeed`

### Current Behavior
Adding `--autoquit 3600` to `masterCommandSeed` results in duplicate flags:
```yaml
args:
  - --run-time
  - 3m
  - --autoquit
  - "3600"        # User's value
  - --master
  - --master-port=5557
  - --expect-workers=1
  - --autostart
  - --autoquit
  - "60"          # Operator's hardcoded value
```

The last value (60) takes effect, ignoring the user's 3600.

## Technical Implications

### Requested Configurability

**@AbdelrhmanHamouda's proposed solution**:
> "looking at the request, i understand that the request is asking for the configurability of the below values:
> - `--expect-workers` -> currently configurable through changing the value for the workerReplicas value <- thus i'll assume we are good here
> - `--autostart` -> _disable | enable_ through helm values override
> - `--autoquit` -> _disable | enable | adjust_ through helm values override"

**User confirmed**: "Yes sounds good"

### Implementation Approach

#### Helm Values Configuration
```yaml
# values.yaml
locust:
  master:
    autostart:
      enabled: true
    autoquit:
      enabled: true
      timeout: 60  # seconds
    expectWorkers:
      # Derived from workerReplicas - no change needed
```

#### Operator Logic
1. Read helm values for autostart/autoquit configuration
2. Conditionally add flags to master command template
3. Use configurable timeout value for autoquit
4. Ensure no duplicate flags

### Use Case Details

**User's scenario**:
- Test runs for 3 minutes
- Default autoquit is 60 seconds after test completion
- User wants UI available for 3600 seconds (1 hour) for analysis
- Current implementation: UI shuts down after ~4 minutes total

**Logs showing issue**:
```
[2025-07-08 15:20:56] Run time limit set to 180 seconds
[2025-07-08 15:23:56] --run-time limit reached, stopping test
[2025-07-08 15:27:53] --autoquit time reached, shutting down  # ~4 mins after start
```

User wants the shutdown at ~1 hour after test completion for longer UI access.

## Why Implement This

### User Experience
- **Post-test Analysis**: Users need time to review UI results
- **Debugging**: Longer UI availability helps troubleshoot issues
- **Flexibility**: Different tests have different requirements

### Configuration Flexibility
- Some users want immediate shutdown (CI/CD)
- Others want extended UI access (manual testing)
- One-size-fits-all doesn't work

### Minimal Implementation Cost
- Low complexity change
- Clear user requirement
- Backward compatible with defaults

## Author's Comments Summary

**@AbdelrhmanHamouda** (2025-07-08):
> "can you kindly share with me the Custom Resource yaml you deploy for the test? it would also be nice if you can kindly include any value overrides you do to the operator during deployment and which version of the operator you are currently running. reasoning: to better understand the use case and better design the solution."

**@AbdelrhmanHamouda** (2025-07-15):
Proposed the solution approach focusing on helm values for `--autostart` and `--autoquit`.

**Key Insight**: You understood the requirement and proposed a clean helm-based configuration approach that the user agreed with.

## Technical Complexity: **Low**

### Required Changes
1. **Helm Values Schema**: Add autostart/autoquit configuration
2. **Constants Refactoring**: Make command template dynamic
3. **Command Builder Logic**: Conditionally add flags
4. **Validation**: Ensure timeout values are positive integers
5. **Documentation**: Update with configuration examples

### Risks
- Low risk - well-defined requirement
- Backward compatible with sensible defaults
- No breaking changes

## Recommended Solution Approach

### Phase 1: Helm Configuration
```yaml
# values.yaml
locust:
  master:
    autostart:
      enabled: true  # default: true (current behavior)
    autoquit:
      enabled: true  # default: true (current behavior)
      timeout: 60    # default: 60 seconds (current behavior)
```

### Phase 2: Command Template Builder
```java
// Pseudocode
StringBuilder masterCommand = new StringBuilder();
masterCommand.append(masterCommandSeed);
masterCommand.append(" --master");
masterCommand.append(" --master-port=5557");
masterCommand.append(" --expect-workers=").append(workerReplicas);

if (config.getMaster().getAutostart().isEnabled()) {
    masterCommand.append(" --autostart");
}

if (config.getMaster().getAutoquit().isEnabled()) {
    masterCommand.append(" --autoquit ").append(config.getMaster().getAutoquit().getTimeout());
}

masterCommand.append(" --enable-rebalancing");
masterCommand.append(" --only-summary");
```

### Phase 3: CR-Level Override (Future)
Consider allowing per-test override:
```yaml
apiVersion: locust.io/v1
kind: LocustTest
spec:
  masterConfig:
    autoquitTimeout: 3600
```

## Related Issues
None directly.

## Estimated Effort
- **Helm Schema**: Low (0.5 day)
- **Code Refactoring**: Low (1 day)
- **Testing**: Low (1 day)
- **Documentation**: Low (0.5 day)
- **Total**: 3 days

## Success Criteria
1. Users can disable/enable `--autostart` via helm values
2. Users can disable/enable `--autoquit` via helm values
3. Users can configure `--autoquit` timeout value
4. No duplicate flags in pod args
5. Backward compatible with current defaults
6. Clear documentation with examples
7. Validation prevents negative timeout values
