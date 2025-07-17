# deco Framework Usage Guide

The **deco** is an annotation-based framework for Gin that simplifies REST API and WebSocket development in Go. This guide shows how to use all available decorators to create robust and well-documented APIs.

## 📋 Table of Contents

- [Basic Concepts](#basic-concepts)
- [Route Decorators](#route-decorators)
- [Middleware Decorators](#middleware-decorators)
- [Documentation Decorators](#documentation-decorators)
- [Validation Decorators](#validation-decorators)
- [WebSocket Decorators](#websocket-decorators)
- [Monitoring Decorators](#monitoring-decorators)
- [Practical Examples](#practical-examples)
- [Advanced Configuration](#advanced-configuration)

## 🎯 Basic Concepts

### Handler Structure

```go
// handlers/user_handlers.go
package handlers

import (
    "github.com/gin-gonic/gin"
    deco "github.com/RodolfoBonis/deco"
)

// @Route("GET", "/users")
// @Description("List all users")
// @Summary("Get users")
// @Tag("users")
// @Response(200, "List of users", "[]User")
func GetUsers(c *gin.Context) {
    // Handler implementation
    c.JSON(200, []User{})
}
```

### Initial Setup

1. **Install CLI:**
```bash
go install github.com/RodolfoBonis/deco/cmd/deco@latest
```

2. **Initialize project:**
```bash
deco init
```

3. **Generate code:**
```bash
deco
```

4. **Import in main.go:**
```go
package main

import (
    "github.com/gin-gonic/gin"
    _ "yourmodule/.deco" // Import generated code
)

func main() {
    r := gin.Default()
    r.Run(":8080")
}
```

## 🛣️ Route Decorators

### @Route

Defines an HTTP route with method and path.

```go
// @Route("GET", "/users")
func GetUsers(c *gin.Context) { }

// @Route("POST", "/users")
func CreateUser(c *gin.Context) { }

// @Route("PUT", "/users/:id")
func UpdateUser(c *gin.Context) { }

// @Route("DELETE", "/users/:id")
func DeleteUser(c *gin.Context) { }
```

### @Group

Groups related routes with a common prefix.

```go
// @Group("users", "/api/v1/users", "User operations")
// @Route("GET", "/")
func GetUsers(c *gin.Context) { }

// @Group("users", "/api/v1/users", "User operations")
// @Route("POST", "/")
func CreateUser(c *gin.Context) { }
```

## 🔧 Middleware Decorators

### @Auth

Authentication and authorization middleware.

```go
// @Route("GET", "/profile")
// @Auth("role=admin")
func GetProfile(c *gin.Context) { }

// @Route("POST", "/admin/users")
// @Auth("role=super_admin")
func CreateUser(c *gin.Context) { }
```

**Parameters:**
- `role`: Defines the required role to access the route

### @Cache

Cache middleware for performance optimization.

```go
// @Route("GET", "/products")
// @Cache("duration=5m,type=memory")
func GetProducts(c *gin.Context) { }

// @Route("GET", "/products/:id")
// @Cache("duration=10m,type=redis")
func GetProduct(c *gin.Context) { }
```

**Parameters:**
- `duration`: Cache lifetime (e.g., `5m`, `1h`)
- `type`: Cache type (`memory`, `redis`)

### @CacheByURL

Cache based on complete URL.

```go
// @Route("GET", "/products")
// @CacheByURL("duration=5m")
func GetProducts(c *gin.Context) { }
```

### @CacheByUser

Cache based on user and URL.

```go
// @Route("GET", "/user/products")
// @CacheByUser("duration=10m")
func GetUserProducts(c *gin.Context) { }
```

### @CacheByEndpoint

Cache based on endpoint (method + path).

```go
// @Route("GET", "/products")
// @CacheByEndpoint("duration=5m")
func GetProducts(c *gin.Context) { }
```

### @RateLimit

Rate limiting to protect against abuse.

```go
// @Route("POST", "/login")
// @RateLimit("limit=5,window=1m")
func Login(c *gin.Context) { }

// @Route("GET", "/api/data")
// @RateLimit("limit=100,window=1h")
func GetData(c *gin.Context) { }
```

**Parameters:**
- `limit`: Maximum number of requests
- `window`: Time window (e.g., `1m`, `1h`)

### @RateLimitByIP

IP-based rate limiting.

```go
// @Route("POST", "/register")
// @RateLimitByIP("limit=3,window=1h")
func Register(c *gin.Context) { }
```

### @RateLimitByUser

User-based rate limiting.

```go
// @Route("POST", "/orders")
// @RateLimitByUser("limit=10,window=1h")
func CreateOrder(c *gin.Context) { }
```

### @RateLimitByEndpoint

Endpoint-based rate limiting.

```go
// @Route("GET", "/api/expensive")
// @RateLimitByEndpoint("limit=50,window=1h")
func ExpensiveOperation(c *gin.Context) { }
```

### @CORS

Cross-Origin Resource Sharing configuration.

```go
// @Route("GET", "/api/public")
// @CORS("origins=*")
func PublicAPI(c *gin.Context) { }

// @Route("GET", "/api/private")
// @CORS("origins=https://app.example.com")
func PrivateAPI(c *gin.Context) { }
```

**Parameters:**
- `origins`: Allowed domains (use `*` for all)

### @Telemetry

Telemetry and observability middleware.

```go
// @Route("GET", "/api/metrics")
// @Telemetry("enabled=true")
func GetMetrics(c *gin.Context) { }
```

## 📝 Documentation Decorators

### @Description

Detailed route description.

```go
// @Route("GET", "/users")
// @Description("Returns a paginated list of all users in the system")
func GetUsers(c *gin.Context) { }
```

### @Summary

Short functionality summary.

```go
// @Route("POST", "/users")
// @Summary("Create a new user")
func CreateUser(c *gin.Context) { }
```

### @Tag

Route categorization for documentation.

```go
// @Route("GET", "/users")
// @Tag("users")
// @Tag("authentication")
func GetUsers(c *gin.Context) { }
```

### @Param

Defines route parameters.

```go
// @Route("GET", "/users/:id")
// @Param("id", "path", "string", true, "User ID", "123")
func GetUser(c *gin.Context) { }

// @Route("GET", "/users")
// @Param("page", "query", "int", false, "Page number", "1")
// @Param("limit", "query", "int", false, "Items per page", "10")
func GetUsers(c *gin.Context) { }
```

**Parameters:**
- Parameter name
- Location (`path`, `query`, `header`, `body`)
- Type (`string`, `int`, `bool`, etc.)
- Required (`true`/`false`)
- Description
- Example

### @Response

Defines possible route responses.

```go
// @Route("GET", "/users")
// @Response(200, "List of users", "[]User")
// @Response(400, "Invalid parameters", "Error")
// @Response(500, "Internal error", "Error")
func GetUsers(c *gin.Context) { }
```

**Parameters:**
- HTTP code
- Description
- Response type

### @Schema

Defines data schemas for documentation.

```go
// @Schema("User", "struct", "Represents a system user")
type User struct {
    ID       string `json:"id" example:"123"`
    Name     string `json:"name" example:"John Doe"`
    Email    string `json:"email" example:"john@example.com"`
    Age      int    `json:"age" example:"30"`
    IsActive bool   `json:"is_active" example:"true"`
}
```

## ✅ Validation Decorators

### @Validate

General input data validation.

```go
// @Route("POST", "/users")
// @Validate("required=name,email;email=email;min=age:18")
func CreateUser(c *gin.Context) { }
```

### @ValidateJSON

JSON-specific validation.

```go
// @Route("POST", "/users")
// @ValidateJSON("required=name,email;email=email")
func CreateUser(c *gin.Context) { }
```

### @ValidateQuery

Query parameter validation.

```go
// @Route("GET", "/users")
// @ValidateQuery("required=page;min=page:1;max=limit:100")
func GetUsers(c *gin.Context) { }
```

### @ValidateParams

Route parameter validation.

```go
// @Route("GET", "/users/:id")
// @ValidateParams("required=id;uuid=id")
func GetUser(c *gin.Context) { }
```

## 🌐 WebSocket Decorators

### @WebSocket

WebSocket configuration for real-time communication.

```go
// @Route("GET", "/ws/chat")
// @WebSocket("chat,notification")
func ChatHandler(c *gin.Context) { }
```

**Parameters:**
- Message types that the handler processes

## 📊 Monitoring Decorators

### @Metrics

Metrics collection for monitoring.

```go
// @Route("GET", "/api/data")
// @Metrics("enabled=true")
func GetData(c *gin.Context) { }
```

### @Prometheus

Prometheus format metrics exposure.

```go
// @Route("GET", "/metrics")
// @Prometheus("enabled=true")
func MetricsHandler(c *gin.Context) { }
```

### @HealthCheck

Health check endpoint.

```go
// @Route("GET", "/health")
// @HealthCheck("enabled=true")
func HealthCheck(c *gin.Context) { }
```

### @HealthCheckWithTracing

Health check with tracing.

```go
// @Route("GET", "/health/trace")
// @HealthCheckWithTracing("enabled=true")
func HealthCheckWithTracing(c *gin.Context) { }
```

### @CacheStats

Cache statistics.

```go
// @Route("GET", "/cache/stats")
// @CacheStats("enabled=true")
func CacheStats(c *gin.Context) { }
```

### @InvalidateCache

Cache invalidation.

```go
// @Route("POST", "/cache/invalidate")
// @InvalidateCache("enabled=true")
func InvalidateCache(c *gin.Context) { }
```

### @WebSocketStats

WebSocket statistics.

```go
// @Route("GET", "/ws/stats")
// @WebSocketStats("enabled=true")
func WebSocketStats(c *gin.Context) { }
```

### @TracingStats

Tracing statistics.

```go
// @Route("GET", "/trace/stats")
// @TracingStats("enabled=true")
func TracingStats(c *gin.Context) { }
```

### @TraceMiddleware

Tracing middleware.

```go
// @Route("GET", "/api/traced")
// @TraceMiddleware("handler_name")
func TracedHandler(c *gin.Context) { }
```

### @InstrumentedHandler

Instrumented handler for observability.

```go
// @Route("GET", "/api/instrumented")
// @InstrumentedHandler("custom_handler")
func InstrumentedHandler(c *gin.Context) { }
```

## 📚 OpenAPI Documentation Decorators

### @OpenAPIJSON

Endpoint for OpenAPI documentation in JSON.

```go
// @Route("GET", "/docs/openapi.json")
// @OpenAPIJSON("enabled=true")
func OpenAPIJSON(c *gin.Context) { }
```

### @OpenAPIYAML

Endpoint for OpenAPI documentation in YAML.

```go
// @Route("GET", "/docs/openapi.yaml")
// @OpenAPIYAML("enabled=true")
func OpenAPIYAML(c *gin.Context) { }
```

### @SwaggerUI

Swagger UI interface for documentation.

```go
// @Route("GET", "/docs/swagger")
// @SwaggerUI("enabled=true")
func SwaggerUI(c *gin.Context) { }
```

## 🚀 Practical Examples

### Complete User API

```go
// handlers/user_handlers.go
package handlers

import (
    "github.com/gin-gonic/gin"
    deco "github.com/RodolfoBonis/deco"
)

// @Schema("User", "struct", "Represents a system user")
type User struct {
    ID       string `json:"id" example:"123"`
    Name     string `json:"name" example:"John Doe"`
    Email    string `json:"email" example:"john@example.com"`
    Age      int    `json:"age" example:"30"`
    IsActive bool   `json:"is_active" example:"true"`
}

// @Group("users", "/api/v1/users", "User operations")
// @Route("GET", "/")
// @Description("Returns a paginated list of all users")
// @Summary("List users")
// @Tag("users")
// @Param("page", "query", "int", false, "Page number", "1")
// @Param("limit", "query", "int", false, "Items per page", "10")
// @ValidateQuery("min=page:1;max=limit:100")
// @Cache("duration=5m,type=memory")
// @RateLimit("limit=100,window=1h")
// @Response(200, "List of users", "[]User")
// @Response(400, "Invalid parameters", "Error")
func GetUsers(c *gin.Context) {
    // Implementation
    c.JSON(200, []User{})
}

// @Group("users", "/api/v1/users", "User operations")
// @Route("POST", "/")
// @Description("Creates a new user in the system")
// @Summary("Create user")
// @Tag("users")
// @ValidateJSON("required=name,email;email=email;min=age:18")
// @Auth("role=admin")
// @RateLimit("limit=10,window=1h")
// @Response(201, "User created", "User")
// @Response(400, "Invalid data", "Error")
// @Response(401, "Unauthorized", "Error")
func CreateUser(c *gin.Context) {
    // Implementation
    c.JSON(201, User{})
}

// @Group("users", "/api/v1/users", "User operations")
// @Route("GET", "/:id")
// @Description("Returns a specific user by ID")
// @Summary("Get user")
// @Tag("users")
// @Param("id", "path", "string", true, "User ID", "123")
// @ValidateParams("required=id;uuid=id")
// @Cache("duration=10m,type=memory")
// @Response(200, "User found", "User")
// @Response(404, "User not found", "Error")
func GetUser(c *gin.Context) {
    // Implementation
    c.JSON(200, User{})
}
```

### Chat API with WebSocket

```go
// handlers/chat_handlers.go
package handlers

import (
    "github.com/gin-gonic/gin"
    deco "github.com/RodolfoBonis/deco"
)

// @Schema("ChatMessage", "struct", "Chat message")
type ChatMessage struct {
    ID        string `json:"id" example:"msg_123"`
    UserID    string `json:"user_id" example:"user_456"`
    Username  string `json:"username" example:"john"`
    Message   string `json:"message" example:"Hello everyone!"`
    Timestamp string `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

// @Route("GET", "/ws/chat")
// @Description("WebSocket endpoint for real-time chat")
// @Summary("Chat WebSocket")
// @Tag("websocket")
// @Tag("chat")
// @WebSocket("chat,notification")
// @Auth("role=user")
func ChatHandler(c *gin.Context) {
    // WebSocket implementation
}
```

### Monitoring API

```go
// handlers/monitoring_handlers.go
package handlers

import (
    "github.com/gin-gonic/gin"
    deco "github.com/RodolfoBonis/deco"
)

// @Route("GET", "/health")
// @Description("Application health check")
// @Summary("Health check")
// @Tag("monitoring")
// @HealthCheck("enabled=true")
// @Response(200, "Application healthy", "HealthStatus")
func HealthCheck(c *gin.Context) {
    c.JSON(200, gin.H{"status": "healthy"})
}

// @Route("GET", "/metrics")
// @Description("Application metrics in Prometheus format")
// @Summary("Prometheus metrics")
// @Tag("monitoring")
// @Prometheus("enabled=true")
func MetricsHandler(c *gin.Context) {
    // Metrics implementation
}

// @Route("GET", "/cache/stats")
// @Description("Cache statistics")
// @Summary("Cache stats")
// @Tag("monitoring")
// @CacheStats("enabled=true")
func CacheStats(c *gin.Context) {
    // Statistics implementation
}
```

## ⚙️ Advanced Configuration

### .deco.yaml File

```yaml
# deco framework configuration
framework:
  name: "deco"
  version: "1.0.0"
  description: "Example API using deco"

# Development settings
dev:
  watch: true
  verbose: true
  auto_reload: true

# Production settings
prod:
  minify: true
  validate: true
  optimize: true

# Cache settings
cache:
  type: "memory"
  default_ttl: "5m"
  max_size: 1000
  redis:
    url: "redis://localhost:6379"

# Rate limiting settings
rate_limit:
  default_limit: 100
  default_window: "1m"
  redis:
    url: "redis://localhost:6379"

# Metrics settings
metrics:
  enabled: true
  prometheus:
    enabled: true
    path: "/metrics"

# WebSocket settings
websocket:
  enabled: true
  read_buffer_size: 1024
  write_buffer_size: 1024
  check_origin: true

# Validation settings
validation:
  enabled: true
  strict: false

# Documentation settings
docs:
  enabled: true
  openapi:
    enabled: true
    title: "API Documentation"
    version: "1.0.0"
  swagger:
    enabled: true
    path: "/docs/swagger"
```

### Custom Hooks

```go
// hooks/custom_hooks.go
package hooks

import (
    deco "github.com/RodolfoBonis/deco"
)

func init() {
    // Hook executed after route parsing
    deco.RegisterParserHook(func(routes []*deco.RouteMeta) error {
        // Custom logic
        return nil
    })

    // Hook executed before code generation
    deco.RegisterGeneratorHook(func(data *deco.GenData) error {
        // Custom logic
        return nil
    })
}
```

### Custom Middlewares

```go
// middleware/custom.go
package middleware

import (
    "regexp"
    deco "github.com/RodolfoBonis/deco"
    "github.com/gin-gonic/gin"
)

func init() {
    // Register custom middleware
    deco.RegisterMarker(deco.MarkerConfig{
        Name:        "Logging",
        Pattern:     regexp.MustCompile(`@Logging\s*\(([^)]*)\)`),
        Factory:     createLoggingMiddleware,
        Description: "Custom logging middleware",
    })
}

func createLoggingMiddleware(args []string) gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        // Middleware implementation
        c.Next()
    })
}
```

## 🔍 Documentation Endpoints

The deco framework automatically generates the following documentation endpoints:

- `/decorators/docs` - HTML documentation
- `/decorators/docs.json` - JSON documentation
- `/decorators/openapi.json` - OpenAPI 3.0 specification (JSON)
- `/decorators/openapi.yaml` - OpenAPI 3.0 specification (YAML)
- `/decorators/swagger-ui` - Swagger UI interface
- `/decorators/swagger` - Swagger UI redirect

## 🎯 Best Practices

1. **Code Organization:**
   - Use `@Group` decorator to organize related routes
   - Keep handlers in separate files by domain
   - Use descriptive names for functions and parameters

2. **Documentation:**
   - Always use `@Description` and `@Summary` to document routes
   - Define `@Param` for all parameters
   - Use `@Response` to document response codes
   - Create `@Schema` for complex data structures

3. **Security:**
   - Use `@Auth` to protect sensitive routes
   - Configure `@RateLimit` to prevent abuse
   - Use `@CORS` appropriately for public APIs

4. **Performance:**
   - Use `@Cache` for data that doesn't change frequently
   - Configure `@Metrics` for monitoring
   - Use `@RateLimit` to protect resources

5. **Validation:**
   - Use appropriate validation decorators
   - Always validate input data
   - Provide clear error messages

## 🚀 Next Steps

1. **Install CLI:** `go install github.com/RodolfoBonis/deco/cmd/deco@latest`
2. **Initialize project:** `deco init`
3. **Create handlers** with appropriate decorators
4. **Generate code:** `deco`
5. **Run application:** `go run main.go`
6. **Access documentation:** `http://localhost:8080/decorators/docs`

For more information, see:
- [API Reference](./api.md) - Complete API reference
- [Examples](./examples.md) - Additional examples
- [CLI Reference](./cli.md) - CLI documentation
