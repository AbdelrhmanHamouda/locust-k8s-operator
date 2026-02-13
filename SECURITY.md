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
