# Examples

## Quick Examples

### Basic API

```go
// handlers/health.go
package handlers

import "github.com/gin-gonic/gin"

// @Route("GET", "/health")
func HealthCheck(c *gin.Context) {
    c.JSON(200, gin.H{"status": "healthy"})
}
```

### User Management

```go
// handlers/users.go
package handlers

import "github.com/gin-gonic/gin"

// @Route("GET", "/users/:id")
// @Cache(ttl="5m")
// @Auth(role="user")
func GetUser(c *gin.Context) {
    c.JSON(200, gin.H{"id": c.Param("id"), "name": "John Doe"})
}

// @Route("POST", "/users")
// @ValidateJSON()
// @RateLimit(limit=10, window="1m")
func CreateUser(c *gin.Context) {
    c.JSON(201, gin.H{"message": "User created"})
}
```

### Security

```go
// handlers/admin.go
package handlers

import "github.com/gin-gonic/gin"

// @Route("GET", "/admin/dashboard")
// @Security(private)
// @Auth(role="admin")
func AdminDashboard(c *gin.Context) {
    c.JSON(200, gin.H{"dashboard": "Admin panel"})
}
```

### API Gateway

```go
// handlers/proxy.go
package handlers

import "github.com/gin-gonic/gin"

// @Route("GET", "/api/users/:id")
// @Proxy(service="user-service")
func GetUserProxy(c *gin.Context) {
    // Automatically forwards to user-service
}
```

## Available Examples

### Basic Examples
- **[basic](../examples/basic/)** - Complete API with all decorators

### Security Examples
- **[security](../examples/security/)** - Security features demonstration

## Complete Examples

For comprehensive examples with real-world scenarios, see the **[EXAMPLES.md](../EXAMPLES.md)** file in the root directory.

## Next Steps

- **[Usage Guide](./usage.md)** - Learn how to use decorators
- **[API Reference](./api.md)** - Complete API documentation
- **[Security Guide](./security.md)** - Security features
- **[Proxy Guide](./proxy.md)** - API Gateway features

