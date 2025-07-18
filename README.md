# Deco Framework 🚀

<div align="center">
  <img src="./docs/images/deco_gopher.png" alt="Go Gopher Artist" width="200" height="200">
  <br>
  <em>The Go Gopher decorator, crafting elegant APIs with simple annotations! ✨</em>
</div>

A modern, annotation-driven Go web framework built on top of Gin. Write web APIs using simple `@` annotations and let deco handle the heavy lifting - automatic route registration, middleware injection, validation, caching, rate limiting, security, and more!

## ✨ Why Deco?

- **🚀 Zero Boilerplate**: Define routes and middleware with simple annotations
- **🛡️ Built-in Security**: Automatic protection for internal endpoints
- **🔄 API Gateway**: Service discovery, load balancing, circuit breakers
- **📊 Observability**: Metrics, tracing, and monitoring out of the box
- **⚡ Production Ready**: Optimized builds with validation and minification

## 🚀 Quick Start

```bash
# Install CLI
go install github.com/RodolfoBonis/deco/cmd/deco@latest

# Initialize project
deco init

# Create your first handler
# @Route("GET", "/health")
# func HealthCheck(c *gin.Context) {
#     c.JSON(200, gin.H{"status": "healthy"})
# }

# Generate and run
deco generate
go run main.go
```

## 🎯 Key Features

### 🔒 Security
```go
// Automatic internal endpoint protection
r := deco.Default()

// Custom security configuration
securityConfig := &deco.SecurityConfig{
    AllowPrivateNetworks: true,
    AllowLocalhost: true,
}
r := deco.DefaultWithSecurity(securityConfig)
```

### 🔄 API Gateway
```go
// @Route("GET", "/api/users/:id")
// @Proxy(service="user-service", discovery="consul")
func GetUserProxy(c *gin.Context) {
    // Automatically forwards to user-service
}
```

### 💾 Caching & Rate Limiting
```go
// @Route("GET", "/users/:id")
// @Cache(ttl="5m")
// @RateLimit(limit=100, window="1m")
func GetUser(c *gin.Context) {
    // Response cached for 5 minutes, 100 requests per minute
}
```

## 📚 Documentation

- **[Installation Guide](docs/installation.md)** - Setup and configuration
- **[Usage Guide](docs/usage.md)** - How to use decorators and features
- **[API Reference](docs/api.md)** - Complete API documentation
- **[Security Guide](docs/security.md)** - Security features and best practices
- **[Proxy Guide](docs/proxy.md)** - API Gateway functionality
- **[Examples](docs/examples.md)** - Code examples and tutorials
- **[CLI Reference](docs/cli.md)** - Command line interface

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## 📚 Documentation

- **[Installation Guide](docs/installation.md)** - Setup and configuration
- **[Usage Guide](docs/usage.md)** - How to use decorators and features
- **[API Reference](docs/api.md)** - Complete API documentation
- **[Security Guide](docs/security.md)** - Security features and best practices
- **[Proxy Guide](docs/proxy.md)** - API Gateway functionality
- **[Examples](docs/examples.md)** - Code examples and tutorials
- **[CLI Reference](docs/cli.md)** - Command line interface

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details. 