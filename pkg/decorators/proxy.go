package decorators

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ProxyConfig configuration for proxy middleware
type ProxyConfig struct {
	// Service Discovery
	Target    string   `json:"target"`    // Direct URL
	Service   string   `json:"service"`   // Service name for discovery
	Discovery string   `json:"discovery"` // dns, consul, kubernetes, static
	Targets   []string `json:"targets"`   // List of URLs for static discovery

	// Load Balancing
	LoadBalancer   string `json:"load_balancer"`   // round_robin, least_connections, ip_hash, weighted
	HealthCheck    string `json:"health_check"`    // Health check endpoint
	HealthInterval string `json:"health_interval"` // Health check interval

	// Resilience
	Timeout      string `json:"timeout"`
	Retries      int    `json:"retries"`
	RetryBackoff string `json:"retry_backoff"` // linear, exponential
	RetryDelay   string `json:"retry_delay"`

	// Circuit Breaker
	CircuitBreaker   string `json:"circuit_breaker"`
	FailureThreshold int    `json:"failure_threshold"`

	// Advanced
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers"`
	Transform string            `json:"transform"` // Request/response transformation

	// Service Discovery specific
	ConsulAddress string `json:"consul_address"`
	K8sNamespace  string `json:"k8s_namespace"`
}

// ProxyInstance represents a service instance
type ProxyInstance struct {
	URL          string            `json:"url"`
	Weight       int               `json:"weight"`
	Healthy      bool              `json:"healthy"`
	LastCheck    time.Time         `json:"last_check"`
	ActiveConns  int               `json:"active_conns"`
	FailureCount int               `json:"failure_count"`
	Metadata     map[string]string `json:"metadata"`
	mu           sync.RWMutex
}

// ProxyManager manages proxy operations
type ProxyManager struct {
	instances      []*ProxyInstance
	loadBalancer   LoadBalancer
	circuitBreaker CircuitBreaker
	healthChecker  HealthChecker
	httpClient     *http.Client
	config         ProxyConfig
	mu             sync.RWMutex
}

// LoadBalancer interface for different load balancing algorithms
type LoadBalancer interface {
	Select(instances []*ProxyInstance, c *gin.Context) *ProxyInstance
}

// CircuitBreaker interface for circuit breaker pattern
type CircuitBreaker interface {
	IsOpen() bool
	RecordSuccess()
	RecordFailure()
	GetState() string
}

// HealthChecker interface for health checking
type HealthChecker interface {
	Check(instance *ProxyInstance) bool
}

// ServiceDiscovery interface for different discovery methods
type ServiceDiscovery interface {
	Discover(service string) ([]*ProxyInstance, error)
}

// Global proxy managers registry
var proxyManagers = make(map[string]*ProxyManager)
var proxyManagersMu sync.RWMutex

// Default configurations
const (
	DefaultTimeout          = "10s"
	DefaultRetries          = 3
	DefaultRetryDelay       = "1s"
	DefaultHealthInterval   = "30s"
	DefaultFailureThreshold = 5
	DefaultCircuitBreaker   = "30s"
)

// createProxyMiddleware creates proxy middleware with configuration
func createProxyMiddleware(args []string) gin.HandlerFunc {
	config := parseProxyConfig(args)
	manager := getOrCreateProxyManager(config)

	return func(c *gin.Context) {
		// 1. Intercept BEFORE (if handler wants)
		c.Next()

		// 2. If handler didn't abort, do proxy
		if !c.IsAborted() {
			manager.Forward(c, config)
		}

		// 3. Intercept AFTER (if handler wants)
		c.Next()
	}
}

// parseProxyConfig parses proxy configuration from arguments
func parseProxyConfig(args []string) ProxyConfig {
	config := ProxyConfig{
		Timeout:          DefaultTimeout,
		Retries:          DefaultRetries,
		RetryDelay:       DefaultRetryDelay,
		RetryBackoff:     "exponential",
		LoadBalancer:     "round_robin",
		HealthInterval:   DefaultHealthInterval,
		FailureThreshold: DefaultFailureThreshold,
		CircuitBreaker:   DefaultCircuitBreaker,
		Headers:          make(map[string]string),
	}

	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		switch key {
		case "target":
			config.Target = value
		case "service":
			config.Service = value
		case "discovery":
			config.Discovery = value
		case "targets":
			config.Targets = strings.Split(value, ",")
		case "load_balancer":
			config.LoadBalancer = value
		case "health_check":
			config.HealthCheck = value
		case "health_interval":
			config.HealthInterval = value
		case "timeout":
			config.Timeout = value
		case "retries":
			if retries, err := strconv.Atoi(value); err == nil {
				config.Retries = retries
			}
		case "retry_backoff":
			config.RetryBackoff = value
		case "retry_delay":
			config.RetryDelay = value
		case "circuit_breaker":
			config.CircuitBreaker = value
		case "failure_threshold":
			if threshold, err := strconv.Atoi(value); err == nil {
				config.FailureThreshold = threshold
			}
		case "path":
			config.Path = value
		case "transform":
			config.Transform = value
		case "consul_address":
			config.ConsulAddress = value
		case "k8s_namespace":
			config.K8sNamespace = value
		}
	}

	return config
}

// getOrCreateProxyManager gets or creates a proxy manager
func getOrCreateProxyManager(config ProxyConfig) *ProxyManager {
	// Create unique key for this configuration
	key := fmt.Sprintf("%s:%s:%s", config.Service, config.Target, config.Discovery)

	proxyManagersMu.RLock()
	if manager, exists := proxyManagers[key]; exists {
		proxyManagersMu.RUnlock()
		return manager
	}
	proxyManagersMu.RUnlock()

	// Create new manager
	proxyManagersMu.Lock()
	defer proxyManagersMu.Unlock()

	// Double-check after acquiring lock
	if manager, exists := proxyManagers[key]; exists {
		return manager
	}

	manager := NewProxyManager(config)
	proxyManagers[key] = manager

	// Start background tasks
	go manager.startHealthChecks()
	go manager.startServiceDiscovery()

	return manager
}

// NewProxyManager creates a new proxy manager
func NewProxyManager(config ProxyConfig) *ProxyManager {
	// Parse timeouts
	timeout, _ := time.ParseDuration(config.Timeout)
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // For development
			},
		},
	}

	manager := &ProxyManager{
		config:     config,
		httpClient: httpClient,
	}

	// Initialize load balancer
	manager.loadBalancer = createLoadBalancer(config.LoadBalancer)

	// Initialize circuit breaker
	manager.circuitBreaker = createCircuitBreaker(config)

	// Initialize health checker
	manager.healthChecker = createHealthChecker(config)

	// Initialize instances
	manager.initializeInstances()

	return manager
}

// initializeInstances initializes service instances
func (pm *ProxyManager) initializeInstances() {
	if pm.config.Target != "" {
		// Single target
		instance := &ProxyInstance{
			URL:       pm.config.Target,
			Weight:    1,
			Healthy:   true,
			LastCheck: time.Now(),
			Metadata:  make(map[string]string),
		}
		pm.instances = append(pm.instances, instance)
	} else if len(pm.config.Targets) > 0 {
		// Static targets
		for _, target := range pm.config.Targets {
			instance := &ProxyInstance{
				URL:       strings.TrimSpace(target),
				Weight:    1,
				Healthy:   true,
				LastCheck: time.Now(),
				Metadata:  make(map[string]string),
			}
			pm.instances = append(pm.instances, instance)
		}
	}
}

// Forward forwards the request to the selected instance
func (pm *ProxyManager) Forward(c *gin.Context, config ProxyConfig) {
	// Check circuit breaker
	if pm.circuitBreaker.IsOpen() {
		c.JSON(503, gin.H{"error": "Service temporarily unavailable"})
		c.Abort()
		return
	}

	// Select instance
	instance := pm.loadBalancer.Select(pm.instances, c)
	if instance == nil {
		c.JSON(503, gin.H{"error": "No healthy instances available"})
		c.Abort()
		return
	}

	// Build target URL
	targetURL := pm.buildTargetURL(instance, c)

	// Create request
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create request"})
		c.Abort()
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add custom headers
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	// Add proxy headers
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", c.Request.URL.Scheme)
	req.Header.Set("X-Forwarded-Host", c.Request.Host)

	// Execute request with retry logic
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= config.Retries; attempt++ {
		// Increment active connections
		instance.mu.Lock()
		instance.ActiveConns++
		instance.mu.Unlock()

		// Execute request
		resp, lastErr = pm.httpClient.Do(req)

		// Decrement active connections
		instance.mu.Lock()
		instance.ActiveConns--
		instance.mu.Unlock()

		if lastErr == nil && resp.StatusCode < 500 {
			// Success
			pm.circuitBreaker.RecordSuccess()
			break
		}

		// Failure
		pm.circuitBreaker.RecordFailure()
		instance.mu.Lock()
		instance.FailureCount++
		instance.mu.Unlock()

		if attempt < config.Retries {
			// Calculate delay
			delay := pm.calculateRetryDelay(attempt, config)
			time.Sleep(delay)
		}
	}

	if lastErr != nil {
		c.JSON(502, gin.H{"error": "Upstream service error"})
		c.Abort()
		return
	}

	// Copy response
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Add proxy headers
	c.Header("X-Proxy-Instance", instance.URL)
	c.Header("X-Proxy-Circuit-Breaker", pm.circuitBreaker.GetState())

	// Copy response body
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		c.JSON(502, gin.H{"error": "Failed to read response"})
		c.Abort()
		return
	}

	// Set response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// buildTargetURL builds the target URL for the request
func (pm *ProxyManager) buildTargetURL(instance *ProxyInstance, c *gin.Context) string {
	baseURL := instance.URL

	// If path is specified, use it
	if pm.config.Path != "" {
		path := pm.config.Path

		// Replace path parameters
		for _, param := range c.Params {
			path = strings.ReplaceAll(path, "{"+param.Key+"}", param.Value)
		}

		// Join with base URL
		if strings.HasSuffix(baseURL, "/") {
			baseURL = baseURL[:len(baseURL)-1]
		}
		if !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		baseURL += path
	} else {
		// Use original path
		baseURL += c.Request.URL.Path
	}

	// Add query parameters
	if c.Request.URL.RawQuery != "" {
		baseURL += "?" + c.Request.URL.RawQuery
	}

	return baseURL
}

// calculateRetryDelay calculates delay for retry attempts
func (pm *ProxyManager) calculateRetryDelay(attempt int, config ProxyConfig) time.Duration {
	baseDelay, _ := time.ParseDuration(config.RetryDelay)
	if baseDelay == 0 {
		baseDelay = time.Second
	}

	if config.RetryBackoff == "exponential" {
		return baseDelay * time.Duration(1<<attempt)
	}

	// Linear backoff
	return baseDelay * time.Duration(attempt+1)
}

// startHealthChecks starts background health checks
func (pm *ProxyManager) startHealthChecks() {
	if pm.config.HealthCheck == "" {
		return
	}

	interval, _ := time.ParseDuration(pm.config.HealthInterval)
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		pm.performHealthChecks()
	}
}

// performHealthChecks performs health checks on all instances
func (pm *ProxyManager) performHealthChecks() {
	pm.mu.RLock()
	instances := make([]*ProxyInstance, len(pm.instances))
	copy(instances, pm.instances)
	pm.mu.RUnlock()

	for _, instance := range instances {
		healthy := pm.healthChecker.Check(instance)

		instance.mu.Lock()
		instance.Healthy = healthy
		instance.LastCheck = time.Now()
		if healthy {
			instance.FailureCount = 0
		}
		instance.mu.Unlock()
	}
}

// startServiceDiscovery starts background service discovery
func (pm *ProxyManager) startServiceDiscovery() {
	if pm.config.Service == "" || pm.config.Discovery == "" {
		return
	}

	interval, _ := time.ParseDuration(pm.config.HealthInterval)
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		pm.performServiceDiscovery()
	}
}

// performServiceDiscovery performs service discovery
func (pm *ProxyManager) performServiceDiscovery() {
	var discovery ServiceDiscovery

	switch pm.config.Discovery {
	case "consul":
		discovery = NewConsulDiscovery(pm.config.ConsulAddress)
	case "dns":
		discovery = NewDNSDiscovery()
	case "kubernetes":
		discovery = NewK8sDiscovery(pm.config.K8sNamespace)
	default:
		return
	}

	instances, err := discovery.Discover(pm.config.Service)
	if err != nil {
		LogVerbose("Service discovery error: %v", err)
		return
	}

	pm.mu.Lock()
	pm.instances = instances
	pm.mu.Unlock()
}
