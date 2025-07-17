# Deco Framework - API Reference 📖

Complete technical reference for all Deco framework decorators and their configuration options.

## Table of Contents

- [Route Definition](#route-definition)
- [Authentication & Authorization](#authentication--authorization)
- [Caching System](#caching-system)
- [Rate Limiting](#rate-limiting)
- [Validation](#validation)
- [CORS & WebSocket](#cors--websocket)
- [Metrics & Monitoring](#metrics--monitoring)
- [Tracing & Observability](#tracing--observability)
- [Documentation](#documentation)
- [Configuration Reference](#configuration-reference)
- [CLI Reference](#cli-reference)

---

## Route Definition

### @Route

**Purpose**: Define HTTP routes with method and path  
**Type**: Core decorator  
**Factory**: Built-in

**Syntax**:
```go
// @Route(method="METHOD", path="PATH")
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `method` | string | Yes | HTTP method: GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD |
| `path` | string | Yes | URL path with support for parameters |

**Examples**:
```go
// Basic routes
// @Route(method="GET", path="/users")
// @Route(method="POST", path="/users")
// @Route(method="PUT", path="/users/:id")
// @Route(method="DELETE", path="/users/:id")

// Path parameters
// @Route(method="GET", path="/users/:id")
// @Route(method="GET", path="/users/:id/posts/:postId")
// @Route(method="GET", path="/files/*filepath")

// Query parameters (handled automatically)
// @Route(method="GET", path="/search")  // ?q=term&page=1&limit=10
```

**Generated Code**:
```go
router.GET("/users/:id", handlers.GetUser)
router.POST("/users", handlers.CreateUser)
```

---

## Authentication & Authorization

### @Auth

**Purpose**: Require authentication with optional role-based access control  
**Type**: Middleware decorator  
**Factory**: `createAuthMiddleware`

**Syntax**:
```go
// @Auth()                    // Basic authentication required
// @Auth(role="ROLE")         // Role-based access control
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `role` | string | No | Required user role (admin, user, moderator, etc.) |

**Examples**:
```go
// Basic authentication
// @Route(method="GET", path="/profile")
// @Auth()
func GetProfile(c *gin.Context) {}

// Role-based access
// @Route(method="DELETE", path="/admin/users/:id")
// @Auth(role="admin")
func DeleteUser(c *gin.Context) {}

// Different roles
// @Auth(role="user")      // Regular users
// @Auth(role="moderator") // Moderators
// @Auth(role="admin")     // Administrators
```

**Behavior**:
- Validates `Authorization` header
- Checks JWT token (configurable)
- Enforces role-based permissions
- Sets user context for downstream handlers
- Returns 401 for missing/invalid tokens
- Returns 403 for insufficient privileges

**Integration**:
```go
// Access user info in handler
func GetProfile(c *gin.Context) {
    user := c.MustGet("user").(*UserClaims)
    // Use user information
}
```

---

## Caching System

### @Cache

**Purpose**: General HTTP response caching  
**Type**: Middleware decorator  
**Factory**: `createCacheMiddleware`

**Syntax**:
```go
// @Cache(ttl="DURATION")
// @Cache(ttl="DURATION", type="TYPE")
```

**Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `ttl` | duration | Yes | - | Time-to-live (30s, 5m, 1h, 24h) |
| `type` | string | No | memory | Cache type: `memory`, `redis` |

**Examples**:
```go
// Basic caching
// @Route(method="GET", path="/products")
// @Cache(ttl="5m")
func GetProducts(c *gin.Context) {}

// Redis caching
// @Cache(ttl="1h", type="redis")
func GetExpensiveData(c *gin.Context) {}

// Different TTL values
// @Cache(ttl="30s")   // 30 seconds
// @Cache(ttl="10m")   // 10 minutes
// @Cache(ttl="2h")    // 2 hours
// @Cache(ttl="24h")   // 24 hours
```

### @CacheByURL

**Purpose**: Cache responses based on URL path only  
**Type**: Middleware decorator  
**Factory**: `createCacheByURLMiddleware`

**Syntax**:
```go
// @CacheByURL(ttl="DURATION")
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `ttl` | duration | Yes | Cache time-to-live |

**Examples**:
```go
// Cache public content by URL
// @Route(method="GET", path="/public/news")
// @CacheByURL(ttl="15m")
func GetNews(c *gin.Context) {}
```

**Behavior**: Cache key = URL path (ignores query parameters and user context)

### @CacheByUser

**Purpose**: Cache responses per authenticated user  
**Type**: Middleware decorator  
**Factory**: `createCacheByUserMiddleware`

**Syntax**:
```go
// @CacheByUser(ttl="DURATION")
```

**Examples**:
```go
// User-specific caching
// @Route(method="GET", path="/recommendations")
// @Auth(role="user")
// @CacheByUser(ttl="1h")
func GetRecommendations(c *gin.Context) {}
```

**Behavior**: Cache key = user_id + URL path

### @CacheByEndpoint

**Purpose**: Cache responses per endpoint  
**Type**: Middleware decorator  
**Factory**: `createCacheByEndpointMiddleware`

**Syntax**:
```go
// @CacheByEndpoint(ttl="DURATION")
```

**Examples**:
```go
// Endpoint-specific caching
// @Route(method="GET", path="/api/stats")
// @CacheByEndpoint(ttl="30m")
func GetStats(c *gin.Context) {}
```

**Behavior**: Cache key = endpoint name

### @CacheStats

**Purpose**: Expose cache statistics endpoint  
**Type**: Handler decorator  
**Factory**: `createCacheStatsMiddleware`

**Syntax**:
```go
// @CacheStats()
```

**Examples**:
```go
// @Route(method="GET", path="/admin/cache/stats")
// @Auth(role="admin")
// @CacheStats()
func CacheStatistics(c *gin.Context) {}
```

**Response**:
```json
{
  "hits": 1542,
  "misses": 234,
  "hit_rate": 86.8,
  "size": 150,
  "max_size": 1000
}
```

### @InvalidateCache

**Purpose**: Cache invalidation endpoint  
**Type**: Handler decorator  
**Factory**: `createInvalidateCacheMiddleware`

**Syntax**:
```go
// @InvalidateCache()
```

**Examples**:
```go
// @Route(method="DELETE", path="/admin/cache")
// @Auth(role="admin")
// @InvalidateCache()
func InvalidateCache(c *gin.Context) {}
```

---

## Rate Limiting

### @RateLimit

**Purpose**: General rate limiting  
**Type**: Middleware decorator  
**Factory**: `createRateLimitMiddlewareInternal`

**Syntax**:
```go
// @RateLimit(limit=COUNT, window="DURATION")
// @RateLimit(limit=COUNT, window="DURATION", type="TYPE")
```

**Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `limit` | int | Yes | - | Maximum requests allowed |
| `window` | duration | Yes | - | Time window (1s, 1m, 1h) |
| `type` | string | No | memory | Limiter type: `memory`, `redis` |

**Examples**:
```go
// Basic rate limiting
// @Route(method="POST", path="/api/upload")
// @RateLimit(limit=10, window="1m")
func UploadFile(c *gin.Context) {}

// Redis-based (for distributed systems)
// @RateLimit(limit=100, window="1h", type="redis")
func ApiEndpoint(c *gin.Context) {}
```

### @RateLimitByIP

**Purpose**: Rate limit by client IP address  
**Type**: Middleware decorator  
**Factory**: `createRateLimitByIPMiddleware`

**Syntax**:
```go
// @RateLimitByIP(limit=COUNT, window="DURATION")
```

**Examples**:
```go
// Limit by IP for public endpoints
// @Route(method="POST", path="/contact")
// @RateLimitByIP(limit=3, window="10m")
func ContactForm(c *gin.Context) {}
```

**Behavior**: Rate limit key = client IP address

### @RateLimitByUser

**Purpose**: Rate limit by authenticated user  
**Type**: Middleware decorator  
**Factory**: `createRateLimitByUserMiddleware`

**Syntax**:
```go
// @RateLimitByUser(limit=COUNT, window="DURATION")
```

**Examples**:
```go
// Per-user rate limiting
// @Route(method="POST", path="/api/actions")
// @Auth(role="user")
// @RateLimitByUser(limit=100, window="1h")
func UserActions(c *gin.Context) {}
```

**Behavior**: Rate limit key = user ID from authentication context

### @RateLimitByEndpoint

**Purpose**: Rate limit by endpoint globally  
**Type**: Middleware decorator  
**Factory**: `createRateLimitByEndpointMiddleware`

**Syntax**:
```go
// @RateLimitByEndpoint(limit=COUNT, window="DURATION")
```

**Examples**:
```go
// Global endpoint rate limiting
// @Route(method="GET", path="/api/search")
// @RateLimitByEndpoint(limit=1000, window="1m")
func Search(c *gin.Context) {}
```

**Behavior**: Rate limit key = endpoint path

---

## Validation

### @Validate

**Purpose**: General struct validation  
**Type**: Middleware decorator  
**Factory**: `createValidateMiddleware`

**Syntax**:
```go
// @Validate()
```

**Examples**:
```go
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required,min=2,max=50"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"min=18,max=120"`
}

// @Route(method="POST", path="/users")
// @Validate()
func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    // Validation automatically applied
}
```

**Validation Tags**:
| Tag | Description | Example |
|-----|-------------|---------|
| `required` | Field is mandatory | `binding:"required"` |
| `email` | Valid email format | `binding:"email"` |
| `min` | Minimum length/value | `binding:"min=3"` |
| `max` | Maximum length/value | `binding:"max=50"` |
| `len` | Exact length | `binding:"len=10"` |
| `oneof` | One of specified values | `binding:"oneof=red green blue"` |
| `url` | Valid URL format | `binding:"url"` |
| `uuid` | Valid UUID format | `binding:"uuid"` |

### @ValidateJSON

**Purpose**: JSON body validation  
**Type**: Middleware decorator  
**Factory**: `createValidateJSONMiddleware`

**Syntax**:
```go
// @ValidateJSON()
```

**Examples**:
```go
// @Route(method="POST", path="/data")
// @ValidateJSON()
func ReceiveData(c *gin.Context) {
    var data map[string]interface{}
    // JSON structure automatically validated
}
```

### @ValidateQuery

**Purpose**: Query parameter validation  
**Type**: Middleware decorator  
**Factory**: `createValidateQueryMiddleware`

**Syntax**:
```go
// @ValidateQuery()
```

**Examples**:
```go
type SearchParams struct {
    Query    string `form:"q" binding:"required"`
    Page     int    `form:"page" binding:"min=1"`
    Limit    int    `form:"limit" binding:"min=1,max=100"`
    Category string `form:"category" binding:"omitempty,oneof=tech business"`
}

// @Route(method="GET", path="/search")
// @ValidateQuery()
func Search(c *gin.Context) {
    var params SearchParams
    // Query parameters automatically validated
}
```

### @ValidateParams

**Purpose**: Path parameter validation  
**Type**: Middleware decorator  
**Factory**: `createValidateParamsMiddleware`

**Syntax**:
```go
// @ValidateParams(param1="rule1", param2="rule2")
```

**Validation Rules**:
| Rule | Description | Example |
|------|-------------|---------|
| `uuid` | Valid UUID format | `id="uuid"` |
| `alpha` | Only alphabetic characters | `slug="alpha"` |
| `alphanum` | Alphanumeric characters | `code="alphanum"` |
| `numeric` | Only numeric characters | `id="numeric"` |
| `email` | Valid email format | `email="email"` |
| `url` | Valid URL format | `url="url"` |
| `phone` | Valid phone number | `phone="phone"` |
| `cpf` | Valid Brazilian CPF | `cpf="cpf"` |
| `cnpj` | Valid Brazilian CNPJ | `cnpj="cnpj"` |
| `datetime` | Valid date/time format | `date="datetime"` |

**Examples**:
```go
// @Route(method="GET", path="/users/:id")
// @ValidateParams(id="uuid")
func GetUser(c *gin.Context) {}

// @Route(method="GET", path="/posts/:slug")
// @ValidateParams(slug="alpha")
func GetPost(c *gin.Context) {}

// @Route(method="GET", path="/products/:id/reviews/:reviewId")
// @ValidateParams(id="numeric", reviewId="uuid")
func GetReview(c *gin.Context) {}
```

---

## CORS & WebSocket

### @CORS

**Purpose**: Cross-Origin Resource Sharing configuration  
**Type**: Middleware decorator  
**Factory**: `createCORSMiddleware`

**Syntax**:
```go
// @CORS(origins="ORIGINS")
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `origins` | string | No | Allowed origins (comma-separated) |

**Examples**:
```go
// Allow all origins (development)
// @Route(method="GET", path="/public/api")
// @CORS(origins="*")
func PublicAPI(c *gin.Context) {}

// Specific origin
// @CORS(origins="https://app.example.com")
func RestrictedAPI(c *gin.Context) {}

// Multiple origins
// @CORS(origins="https://app.example.com,https://admin.example.com")
func MultiOriginAPI(c *gin.Context) {}

// Wildcard domains
// @CORS(origins="*.example.com")
func SubdomainAPI(c *gin.Context) {}
```

**Default Headers**:
- `Access-Control-Allow-Origin`
- `Access-Control-Allow-Methods`: GET, POST, PUT, DELETE, OPTIONS
- `Access-Control-Allow-Headers`: Origin, Content-Type, Authorization

### @WebSocket

**Purpose**: WebSocket connection handler  
**Type**: Handler decorator  
**Factory**: `createWebSocketMiddleware`

**Syntax**:
```go
// @WebSocket()
```

**Examples**:
```go
// Basic WebSocket
// @Route(method="GET", path="/ws")
// @WebSocket()
func WebSocketHandler(c *gin.Context) {}

// Authenticated WebSocket
// @Route(method="GET", path="/ws/private")
// @WebSocket()
// @Auth(role="user")
func AuthenticatedWebSocket(c *gin.Context) {}
```

**Features**:
- Automatic connection upgrade
- Message broadcasting
- Connection management
- Group/room support
- JSON message handling

### @WebSocketStats

**Purpose**: WebSocket statistics endpoint  
**Type**: Handler decorator  
**Factory**: `createWebSocketStatsMiddleware`

**Syntax**:
```go
// @WebSocketStats()
```

**Examples**:
```go
// @Route(method="GET", path="/admin/ws/stats")
// @Auth(role="admin")
// @WebSocketStats()
func WebSocketStats(c *gin.Context) {}
```

**Response**:
```json
{
  "active_connections": 142,
  "total_connections": 1523,
  "messages_sent": 5437,
  "messages_received": 4891,
  "groups": {
    "chat_room_1": 23,
    "chat_room_2": 15
  }
}
```

---

## Metrics & Monitoring

### @Metrics

**Purpose**: Automatic metrics collection  
**Type**: Middleware decorator  
**Factory**: `createMetricsMiddleware`

**Syntax**:
```go
// @Metrics()
```

**Examples**:
```go
// @Route(method="POST", path="/api/orders")
// @Auth(role="user")
// @Metrics()
func CreateOrder(c *gin.Context) {}
```

**Collected Metrics**:
- Request count by method, endpoint, status
- Request duration histogram
- Request/response size
- Active requests gauge
- Error rate

### @Prometheus

**Purpose**: Prometheus metrics endpoint  
**Type**: Handler decorator  
**Factory**: `createPrometheusMiddleware`

**Syntax**:
```go
// @Prometheus()
```

**Examples**:
```go
// @Route(method="GET", path="/metrics")
// @Prometheus()
func MetricsEndpoint(c *gin.Context) {}
```

**Exposed Metrics**:
```
# HELP http_requests_total Total HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",endpoint="/users",status="200"} 1523

# HELP http_request_duration_seconds HTTP request duration
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{le="0.1"} 1234
```

### @HealthCheck

**Purpose**: Health check endpoint  
**Type**: Handler decorator  
**Factory**: `createHealthCheckMiddleware`

**Syntax**:
```go
// @HealthCheck()
```

**Examples**:
```go
// @Route(method="GET", path="/health")
// @HealthCheck()
func HealthCheck(c *gin.Context) {}
```

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0",
  "checks": {
    "database": "healthy",
    "redis": "healthy",
    "external_api": "healthy"
  }
}
```

### @HealthCheckWithTracing

**Purpose**: Health check with distributed tracing  
**Type**: Handler decorator  
**Factory**: `createHealthCheckWithTracingMiddleware`

**Syntax**:
```go
// @HealthCheckWithTracing()
```

**Examples**:
```go
// @Route(method="GET", path="/health/traced")
// @HealthCheckWithTracing()
func TracedHealthCheck(c *gin.Context) {}
```

---

## Tracing & Observability

### @Telemetry

**Purpose**: OpenTelemetry distributed tracing  
**Type**: Middleware decorator  
**Factory**: `createTelemetryMiddleware`

**Syntax**:
```go
// @Telemetry()
```

**Examples**:
```go
// @Route(method="POST", path="/api/process")
// @Auth(role="user")
// @Telemetry()
func ProcessRequest(c *gin.Context) {}
```

**Features**:
- Automatic span creation
- Request/response data capture
- Error tracking
- Propagation context handling
- Integration with Jaeger/Zipkin

### @TraceMiddleware

**Purpose**: Named middleware tracing  
**Type**: Middleware decorator  
**Factory**: `createTraceMiddlewareWrapper`

**Syntax**:
```go
// @TraceMiddleware(name="SPAN_NAME")
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | No | Custom span name |

**Examples**:
```go
// @Route(method="POST", path="/auth/login")
// @TraceMiddleware(name="authentication")
func Login(c *gin.Context) {}

// @Route(method="POST", path="/payments")
// @TraceMiddleware(name="payment_processing")
func ProcessPayment(c *gin.Context) {}
```

### @TracingStats

**Purpose**: Tracing statistics endpoint  
**Type**: Handler decorator  
**Factory**: `createTracingStatsMiddleware`

**Syntax**:
```go
// @TracingStats()
```

**Examples**:
```go
// @Route(method="GET", path="/admin/tracing/stats")
// @Auth(role="admin")
// @TracingStats()
func TracingStats(c *gin.Context) {}
```

**Response**:
```json
{
  "enabled": true,
  "service_name": "my-api",
  "spans_created": 15234,
  "spans_exported": 15198,
  "export_errors": 36,
  "sampling_rate": 1.0
}
```

### @InstrumentedHandler

**Purpose**: Custom handler instrumentation  
**Type**: Middleware decorator  
**Factory**: `createInstrumentedHandlerMiddleware`

**Syntax**:
```go
// @InstrumentedHandler(name="HANDLER_NAME")
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | No | Custom handler name for instrumentation |

**Examples**:
```go
// @Route(method="GET", path="/api/data")
// @InstrumentedHandler(name="data_fetcher")
func GetData(c *gin.Context) {}
```

---

## Documentation

### @OpenAPIJSON

**Purpose**: OpenAPI 3.0 JSON specification endpoint  
**Type**: Handler decorator  
**Factory**: `createOpenAPIJSONMiddleware`

**Syntax**:
```go
// @OpenAPIJSON()
```

**Examples**:
```go
// @Route(method="GET", path="/docs/openapi.json")
// @OpenAPIJSON()
func OpenAPISpec(c *gin.Context) {}
```

### @OpenAPIYAML

**Purpose**: OpenAPI 3.0 YAML specification endpoint  
**Type**: Handler decorator  
**Factory**: `createOpenAPIYAMLMiddleware`

**Syntax**:
```go
// @OpenAPIYAML()
```

**Examples**:
```go
// @Route(method="GET", path="/docs/openapi.yaml")
// @OpenAPIYAML()
func OpenAPIYAML(c *gin.Context) {}
```

### @SwaggerUI

**Purpose**: Swagger UI interface for interactive API documentation  
**Type**: Handler decorator  
**Factory**: `createSwaggerUIMiddleware`

**Syntax**:
```go
// @SwaggerUI()
```

**Examples**:
```go
// @Route(method="GET", path="/swagger")
// @SwaggerUI()
func SwaggerInterface(c *gin.Context) {}

// @Route(method="GET", path="/api-docs")
// @SwaggerUI()
func APIDocs(c *gin.Context) {}
```

**Features**:
- Interactive API testing interface
- Automatic integration with OpenAPI specification from `/decorators/openapi.json`
- Modern responsive UI
- Try-it-out functionality for all HTTP methods
- Syntax highlighting and validation

### @Schema

**Purpose**: Entity/struct schema definition for OpenAPI documentation  
**Type**: Entity decorator  
**Factory**: None (Documentation only)

**Syntax**:
```go
// @Schema()
// @Description("Entity description")
type EntityName struct {
    Field1 type `json:"field1" validate:"validation_rules"` // Field description
    Field2 type `json:"field2"` // Another field
}
```

**Examples**:
```go
// User entity with validation
// @Schema()
// @Description("User entity representing a registered user in the system")
type User struct {
    ID       int    `json:"id" validate:"required"`                    // User unique identifier
    Name     string `json:"name" validate:"required,min=2,max=100"`    // Full name of the user
    Email    string `json:"email" validate:"required,email"`           // Email address (must be unique)
    Age      *int   `json:"age,omitempty" validate:"min=18,max=120"`   // User age (optional)
    IsActive bool   `json:"isActive"`                                  // Whether the user is active
    Role     string `json:"role" validate:"oneof=admin user guest"`    // User role in the system
}

// Request/Response schemas
// @Schema()
// @Description("Request payload for creating a new user")
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`    // Full name
    Email    string `json:"email" validate:"required,email"`           // Email address
    Password string `json:"password" validate:"required,min=8"`        // Password (minimum 8 characters)
    Role     string `json:"role" validate:"oneof=admin user guest"`    // User role
}

// Error response schema
// @Schema()
// @Description("Standard error response format")
type ErrorResponse struct {
    Error   string                 `json:"error" validate:"required"`     // Error message
    Code    int                    `json:"code" validate:"required"`      // HTTP status code
    Details map[string]interface{} `json:"details,omitempty"`             // Additional error details
}

// Array response schemas
// @Schema()
// @Description("Paginated response containing a list of users")
type ListUsersResponse struct {
    Users   []UserResponse `json:"users" validate:"required"`       // List of users
    Total   int            `json:"total" validate:"required"`       // Total number of users
    Page    int            `json:"page" validate:"required,min=1"`  // Current page number
    Limit   int            `json:"limit" validate:"required,min=1"` // Number of items per page
    HasNext bool           `json:"hasNext"`                         // Whether there are more pages
    HasPrev bool           `json:"hasPrev"`                         // Whether there are previous pages
}
```

**Features**:
- **Automatic OpenAPI integration**: Schemas are automatically included in OpenAPI specification
- **Validation extraction**: Uses `validate` struct tags for field constraints
- **JSON tag support**: Respects `json` tags for field naming and serialization
- **Field documentation**: Extracts descriptions from inline comments after field declarations
- **Type mapping**: Automatically maps Go types to OpenAPI types (string, integer, number, boolean, array, object)
- **Nested schemas**: Supports complex nested structures and references
- **Constraint support**: Extracts min/max values, string lengths, enum values from validation tags
- **Required fields**: Marks fields as required based on validation tags

**Supported Validation Tags**:
- `required`: Marks field as required
- `min=N`, `max=N`: Sets minimum/maximum values for numbers or string lengths
- `oneof=val1 val2`: Creates enum constraint with allowed values
- `email`: Adds email format validation (OpenAPI format: email)
- `len=N`: Sets exact length requirement

**Go Type Mapping**:
- `string` → `string`
- `int`, `int32`, `int64` → `integer` (with appropriate format)
- `float32`, `float64` → `number` (with float/double format)
- `bool` → `boolean`
- `[]T` → `array` (with items of type T)
- `map[K]V` → `object`
- `*T` → Same as T (pointer types are unwrapped)
- Custom structs → `object` (with reference to schema)

### @Description

**Purpose**: Handler description for documentation  
**Type**: Documentation decorator  
**Factory**: None (documentation only)

**Syntax**:
```go
// @Description("DESCRIPTION")
```

**Examples**:
```go
// @Route(method="POST", path="/users")
// @Description("Create a new user account with validation")
func CreateUser(c *gin.Context) {}
```

### @Summary

**Purpose**: Short handler summary  
**Type**: Documentation decorator  
**Factory**: None (documentation only)

**Syntax**:
```go
// @Summary("SUMMARY")
```

**Examples**:
```go
// @Route(method="GET", path="/users/:id")
// @Summary("Get user by ID")
func GetUser(c *gin.Context) {}
```

### @Tag

**Purpose**: API grouping tag  
**Type**: Documentation decorator  
**Factory**: None (documentation only)

**Syntax**:
```go
// @Tag("TAG_NAME")
```

**Examples**:
```go
// @Route(method="GET", path="/users")
// @Tag("User Management")
func GetUsers(c *gin.Context) {}
```

### @Group

**Purpose**: Route grouping  
**Type**: Documentation decorator  
**Factory**: None (documentation only)

**Syntax**:
```go
// @Group("GROUP_PATH")
```

**Examples**:
```go
// @Route(method="GET", path="/api/v1/users")
// @Group("api/v1")
func GetUsers(c *gin.Context) {}
```

### @Param

**Purpose**: Parameter documentation  
**Type**: Documentation decorator  
**Factory**: None (documentation only)

**Syntax**:
```go
// @Param(name="NAME", type="TYPE", location="LOCATION", required=BOOL, description="DESC")
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Parameter name |
| `type` | string | Yes | Parameter type (string, integer, boolean, etc.) |
| `location` | string | Yes | Parameter location (path, query, header, body) |
| `required` | boolean | No | Whether parameter is required |
| `description` | string | No | Parameter description |

**Examples**:
```go
// @Route(method="GET", path="/users/:id")
// @Param(name="id", type="string", location="path", required=true, description="User ID")
// @Param(name="include", type="string", location="query", required=false, description="Related data to include")
func GetUser(c *gin.Context) {}
```

### @Response

**Purpose**: Response documentation with schema linking  
**Type**: Documentation decorator  
**Factory**: None (documentation only)

**Syntax**:
```go
// @Response(code=CODE, description="DESCRIPTION", type="SCHEMA_NAME", example="EXAMPLE")
```

**Parameters**:
| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `code` | integer | Yes | HTTP status code |
| `description` | string | Yes | Response description |
| `type` | string | No | Schema name to reference (must be registered with @Schema) |
| `example` | string | No | Example response data |

**Examples**:
```go
// Basic response documentation
// @Route(method="POST", path="/users")
// @Response(code=201, description="User created successfully")
// @Response(code=400, description="Invalid input data")
// @Response(code=409, description="User already exists")
func CreateUser(c *gin.Context) {}

// Response with schema references
// @Route(method="POST", path="/api/users")
// @Response(code=201, description="User created successfully", type="UserResponse")
// @Response(code=400, description="Invalid user data", type="ErrorResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
func CreateUserWithSchema(c *gin.Context) {}

// Array response with schema
// @Route(method="GET", path="/api/users")
// @Response(code=200, description="List of users retrieved successfully", type="ListUsersResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
func ListUsers(c *gin.Context) {}
```

**Features**:
- **Schema Integration**: When `type` parameter is provided, automatically links to registered @Schema
- **Array Support**: Supports array schemas like `ListUsersResponse` that contain arrays of other schemas
- **Swagger UI Integration**: Responses with schemas are fully interactive in Swagger UI
- **Automatic Reference**: Creates OpenAPI `$ref` to `#/components/schemas/SchemaName`
- **Type Validation**: Validates that referenced schema is registered during code generation

---

## Configuration Reference

### .deco.yaml Structure

```yaml
# Handler discovery
handlers:
  include:
    - "handlers/**/*.go"
    - "api/**/*.go"
  exclude:
    - "**/*_test.go"

# Code generation
generation:
  output: ".deco/init_decorators.go"
  package: "deco"

# Development settings
dev:
  watch: true
  hot_reload: true
  port: 8080

# Production settings
prod:
  minify: true
  validate: true
  optimize: true

# Cache configuration
cache:
  type: "memory"  # or "redis"
  default_ttl: "5m"
  max_size: 1000
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0

# Rate limiting
rate_limit:
  type: "memory"  # or "redis"
  default_limit: 100
  default_window: "1m"
  redis:
    addr: "localhost:6379"
    password: ""
    db: 1

# Telemetry
telemetry:
  enabled: true
  service_name: "my-api"
  service_version: "1.0.0"
  jaeger:
    endpoint: "http://localhost:14268/api/traces"
  otlp:
    endpoint: "http://localhost:4317"

# Metrics
metrics:
  enabled: true
  path: "/metrics"
  namespace: "myapp"

# WebSocket
websocket:
  read_buffer_size: 1024
  write_buffer_size: 1024
  handshake_timeout: "10s"

# Validation
validation:
  enabled: true
  stop_on_first_error: false

# OpenAPI documentation
openapi:
  title: "My API"
  description: "API documentation"
  version: "1.0.0"
  contact:
    name: "API Support"
    email: "support@example.com"
  license:
    name: "MIT"
```

---

## CLI Reference

### Commands

#### deco init
Initialize a new Deco project.

```bash
deco init [flags]
```

**Flags**:
- `--config string`: Configuration file name (default: .deco.yaml)
- `--force`: Overwrite existing files

#### deco generate
Generate route registration code.

```bash
deco generate [flags]
```

**Flags**:
- `--config string`: Configuration file path
- `--output string`: Output file path
- `--package string`: Package name for generated code
- `--watch`: Watch for file changes and regenerate
- `--minify`: Minify generated code
- `--validate`: Validate generated code
- `--verbose`: Verbose output

#### deco version
Show version information.

```bash
deco version
```

#### deco help
Show help information.

```bash
deco help [command]
```

### Global Flags

- `--config string`: Configuration file path
- `--verbose`: Enable verbose logging
- `--silent`: Suppress all output except errors

### Examples

```bash
# Initialize new project
deco init

# Generate with default config
deco generate

# Generate with custom config
deco generate --config .deco.production.yaml

# Watch mode for development
deco generate --watch

# Production build
deco generate --minify --validate

# Generate to specific output
deco generate --output ./routes/generated.go --package routes
```

---

## Error Handling

### Common Error Responses

#### Authentication Errors
```json
{
  "error": "Authentication required",
  "code": 401
}
```

#### Authorization Errors
```json
{
  "error": "Insufficient privileges",
  "code": 403,
  "required_role": "admin"
}
```

#### Validation Errors
```json
{
  "error": "Validation failed",
  "code": 400,
  "fields": [
    {
      "field": "email",
      "tag": "email",
      "message": "must be a valid email"
    }
  ]
}
```

#### Rate Limit Errors
```json
{
  "error": "Rate limit exceeded",
  "code": 429,
  "retry_after": 60
}
```

---

This completes the comprehensive API reference for the Deco framework. Each decorator is fully documented with syntax, parameters, examples, and behavior specifications. 