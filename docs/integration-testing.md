# Integration Testing Guide

This document describes the comprehensive integration testing setup for the Locust K8s Operator, which validates the complete end-to-end functionality beyond unit tests.

## Overview

The integration test suite performs the following workflow:
1. **Build** - Creates the operator Docker image
2. **Package** - Packages the Helm chart
3. **Deploy** - Spins up a K8s cluster (K3s) and installs the operator
4. **Test** - Deploys a LocustTest CR and validates operator behavior
5. **Validate** - Ensures Locust master/workers are running correctly
6. **Cleanup** - Removes all resources and tears down the environment

## Architecture

### Test Framework
- **Testing Framework**: JUnit 5 with Testcontainers
- **Kubernetes Cluster**: K3s via Testcontainers for local development, KinD in CI environment
- **Build System**: Gradle with custom integration test source set
- **Container Management**: Docker with Jib plugin for image building

### Test Structure
```
src/integrationTest/
├── java/com/locust/operator/
│   └── LocustOperatorIntegrationTest.java    # Main integration test
└── resources/
    └── application-test.yml                   # Test configuration
```

## Prerequisites

### Local Development
- **Docker**: Running Docker daemon
- **Java 21**: Required for building the operator
- **Helm 3.x**: For chart packaging and installation
- **Gradle**: Uses project's gradle wrapper

### CI/CD (GitHub Actions)
- Uses Ubuntu latest runner
- Automatically installs all dependencies
- Runs on PR and push to main branch

## Running Integration Tests

### Option 1: Using the Integration Test Script (Recommended)
```bash
# Make script executable (first time only)
chmod +x scripts/run-integration-test.sh

# Run integration tests
./scripts/run-integration-test.sh
```

The script performs several helpful functions:
- Checks prerequisites (Docker, Helm, Java)
- Cleans up previous runs and Docker resources
- Runs the integration tests with proper error handling
- Provides detailed error reporting and logs
- Shows test results and report locations

### Option 2: Using Gradle Directly
```bash
# Run integration tests
./gradlew integrationTest -PrunIntegrationTests

# Run with verbose output
./gradlew integrationTest -PrunIntegrationTests --info

# Run specific test class
./gradlew integrationTest -PrunIntegrationTests --tests="LocustOperatorIntegrationTest"
```

### Option 3: In CI/CD
Integration tests run automatically in GitHub Actions:
- On pull requests to `main` or `master`
- On pushes to `main` or `master`
- Can be triggered manually via `workflow_dispatch`

## Test Scenarios

### Test 1: Operator Deployment
- Creates operator namespace
- Installs operator via Helm chart
- Validates operator deployment is ready
- Verifies operator pod is running

### Test 2: LocustTest Deployment
- Creates test namespace and ConfigMap with simple Locust script
- Deploys LocustTest custom resource
- Validates master and worker deployments are created
- Ensures all pods reach Running state

### Test 3: LocustTest Execution
- Verifies Locust master web interface starts
- Checks master logs for successful initialization
- Validates workers connect to master
- Confirms test environment is functional

### Test 4: Cleanup
- Deletes LocustTest custom resource
- Verifies all managed resources are cleaned up
- Uninstalls operator
- Validates complete cleanup

## Configuration

### Integration Test Configuration
Located in `gradle/integration-test.gradle`:
- Defines separate source set for integration tests
- Configures dependencies (Testcontainers, Kubernetes client, etc.)
- Sets up test reporting and timeouts
- Links to main build pipeline

### Test Application Configuration
Located in `src/integrationTest/resources/application-test.yml`:
- Configures logging levels for test visibility
- Sets timeouts for different test phases
- Defines resource locations and image names

### CI Configuration
Located in `.github/workflows/integration-test.yml`:
- GitHub Actions workflow for automated testing
- Includes caching for Gradle and Docker layers
- Uploads test results as artifacts
- Uses **KinD (Kubernetes in Docker)** cluster with custom configuration in `.github/kind-config.yaml`
- Uses Helm 3.12.0 for chart installation

## Sample LocustTest Resource

The integration test creates this sample LocustTest CR:

```yaml
apiVersion: locust.io/v1
kind: LocustTest
metadata:
  name: integration-test
  namespace: locust-tests
spec:
  masterConfig:
    replicas: 1
    image: locustio/locust:2.15.1
    resources:
      requests:
        memory: "128Mi"
        cpu: "100m"
      limits:
        memory: "256Mi"
        cpu: "200m"
  workerConfig:
    replicas: 2
    image: locustio/locust:2.15.1
    resources:
      requests:
        memory: "128Mi"
        cpu: "100m"
      limits:
        memory: "256Mi"
        cpu: "200m"
  configMap: locust-test-scripts
```

## Test Reports and Artifacts

### Local Testing
- **HTML Report**: `build/reports/integration-tests/index.html`
- **JUnit XML**: `build/test-results/integration-test/`
- **Logs**: `/tmp/locust-integration-test-{timestamp}.log`

### CI Testing
- Test results uploaded as GitHub Actions artifacts
- Available for download from the Actions run page
- Includes both HTML reports and raw XML results

## Troubleshooting

### Common Issues

#### Docker Permission Errors
```bash
# On Linux, ensure user is in docker group
sudo usermod -aG docker $USER
# Then logout and login again
```

#### K3s Container Startup Issues
- Ensure Docker has enough resources (4GB+ RAM recommended)
- Check Docker daemon is running: `docker info`
- Verify no conflicting containers: `docker ps -a`

#### Helm Chart Packaging Failures
- Ensure Helm is installed: `helm version`
- Check chart syntax: `helm lint charts/locust-k8s-operator`
- Verify chart dependencies: `helm dependency list charts/locust-k8s-operator`

#### Integration Test Timeouts
- Tests have generous timeouts but may need adjustment for slower systems
- Modify timeouts in `application-test.yml` if needed
- Check system resources during test execution

### Debug Mode
Enable debug logging by setting:
```yaml
logger:
  levels:
    com.locust: DEBUG
    org.testcontainers: DEBUG
```

### Manual Debugging
If tests fail, you can manually inspect the K3s cluster:
1. The test creates temporary kubeconfig files
2. Look for log messages indicating kubeconfig location
3. Use `kubectl` with the temporary kubeconfig to inspect cluster state

## Performance Considerations

### Resource Requirements
- **Memory**: ~4GB available RAM recommended
- **CPU**: 2+ cores for reasonable performance
- **Disk**: ~10GB for images and temporary files
- **Network**: Internet access for pulling images

### Execution Time
- Full test suite: ~10-15 minutes
- Individual test phases:
  - Cluster startup: ~2-3 minutes
  - Image building: ~3-5 minutes
  - Deployment validation: ~2-3 minutes
  - Test execution: ~2-3 minutes
  - Cleanup: ~1-2 minutes

## Future Enhancements

### Planned Improvements
- [ ] Multi-scenario testing (different LocustTest configurations)
- [ ] Performance benchmarking integration
- [ ] Integration with libConfigMap feature testing
- [ ] Cross-platform testing (ARM64 support)
- [ ] Parallel test execution for faster CI

### Extension Points
- Add custom test scenarios in separate test classes
- Extend with custom Kubernetes resources validation
- Integrate with monitoring and observability testing
- Add chaos engineering tests for resilience validation

## Related Documentation
- [How It Works](how_does_it_work.md) - Operator architecture overview
- [Contributing](contribute.md) - Development guidelines
- [LibConfigMap Feature](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/blob/master/LIBCONFIGMAP_FEATURE_IMPLEMENTATION.md) - Feature implementation details
