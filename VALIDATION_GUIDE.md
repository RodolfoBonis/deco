# Decorator Validation Guide

The **deco** framework now includes an advanced validation system that detects and reports specific errors in decorators, providing precise information about the location and nature of problems.

## 🎯 Validation Features

### 1. Decorator Syntax Validation

The system automatically detects the following types of errors:

#### ❌ Missing Parentheses
```go
// ERROR: Missing parentheses
// @Route("GET", "/users"
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:15 - Unmatched parentheses in: '@Route("GET", "/users"'
```

#### ❌ Unmatched Quotes
```go
// ERROR: Unclosed quotes
// @Summary("User endpoint)  
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:16 - Unmatched quotes in: '@Summary("User endpoint)'
```

#### ❌ Unbalanced Parentheses
```go
// ERROR: Extra parentheses
// @Route("GET", "/users")))
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:17 - Unmatched parentheses in: '@Route("GET", "/users")))'
```

### 2. HTTP Method Validation

#### ❌ Invalid Methods
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

#### ❌ Invalid Path
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

#### ❌ Insufficient Arguments
```go
// ERROR: @Route requires 2 arguments
// @Route("GET")
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:20 - Invalid @Route syntax in function GetUsers. Use: @Route("METHOD", "/path")
```

#### ❌ @Response Arguments
```go
// ERROR: @Response without arguments
// @Response()
func GetUsers(c *gin.Context) { ... }
```

**Error reported:**
```
user_handlers.go:21 - Error in @Response decorator arguments: @Response requires at least 1 argument (status code)
```

## 🔧 How to Use the Validation System

### 1. Automatic Validation
Validation runs automatically during code generation:

```bash
deco generate
```

### 2. Verbose Validation
To see more details about the process:

```bash
deco generate --verbose
```

### 3. Multiple Errors
The system reports all errors found:

```
❌ Decorator errors found:
user_handlers.go:15 - Unmatched parentheses in: '@Route("GET", "/users"'
user_handlers.go:18 - Invalid HTTP method 'INVALID' in function GetUsers
user_handlers.go:19 - Invalid path 'users' in function GetUsers
```

## 📋 Best Practices Checklist

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

## 🚀 Supported Error Types

| Code | Description | Example |
|------|-------------|---------|
| `MALFORMED_DECORATOR` | Malformed decorator | Missing parentheses |
| `UNMATCHED_QUOTES` | Unmatched quotes | `"text without closing` |
| `UNMATCHED_PARENTHESES` | Unmatched parentheses | `@Route("GET", "/path"))` |
| `INVALID_ROUTE_SYNTAX` | Invalid @Route syntax | Insufficient arguments |
| `INVALID_HTTP_METHOD` | Invalid HTTP method | `INVALID`, `CUSTOM` |
| `INVALID_PATH` | Invalid path | Path without initial `/` |
| `INVALID_ARGUMENTS` | Invalid arguments | Empty or malformed arguments |

## 🛠️ Validation Configuration

### Enabling Production Validation
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

## 📚 Complete Examples

### Valid Handler Example
```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// UserResponse represents a user
// @Schema
type UserResponse struct {
    ID    int    `json:"id" validate:"required"`
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
}

// GetUsers lists all users
// @Route("GET", "/users")
// @Summary("List users")
// @Description("Returns a paginated list of users")
// @Tag("users")
// @Response(200, type="UserResponse", description="List of users")
// @Response(500, type="ErrorResponse", description="Internal server error")
func GetUsers(c *gin.Context) {
    users := []UserResponse{
        {ID: 1, Name: "John", Email: "john@example.com"},
        {ID: 2, Name: "Mary", Email: "mary@example.com"},
    }
    c.JSON(http.StatusOK, users)
}
```

### Error Handling Example
```go
// When there's a validation error, the system returns:
type ValidationError struct {
    File    string `json:"file"`    // File where error occurred
    Line    int    `json:"line"`    // Error line
    Message string `json:"message"` // Problem description
    Code    string `json:"code"`    // Error type code
}
```

## 🔍 Troubleshooting

### 1. "Invalid @Route syntax"
- **Problem**: Malformed @Route
- **Solution**: Use `@Route("METHOD", "/path")`

### 2. "Invalid HTTP method"
- **Problem**: Unsupported method
- **Solution**: Use GET, POST, PUT, DELETE, PATCH, OPTIONS or HEAD

### 3. "Invalid path"
- **Problem**: Path doesn't start with `/`
- **Solution**: Always start paths with `/`

### 4. "Unmatched parentheses"
- **Problem**: Parentheses not properly closed
- **Solution**: Check that all `(` have a corresponding `)`

### 5. "Unmatched quotes"
- **Problem**: Unclosed quotes
- **Solution**: Check that all `"` are in pairs

---

This validation system ensures your decorators are always correct and well-formatted, preventing runtime errors and improving the quality of generated code. 