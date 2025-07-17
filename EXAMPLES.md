# 🚀 Examples - deco

Real-world examples and usage patterns for deco.

## 📖 Table of Contents

1. [Quick Start](#-quick-start)
2. [Basic REST API](#-basic-rest-api)
3. [Schema System & OpenAPI](#-schema-system--openapi)
4. [User Management System](#-user-management-system)
5. [E-commerce API](#-e-commerce-api)
6. [Admin Dashboard](#-admin-dashboard)
7. [File Upload Service](#-file-upload-service)
8. [Real-time Chat API](#-real-time-chat-api)
9. [Microservice Integration](#-microservice-integration)
10. [Custom Middlewares](#-custom-middlewares)
11. [Advanced Patterns](#-advanced-patterns)

## 🚀 Quick Start

### Minimal Example

```go
// main.go
package main

import (
    _ "myapp/handlers"
    deco "github.com/yourusername/deco"
)

func main() {
    r := deco.Default()
    r.Run(":8080")
}
```

```go
// handlers/health.go
package handlers

import "github.com/gin-gonic/gin"

//go:generate deco --root ./handlers --out ./init_decorators.go --pkg handlers

// @Route("GET", "/health")
func HealthCheck(c *gin.Context) {
    c.JSON(200, gin.H{"status": "healthy"})
}
```

## 🔧 Basic REST API

### Simple CRUD Operations

```go
// handlers/users.go
package handlers

import (
    "net/http"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
)

type User struct {
    ID        int       `json:"id"`
    Name      string    `json:"name" binding:"required"`
    Email     string    `json:"email" binding:"required,email"`
    CreatedAt time.Time `json:"created_at"`
}

var users = []User{
    {ID: 1, Name: "John Doe", Email: "john@example.com", CreatedAt: time.Now()},
    {ID: 2, Name: "Jane Smith", Email: "jane@example.com", CreatedAt: time.Now()},
}
var nextID = 3

// @Route("GET", "/api/users")
// @Cache(duration="5m")
// @Metrics(name="list_users")
func GetUsers(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "users": users,
        "total": len(users),
    })
}

// @Route("GET", "/api/users/:id")
// @Cache(duration="10m")
// @Metrics(name="get_user")
func GetUser(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    for _, user := range users {
        if user.ID == id {
            c.JSON(http.StatusOK, gin.H{"user": user})
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

// @Route("POST", "/api/users")
// @RateLimit(limit=10, window="1m")
// @Metrics(name="create_user")
func CreateUser(c *gin.Context) {
    var newUser User
    if err := c.ShouldBindJSON(&newUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    newUser.ID = nextID
    nextID++
    newUser.CreatedAt = time.Now()
    
    users = append(users, newUser)
    
    c.JSON(http.StatusCreated, gin.H{
        "message": "User created successfully",
        "user":    newUser,
    })
}

// @Route("PUT", "/api/users/:id")
// @RateLimit(limit=20, window="1m")
// @Metrics(name="update_user")
func UpdateUser(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    var updatedUser User
    if err := c.ShouldBindJSON(&updatedUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    for i, user := range users {
        if user.ID == id {
            users[i].Name = updatedUser.Name
            users[i].Email = updatedUser.Email
            
            c.JSON(http.StatusOK, gin.H{
                "message": "User updated successfully",
                "user":    users[i],
            })
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

// @Route("DELETE", "/api/users/:id")
// @RateLimit(limit=5, window="1m")
// @Metrics(name="delete_user")
func DeleteUser(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    for i, user := range users {
        if user.ID == id {
            users = append(users[:i], users[i+1:]...)
            c.JSON(http.StatusOK, gin.H{
                "message": "User deleted successfully",
                "id":      id,
            })
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}
```

## 📋 Schema System & OpenAPI

### Complete Schema-Driven API

This example demonstrates a fully schema-driven API with automatic OpenAPI generation, array support, and interactive Swagger UI.

```go
// schemas/entities.go
package schemas

import "time"

// @Schema()
// @Description("User entity representing a registered user in the system")
type User struct {
    ID          int       `json:"id" validate:"required"`                  // User unique identifier
    Name        string    `json:"name" validate:"required,min=2,max=100"`  // Full name of the user
    Email       string    `json:"email" validate:"required,email"`         // Email address (must be unique)
    Age         *int      `json:"age,omitempty" validate:"min=18,max=120"` // User age (optional)
    IsActive    bool      `json:"isActive"`                                // Whether the user is active
    Role        string    `json:"role" validate:"oneof=admin user guest"`  // User role in the system
    CreatedAt   time.Time `json:"createdAt"`                               // Account creation timestamp
    UpdatedAt   time.Time `json:"updatedAt"`                               // Last update timestamp
    Preferences UserPrefs `json:"preferences,omitempty"`                   // User preferences
}

// @Schema()
// @Description("User preferences and settings")
type UserPrefs struct {
    Theme         string   `json:"theme" validate:"oneof=light dark auto"`   // UI theme preference
    Language      string   `json:"language" validate:"required,min=2,max=5"` // Preferred language code
    Notifications bool     `json:"notifications"`                            // Enable notifications
    Tags          []string `json:"tags,omitempty"`                           // User-defined tags
}

// @Schema()
// @Description("Request payload for creating a new user account")
type CreateUserRequest struct {
    Name     string    `json:"name" validate:"required,min=2,max=100"`  // Full name of the user
    Email    string    `json:"email" validate:"required,email"`         // Email address (must be unique)
    Password string    `json:"password" validate:"required,min=8"`      // Password (minimum 8 characters)
    Age      *int      `json:"age,omitempty" validate:"min=18,max=120"` // User age (optional)
    Role     string    `json:"role" validate:"oneof=admin user guest"`  // User role (defaults to 'user')
    Prefs    UserPrefs `json:"preferences,omitempty"`                   // Initial user preferences
}

// @Schema()
// @Description("Response containing user information (without sensitive data)")
type UserResponse struct {
    ID        int       `json:"id"`            // User unique identifier
    Name      string    `json:"name"`          // Full name of the user
    Email     string    `json:"email"`         // Email address
    Age       *int      `json:"age,omitempty"` // User age
    IsActive  bool      `json:"isActive"`      // Whether the user is active
    Role      string    `json:"role"`          // User role in the system
    CreatedAt time.Time `json:"createdAt"`     // Account creation timestamp
    UpdatedAt time.Time `json:"updatedAt"`     // Last update timestamp
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

// @Schema()
// @Description("Standard error response format")
type ErrorResponse struct {
    Error   string                 `json:"error" validate:"required"` // Error message
    Code    int                    `json:"code" validate:"required"`  // HTTP status code
    Details map[string]interface{} `json:"details,omitempty"`         // Additional error details
}
```

```go
// handlers/user_api.go
package handlers

import (
    "time"
    "github.com/gin-gonic/gin"
    "myapp/schemas"
)

// @Route("POST", "/api/users")
// @Auth(role="admin")
// @ValidateJSON()
// @RateLimit(limit=10, window="1m")
// @Description("Create a new user account with full validation")
// @Summary("Create User")
// @Tag("Users")
// @Param(name="user", type="CreateUserRequest", location="body", required=true, description="User creation data")
// @Response(code=201, description="User created successfully", type="UserResponse")
// @Response(code=400, description="Invalid user data", type="ErrorResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
// @Response(code=403, description="Admin role required", type="ErrorResponse")
func CreateUser(c *gin.Context) {
    var req schemas.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, schemas.ErrorResponse{
            Error:   "Invalid request payload",
            Code:    400,
            Details: map[string]interface{}{"validation_error": err.Error()},
        })
        return
    }
    
    // Create user logic here...
    user := schemas.UserResponse{
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
// @Description("List all users with pagination support")
// @Summary("List Users")
// @Tag("Users")
// @Param(name="page", type="int", location="query", description="Page number", example="1")
// @Param(name="limit", type="int", location="query", description="Items per page", example="10")
// @Param(name="role", type="string", location="query", description="Filter by role", example="admin")
// @Response(code=200, description="Users retrieved successfully", type="ListUsersResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
func ListUsers(c *gin.Context) {
    users := []schemas.UserResponse{
        {ID: 1, Name: "John Doe", Email: "john@example.com", IsActive: true, Role: "user"},
        {ID: 2, Name: "Jane Smith", Email: "jane@example.com", IsActive: true, Role: "admin"},
        {ID: 3, Name: "Bob Wilson", Email: "bob@example.com", IsActive: false, Role: "guest"},
    }
    
    response := schemas.ListUsersResponse{
        Users:   users,
        Total:   len(users),
        Page:    1,
        Limit:   10,
        HasNext: false,
        HasPrev: false,
    }
    
    c.JSON(200, response)
}

// @Route("GET", "/api/users/:id")
// @Auth()
// @Cache(ttl="10m")
// @Description("Get user details by ID")
// @Summary("Get User")
// @Tag("Users")
// @Param(name="id", type="int", location="path", required=true, description="User ID")
// @Response(code=200, description="User retrieved successfully", type="UserResponse")
// @Response(code=404, description="User not found", type="ErrorResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
func GetUser(c *gin.Context) {
    // Get user logic here...
    user := schemas.UserResponse{
        ID:        1,
        Name:      "John Doe",
        Email:     "john@example.com",
        IsActive:  true,
        Role:      "user",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    c.JSON(200, user)
}

// @Route("GET", "/swagger")
// @SwaggerUI()
// @Description("Interactive API documentation and testing interface")
func SwaggerUI(c *gin.Context) {
    // Framework automatically serves Swagger UI
}

// @Route("GET", "/api-spec")
// @OpenAPIJSON()
// @Description("OpenAPI 3.0 specification in JSON format")
func OpenAPISpec(c *gin.Context) {
    // Framework automatically serves OpenAPI spec
}
```

### Key Features Demonstrated

1. **Schema Definition**: Entities defined with `@Schema()` decorator and rich field documentation
2. **Array Support**: `ListUsersResponse` contains `[]UserResponse` with automatic reference resolution
3. **Validation Integration**: `validate` tags automatically appear in OpenAPI constraints
4. **Response Linking**: `@Response(type="SchemaName")` creates proper OpenAPI references
5. **Interactive Testing**: Swagger UI provides full testing capabilities with schema-aware forms
6. **Nested Schemas**: `UserPrefs` nested within `User` with proper type resolution

### Generated Documentation Features

- **Complete OpenAPI 3.0 Specification**: All schemas, endpoints, and relationships
- **Interactive Swagger UI**: Test endpoints with proper request/response schemas
- **Array Visualization**: Swagger UI shows expandable arrays with item schema details
- **Validation Constraints**: Min/max values, required fields, enum options
- **Type Safety**: Proper Go type to OpenAPI type mapping

### Access Points

- **Swagger UI**: `http://localhost:8080/swagger`
- **OpenAPI JSON**: `http://localhost:8080/api-spec`
- **Framework Stats**: `http://localhost:8080/decorators/docs`

## 👤 User Management System

### Authentication & Authorization

```go
// handlers/auth.go
package handlers

import (
    "crypto/sha256"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v4"
)

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
    Name     string `json:"name" binding:"required,min=2"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
    Role     string `json:"role,omitempty"`
}

type AuthResponse struct {
    Token string `json:"token"`
    User  User   `json:"user"`
}

var jwtSecret = []byte("your-secret-key")

// @Route("POST", "/api/auth/register")
// @CORS(origins="*")
// @RateLimit(limit=5, window="1m")
// @Metrics(name="user_registration")
func Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Check if user already exists
    for _, user := range users {
        if user.Email == req.Email {
            c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
            return
        }
    }
    
    // Hash password (simplified - use bcrypt in production)
    hasher := sha256.New()
    hasher.Write([]byte(req.Password))
    hashedPassword := fmt.Sprintf("%x", hasher.Sum(nil))
    
    // Create user
    user := User{
        ID:        nextID,
        Name:      req.Name,
        Email:     req.Email,
        CreatedAt: time.Now(),
    }
    nextID++
    
    users = append(users, user)
    
    // Generate JWT token
    token, err := generateJWT(user)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
        return
    }
    
    c.JSON(http.StatusCreated, AuthResponse{
        Token: token,
        User:  user,
    })
}

// @Route("POST", "/api/auth/login")
// @CORS(origins="*")
// @RateLimit(limit=10, window="1m")
// @Metrics(name="user_login")
func Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // Find user by email (simplified authentication)
    for _, user := range users {
        if user.Email == req.Email {
            // In production, compare hashed passwords
            token, err := generateJWT(user)
            if err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
                return
            }
            
            c.JSON(http.StatusOK, AuthResponse{
                Token: token,
                User:  user,
            })
            return
        }
    }
    
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}

// @Route("GET", "/api/auth/profile")
// @Auth(role="user")
// @Cache(duration="5m")
// @Metrics(name="profile_access")
func GetProfile(c *gin.Context) {
    // User information should be set by auth middleware
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    
    id := userID.(int)
    for _, user := range users {
        if user.ID == id {
            c.JSON(http.StatusOK, gin.H{"profile": user})
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

// @Route("PUT", "/api/auth/profile")
// @Auth(role="user")
// @RateLimit(limit=5, window="1m")
// @Metrics(name="profile_update")
func UpdateProfile(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    
    var updateReq struct {
        Name string `json:"name" binding:"required,min=2"`
    }
    
    if err := c.ShouldBindJSON(&updateReq); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    id := userID.(int)
    for i, user := range users {
        if user.ID == id {
            users[i].Name = updateReq.Name
            c.JSON(http.StatusOK, gin.H{
                "message": "Profile updated successfully",
                "user":    users[i],
            })
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func generateJWT(user User) (string, error) {
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "email":   user.Email,
        "exp":     time.Now().Add(time.Hour * 24).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}
```

## 🛒 E-commerce API

### Product Management

```go
// handlers/products.go
package handlers

import (
    "net/http"
    "strconv"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
)

type Product struct {
    ID          int       `json:"id"`
    Name        string    `json:"name" binding:"required"`
    Description string    `json:"description"`
    Price       float64   `json:"price" binding:"required,gt=0"`
    Category    string    `json:"category" binding:"required"`
    Stock       int       `json:"stock" binding:"min=0"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

var products = []Product{
    {ID: 1, Name: "Laptop", Description: "High-performance laptop", Price: 999.99, Category: "electronics", Stock: 10, CreatedAt: time.Now()},
    {ID: 2, Name: "Coffee Mug", Description: "Ceramic coffee mug", Price: 12.99, Category: "kitchen", Stock: 50, CreatedAt: time.Now()},
}
var nextProductID = 3

// @Route("GET", "/api/products")
// @CORS(origins="*")
// @Cache(duration="10m")
// @RateLimit(limit=100, window="1m")
// @Metrics(name="product_listing")
func ListProducts(c *gin.Context) {
    // Pagination
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    category := c.Query("category")
    search := c.Query("search")
    
    filtered := make([]Product, 0)
    
    for _, product := range products {
        // Category filter
        if category != "" && product.Category != category {
            continue
        }
        
        // Search filter
        if search != "" && !strings.Contains(strings.ToLower(product.Name), strings.ToLower(search)) {
            continue
        }
        
        filtered = append(filtered, product)
    }
    
    // Pagination
    start := (page - 1) * limit
    end := start + limit
    
    if start >= len(filtered) {
        start = len(filtered)
    }
    if end > len(filtered) {
        end = len(filtered)
    }
    
    result := filtered[start:end]
    
    c.JSON(http.StatusOK, gin.H{
        "products": result,
        "pagination": gin.H{
            "page":       page,
            "limit":      limit,
            "total":      len(filtered),
            "total_pages": (len(filtered) + limit - 1) / limit,
        },
    })
}

// @Route("GET", "/api/products/:id")
// @CORS(origins="*")
// @Cache(duration="15m")
// @Metrics(name="product_detail")
func GetProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }
    
    for _, product := range products {
        if product.ID == id {
            c.JSON(http.StatusOK, gin.H{"product": product})
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
}

// @Route("POST", "/api/products")
// @Auth(role="admin")
// @RateLimit(limit=10, window="1m")
// @Metrics(name="product_creation")
func CreateProduct(c *gin.Context) {
    var newProduct Product
    if err := c.ShouldBindJSON(&newProduct); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    newProduct.ID = nextProductID
    nextProductID++
    newProduct.CreatedAt = time.Now()
    newProduct.UpdatedAt = time.Now()
    
    products = append(products, newProduct)
    
    c.JSON(http.StatusCreated, gin.H{
        "message": "Product created successfully",
        "product": newProduct,
    })
}

// @Route("PUT", "/api/products/:id")
// @Auth(role="admin")
// @RateLimit(limit=20, window="1m")
// @Metrics(name="product_update")
func UpdateProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }
    
    var updates Product
    if err := c.ShouldBindJSON(&updates); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    for i, product := range products {
        if product.ID == id {
            products[i].Name = updates.Name
            products[i].Description = updates.Description
            products[i].Price = updates.Price
            products[i].Category = updates.Category
            products[i].Stock = updates.Stock
            products[i].UpdatedAt = time.Now()
            
            c.JSON(http.StatusOK, gin.H{
                "message": "Product updated successfully",
                "product": products[i],
            })
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
}

// @Route("DELETE", "/api/products/:id")
// @Auth(role="admin")
// @RateLimit(limit=5, window="1m")
// @Metrics(name="product_deletion")
func DeleteProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
        return
    }
    
    for i, product := range products {
        if product.ID == id {
            products = append(products[:i], products[i+1:]...)
            c.JSON(http.StatusOK, gin.H{
                "message": "Product deleted successfully",
                "id":      id,
            })
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
}

// @Route("GET", "/api/categories")
// @CORS(origins="*")
// @Cache(duration="1h")
// @Metrics(name="category_list")
func GetCategories(c *gin.Context) {
    categories := make(map[string]int)
    
    for _, product := range products {
        categories[product.Category]++
    }
    
    result := make([]gin.H, 0)
    for category, count := range categories {
        result = append(result, gin.H{
            "name":  category,
            "count": count,
        })
    }
    
    c.JSON(http.StatusOK, gin.H{"categories": result})
}
```

## 🎛️ Admin Dashboard

### Administration Endpoints

```go
// handlers/admin.go
package handlers

import (
    "net/http"
    "runtime"
    "time"
    
    "github.com/gin-gonic/gin"
)

type DashboardStats struct {
    TotalUsers    int     `json:"total_users"`
    TotalProducts int     `json:"total_products"`
    TotalOrders   int     `json:"total_orders"`
    Revenue       float64 `json:"revenue"`
    ActiveUsers   int     `json:"active_users"`
    ServerUptime  string  `json:"server_uptime"`
    MemoryUsage   string  `json:"memory_usage"`
}

var serverStartTime = time.Now()

// @Route("GET", "/api/admin/dashboard")
// @Auth(role="admin")
// @Cache(duration="1m")
// @Metrics(name="admin_dashboard")
func GetDashboard(c *gin.Context) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    stats := DashboardStats{
        TotalUsers:    len(users),
        TotalProducts: len(products),
        TotalOrders:   0, // Would come from orders database
        Revenue:       0, // Would be calculated from orders
        ActiveUsers:   len(users), // Simplified
        ServerUptime:  time.Since(serverStartTime).String(),
        MemoryUsage:   formatBytes(memStats.Alloc),
    }
    
    c.JSON(http.StatusOK, gin.H{"dashboard": stats})
}

// @Route("GET", "/api/admin/users")
// @Auth(role="admin")
// @Cache(duration="2m")
// @RateLimit(limit=50, window="1m")
// @Metrics(name="admin_user_list")
func AdminListUsers(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    search := c.Query("search")
    
    filtered := make([]User, 0)
    
    for _, user := range users {
        if search != "" && !strings.Contains(strings.ToLower(user.Name), strings.ToLower(search)) &&
           !strings.Contains(strings.ToLower(user.Email), strings.ToLower(search)) {
            continue
        }
        filtered = append(filtered, user)
    }
    
    // Pagination
    start := (page - 1) * limit
    end := start + limit
    
    if start >= len(filtered) {
        start = len(filtered)
    }
    if end > len(filtered) {
        end = len(filtered)
    }
    
    result := filtered[start:end]
    
    c.JSON(http.StatusOK, gin.H{
        "users": result,
        "pagination": gin.H{
            "page":       page,
            "limit":      limit,
            "total":      len(filtered),
            "total_pages": (len(filtered) + limit - 1) / limit,
        },
    })
}

// @Route("POST", "/api/admin/users/:id/suspend")
// @Auth(role="admin")
// @RateLimit(limit=10, window="1m")
// @Metrics(name="admin_user_suspend")
func SuspendUser(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    
    for i, user := range users {
        if user.ID == id {
            // In a real application, you'd set a suspended flag
            c.JSON(http.StatusOK, gin.H{
                "message": "User suspended successfully",
                "user":    users[i],
            })
            return
        }
    }
    
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

// @Route("GET", "/api/admin/logs")
// @Auth(role="admin")
// @RateLimit(limit=20, window="1m")
// @Metrics(name="admin_logs")
func GetSystemLogs(c *gin.Context) {
    // Mock log entries
    logs := []gin.H{
        {
            "timestamp": time.Now().Add(-time.Hour),
            "level":     "INFO",
            "message":   "Server started successfully",
            "source":    "main.go",
        },
        {
            "timestamp": time.Now().Add(-30 * time.Minute),
            "level":     "WARN",
            "message":   "High memory usage detected",
            "source":    "monitor.go",
        },
        {
            "timestamp": time.Now().Add(-10 * time.Minute),
            "level":     "ERROR",
            "message":   "Failed to connect to external API",
            "source":    "api_client.go",
        },
    }
    
    c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// @Route("GET", "/api/admin/system/health")
// @Auth(role="admin")
// @Metrics(name="admin_system_health")
func SystemHealth(c *gin.Context) {
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    
    health := gin.H{
        "status":       "healthy",
        "timestamp":    time.Now(),
        "uptime":       time.Since(serverStartTime).String(),
        "memory": gin.H{
            "allocated":     formatBytes(memStats.Alloc),
            "total_alloc":   formatBytes(memStats.TotalAlloc),
            "sys":           formatBytes(memStats.Sys),
            "num_gc":        memStats.NumGC,
        },
        "runtime": gin.H{
            "goroutines": runtime.NumGoroutine(),
            "cpu_count":  runtime.NumCPU(),
            "go_version": runtime.Version(),
        },
    }
    
    c.JSON(http.StatusOK, gin.H{"health": health})
}

func formatBytes(bytes uint64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := uint64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

## 📁 File Upload Service

### File Handling with Custom Middleware

```go
// handlers/files.go
package handlers

import (
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
)

// @Route("POST", "/api/upload")
// @Auth(role="user")
// @RateLimit(limit=5, window="1m")
// @Metrics(name="file_upload")
func UploadFile(c *gin.Context) {
    file, header, err := c.Request.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
        return
    }
    defer file.Close()
    
    // Validate file type
    allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"}
    ext := strings.ToLower(filepath.Ext(header.Filename))
    
    allowed := false
    for _, allowedType := range allowedTypes {
        if ext == allowedType {
            allowed = true
            break
        }
    }
    
    if !allowed {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "File type not allowed",
            "allowed_types": allowedTypes,
        })
        return
    }
    
    // Validate file size (5MB limit)
    if header.Size > 5*1024*1024 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 5MB)"})
        return
    }
    
    // Create uploads directory if it doesn't exist
    uploadDir := "uploads"
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        os.MkdirAll(uploadDir, 0755)
    }
    
    // Generate unique filename
    filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
    filepath := filepath.Join(uploadDir, filename)
    
    // Save file
    out, err := os.Create(filepath)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }
    defer out.Close()
    
    _, err = io.Copy(out, file)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message":  "File uploaded successfully",
        "filename": filename,
        "size":     header.Size,
        "url":      fmt.Sprintf("/api/files/%s", filename),
    })
}

// @Route("GET", "/api/files/:filename")
// @Cache(duration="1h")
// @Metrics(name="file_download")
func DownloadFile(c *gin.Context) {
    filename := c.Param("filename")
    filepath := filepath.Join("uploads", filename)
    
    // Check if file exists
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
        return
    }
    
    c.File(filepath)
}

// @Route("DELETE", "/api/files/:filename")
// @Auth(role="user")
// @RateLimit(limit=10, window="1m")
// @Metrics(name="file_delete")
func DeleteFile(c *gin.Context) {
    filename := c.Param("filename")
    filepath := filepath.Join("uploads", filename)
    
    // Check if file exists
    if _, err := os.Stat(filepath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
        return
    }
    
    // Delete file
    err := os.Remove(filepath)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message":  "File deleted successfully",
        "filename": filename,
    })
}

// @Route("GET", "/api/files")
// @Auth(role="user")
// @Cache(duration="5m")
// @Metrics(name="file_list")
func ListFiles(c *gin.Context) {
    uploadDir := "uploads"
    
    files, err := os.ReadDir(uploadDir)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
        return
    }
    
    var fileList []gin.H
    for _, file := range files {
        if !file.IsDir() {
            info, err := file.Info()
            if err != nil {
                continue
            }
            
            fileList = append(fileList, gin.H{
                "name":      file.Name(),
                "size":      info.Size(),
                "modified":  info.ModTime(),
                "url":       fmt.Sprintf("/api/files/%s", file.Name()),
            })
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "files": fileList,
        "total": len(fileList),
    })
}
```

## 💬 Real-time Chat API

### WebSocket Integration

```go
// handlers/chat.go
package handlers

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins in development
    },
}

type ChatMessage struct {
    ID        string    `json:"id"`
    UserID    int       `json:"user_id"`
    Username  string    `json:"username"`
    Message   string    `json:"message"`
    Timestamp time.Time `json:"timestamp"`
    Type      string    `json:"type"` // "message", "join", "leave"
}

type ChatRoom struct {
    clients   map[*websocket.Conn]*Client
    broadcast chan ChatMessage
    register  chan *Client
    unregister chan *Client
}

type Client struct {
    conn   *websocket.Conn
    userID int
    username string
    room   *ChatRoom
}

var chatRoom = &ChatRoom{
    clients:    make(map[*websocket.Conn]*Client),
    broadcast:  make(chan ChatMessage),
    register:   make(chan *Client),
    unregister: make(chan *Client),
}

// Start the chat room hub
func init() {
    go chatRoom.run()
}

// @Route("GET", "/api/chat/ws")
// @Auth(role="user")
func WebSocketChat(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
        return
    }
    
    // Get user info from context (set by auth middleware)
    userID, _ := c.Get("user_id")
    username, _ := c.Get("username")
    
    client := &Client{
        conn:     conn,
        userID:   userID.(int),
        username: username.(string),
        room:     chatRoom,
    }
    
    chatRoom.register <- client
    
    // Start goroutines for reading and writing
    go client.writePump()
    go client.readPump()
}

// @Route("GET", "/api/chat/messages")
// @Auth(role="user")
// @Cache(duration="1m")
// @Metrics(name="chat_history")
func GetChatHistory(c *gin.Context) {
    // In a real application, you'd fetch from database
    // This is a simplified example
    messages := []ChatMessage{
        {
            ID:        "1",
            UserID:    1,
            Username:  "john_doe",
            Message:   "Hello everyone!",
            Timestamp: time.Now().Add(-time.Hour),
            Type:      "message",
        },
        {
            ID:        "2",
            UserID:    2,
            Username:  "jane_smith",
            Message:   "Hi John! How are you?",
            Timestamp: time.Now().Add(-30 * time.Minute),
            Type:      "message",
        },
    }
    
    c.JSON(http.StatusOK, gin.H{
        "messages": messages,
        "total":    len(messages),
    })
}

// @Route("POST", "/api/chat/messages")
// @Auth(role="user")
// @RateLimit(limit=30, window="1m")
// @Metrics(name="chat_send_message")
func SendMessage(c *gin.Context) {
    var req struct {
        Message string `json:"message" binding:"required,max=500"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    userID, _ := c.Get("user_id")
    username, _ := c.Get("username")
    
    message := ChatMessage{
        ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
        UserID:    userID.(int),
        Username:  username.(string),
        Message:   req.Message,
        Timestamp: time.Now(),
        Type:      "message",
    }
    
    // Broadcast to all connected clients
    chatRoom.broadcast <- message
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Message sent successfully",
        "data":    message,
    })
}

func (room *ChatRoom) run() {
    for {
        select {
        case client := <-room.register:
            room.clients[client.conn] = client
            
            // Send join notification
            joinMessage := ChatMessage{
                ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
                UserID:    client.userID,
                Username:  client.username,
                Message:   fmt.Sprintf("%s joined the chat", client.username),
                Timestamp: time.Now(),
                Type:      "join",
            }
            room.broadcast <- joinMessage
            
        case client := <-room.unregister:
            if _, ok := room.clients[client.conn]; ok {
                delete(room.clients, client.conn)
                client.conn.Close()
                
                // Send leave notification
                leaveMessage := ChatMessage{
                    ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
                    UserID:    client.userID,
                    Username:  client.username,
                    Message:   fmt.Sprintf("%s left the chat", client.username),
                    Timestamp: time.Now(),
                    Type:      "leave",
                }
                room.broadcast <- leaveMessage
            }
            
        case message := <-room.broadcast:
            for conn, client := range room.clients {
                err := conn.WriteJSON(message)
                if err != nil {
                    delete(room.clients, conn)
                    conn.Close()
                }
            }
        }
    }
}

func (c *Client) readPump() {
    defer func() {
        c.room.unregister <- c
        c.conn.Close()
    }()
    
    for {
        var message ChatMessage
        err := c.conn.ReadJSON(&message)
        if err != nil {
            break
        }
        
        // Set user info
        message.UserID = c.userID
        message.Username = c.username
        message.Timestamp = time.Now()
        message.ID = fmt.Sprintf("%d", time.Now().UnixNano())
        
        c.room.broadcast <- message
    }
}

func (c *Client) writePump() {
    defer c.conn.Close()
    
    for {
        select {
        case <-time.After(60 * time.Second):
            c.conn.WriteMessage(websocket.PingMessage, nil)
        }
    }
}
```

## 🔧 Custom Middlewares

### Creating and Using Custom Annotations

```go
// middleware/custom.go
package middleware

import (
    "context"
    "regexp"
    "time"
    
    deco "github.com/yourusername/deco"
    "github.com/gin-gonic/gin"
)

func init() {
    // Register custom @Timeout annotation
    deco.RegisterMarker(deco.MarkerConfig{
        Name:        "Timeout",
        Pattern:     regexp.MustCompile(`@Timeout\(duration="([^"]+)"\)`),
        Factory:     createTimeoutMiddleware,
        Description: "Request timeout middleware",
    })
    
    // Register custom @Log annotation
    deco.RegisterMarker(deco.MarkerConfig{
        Name:        "Log",
        Pattern:     regexp.MustCompile(`@Log\(level="([^"]+)"\)`),
        Factory:     createLogMiddleware,
        Description: "Custom logging middleware",
    })
    
    // Register custom @Retry annotation
    deco.RegisterMarker(deco.MarkerConfig{
        Name:        "Retry",
        Pattern:     regexp.MustCompile(`@Retry\(attempts=(\d+)\)`),
        Factory:     createRetryMiddleware,
        Description: "Request retry middleware",
    })
}

func createTimeoutMiddleware(args []string) gin.HandlerFunc {
    duration := "30s" // default
    if len(args) > 0 {
        duration = args[0]
    }
    
    timeout, _ := time.ParseDuration(duration)
    
    return gin.HandlerFunc(func(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
        defer cancel()
        
        c.Request = c.Request.WithContext(ctx)
        
        done := make(chan struct{})
        go func() {
            c.Next()
            done <- struct{}{}
        }()
        
        select {
        case <-done:
            // Request completed successfully
        case <-ctx.Done():
            // Request timed out
            c.JSON(408, gin.H{
                "error":   "Request timeout",
                "timeout": duration,
            })
            c.Abort()
        }
    })
}

func createLogMiddleware(args []string) gin.HandlerFunc {
    level := "info"
    if len(args) > 0 {
        level = args[0]
    }
    
    return gin.HandlerFunc(func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        latency := time.Since(start)
        status := c.Writer.Status()
        
        if level == "debug" || (level == "info" && status >= 400) {
            log.Printf("[%s] %s %s %d %v - %s",
                level,
                c.Request.Method,
                c.Request.URL.Path,
                status,
                latency,
                c.ClientIP(),
            )
        }
    })
}

func createRetryMiddleware(args []string) gin.HandlerFunc {
    attempts := 3 // default
    if len(args) > 0 {
        if parsed, err := strconv.Atoi(args[0]); err == nil {
            attempts = parsed
        }
    }
    
    return gin.HandlerFunc(func(c *gin.Context) {
        for i := 0; i < attempts; i++ {
            // Create a copy of the response writer to capture the response
            recorder := &responseRecorder{
                ResponseWriter: c.Writer,
                status:         200,
            }
            c.Writer = recorder
            
            c.Next()
            
            // If successful (status < 500), break the retry loop
            if recorder.status < 500 {
                break
            }
            
            // If this is the last attempt, let the error response through
            if i == attempts-1 {
                break
            }
            
            // Reset for retry
            c.Writer = recorder.ResponseWriter
            
            // Wait before retry (exponential backoff)
            time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
        }
    })
}

type responseRecorder struct {
    gin.ResponseWriter
    status int
}

func (r *responseRecorder) WriteHeader(status int) {
    r.status = status
    r.ResponseWriter.WriteHeader(status)
}
```

### Using Custom Middlewares

```go
// handlers/custom_example.go
package handlers

import _ "myapp/middleware" // Import to register custom middlewares

// @Route("POST", "/api/heavy-operation")
// @Auth(role="user")
// @Timeout(duration="60s")
// @Log(level="debug")
// @Retry(attempts=3)
// @Metrics(name="heavy_operation")
func HeavyOperation(c *gin.Context) {
    // Simulate heavy processing
    time.Sleep(5 * time.Second)
    
    c.JSON(200, gin.H{
        "message": "Operation completed successfully",
        "duration": "5s",
    })
}

// @Route("GET", "/api/external-data")
// @Auth(role="user")
// @Timeout(duration="10s")
// @Retry(attempts=5)
// @Cache(duration="5m")
func FetchExternalData(c *gin.Context) {
    // Simulate external API call that might fail
    if time.Now().Second()%3 == 0 {
        c.JSON(500, gin.H{"error": "External service unavailable"})
        return
    }
    
    c.JSON(200, gin.H{
        "data": "External data retrieved successfully",
        "timestamp": time.Now(),
    })
}
```

## 🚀 Advanced Patterns

### Microservice Integration

```go
// handlers/microservice.go
package handlers

import (
    "bytes"
    "encoding/json"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
)

// @Route("GET", "/api/user-service/users/:id")
// @Auth(role="user")
// @Cache(duration="5m")
// @Timeout(duration="10s")
// @Metrics(name="user_service_proxy")
func ProxyToUserService(c *gin.Context) {
    userID := c.Param("id")
    
    // Forward request to user microservice
    resp, err := http.Get(fmt.Sprintf("http://user-service:8081/users/%s", userID))
    if err != nil {
        c.JSON(500, gin.H{"error": "User service unavailable"})
        return
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    c.JSON(resp.StatusCode, result)
}

// @Route("POST", "/api/notification-service/send")
// @Auth(role="admin")
// @RateLimit(limit=10, window="1m")
// @Retry(attempts=3)
// @Metrics(name="notification_service")
func SendNotification(c *gin.Context) {
    var notification struct {
        Recipients []string `json:"recipients" binding:"required"`
        Subject    string   `json:"subject" binding:"required"`
        Message    string   `json:"message" binding:"required"`
        Type       string   `json:"type" binding:"required,oneof=email sms push"`
    }
    
    if err := c.ShouldBindJSON(&notification); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    // Forward to notification service
    jsonData, _ := json.Marshal(notification)
    resp, err := http.Post(
        "http://notification-service:8082/send",
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    
    if err != nil {
        c.JSON(500, gin.H{"error": "Notification service unavailable"})
        return
    }
    defer resp.Body.Close()
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    c.JSON(resp.StatusCode, result)
}
```

### API Versioning

```go
// handlers/v1/users.go
package v1

import "github.com/gin-gonic/gin"

// @Route("GET", "/api/v1/users")
// @Cache(duration="5m")
func GetUsersV1(c *gin.Context) {
    // Version 1 implementation
    c.JSON(200, gin.H{
        "version": "1.0",
        "users":   []string{"user1", "user2"},
    })
}

// handlers/v2/users.go
package v2

import "github.com/gin-gonic/gin"

// @Route("GET", "/api/v2/users")
// @Cache(duration="10m")
func GetUsersV2(c *gin.Context) {
    // Version 2 implementation with enhanced features
    c.JSON(200, gin.H{
        "version": "2.0",
        "users": []gin.H{
            {"id": 1, "name": "User 1", "created_at": "2023-01-01"},
            {"id": 2, "name": "User 2", "created_at": "2023-01-02"},
        },
        "pagination": gin.H{
            "page":  1,
            "limit": 20,
            "total": 2,
        },
    })
}
```

This comprehensive examples guide demonstrates the power and flexibility of deco for building modern web APIs with minimal boilerplate code. Each example showcases different aspects of the framework while maintaining clean, readable, and maintainable code.

---

**🎯 deco** - Transform your development experience with annotation-driven APIs! 