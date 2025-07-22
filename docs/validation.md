# Validation Guide

## Overview

The deco framework includes an advanced validation system that detects and reports specific errors in decorators, providing precise information about the location and nature of problems.

## Validation Features

### 1. Decorator Syntax Validation

The system automatically detects syntax errors:

#### Missing Parentheses
```go
// ERROR: Missing parentheses
// @Route("GET", "/users"
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:15 - Unmatched parentheses in: '@Route("GET", "/users"'
```

#### Unmatched Quotes
```go
// ERROR: Unclosed quotes
// @Summary("User endpoint)  
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:16 - Unmatched quotes in: '@Summary("User endpoint)'
```

### 2. HTTP Method Validation

#### Invalid Methods
```go
// ERROR: Invalid HTTP method
// @Route("INVALID", "/users")
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:18 - Invalid HTTP method 'INVALID' in function GetUsers. Valid methods: [GET POST PUT DELETE PATCH OPTIONS HEAD]
```

### 3. Path Validation

#### Invalid Path
```go
// ERROR: Path must start with '/'
// @Route("GET", "users")
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:19 - Invalid path 'users' in function GetUsers. Path must start with '/'
```

### 4. Argument Validation

#### Insufficient Arguments
```go
// ERROR: @Route requires 2 arguments
// @Route("GET")
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:20 - Invalid @Route syntax in function GetUsers. Use: @Route("METHOD", "/path")
```

## Using the Validation System

### Automatic Validation
Validation runs automatically during code generation:

```bash
deco generate
```

### Verbose Validation
To see more details about the process:

```bash
deco generate --verbose
```

### Multiple Errors
The system reports all errors found:

```
❌ Decorator errors found:
user_handlers.go:15 - Unmatched parentheses in: '@Route("GET", "/users"'
user_handlers.go:18 - Invalid HTTP method 'INVALID' in function GetUsers
user_handlers.go:19 - Invalid path 'users' in function GetUsers
```

## Best Practices

### ✅ Correct Decorators

```go
// ✅ CORRECT: @Route with valid method and path
// @Route("GET", "/users")
// @Summary("List users")
// @Description("Returns paginated list of users")
// @Response(200, type="UserResponse", description="Success")
// @Response(500, type="ErrorResponse", description="Internal error")
func GetUsers(c *gin.Context) {
    // implementation...
}

// ✅ CORRECT: Schema with validations
// @Schema
type UserRequest struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
}
```

### ❌ Common Errors to Avoid

```go
// ❌ AVOID: Missing parentheses
// @Route("GET", "/users"

// ❌ AVOID: Unmatched quotes
// @Summary("List users)

// ❌ AVOID: Invalid methods
// @Route("INVALID", "/users")

// ❌ AVOID: Paths without '/'
// @Route("GET", "users")

// ❌ AVOID: Insufficient arguments
// @Route("GET")
// @Response()
```

## Supported Error Types

| Code | Description | Example |
|------|-------------|---------|
| `MALFORMED_DECORATOR` | Malformed decorator | Missing parentheses |
| `UNMATCHED_QUOTES` | Unmatched quotes | `"text without closing` |
| `UNMATCHED_PARENTHESES` | Unmatched parentheses | `@Route("GET", "/path"))` |
| `INVALID_ROUTE_SYNTAX` | Invalid @Route syntax | Insufficient arguments |
| `INVALID_HTTP_METHOD` | Invalid HTTP method | `INVALID`, `CUSTOM` |
| `INVALID_PATH` | Invalid path | Path without initial `/` |
| `INVALID_ARGUMENTS` | Invalid arguments | Empty or malformed arguments |

## Configuration

### Production Validation
In `.deco.yaml` file:

```yaml
prod:
  validate: true  # Enable strict validation
  minify: true    # Minify generated code
```

### Development Configuration
```yaml
dev:
  auto_discover: true  # Auto-discover handlers
  watch: true         # Watch file changes
```

## Next Steps

- **[Usage Guide](./usage.md)** - How to use decorators
- **[API Reference](./api.md)** - Complete API documentation
- **[Examples](./examples.md)** - Code examples 