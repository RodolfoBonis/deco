// Tests for metrics logic in gin-decorators framework
package decorators

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInitMetrics(t *testing.T) {
	// Reset metrics state before test
	metricsInitMutex.Lock()
	defaultMetricsCollector = nil
	metricsInitialized = false
	metricsInitMutex.Unlock()

	config := &MetricsConfig{
		Enabled:   true,
		Endpoint:  "/metrics",
		Namespace: "test",
		Subsystem: "api",
		Buckets:   []float64{0.1, 0.5, 1.0, 2.0, 5.0},
	}

	collector := InitMetrics(config)
	assert.NotNil(t, collector)

	// Check that metrics are properly initialized
	assert.NotNil(t, collector.httpRequestsTotal)
	assert.NotNil(t, collector.httpRequestDuration)
	assert.NotNil(t, collector.httpRequestSize)
	assert.NotNil(t, collector.httpResponseSize)
	assert.NotNil(t, collector.httpActiveRequests)
	assert.NotNil(t, collector.middlewareExecutionTime)
	assert.NotNil(t, collector.middlewareErrors)
	assert.NotNil(t, collector.cacheHits)
	assert.NotNil(t, collector.cacheMisses)
	assert.NotNil(t, collector.cacheSize)
	assert.NotNil(t, collector.rateLimitHits)
	assert.NotNil(t, collector.rateLimitExceeded)
	assert.NotNil(t, collector.validationErrors)
	assert.NotNil(t, collector.validationTime)
}

func TestMetricsMiddleware(t *testing.T) {
	// Reset metrics state before test
	metricsInitMutex.Lock()
	defaultMetricsCollector = nil
	metricsInitialized = false
	metricsInitMutex.Unlock()

	config := &MetricsConfig{
		Enabled:   true,
		Namespace: "test",
		Subsystem: "api",
		Buckets:   []float64{0.1, 0.5, 1.0},
	}

	middleware := MetricsMiddleware(config)
	assert.NotNil(t, middleware)

	// Test middleware execution
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set up request
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	c.Request = req

	// Execute middleware
	assert.NotPanics(t, func() {
		middleware(c)
	})
}

func TestMetricsResponseWriter(t *testing.T) {

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create metrics response writer
	mrw := &metricsResponseWriter{
		ResponseWriter: c.Writer,
		size:           0,
		status:         0,
	}

	// Test Write method
	data := []byte("test data")
	n, err := mrw.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, len(data), mrw.size)

	// Test WriteHeader method
	mrw.WriteHeader(http.StatusOK)
	assert.Equal(t, http.StatusOK, mrw.status)
}

func TestGetEndpointPattern(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test with simple path
	req, _ := http.NewRequest("GET", "/users", http.NoBody)
	c.Request = req
	c.Params = []gin.Param{}

	pattern := getEndpointPattern(c)
	assert.Equal(t, "/users", pattern)

	// Test with path parameters
	req, _ = http.NewRequest("GET", "/users/123", http.NoBody)
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: "123"}}

	pattern = getEndpointPattern(c)
	assert.Equal(t, "/users/:id", pattern)

	// Test with multiple parameters
	req, _ = http.NewRequest("GET", "/users/123/posts/456", http.NoBody)
	c.Request = req
	c.Params = []gin.Param{
		{Key: "user_id", Value: "123"},
		{Key: "post_id", Value: "456"},
	}

	pattern = getEndpointPattern(c)
	assert.Equal(t, "/users/:user_id/posts/:post_id", pattern)
}

func TestRecordCacheHit(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordCacheHit("memory", "url")
	})

	assert.NotPanics(t, func() {
		RecordCacheHit("redis", "user")
	})
}

func TestRecordCacheMiss(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordCacheMiss("memory", "url")
	})

	assert.NotPanics(t, func() {
		RecordCacheMiss("redis", "user")
	})
}

func TestRecordCacheSize(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordCacheSize("memory", 100.0)
	})

	assert.NotPanics(t, func() {
		RecordCacheSize("redis", 500.0)
	})
}

func TestRecordRateLimitHit(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordRateLimitHit("/api/users", "ip")
	})

	assert.NotPanics(t, func() {
		RecordRateLimitHit("/api/posts", "user")
	})
}

func TestRecordRateLimitExceeded(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordRateLimitExceeded("/api/users", "ip")
	})

	assert.NotPanics(t, func() {
		RecordRateLimitExceeded("/api/posts", "user")
	})
}

func TestRecordValidationError(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordValidationError("json", "email")
	})

	assert.NotPanics(t, func() {
		RecordValidationError("query", "page")
	})
}

func TestRecordValidationTime(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordValidationTime("json", 100*time.Millisecond)
	})

	assert.NotPanics(t, func() {
		RecordValidationTime("query", 50*time.Millisecond)
	})
}

func TestRecordMiddlewareTime(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordMiddlewareTime("auth", "/api/users", 10*time.Millisecond)
	})

	assert.NotPanics(t, func() {
		RecordMiddlewareTime("cache", "/api/posts", 5*time.Millisecond)
	})
}

func TestRecordMiddlewareError(t *testing.T) {

	// Test that function doesn't panic
	assert.NotPanics(t, func() {
		RecordMiddlewareError("auth", "invalid_token")
	})

	assert.NotPanics(t, func() {
		RecordMiddlewareError("rate_limit", "quota_exceeded")
	})
}

func TestPrometheusHandler(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	handler := PrometheusHandler()
	assert.NotNil(t, handler)

	// Test handler execution
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/metrics", http.NoBody)
	c.Request = req

	assert.NotPanics(t, func() {
		handler(c)
	})

	// Check that response is not empty
	assert.NotEqual(t, 0, w.Body.Len())
}

func TestHealthCheckHandler(t *testing.T) {

	handler := HealthCheckHandler()
	assert.NotNil(t, handler)

	// Test handler execution
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/health", http.NoBody)
	c.Request = req

	assert.NotPanics(t, func() {
		handler(c)
	})

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")
}

func TestCreateMetricsMiddlewareFromArgs(t *testing.T) {

	// Test with valid args
	args := []string{"enabled=true", "namespace=test", "subsystem=api"}
	middleware := createMetricsMiddleware(args)
	assert.NotNil(t, middleware)

	// Test with empty args
	args = []string{}
	middleware = createMetricsMiddleware(args)
	assert.NotNil(t, middleware)

	// Test with invalid args
	args = []string{"invalid=arg"}
	middleware = createMetricsMiddleware(args)
	assert.NotNil(t, middleware)
}

func TestGetMetricsInfo(t *testing.T) {

	config := &MetricsConfig{
		Enabled:   true,
		Endpoint:  "/metrics",
		Namespace: "test",
		Subsystem: "api",
	}

	info := GetMetricsInfo(config)
	assert.Equal(t, true, info.Enabled)
	assert.Equal(t, "/metrics", info.Endpoint)
	assert.Equal(t, "test", info.Namespace)
	assert.Equal(t, "api", info.Subsystem)
	assert.NotEmpty(t, info.Metrics)
}

func TestMetricsCollector_ConcurrentAccess(_ *testing.T) {
	// Reset metrics state before test
	metricsInitMutex.Lock()
	defaultMetricsCollector = nil
	metricsInitialized = false
	metricsInitMutex.Unlock()

	config := &MetricsConfig{
		Enabled:   true,
		Namespace: "test",
		Subsystem: "api",
	}

	// Initialize metrics
	InitMetrics(config)

	// Test concurrent access to metrics functions
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			RecordCacheHit("memory", "url")
			RecordCacheMiss("memory", "url")
			RecordRateLimitHit("/api/users", "ip")
			RecordValidationError("json", "email")
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMetricsMiddleware_WithRealRequest(t *testing.T) {
	// Reset metrics state before test
	metricsInitMutex.Lock()
	defaultMetricsCollector = nil
	metricsInitialized = false
	metricsInitMutex.Unlock()

	config := &MetricsConfig{
		Enabled:   true,
		Namespace: "test",
		Subsystem: "api",
	}

	middleware := MetricsMiddleware(config)

	// Create test server
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	// Make request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test")
}

func TestMetrics_Disabled(t *testing.T) {
	// Reset metrics state before test
	metricsInitMutex.Lock()
	defaultMetricsCollector = nil
	metricsInitialized = false
	metricsInitMutex.Unlock()

	config := &MetricsConfig{
		Enabled: false,
	}

	middleware := MetricsMiddleware(config)
	assert.NotNil(t, middleware)

	// Test middleware execution (should not panic)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	c.Request = req

	assert.NotPanics(t, func() {
		middleware(c)
	})
}
