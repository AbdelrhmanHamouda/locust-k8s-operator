# Security Policy

## Supported Versions

We release security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 2.x     | :white_check_mark: |
| 1.x     | :x:                |

Version 1.x (Java operator) is no longer maintained. Please upgrade to version 2.x (Go operator) to receive security updates.

## Reporting a Vulnerability

We take the security of the Locust Kubernetes Operator seriously. If you discover a security vulnerability, please report it responsibly.

### Preferred Method: GitHub Security Advisories

The preferred way to report security vulnerabilities is through [GitHub Security Advisories](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/security/advisories/new).

This allows us to:
- Discuss the vulnerability privately
- Work on a fix before public disclosure
- Coordinate the release and announcement

### Alternative: Private Email

If you prefer email or cannot use GitHub Security Advisories, please contact the project maintainers directly. You can find maintainer contact information in the project's GitHub repository.

### What to Include

When reporting a vulnerability, please include:

1. **Description**: A clear description of the vulnerability
2. **Impact**: What an attacker could achieve by exploiting it
3. **Reproduction**: Step-by-step instructions to reproduce the issue
4. **Affected versions**: Which versions are vulnerable
5. **Suggested fix** (optional): If you have ideas for remediation

### Response Timeline

- **Initial response**: Within 48 hours of report
- **Status update**: Within 7 days with assessment and next steps
- **Fix timeline**: Depends on severity; critical issues prioritized immediately

## Scope

Security reports should relate to:

- The operator's Go code (controller logic, webhook validation, resource management)
- The Helm chart (RBAC, security contexts, defaults)
- CI/CD pipeline security (supply chain, artifact integrity)
- Dependencies (Go modules, container base images)

Out of scope:
- Issues in Locust itself (report to [locustio/locust](https://github.com/locustio/locust))
- Issues in Kubernetes core (report to [kubernetes/kubernetes](https://github.com/kubernetes/kubernetes))
- General usage questions (use GitHub Discussions or Issues)

## Security Best Practices

When deploying the operator, we recommend:

1. **Least Privilege RBAC**: The Helm chart provides minimal required permissions by default
2. **Read-only Root Filesystem**: Enabled by default in the operator Pod
3. **Network Policies**: Consider adding NetworkPolicy resources to restrict operator traffic
4. **Image Verification**: Use image digests or verify signatures (cosign) for supply chain security
5. **Keep Updated**: Regularly update to the latest patch version for security fixes

## Disclosure Policy

Once a security fix is released:

1. We will publish a GitHub Security Advisory with details
2. The advisory will be linked in release notes
3. Credit will be given to the reporter (unless anonymity is requested)
4. CVE assignment will be requested for critical or high-severity issues

We follow a coordinated disclosure approach, allowing time for users to update before full public disclosure.

## Vulnerability Scanning & CVE Management

### Automated Security Scanning

We maintain continuous security monitoring through:

1. **Pull Request Scans**: Every PR is automatically scanned with Trivy for CRITICAL and HIGH vulnerabilities
   - Scans container images built during CI
   - PRs are blocked if HIGH/CRITICAL vulnerabilities are detected
   - Results uploaded to [GitHub Security tab](https://github.com/AbdelrhmanHamouda/locust-k8s-operator/security/code-scanning)

2. **Daily Scheduled Scans**: Published `latest` Docker image is scanned daily at 6 AM UTC
   - Auto-creates GitHub issues when new vulnerabilities are found
   - Provides early warning for newly disclosed CVEs
   - Results tracked in Security tab

3. **Dependabot Monitoring**: Automated dependency updates for:
   - Go modules (weekly)
   - Docker base images (weekly)
   - GitHub Actions (weekly)
   - Grouped updates for K8s, OpenTelemetry, and golang.org/x/* packages

### CVE Remediation Process

When vulnerabilities are discovered:

1. **Severity Assessment** (within 24 hours)
   - Review GitHub Security tab for CVE details
   - Assess impact on operator functionality
   - Determine if vulnerability is exploitable in our context

2. **Remediation Timeline** (based on severity)
   - **CRITICAL**: Fix within 48 hours
   - **HIGH**: Fix within 7 days
   - **MEDIUM**: Fix within 30 days
   - **LOW**: Fix in next regular release

3. **Fix & Release**
   - Update affected dependencies or Go version
   - Run full test suite to verify compatibility
   - Build and scan new image to confirm CVEs resolved
   - Release updated Helm chart with new image version
   - Document fixes in release notes

4. **Communication**
   - Security fixes mentioned prominently in release notes
   - GitHub Security Advisory created for significant CVEs
   - Users notified via GitHub Releases

### Updating Base Image Digest

The Dockerfile pins the distroless base image by SHA256 digest for reproducibility and security. To update the digest:

```bash
# Pull latest image
docker pull gcr.io/distroless/static:nonroot

# Get SHA256 digest
docker inspect gcr.io/distroless/static:nonroot --format='{{index .RepoDigests 0}}'

# Update Dockerfile FROM line with new digest
# Example: FROM gcr.io/distroless/static:nonroot@sha256:abc123...
```

Dependabot will automatically create PRs when the base image updates. Review and merge these PRs promptly to stay current with security patches.

### False Positive Management

If a reported vulnerability is a false positive or does not affect our usage:

1. Document justification in `.trivyignore` with expiry date
2. Add comment explaining why it's safe to ignore
3. Set review date (typically 3-6 months)
4. Re-evaluate on review date

Example `.trivyignore` entry:
```
# CVE-2024-1234: False positive, function not used in operator code
# Review date: 2025-06-01
CVE-2024-1234
```

### Verifying Security Status

Check current security status:

```bash
# Scan published image
docker pull lotest/locust-k8s-operator:latest
trivy image lotest/locust-k8s-operator:latest --severity CRITICAL,HIGH

# View GitHub Security alerts
# Visit: https://github.com/AbdelrhmanHamouda/locust-k8s-operator/security/code-scanning
```
