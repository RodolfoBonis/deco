# 🔄 CI/CD for deco Framework

This document describes the CI/CD flows specifically adapted for the **deco** framework, which is a Go package, not an application.

## 📋 Overview

The deco framework uses CI/CD flows optimized for Go packages, focusing on:

- ✅ **Multi-platform testing** (Linux, Windows, macOS)
- ✅ **Code linting and validation**
- ✅ **Security verification**
- ✅ **Binary build and distribution**
- ✅ **Go Proxy publication**
- ✅ **Automatic documentation generation**
- ✅ **Release management**

## 🚀 Available Workflows

### 1. CI Package (`.github/workflows/ci-package.yaml`)

**Trigger:** Push to `main` or Pull Requests

**Jobs:**
- **test**: Tests on multiple platforms and Go versions
- **lint**: Linting with golangci-lint, goimports, go vet
- **security**: Vulnerability checking with govulncheck
- **build**: Binary build on multiple platforms
- **validate**: go.mod and dependencies validation
- **notify**: Telegram notifications

### 2. CD Package (`.github/workflows/cd-package.yaml`)

**Trigger:** After successful CI Package on `main` branch

**Jobs:**
- **get_commit_messages**: Collect commit information
- **build_and_release**: Build, versioning and release creation
- **publish_to_go_proxy**: Publication to Go Proxy
- **generate_documentation**: Automatic documentation update
- **notify**: Success/error notifications

### 3. Documentation (`.github/workflows/docs.yaml`)

**Trigger:** Changes in code files or documentation

**Jobs:**
- **generate_docs**: Automatic documentation generation
- **validate_docs**: Generated documentation validation
- **update_main_readme**: Main README update
- **notify**: Documentation update notifications

### 4. Release Drafter (`.github/workflows/release-drafter.yml`)

**Trigger:** Push to `main` or Pull Requests

**Jobs:**
- **update_release_draft**: Automatic release notes generation

## 🔧 Configurations

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

### Dependabot (`.github/dependabot.yml`)

```yaml
# Automatic updates
- Go modules: Weekly
- GitHub Actions: Weekly
- Ignore major updates of critical dependencies
```

## 📦 Release Process

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

## 🛠️ Local Commands

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

# Documentation generation
go doc -all ./pkg/decorators > docs/api.md
```

## 🔍 Monitoring

### Telegram Notifications

- ✅ **Success**: Release details, version, links
- ❌ **Error**: Debug information, logs, troubleshooting
- 📚 **Documentation**: Documentation update status

### Metrics

- **Test coverage**: Target 80%
- **Build time**: Monitored per job
- **Vulnerabilities**: Blocks release if found

## 🚨 Troubleshooting

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

4. **Documentation doesn't generate**
   - Check if binary compiles
   - Check write permissions

### Logs and Debug

```bash
# See detailed CI logs
# GitHub Actions > Workflows > [Workflow] > [Job] > [Step]

# Test locally
make all

# Check configurations
cat .golangci.yml
cat .codecov.yml
```

## 🔗 Useful Links

- [GitHub Actions](https://github.com/RodolfoBonis/deco/actions)
- [Releases](https://github.com/RodolfoBonis/deco/releases)
- [Go Package](https://pkg.go.dev/github.com/RodolfoBonis/deco)
- [Documentation](https://github.com/RodolfoBonis/deco/tree/main/docs)

## 📝 Important Notes

1. **Not an application**: This framework is not deployed to AWS
2. **Go Package**: Focus on distribution via Go Proxy
3. **CLI Binary**: The main product is a CLI command
4. **Multi-platform**: Build for Linux, Windows, macOS
5. **Documentation**: Automatically generated on each change

---

**Last updated:** $(date)
**Framework version:** $(cat version.txt 2>/dev/null || echo "dev") 