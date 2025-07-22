package decorators

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegisterGroup(t *testing.T) {
	// Clear existing groups
	groups = make(map[string]*GroupInfo)

	// Test registering a group
	group := RegisterGroup("api", "/api/v1", "API group")
	assert.NotNil(t, group)
	assert.Equal(t, "api", group.Name)
	assert.Equal(t, "/api/v1", group.Prefix)
	assert.Equal(t, "API group", group.Description)

	// Verify group is stored
	storedGroup := groups["api"]
	assert.Equal(t, group, storedGroup)
}

func TestGetGroup(t *testing.T) {
	// Clear existing groups
	groups = make(map[string]*GroupInfo)

	// Register a group
	group := RegisterGroup("test", "/test", "Test group")

	// Test getting the group
	retrievedGroup := GetGroup("test")
	assert.Equal(t, group, retrievedGroup)

	// Test getting non-existent group
	nonExistentGroup := GetGroup("non-existent")
	assert.Nil(t, nonExistentGroup)
}

func TestGetGroups(t *testing.T) {
	// Clear existing groups
	groups = make(map[string]*GroupInfo)

	// Register multiple groups
	group1 := RegisterGroup("api", "/api", "API group")
	group2 := RegisterGroup("admin", "/admin", "Admin group")

	// Test getting all groups
	allGroups := GetGroups()
	assert.Len(t, allGroups, 2)
	assert.Equal(t, group1, allGroups["api"])
	assert.Equal(t, group2, allGroups["admin"])
}

func TestRegisterRoute(t *testing.T) {
	// Clear existing routes
	routes = nil

	// Test registering a route
	handler := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	}

	RegisterRoute("GET", "/test", handler)
	assert.Len(t, routes, 1)
	assert.Equal(t, "GET", routes[0].Method)
	assert.Equal(t, "/test", routes[0].Path)
	assert.NotNil(t, routes[0].Handler)
}

func TestRegisterRouteWithMeta(t *testing.T) {
	// Clear existing routes
	routes = nil

	// Test registering a route with metadata
	handler := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	}

	entry := &RouteEntry{
		Method:      "POST",
		Path:        "/test",
		Handler:     handler,
		FuncName:    "TestHandler",
		PackageName: "test",
		FileName:    "test.go",
		Description: "Test route",
		Summary:     "Test summary",
		Tags:        []string{"test"},
		Parameters: []ParameterInfo{
			{Name: "id", Type: "string", Location: "path", Required: true},
		},
		Responses: []ResponseInfo{
			{Code: "200", Description: "Success", Type: "TestResponse"},
		},
	}

	RegisterRouteWithMeta(entry)
	assert.Len(t, routes, 1)
	assert.Equal(t, "POST", routes[0].Method)
	assert.Equal(t, "/test", routes[0].Path)
	assert.Equal(t, "TestHandler", routes[0].FuncName)
	assert.Equal(t, "test", routes[0].PackageName)
	assert.Equal(t, "test.go", routes[0].FileName)
	assert.Equal(t, "Test route", routes[0].Description)
	assert.Equal(t, "Test summary", routes[0].Summary)
	assert.Len(t, routes[0].Tags, 1)
	assert.Equal(t, "test", routes[0].Tags[0])
	assert.Len(t, routes[0].Parameters, 1)
	assert.Equal(t, "id", routes[0].Parameters[0].Name)
	assert.Len(t, routes[0].Responses, 1)
	assert.Equal(t, "200", routes[0].Responses[0].Code)
}

func TestGetRoutes(t *testing.T) {
	// Clear existing routes
	routes = nil

	// Register multiple routes
	handler1 := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test1"})
	}
	handler2 := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test2"})
	}

	RegisterRoute("GET", "/test1", handler1)
	RegisterRoute("POST", "/test2", handler2)

	// Test getting all routes
	allRoutes := GetRoutes()
	assert.Len(t, allRoutes, 2)
	assert.Equal(t, "GET", allRoutes[0].Method)
	assert.Equal(t, "/test1", allRoutes[0].Path)
	assert.Equal(t, "POST", allRoutes[1].Method)
	assert.Equal(t, "/test2", allRoutes[1].Path)
}

func TestGetFuncName(t *testing.T) {
	// Test getting function name
	handler := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	}

	funcName := getFuncName(handler)
	assert.NotEmpty(t, funcName)
}

func TestDefault(t *testing.T) {
	// Test creating default engine
	engine := Default()
	assert.NotNil(t, engine)
	assert.IsType(t, &gin.Engine{}, engine)
}

func TestDefaultWithSecurity(t *testing.T) {
	// Test creating default engine with security
	securityConfig := &SecurityConfig{
		AllowedHosts: []string{"localhost"},
	}

	engine := DefaultWithSecurity(securityConfig)
	assert.NotNil(t, engine)
	assert.IsType(t, &gin.Engine{}, engine)
}

func TestParameterInfo_Structure(t *testing.T) {
	// Test ParameterInfo structure
	param := ParameterInfo{
		Name:        "id",
		Type:        "string",
		Location:    "path",
		Required:    true,
		Description: "User ID",
		Example:     "123",
	}

	assert.Equal(t, "id", param.Name)
	assert.Equal(t, "string", param.Type)
	assert.Equal(t, "path", param.Location)
	assert.True(t, param.Required)
	assert.Equal(t, "User ID", param.Description)
	assert.Equal(t, "123", param.Example)
}

func TestResponseInfo_Structure(t *testing.T) {
	// Test ResponseInfo structure
	response := ResponseInfo{
		Code:        "200",
		Description: "Success",
		Type:        "UserResponse",
		Example:     `{"id": 1, "name": "John"}`,
	}

	assert.Equal(t, "200", response.Code)
	assert.Equal(t, "Success", response.Description)
	assert.Equal(t, "UserResponse", response.Type)
	assert.Equal(t, `{"id": 1, "name": "John"}`, response.Example)
}

func TestGroupInfo_Structure(t *testing.T) {
	// Test GroupInfo structure
	group := GroupInfo{
		Name:        "api",
		Prefix:      "/api/v1",
		Description: "API group",
	}

	assert.Equal(t, "api", group.Name)
	assert.Equal(t, "/api/v1", group.Prefix)
	assert.Equal(t, "API group", group.Description)
}

func TestRouteEntry_Structure(t *testing.T) {
	// Test RouteEntry structure
	handler := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	}

	entry := RouteEntry{
		Method:            "GET",
		Path:              "/test",
		Handler:           handler,
		Middlewares:       []gin.HandlerFunc{},
		FuncName:          "TestHandler",
		PackageName:       "test",
		FileName:          "test.go",
		Description:       "Test route",
		Summary:           "Test summary",
		Tags:              []string{"test"},
		MiddlewareInfo:    []MiddlewareInfo{},
		Parameters:        []ParameterInfo{},
		Group:             &GroupInfo{},
		Responses:         []ResponseInfo{},
		WebSocketHandlers: []string{},
	}

	assert.Equal(t, "GET", entry.Method)
	assert.Equal(t, "/test", entry.Path)
	assert.NotNil(t, entry.Handler)
	assert.Equal(t, "TestHandler", entry.FuncName)
	assert.Equal(t, "test", entry.PackageName)
	assert.Equal(t, "test.go", entry.FileName)
	assert.Equal(t, "Test route", entry.Description)
	assert.Equal(t, "Test summary", entry.Summary)
	assert.Len(t, entry.Tags, 1)
}

func TestRegistry_Concurrency(t *testing.T) {
	// Test registry concurrency
	// Clear existing data
	routes = nil
	groups = make(map[string]*GroupInfo)

	// Run concurrent operations
	done := make(chan bool, 4)

	go func() {
		RegisterGroup("group1", "/group1", "Group 1")
		done <- true
	}()

	go func() {
		RegisterGroup("group2", "/group2", "Group 2")
		done <- true
	}()

	go func() {
		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		}
		RegisterRoute("GET", "/test1", handler)
		done <- true
	}()

	go func() {
		handler := func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		}
		RegisterRoute("POST", "/test2", handler)
		done <- true
	}()

	// Wait for all operations to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Verify results
	allGroups := GetGroups()
	allRoutes := GetRoutes()

	assert.Len(t, allGroups, 2)
	assert.Len(t, allRoutes, 2)
}

func TestRegistry_EdgeCases(t *testing.T) {
	// Test edge cases
	// Clear existing data
	routes = nil
	groups = make(map[string]*GroupInfo)

	// Test registering empty group
	group := RegisterGroup("", "", "")
	assert.NotNil(t, group)
	assert.Equal(t, "", group.Name)
	assert.Equal(t, "", group.Prefix)

	// Test registering route with empty method/path
	handler := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	}
	RegisterRoute("", "", handler)
	assert.Len(t, routes, 1)
}

func TestRegistry_ErrorHandling(t *testing.T) {
	// Test error handling scenarios
	// Clear existing data
	routes = nil
	groups = make(map[string]*GroupInfo)

	// Test registering duplicate group
	RegisterGroup("test", "/test", "Test group")
	group2 := RegisterGroup("test", "/test2", "Test group 2")

	// Should overwrite the previous group
	assert.Equal(t, group2, GetGroup("test"))
	assert.Equal(t, "/test2", GetGroup("test").Prefix)

	// Test registering multiple routes with same method/path
	handler1 := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test1"})
	}
	handler2 := func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test2"})
	}

	RegisterRoute("GET", "/test", handler1)
	RegisterRoute("GET", "/test", handler2)

	// Both routes should be registered
	allRoutes := GetRoutes()
	assert.Len(t, allRoutes, 2)
	assert.Equal(t, "GET", allRoutes[0].Method)
	assert.Equal(t, "/test", allRoutes[0].Path)
	assert.Equal(t, "GET", allRoutes[1].Method)
	assert.Equal(t, "/test", allRoutes[1].Path)
}
