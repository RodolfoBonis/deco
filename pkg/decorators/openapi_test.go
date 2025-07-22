// Tests for OpenAPI specification logic in gin-decorators framework
package decorators

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGenerateOpenAPISpec(t *testing.T) {
	// Remove  to avoid race conditions

	// Register test route
	RegisterRoute("GET", "/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "users"})
	})

	// Create test config
	config := &Config{
		Version: "3.0.0", // Use the actual default version
		OpenAPI: OpenAPIConfig{
			Version:     "3.0.0",
			Title:       "Test API",
			Description: "Test API Description",
			Host:        "localhost:8080",
			BasePath:    "/api",
			Schemes:     []string{"http", "https"},
		},
	}

	// Generate OpenAPI spec
	spec := GenerateOpenAPISpec(config)

	// Check basic structure
	assert.Equal(t, "3.0.0", spec.OpenAPI)
	assert.Equal(t, "Test API", spec.Info.Title)
	assert.Equal(t, "3.0.0", spec.Info.Version)

	// Check that routes are included
	assert.NotEmpty(t, spec.Paths)
	assert.Contains(t, spec.Paths, "/users")
}

func TestConvertRouteToOperation(t *testing.T) {
	route := &RouteEntry{
		Method:      "POST",
		Path:        "/users",
		Handler:     func(_ *gin.Context) {},
		Tags:        []string{"users"},
		Summary:     "Create user",
		Description: "Create a new user",
		Parameters: []ParameterInfo{
			{Name: "name", Type: "string", Required: true, Description: "User name"},
			{Name: "email", Type: "string", Required: true, Description: "User email"},
		},
		Responses: []ResponseInfo{
			{Code: "201", Description: "User created", Type: "object"},
			{Code: "400", Description: "Bad request", Type: "object"},
		},
	}

	components := &OpenAPIComponents{}
	operation := convertRouteToOperation(route, components)

	assert.NotNil(t, operation)
	assert.Contains(t, operation.Tags, "users")
	assert.Equal(t, "Create user", operation.Summary)
	assert.Equal(t, "Create a new user", operation.Description)
	assert.Equal(t, "postUsers", operation.OperationID)

	// Check parameters
	assert.Len(t, operation.Parameters, 2)

	// Check responses
	assert.Len(t, operation.Responses, 2)
	assert.Contains(t, operation.Responses, "201")
	assert.Contains(t, operation.Responses, "400")
}

func TestConvertTypeToSchema(t *testing.T) {
	// Remove  to avoid race conditions

	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"int", "integer"},
		{"int64", "integer"},
		{"float64", "number"},
		{"bool", "boolean"},
		{"[]string", "array"},
		{"map[string]interface{}", "object"},
		{"User", "object"},
		{"*User", "object"},
		{"interface{}", "object"},
		{"time.Time", "string"},
		{"uuid.UUID", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertTypeToSchema(tt.input)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expected, result.Type)
		})
	}
}

func TestConvertTypeToSchema_ArrayTypes(t *testing.T) {
	// Remove  to avoid race conditions

	result := convertTypeToSchema("[]User")
	assert.NotNil(t, result)
	assert.Equal(t, "array", result.Type)
	assert.NotNil(t, result.Items)
}

func TestConvertTypeToSchema_MapTypes(t *testing.T) {
	// Remove  to avoid race conditions

	result := convertTypeToSchema("map[string]interface{}")
	assert.NotNil(t, result)
	assert.Equal(t, "object", result.Type)
}

func TestGenerateOperationID(t *testing.T) {
	// Remove  to avoid race conditions

	tests := []struct {
		method   string
		path     string
		expected string
	}{
		{"GET", "/users", "getUsers"},
		{"POST", "/users", "postUsers"},
		{"PUT", "/users/{id}", "putUsersid"},
		{"DELETE", "/users/{id}", "deleteUsersid"},
		{"GET", "/api/v1/products", "getApiv1products"},
		{"PATCH", "/users/{id}/profile", "patchUsersidprofile"},
	}

	for _, tt := range tests {
		t.Run(tt.method+"_"+tt.path, func(t *testing.T) {
			route := &RouteEntry{
				Method: tt.method,
				Path:   tt.path,
			}
			result := generateOperationID(route)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertToOpenAPIParameter(t *testing.T) {
	// Remove  to avoid race conditions

	param := &ParameterInfo{
		Name:        "id",
		Type:        "int",
		Location:    "path",
		Required:    true,
		Description: "User ID",
	}
	components := &OpenAPIComponents{}
	openAPIParam := convertToOpenAPIParameter(param, components)
	assert.Equal(t, "id", openAPIParam.Name)
	assert.Equal(t, "path", openAPIParam.In)
	assert.True(t, openAPIParam.Required)
	assert.Equal(t, "User ID", openAPIParam.Description)
}

func TestCreateRequestBodyFromParameters(t *testing.T) {
	// Remove  to avoid race conditions

	params := []ParameterInfo{
		{Name: "name", Type: "string", Location: "body", Required: true},
		{Name: "email", Type: "string", Location: "body", Required: true},
		{Name: "age", Type: "int", Location: "body", Required: false},
	}

	components := &OpenAPIComponents{}
	body := createRequestBodyFromParameters(params, components)
	assert.NotNil(t, body)
	assert.True(t, body.Required)
	assert.NotNil(t, body.Content)
}

func TestCreateResponseWithSchemaAndType(t *testing.T) {
	// Remove  to avoid race conditions

	responseInfo := ResponseInfo{
		Code:        "200",
		Description: "Success",
		Type:        "User",
	}
	components := &OpenAPIComponents{}
	response := createResponseWithSchemaAndType(responseInfo, components)
	assert.Equal(t, "Success", response.Description)
	assert.NotNil(t, response.Content)
}

func TestFindSchemaByName(t *testing.T) {
	// Remove  to avoid race conditions

	// Clear schemas before test
	schemasMutex.Lock()
	schemas = make(map[string]*SchemaInfo)
	schemasMutex.Unlock()

	// Register a test schema
	testSchema := &SchemaInfo{
		Name:        "User",
		Description: "User entity",
		Type:        "object",
	}
	RegisterSchema(testSchema)

	// Test finding existing schema
	schema := findSchemaByName("User")
	assert.NotNil(t, schema)
	assert.Equal(t, "User", schema.Name)

	// Test finding non-existing schema
	schema = findSchemaByName("NonExistent")
	assert.Nil(t, schema)
}

func TestFindSchemaByPattern(t *testing.T) {
	// Remove  to avoid race conditions

	// Clear schemas before test
	schemasMutex.Lock()
	schemas = make(map[string]*SchemaInfo)
	schemasMutex.Unlock()

	// Register test schemas
	RegisterSchema(&SchemaInfo{
		Name:        "User",
		Description: "User entity",
		Type:        "object",
	})
	RegisterSchema(&SchemaInfo{
		Name:        "UserResponse",
		Description: "User response",
		Type:        "object",
	})

	// Test finding schema by pattern
	schema := findSchemaByPattern("User")
	assert.NotNil(t, schema)
	assert.Equal(t, "User", schema.Name)

	// Test finding schema with partial match
	schema = findSchemaByPattern("UserResponse")
	assert.NotNil(t, schema)
	assert.Equal(t, "UserResponse", schema.Name)
}

func TestAddDefaultSecuritySchemes(t *testing.T) {
	// Remove  to avoid race conditions

	components := &OpenAPIComponents{
		SecuritySchemes: make(map[string]SecurityScheme), // Initialize the map
	}
	addDefaultSecuritySchemes(components)

	assert.NotNil(t, components.SecuritySchemes)
	assert.Contains(t, components.SecuritySchemes, "BearerAuth")
	assert.Contains(t, components.SecuritySchemes, "ApiKeyAuth")

	bearerScheme := components.SecuritySchemes["BearerAuth"]
	assert.Equal(t, "http", bearerScheme.Type)
	assert.Equal(t, "bearer", bearerScheme.Scheme)
	assert.Equal(t, "JWT", bearerScheme.BearerFormat)

	apiKeyScheme := components.SecuritySchemes["ApiKeyAuth"]
	assert.Equal(t, "apiKey", apiKeyScheme.Type)
	assert.Equal(t, "X-API-Key", apiKeyScheme.Name)
	assert.Equal(t, "header", apiKeyScheme.In)
}

func TestConvertSchemaInfoToOpenAPISchema(t *testing.T) {
	// Remove  to avoid race conditions

	schemaInfo := &SchemaInfo{
		Name:        "User",
		Description: "User entity",
		Type:        "object",
		Properties: map[string]*PropertyInfo{
			"id": {
				Name:        "id",
				Type:        "integer",
				Description: "User ID",
				Required:    true,
			},
			"name": {
				Name:        "name",
				Type:        "string",
				Description: "User name",
				Required:    true,
			},
			"email": {
				Name:        "email",
				Type:        "string",
				Format:      "email",
				Description: "User email",
				Required:    true,
			},
		},
		Required: []string{"name", "email"},
	}

	schema := convertSchemaInfoToOpenAPISchema(schemaInfo)

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema.Type)
	assert.Equal(t, "User entity", schema.Description)
	assert.Len(t, schema.Properties, 3)
	assert.Len(t, schema.Required, 2)
	assert.Contains(t, schema.Required, "name")
	assert.Contains(t, schema.Required, "email")

	// Check properties
	idProp := schema.Properties["id"]
	assert.Equal(t, "integer", idProp.Type)
	assert.Equal(t, "User ID", idProp.Description)

	nameProp := schema.Properties["name"]
	assert.Equal(t, "string", nameProp.Type)
	assert.Equal(t, "User name", nameProp.Description)

	emailProp := schema.Properties["email"]
	assert.Equal(t, "string", emailProp.Type)
	assert.Equal(t, "email", emailProp.Format)
	assert.Equal(t, "User email", emailProp.Description)
}

func TestOpenAPIJSONHandler(t *testing.T) {
	// Remove  to avoid race conditions

	config := &Config{
		OpenAPI: OpenAPIConfig{
			Title:       "Test API",
			Description: "Test API Description",
			Version:     "1.0.0",
		},
	}

	setupGinTestMode(t)
	router := gin.New()
	router.GET("/openapi.json", OpenAPIJSONHandler(config))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/openapi.json", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var spec map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &spec)
	assert.NoError(t, err)
	assert.Equal(t, "3.0.0", spec["openapi"])
	assert.Equal(t, "Test API", spec["info"].(map[string]interface{})["title"])
}

func TestOpenAPIYAMLHandler(t *testing.T) {
	// Remove  to avoid race conditions

	config := &Config{
		OpenAPI: OpenAPIConfig{
			Title:       "Test API",
			Description: "Test API Description",
			Version:     "1.0.0",
		},
	}

	setupGinTestMode(t)
	router := gin.New()
	router.GET("/openapi.yaml", OpenAPIYAMLHandler(config))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/openapi.yaml", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/yaml")

	// Verify YAML content starts with expected structure
	body := w.Body.String()
	assert.Contains(t, body, "openapi: 3.0.0")
	assert.Contains(t, body, "title: Test API")
}

func TestSwaggerUIHandler(t *testing.T) {
	// Remove  to avoid race conditions

	config := &Config{
		OpenAPI: OpenAPIConfig{
			Title:       "Test API",
			Description: "Test API Description",
			Version:     "1.0.0",
		},
	}

	setupGinTestMode(t)
	router := gin.New()
	router.GET("/swagger-ui", SwaggerUIHandler(config))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger-ui", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")

	// Verify HTML content contains Swagger UI elements
	body := w.Body.String()
	assert.Contains(t, body, "<!DOCTYPE html>")
	assert.Contains(t, body, "swagger-ui")
	assert.Contains(t, body, "/openapi.json")
}

func TestSwaggerRedirectHandler(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()

	setupGinTestMode(t)
	router := gin.New()
	router.GET("/swagger", SwaggerRedirectHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/swagger", http.NoBody)
	router.ServeHTTP(w, req)

	// Accept both 301 and 302 as valid redirect codes
	assert.True(t, w.Code == http.StatusMovedPermanently || w.Code == http.StatusFound)
	assert.Equal(t, "/decorators/swagger-ui", w.Header().Get("Location"))
}

func TestOpenAPISpec_ComplexExample(t *testing.T) {
	// Clear global state before test
	registryMutex.Lock()
	routes = routes[:0]                  // Clear routes
	groups = make(map[string]*GroupInfo) // Clear groups
	registryMutex.Unlock()

	config := &Config{
		OpenAPI: OpenAPIConfig{
			Title:       "E-Commerce API",
			Description: "Complete e-commerce API with authentication, products, orders, and payments",
			Version:     "2.0.0",
			Host:        "api.ecommerce.com",
			BasePath:    "/api/v1",
			Schemes:     []string{"https"},
		},
	}

	// Register complex routes
	routes := []*RouteEntry{
		{
			Method:      "POST",
			Path:        "/auth/login",
			Handler:     func(_ *gin.Context) {},
			Summary:     "User login",
			Description: "Authenticate user and return JWT token",
			Tags:        []string{"authentication"},
			Parameters: []ParameterInfo{
				{Name: "email", Type: "string", Location: "body", Required: true, Description: "User email"},
				{Name: "password", Type: "string", Location: "body", Required: true, Description: "User password"},
			},
			Responses: []ResponseInfo{
				{Code: "200", Description: "Login successful", Type: "User"},
				{Code: "401", Description: "Invalid credentials", Type: "Error"},
			},
		},
		{
			Method:      "GET",
			Path:        "/users",
			Handler:     func(_ *gin.Context) {},
			Summary:     "Get users",
			Description: "Retrieve list of users (admin only)",
			Tags:        []string{"users"},
			Parameters: []ParameterInfo{
				{Name: "page", Type: "integer", Location: "query", Required: false, Description: "Page number"},
				{Name: "limit", Type: "integer", Location: "query", Required: false, Description: "Items per page"},
			},
			Responses: []ResponseInfo{
				{Code: "200", Description: "List of users", Type: "UserList"},
				{Code: "403", Description: "Access denied", Type: "Error"},
			},
		},
		{
			Method:      "POST",
			Path:        "/products",
			Handler:     func(_ *gin.Context) {},
			Summary:     "Create product",
			Description: "Create a new product (admin only)",
			Tags:        []string{"products"},
			Parameters: []ParameterInfo{
				{Name: "name", Type: "string", Location: "body", Required: true, Description: "Product name"},
				{Name: "price", Type: "number", Location: "body", Required: true, Description: "Product price"},
				{Name: "description", Type: "string", Location: "body", Required: false, Description: "Product description"},
			},
			Responses: []ResponseInfo{
				{Code: "201", Description: "Product created", Type: "Product"},
				{Code: "400", Description: "Invalid product data", Type: "Error"},
				{Code: "403", Description: "Access denied", Type: "Error"},
			},
		},
	}

	for _, route := range routes {
		RegisterRouteWithMeta(route)
	}

	spec := GenerateOpenAPISpec(config)

	// Verify spec structure
	assert.Equal(t, "3.0.0", spec.OpenAPI)
	assert.Equal(t, "E-Commerce API", spec.Info.Title)
	assert.Equal(t, "2.0.0", spec.Info.Version)
	assert.Equal(t, "Complete e-commerce API with authentication, products, orders, and payments", spec.Info.Description)

	// Verify paths
	assert.Len(t, spec.Paths, 3)
	assert.NotNil(t, spec.Paths["/auth/login"])
	assert.NotNil(t, spec.Paths["/users"])
	assert.NotNil(t, spec.Paths["/products"])

	// Verify specific path details
	loginPath := spec.Paths["/auth/login"]
	assert.NotNil(t, loginPath["post"])
	assert.Equal(t, "User login", loginPath["post"].Summary)
	assert.Equal(t, []string{"authentication"}, loginPath["post"].Tags)
	assert.NotNil(t, loginPath["post"].RequestBody)
	assert.Len(t, loginPath["post"].Responses, 2)

	usersPath := spec.Paths["/users"]
	assert.NotNil(t, usersPath["get"])
	assert.Equal(t, "Get users", usersPath["get"].Summary)
	assert.Equal(t, []string{"users"}, usersPath["get"].Tags)
	assert.Len(t, usersPath["get"].Parameters, 2)
	assert.Len(t, usersPath["get"].Responses, 2)

	productsPath := spec.Paths["/products"]
	assert.NotNil(t, productsPath["post"])
	assert.Equal(t, "Create product", productsPath["post"].Summary)
	assert.Equal(t, []string{"products"}, productsPath["post"].Tags)
	assert.NotNil(t, productsPath["post"].RequestBody)
	assert.Len(t, productsPath["post"].Responses, 3)
}

func TestOpenAPISpec_Concurrency(t *testing.T) {
	// Clear global state before test
	registryMutex.Lock()
	routes = routes[:0]                  // Clear routes
	groups = make(map[string]*GroupInfo) // Clear groups
	registryMutex.Unlock()

	config := &Config{
		OpenAPI: OpenAPIConfig{
			Title:   "Concurrent API",
			Version: "1.0.0",
		},
	}

	// Test concurrent spec generation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			spec := GenerateOpenAPISpec(config)
			assert.Equal(t, "Concurrent API", spec.Info.Title)
			assert.Equal(t, "1.0.0", spec.Info.Version)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestOpenAPISpec_ErrorHandling(t *testing.T) {
	// Test with nil config
	t.Run("should handle nil config", func(t *testing.T) {
		spec := GenerateOpenAPISpec(nil)
		assert.NotNil(t, spec)
		assert.Equal(t, "3.0.0", spec.OpenAPI)
		assert.Equal(t, "gin-decorators API", spec.Info.Title)
		assert.Equal(t, "1.0.0", spec.Info.Version)
	})

	// Test with empty config
	t.Run("should handle empty config", func(t *testing.T) {
		config := &Config{}
		spec := GenerateOpenAPISpec(config)
		assert.NotNil(t, spec)
		assert.Equal(t, "3.0.0", spec.OpenAPI)
		assert.Equal(t, "gin-decorators API", spec.Info.Title)
		assert.Equal(t, "1.0.0", spec.Info.Version)
	})

	// Test handler with invalid JSON
	t.Run("should handle invalid JSON in handler", func(t *testing.T) {
		config := &Config{
			OpenAPI: OpenAPIConfig{
				Title:   "Test API",
				Version: "1.0.0",
			},
		}

		setupGinTestMode(t)
		router := gin.New()
		router.GET("/openapi.json", OpenAPIJSONHandler(config))

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/openapi.json", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		// Verify response is valid JSON
		var spec map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &spec)
		assert.NoError(t, err)
		assert.Equal(t, "3.0.0", spec["openapi"])
	})
}

func TestOpenAPISpec_Integration(t *testing.T) {
	// Clear global state before test
	registryMutex.Lock()
	routes = routes[:0]                  // Clear routes
	groups = make(map[string]*GroupInfo) // Clear groups
	registryMutex.Unlock()

	// Test integration with Gin router
	t.Run("should integrate with Gin router", func(t *testing.T) {
		config := &Config{
			OpenAPI: OpenAPIConfig{
				Title:   "Integration Test API",
				Version: "1.0.0",
			},
		}

		// Register routes
		RegisterRouteWithMeta(&RouteEntry{
			Method:      "GET",
			Path:        "/health",
			Handler:     func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "healthy"}) },
			Summary:     "Health check",
			Description: "Check API health status",
			Tags:        []string{"health"},
			Responses: []ResponseInfo{
				{Code: "200", Description: "API is healthy", Type: "Health"},
			},
		})

		RegisterRouteWithMeta(&RouteEntry{
			Method:      "POST",
			Path:        "/data",
			Handler:     func(c *gin.Context) { c.JSON(http.StatusCreated, gin.H{"message": "created"}) },
			Summary:     "Create data",
			Description: "Create new data entry",
			Tags:        []string{"data"},
			Parameters: []ParameterInfo{
				{Name: "name", Type: "string", Location: "body", Required: true, Description: "Data name"},
				{Name: "value", Type: "number", Location: "body", Required: false, Description: "Data value"},
			},
			Responses: []ResponseInfo{
				{Code: "201", Description: "Data created", Type: "Data"},
				{Code: "400", Description: "Invalid data", Type: "Error"},
			},
		})

		// Create Gin router
		setupGinTestMode(t)
		router := gin.New()

		// Add OpenAPI handler
		router.GET("/openapi.json", OpenAPIJSONHandler(config))

		// Add actual route handlers
		router.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})

		router.POST("/data", func(c *gin.Context) {
			var data map[string]interface{}
			if err := c.ShouldBindJSON(&data); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, data)
		})

		// Test OpenAPI endpoint
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/openapi.json", http.NoBody)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

		var spec map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &spec)
		assert.NoError(t, err)
		assert.Equal(t, "Integration Test API", spec["info"].(map[string]interface{})["title"])

		// Test actual endpoints
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/health", http.NoBody)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/data", strings.NewReader(`{"name":"test","value":42}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
