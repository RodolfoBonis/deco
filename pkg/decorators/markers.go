package decorators

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// MarkerFactory function that creates a middleware based on arguments
type MarkerFactory func(args []string) gin.HandlerFunc

// MarkerConfig configuration of a marker
type MarkerConfig struct {
	Name        string                              // Marker name (ex: "Auth")
	Pattern     *regexp.Regexp                      // Regex to detect the marker
	Factory     func(args []string) gin.HandlerFunc // Factory to create middleware
	Description string                              // Marker description
}

// global markers registry
var markers = make(map[string]MarkerConfig)

// init registers default markers automatically
func init() {
	initDefaultMarkers()
}

// RegisterMarker registers a new marker in the framework
func RegisterMarker(config MarkerConfig) {
	markers[config.Name] = config
	LogVerbose("Marker registered: %s", config.Name)
}

// GetMarkers returns all registered markers
func GetMarkers() map[string]MarkerConfig {
	return markers
}

// initDefaultMarkers registers framework default markers
func initDefaultMarkers() {
	// Middleware markers
	RegisterMarker(MarkerConfig{
		Name:    "Auth",
		Pattern: regexp.MustCompile(`@Auth\s*\(([^)]*)\)`),
		Factory: createAuthMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "Cache",
		Pattern: regexp.MustCompile(`@Cache\s*\(([^)]*)\)`),
		Factory: createCacheMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "CacheByURL",
		Pattern: regexp.MustCompile(`@CacheByURL\s*\(([^)]*)\)`),
		Factory: createCacheByURLMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "CacheByUser",
		Pattern: regexp.MustCompile(`@CacheByUser\s*\(([^)]*)\)`),
		Factory: createCacheByUserMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "CacheByEndpoint",
		Pattern: regexp.MustCompile(`@CacheByEndpoint\s*\(([^)]*)\)`),
		Factory: createCacheByEndpointMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "RateLimit",
		Pattern: regexp.MustCompile(`@RateLimit\s*\(([^)]*)\)`),
		Factory: createRateLimitMiddlewareInternal,
	})

	RegisterMarker(MarkerConfig{
		Name:    "RateLimitByIP",
		Pattern: regexp.MustCompile(`@RateLimitByIP\s*\(([^)]*)\)`),
		Factory: createRateLimitByIPMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "RateLimitByUser",
		Pattern: regexp.MustCompile(`@RateLimitByUser\s*\(([^)]*)\)`),
		Factory: createRateLimitByUserMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "RateLimitByEndpoint",
		Pattern: regexp.MustCompile(`@RateLimitByEndpoint\s*\(([^)]*)\)`),
		Factory: createRateLimitByEndpointMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "Metrics",
		Pattern: regexp.MustCompile(`@Metrics\s*\(([^)]*)\)`),
		Factory: createMetricsMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "Prometheus",
		Pattern: regexp.MustCompile(`@Prometheus\s*\(([^)]*)\)`),
		Factory: createPrometheusMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "HealthCheck",
		Pattern: regexp.MustCompile(`@HealthCheck\s*\(([^)]*)\)`),
		Factory: createHealthCheckMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "CacheStats",
		Pattern: regexp.MustCompile(`@CacheStats\s*\(([^)]*)\)`),
		Factory: createCacheStatsMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "InvalidateCache",
		Pattern: regexp.MustCompile(`@InvalidateCache\s*\(([^)]*)\)`),
		Factory: createInvalidateCacheMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "WebSocketStats",
		Pattern: regexp.MustCompile(`@WebSocketStats\s*\(([^)]*)\)`),
		Factory: createWebSocketStatsMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "TracingStats",
		Pattern: regexp.MustCompile(`@TracingStats\s*\(([^)]*)\)`),
		Factory: createTracingStatsMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "OpenAPIJSON",
		Pattern: regexp.MustCompile(`@OpenAPIJSON\s*\(([^)]*)\)`),
		Factory: createOpenAPIJSONMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "OpenAPIYAML",
		Pattern: regexp.MustCompile(`@OpenAPIYAML\s*\(([^)]*)\)`),
		Factory: createOpenAPIYAMLMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "SwaggerUI",
		Pattern: regexp.MustCompile(`@SwaggerUI\s*\(([^)]*)\)`),
		Factory: createSwaggerUIMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "TraceMiddleware",
		Pattern: regexp.MustCompile(`@TraceMiddleware\s*\(([^)]*)\)`),
		Factory: createTraceMiddlewareWrapper,
	})

	RegisterMarker(MarkerConfig{
		Name:    "HealthCheckWithTracing",
		Pattern: regexp.MustCompile(`@HealthCheckWithTracing\s*\(([^)]*)\)`),
		Factory: createHealthCheckWithTracingMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "InstrumentedHandler",
		Pattern: regexp.MustCompile(`@InstrumentedHandler\s*\(([^)]*)\)`),
		Factory: createInstrumentedHandlerMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "Validate",
		Pattern: regexp.MustCompile(`@Validate\s*\(([^)]*)\)`),
		Factory: createValidateMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "ValidateJSON",
		Pattern: regexp.MustCompile(`@ValidateJSON\s*\(([^)]*)\)`),
		Factory: createValidateJSONMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "ValidateQuery",
		Pattern: regexp.MustCompile(`@ValidateQuery\s*\(([^)]*)\)`),
		Factory: createValidateQueryMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "ValidateParams",
		Pattern: regexp.MustCompile(`@ValidateParams\s*\(([^)]*)\)`),
		Factory: createValidateParamsMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "CORS",
		Pattern: regexp.MustCompile(`@CORS\s*\(([^)]*)\)`),
		Factory: createCORSMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "Telemetry",
		Pattern: regexp.MustCompile(`@Telemetry\s*\(([^)]*)\)`),
		Factory: createTelemetryMiddleware,
	})

	RegisterMarker(MarkerConfig{
		Name:    "WebSocket",
		Pattern: regexp.MustCompile(`@WebSocket\s*\(([^)]*)\)`),
		Factory: createWebSocketMiddleware,
	})

	// Documentation markers
	RegisterMarker(MarkerConfig{
		Name:    "Group",
		Pattern: regexp.MustCompile(`@Group\s*\(([^)]*)\)`),
		Factory: nil, // Does not generate middleware
	})

	RegisterMarker(MarkerConfig{
		Name:    "Param",
		Pattern: regexp.MustCompile(`@Param\s*\(([^)]*)\)`),
		Factory: nil, // Does not generate middleware
	})

	RegisterMarker(MarkerConfig{
		Name:    "Description",
		Pattern: regexp.MustCompile(`@Description\s*\(([^)]*)\)`),
		Factory: nil, // Does not generate middleware
	})

	RegisterMarker(MarkerConfig{
		Name:    "Summary",
		Pattern: regexp.MustCompile(`@Summary\s*\(([^)]*)\)`),
		Factory: nil, // Does not generate middleware
	})

	RegisterMarker(MarkerConfig{
		Name:    "Schema",
		Pattern: regexp.MustCompile(`@Schema\s*\(([^)]*)\)`),
		Factory: nil, // Documentation only - does not generate middleware
	})

	RegisterMarker(MarkerConfig{
		Name:    "Tag",
		Pattern: regexp.MustCompile(`@Tag\s*\(([^)]*)\)`),
		Factory: nil, // Does not generate middleware
	})

	RegisterMarker(MarkerConfig{
		Name:    "Response",
		Pattern: regexp.MustCompile(`@Response\s*\(([^)]*)\)`),
		Factory: nil, // Does not generate middleware
	})
}

// createAuthMiddleware creates authentication middleware
func createAuthMiddleware(args []string) gin.HandlerFunc {
	var role string
	if len(args) > 0 && args[0] != "" {
		role = parseKeyValue(args[0], "role")
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "Token de autorização requerido"})
			c.Abort()
			return
		}

		// Basic token validation (in production use JWT)
		if !strings.HasPrefix(token, "Bearer ") {
			c.JSON(401, gin.H{"error": "Token inválido"})
			c.Abort()
			return
		}

		// If role specified, validate
		if role != "" {
			// Role validation logic (simulated)
			c.Set("user_role", role)
		}

		c.Set("authenticated", true)
		c.Next()
	})
}

// createCacheMiddleware creates cache middleware
func createCacheMiddleware(args []string) gin.HandlerFunc {
	duration, cacheType, keyGen := ParseCacheArgs(args)

	config := &CacheConfig{
		Type:       cacheType,
		DefaultTTL: duration.String(),
		MaxSize:    1000,
	}

	return CacheMiddleware(config, keyGen)
}

// createCORSMiddleware creates CORS middleware
func createCORSMiddleware(args []string) gin.HandlerFunc {
	origins := "*"
	if len(args) > 0 && args[0] != "" {
		origins = parseKeyValue(args[0], "origins")
	}

	return gin.HandlerFunc(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origins)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

// parseKeyValue extracts value from a key=value string
func parseKeyValue(input, key string) string {
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		kv := strings.Split(strings.TrimSpace(pair), "=")
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == key {
			return strings.Trim(strings.TrimSpace(kv[1]), `"'`)
		}
	}
	return ""
}

// createValidateMiddleware creates general validation middleware
func createValidateMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Validation
	return ValidateStruct(&config)
}

// createValidateJSONMiddleware creates JSON validation middleware
func createValidateJSONMiddleware(args []string) gin.HandlerFunc {
	// For now, return basic middleware - target will be inferred at runtime

	return gin.HandlerFunc(func(c *gin.Context) {
		// Generic middleware that tries to validate any JSON struct
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			response := ValidationResponse{
				Error:   "validation_failed",
				Message: "Formato JSON inválido",
				Fields: []ValidationField{{
					Field:   "json",
					Message: err.Error(),
				}},
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, response)
			return
		}
		c.Set("validated_data", data)
		c.Next()
	})
}

// createValidateQueryMiddleware creates query parameter validation middleware
func createValidateQueryMiddleware(args []string) gin.HandlerFunc {

	return gin.HandlerFunc(func(c *gin.Context) {
		// Validate basic query parameters
		query := c.Request.URL.Query()
		validatedQuery := make(map[string]string)

		for key, values := range query {
			if len(values) > 0 {
				validatedQuery[key] = values[0]
			}
		}

		c.Set("validated_query", validatedQuery)
		c.Next()
	})
}

// createValidateParamsMiddleware creates path parameter validation middleware
func createValidateParamsMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Validation

	// Extract rules from arguments
	rules := make(map[string]string)
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			rules[key] = value
		}
	}

	return ValidateParams(rules, &config)
}

// createWebSocketMiddleware creates WebSocket middleware
func createWebSocketMiddleware(args []string) gin.HandlerFunc {
	// Load configuration from file instead of using defaults
	fullConfig, err := LoadConfig("")
	if err != nil {
		// Fallback to default if error
		config := DefaultConfig().WebSocket
		return CreateWebSocketHandler(&config)
	}
	return CreateWebSocketHandler(&fullConfig.WebSocket)
}

// createCacheByURLMiddleware creates URL-based cache middleware
func createCacheByURLMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Cache
	return CacheByURL(&config)
}

// createCacheByUserMiddleware creates user+URL-based cache middleware
func createCacheByUserMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Cache
	return CacheByUserURL(&config)
}

// createPrometheusMiddleware creates Prometheus metrics middleware
func createPrometheusMiddleware(args []string) gin.HandlerFunc {
	return PrometheusHandler()
}

// createHealthCheckMiddleware creates health check middleware
func createHealthCheckMiddleware(args []string) gin.HandlerFunc {
	return HealthCheckHandler()
}

// createCacheStatsMiddleware creates cache statistics middleware
func createCacheStatsMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Cache
	store := NewMemoryCache(config.MaxSize)
	return CacheStatsHandler(store)
}

// createInvalidateCacheMiddleware creates cache invalidation middleware
func createInvalidateCacheMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Cache
	store := NewMemoryCache(config.MaxSize)
	return InvalidateCacheHandler(store)
}

// createWebSocketStatsMiddleware creates WebSocket statistics middleware
func createWebSocketStatsMiddleware(args []string) gin.HandlerFunc {
	return WebSocketStatsHandler()
}

// createTracingStatsMiddleware creates tracing statistics middleware
func createTracingStatsMiddleware(args []string) gin.HandlerFunc {
	return TracingStatsHandler()
}

// createOpenAPIJSONMiddleware creates OpenAPI JSON middleware
func createOpenAPIJSONMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig()
	return OpenAPIJSONHandler(config)
}

// createOpenAPIYAMLMiddleware creates OpenAPI YAML middleware
func createOpenAPIYAMLMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig()
	return OpenAPIYAMLHandler(config)
}

// createSwaggerUIMiddleware creates Swagger UI middleware
func createSwaggerUIMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig()
	return SwaggerUIHandler(config)
}

// createTraceMiddlewareWrapper creates trace middleware wrapper
func createTraceMiddlewareWrapper(args []string) gin.HandlerFunc {
	middlewareName := "middleware"
	if len(args) > 0 && args[0] != "" {
		middlewareName = strings.Trim(args[0], `"'`)
	}
	return TraceMiddleware(middlewareName)
}

// createCacheByEndpointMiddleware creates endpoint-based cache middleware
func createCacheByEndpointMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Cache
	return CacheByEndpoint(&config)
}

// createRateLimitByIPMiddleware creates IP-based rate limiting middleware
func createRateLimitByIPMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().RateLimit
	return RateLimitByIP(&config)
}

// createRateLimitByUserMiddleware creates user-based rate limiting middleware
func createRateLimitByUserMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().RateLimit
	return RateLimitByUser(&config)
}

// createRateLimitByEndpointMiddleware creates endpoint-based rate limiting middleware
func createRateLimitByEndpointMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().RateLimit
	return RateLimitByEndpoint(&config)
}

// createHealthCheckWithTracingMiddleware creates health check with tracing middleware
func createHealthCheckWithTracingMiddleware(args []string) gin.HandlerFunc {
	return HealthCheckWithTracing()
}

// createInstrumentedHandlerMiddleware creates instrumented handler middleware
func createInstrumentedHandlerMiddleware(args []string) gin.HandlerFunc {
	handlerName := "handler"
	if len(args) > 0 && args[0] != "" {
		handlerName = strings.Trim(args[0], `"'`)
	}
	// Return a middleware that instruments the next handler
	return InstrumentedHandler(handlerName, func(c *gin.Context) {
		c.Next()
	})
}
