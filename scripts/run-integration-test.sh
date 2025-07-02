#!/bin/bash

# Locust K8s Operator Integration Test Runner
# This script runs the complete integration test suite

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LOG_FILE="/tmp/locust-integration-test-$(date +%Y%m%d-%H%M%S).log"

echo -e "${GREEN}🚀 Starting Locust K8s Operator Integration Test${NC}"
echo -e "${YELLOW}📁 Project Root: $PROJECT_ROOT${NC}"
echo -e "${YELLOW}📝 Log File: $LOG_FILE${NC}"
echo ""

# Function to print colored output
print_step() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Function to check prerequisites
check_prerequisites() {
    print_step "Checking prerequisites..."

    # Check Docker
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed or not in PATH"
        exit 1
    fi

    # Check Docker daemon
    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        print_warning "Please start Docker and try again"
        print_warning "On macOS: Start Docker Desktop"
        print_warning "On Linux: sudo systemctl start docker"
        exit 1
    fi

    print_step "Docker daemon is running"

    # Check Helm
    if ! command -v helm &> /dev/null; then
        print_error "Helm is not installed or not in PATH"
        exit 1
    fi

    # Check Java
    if ! command -v java &> /dev/null; then
        print_error "Java is not installed or not in PATH"
        exit 1
    fi

    # Check Gradle wrapper
    if [ ! -f "$PROJECT_ROOT/gradlew" ]; then
        print_error "Gradle wrapper not found in project root"
        exit 1
    fi

    print_step "All prerequisites check passed!"
}

# Function to clean up previous runs
cleanup_previous_runs() {
    print_step "Cleaning up previous runs..."

    # Remove any existing integration test containers
    if docker ps -a --filter "name=locust" --format "{{.ID}}" | grep -q .; then
        print_warning "Removing existing locust containers..."
        docker ps -a --filter "name=locust" --format "{{.ID}}" | xargs docker rm -f || true
    fi

    # Clean up testcontainers
    if docker ps -a --filter "label=org.testcontainers=true" --format "{{.ID}}" | grep -q .; then
        print_warning "Removing testcontainers..."
        docker ps -a --filter "label=org.testcontainers=true" --format "{{.ID}}" | xargs docker rm -f || true
    fi

    # Clean up any leftover test images - more thorough pattern matching
    if docker images "locust-k8s-operator*" --format "{{.ID}}" | grep -q .; then
        print_warning "Removing leftover operator test images..."
        docker images "locust-k8s-operator*" --format "{{.ID}}" | xargs docker rmi -f || true
    fi

    # Clean Gradle build cache completely and force daemon restart
    cd "$PROJECT_ROOT"
    print_warning "Cleaning Gradle cache and stopping daemons..."
    ./gradlew --stop || true  # Stop gradle daemon to release any locks
    ./gradlew clean cleanBuildCache || true  # More thorough cleaning

    # Remove any cached test results
    rm -rf "$PROJECT_ROOT/build/test-results" || true
    rm -rf "$PROJECT_ROOT/build/reports/integration-tests" || true

    print_step "Cleanup completed!"
}

# Function to run the integration test
run_integration_test() {
    print_step "Running integration test suite..."

    cd "$PROJECT_ROOT"

    # Run integration test with proper error handling
    # Use PIPESTATUS to capture the actual gradle exit code, not tee's exit code
    set -o pipefail
    ./gradlew integrationTest -PrunIntegrationTests 2>&1 | tee "$LOG_FILE"
    local gradle_exit_code=$?
    set +o pipefail

    if [ $gradle_exit_code -eq 0 ]; then
        print_step "Integration test suite completed successfully!"
        return 0
    else
        print_error "Integration test suite failed with exit code: $gradle_exit_code"
        print_warning "Check the log file for details: $LOG_FILE"

        # Show last few lines of the log for immediate feedback
        echo ""
        print_warning "Last 20 lines of the test output:"
        tail -n 20 "$LOG_FILE" || true
        echo ""

        return 1
    fi
}

# Function to show test results
show_test_results() {
    print_step "Integration test results:"
    echo ""

    # Show test report location
    if [ -d "$PROJECT_ROOT/build/reports/integration-tests" ]; then
        echo -e "${GREEN}📊 HTML Report: $PROJECT_ROOT/build/reports/integration-tests/index.html${NC}"
    fi

    # Show test results location
    if [ -d "$PROJECT_ROOT/build/test-results/integration-test" ]; then
        echo -e "${GREEN}📋 XML Results: $PROJECT_ROOT/build/test-results/integration-test/${NC}"
    fi

    # Show log file
    echo -e "${GREEN}📝 Detailed Logs: $LOG_FILE${NC}"
    echo ""
}

# Function to handle cleanup on exit
cleanup_on_exit() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        print_error "Integration test failed with exit code $exit_code"
        print_warning "Performing emergency cleanup..."

        # Kill any hanging processes
        pkill -f "locust-integration-test" || true

        # Clean up Docker resources
        docker ps -a --filter "name=locust-integration-test" --format "{{.ID}}" | xargs -r docker rm -f || true
    fi
}

# Set up cleanup trap
trap cleanup_on_exit EXIT

# Main execution
main() {
    echo -e "${GREEN}🔍 Locust K8s Operator Integration Test Runner${NC}"
    echo "================================================="
    echo ""

    check_prerequisites
    cleanup_previous_runs

    print_step "Starting integration test..."
    echo ""

    if run_integration_test; then
        echo ""
        print_step "🎉 Integration test completed successfully!"
        show_test_results
        exit_code=0
    else
        echo ""
        print_error "💥 Integration test failed!"
        show_test_results
        exit_code=1
    fi

    echo ""
    echo "================================================="
    print_step "Integration test runner finished"

    exit $exit_code
}

# Run main function
main "$@"
