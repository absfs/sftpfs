# Contributing to sftpfs

Thank you for your interest in contributing to sftpfs! This document provides guidelines and instructions for contributing.

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for integration tests)
- Git

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/absfs/sftpfs.git
   cd sftpfs
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Verify the build:
   ```bash
   go build ./...
   ```

## Running Tests

### Unit Tests

Unit tests use mock interfaces and don't require an SFTP server:

```bash
go test -v ./...
```

### Integration Tests

Integration tests require a running SFTP server:

```bash
# Start the SFTP server
docker-compose up -d

# Wait for the server to be ready (usually a few seconds)
sleep 5

# Run integration tests
go test -v -tags=integration ./...

# Stop the server when done
docker-compose down
```

### Coverage

Check and view coverage:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage summary
go tool cover -func=coverage.out
```

Target coverage: >40% for unit tests, >65% overall.

### Benchmarks

Run performance benchmarks:

```bash
go test -bench=. -benchmem ./...
```

## Code Style

### Go Conventions

- Follow standard Go conventions and idioms
- Use `gofmt` to format code
- Use `go vet` to check for issues
- Run `golint` for style suggestions (optional but recommended)

### Error Handling

- Always check and handle errors
- Return errors with context when appropriate
- Use standard library errors or `errors.Is`/`errors.As` for error checking

### Testing

- Write tests for new functionality
- Use table-driven tests where appropriate
- Name test functions descriptively: `TestFunctionName_Scenario`
- Use mocks for unit tests (see `internal/mocks` package)

## Project Structure

```
sftpfs/
├── sftpfs.go           # Main FileSystem implementation
├── sftpfile.go         # File wrapper implementation
├── interfaces.go       # Internal interfaces for mocking
├── wrappers.go         # Wrappers around real SFTP client
├── sftpfs_test.go      # Unit tests
├── compliance_test.go  # absfs.Filer compliance tests
├── benchmark_test.go   # Performance benchmarks
├── integration_test.go # Integration tests (build tag: integration)
├── docker-compose.yml  # Docker setup for integration tests
├── internal/
│   └── mocks/          # Mock implementations for testing
└── go.mod
```

## Mock Infrastructure

The project uses custom mocks in `internal/mocks` for testing. When adding new functionality:

1. Add any new methods to the interface in `interfaces.go`
2. Implement the method in the wrapper in `wrappers.go`
3. Add mock implementation in `internal/mocks/mocks.go`
4. Update `mockSFTPClient` in `sftpfs_test.go`
5. Write tests for the new functionality

## Submitting Changes

### Pull Request Process

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Ensure all tests pass
5. Update documentation if needed
6. Submit a pull request

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb (Add, Fix, Update, Remove, etc.)
- Reference issues if applicable

### Code Review

- All changes require review before merging
- Address review feedback promptly
- Keep PRs focused and reasonably sized

## Security

### Reporting Security Issues

If you discover a security vulnerability, please report it privately to the maintainers rather than opening a public issue.

### Security Best Practices

- Never commit credentials or secrets
- Be aware that the current implementation skips host key verification
- Consider security implications when adding new features

## Questions?

If you have questions about contributing, feel free to open an issue for discussion.
