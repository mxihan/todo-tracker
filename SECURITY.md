# Security Policy

## Supported Versions

We actively support the following versions of TODO Tracker with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability in TODO Tracker, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Security Advisory** (Preferred)
   - Go to the [Security Advisories](https://github.com/mxihan/todo-tracker/security/advisories) page
   - Click "Report a vulnerability"
   - Fill in the details

2. **Email**
   - Send an email to: security@todo-tracker.dev
   - Subject: [Security] Vulnerability in TODO Tracker
   - Include detailed description and steps to reproduce

### What to Include

Please include the following information in your report:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Affected versions
- Suggested fix (if any)

### Response Timeline

We aim to respond to security reports within:

- **Initial Response**: 48 hours
- **Status Update**: 7 days
- **Fix Release**: Depends on severity

| Severity | Target Fix Time |
|----------|----------------|
| Critical | 24-48 hours |
| High | 1 week |
| Medium | 2 weeks |
| Low | Next release |

### Disclosure Policy

We follow coordinated disclosure:

1. Report received and acknowledged
2. Vulnerability verified
3. Fix developed and tested
4. Security advisory prepared
5. Fix released
6. Advisory published (after users have had time to update)

## Security Best Practices

When using TODO Tracker, please follow these security practices:

### Configuration Security

- Do not commit sensitive configuration to version control
- Use environment variables for sensitive settings
- Review `.todoignore` to ensure sensitive files are excluded

### Git Repository Access

- TODO Tracker uses `git blame` which requires read access to your repository
- Ensure proper access controls are in place
- Be aware that TODO metadata may include author information

### CI/CD Integration

- Use secrets for API tokens in GitHub Actions
- Limit repository access for CI/CD jobs
- Review workflow permissions

### Self-Hosted Deployments

- Keep your deployment updated
- Use the latest stable release
- Monitor security advisories

## Security Features

TODO Tracker includes the following security features:

- **No external network calls**: All operations are local
- **No telemetry**: We don't collect any usage data
- **Sandboxed file access**: Only scans specified directories
- **Safe git operations**: Read-only git operations

## Known Security Considerations

### Git Blame Information

TODO Tracker extracts author information from git history. Be aware that:

- Author names and emails from git history will be processed
- This information may appear in reports
- Ensure your git history doesn't contain sensitive information

### File System Access

TODO Tracker reads source code files. Consider:

- Using `.todoignore` to exclude sensitive files
- Running with minimal necessary permissions
- Reviewing scanned directories

## Contact

For security concerns, contact:

- **Security Email**: security@todo-tracker.dev
- **PGP Key**: [Available upon request]
- **GitHub Security**: https://github.com/mxihan/todo-tracker/security

---

Thank you for helping keep TODO Tracker and its users safe!