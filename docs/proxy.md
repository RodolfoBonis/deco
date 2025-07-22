# Proxy Guide

## Overview

The `@Proxy()` decorator enables **API Gateway** and **Service Discovery** functionality with automatic service management, load balancing, circuit breakers, and resilience.

## Basic Usage

### Simple Proxy

```go
// @Route("GET", "/api/users/:id")
// @Proxy(service="user-service")
func GetUser(c *gin.Context) {
    // Proxy handles forwarding automatically
}
```

### Advanced Configuration

```go
// @Route("GET", "/api/products")
// @Proxy(
//     service="product-service",
//     discovery="consul",
//     load_balancer="round_robin",
//     health_interval="30s",
//     circuit_breaker="30s"
// )
func GetProducts(c *gin.Context) {
    // Advanced proxy with service discovery and load balancing
}
```

## Service Discovery

### Consul Service Discovery

```go
// @Route("GET", "/api/notifications")
// @Proxy(
//     service="notification-service",
//     discovery="consul",
//     consul_address="localhost:8500",
//     load_balancer="round_robin"
// )
func GetNotifications(c *gin.Context) {}
```

### DNS Service Discovery

```go
// @Route("GET", "/api/inventory")
// @Proxy(
//     service="inventory-service.default.svc.cluster.local",
//     discovery="dns",
//     load_balancer="ip_hash"
// )
func GetInventory(c *gin.Context) {}
```

### Kubernetes Service Discovery

```go
// @Route("GET", "/api/payments")
// @Proxy(
//     service="payment-service",
//     discovery="kubernetes",
//     k8s_namespace="production",
//     load_balancer="weighted"
// )
func ProcessPayment(c *gin.Context) {}
```

### Static Targets

```go
// @Route("GET", "/api/reviews")
// @Proxy(
//     discovery="static",
//     instances="http://review-1:8083,http://review-2:8083,http://review-3:8083",
//     load_balancer="least_connections"
// )
func GetReviews(c *gin.Context) {}
```

## Load Balancing

### Round Robin

```go
// @Proxy(load_balancer="round_robin")
```
Distributes requests sequentially between instances.

### Least Connections

```go
// @Proxy(load_balancer="least_connections")
```
Sends requests to the instance with the fewest active connections.

### IP Hash

```go
// @Proxy(load_balancer="ip_hash")
```
Distributes based on client IP hash (consistent).

### Weighted Round Robin

```go
// @Proxy(load_balancer="weighted")
```
Round robin with configurable weights per instance.

## Resilience Patterns

### Circuit Breaker

```go
// @Proxy(
//     circuit_breaker="30s",      // Recovery time
//     failure_threshold=5         // Failures before opening
// )
```

The circuit breaker has three states:
1. **Closed**: Normal operation
2. **Open**: Blocks requests after failure threshold
3. **Half-Open**: Allows one test request

### Retry Logic

```go
// @Proxy(
//     retries=3,
//     retry_backoff="exponential",
//     retry_delay="1s"
// )
// Delays: 1s, 2s, 4s
```

### Health Checks

```go
// @Proxy(
//     health_interval="30s",
//     health_check="/health"
// )
```

## Configuration Parameters

### Service Discovery
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `service` | string | Service name for discovery | `"user-service"` |
| `discovery` | string | Discovery method | `"consul"`, `"dns"`, `"kubernetes"` |
| `instances` | string | Static instances list | `"http://s1:8080,http://s2:8080"` |
| `consul_address` | string | Consul address | `"localhost:8500"` |
| `k8s_namespace` | string | Kubernetes namespace | `"production"` |

### Load Balancing
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `load_balancer` | string | Load balancing algorithm | `"round_robin"`, `"least_connections"` |
| `health_interval` | string | Health check interval | `"30s"` |

### Resilience
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `timeout` | string | Request timeout | `"10s"` |
| `retries` | int | Number of retries | `3` |
| `retry_backoff` | string | Backoff type | `"exponential"`, `"linear"` |
| `circuit_breaker` | string | Circuit breaker timeout | `"30s"` |
| `failure_threshold` | int | Failure threshold | `5` |

## Examples

### Basic API Gateway

```go
// @Route("GET", "/api/user/:id")
// @Proxy(service="user-service")
func GetUserProxy(c *gin.Context) {
    // Forwards to user-service
}
```

### Microservice Integration

```go
// @Route("POST", "/api/orders")
// @Proxy(
//     service="order-service",
//     discovery="consul",
//     timeout="15s",
//     retries=3,
//     circuit_breaker="30s"
// )
func CreateOrder(c *gin.Context) {
    // Forwards to order-service with resilience
}
```

### Legacy System Integration

```go
// @Route("GET", "/api/legacy/users")
// @Proxy(
//     discovery="static",
//     instances="http://legacy-1:8080,http://legacy-2:8080",
//     load_balancer="weighted",
//     weights="3,1"
// )
func LegacyUsers(c *gin.Context) {
    // Forwards to legacy systems with weighted load balancing
}
```

## Best Practices

### 1. Service Discovery

```go
// Use appropriate discovery method
// @Proxy(discovery="consul")     // For microservices
// @Proxy(discovery="kubernetes") // For K8s environments
// @Proxy(discovery="static")     // For legacy systems
```

### 2. Load Balancing

```go
// Choose based on requirements
// @Proxy(load_balancer="round_robin")      // General purpose
// @Proxy(load_balancer="least_connections") // For long-running requests
// @Proxy(load_balancer="ip_hash")          // For session affinity
```

### 3. Resilience

```go
// Always configure timeouts and retries
// @Proxy(
//     timeout="10s",
//     retries=3,
//     circuit_breaker="30s"
// )
```

## Next Steps

- **[Usage Guide](./usage.md)** - General usage guide
- **[API Reference](./api.md)** - Complete API documentation
- **[Examples](./examples.md)** - Proxy examples 