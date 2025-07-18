# Usage Guide

## Overview

The deco framework uses annotations to automatically generate routes, middleware, and documentation. This guide shows you how to use decorators effectively.

## Basic Concepts

### Route Registration

Routes are automatically registered using `@Route` decorators:

```go
// @Route("GET", "/users/:id")
func GetUser(c *gin.Context) {
    // Handler implementation
}
```

### Middleware Integration

Middlewares are applied using decorators:

```go
// @Route("GET", "/users/:id")
// @Auth(role="user")
// @Cache(ttl="5m")
// @RateLimit(limit=100, window="1m")
func GetUser(c *gin.Context) {
    // Handler with authentication, caching, and rate limiting
}
```

## Core Decorators

### 🔒 Authentication

```go
// @Route("GET", "/admin/users")
// @Auth(role="admin")
func AdminUsers(c *gin.Context) {
    // Only accessible by admin users
}
```

### 💾 Caching

```go
// Basic caching
// @Route("GET", "/users/:id")
// @Cache(ttl="5m")
func GetUser(c *gin.Context) {
    // Response cached for 5 minutes
}

// Cache by user
// @Route("GET", "/profile")
// @CacheByUser(ttl="1h")
func GetProfile(c *gin.Context) {
    // Cached per user
}
```

### 🛡️ Rate Limiting

```go
// @Route("POST", "/users")
// @RateLimit(limit=10, window="1m")
func CreateUser(c *gin.Context) {
    // 10 requests per minute per client
}
```

### ✅ Validation

```go
// @Route("POST", "/users")
// @ValidateJSON()
func CreateUser(c *gin.Context) {
    var user struct {
        Name  string `json:"name" binding:"required"`
        Email string `json:"email" binding:"required,email"`
    }
    // Automatic validation of JSON body
}
```

## Security Features

### Automatic Protection

```go
// Automatic protection (localhost only)
r := deco.Default()

// Custom security configuration
securityConfig := &deco.SecurityConfig{
    AllowPrivateNetworks: true,
    AllowLocalhost: true,
}
r := deco.DefaultWithSecurity(securityConfig)
```

### Application Security

```go
// @Route("GET", "/admin/dashboard")
// @Security(private)
func AdminDashboard(c *gin.Context) {
    // Only accessible from private networks
}
```

## API Gateway & Proxy

### Basic Proxy

```go
// @Route("GET", "/api/users/:id")
// @Proxy(service="user-service")
func GetUserProxy(c *gin.Context) {
    // Automatically forwards to user-service
}
```

### Advanced Proxy

```go
// @Route("GET", "/api/products")
// @Proxy(
//     service="product-service",
//     discovery="consul",
//     load_balancer="round_robin"
// )
func GetProductsProxy(c *gin.Context) {
    // Advanced proxy with service discovery
}
```

## Observability

### Health Checks

```go
// @Route("GET", "/health")
// @HealthCheck()
func HealthCheck(c *gin.Context) {
    // Basic health check
}
```

### Metrics

```go
// @Route("GET", "/metrics")
// @Prometheus()
func MetricsHandler(c *gin.Context) {
    // Prometheus metrics endpoint
}
```

### Tracing

```go
// @Route("GET", "/api/data")
// @Telemetry()
func GetData(c *gin.Context) {
    // Automatic tracing and metrics
}
```

## Documentation

### OpenAPI Generation

```go
// @Route("GET", "/users/:id")
// @Description("Get user by ID")
// @Summary("Retrieve user information")
// @Tag("users")
// @Param(name="id", type="string", location="path")
// @Response(code=200, description="User found")
func GetUser(c *gin.Context) {
    // Automatic OpenAPI documentation
}
```

### Schema Definition

```go
// @Schema()
type User struct {
    ID    int    `json:"id" example:"1"`
    Name  string `json:"name" example:"John Doe"`
    Email string `json:"email" example:"john@example.com"`
}
```

## Configuration

### YAML Configuration

Create `.deco.yaml` in your project root:

```yaml
handlers:
  include:
    - "handlers/**/*.go"
    - "api/**/*.go"
  exclude:
    - "**/*_test.go"

generation:
  output: ".deco/init_decorators.go"
  package: "deco"

dev:
  watch: true
  hot_reload: true

prod:
  minify: true
  validate: true
```

## Development Workflow

### 1. Generate Code

```bash
deco generate
```

### 2. Development Mode

```bash
deco dev
```

### 3. Production Build

```bash
deco build
```

### 4. Validation

```bash
deco validate
```

## Best Practices

### 1. Organize Handlers

```bash
handlers/
├── users.go
├── orders.go
├── auth.go
└── admin.go
```

### 2. Use Descriptive Decorators

```go
// Good
// @Route("GET", "/users/:id")
// @Description("Retrieve user information by ID")
// @Auth(role="user")
// @Cache(ttl="5m")

// Avoid
// @Route("GET", "/u/:id")
// @Cache(ttl="5m")
```

### 3. Implement Error Handling

```go
// @Route("GET", "/users/:id")
func GetUser(c *gin.Context) {
    id := c.Param("id")
    user, err := userService.GetByID(id)
    if err != nil {
        c.JSON(404, gin.H{"error": "User not found"})
        return
    }
    c.JSON(200, user)
}
```

## Troubleshooting

### Common Issues

1. **Routes not registered**: Ensure handlers are imported in main.go
2. **Middleware not working**: Check decorator syntax and parameters
3. **Generation errors**: Run `deco validate` to check for issues

### Debug Mode

```bash
deco generate --verbose
```

## Next Steps

- **[API Reference](./api.md)** - Complete API documentation
- **[Examples](./examples.md)** - Code examples and tutorials
- **[Security Guide](./security.md)** - Security features
- **[Proxy Guide](./proxy.md)** - API Gateway features
