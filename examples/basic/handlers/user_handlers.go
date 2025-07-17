package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Route("GET", "/api/health")
// @Summary("Health check endpoint")
// @Description("Checks if the service is working correctly")
// @Tag("health")
// @Response(code="200", description="Service is working")
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service":   "gin-decorators-example",
		"status":    "ok",
		"timestamp": 1234567890,
	})
}

// @Route("GET", "/users")
// @Summary("List all users")
// @Description("Returns a paginated list of all users in the system")
// @Tag("users")
// @Response(code="200", description="List of users returned successfully")
func ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users := []gin.H{
		{"id": 1, "name": "John Silva", "email": "john@example.com"},
		{"id": 2, "name": "Mary Santos", "email": "mary@example.com"},
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(users),
		},
	})
}

// @Route("POST", "/users")
// @Summary("Create a new user")
// @Description("Creates a new user in the system with the provided data")
// @Tag("users")
// @Response(code="201", description="User created successfully")
// @Response(code="400", description="Invalid data provided")
func CreateUser(c *gin.Context) {
	var body struct {
		Name     string `json:"name" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser := gin.H{
		"id":    3,
		"name":  body.Name,
		"email": body.Email,
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    newUser,
	})
}

// @Route("GET", "/users/:id")
// @Summary("Get a specific user")
// @Description("Returns data for a specific user by ID")
// @Tag("users")
// @Response(code="200", description="User data")
// @Response(code="404", description="User not found")
func GetUser(c *gin.Context) {
	id := c.Param("id")
	userId, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	user := gin.H{
		"id":    userId,
		"name":  "John Silva",
		"email": "john@example.com",
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
