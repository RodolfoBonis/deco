package decorators

import (
	"net/http"
	"strings"
	"time"
)

// HealthCheckerImpl implements health checking for service instances
type HealthCheckerImpl struct {
	config ProxyConfig
	client *http.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(config *ProxyConfig) *HealthCheckerImpl {
	timeout, _ := time.ParseDuration(config.Timeout)
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	return &HealthCheckerImpl{
		config: *config,
		client: client,
	}
}

// Check performs a health check on the given instance
func (hc *HealthCheckerImpl) Check(instance *ProxyInstance) bool {
	if hc.config.HealthCheck == "" {
		// No health check configured, assume healthy
		return true
	}

	// Build health check URL
	healthURL := instance.URL
	if !strings.HasSuffix(healthURL, "/") {
		healthURL += "/"
	}
	healthURL += strings.TrimPrefix(hc.config.HealthCheck, "/")

	// Perform HTTP health check
	resp, err := hc.client.Get(healthURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Consider healthy if status code is 2xx
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// createHealthChecker creates a health checker from configuration
func createHealthChecker(config *ProxyConfig) HealthChecker {
	return NewHealthChecker(config)
}
