# 🤝 Contributing to deco

Thank you for your interest in contributing to deco! This guide will help you get started with contributing to the project.

## 📋 Table of Contents

1. [Code of Conduct](#-code-of-conduct)
2. [Getting Started](#-getting-started)
3. [Development Setup](#-development-setup)
4. [Making Changes](#-making-changes)
5. [Testing](#-testing)
6. [Submitting Changes](#-submitting-changes)
7. [Code Style](#-code-style)
8. [Documentation](#-documentation)
9. [Community](#-community)

## 📜 Code of Conduct

This project and everyone participating in it is governed by our Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

### Our Standards

- **Be respectful**: Treat everyone with respect and consideration
- **Be inclusive**: Welcome diverse perspectives and backgrounds
- **Be constructive**: Provide helpful feedback and suggestions
- **Be patient**: Help newcomers learn and grow
- **Be collaborative**: Work together towards common goals

## 🚀 Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.19 or higher**
- **Git**
- **Make** (optional, but recommended)

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/your-username/deco.git
cd deco
```

3. Add the upstream repository:

```bash
git remote add upstream https://github.com/original-owner/deco.git
```

## 🛠️ Development Setup

### Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Verify Installation

```bash
# Run tests to ensure everything is working
go test ./...

# Build the CLI tool
go build -o deco ./cmd/deco

# Test the CLI
./deco --version
```

### Project Structure

```
deco/
├── cmd/
│   └── deco/           # CLI application
├── pkg/
│   └── decorators/             # Core framework code
├── examples/                   # Example applications
├── templates/                  # Code generation templates
├── docs/                       # Documentation
├── tests/                      # Integration tests
└── scripts/                    # Build and utility scripts
```

## 🔄 Making Changes

### Branching Strategy

1. **Create a feature branch** from `main`:

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

2. **Branch naming conventions**:
   - `feature/` - New features
   - `fix/` - Bug fixes
   - `docs/` - Documentation updates
   - `refactor/` - Code refactoring
   - `test/` - Adding or updating tests

### Development Workflow

1. **Make your changes** in small, logical commits
2. **Write tests** for your changes
3. **Update documentation** if needed
4. **Run tests** regularly during development
5. **Format your code** before committing

```bash
# Format code
gofmt -w .
goimports -w .

# Run linter
golangci-lint run

# Run tests
go test ./...
```

### Commit Messages

Follow conventional commit format:

```
type(scope): short description

Longer description if necessary

Closes #issue-number
```

**Types:**
- `feat` - New features
- `fix` - Bug fixes
- `docs` - Documentation changes
- `style` - Code style changes (formatting, etc.)
- `refactor` - Code refactoring
- `test` - Adding or updating tests
- `chore` - Maintenance tasks

**Examples:**
```
feat(parser): add support for @Timeout annotation

Add timeout middleware marker that allows setting request timeouts
via comment annotations. Includes validation and error handling.

Closes #123
```

```
fix(generator): escape strings properly in templates

Template generation was failing when route names contained quotes.
Added escapeGoString function to properly handle string escaping.

Fixes #456
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./pkg/decorators

# Run tests with race detection
go test -race ./...
```

### Writing Tests

1. **Unit tests** for individual functions and methods
2. **Integration tests** for complete workflows
3. **Example tests** to ensure documentation examples work

#### Test File Naming

- `*_test.go` - Standard Go test files
- `example_*_test.go` - Example tests that appear in documentation

#### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"
    
    // Act
    result := FunctionToTest(input)
    
    // Assert
    if result != expected {
        t.Errorf("Expected %s, got %s", expected, result)
    }
}
```

#### Table-Driven Tests

```go
func TestParseRoute(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected Route
        hasError bool
    }{
        {
            name:     "valid GET route",
            input:    `@Route("GET", "/api/users")`,
            expected: Route{Method: "GET", Path: "/api/users"},
            hasError: false,
        },
        {
            name:     "invalid method",
            input:    `@Route("INVALID", "/api/users")`,
            expected: Route{},
            hasError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ParseRoute(tt.input)
            
            if tt.hasError && err == nil {
                t.Error("Expected error but got none")
            }
            
            if !tt.hasError && err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("Expected %+v, got %+v", tt.expected, result)
            }
        })
    }
}
```

### Testing Best Practices

1. **Test behavior, not implementation**
2. **Use meaningful test names**
3. **Keep tests simple and focused**
4. **Test edge cases and error conditions**
5. **Use test helpers to reduce duplication**
6. **Mock external dependencies**

## 📤 Submitting Changes

### Before Submitting

1. **Rebase your branch** on the latest main:

```bash
git fetch upstream
git rebase upstream/main
```

2. **Run the full test suite**:

```bash
go test ./...
golangci-lint run
```

3. **Update documentation** if your changes affect the public API

### Creating a Pull Request

1. **Push your branch** to your fork:

```bash
git push origin feature/your-feature-name
```

2. **Create a pull request** on GitHub with:
   - Clear title and description
   - Reference to related issues
   - List of changes made
   - Screenshots if applicable (for UI changes)

### Pull Request Template

```markdown
## Description
Brief description of the changes made.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added tests for new functionality
- [ ] Updated existing tests if needed

## Checklist
- [ ] Code follows the project's style guidelines
- [ ] Self-review of code completed
- [ ] Documentation updated if needed
- [ ] No new warnings introduced
```

### Review Process

1. **Automated checks** will run on your PR
2. **Maintainers will review** your code
3. **Address feedback** by pushing new commits
4. **Squash and merge** once approved

## 🎨 Code Style

### Go Style Guidelines

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and these additional guidelines:

#### Formatting

```bash
# Format your code
gofmt -w .
goimports -w .
```

#### Naming Conventions

- **Functions**: Use camelCase (`parseRoute`, `generateCode`)
- **Types**: Use PascalCase (`RouteMeta`, `ConfigOption`)
- **Constants**: Use PascalCase (`DefaultTimeout`, `MaxRetries`)
- **Interfaces**: Use PascalCase, often ending with -er (`Parser`, `Generator`)

#### Comments

```go
// Package decorators provides annotation-based route handling for Gin.
package decorators

// RouteMeta represents metadata for a discovered route.
type RouteMeta struct {
    Method string // HTTP method (GET, POST, etc.)
    Path   string // Route path with parameters
}

// ParseRoute extracts route information from comment annotations.
// It returns an error if the annotation format is invalid.
func ParseRoute(comment string) (*RouteMeta, error) {
    // Implementation details...
}
```

#### Error Handling

```go
// Prefer explicit error handling
result, err := someOperation()
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Use meaningful error messages
if input == "" {
    return fmt.Errorf("input cannot be empty")
}
```

#### Package Organization

- Keep packages focused and cohesive
- Avoid circular dependencies
- Use internal packages for implementation details
- Export only what's necessary

### Linting

The project uses `golangci-lint` with the following configuration:

```yaml
# .golangci.yml
linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - structcheck
    - varcheck
    - ineffassign
    - deadcode
```

Run the linter before submitting:

```bash
golangci-lint run
```

## 📚 Documentation

### Types of Documentation

1. **Code comments** - Explain complex logic
2. **API documentation** - GoDoc comments for public APIs
3. **User guides** - How-to articles and tutorials
4. **Reference docs** - Complete API reference
5. **Examples** - Working code examples

### Writing Documentation

#### GoDoc Comments

```go
// ParseDirectory analyzes a directory and extracts route metadata.
//
// The function recursively scans the specified directory for Go files
// containing route annotations. It returns a slice of RouteMeta objects
// representing the discovered routes.
//
// Example:
//   routes, err := ParseDirectory("./handlers")
//   if err != nil {
//       log.Fatal(err)
//   }
//   for _, route := range routes {
//       fmt.Printf("%s %s\n", route.Method, route.Path)
//   }
//
// Returns an error if the directory cannot be read or if any files
// contain invalid annotations.
func ParseDirectory(dir string) ([]*RouteMeta, error) {
    // Implementation...
}
```

#### Markdown Documentation

- Use clear headings and structure
- Include code examples
- Add links to related documentation
- Use tables for reference information
- Include screenshots for visual features

#### Example Code

- Ensure all examples compile and run
- Use realistic, practical scenarios
- Include error handling
- Add comments explaining key concepts

### Documentation Standards

1. **Keep it up to date** - Update docs with code changes
2. **Be clear and concise** - Avoid unnecessary complexity
3. **Use examples** - Show, don't just tell
4. **Link related concepts** - Help users navigate
5. **Test examples** - Ensure code examples work

## 👥 Community

### Getting Help

- **GitHub Issues** - Report bugs and request features
- **GitHub Discussions** - Ask questions and share ideas
- **Documentation** - Check existing docs first
- **Examples** - Look at example applications

### Helping Others

- **Answer questions** in discussions
- **Review pull requests** from other contributors
- **Improve documentation** based on common questions
- **Share your experience** using the framework

### Communication Guidelines

- **Be respectful** and professional
- **Search existing issues** before creating new ones
- **Provide clear reproduction steps** for bugs
- **Use descriptive titles** for issues and PRs
- **Follow up** on your issues and PRs

## 🏷️ Release Process

### Versioning

The project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality
- **PATCH** version for backwards-compatible bug fixes

### Release Checklist

1. Update version numbers
2. Update CHANGELOG.md
3. Run full test suite
4. Create and test release build
5. Tag the release
6. Update documentation
7. Announce the release

## 🙏 Recognition

Contributors are recognized in:

- **CONTRIBUTORS.md** file
- **Release notes** for significant contributions
- **GitHub contributors** page
- **Documentation credits** for doc contributions

## 📞 Contact

- **Project Maintainers**: Listed in README.md
- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and general discussion

---

**Thank you for contributing to deco!** 🎉

Your contributions help make the framework better for everyone in the Go community. 