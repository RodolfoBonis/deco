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
	manager := &ProxyManager{
		config: ProxyConfig{
			Path: "/users/{id}",
		},
	}

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
