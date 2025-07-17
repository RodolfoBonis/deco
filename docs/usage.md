# deco Framework Usage Guide

This guide provides detailed examples and usage patterns for all Deco framework decorators. Learn how to build powerful APIs with simple annotations.

## Table of Contents

- [Getting Started](#getting-started)
- [Route Definition](#route-definition)
- [Authentication & Authorization](#authentication--authorization)
- [Caching Strategies](#caching-strategies)
- [Rate Limiting](#rate-limiting)
- [Validation](#validation)
- [CORS & WebSocket](#cors--websocket)
- [Metrics & Monitoring](#metrics--monitoring)
- [Tracing & Observability](#tracing--observability)
- [Documentation](#documentation)
- [Advanced Examples](#advanced-examples)
- [Configuration](#configuration)

## Getting Started

### Project Setup

```bash
# Install deco CLI
go install github.com/RodolfoBonis/deco/cmd/deco@latest

# Initialize new project
mkdir my-api && cd my-api
deco init

# Create basic structure
mkdir handlers
```

### Basic Handler

```go
// handlers/hello.go
package handlers

import "github.com/gin-gonic/gin"

// @Route(method="GET", path="/hello")
// @Description("Simple hello world endpoint")
func Hello(c *gin.Context) {
    c.JSON(200, gin.H{"message": "Hello, World!"})
}
```

### Generate and Run

```bash
# Generate route registration code
deco generate

# Run your application
go run main.go
```

## Route Definition

### Basic Routes

```go
// @Route(method="GET", path="/users")
func GetUsers(c *gin.Context) {}

// @Route(method="POST", path="/users")
func CreateUser(c *gin.Context) {}

// @Route(method="PUT", path="/users/:id")
func UpdateUser(c *gin.Context) {}

// @Route(method="DELETE", path="/users/:id")
func DeleteUser(c *gin.Context) {}
```

### Path Parameters

```go
// @Route(method="GET", path="/users/:id")
// @Route(method="GET", path="/users/:id/posts/:postId")
// @Route(method="GET", path="/files/*filepath")
func HandleWithParams(c *gin.Context) {
    id := c.Param("id")
    postId := c.Param("postId")
    filepath := c.Param("filepath")
}
```

### Query Parameters

```go
// @Route(method="GET", path="/search")
func SearchUsers(c *gin.Context) {
    query := c.Query("q")
    page := c.DefaultQuery("page", "1")
    limit := c.DefaultQuery("limit", "10")
}
```

## Authentication & Authorization

### Basic Authentication

```go
// @Route(method="GET", path="/profile")
// @Auth()
// @Description("Get user profile - requires authentication")
func GetProfile(c *gin.Context) {
    // Authentication automatically verified
    c.JSON(200, gin.H{"message": "Profile data"})
}
```

### Role-Based Access Control

```go
// @Route(method="GET", path="/admin/users")
// @Auth(role="admin")
// @Description("Admin only - list all users")
func AdminListUsers(c *gin.Context) {
    // Only admin role can access
    c.JSON(200, gin.H{"users": []string{}})
}

// @Route(method="POST", path="/posts")
// @Auth(role="user")
// @Description("User role required to create posts")
func CreatePost(c *gin.Context) {
    // User or higher role required
    c.JSON(201, gin.H{"message": "Post created"})
}
```

### Multiple Authorization Examples

```go
// @Route(method="GET", path="/moderator/reports")
// @Auth(role="moderator")
func ModeratorReports(c *gin.Context) {}

// @Route(method="DELETE", path="/admin/cleanup")
// @Auth(role="admin")
func AdminCleanup(c *gin.Context) {}
```

## Caching Strategies

### Basic Caching

```go
// @Route(method="GET", path="/products")
// @Cache(ttl="5m")
// @Description("Cache product list for 5 minutes")
func GetProducts(c *gin.Context) {
    // Expensive database query cached for 5 minutes
    c.JSON(200, gin.H{"products": []string{}})
}
```

### Cache Types and TTL Options

```go
// Memory cache (default)
// @Route(method="GET", path="/categories")
// @Cache(ttl="10m", type="memory")
func GetCategories(c *gin.Context) {}

// Redis cache
// @Route(method="GET", path="/popular-items")
// @Cache(ttl="1h", type="redis")
func GetPopularItems(c *gin.Context) {}

// Different TTL examples
// @Route(method="GET", path="/flash-sales")
// @Cache(ttl="30s")  // 30 seconds
func GetFlashSales(c *gin.Context) {}

// @Route(method="GET", path="/daily-stats")
// @Cache(ttl="24h")  // 24 hours
func GetDailyStats(c *gin.Context) {}
```

### Advanced Caching Strategies

```go
// Cache by URL only
// @Route(method="GET", path="/public/news")
// @CacheByURL(ttl="15m")
// @Description("Cache based on URL path")
func GetNews(c *gin.Context) {}

// Cache per user
// @Route(method="GET", path="/recommendations")
// @Auth(role="user")
// @CacheByUser(ttl="1h")
// @Description("Cache recommendations per user")
func GetRecommendations(c *gin.Context) {}

// Cache per endpoint
// @Route(method="GET", path="/api/v1/trending")
// @CacheByEndpoint(ttl="30m")
// @Description("Cache trending data per endpoint")
func GetTrending(c *gin.Context) {}
```

### Cache Management

```go
// Cache statistics endpoint
// @Route(method="GET", path="/admin/cache/stats")
// @Auth(role="admin")
// @CacheStats()
// @Description("Get cache statistics")
func CacheStats(c *gin.Context) {}

// Cache invalidation endpoint
// @Route(method="DELETE", path="/admin/cache/invalidate")
// @Auth(role="admin")
// @InvalidateCache()
// @Description("Invalidate cache entries")
func InvalidateCache(c *gin.Context) {}
```

## Rate Limiting

### Basic Rate Limiting

```go
// @Route(method="POST", path="/api/upload")
// @Auth(role="user")
// @RateLimit(limit=10, window="1m")
// @Description("Upload file - max 10 requests per minute")
func UploadFile(c *gin.Context) {}

// @Route(method="POST", path="/api/send-email")
// @RateLimit(limit=5, window="1h")
// @Description("Send email - max 5 per hour")
func SendEmail(c *gin.Context) {}
```

### Rate Limiting Strategies

```go
// Rate limit by client IP
// @Route(method="POST", path="/public/contact")
// @RateLimitByIP(limit=3, window="10m")
// @Description("Contact form - 3 submissions per IP per 10 minutes")
func ContactForm(c *gin.Context) {}

// Rate limit by authenticated user
// @Route(method="POST", path="/api/actions")
// @Auth(role="user")
// @RateLimitByUser(limit=100, window="1h")
// @Description("API actions - 100 per user per hour")
func ApiActions(c *gin.Context) {}

// Rate limit by endpoint
// @Route(method="GET", path="/api/search")
// @RateLimitByEndpoint(limit=1000, window="1m")
// @Description("Search endpoint - 1000 total requests per minute")
func Search(c *gin.Context) {}
```

### Rate Limit Examples by Use Case

```go
// Authentication endpoints
// @Route(method="POST", path="/auth/login")
// @RateLimit(limit=5, window="15m")
func Login(c *gin.Context) {}

// @Route(method="POST", path="/auth/register")
// @RateLimitByIP(limit=3, window="1h")
func Register(c *gin.Context) {}

// API endpoints
// @Route(method="GET", path="/api/data")
// @Auth(role="user")
// @RateLimitByUser(limit=1000, window="1h")
func GetData(c *gin.Context) {}

// Public endpoints
// @Route(method="GET", path="/public/content")
// @RateLimitByIP(limit=100, window="1m")
func PublicContent(c *gin.Context) {}
```

## Validation

### Struct Validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Age      int    `json:"age" binding:"min=18,max=120"`
}

// @Route(method="POST", path="/users")
// @Validate()
// @Description("Create user with automatic validation")
func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // Validation errors automatically handled
        return
    }
    c.JSON(201, gin.H{"message": "User created"})
}
```

### JSON Validation

```go
// @Route(method="POST", path="/api/data")
// @ValidateJSON()
// @Description("Validate JSON body structure")
func ReceiveData(c *gin.Context) {
    var data map[string]interface{}
    // JSON automatically validated and parsed
    c.ShouldBindJSON(&data)
    c.JSON(200, gin.H{"received": data})
}
```

### Query Parameter Validation

```go
type SearchQuery struct {
    Query    string `form:"q" binding:"required"`
    Page     int    `form:"page" binding:"min=1"`
    Limit    int    `form:"limit" binding:"min=1,max=100"`
    Category string `form:"category" binding:"omitempty,oneof=tech business sports"`
}

// @Route(method="GET", path="/search")
// @ValidateQuery()
// @Description("Search with validated query parameters")
func SearchWithValidation(c *gin.Context) {
    var query SearchQuery
    // Query parameters automatically validated
    c.ShouldBindQuery(&query)
    c.JSON(200, gin.H{"results": []string{}})
}
```

### Path Parameter Validation

```go
// @Route(method="GET", path="/users/:id")
// @ValidateParams(id="uuid")
// @Description("Get user by UUID")
func GetUserByUUID(c *gin.Context) {
    id := c.Param("id") // Validated as UUID
    c.JSON(200, gin.H{"user_id": id})
}

// @Route(method="GET", path="/posts/:slug")
// @ValidateParams(slug="alpha")
// @Description("Get post by alphabetic slug")
func GetPostBySlug(c *gin.Context) {
    slug := c.Param("slug") // Validated as alphabetic
    c.JSON(200, gin.H{"post_slug": slug})
}

// @Route(method="GET", path="/products/:id/reviews/:reviewId")
// @ValidateParams(id="numeric", reviewId="uuid")
// @Description("Get review with validated parameters")
func GetProductReview(c *gin.Context) {
    productId := c.Param("id")      // Validated as numeric
    reviewId := c.Param("reviewId") // Validated as UUID
    c.JSON(200, gin.H{"product_id": productId, "review_id": reviewId})
}
```

### Custom Validation Rules

```go
// @Route(method="POST", path="/users/profile")
// @ValidateParams(phone="phone", cpf="cpf", cnpj="cnpj")
// @Description("Brazilian-specific validations")
func UpdateBrazilianProfile(c *gin.Context) {
    // phone: validates Brazilian phone format
    // cpf: validates Brazilian CPF
    // cnpj: validates Brazilian CNPJ
}
```

## CORS & WebSocket

### CORS Configuration

```go
// Allow all origins (development)
// @Route(method="GET", path="/public/api")
// @CORS(origins="*")
// @Description("Public API with open CORS")
func PublicAPI(c *gin.Context) {}

// Specific origins
// @Route(method="POST", path="/api/data")
// @CORS(origins="https://app.example.com")
// @Description("API restricted to specific domain")
func RestrictedAPI(c *gin.Context) {}

// Multiple origins
// @Route(method="GET", path="/api/shared")
// @CORS(origins="https://app.example.com,https://admin.example.com")
// @Description("API accessible from multiple domains")
func SharedAPI(c *gin.Context) {}

// Wildcard domains
// @Route(method="GET", path="/api/subdomains")
// @CORS(origins="*.example.com")
// @Description("API accessible from all subdomains")
func SubdomainAPI(c *gin.Context) {}
```

### WebSocket Handlers

```go
// Basic WebSocket connection
// @Route(method="GET", path="/ws")
// @WebSocket()
// @Description("WebSocket connection endpoint")
func BasicWebSocket(c *gin.Context) {
    // WebSocket connection automatically handled
    // Real-time messaging enabled
}

// Authenticated WebSocket
// @Route(method="GET", path="/ws/private")
// @WebSocket()
// @Auth(role="user")
// @Description("Authenticated WebSocket connection")
func AuthenticatedWebSocket(c *gin.Context) {
    // Only authenticated users can connect
}

// WebSocket with metrics
// @Route(method="GET", path="/ws/monitored")
// @WebSocket()
// @Metrics()
// @Description("WebSocket with monitoring")
func MonitoredWebSocket(c *gin.Context) {
    // Connection metrics automatically collected
}
```

### WebSocket Statistics

```go
// @Route(method="GET", path="/admin/websocket/stats")
// @Auth(role="admin")
// @WebSocketStats()
// @Description("WebSocket connection statistics")
func WebSocketStatistics(c *gin.Context) {
    // Returns active connections, messages sent/received, etc.
}
```

## Metrics & Monitoring

### Basic Metrics

```go
// @Route(method="POST", path="/api/orders")
// @Auth(role="user")
// @Metrics()
// @Description("Create order with metrics collection")
func CreateOrder(c *gin.Context) {
    // Automatically tracks:
    // - Request count
    // - Response time
    // - Status codes
    // - Error rates
}
```

### Prometheus Integration

```go
// @Route(method="GET", path="/metrics")
// @Prometheus()
// @Description("Prometheus metrics endpoint")
func PrometheusMetrics(c *gin.Context) {
    // Exposes metrics in Prometheus format
    // Accessible at /metrics for scraping
}
```

### Health Checks

```go
// Basic health check
// @Route(method="GET", path="/health")
// @HealthCheck()
// @Description("Basic health check endpoint")
func HealthCheck(c *gin.Context) {
    // Returns service health status
}

// Health check with tracing
// @Route(method="GET", path="/health/traced")
// @HealthCheckWithTracing()
// @Description("Health check with distributed tracing")
func TracedHealthCheck(c *gin.Context) {
    // Health check with OpenTelemetry tracing
}
```

### Custom Metrics Examples

```go
// Payment processing metrics
// @Route(method="POST", path="/payments")
// @Auth(role="user")
// @Metrics()
// @RateLimit(limit=10, window="1m")
func ProcessPayment(c *gin.Context) {}

// File upload metrics
// @Route(method="POST", path="/upload")
// @Auth(role="user")
// @Metrics()
// @RateLimit(limit=5, window="1m")
func UploadFile(c *gin.Context) {}

// Search metrics
// @Route(method="GET", path="/search")
// @Metrics()
// @Cache(ttl="5m")
func SearchWithMetrics(c *gin.Context) {}
```

## Tracing & Observability

### Basic Tracing

```go
// @Route(method="POST", path="/api/process")
// @Auth(role="user")
// @Telemetry()
// @Description("Process request with distributed tracing")
func ProcessWithTracing(c *gin.Context) {
    // OpenTelemetry traces automatically created
    // Spans include request/response data
}
```

### Named Trace Middleware

```go
// @Route(method="POST", path="/auth/login")
// @TraceMiddleware(name="authentication")
// @Description("Login with named tracing")
func LoginWithNamedTrace(c *gin.Context) {
    // Creates trace span named "authentication"
}

// @Route(method="POST", path="/payments/process")
// @Auth(role="user")
// @TraceMiddleware(name="payment_processing")
// @Description("Payment processing with named trace")
func PaymentWithTrace(c *gin.Context) {
    // Creates trace span named "payment_processing"
}
```

### Instrumented Handlers

```go
// @Route(method="GET", path="/api/data")
// @InstrumentedHandler(name="data_fetcher")
// @Description("Handler with custom instrumentation")
func InstrumentedDataFetcher(c *gin.Context) {
    // Handler wrapped with custom instrumentation
}
```

### Tracing Statistics

```go
// @Route(method="GET", path="/admin/tracing/stats")
// @Auth(role="admin")
// @TracingStats()
// @Description("Get distributed tracing statistics")
func TracingStatistics(c *gin.Context) {
    // Returns tracing configuration and stats
}
```

### Combined Observability

```go
// @Route(method="POST", path="/api/critical-operation")
// @Auth(role="admin")
// @Telemetry()
// @Metrics()
// @TraceMiddleware(name="critical_operation")
// @RateLimit(limit=5, window="1m")
// @Description("Critical operation with full observability")
func CriticalOperation(c *gin.Context) {
    // Full observability stack:
    // - Distributed tracing
    // - Metrics collection
    // - Named traces
    // - Rate limiting
}
```

## Documentation

### OpenAPI Integration

```go
// OpenAPI JSON endpoint
// @Route(method="GET", path="/docs/json")
// @OpenAPIJSON()
// @Description("OpenAPI 3.0 specification in JSON format")
func OpenAPIJSONDocs(c *gin.Context) {
    // Returns complete OpenAPI 3.0 spec
}

// OpenAPI YAML endpoint
// @Route(method="GET", path="/docs/yaml")
// @OpenAPIYAML()
// @Description("OpenAPI 3.0 specification in YAML format")
func OpenAPIYAMLDocs(c *gin.Context) {
    // Returns OpenAPI spec in YAML format
}

// Swagger UI interactive documentation
// @Route(method="GET", path="/swagger")
// @SwaggerUI()
// @Description("Interactive API documentation with Swagger UI")
func SwaggerUIDocs(c *gin.Context) {
    // Returns interactive Swagger UI interface
}
```

### Entity Schemas

The framework automatically registers struct schemas for Swagger documentation:

```go
// User entity for API documentation
// @Schema()
// @Description("User entity representing a registered user")
type User struct {
    ID       int    `json:"id" validate:"required"`                    // User unique identifier
    Name     string `json:"name" validate:"required,min=2,max=100"`    // Full name
    Email    string `json:"email" validate:"required,email"`           // Email address
    Age      *int   `json:"age,omitempty" validate:"min=18,max=120"`   // User age (optional)
    IsActive bool   `json:"isActive"`                                  // Account status
    Role     string `json:"role" validate:"oneof=admin user guest"`    // User role
}

// Request payload for user creation
// @Schema()  
// @Description("Request payload for creating a new user")
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`    // Full name
    Email    string `json:"email" validate:"required,email"`           // Email address
    Password string `json:"password" validate:"required,min=8"`        // Password
    Role     string `json:"role" validate:"oneof=admin user guest"`    // User role
}

// Standard error response
// @Schema()
// @Description("Standard error response format")
type ErrorResponse struct {
    Error   string                 `json:"error" validate:"required"`     // Error message
    Code    int                    `json:"code" validate:"required"`      // HTTP status code
    Details map[string]interface{} `json:"details,omitempty"`             // Additional details
}
```

**Schema Features:**
- **Automatic validation extraction**: Uses `validate` tags for constraints
- **JSON tag support**: Respects `json` tags for field names
- **Field descriptions**: Extracts from inline comments
- **Type mapping**: Converts Go types to OpenAPI types
- **Nested schemas**: Supports complex nested structures
- **Automatic endpoint integration**: Schemas are automatically linked to request/response bodies

### Schema-Endpoint Integration

When you use a schema name in `@Param(type="SchemaName")` or when the framework detects common schema patterns, it automatically connects them to endpoints:

```go
// This endpoint will automatically use the CreateUserRequest schema for request body
// @Route("POST", "/api/users")
// @Param(name="user", type="CreateUserRequest", location="body", required=true, description="User data")
// @Response(code=201, description="User created successfully")
// @Response(code=400, description="Invalid user data")
func CreateUser(c *gin.Context) {
    var req CreateUserRequest // Schema will be referenced in OpenAPI
    // ... handler logic
}
```

**Automatic Schema Detection:**
- **Request Bodies**: When `location="body"` and `type` matches a registered schema
- **Response Bodies**: Automatically maps common patterns:
  - 2xx responses → Tries to find schemas ending with "Response"
  - 4xx/5xx responses → Tries to find "ErrorResponse" or "Error" schemas
- **Schema References**: Uses `$ref` to reference schemas in `components.schemas`

**Result in Swagger UI:**
- Interactive forms with proper field validation
- Dropdown menus for enum values
- Required field indicators
- Field descriptions and examples
- Proper request/response body documentation

### Detailed Documentation

```go
// @Route(method="POST", path="/users")
// @Auth(role="admin")
// @ValidateJSON()
// @RateLimit(limit=10, window="1m")
// @Cache(ttl="1m")
// @Metrics()
// @Description("Create a new user account with full validation and security")
// @Summary("Create User")
// @Tag("User Management")
// @Param(name="user", type="object", location="body", required=true, description="User data object")
// @Response(code=201, description="User created successfully")
// @Response(code=400, description="Invalid request data or validation errors")
// @Response(code=401, description="Authentication required")
// @Response(code=403, description="Insufficient privileges - admin role required")
// @Response(code=429, description="Rate limit exceeded - max 10 requests per minute")
func CreateUserWithDocs(c *gin.Context) {
    c.JSON(201, gin.H{"message": "User created"})
}
```

### API Grouping

```go
// @Route(method="GET", path="/api/v1/users")
// @Group("api/v1")
// @Tag("Users")
// @Description("List all users")
func ListUsers(c *gin.Context) {}

// @Route(method="GET", path="/api/v1/products")
// @Group("api/v1")
// @Tag("Products")
// @Description("List all products")
func ListProducts(c *gin.Context) {}

// @Route(method="GET", path="/api/v2/users")
// @Group("api/v2")
// @Tag("Users V2")
// @Description("List users - version 2 with enhanced data")
func ListUsersV2(c *gin.Context) {}
```

### Parameter Documentation

```go
// @Route(method="GET", path="/users/:id")
// @Param(name="id", type="string", location="path", required=true, description="User unique identifier")
// @Param(name="include", type="string", location="query", required=false, description="Related data to include (profile,preferences)")
// @Param(name="format", type="string", location="query", required=false, description="Response format (json,xml)")
// @Response(code=200, description="User data retrieved successfully")
// @Response(code=404, description="User not found")
// @Tag("Users")
// @Summary("Get User by ID")
// @Description("Retrieve detailed user information by unique identifier")
func GetUserWithDetailedDocs(c *gin.Context) {
    c.JSON(200, gin.H{"user": "data"})
}
```

## Advanced Examples

### E-commerce API

```go
// Product listing with caching and rate limiting
// @Route(method="GET", path="/api/products")
// @Cache(ttl="10m", type="redis")
// @RateLimit(limit=100, window="1m")
// @Metrics()
// @CORS(origins="*.shop.example.com")
// @Description("Get product catalog with caching")
// @Tag("Products")
func GetProducts(c *gin.Context) {}

// Order creation with full security
// @Route(method="POST", path="/api/orders")
// @Auth(role="user")
// @ValidateJSON()
// @RateLimit(limit=5, window="1m")
// @Metrics()
// @Telemetry()
// @TraceMiddleware(name="order_creation")
// @Description("Create new order with payment processing")
// @Tag("Orders")
func CreateOrder(c *gin.Context) {}

// Admin dashboard with comprehensive monitoring
// @Route(method="GET", path="/admin/dashboard")
// @Auth(role="admin")
// @Cache(ttl="1m")
// @Metrics()
// @Telemetry()
// @RateLimit(limit=50, window="1m")
// @Description("Admin dashboard with real-time metrics")
// @Tag("Admin")
func AdminDashboard(c *gin.Context) {}
```

### Real-time Chat API

```go
// WebSocket chat connection
// @Route(method="GET", path="/chat/ws")
// @WebSocket()
// @Auth(role="user")
// @Metrics()
// @Telemetry()
// @Description("Real-time chat WebSocket connection")
func ChatWebSocket(c *gin.Context) {}

// Chat history with caching
// @Route(method="GET", path="/chat/history/:roomId")
// @Auth(role="user")
// @Cache(ttl="5m")
// @ValidateParams(roomId="uuid")
// @Metrics()
// @Description("Get chat room message history")
func ChatHistory(c *gin.Context) {}

// Send message with rate limiting
// @Route(method="POST", path="/chat/send")
// @Auth(role="user")
// @ValidateJSON()
// @RateLimit(limit=60, window="1m")
// @Metrics()
// @Telemetry()
// @Description("Send chat message")
func SendMessage(c *gin.Context) {}
```

### Payment Processing API

```go
// Process payment with full security and monitoring
// @Route(method="POST", path="/payments/process")
// @Auth(role="user")
// @ValidateJSON()
// @RateLimit(limit=10, window="1h")
// @Metrics()
// @Telemetry()
// @TraceMiddleware(name="payment_processing")
// @CORS(origins="https://checkout.example.com")
// @Description("Process payment with fraud detection")
// @Tag("Payments")
// @Response(code=200, description="Payment processed successfully")
// @Response(code=400, description="Invalid payment data")
// @Response(code=402, description="Payment declined")
// @Response(code=429, description="Rate limit exceeded")
func ProcessPayment(c *gin.Context) {}

// Payment webhook (no auth but with validation)
// @Route(method="POST", path="/payments/webhook")
// @ValidateJSON()
// @RateLimit(limit=1000, window="1m")
// @Metrics()
// @Telemetry()
// @Description("Payment provider webhook")
func PaymentWebhook(c *gin.Context) {}
```

### Monitoring and Management Endpoints

```go
// Comprehensive health check
// @Route(method="GET", path="/system/health")
// @HealthCheckWithTracing()
// @Metrics()
// @Cache(ttl="30s")
// @Description("System health check with dependencies")
func SystemHealth(c *gin.Context) {}

// Application metrics
// @Route(method="GET", path="/system/metrics")
// @Auth(role="admin")
// @Prometheus()
// @Description("Prometheus metrics for monitoring")
func SystemMetrics(c *gin.Context) {}

// Cache management
// @Route(method="GET", path="/admin/cache/stats")
// @Auth(role="admin")
// @CacheStats()
// @Description("Cache performance statistics")
func CacheManagement(c *gin.Context) {}

// WebSocket monitoring
// @Route(method="GET", path="/admin/websocket/stats")
// @Auth(role="admin")
// @WebSocketStats()
// @Description("WebSocket connection statistics")
func WebSocketMonitoring(c *gin.Context) {}

// Distributed tracing stats
// @Route(method="GET", path="/admin/tracing/stats")
// @Auth(role="admin")
// @TracingStats()
// @Description("Distributed tracing statistics")
func TracingMonitoring(c *gin.Context) {}
```

## Configuration

### Basic Configuration (.deco.yaml)

```yaml
# Basic project configuration
handlers:
  include:
    - "handlers/**/*.go"
    - "api/**/*.go"
    - "controllers/**/*.go"
  exclude:
    - "**/*_test.go"
    - "**/*_mock.go"
    - "**/vendor/**"

generation:
  output: ".deco/init_decorators.go"
  package: "deco"

# Development settings
dev:
  watch: true
  hot_reload: true
  port: 8080

# Production optimizations
prod:
  minify: true
  validate: true
  optimize: true
```

### Advanced Configuration

```yaml
# Cache configuration
cache:
  type: "redis"  # "memory" or "redis"
  default_ttl: "5m"
  max_size: 10000
  
  # Redis-specific settings
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
    pool_size: 10

# Rate limiting configuration
rate_limit:
  type: "redis"  # "memory" or "redis"
  default_limit: 100
  default_window: "1m"
  
  # Redis settings (shared with cache)
  redis:
    addr: "localhost:6379"
    password: ""
    db: 1

# Telemetry and tracing
telemetry:
  enabled: true
  service_name: "my-api"
  service_version: "1.0.0"
  environment: "production"
  
  # Jaeger configuration
  jaeger:
    endpoint: "http://localhost:14268/api/traces"
    agent_endpoint: "localhost:6831"
  
  # OTLP configuration
  otlp:
    endpoint: "http://localhost:4317"
    headers:
      api-key: "your-api-key"

# Metrics configuration
metrics:
  enabled: true
  path: "/metrics"
  namespace: "myapp"
  subsystem: "api"
  
  # Custom labels
  labels:
    service: "user-api"
    version: "1.0.0"

# WebSocket configuration
websocket:
  read_buffer_size: 1024
  write_buffer_size: 1024
  handshake_timeout: "10s"
  
  # CORS for WebSocket
  check_origin: true
  allowed_origins:
    - "https://app.example.com"
    - "https://admin.example.com"

# Validation configuration
validation:
  enabled: true
  stop_on_first_error: false
  
  # Custom error messages
  messages:
    required: "This field is required"
    email: "Please provide a valid email address"
    min: "Value must be at least {0} characters"
    max: "Value must be at most {0} characters"

# OpenAPI documentation
openapi:
  title: "My API"
  description: "API documentation generated by Deco"
  version: "1.0.0"
  
  contact:
    name: "API Support"
    email: "support@example.com"
    url: "https://example.com/support"
  
  license:
    name: "MIT"
    url: "https://opensource.org/licenses/MIT"
  
  servers:
    - url: "https://api.example.com"
      description: "Production server"
    - url: "https://staging-api.example.com"
      description: "Staging server"
```

### Environment-Specific Configuration

```yaml
# .deco.yaml (base configuration)
handlers:
  include: ["handlers/**/*.go"]

# .deco.development.yaml
dev:
  watch: true
  hot_reload: true
cache:
  type: "memory"
rate_limit:
  type: "memory"
telemetry:
  enabled: false

# .deco.production.yaml
prod:
  minify: true
  validate: true
cache:
  type: "redis"
  redis:
    addr: "${REDIS_URL}"
rate_limit:
  type: "redis"
telemetry:
  enabled: true
  jaeger:
    endpoint: "${JAEGER_ENDPOINT}"
```

### CLI Usage with Configuration

```bash
# Use specific configuration file
deco generate --config .deco.production.yaml

# Override configuration values
deco generate --config .deco.yaml --cache-type redis --telemetry-enabled true

# Environment-based configuration
APP_ENV=production deco generate

# Development mode with file watching
deco generate --watch --config .deco.development.yaml

# Production build with optimizations
deco generate --minify --validate --config .deco.production.yaml
```

## Complete Documentation Setup

### API Documentation with Swagger UI

Here's a complete example showing how to set up full API documentation with Swagger UI:

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/RodolfoBonis/deco/pkg/decorators"
)

// Documentation endpoints
// @Route(method="GET", path="/docs")
// @Description("HTML documentation page")
func DocsHandler(c *gin.Context) {
    // Serves auto-generated HTML documentation
}

// @Route(method="GET", path="/swagger")
// @SwaggerUI()
// @Description("Interactive Swagger UI documentation")
func SwaggerUIHandler(c *gin.Context) {
    // Serves interactive Swagger UI
}

// @Route(method="GET", path="/api/openapi.json")
// @OpenAPIJSON()
// @Description("OpenAPI 3.0 specification in JSON")
func OpenAPISpecHandler(c *gin.Context) {
    // Serves OpenAPI JSON specification
}

// Example API endpoint with full documentation
// @Route(method="POST", path="/api/users")
// @Auth(role="admin")
// @ValidateJSON()
// @RateLimit(limit=10, window="1m")
// @Cache(ttl="5m")
// @Prometheus()
// @Description("Create a new user with validation and security")
// @Summary("Create User")
// @Tag("Users")
// @Param(name="user", type="object", location="body", required=true, description="User creation data")
// @Response(code=201, description="User created successfully")
// @Response(code=400, description="Invalid user data")
// @Response(code=401, description="Authentication required")
// @Response(code=403, description="Admin role required")
// @Response(code=429, description="Rate limit exceeded")
func CreateUser(c *gin.Context) {
    // Implementation here
}

func main() {
    r := gin.Default()
    
    // Generate and register all decorated routes
    decorators.RegisterRoutes(r)
    
    r.Run(":8080")
}
```

### Available Documentation Endpoints

After running `deco generate`, your application will automatically provide:

| Endpoint | Description | Content |
|----------|-------------|---------|
| `/decorators/docs` | HTML documentation | Interactive route browser |
| `/decorators/swagger-ui` | Swagger UI | Interactive API testing |
| `/decorators/swagger` | Swagger redirect | Redirects to Swagger UI |
| `/decorators/openapi.json` | OpenAPI JSON | Machine-readable API spec |
| `/decorators/openapi.yaml` | OpenAPI YAML | Human-readable API spec |
| `/decorators/docs.json` | Route metadata | Framework route information |

### Accessing Your Documentation

1. **Swagger UI**: Navigate to `http://localhost:8080/decorators/swagger-ui`
   - Interactive interface to test your API
   - Try-it-out functionality for all endpoints
   - Automatic authentication handling
   - Real-time API validation

2. **HTML Documentation**: Visit `http://localhost:8080/decorators/docs`
   - Framework-generated documentation
   - Route statistics and middleware information
   - Parameter and response details

3. **OpenAPI Specification**: 
   - JSON: `http://localhost:8080/decorators/openapi.json`
   - YAML: `http://localhost:8080/decorators/openapi.yaml`
   - Import into any OpenAPI-compatible tool

## Complete API Example with Schemas

Here's a complete example showing how to build a REST API with entity schemas:

```go
package main

import (
    "time"
    "github.com/gin-gonic/gin"
    "github.com/RodolfoBonis/deco/pkg/decorators"
)

// Define your entities with @Schema decorators
// @Schema()
// @Description("User entity representing a registered user")
type User struct {
    ID        int       `json:"id" validate:"required"`                    // User unique identifier
    Name      string    `json:"name" validate:"required,min=2,max=100"`    // Full name
    Email     string    `json:"email" validate:"required,email"`           // Email address
    Age       *int      `json:"age,omitempty" validate:"min=18,max=120"`   // User age (optional)
    IsActive  bool      `json:"isActive"`                                  // Account status
    Role      string    `json:"role" validate:"oneof=admin user guest"`    // User role
    CreatedAt time.Time `json:"createdAt"`                                 // Creation timestamp
    UpdatedAt time.Time `json:"updatedAt"`                                 // Last update timestamp
}

// @Schema()
// @Description("Request payload for creating a new user")
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`    // Full name
    Email    string `json:"email" validate:"required,email"`           // Email address
    Password string `json:"password" validate:"required,min=8"`        // Password (minimum 8 characters)
    Age      *int   `json:"age,omitempty" validate:"min=18,max=120"`   // User age (optional)
    Role     string `json:"role" validate:"oneof=admin user guest"`    // User role (defaults to 'user')
}

// @Schema()
// @Description("User response without sensitive data")
type UserResponse struct {
    ID        int       `json:"id"`        // User unique identifier
    Name      string    `json:"name"`      // Full name
    Email     string    `json:"email"`     // Email address
    Age       *int      `json:"age,omitempty"` // User age
    IsActive  bool      `json:"isActive"`  // Account status
    Role      string    `json:"role"`      // User role
    CreatedAt time.Time `json:"createdAt"` // Creation timestamp
    UpdatedAt time.Time `json:"updatedAt"` // Last update timestamp
}

// @Schema()
// @Description("Standard error response format")
type ErrorResponse struct {
    Error   string                 `json:"error" validate:"required"`     // Error message
    Code    int                    `json:"code" validate:"required"`      // HTTP status code
    Details map[string]interface{} `json:"details,omitempty"`             // Additional error details
}

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

// API handlers that reference the schemas
// @Route("POST", "/api/users")
// @Auth(role="admin")
// @ValidateJSON()
// @RateLimit(limit=10, window="1m")
// @Description("Create a new user account")
// @Summary("Create User")
// @Tag("Users")
// @Param(name="user", type="CreateUserRequest", location="body", required=true, description="User creation data")
// @Response(code=201, description="User created successfully", type="UserResponse")
// @Response(code=400, description="Invalid user data", type="ErrorResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
// @Response(code=403, description="Admin role required", type="ErrorResponse")
func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{
            Error: "Invalid request payload",
            Code:  400,
            Details: map[string]interface{}{"validation_error": err.Error()},
        })
        return
    }
    
    // Create user logic here...
    user := UserResponse{
        ID:        123,
        Name:      req.Name,
        Email:     req.Email,
        Age:       req.Age,
        IsActive:  true,
        Role:      req.Role,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    c.JSON(201, user)
}

// @Route("GET", "/api/users")
// @Auth()
// @Cache(ttl="5m")
// @Description("List all users with pagination")
// @Summary("List Users")
// @Tag("Users")
// @Param(name="page", type="int", location="query", description="Page number", example="1")
// @Param(name="limit", type="int", location="query", description="Items per page", example="10")
// @Response(code=200, description="List of users retrieved successfully", type="ListUsersResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
func ListUsers(c *gin.Context) {
    users := []UserResponse{
        {ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true, Role: "user"},
        {ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: true, Role: "admin"},
    }
    
    response := ListUsersResponse{
        Users:   users,
        Total:   2,
        Page:    1,
        Limit:   10,
        HasNext: false,
        HasPrev: false,
    }
    
    c.JSON(200, response)
}

func main() {
    r := gin.Default()
    
    // Generate and register all decorated routes and schemas
    decorators.RegisterRoutes(r)
    
    r.Run(":8080")
}
```

**Benefits of this approach:**

1. **Automatic Schema Detection**: All structs with `@Schema()` are automatically registered
2. **Rich Documentation**: Field descriptions and validation rules appear in Swagger UI
3. **Type Safety**: Strong typing with validation constraints
4. **Array Support**: Automatic reference resolution for arrays of schemas (e.g., `[]UserResponse` in `ListUsersResponse`)
5. **API Testing**: Swagger UI provides interactive testing with proper request/response schemas
6. **Client Generation**: OpenAPI spec can generate client SDKs for multiple languages

**Result**: Your API documentation will include:
- Complete request/response schemas with field descriptions
- Validation constraints (required fields, min/max values, enums)
- Array schemas with proper item type references
- Interactive forms in Swagger UI for testing

## Interactive Swagger UI

The framework automatically provides a complete Swagger UI interface accessible at `/decorators/swagger-ui` when you include the `@SwaggerUI()` decorator:

```go
// @SwaggerUI()
// @Description("Interactive API documentation and testing interface")
func SwaggerUIHandler(c *gin.Context) {
    // This endpoint automatically serves the Swagger UI
    // No implementation needed - handled by the framework
}
```

**Swagger UI Features:**

1. **Interactive Testing**: Test all endpoints directly from the browser
2. **Schema Visualization**: See complete request/response schemas with examples
3. **Array Support**: Properly displays arrays of complex objects
4. **Validation Feedback**: Shows validation constraints and requirements
5. **Authentication Testing**: Test protected endpoints with auth tokens
6. **Response Examples**: View actual response structures

**Access Points:**
- **Swagger UI**: `http://localhost:8080/decorators/swagger-ui`
- **OpenAPI JSON**: `http://localhost:8080/decorators/openapi.json`
- **OpenAPI YAML**: `http://localhost:8080/decorators/openapi.yaml`

**Schema Integration in Swagger UI:**
When you use the `type` parameter in `@Response` decorators, Swagger UI automatically:
- Links response examples to the actual schema
- Shows expandable schema definitions with all fields
- Displays validation rules and constraints
- Provides proper type information for arrays and nested objects
- Enables "Try it out" functionality with pre-filled forms
- Proper error response formats
- Type information for all fields

**Access your documentation:**
- Swagger UI: `http://localhost:8080/decorators/swagger-ui`
- OpenAPI JSON: `http://localhost:8080/decorators/openapi.json`
- HTML Docs: `http://localhost:8080/decorators/docs`

---

This completes the comprehensive usage guide for the Deco framework. Each decorator provides powerful functionality with simple annotations, enabling you to build robust, scalable APIs with minimal boilerplate code. 
---
*This documentation is automatically generated from the main USAGE.md file.*
