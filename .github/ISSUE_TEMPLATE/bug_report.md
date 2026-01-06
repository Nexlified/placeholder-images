---
name: Bug Report
about: Report a bug to help us improve Grout
title: '[BUG] '
labels: ['bug']
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## Steps to Reproduce

1. Go to '...'
2. Execute command '...'
3. Send request to '...'
4. See error

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

A clear and concise description of what actually happened.

## Screenshots/Error Messages

If applicable, add screenshots or error messages to help explain your problem.

```
Paste error messages or logs here
```

## Environment

- **Grout Version**: [e.g., v1.0.0, latest, commit hash]
- **Deployment Method**: [e.g., Docker, Docker Compose, binary]
- **Go Version** (if building from source): [e.g., 1.24]
- **OS**: [e.g., Ubuntu 22.04, macOS 13, Windows 11]
- **Architecture**: [e.g., amd64, arm64]

## Configuration

```yaml
# If using Docker Compose, paste your docker-compose.yml
# Or list environment variables and flags used
ADDR: ":8080"
CACHE_SIZE: "2000"
```

## Request Details (if applicable)

```bash
# Example request that produces the bug
curl "http://localhost:8080/avatar/John+Doe?size=256"
```

## Additional Context

Add any other context about the problem here. This might include:
- Does this happen consistently or intermittently?
- Did this work in a previous version?
- Any workarounds you've found?

## Possible Solution (Optional)

If you have suggestions on how to fix the bug, please describe them here.
