# Deco Framework 🚀

A modern, annotation-driven Go web framework built on top of Gin. Write web APIs using simple annotations and let deco handle the heavy lifting - automatic route registration, middleware injection, validation, caching, rate limiting, and more!

## ✨ Features

- **Annotation-Driven**: Define routes and middleware with simple `@` annotations
- **Zero Configuration**: Works out of the box with sensible defaults
- **Development Mode**: Auto-reload with file watching for rapid development
- **Production Ready**: Optimized builds with minification and validation
- **Comprehensive Middleware**: Authentication, caching, rate limiting, validation, CORS, metrics, tracing
- **Schema System**: Define entities with `@Schema` for automatic OpenAPI integration
- **Interactive Swagger UI**: Built-in Swagger interface with schema visualization and API testing
- **OpenAPI 3.0**: Complete specification generation with schema references and array support
- **WebSocket Support**: Real-time communication with built-in handlers
- **Flexible Caching**: In-memory and Redis support with multiple strategies
- **Rate Limiting**: IP, user, and endpoint-based limiting
- **Observability**: Prometheus metrics and OpenTelemetry tracing
- **Type-Safe Validation**: Automatic request/response validation with schema linking
- **Advanced Error Detection**: Comprehensive decorator syntax validation with precise error location reporting

## 🚀 Quick Start

### Installation

```bash
go install github.com/RodolfoBonis/deco/cmd/decorate-gen@latest
```

### Initialize Project

```bash
mkdir my-api && cd my-api
deco init
```

### Create Your First Handler

```go
// handlers/user_handlers.go
package handlers

import "github.com/gin-gonic/gin"

// @Route(method="GET", path="/users/:id")
// @Auth(role="user")
// @Cache(ttl="5m")
// @Validate()
// @Description("Get user by ID")
func GetUser(c *gin.Context) {
    c.JSON(200, gin.H{"id": c.Param("id"), "name": "John Doe"})
}

// @Route(method="POST", path="/users")
// @Auth(role="admin")
// @ValidateJSON()
// @RateLimit(limit=10, window="1m")
// @Description("Create new user")
func CreateUser(c *gin.Context) {
    c.JSON(201, gin.H{"message": "User created"})
}
```

### Generate and Run

```bash
deco generate
go run main.go
```

## 📋 Available Decorators

### 🔒 Authentication & Authorization

| Decorator | Description | Example |
|-----------|-------------|---------|
| `@Auth()` | Authentication with optional role-based access | `@Auth(role="admin")` |

### 💾 Caching

| Decorator | Description | Example |
|-----------|-------------|---------|
| `@Cache()` | General caching with TTL | `@Cache(ttl="5m", type="memory")` |
| `@CacheByURL()` | Cache based on URL path | `@CacheByURL(ttl="10m")` |
| `@CacheByUser()` | Cache per user and URL | `@CacheByUser(ttl="1h")` |
| `@CacheByEndpoint()` | Cache per endpoint | `@CacheByEndpoint(ttl="30m")` |
| `@CacheStats()` | Cache statistics endpoint | `@CacheStats()` |
| `@InvalidateCache()` | Cache invalidation endpoint | `@InvalidateCache()` |

### 🛡️ Rate Limiting

| Decorator | Description | Example |
|-----------|-------------|---------|
| `@RateLimit()` | General rate limiting | `@RateLimit(limit=100, window="1m")` |
| `@RateLimitByIP()` | Rate limit by client IP | `@RateLimitByIP(limit=50, window="1m")` |
| `@RateLimitByUser()` | Rate limit by user ID | `@RateLimitByUser(limit=1000, window="1h")` |
| `@RateLimitByEndpoint()` | Rate limit per endpoint | `@RateLimitByEndpoint(limit=200, window="1m")` |

### ✅ Validation

| Decorator | Description | Example |
|-----------|-------------|---------|
| `@Validate()` | General struct validation | `@Validate()` |
| `@ValidateJSON()` | JSON body validation | `@ValidateJSON()` |
| `@ValidateQuery()` | Query parameter validation | `@ValidateQuery()` |
| `@ValidateParams()` | Path parameter validation | `@ValidateParams(id="uuid", name="alpha")` |

### 🌐 CORS & WebSocket

| Decorator | Description | Example |
|-----------|-------------|---------|
| `@CORS()` | Cross-Origin Resource Sharing | `@CORS(origins="*.example.com")` |
| `@WebSocket()` | WebSocket connection handler | `@WebSocket(path="/ws")` |
| `@WebSocketStats()` | WebSocket statistics | `@WebSocketStats()` |

### 📊 Metrics & Monitoring

| Decorator | Description | Example |
|-----------|-------------|---------|
| `@Metrics()` | Custom metrics collection | `@Metrics()` |
| `@Prometheus()` | Prometheus metrics endpoint | `@Prometheus()` |
| `@HealthCheck()` | Health check endpoint | `@HealthCheck()` |
| `@HealthCheckWithTracing()` | Health check with tracing | `@HealthCheckWithTracing()` |

### 🔍 Tracing & Observability

| Decorator | Description | Example |
|-----------|-------------|---------|
| `@Telemetry()` | OpenTelemetry tracing | `@Telemetry()` |
| `@TraceMiddleware()` | Named trace middleware | `@TraceMiddleware(name="auth")` |
| `@TracingStats()` | Tracing statistics | `@TracingStats()` |
| `@InstrumentedHandler()` | Handler instrumentation | `@InstrumentedHandler(name="api")` |

### 📖 Documentation

| Decorator | Description | Example |
|-----------|-------------|---------|
| `@OpenAPIJSON()` | OpenAPI JSON endpoint | `@OpenAPIJSON()` |
| `@OpenAPIYAML()` | OpenAPI YAML endpoint | `@OpenAPIYAML()` |
| `@SwaggerUI()` | Swagger UI interface | `@SwaggerUI()` |
| `@Schema()` | Entity/struct definition | `@Schema()` |
| `@Description()` | Handler description | `@Description("User management endpoint")` |
| `@Summary()` | Handler summary | `@Summary("Get user")` |
| `@Tag()` | API grouping tag | `@Tag("users")` |
| `@Group()` | Route grouping | `@Group("api/v1")` |
| `@Param()` | Parameter documentation | `@Param(name="id", type="string", location="path")` |
| `@Response()` | Response documentation | `@Response(code=200, description="Success")` |

## 🔧 Configuration

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

cache:
  type: "memory"  # or "redis"
  default_ttl: "5m"
  max_size: 1000

rate_limit:
  type: "memory"  # or "redis" 
  default_limit: 100
  default_window: "1m"

telemetry:
  enabled: true
  service_name: "my-api"
  endpoint: "http://localhost:14268/api/traces"

metrics:
  enabled: true
  path: "/metrics"
  namespace: "myapp"
```

## 📚 Advanced Examples

### Complete API Endpoint

```go
// @Route(method="POST", path="/api/v1/users")
// @Auth(role="admin")
// @CORS(origins="https://app.example.com")
// @RateLimit(limit=10, window="1m")
// @ValidateJSON()
// @Cache(ttl="1m", type="redis")
// @Metrics()
// @Telemetry()
// @Description("Create a new user with full validation and monitoring")
// @Tag("users")
// @Response(code=201, description="User created successfully")
// @Response(code=400, description="Invalid input")
// @Response(code=401, description="Unauthorized")
// @Response(code=429, description="Rate limit exceeded")
func CreateUser(c *gin.Context) {
    // Your handler logic here
    c.JSON(201, gin.H{"message": "User created successfully"})
}
```

### WebSocket Handler

```go
// @Route(method="GET", path="/ws")
// @WebSocket()
// @Auth(role="user")
// @Metrics()
func HandleWebSocket(c *gin.Context) {
    // WebSocket connection automatically handled
    // Real-time messaging, broadcasting, etc.
}
```

### Monitoring Endpoints

```go
// @Route(method="GET", path="/health")
// @HealthCheck()
func HealthEndpoint(c *gin.Context) {}

// @Route(method="GET", path="/metrics")
// @Prometheus()
func MetricsEndpoint(c *gin.Context) {}

// @Route(method="GET", path="/docs")
// @OpenAPIJSON()
func DocsEndpoint(c *gin.Context) {}

// @Route(method="GET", path="/swagger")
// @SwaggerUI()
func SwaggerEndpoint(c *gin.Context) {}
```

## 📋 Schema System & API Documentation

Deco provides a powerful schema system that automatically generates comprehensive OpenAPI documentation:

### Define Schemas

```go
// @Schema()
// @Description("User entity representing a registered user")
type User struct {
    ID    int    `json:"id" validate:"required"`                 // User unique identifier
    Name  string `json:"name" validate:"required,min=2,max=100"` // Full name of the user
    Email string `json:"email" validate:"required,email"`        // Email address (must be unique)
    Role  string `json:"role" validate:"oneof=admin user guest"` // User role in the system
}

// @Schema()
// @Description("Paginated response containing a list of users")
type ListUsersResponse struct {
    Users   []User `json:"users" validate:"required"`       // List of users
    Total   int    `json:"total" validate:"required"`       // Total number of users
    Page    int    `json:"page" validate:"required,min=1"`  // Current page number
    HasNext bool   `json:"hasNext"`                         // Whether there are more pages
}
```

### Link Schemas to Endpoints

```go
// @Route(method="GET", path="/api/users")
// @Description("List all users with pagination")
// @Response(code=200, description="Users retrieved successfully", type="ListUsersResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
func ListUsers(c *gin.Context) {
    response := ListUsersResponse{
        Users: []User{{ID: 1, Name: "John", Email: "john@example.com", Role: "user"}},
        Total: 1,
        Page:  1,
        HasNext: false,
    }
    c.JSON(200, response)
}
```

### Interactive Documentation

- **Swagger UI**: `http://localhost:8080/decorators/swagger-ui` - Interactive API testing
- **OpenAPI JSON**: `http://localhost:8080/decorators/openapi.json` - Complete specification
- **Framework Docs**: `http://localhost:8080/decorators/docs` - Framework statistics

**Features:**
- Automatic schema detection and registration
- Array support with proper item type references
- Validation constraints in documentation
- Interactive testing in Swagger UI
- Type-safe request/response mapping

## 🛠️ CLI Commands

```bash
# Initialize new project
deco init

# Generate code
deco generate

# Generate with specific config
deco generate --config custom.yaml

# Generate with output path
deco generate --output ./generated/routes.go

# Watch mode (development)
deco generate --watch

# Production build
deco generate --minify --validate

# Show version
deco --version

# Show help
deco --help
```

## 🔄 Development Workflow

1. **Write handlers** with annotations
2. **Run `deco generate`** to generate route registration
3. **Start your app** - routes are automatically registered
4. **Make changes** - file watcher auto-regenerates code
5. **Deploy to production** - minified and validated code

## 🌟 Why Choose Deco?

- **🚀 Fast Development**: Write less boilerplate, focus on business logic
- **📖 Self-Documenting**: Annotations serve as inline documentation
- **🔧 Flexible**: Use only the features you need
- **📊 Observable**: Built-in metrics, tracing, and monitoring
- **🛡️ Secure**: Authentication, validation, and rate limiting out of the box
- **⚡ Performant**: Optimized for production with caching and minification
- **🔌 Extensible**: Easy to add custom decorators and middleware

## 📖 Documentation

- [Usage Guide](USAGE.md) - Detailed examples and tutorials
- [API Reference](API_REFERENCE.md) - Complete decorator documentation
- [Examples](EXAMPLES.md) - Real-world usage patterns
- [Validation Guide](VALIDATION_GUIDE.md) - Error detection and decorator validation
- [Contributing](CONTRIBUTING.md) - Development guidelines

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

**Made with ❤️ for the Go community** 