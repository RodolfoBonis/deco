package main

import (
	"fmt"

	deco "github.com/RodolfoBonis/deco"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Apply security middleware to internal endpoints
	// This will block access from external networks
	internalGroup := r.Group("/internal")
	internalGroup.Use(deco.AllowLocalhostOnly()) // Only localhost can access

	// Alternative: Allow private networks (VPN, internal networks)
	// internalGroup.Use(deco.AllowPrivateNetworks())

	// Alternative: Allow specific networks
	// internalGroup.Use(deco.AllowSpecificNetworks([]string{"192.168.1.0/24", "10.0.0.0/8"}))

	// Alternative: Allow specific IPs
	// internalGroup.Use(deco.AllowSpecificIPs([]string{"192.168.1.100", "10.0.0.50"}))

	// Register routes manually for this example
	r.GET("/internal/health", deco.AllowLocalhostOnly(), InternalHealthCheck)
	r.GET("/internal/metrics", deco.AllowPrivateNetworks(), InternalMetrics)
	r.GET("/public/info", PublicInfo)

	fmt.Println("🚀 Server running on :8080")
	fmt.Println("🔒 Internal endpoints protected:")
	fmt.Println("   - /internal/health (localhost only)")
	fmt.Println("   - /internal/metrics (private networks only)")
	fmt.Println("   - /public/info (public access)")

	r.Run(":8080")
}

// @Route(method="GET", path="/internal/health")
// @Security(localhost)
// @Description("Internal health check - localhost only")
func InternalHealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "healthy", "internal": true})
}

// @Route(method="GET", path="/internal/metrics")
// @Security(private)
// @Description("Internal metrics - private networks only")
func InternalMetrics(c *gin.Context) {
	c.JSON(200, gin.H{"metrics": "internal_data"})
}

// Public endpoints (no security restrictions)
// @Route(method="GET", path="/public/info")
// @Description("Public information endpoint")
func PublicInfo(c *gin.Context) {
	c.JSON(200, gin.H{"info": "public_data"})
}
