package decorators

import (
	"crypto/sha256"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

// RoundRobinLoadBalancer implements round-robin load balancing
type RoundRobinLoadBalancer struct {
	current uint64
}

// Select selects the next instance in round-robin fashion
func (lb *RoundRobinLoadBalancer) Select(instances []*ProxyInstance, _ *gin.Context) *ProxyInstance {
	if len(instances) == 0 {
		return nil
	}

	// Filter healthy instances
	var healthyInstances []*ProxyInstance
	for _, instance := range instances {
		instance.mu.RLock()
		if instance.Healthy {
			healthyInstances = append(healthyInstances, instance)
		}
		instance.mu.RUnlock()
	}

	if len(healthyInstances) == 0 {
		return nil
	}

	// Get next index
	next := atomic.AddUint64(&lb.current, 1)
	// Safe conversion: len() returns int, which is always positive and small
	// This conversion is safe because len() is always >= 0 and typically small
	instanceCount := len(healthyInstances)
	if instanceCount == 0 {
		return nil
	}
	// Safe conversion: instanceCount is always positive and small
	index := int(next % uint64(instanceCount)) // nolint:gosec // Safe: instanceCount is small

	return healthyInstances[index]
}

// LeastConnectionsLoadBalancer implements least connections load balancing
type LeastConnectionsLoadBalancer struct{}

// Select selects the instance with the least active connections
func (lb *LeastConnectionsLoadBalancer) Select(instances []*ProxyInstance, _ *gin.Context) *ProxyInstance {
	if len(instances) == 0 {
		return nil
	}

	var selected *ProxyInstance
	minConns := int(^uint(0) >> 1) // Max int

	for _, instance := range instances {
		instance.mu.RLock()
		if instance.Healthy && instance.ActiveConns < minConns {
			minConns = instance.ActiveConns
			selected = instance
		}
		instance.mu.RUnlock()
	}

	return selected
}

// IPHashLoadBalancer implements IP hash load balancing
type IPHashLoadBalancer struct{}

// Select selects instance based on client IP hash
func (lb *IPHashLoadBalancer) Select(instances []*ProxyInstance, c *gin.Context) *ProxyInstance {
	if len(instances) == 0 {
		return nil
	}

	// Filter healthy instances
	var healthyInstances []*ProxyInstance
	for _, instance := range instances {
		instance.mu.RLock()
		if instance.Healthy {
			healthyInstances = append(healthyInstances, instance)
		}
		instance.mu.RUnlock()
	}

	if len(healthyInstances) == 0 {
		return nil
	}

	// Hash client IP using SHA-256
	clientIP := c.ClientIP()
	hash := sha256.Sum256([]byte(clientIP))
	// Use first 8 bytes of hash for consistency
	hashValue := uint64(hash[0])<<56 | uint64(hash[1])<<48 | uint64(hash[2])<<40 | uint64(hash[3])<<32 |
		uint64(hash[4])<<24 | uint64(hash[5])<<16 | uint64(hash[6])<<8 | uint64(hash[7])

	// Safe conversion: len() returns int, which is always positive and small
	instanceCount := len(healthyInstances)
	if instanceCount == 0 {
		return nil
	}
	// Safe conversion: instanceCount is always positive and small
	index := int(hashValue % uint64(instanceCount)) // nolint:gosec // Safe: instanceCount is small
	return healthyInstances[index]
}

// WeightedRoundRobinLoadBalancer implements weighted round-robin load balancing
type WeightedRoundRobinLoadBalancer struct {
	current uint64
}

// Select selects instance based on weighted round-robin
func (lb *WeightedRoundRobinLoadBalancer) Select(instances []*ProxyInstance, _ *gin.Context) *ProxyInstance {
	if len(instances) == 0 {
		return nil
	}

	// Filter healthy instances
	var healthyInstances []*ProxyInstance
	totalWeight := 0

	for _, instance := range instances {
		instance.mu.RLock()
		if instance.Healthy {
			healthyInstances = append(healthyInstances, instance)
			totalWeight += instance.Weight
		}
		instance.mu.RUnlock()
	}

	if len(healthyInstances) == 0 {
		return nil
	}

	// Get next index
	next := atomic.AddUint64(&lb.current, 1)
	// Safe conversion: totalWeight is always positive and typically small
	if totalWeight <= 0 {
		return healthyInstances[0]
	}
	// Safe conversion: totalWeight is always positive and small
	weightedIndex := int(next % uint64(totalWeight)) // nolint:gosec // Safe: totalWeight is small

	// Find instance based on weight
	currentWeight := 0
	for _, instance := range healthyInstances {
		currentWeight += instance.Weight
		if weightedIndex < currentWeight {
			return instance
		}
	}

	// Fallback to first instance
	return healthyInstances[0]
}

// createLoadBalancer creates a load balancer based on the algorithm name
func createLoadBalancer(algorithm string) LoadBalancer {
	switch algorithm {
	case "round_robin":
		return &RoundRobinLoadBalancer{}
	case "least_connections":
		return &LeastConnectionsLoadBalancer{}
	case "ip_hash":
		return &IPHashLoadBalancer{}
	case "weighted":
		return &WeightedRoundRobinLoadBalancer{}
	default:
		// Default to round-robin
		return &RoundRobinLoadBalancer{}
	}
}
