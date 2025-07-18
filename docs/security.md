# Security Guide

## Overview

The deco framework provides multiple layers of security:

1. **Automatic Internal Protection** - Protects framework endpoints
2. **Network-Based Access Control** - IP and network restrictions
3. **Application-Level Security** - Endpoint-specific protection

## Automatic Internal Protection

### Default Protection

When using `deco.Default()`, internal endpoints are automatically protected:

```go
package main

import deco "github.com/RodolfoBonis/deco"

func main() {
    // Automatic localhost-only protection
    r := deco.Default()
    r.Run(":8080")
}
```

**Protected endpoints**:
- `/decorators/docs` - Route documentation
- `/decorators/docs.json` - Documentation JSON
- `/decorators/openapi.json` - OpenAPI specification
- `/decorators/openapi.yaml` - OpenAPI YAML
- `/decorators/swagger-ui` - Swagger interface
- `/decorators/swagger` - Swagger redirect

### Custom Security Configuration

```go
package main

import deco "github.com/RodolfoBonis/deco"

func main() {
    // Custom security configuration
    securityConfig := &deco.SecurityConfig{
        AllowPrivateNetworks: true,  // VPN, Docker, etc.
        AllowLocalhost: true,        // Development
        AllowedNetworks: []string{"10.0.0.0/8", "172.16.0.0/12"},
        AllowedIPs: []string{"192.168.1.100", "192.168.1.101"},
        ErrorMessage: "Access denied: Internal endpoints restricted",
        LogBlockedAttempts: true,
    }
    
    r := deco.DefaultWithSecurity(securityConfig)
    r.Run(":8080")
}
```

## Security Configuration

### SecurityConfig Structure

```go
type SecurityConfig struct {
    AllowedNetworks     []string `yaml:"allowed_networks"`     // Networks in CIDR notation
    AllowedIPs          []string `yaml:"allowed_ips"`          // Individual IP addresses
    AllowedHosts        []string `yaml:"allowed_hosts"`        // Hostnames/domains
    AllowLocalhost      bool     `yaml:"allow_localhost"`      // Allow localhost/127.0.0.1
    AllowPrivateNetworks bool    `yaml:"allow_private_networks"` // Allow private networks
    ErrorMessage        string   `yaml:"error_message"`        // Custom error message
    LogBlockedAttempts  bool     `yaml:"log_blocked_attempts"` // Log blocked attempts
}
```

### Configuration Examples

#### Development Environment

```go
devSecurity := &deco.SecurityConfig{
    AllowPrivateNetworks: true,  // VPN, Docker networks
    AllowLocalhost: true,        // Local development
    ErrorMessage: "Access denied: Development environment only",
    LogBlockedAttempts: true,
}
```

#### Production Environment

```go
prodSecurity := &deco.SecurityConfig{
    AllowedNetworks: []string{
        "10.0.0.0/8",      // Company network
        "172.16.0.0/12",   // Docker networks
    },
    AllowLocalhost: true,  // For debugging
    AllowedIPs: []string{
        "192.168.1.100",   // Monitoring server
        "192.168.1.101",   // Backup server
    },
    ErrorMessage: "Access denied: Corporate network only",
    LogBlockedAttempts: true,
}
```

## Application-Level Security

### @Security Decorator

Use the `@Security` decorator to protect specific application endpoints:

```go
// handlers/admin.go
package handlers

import "github.com/gin-gonic/gin"

// @Route("GET", "/admin/dashboard")
// @Security(private)
// @Auth(role="admin")
func AdminDashboard(c *gin.Context) {
    // Only accessible from private networks
    c.JSON(200, gin.H{"dashboard": "Admin panel"})
}

// @Route("GET", "/admin/users")
// @Security(networks="192.168.1.0/24")
// @Auth(role="admin")
func AdminUsers(c *gin.Context) {
    // Only accessible from specific network
    c.JSON(200, gin.H{"users": []string{"user1", "user2"}})
}

// @Route("GET", "/admin/logs")
// @Security(ips="192.168.1.100,10.0.0.50")
// @Auth(role="admin")
func AdminLogs(c *gin.Context) {
    // Only accessible from specific IPs
    c.JSON(200, gin.H{"logs": []string{"log1", "log2"}})
}
```

### Security Patterns

#### Private Network Access

```go
// @Route("GET", "/internal/api")
// @Security(private)
func InternalAPI(c *gin.Context) {
    // Allows private networks (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
}
```

#### Specific Network Access

```go
// @Route("GET", "/corporate/api")
// @Security(networks="10.0.0.0/8,172.16.0.0/12")
func CorporateAPI(c *gin.Context) {
    // Only specific corporate networks
}
```

#### IP-Based Access

```go
// @Route("GET", "/monitoring/api")
// @Security(ips="192.168.1.100,192.168.1.101,10.0.0.50")
func MonitoringAPI(c *gin.Context) {
    // Only specific monitoring servers
}
```

## Security Functions

### Convenience Functions

```go
// Allow only localhost
middleware := deco.AllowLocalhostOnly()

// Allow private networks
middleware := deco.AllowPrivateNetworks()

// Allow specific networks
middleware := deco.AllowSpecificNetworks([]string{"10.0.0.0/8", "192.168.1.0/24"})

// Allow specific IPs
middleware := deco.AllowSpecificIPs([]string{"192.168.1.100", "10.0.0.50"})
```

## Best Practices

### 1. Principle of Least Privilege

```go
// Good: Specific network access
// @Security(networks="192.168.1.0/24")

// Avoid: Too permissive
// @Security(private)  // Unless really needed
```

### 2. Logging and Monitoring

```go
securityConfig := &deco.SecurityConfig{
    LogBlockedAttempts: true,  // Always enable logging
    ErrorMessage: "Access denied",  // Generic message for security
}
```

### 3. Environment-Specific Configuration

```go
func getSecurityConfig() *deco.SecurityConfig {
    env := os.Getenv("ENVIRONMENT")
    
    switch env {
    case "development":
        return &deco.SecurityConfig{
            AllowPrivateNetworks: true,
            AllowLocalhost: true,
        }
    case "production":
        return &deco.SecurityConfig{
            AllowedNetworks: []string{"10.0.0.0/8"},
            AllowLocalhost: false,
        }
    default:
        return deco.DefaultSecurityConfig()
    }
}
```

## Common Security Issues

### 1. Overly Permissive Access

```go
// ❌ Bad: Too permissive
securityConfig := &deco.SecurityConfig{
    AllowPrivateNetworks: true,  // Allows all private networks
}

// ✅ Good: Specific access
securityConfig := &deco.SecurityConfig{
    AllowedNetworks: []string{"10.0.0.0/8"},  // Only corporate network
}
```

### 2. Missing Logging

```go
// ❌ Bad: No logging
securityConfig := &deco.SecurityConfig{
    LogBlockedAttempts: false,
}

// ✅ Good: With logging
securityConfig := &deco.SecurityConfig{
    LogBlockedAttempts: true,
}
```

## Next Steps

- **[Usage Guide](./usage.md)** - General usage guide
- **[API Reference](./api.md)** - Complete API documentation
- **[Examples](./examples.md)** - Security examples 