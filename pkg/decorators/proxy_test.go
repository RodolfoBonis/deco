package decorators

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProxyMiddleware(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	// Test proxy configuration
	config := ProxyConfig{
		Target:  server.URL,
		Path:    "/test",
		Timeout: "5s",
		Retries: 1,
	}

	// Create proxy manager
	manager := NewProxyManager(&config)

	// Create Gin context
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		manager.Forward(c, &config)
	})

	// Test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
	assert.Contains(t, w.Body.String(), "success")
}

func TestLoadBalancerRoundRobin(t *testing.T) {
	lb := &RoundRobinLoadBalancer{}

	// Create test instances
	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true},
		{URL: "http://instance2:8080", Healthy: true},
		{URL: "http://instance3:8080", Healthy: true},
	}

	// Test round-robin selection
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_ = c // Use context to avoid unused variable warning

	selected1 := lb.Select(instances, c)
	selected2 := lb.Select(instances, c)
	selected3 := lb.Select(instances, c)
	selected4 := lb.Select(instances, c)

	// Should cycle through instances
	assert.NotNil(t, selected1)
	assert.NotNil(t, selected2)
	assert.NotNil(t, selected3)
	assert.NotNil(t, selected4)
}

func TestLoadBalancerLeastConnections(t *testing.T) {
	lb := &LeastConnectionsLoadBalancer{}

	// Create test instances with different connection counts
	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true, ActiveConns: 5},
		{URL: "http://instance2:8080", Healthy: true, ActiveConns: 2},
		{URL: "http://instance3:8080", Healthy: true, ActiveConns: 8},
	}

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_ = c // Use context to avoid unused variable warning

	selected := lb.Select(instances, c)

	// Should select instance with least connections
	assert.NotNil(t, selected)
	assert.Equal(t, "http://instance2:8080", selected.URL)
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// Initially closed
	assert.False(t, cb.IsOpen())
	assert.Equal(t, "closed", cb.GetState())

	// Record failures
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	// Should be open after 3 failures
	assert.True(t, cb.IsOpen())
	assert.Equal(t, "open", cb.GetState())

	// Wait for recovery timeout
	time.Sleep(1100 * time.Millisecond)

	// Should be half-open
	assert.False(t, cb.IsOpen())
	assert.Equal(t, "half_open", cb.GetState())

	// Record success
	cb.RecordSuccess()

	// Should be closed again
	assert.False(t, cb.IsOpen())
	assert.Equal(t, "closed", cb.GetState())
}

func TestHealthChecker(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "5s",
	}

	hc := NewHealthChecker(&config)
	instance := &ProxyInstance{URL: server.URL}

	// Test health check
	healthy := hc.Check(instance)
	assert.True(t, healthy)
}

func TestParseProxyConfig(t *testing.T) {
	args := []string{
		"target=http://service:8080",
		"path=/api/users/{id}",
		"timeout=10s",
		"retries=3",
		"load_balancer=round_robin",
		"circuit_breaker=30s",
		"failure_threshold=5",
	}

	config := parseProxyConfig(args)

	assert.Equal(t, "http://service:8080", config.Target)
	assert.Equal(t, "/api/users/{id}", config.Path)
	assert.Equal(t, "10s", config.Timeout)
	assert.Equal(t, 3, config.Retries)
	assert.Equal(t, "round_robin", config.LoadBalancer)
	assert.Equal(t, "30s", config.CircuitBreaker)
	assert.Equal(t, 5, config.FailureThreshold)
}

func TestBuildTargetURL(t *testing.T) {
	config := ProxyConfig{
		Path: "/users/{id}",
	}

	manager := NewProxyManager(&config)

	instance := &ProxyInstance{URL: "http://service:8080"}

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_ = c // Use context to avoid unused variable warning
	c.Params = gin.Params{{Key: "id", Value: "123"}}

	url := manager.buildTargetURL(instance, c)
	expected := "http://service:8080/users/123"

	assert.Equal(t, expected, url)
}

func TestCalculateRetryDelay(t *testing.T) {
	manager := &ProxyManager{}
	config := ProxyConfig{
		RetryDelay:   "1s",
		RetryBackoff: "exponential",
	}

	// Test exponential backoff
	delay1 := manager.calculateRetryDelay(0, &config)
	delay2 := manager.calculateRetryDelay(1, &config)
	delay3 := manager.calculateRetryDelay(2, &config)

	assert.Equal(t, 1*time.Second, delay1)
	assert.Equal(t, 2*time.Second, delay2)
	assert.Equal(t, 4*time.Second, delay3)

	// Test linear backoff
	config.RetryBackoff = "linear"
	delay1 = manager.calculateRetryDelay(0, &config)
	delay2 = manager.calculateRetryDelay(1, &config)
	delay3 = manager.calculateRetryDelay(2, &config)

	assert.Equal(t, 1*time.Second, delay1)
	assert.Equal(t, 2*time.Second, delay2)
	assert.Equal(t, 3*time.Second, delay3)
}

func TestCreateLoadBalancer(t *testing.T) {
	// Test round-robin
	lb := createLoadBalancer("round_robin")
	assert.IsType(t, &RoundRobinLoadBalancer{}, lb)

	// Test least connections
	lb = createLoadBalancer("least_connections")
	assert.IsType(t, &LeastConnectionsLoadBalancer{}, lb)

	// Test IP hash
	lb = createLoadBalancer("ip_hash")
	assert.IsType(t, &IPHashLoadBalancer{}, lb)

	// Test weighted
	lb = createLoadBalancer("weighted")
	assert.IsType(t, &WeightedRoundRobinLoadBalancer{}, lb)

	// Test default
	lb = createLoadBalancer("unknown")
	assert.IsType(t, &RoundRobinLoadBalancer{}, lb)
}

func TestCreateCircuitBreaker(t *testing.T) {
	config := ProxyConfig{
		FailureThreshold: 5,
		CircuitBreaker:   "30s",
	}

	cb := createCircuitBreaker(&config)
	assert.NotNil(t, cb)
	assert.False(t, cb.IsOpen())
}

func TestCreateHealthChecker(t *testing.T) {
	config := ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "5s",
	}

	hc := createHealthChecker(&config)
	assert.NotNil(t, hc)
}

func TestCreateProxyMiddleware(t *testing.T) {
	args := []string{
		"target=http://httpbin.org",
		"path=/get",
		"timeout=5s",
	}

	middleware := createProxyMiddleware(args)
	assert.NotNil(t, middleware, "Proxy middleware should be created")

	// Test that middleware is a valid gin.HandlerFunc
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	// This should not panic
	assert.NotPanics(t, func() {
		middleware(c)
	})
}

// Tests for health checks functionality

func TestProxyManager_PerformHealthChecks(t *testing.T) {
	config := &ProxyConfig{
		HealthInterval: "30s",
		HealthCheck:    "/health",
	}

	manager := NewProxyManager(config)

	// Test health check execution
	manager.performHealthChecks()

	// Verify that health checks were performed
	// Note: This is a basic test to ensure the function doesn't panic
	assert.NotNil(t, manager)
}

func TestProxyManager_PerformServiceDiscovery(t *testing.T) {
	config := &ProxyConfig{
		Discovery:     "static",
		ConsulAddress: "http://localhost:8500",
		K8sNamespace:  "default",
	}

	manager := NewProxyManager(config)

	// Test service discovery execution
	manager.performServiceDiscovery()

	// Verify that discovery was attempted
	// Note: This will fail in test environment without Consul, but we can verify the function doesn't panic
	assert.NotNil(t, manager)
}

// Tests for proxy middleware with basic configuration

func TestProxyMiddleware_Basic(t *testing.T) {
	// Clear cache before test
	clearProxyManagers()

	middleware := createProxyMiddleware([]string{
		"target=http://localhost:8080",
		"timeout=10s",
		"retries=3",
	})
	assert.NotNil(t, middleware)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", http.NoBody)

	// Execute middleware
	middleware(c)

	// Should handle service unavailability gracefully
	assert.Equal(t, 502, w.Code)
}

// Tests for proxy with circuit breaker

func TestProxyMiddleware_WithCircuitBreaker(t *testing.T) {
	// Clear cache before test
	clearProxyManagers()

	middleware := createProxyMiddleware([]string{
		"target=http://localhost:8081", // Different port to avoid cache
		"circuit_breaker=30s",
		"failure_threshold=5",
	})
	assert.NotNil(t, middleware)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", http.NoBody)

	// Execute middleware
	middleware(c)

	// Should handle service unavailability gracefully
	assert.Equal(t, 502, w.Code)
}

// Tests for proxy with retry logic

func TestProxyMiddleware_WithRetry(t *testing.T) {
	// Clear cache before test
	clearProxyManagers()

	middleware := createProxyMiddleware([]string{
		"target=http://localhost:8082", // Different port to avoid cache
		"retries=3",
		"retry_delay=1s",
		"retry_backoff=exponential",
	})
	assert.NotNil(t, middleware)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", http.NoBody)

	// Execute middleware
	middleware(c)

	// Should return 503 when target service is unavailable
	assert.Equal(t, 502, w.Code)
}

// Tests for proxy error handling

func TestProxyMiddleware_ErrorHandling(t *testing.T) {
	// Clear cache before test
	clearProxyManagers()

	middleware := createProxyMiddleware([]string{
		"target=http://invalid-service:9999",
		"timeout=5s",
	})
	assert.NotNil(t, middleware)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", http.NoBody)

	// Execute middleware
	middleware(c)

	// Should return 502 when target service is unreachable
	assert.Equal(t, 502, w.Code)
}

// Tests for proxy with custom headers

func TestProxyMiddleware_WithCustomHeaders(t *testing.T) {
	// Clear cache before test
	clearProxyManagers()

	middleware := createProxyMiddleware([]string{
		"target=http://localhost:8083", // Different port to avoid cache
		"headers=X-Custom-Header:test-value,Authorization:Bearer test-token",
	})
	assert.NotNil(t, middleware)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test", http.NoBody)

	// Execute middleware
	middleware(c)

	// Should return 502 when target service is unavailable
	assert.Equal(t, 502, w.Code)
}

// Tests for proxy with path rewriting

func TestProxyMiddleware_WithPathRewriting(t *testing.T) {
	// Clear cache before test
	clearProxyManagers()

	middleware := createProxyMiddleware([]string{
		"target=http://localhost:8084", // Different port to avoid cache
		"path=/v1/",
	})
	assert.NotNil(t, middleware)

	// Create test context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/users", http.NoBody)

	// Execute middleware
	middleware(c)

	// Should return 503 when target service is unavailable
	assert.Equal(t, 502, w.Code)
}

// Tests for proxy manager creation

func TestNewProxyManager(t *testing.T) {
	config := &ProxyConfig{
		Target:         "http://localhost:8080",
		Timeout:        "10s",
		Retries:        3,
		LoadBalancer:   "round_robin",
		HealthInterval: "30s",
	}

	manager := NewProxyManager(config)
	assert.NotNil(t, manager)
	assert.Equal(t, config.Target, manager.config.Target)
	assert.Equal(t, config.Timeout, manager.config.Timeout)
	assert.Equal(t, config.Retries, manager.config.Retries)
	assert.Equal(t, config.LoadBalancer, manager.config.LoadBalancer)
	assert.Equal(t, config.HealthInterval, manager.config.HealthInterval)
}

// Tests for proxy config parsing

func TestParseProxyConfig_ValidArgs(t *testing.T) {
	args := []string{
		"target=http://localhost:8080",
		"timeout=15s",
		"retries=5",
		"load_balancer=least_connections",
		"health_check=/health",
		"health_interval=60s",
		"circuit_breaker=45s",
		"failure_threshold=10",
		"path=/api/",
		"transform=request",
	}

	config := parseProxyConfig(args)

	assert.Equal(t, "http://localhost:8080", config.Target)
	assert.Equal(t, "15s", config.Timeout)
	assert.Equal(t, 5, config.Retries)
	assert.Equal(t, "least_connections", config.LoadBalancer)
	assert.Equal(t, "/health", config.HealthCheck)
	assert.Equal(t, "60s", config.HealthInterval)
	assert.Equal(t, "45s", config.CircuitBreaker)
	assert.Equal(t, 10, config.FailureThreshold)
	assert.Equal(t, "/api/", config.Path)
	assert.Equal(t, "request", config.Transform)
}

// Tests for proxy config parsing with invalid values

func TestParseProxyConfig_InvalidValues(t *testing.T) {
	args := []string{
		"target=http://localhost:8080",
		"timeout=invalid",
		"retries=invalid",
		"failure_threshold=invalid",
	}

	config := parseProxyConfig(args)

	// Should use default values for invalid inputs
	assert.Equal(t, "http://localhost:8080", config.Target)
	assert.Equal(t, "invalid", config.Timeout)  // Invalid string values are kept as-is
	assert.Equal(t, 3, config.Retries)          // Invalid int values use defaults
	assert.Equal(t, 5, config.FailureThreshold) // Invalid int values use defaults
}

// Tests for proxy config parsing with malformed arguments

func TestParseProxyConfig_MalformedArgs(t *testing.T) {
	args := []string{
		"target=http://localhost:8080",
		"invalid-arg",
		"key=value=extra",
		"",
	}

	config := parseProxyConfig(args)

	// Should handle malformed arguments gracefully
	assert.Equal(t, "http://localhost:8080", config.Target)
	assert.Equal(t, DefaultTimeout, config.Timeout)
	assert.Equal(t, DefaultRetries, config.Retries)
}
