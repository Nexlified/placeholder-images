# Contributing to Grout

Thank you for your interest in contributing to Grout! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Features](#suggesting-features)

## Code of Conduct

This project adheres to a Code of Conduct that all contributors are expected to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/grout.git
   cd grout
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/Nexlified/grout.git
   ```

## Development Setup

### Prerequisites

- Go 1.24 or later
- Docker and Docker Compose (optional, for container-based development)
- Git

### Local Development

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Run the server**:
   ```bash
   go run ./cmd/grout
   ```
   The server will start on `http://localhost:8080`

3. **Using Docker Compose**:
   ```bash
   docker compose up --build
   ```

### Project Structure

```
grout/
├── cmd/grout/          # Application entry point
├── internal/           # Private application code
│   ├── config/        # Configuration management
│   ├── handlers/      # HTTP request handlers
│   ├── render/        # Image rendering logic
│   └── utils/         # Shared utilities
├── static/            # Static files (robots.txt, sitemap.xml)
├── .github/           # GitHub workflows and templates
└── docs/              # Additional documentation
```

## Making Changes

### Branching Strategy

1. **Create a feature branch** from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
   or for bug fixes:
   ```bash
   git checkout -b fix/bug-description
   ```

2. **Keep your branch up to date**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

### Commit Messages

Write clear, descriptive commit messages following these guidelines:

- Use the present tense ("Add feature" not "Added feature")
- Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit the first line to 72 characters or less
- Reference issues and pull requests after the first line

Example:
```
Add gradient support for placeholder backgrounds

- Implement gradient parsing in render package
- Update handlers to accept comma-separated colors
- Add tests for gradient rendering
- Update documentation with examples

Fixes #123
```

## Coding Standards

### Go Best Practices

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Run `gofmt` to format your code (or use `go fmt ./...`)
- Run `go vet ./...` to check for common mistakes
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused

### Code Style

- **Imports**: Group standard library, external packages, and internal packages separately
- **Error Handling**: Always handle errors explicitly; wrap with context using `fmt.Errorf`
- **Naming**:
  - Use `camelCase` for unexported identifiers
  - Use `PascalCase` for exported identifiers
  - Use descriptive names (avoid single letters except for loops)

### Package Organization

- Keep code in `internal/` to prevent external imports
- Separate concerns: config, rendering, HTTP handling, utilities
- Use dependency injection for testability

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run tests with race detection
go test -race ./...
```

### Writing Tests

- Write table-driven tests for multiple scenarios
- Test both success and failure cases
- Use meaningful test names that describe what's being tested
- Aim for at least 80% code coverage
- Mock external dependencies

Example:
```go
func TestParseHexColor(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected color.Color
    }{
        {"3-digit hex", "f00", color.RGBA{255, 0, 0, 255}},
        {"6-digit hex", "ff0000", color.RGBA{255, 0, 0, 255}},
        {"invalid", "xyz", color.RGBA{128, 128, 128, 255}},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ParseHexColor(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Test Requirements

- All new features must include tests
- Bug fixes should include regression tests
- Maintain or improve existing code coverage
- Tests must pass before submitting a PR

## Submitting Changes

### Pull Request Process

1. **Update your branch**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests and linting**:
   ```bash
   go test ./...
   go fmt ./...
   go vet ./...
   ```

3. **Push your changes**:
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Create a Pull Request** on GitHub with:
   - Clear title describing the change
   - Detailed description of what changed and why
   - Reference to related issues (e.g., "Fixes #123")
   - Screenshots for UI changes (if applicable)

5. **Address review feedback**:
   - Make requested changes
   - Push additional commits to your branch
   - Respond to reviewer comments

### Pull Request Checklist

- [ ] Code follows the project's style guidelines
- [ ] Self-review of code completed
- [ ] Comments added for complex logic
- [ ] Documentation updated (if needed)
- [ ] Tests added/updated and passing
- [ ] No new warnings introduced
- [ ] Related issues linked in PR description

## Reporting Bugs

Found a bug? Please help us fix it!

1. **Check existing issues** to avoid duplicates
2. **Use the bug report template** when creating a new issue
3. **Include**:
   - Clear description of the bug
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)
   - Screenshots or error messages (if applicable)

## Suggesting Features

Have an idea for a new feature?

1. **Check existing issues** to see if it's already suggested
2. **Use the feature request template** when creating a new issue
3. **Include**:
   - Clear description of the feature
   - Use cases and benefits
   - Proposed implementation (optional)
   - Examples from other projects (if applicable)

## Development Guidelines

### Adding New Endpoints

1. Add handler method to `handlers.Service`
2. Register route in `RegisterRoutes()`
3. Parse query parameters using `utils.ParseIntOrDefault()`
4. Implement caching with appropriate cache keys
5. Use `serveImage()` helper for consistent behavior
6. Add comprehensive tests
7. Update API documentation in README.md

### Adding Configuration Options

1. Add constant to `internal/config/config.go`
2. Add field to `ServerConfig` struct (if runtime-configurable)
3. Add environment variable parsing in `LoadServerConfig()`
4. Add CLI flag support (if needed)
5. Document in README.md

### Performance Considerations

- Use the LRU cache for expensive operations
- Set appropriate cache keys to maximize hit rate
- Consider memory usage for large operations
- Profile code for performance bottlenecks

## Security

If you discover a security vulnerability, please follow our [Security Policy](SECURITY.md). **Do not** open a public issue.

## Questions?

- Open a discussion in GitHub Discussions
- Use the question/help issue template
- Check existing documentation in the repository

## License

By contributing to Grout, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Grout! 🎨
