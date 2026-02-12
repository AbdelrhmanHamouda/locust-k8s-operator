---
title: CI/CD Integration (15 minutes)
description: Automate load tests in GitHub Actions and GitLab CI pipelines
tags:
  - tutorial
  - ci-cd
  - automation
  - github-actions
  - gitlab-ci
---

# CI/CD Integration (15 minutes)

Automate your load tests to run on every deployment or on a schedule.

## What you'll learn

- How to run load tests in CI/CD pipelines
- How to create unique test runs per pipeline execution
- How to collect and store test results
- How to fail a pipeline on performance regression

## Prerequisites

- Completed the [Your First Load Test](first-load-test.md) tutorial
- A Kubernetes cluster accessible from CI (kubeconfig or service account)
- GitHub repository (for GitHub Actions example)

## The scenario

You want nightly load tests against your staging environment, plus on-demand tests before releases. Tests should fail the pipeline if error rate exceeds 1%.

## Step 1: Prepare the test script

We'll reuse the `ecommerce_test.py` from Tutorial 1. Store it in your repository:

```
your-repo/
├── .github/
│   └── workflows/
│       └── load-test.yaml
└── tests/
    └── locust/
        └── ecommerce_test.py
```

The test script should be checked into your repository at `tests/locust/ecommerce_test.py` (same content from Tutorial 1). This ensures version control and consistency across pipeline runs.

## Step 2: Create the GitHub Actions workflow

Create `.github/workflows/load-test.yaml`:

```yaml
name: Nightly Load Test

on:
  schedule:
    - cron: '0 2 * * 1'  # Every Monday at 2 AM UTC
  workflow_dispatch:  # Allow manual trigger from GitHub UI

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Set up kubectl
        uses: azure/setup-kubectl@v3
        with:
          version: 'v1.28.0'

      - name: Configure kubeconfig
        run: |
          # Create .kube directory
          mkdir -p $HOME/.kube
          # Write kubeconfig from GitHub secret
          echo "${{ secrets.KUBECONFIG }}" > $HOME/.kube/config
          # Verify connectivity
          kubectl cluster-info

      - name: Create/update test ConfigMap
        run: |
          # Use --dry-run + kubectl apply for idempotency
          kubectl create configmap ecommerce-test \
            --from-file=tests/locust/ecommerce_test.py \
            --dry-run=client -o yaml | kubectl apply -f -

      - name: Deploy LocustTest with unique name
        run: |
          # Generate unique test name with timestamp
          TEST_NAME="ecommerce-ci-$(date +%Y%m%d-%H%M%S)"

          kubectl apply -f - <<EOF
          apiVersion: locust.io/v2
          kind: LocustTest
          metadata:
            name: ${TEST_NAME}
          spec:
            image: locustio/locust:2.20.0
            testFiles:
              configMapRef: ecommerce-test
            master:
              command: |
                --locustfile /lotest/src/ecommerce_test.py
                --host https://api.staging.example.com
                --users 100
                --spawn-rate 10
                --run-time 5m
            worker:
              command: "--locustfile /lotest/src/ecommerce_test.py"
              replicas: 5
          EOF

          # Store test name for later steps
          echo "TEST_NAME=${TEST_NAME}" >> $GITHUB_ENV

      - name: Wait for test completion
        run: |
          # Wait up to 10 minutes for test to succeed
          kubectl wait --for=jsonpath='{.status.phase}'=Succeeded \
            locusttest/${TEST_NAME} --timeout=10m

      - name: Collect test results
        if: always()  # Run even if test fails
        run: |
          # Get master pod logs
          kubectl logs job/${TEST_NAME}-master > results.log

          # Get test status YAML
          kubectl get locusttest ${TEST_NAME} -o yaml > test-status.yaml

          # Display summary
          echo "=== Test Summary ==="
          kubectl get locusttest ${TEST_NAME}

      - name: Check for performance regression
        run: |
          # Extract final statistics from master logs
          FAILURE_RATE=$(kubectl logs job/${TEST_NAME}-master | \
            grep -oP 'Total.*Failures.*\K[\d.]+%' | tail -1 | sed 's/%//')

          echo "Failure rate: ${FAILURE_RATE}%"

          # Fail pipeline if error rate > 1%
          if (( $(echo "$FAILURE_RATE > 1.0" | bc -l) )); then
            echo "ERROR: Failure rate ${FAILURE_RATE}% exceeds threshold of 1%"
            exit 1
          fi

          echo "✓ Performance acceptable: ${FAILURE_RATE}% failures"

      - name: Upload test artifacts
        if: always()
        uses: actions/upload-artifact@v4
        with:
          name: load-test-results-${{ env.TEST_NAME }}
          path: |
            results.log
            test-status.yaml
          retention-days: 30
```

**Key workflow features:**

- **Scheduled execution**: `cron: '0 2 * * 1'` runs every Monday at 2 AM
- **Manual trigger**: `workflow_dispatch` allows on-demand runs from GitHub UI
- **Unique test names**: `$(date +%Y%m%d-%H%M%S)` prevents name conflicts
- **Idempotent ConfigMap**: `--dry-run=client -o yaml | kubectl apply` updates existing ConfigMap
- **Result collection**: Logs and YAML saved as GitHub artifacts
- **Regression detection**: Pipeline fails if error rate exceeds 1%

## Step 3: Configure GitHub secrets

Your workflow needs a kubeconfig to access the cluster. Add it as a GitHub secret:

1. Get your kubeconfig:
   ```bash
   cat ~/.kube/config | base64
   ```

2. In GitHub: Go to **Settings** → **Secrets and variables** → **Actions** → **New repository secret**

3. Create secret `KUBECONFIG` with the base64-encoded content

**Security note:** For production, use a service account with minimal permissions instead of full admin kubeconfig.

## Step 4: Run and verify

### Trigger the workflow manually

1. Go to **Actions** tab in GitHub
2. Select **Nightly Load Test** workflow
3. Click **Run workflow** → **Run workflow**

### Monitor execution

Watch the workflow run in real-time. Check each step's output:

- ✓ ConfigMap created/updated
- ✓ LocustTest deployed with unique name
- ✓ Test completed successfully
- ✓ Performance within acceptable limits

### Check artifacts

After completion (or failure), download artifacts:

1. Click on the workflow run
2. Scroll to **Artifacts** section
3. Download `load-test-results-ecommerce-ci-YYYYMMDD-HHMMSS.zip`

The artifact contains:
- `results.log` — Full Locust master output with statistics
- `test-status.yaml` — Complete LocustTest CR status

## Step 5: Make tests fail on regression

The workflow already includes regression detection in the "Check for performance regression" step. It:

1. **Extracts error rate** from master logs using `grep`
2. **Compares to threshold** (1% in this example)
3. **Fails pipeline** with `exit 1` if threshold exceeded

**Customizing thresholds:**

```bash
# Fail on error rate > 1%
if (( $(echo "$FAILURE_RATE > 1.0" | bc -l) )); then
  exit 1
fi

# Or fail on response time > 500ms
AVG_RESPONSE=$(kubectl logs job/${TEST_NAME}-master | \
  grep -oP 'Average response time.*\K[\d.]+' | tail -1)
if (( $(echo "$AVG_RESPONSE > 500" | bc -l) )); then
  exit 1
fi
```

## Alternative: GitLab CI

For GitLab users, here's an equivalent `.gitlab-ci.yml`:

??? example "GitLab CI configuration"
    ```yaml
    nightly-load-test:
      stage: test
      image: bitnami/kubectl:1.28
      script:
        # Configure kubectl
        - mkdir -p ~/.kube
        - echo "$KUBECONFIG" > ~/.kube/config
        - kubectl cluster-info

        # Create/update ConfigMap
        - kubectl create configmap ecommerce-test
            --from-file=tests/locust/ecommerce_test.py
            --dry-run=client -o yaml | kubectl apply -f -

        # Deploy LocustTest
        - TEST_NAME="ecommerce-ci-$(date +%Y%m%d-%H%M%S)"
        - |
          kubectl apply -f - <<EOF
          apiVersion: locust.io/v2
          kind: LocustTest
          metadata:
            name: ${TEST_NAME}
          spec:
            image: locustio/locust:2.20.0
            testFiles:
              configMapRef: ecommerce-test
            master:
              command: |
                --locustfile /lotest/src/ecommerce_test.py
                --host https://api.staging.example.com
                --users 100
                --spawn-rate 10
                --run-time 5m
            worker:
              command: "--locustfile /lotest/src/ecommerce_test.py"
              replicas: 5
          EOF

        # Wait for completion
        - kubectl wait --for=jsonpath='{.status.phase}'=Succeeded
            locusttest/${TEST_NAME} --timeout=10m

        # Collect results
        - kubectl logs job/${TEST_NAME}-master > results.log
        - kubectl get locusttest ${TEST_NAME} -o yaml > test-status.yaml

        # Check regression
        - |
          FAILURE_RATE=$(grep -oP 'Total.*Failures.*\K[\d.]+%' results.log | tail -1 | sed 's/%//')
          if (( $(echo "$FAILURE_RATE > 1.0" | bc -l) )); then
            echo "ERROR: Failure rate ${FAILURE_RATE}% exceeds threshold"
            exit 1
          fi

      artifacts:
        when: always
        paths:
          - results.log
          - test-status.yaml
        expire_in: 30 days

      only:
        - schedules
        - web  # Manual trigger from GitLab UI
    ```

    **GitLab-specific setup:**

    1. Add `KUBECONFIG` as a GitLab CI/CD variable (Settings → CI/CD → Variables)
    2. Create a pipeline schedule (CI/CD → Schedules → New schedule)
    3. Set cron expression: `0 2 * * 1` for weekly Monday 2 AM runs

## What you learned

✓ How to run Kubernetes-based load tests in CI/CD pipelines
✓ How to create unique test names for traceability
✓ How to collect and store test results as artifacts
✓ How to fail pipelines on performance regression
✓ How to configure scheduled and manual test execution

## Next steps

- [Production Deployment](production-deployment.md) — Configure production-grade load tests
- [Configure resources](../how-to-guides/configuration/configure-resources.md) — Optimize pod resource allocation
- [Set up OpenTelemetry](../how-to-guides/observability/configure-opentelemetry.md) — Export metrics for long-term analysis
