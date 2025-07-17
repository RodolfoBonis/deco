package handlers

import (
	"time"

	"github.com/gin-gonic/gin"
)

// User represents a user in the system
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

// UserPrefs represents user preferences
// @Schema()
// @Description("User preferences and settings")
type UserPrefs struct {
	Theme         string   `json:"theme" validate:"oneof=light dark auto"`   // UI theme preference
	Language      string   `json:"language" validate:"required,min=2,max=5"` // Preferred language code
	Notifications bool     `json:"notifications"`                            // Enable notifications
	Tags          []string `json:"tags,omitempty"`                           // User-defined tags
}

// CreateUserRequest represents the request body for creating a user
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

// UserResponse represents the response when returning user data
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

// ErrorResponse represents an error response
// @Schema()
// @Description("Standard error response format")
type ErrorResponse struct {
	Error   string                 `json:"error" validate:"required"` // Error message
	Code    int                    `json:"code" validate:"required"`  // HTTP status code
	Details map[string]interface{} `json:"details,omitempty"`         // Additional error details
}

// ListUsersResponse represents a paginated list of users
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

// Product represents a product in an e-commerce system
// @Schema()
// @Description("Product entity for e-commerce operations")
type Product struct {
	ID          int       `json:"id" validate:"required"`                 // Product unique identifier
	Name        string    `json:"name" validate:"required,min=1,max=200"` // Product name
	Description string    `json:"description" validate:"max=1000"`        // Product description
	Price       float64   `json:"price" validate:"required,min=0"`        // Product price
	Currency    string    `json:"currency" validate:"required,len=3"`     // Currency code (e.g., USD, EUR)
	Stock       int       `json:"stock" validate:"min=0"`                 // Available stock quantity
	Category    string    `json:"category" validate:"required"`           // Product category
	Tags        []string  `json:"tags,omitempty"`                         // Product tags for search
	IsActive    bool      `json:"isActive"`                               // Whether product is available
	CreatedAt   time.Time `json:"createdAt"`                              // Product creation timestamp
	UpdatedAt   time.Time `json:"updatedAt"`                              // Last update timestamp
}

// Example handlers that use the schemas defined above

// CreateUserWithSchema demonstrates how to use schemas in request/response
// @Route("POST", "/api/users/schema")
// @Auth(role="admin")
// @ValidateJSON()
// @Description("Create a new user using schema-defined request and response")
// @Summary("Create User (with schemas)")
// @Tag("Users")
// @Param(name="user", type="CreateUserRequest", location="body", required=true, description="User creation data")
// @Response(code=201, description="User created successfully", type="UserResponse")
// @Response(code=400, description="Invalid user data", type="ErrorResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
func CreateUserWithSchema(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrorResponse{
			Error:   "Invalid request payload",
			Code:    400,
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// Simulate user creation
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

// GetUserWithSchema demonstrates schema-based response
// @Route("GET", "/api/users/schema/:id")
// @Auth()
// @Cache(ttl="5m")
// @Description("Get user by ID with schema-defined response")
// @Summary("Get User (with schemas)")
// @Tag("Users")
// @Param(name="id", type="int", location="path", required=true, description="User ID")
// @Response(code=200, description="User found", type="UserResponse")
// @Response(code=404, description="User not found", type="ErrorResponse")
func GetUserWithSchema(c *gin.Context) {
	userID := c.Param("id")

	// Simulate user lookup
	if userID == "999" {
		c.JSON(404, ErrorResponse{
			Error: "User not found",
			Code:  404,
		})
		return
	}

	user := UserResponse{
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

// UpdateUserWithSchema demonstrates schema-based request and response
// @Route("PUT", "/api/users/schema/:id")
// @Auth(role="admin")
// @ValidateJSON()
// @Description("Update user with schema-defined request and response")
// @Summary("Update User (with schemas)")
// @Tag("Users")
// @Param(name="id", type="int", location="path", required=true, description="User ID")
// @Param(name="user", type="CreateUserRequest", location="body", required=true, description="Updated user data")
// @Response(code=200, description="User updated successfully", type="UserResponse")
// @Response(code=404, description="User not found", type="ErrorResponse")
// @Response(code=400, description="Invalid user data", type="ErrorResponse")
func UpdateUserWithSchema(c *gin.Context) {
	userID := c.Param("id")
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, ErrorResponse{
			Error:   "Invalid request payload",
			Code:    400,
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// Simulate user update
	if userID == "999" {
		c.JSON(404, ErrorResponse{
			Error: "User not found",
			Code:  404,
		})
		return
	}

	user := UserResponse{
		ID:        1,
		Name:      req.Name,
		Email:     req.Email,
		Age:       req.Age,
		IsActive:  true,
		Role:      req.Role,
		CreatedAt: time.Now().Add(-24 * time.Hour), // Created yesterday
		UpdatedAt: time.Now(),                      // Updated now
	}

	c.JSON(200, user)
}

// ListUsersWithSchema demonstrates schema-based list response
// @Route("GET", "/api/users/schema")
// @Auth()
// @Cache(ttl="2m")
// @Description("List users with schema-defined paginated response")
// @Summary("List Users (with schemas)")
// @Tag("Users")
// @Param(name="page", type="int", location="query", description="Page number")
// @Param(name="limit", type="int", location="query", description="Items per page")
// @Response(code=200, description="List of users with pagination", type="ListUsersResponse")
// @Response(code=401, description="Authentication required", type="ErrorResponse")
func ListUsersWithSchema(c *gin.Context) {
	// Simulate pagination
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

// CreateProductWithSchema demonstrates e-commerce schema usage
// @Route("POST", "/api/products")
// @Auth(role="admin")
// @ValidateJSON()
// @Description("Create a new product with schema validation")
// @Summary("Create Product")
// @Tag("Products")
// @Param(name="product", type="Product", location="body", required=true, description="Product data")
// @Response(code=201, description="Product created successfully", type="Product")
// @Response(code=400, description="Invalid product data", type="ErrorResponse")
func CreateProductWithSchema(c *gin.Context) {
	var product Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(400, ErrorResponse{
			Error:   "Invalid product data",
			Code:    400,
			Details: map[string]interface{}{"validation_error": err.Error()},
		})
		return
	}

	// Set timestamps
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.ID = 123 // Simulate database ID

	c.JSON(201, product)
}
