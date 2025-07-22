// Tests for health checker logic in gin-decorators framework
package decorators

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHealthChecker(t *testing.T) {
	// Test with default timeout
	config := &ProxyConfig{
		Timeout: "",
	}
	hc := NewHealthChecker(config)
	assert.NotNil(t, hc)
	assert.Equal(t, 5*time.Second, hc.client.Timeout)
	// Test with custom timeout
	config = &ProxyConfig{
		Timeout: "10s",
	}
	hc = NewHealthChecker(config)
	assert.NotNil(t, hc)
	assert.Equal(t, 10*time.Second, hc.client.Timeout)
}

func TestHealthChecker_Check_NoHealthCheckConfigured(t *testing.T) {
	config := &ProxyConfig{
		HealthCheck: "",
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: "http://localhost:8080",
	}
	// Should return true when no health check is configured
	result := hc.Check(instance)
	assert.True(t, result)
}

func TestHealthChecker_Check_HealthyInstance(t *testing.T) {
	// Create test server that returns 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("healthy"))
	}))
	defer server.Close()
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "5s",
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: server.URL,
	}
	// Should return true for healthy instance
	result := hc.Check(instance)
	assert.True(t, result)
}

func TestHealthChecker_Check_UnhealthyInstance(t *testing.T) {
	// Create test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("unhealthy"))
	}))
	defer server.Close()
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "5s",
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: server.URL,
	}
	// Should return false for unhealthy instance
	result := hc.Check(instance)
	assert.False(t, result)
}

func TestHealthChecker_Check_UnreachableInstance_Original(t *testing.T) {
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "1s", // Short timeout for test
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: "http://unreachable-host:9999",
	}
	// Should return false for unreachable instance
	result := hc.Check(instance)
	assert.False(t, result)
}

func TestHealthChecker_Check_URLConstruction(t *testing.T) {
	// Create test server that checks the request path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the health check path is correct
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "5s",
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: server.URL,
	}
	// Should return true when health check path is correct
	result := hc.Check(instance)
	assert.True(t, result)
}

func TestHealthChecker_Check_URLWithTrailingSlash(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "5s",
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: server.URL + "/", // URL with trailing slash
	}
	// Should return true and handle trailing slash correctly
	result := hc.Check(instance)
	assert.True(t, result)
}

func TestHealthChecker_Check_HealthCheckPathWithoutSlash(t *testing.T) {
	// Create test server that checks the request path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the health check path is correct
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	config := &ProxyConfig{
		HealthCheck: "health", // Without leading slash
		Timeout:     "5s",
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: server.URL,
	}
	// Should return true when health check path is constructed correctly
	result := hc.Check(instance)
	assert.True(t, result)
}

func TestHealthChecker_Check_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, true},
		{201, true},
		{204, true},
		{299, true},
		{300, false},
		{400, false},
		{401, false},
		{500, false},
		{503, false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			// Create test server with specific status code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()
			config := &ProxyConfig{
				HealthCheck: "/health",
				Timeout:     "5s",
			}
			hc := NewHealthChecker(config)
			instance := &ProxyInstance{
				URL: server.URL,
			}
			result := hc.Check(instance)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHealthChecker_Check_Timeout(t *testing.T) {
	// Remove  to avoid race conditions
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "1s", // Short timeout
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: server.URL,
	}
	// Should return false due to timeout
	result := hc.Check(instance)
	assert.False(t, result)
}

func TestHealthChecker_Check_UnreachableInstance(t *testing.T) {
	// Remove  to avoid race conditions
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "1s",
	}
	hc := NewHealthChecker(config)
	instance := &ProxyInstance{
		URL: "http://localhost:99999",
	}
	// Should return false due to connection refused
	result := hc.Check(instance)
	assert.False(t, result)
}

func TestCreateHealthCheckerFromConfig(t *testing.T) {
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "10s",
	}
	hc := createHealthChecker(config)
	assert.NotNil(t, hc)
	// Test that it implements the HealthChecker interface
	_ = hc
}

func TestHealthChecker_Check_MultipleInstances(t *testing.T) {
	// Remove  to avoid race conditions
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	config := &ProxyConfig{
		HealthCheck: "/health",
		Timeout:     "5s",
	}
	hc := NewHealthChecker(config)
	// Test multiple instances
	instances := []*ProxyInstance{
		{URL: server.URL},
		{URL: server.URL},
		{URL: server.URL},
	}
	for _, instance := range instances {
		result := hc.Check(instance)
		assert.True(t, result)
	}
}
