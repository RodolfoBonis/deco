# CI/CD Guide

## Overview

The deco framework uses CI/CD flows optimized for Go packages, focusing on multi-platform testing, code quality, security verification, and automated releases.

## Available Workflows

### 1. CI Package (`.github/workflows/ci-package.yaml`)

**Trigger:** Push to `main` or Pull Requests

**Jobs:**
- **test**: Tests on multiple platforms and Go versions
- **lint**: Linting with golangci-lint, goimports, go vet
- **security**: Vulnerability checking with govulncheck
- **build**: Binary build on multiple platforms
- **validate**: go.mod and dependencies validation

### 2. CD Package (`.github/workflows/cd-package.yaml`)

**Trigger:** After successful CI Package on `main` branch

**Jobs:**
- **build_and_release**: Build, versioning and release creation
- **publish_to_go_proxy**: Publication to Go Proxy
- **generate_documentation**: Automatic documentation update

### 3. Release Drafter (`.github/workflows/release-drafter.yml`)

**Trigger:** Push to `main` or Pull Requests

**Jobs:**
- **update_release_draft**: Automatic release notes generation

## Configurations

### GolangCI-Lint (`.golangci.yml`)

```yaml
# Enabled linters
- gofmt, goimports, govet
- staticcheck, gosimple, ineffassign
- unused, misspell, gosec
- errcheck, gocritic

# Specific settings
- Timeout: 5 minutes
- Go version: 1.23
- Exclusions for test files
```

### Codecov (`.codecov.yml`)

```yaml
# Coverage settings
- Target: 80%
- Threshold: 5%
- Ignore: main.go, examples, tests
```

## Release Process

### 1. Automatic Versioning

```bash
# Automatic version increment
./.config/scripts/increment_version.sh
```

### 2. Multi-platform Build

```bash
# Build for Linux, Windows, macOS
go build -ldflags="-s -w -X main.version=$VERSION" -o deco ./cmd/deco
```

### 3. Distribution

- **GitHub Releases**: Binaries for download
- **Go Proxy**: Package available via `go install`
- **Documentation**: Automatically updated

### 4. Installation

```bash
# Install latest version
go install github.com/RodolfoBonis/deco/cmd/deco@latest

# Install specific version
go install github.com/RodolfoBonis/deco/cmd/deco@v1.0.0
```

## Local Commands

### Makefile

```bash
# See all available commands
make help

# Complete pipeline
make all

# Build only
make build

# Tests with coverage
make test-coverage

# Linting
make lint

# Security check
make security

# Development mode
make dev
```

### Manual Commands

```bash
# Local build
go build -o deco ./cmd/deco

# Tests
go test -v -race ./...

# Linting
golangci-lint run

# Security check
govulncheck ./...
```

## Monitoring

### Metrics

- **Test coverage**: Target 80%
- **Build time**: Monitored per job
- **Vulnerabilities**: Blocks release if found

## Troubleshooting

### Common Issues

1. **Build fails on Windows/macOS**
   - Check code compatibility
   - Test locally on different OS

2. **Linting fails**
   - Run `make lint-fix`
   - Check golangci-lint configuration

3. **Vulnerabilities detected**
   - Update dependencies
   - Check if they are false positives

## Next Steps

- **[Installation Guide](./installation.md)** - Setup instructions
- **[Usage Guide](./usage.md)** - How to use decorators
- **[API Reference](./api.md)** - Complete API documentation 