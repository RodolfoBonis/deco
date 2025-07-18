package decorators

import (
	"log"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParameterInfo represents information of a route parameter
type ParameterInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`     // string, int, bool, etc.
	Location    string `json:"location"` // query, path, body, header
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example"`
}

// ResponseInfo represents information of a route response
type ResponseInfo struct {
	Code        string `json:"code"`        // HTTP status code (200, 404, etc.)
	Description string `json:"description"` // Response description
	Type        string `json:"type"`        // Schema type name (UserResponse, ErrorResponse, etc.)
	Example     string `json:"example"`     // Response example
}

// GroupInfo represents information of a route group
type GroupInfo struct {
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Description string `json:"description"`
}

// RouteEntry represents complete information about a route
type RouteEntry struct {
	Method            string            `json:"method"`
	Path              string            `json:"path"`
	Handler           gin.HandlerFunc   `json:"-"`
	Middlewares       []gin.HandlerFunc `json:"-"`
	FuncName          string            `json:"func_name"`
	PackageName       string            `json:"package_name"`
	FileName          string            `json:"file_name"`
	Description       string            `json:"description"`
	Summary           string            `json:"summary"`
	Tags              []string          `json:"tags"`
	MiddlewareInfo    []MiddlewareInfo  `json:"middleware_info"`
	Parameters        []ParameterInfo   `json:"parameters"`
	Group             *GroupInfo        `json:"group,omitempty"`
	Responses         []ResponseInfo    `json:"responses,omitempty"`         // Updated to use ResponseInfo
	WebSocketHandlers []string          `json:"websocketHandlers,omitempty"` // WebSocket message types this function handles
}

// global route registry
var routes []RouteEntry

// global groups registry
var groups = make(map[string]*GroupInfo)

// RegisterGroup registers a new route group
func RegisterGroup(name, prefix, description string) *GroupInfo {
	group := &GroupInfo{
		Name:        name,
		Prefix:      prefix,
		Description: description,
	}
	groups[name] = group
	LogVerbose("Grupo registrado: %s -> %s", name, prefix)
	return group
}

// GetGroup returns information of a group
func GetGroup(name string) *GroupInfo {
	return groups[name]
}

// GetGroups returns all registered groups
func GetGroups() map[string]*GroupInfo {
	return groups
}

// RegisterRoute registers a new route in the framework
func RegisterRoute(method, path string, handlers ...gin.HandlerFunc) {
	RegisterRouteWithMeta(&RouteEntry{
		Method:      method,
		Path:        path,
		Middlewares: handlers[:len(handlers)-1],
		Handler:     handlers[len(handlers)-1],
		FuncName:    getFuncName(handlers[len(handlers)-1]),
	})
}

// RegisterRouteWithMeta registers a route with complete metadata
func RegisterRouteWithMeta(entry *RouteEntry) {
	if entry.Handler == nil {
		log.Fatalf("RegisterRoute: handler is required for %s %s", entry.Method, entry.Path)
	}

	// If FuncName was not defined, extract from function
	if entry.FuncName == "" {
		entry.FuncName = getFuncName(entry.Handler)
	}

	// Apply group prefix if defined
	if entry.Group != nil {
		if entry.Group.Prefix != "" && !strings.HasPrefix(entry.Path, entry.Group.Prefix) {
			entry.Path = entry.Group.Prefix + entry.Path
		}
		// Add tag do grupo
		entry.Tags = append(entry.Tags, entry.Group.Name)
	}

	routes = append(routes, *entry)
	LogVerbose("Route registrada: %s %s -> %s", entry.Method, entry.Path, entry.FuncName)
}

// Default creates a gin.Engine with all registered routes
func Default() *gin.Engine {
	return DefaultWithSecurity(nil)
}

// DefaultWithSecurity creates a gin.Engine with security configuration for internal endpoints
func DefaultWithSecurity(securityConfig *SecurityConfig) *gin.Engine {
	r := gin.Default()

	// Use default security config if not provided
	if securityConfig == nil {
		securityConfig = DefaultSecurityConfig()
	}

	// Create security middleware for internal endpoints
	securityMiddleware := SecureInternalEndpoints(securityConfig)

	// Register documentation routes with security
	config := DefaultConfig()
	r.GET("/decorators/docs", securityMiddleware, DocsHandler)
	r.GET("/decorators/docs.json", securityMiddleware, DocsJSONHandler)
	r.GET("/decorators/openapi.json", securityMiddleware, OpenAPIJSONHandler(config))
	r.GET("/decorators/openapi.yaml", securityMiddleware, OpenAPIYAMLHandler(config))
	r.GET("/decorators/swagger-ui", securityMiddleware, SwaggerUIHandler(config))
	r.GET("/decorators/swagger", securityMiddleware, SwaggerRedirectHandler)

	// Register all framework routes
	for i := range routes {
		route := &routes[i]
		// Combine middlewares + main handler
		handlers := make([]gin.HandlerFunc, 0, len(route.Middlewares)+1)
		handlers = append(handlers, route.Middlewares...)
		handlers = append(handlers, route.Handler)
		r.Handle(route.Method, route.Path, handlers...)
	}

	LogNormal("Framework gin-decorators inicializado com %d routes", len(routes))
	return r
}

// GetRoutes returns all registered routes (used for documentation)
func GetRoutes() []RouteEntry {
	return routes
}

// getFuncName extracts function name from a handler
func getFuncName(handler gin.HandlerFunc) string {
	value := reflect.ValueOf(handler)
	if value.Kind() == reflect.Func {
		return value.Type().String()
	}
	return "unknown"
}
