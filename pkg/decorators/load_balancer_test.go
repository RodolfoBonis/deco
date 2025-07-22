// Tests for load balancer logic in gin-decorators framework
package decorators

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRoundRobinLoadBalancer_Select(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()

	lb := &RoundRobinLoadBalancer{}

	// Create test instances
	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true},
		{URL: "http://instance2:8080", Healthy: true},
		{URL: "http://instance3:8080", Healthy: true},
	}

	// Create Gin context
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test round-robin selection
	selected1 := lb.Select(instances, c)
	selected2 := lb.Select(instances, c)
	selected3 := lb.Select(instances, c)
	selected4 := lb.Select(instances, c)

	// Should cycle through instances
	assert.NotNil(t, selected1)
	assert.NotNil(t, selected2)
	assert.NotNil(t, selected3)
	assert.NotNil(t, selected4)

	// Should be different instances in sequence
	assert.NotEqual(t, selected1.URL, selected2.URL)
	assert.NotEqual(t, selected2.URL, selected3.URL)
}

func TestRoundRobinLoadBalancer_Select_EmptyInstances(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	gin.SetMode(gin.TestMode)

	lb := &RoundRobinLoadBalancer{}
	instance := lb.Select([]*ProxyInstance{}, nil)
	assert.Nil(t, instance)
}

func TestRoundRobinLoadBalancer_Select_NoHealthyInstances(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	gin.SetMode(gin.TestMode)

	lb := &RoundRobinLoadBalancer{}
	instances := []*ProxyInstance{
		{URL: "http://unhealthy1:8080", Healthy: false},
		{URL: "http://unhealthy2:8080", Healthy: false},
	}
	instance := lb.Select(instances, nil)
	assert.Nil(t, instance)
}

func TestRoundRobinLoadBalancer_Select_MixedHealthyInstances(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	gin.SetMode(gin.TestMode)

	lb := &RoundRobinLoadBalancer{}
	instances := []*ProxyInstance{
		{URL: "http://healthy1:8080", Healthy: true},
		{URL: "http://unhealthy1:8080", Healthy: false},
		{URL: "http://healthy2:8080", Healthy: true},
	}

	// First selection
	instance1 := lb.Select(instances, nil)
	assert.NotNil(t, instance1)
	assert.True(t, instance1.Healthy)

	// Second selection
	instance2 := lb.Select(instances, nil)
	assert.NotNil(t, instance2)
	assert.True(t, instance2.Healthy)

	// Should be different instances
	assert.NotEqual(t, instance1.URL, instance2.URL)
}

func TestLeastConnectionsLoadBalancer_Select(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	gin.SetMode(gin.TestMode)

	lb := &LeastConnectionsLoadBalancer{}
	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true, ActiveConns: 5},
		{URL: "http://instance2:8080", Healthy: true, ActiveConns: 2},
		{URL: "http://instance3:8080", Healthy: true, ActiveConns: 8},
	}

	selected := lb.Select(instances, nil)
	assert.NotNil(t, selected)
	assert.Equal(t, "http://instance2:8080", selected.URL)
	assert.Equal(t, 2, selected.ActiveConns)
}

func TestLeastConnectionsLoadBalancer_Select_EmptyInstances(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	gin.SetMode(gin.TestMode)

	lb := &LeastConnectionsLoadBalancer{}
	selected := lb.Select([]*ProxyInstance{}, nil)
	assert.Nil(t, selected)
}

func TestLeastConnectionsLoadBalancer_Select_NoHealthyInstances(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	gin.SetMode(gin.TestMode)

	lb := &LeastConnectionsLoadBalancer{}
	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: false, ActiveConns: 1},
		{URL: "http://instance2:8080", Healthy: false, ActiveConns: 2},
	}

	selected := lb.Select(instances, nil)
	assert.Nil(t, selected)
}

func TestIPHashLoadBalancer_Select(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()
	gin.SetMode(gin.TestMode)

	lb := &IPHashLoadBalancer{}
	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true},
		{URL: "http://instance2:8080", Healthy: true},
		{URL: "http://instance3:8080", Healthy: true},
	}

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/", http.NoBody)
	c.Request.RemoteAddr = "192.168.1.100:12345"

	selected := lb.Select(instances, c)
	assert.NotNil(t, selected)
	assert.True(t, selected.Healthy)
}

func TestIPHashLoadBalancer_Select_DifferentIPs(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()

	lb := &IPHashLoadBalancer{}

	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true},
		{URL: "http://instance2:8080", Healthy: true},
	}

	// Remove gin.SetMode() to avoid race conditions

	// Test with different IPs
	ip1 := "192.168.1.100"
	ip2 := "192.168.1.101"

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/", http.NoBody)
	req1.RemoteAddr = ip1 + ":12345"
	c1, _ := gin.CreateTestContext(w1)
	c1.Request = req1

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/", http.NoBody)
	req2.RemoteAddr = ip2 + ":12345"
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = req2

	selected1 := lb.Select(instances, c1)
	selected2 := lb.Select(instances, c2)

	assert.NotNil(t, selected1)
	assert.NotNil(t, selected2)

	// Different IPs might select different instances (hash-based)
	// We can't guarantee they'll be different, but we can test that both are valid
	assert.Contains(t, []string{"http://instance1:8080", "http://instance2:8080"}, selected1.URL)
	assert.Contains(t, []string{"http://instance1:8080", "http://instance2:8080"}, selected2.URL)
}

func TestIPHashLoadBalancer_Select_EmptyInstances(t *testing.T) {
	// Remove  to avoid race conditions

	lb := &IPHashLoadBalancer{}
	instances := []*ProxyInstance{}

	// Remove gin.SetMode() to avoid race conditions
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	selected := lb.Select(instances, c)
	assert.Nil(t, selected)
}

func TestWeightedRoundRobinLoadBalancer_Select(t *testing.T) {
	// Remove  to avoid race conditions

	lb := &WeightedRoundRobinLoadBalancer{}

	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true, Weight: 3},
		{URL: "http://instance2:8080", Healthy: true, Weight: 1},
		{URL: "http://instance3:8080", Healthy: true, Weight: 2},
	}

	// Remove gin.SetMode() to avoid race conditions
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test multiple selections to see weighted distribution
	selections := make(map[string]int)
	for i := 0; i < 12; i++ { // Test multiple rounds
		selected := lb.Select(instances, c)
		assert.NotNil(t, selected)
		selections[selected.URL]++
	}

	// Should have some distribution based on weights
	assert.Greater(t, selections["http://instance1:8080"], 0) // Weight 3
	assert.Greater(t, selections["http://instance2:8080"], 0) // Weight 1
	assert.Greater(t, selections["http://instance3:8080"], 0) // Weight 2
}

func TestWeightedRoundRobinLoadBalancer_Select_EmptyInstances(t *testing.T) {
	// Remove  to avoid race conditions

	lb := &WeightedRoundRobinLoadBalancer{}
	instances := []*ProxyInstance{}

	// Remove gin.SetMode() to avoid race conditions
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	selected := lb.Select(instances, c)
	assert.Nil(t, selected)
}

func TestWeightedRoundRobinLoadBalancer_Select_NoHealthyInstances(t *testing.T) {
	// Remove  to avoid race conditions

	lb := &WeightedRoundRobinLoadBalancer{}

	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: false, Weight: 3},
		{URL: "http://instance2:8080", Healthy: false, Weight: 1},
	}

	// Remove gin.SetMode() to avoid race conditions
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	selected := lb.Select(instances, c)
	assert.Nil(t, selected)
}

func TestWeightedRoundRobinLoadBalancer_Select_ZeroWeight(t *testing.T) {
	// Remove  to avoid race conditions

	lb := &WeightedRoundRobinLoadBalancer{}

	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true, Weight: 0},
		{URL: "http://instance2:8080", Healthy: true, Weight: 0},
	}

	// Remove gin.SetMode() to avoid race conditions
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	selected := lb.Select(instances, c)
	assert.NotNil(t, selected) // Should fallback to first instance
}

func TestCreateLoadBalancerFromAlgorithm(t *testing.T) {
	// Remove  to avoid race conditions

	tests := []struct {
		algorithm string
		expected  string
	}{
		{"round_robin", "*decorators.RoundRobinLoadBalancer"},
		{"least_connections", "*decorators.LeastConnectionsLoadBalancer"},
		{"ip_hash", "*decorators.IPHashLoadBalancer"},
		{"weighted", "*decorators.WeightedRoundRobinLoadBalancer"},
		{"unknown", "*decorators.RoundRobinLoadBalancer"}, // Default
		{"", "*decorators.RoundRobinLoadBalancer"},        // Empty string
	}

	for _, tt := range tests {
		t.Run(tt.algorithm, func(t *testing.T) {
			lb := createLoadBalancer(tt.algorithm)
			assert.NotNil(t, lb)
			assert.Equal(t, tt.expected, fmt.Sprintf("%T", lb))
		})
	}
}

func TestLoadBalancer_ConcurrentAccess(t *testing.T) {
	// Remove  to avoid race conditions

	lb := &RoundRobinLoadBalancer{}

	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true},
		{URL: "http://instance2:8080", Healthy: true},
		{URL: "http://instance3:8080", Healthy: true},
	}

	// Remove gin.SetMode() to avoid race conditions
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test concurrent access
	var wg sync.WaitGroup
	results := make(chan *ProxyInstance, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			selected := lb.Select(instances, c)
			results <- selected
		}()
	}

	wg.Wait()
	close(results)

	// Check that all selections returned valid instances
	for selected := range results {
		assert.NotNil(t, selected)
		assert.True(t, selected.Healthy)
		assert.Contains(t, []string{
			"http://instance1:8080",
			"http://instance2:8080",
			"http://instance3:8080",
		}, selected.URL)
	}
}

func TestLoadBalancer_Interface(t *testing.T) {
	// Remove  to avoid race conditions

	// Test that all load balancers implement the LoadBalancer interface
	loadBalancers := []LoadBalancer{
		&RoundRobinLoadBalancer{},
		&LeastConnectionsLoadBalancer{},
		&IPHashLoadBalancer{},
		&WeightedRoundRobinLoadBalancer{},
	}

	instances := []*ProxyInstance{
		{URL: "http://instance1:8080", Healthy: true},
	}

	// Remove gin.SetMode() to avoid race conditions
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345" // Set RemoteAddr to avoid panic
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	for _, lb := range loadBalancers {
		selected := lb.Select(instances, c)
		assert.NotNil(t, selected)
	}
}
