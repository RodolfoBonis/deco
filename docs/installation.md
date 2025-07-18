# Installation Guide

## Prerequisites

- Go 1.22 or later
- Git

## Quick Installation

### 1. Install CLI Tool

```bash
go install github.com/RodolfoBonis/deco/cmd/deco@latest
```

### 2. Verify Installation

```bash
deco --version
```

## Project Setup

### 1. Initialize New Project

```bash
# Create project directory
mkdir my-api && cd my-api

# Initialize deco
deco init
```

This creates:
- `.deco.yaml` - Configuration file
- `.gitignore` - Git ignore file (if not exists)

### 2. Create Your First Handler

```go
// handlers/health.go
package handlers

import "github.com/gin-gonic/gin"

// @Route("GET", "/health")
func HealthCheck(c *gin.Context) {
    c.JSON(200, gin.H{"status": "healthy"})
}
```

### 3. Generate and Run

```bash
# Generate decorators
deco generate

# Run application
go run main.go
```

## Go Module Integration

### Initialize Go Module

```bash
go mod init my-api
go get github.com/RodolfoBonis/deco
```

### Import in main.go

```go
package main

import (
    _ "my-api/.deco"  // Import generated decorators
    deco "github.com/RodolfoBonis/deco"
)

func main() {
    r := deco.Default()
    r.Run(":8080")
}
```

## Development Workflow

### Development Mode

```bash
# Watch for changes and auto-regenerate
deco dev
```

### Production Build

```bash
# Build with optimizations
deco build
```

### Validation

```bash
# Validate decorators
deco validate
```

## Next Steps

- **[Usage Guide](./usage.md)** - Learn how to use decorators
- **[Examples](./examples.md)** - See examples in action
- **[API Reference](./api.md)** - Complete API documentation
