package decorators

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityConfig holds security configuration for internal endpoints
type SecurityConfig struct {
	// Allowed networks in CIDR notation (e.g., "192.168.1.0/24", "10.0.0.0/8")
	AllowedNetworks []string
	// Allowed IP addresses (individual IPs)
	AllowedIPs []string
	// Allowed hostnames/domains
	AllowedHosts []string
	// Whether to allow localhost/127.0.0.1
	AllowLocalhost bool
	// Whether to allow private networks (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
	AllowPrivateNetworks bool
	// Custom error message
	ErrorMessage string
	// Whether to log blocked attempts
	LogBlockedAttempts bool
}

// DefaultSecurityConfig returns a secure default configuration
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		AllowedNetworks:      []string{"127.0.0.1/32", "::1/128"}, // Only localhost by default
		AllowedIPs:           []string{},
		AllowedHosts:         []string{},
		AllowLocalhost:       true,
		AllowPrivateNetworks: false, // Disabled by default for security
		ErrorMessage:         "Access denied: This endpoint is restricted to internal networks",
		LogBlockedAttempts:   true,
	}
}

// SecureInternalEndpoints creates a middleware to secure internal gin-decorators endpoints
func SecureInternalEndpoints(config *SecurityConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultSecurityConfig()
	}

	// Build allowed networks list
	allowedNetworks := make([]*net.IPNet, 0)

	// Add configured networks
	for _, network := range config.AllowedNetworks {
		if _, ipNet, err := net.ParseCIDR(network); err == nil {
			allowedNetworks = append(allowedNetworks, ipNet)
		}
	}

	// Add localhost if enabled
	if config.AllowLocalhost {
		if _, localhostIPv4, err := net.ParseCIDR("127.0.0.1/32"); err == nil {
			allowedNetworks = append(allowedNetworks, localhostIPv4)
		}
		if _, localhostIPv6, err := net.ParseCIDR("::1/128"); err == nil {
			allowedNetworks = append(allowedNetworks, localhostIPv6)
		}
	}

	// Add private networks if enabled
	if config.AllowPrivateNetworks {
		privateNetworks := []string{
			"10.0.0.0/8",     // Class A private
			"172.16.0.0/12",  // Class B private
			"192.168.0.0/16", // Class C private
		}
		for _, network := range privateNetworks {
			if _, ipNet, err := net.ParseCIDR(network); err == nil {
				allowedNetworks = append(allowedNetworks, ipNet)
			}
		}
	}

	return func(c *gin.Context) {
		clientIP := getClientIP(c)

		// Check if IP is allowed
		if isIPAllowed(clientIP, allowedNetworks, config.AllowedIPs) {
			c.Next()
			return
		}

		// Check if hostname is allowed
		if isHostnameAllowed(c.Request.Host, config.AllowedHosts) {
			c.Next()
			return
		}

		// Log blocked attempt if enabled
		if config.LogBlockedAttempts {
			fmt.Printf("🔒 SECURITY: Blocked access to internal endpoint from %s (Host: %s, Path: %s)\n",
				clientIP, c.Request.Host, c.Request.URL.Path)
		}

		// Return access denied
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "access_denied",
			"message": config.ErrorMessage,
			"details": "This endpoint is restricted to internal networks only",
		})
		c.Abort()
	}
}

// getClientIP extracts the real client IP from various headers
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header (common in reverse proxies)
	if forwardedFor := c.GetHeader("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Check X-Client-IP header
	if clientIP := c.GetHeader("X-Client-IP"); clientIP != "" {
		return clientIP
	}

	// Fallback to gin's ClientIP method
	return c.ClientIP()
}

// isIPAllowed checks if an IP address is in the allowed networks or IP list
func isIPAllowed(clientIP string, allowedNetworks []*net.IPNet, allowedIPs []string) bool {
	// Parse client IP
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	// Check against allowed networks
	for _, network := range allowedNetworks {
		if network.Contains(ip) {
			return true
		}
	}

	// Check against allowed individual IPs
	for _, allowedIP := range allowedIPs {
		if allowedIP == clientIP {
			return true
		}
	}

	return false
}

// isHostnameAllowed checks if a hostname is in the allowed hosts list
func isHostnameAllowed(hostname string, allowedHosts []string) bool {
	if len(allowedHosts) == 0 {
		return false
	}

	// Remove port if present
	if colonIndex := strings.Index(hostname, ":"); colonIndex != -1 {
		hostname = hostname[:colonIndex]
	}

	for _, allowedHost := range allowedHosts {
		if allowedHost == hostname {
			return true
		}
		// Support wildcard domains (e.g., "*.example.com")
		if strings.HasPrefix(allowedHost, "*.") {
			domain := strings.TrimPrefix(allowedHost, "*.")
			if strings.HasSuffix(hostname, "."+domain) || hostname == domain {
				return true
			}
		}
	}

	return false
}

// Convenience functions for common security configurations

// AllowLocalhostOnly creates a middleware that only allows localhost access
func AllowLocalhostOnly() gin.HandlerFunc {
	config := &SecurityConfig{
		AllowLocalhost:       true,
		AllowPrivateNetworks: false,
		ErrorMessage:         "Access denied: This endpoint is restricted to localhost only",
		LogBlockedAttempts:   true,
	}
	return SecureInternalEndpoints(config)
}

// AllowPrivateNetworks creates a middleware that allows private network access
func AllowPrivateNetworks() gin.HandlerFunc {
	config := &SecurityConfig{
		AllowLocalhost:       true,
		AllowPrivateNetworks: true,
		ErrorMessage:         "Access denied: This endpoint is restricted to private networks only",
		LogBlockedAttempts:   true,
	}
	return SecureInternalEndpoints(config)
}

// AllowSpecificNetworks creates a middleware that allows specific networks
func AllowSpecificNetworks(networks []string) gin.HandlerFunc {
	config := &SecurityConfig{
		AllowedNetworks:      networks,
		AllowLocalhost:       true,
		AllowPrivateNetworks: false,
		ErrorMessage:         "Access denied: This endpoint is restricted to authorized networks only",
		LogBlockedAttempts:   true,
	}
	return SecureInternalEndpoints(config)
}

// AllowSpecificIPs creates a middleware that allows specific IP addresses
func AllowSpecificIPs(ips []string) gin.HandlerFunc {
	config := &SecurityConfig{
		AllowedIPs:           ips,
		AllowLocalhost:       true,
		AllowPrivateNetworks: false,
		ErrorMessage:         "Access denied: This endpoint is restricted to authorized IPs only",
		LogBlockedAttempts:   true,
	}
	return SecureInternalEndpoints(config)
}
