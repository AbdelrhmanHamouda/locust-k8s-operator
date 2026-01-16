# Issue #253: Master-Worker Connection Delay (2.5+ Minutes)

**Priority**: P3 - Low  
**Issue URL**: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/issues/253  
**Status**: Open - Pending Feedback  
**Created**: 2025-09-22  

## Problem
Workers take 2.5+ minutes to connect to master. When connection isn't immediate, it consistently takes exactly ~150-160 seconds.

## Your Response
Requested more context:
- Operator version
- Locust version
- Example CR
- Cluster context/setup

## Technical Implications
- Likely environment-specific (networking, DNS, service discovery)
- Not reproducible without user's specific setup
- May be related to Kubernetes networking configuration

## Why Low Priority
- Only one report
- Needs user feedback to debug
- Potentially environment-specific
- Immediate connections work fine

## Action Required
Wait for user to provide requested debugging information.
