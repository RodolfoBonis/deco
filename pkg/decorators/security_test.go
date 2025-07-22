package decorators

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Tests for security functionality
func TestDefaultSecurityConfig(t *testing.T) {

	config := DefaultSecurityConfig()
	assert.NotNil(t, config)
	assert.True(t, config.AllowLocalhost)
	assert.False(t, config.AllowPrivateNetworks)
	assert.Contains(t, config.AllowedNetworks, "127.0.0.1/32")
	assert.Contains(t, config.AllowedNetworks, "::1/128")
	assert.Equal(t, "Access denied: This endpoint is restricted to internal networks", config.ErrorMessage)
	assert.True(t, config.LogBlockedAttempts)
}

func TestSecureInternalEndpoints_AllowLocalhost(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := SecureInternalEndpoints(DefaultSecurityConfig())
	router.Use(middleware)
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test localhost access
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", http.NoBody)
	req.RemoteAddr = "127.0.0.1:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecureInternalEndpoints_BlockExternalIP(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := SecureInternalEndpoints(DefaultSecurityConfig())
	router.Use(middleware)
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test external IP access
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", http.NoBody)
	req.RemoteAddr = "203.0.113.1:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "access_denied")
}

func TestSecureInternalEndpoints_AllowSpecificIPs(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &SecurityConfig{
		AllowedIPs:         []string{"192.168.1.100", "10.0.0.50"},
		AllowLocalhost:     true,
		LogBlockedAttempts: true,
	}

	middleware := SecureInternalEndpoints(config)
	router.Use(middleware)
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test allowed specific IP
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecureInternalEndpoints_AllowPrivateNetworks(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &SecurityConfig{
		AllowPrivateNetworks: true,
		AllowLocalhost:       true,
		LogBlockedAttempts:   true,
	}

	middleware := SecureInternalEndpoints(config)
	router.Use(middleware)
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test private network access
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", http.NoBody)
	req.RemoteAddr = "192.168.1.50:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecureInternalEndpoints_AllowHostname(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &SecurityConfig{
		AllowedHosts:       []string{"internal.example.com"},
		AllowLocalhost:     true,
		LogBlockedAttempts: true,
	}

	middleware := SecureInternalEndpoints(config)
	router.Use(middleware)
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test allowed hostname
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", http.NoBody)
	req.RemoteAddr = "203.0.113.1:12345"
	req.Host = "internal.example.com"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecureInternalEndpoints_AllowWildcardHostname(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &SecurityConfig{
		AllowedHosts:       []string{"*.example.com"},
		AllowLocalhost:     true,
		LogBlockedAttempts: true,
	}

	middleware := SecureInternalEndpoints(config)
	router.Use(middleware)
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test wildcard hostname
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", http.NoBody)
	req.RemoteAddr = "203.0.113.1:12345"
	req.Host = "api.example.com"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetClientIP_Headers(t *testing.T) {

	t.Run("X-Forwarded-For header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", http.NoBody)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Request.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1")
		clientIP := getClientIP(c)
		assert.Equal(t, "192.168.1.100", clientIP)
	})

	t.Run("X-Real-IP header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", http.NoBody)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Request.Header.Set("X-Real-IP", "203.0.113.1")
		clientIP := getClientIP(c)
		assert.Equal(t, "203.0.113.1", clientIP)
	})

	t.Run("X-Client-IP header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", http.NoBody)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Request.Header.Set("X-Client-IP", "172.16.0.1")
		clientIP := getClientIP(c)
		assert.Equal(t, "172.16.0.1", clientIP)
	})

	t.Run("fallback to ClientIP", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", http.NoBody)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Request.RemoteAddr = "127.0.0.1:12345"
		clientIP := getClientIP(c)
		assert.Equal(t, "127.0.0.1", clientIP)
	})
}

func TestIsIPAllowed(t *testing.T) {

	tests := []struct {
		name            string
		clientIP        string
		allowedNetworks []string
		allowedIPs      []string
		expected        bool
	}{
		{
			name:            "allowed specific IP",
			clientIP:        "192.168.1.100",
			allowedNetworks: []string{},
			allowedIPs:      []string{"192.168.1.100"},
			expected:        true,
		},
		{
			name:            "allowed network",
			clientIP:        "192.168.1.50",
			allowedNetworks: []string{"192.168.1.0/24"},
			allowedIPs:      []string{},
			expected:        true,
		},
		{
			name:            "not allowed",
			clientIP:        "203.0.113.1",
			allowedNetworks: []string{"192.168.1.0/24"},
			allowedIPs:      []string{},
			expected:        false,
		},
		{
			name:            "invalid IP",
			clientIP:        "invalid-ip",
			allowedNetworks: []string{"192.168.1.0/24"},
			allowedIPs:      []string{},
			expected:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse networks
			networks := make([]*net.IPNet, 0)
			for _, network := range tt.allowedNetworks {
				if _, ipNet, err := net.ParseCIDR(network); err == nil {
					networks = append(networks, ipNet)
				}
			}

			result := isIPAllowed(tt.clientIP, networks, tt.allowedIPs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsHostnameAllowed(t *testing.T) {

	tests := []struct {
		name         string
		hostname     string
		allowedHosts []string
		expected     bool
	}{
		{
			name:         "exact match",
			hostname:     "api.example.com",
			allowedHosts: []string{"api.example.com"},
			expected:     true,
		},
		{
			name:         "wildcard match",
			hostname:     "api.example.com",
			allowedHosts: []string{"*.example.com"},
			expected:     true,
		},
		{
			name:         "wildcard root domain",
			hostname:     "example.com",
			allowedHosts: []string{"*.example.com"},
			expected:     true,
		},
		{
			name:         "with port",
			hostname:     "api.example.com:8080",
			allowedHosts: []string{"api.example.com"},
			expected:     true,
		},
		{
			name:         "no match",
			hostname:     "other.com",
			allowedHosts: []string{"*.example.com"},
			expected:     false,
		},
		{
			name:         "empty allowed hosts",
			hostname:     "api.example.com",
			allowedHosts: []string{},
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isHostnameAllowed(tt.hostname, tt.allowedHosts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAllowLocalhostOnly(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := AllowLocalhostOnly()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test localhost access
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "127.0.0.1:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAllowPrivateNetworks(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := AllowPrivateNetworks()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test private network access
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAllowSpecificNetworks(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	networks := []string{"192.168.1.0/24", "10.0.0.0/8"}
	middleware := AllowSpecificNetworks(networks)
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test allowed network
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.50:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAllowSpecificIPs(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	ips := []string{"192.168.1.100", "10.0.0.50"}
	middleware := AllowSpecificIPs(ips)
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test allowed IP
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSecurityConfig_Structure(t *testing.T) {

	config := &SecurityConfig{
		AllowedNetworks:      []string{"192.168.1.0/24"},
		AllowedIPs:           []string{"192.168.1.100"},
		AllowedHosts:         []string{"api.example.com"},
		AllowLocalhost:       true,
		AllowPrivateNetworks: false,
		ErrorMessage:         "Custom error message",
		LogBlockedAttempts:   true,
	}

	assert.Equal(t, []string{"192.168.1.0/24"}, config.AllowedNetworks)
	assert.Equal(t, []string{"192.168.1.100"}, config.AllowedIPs)
	assert.Equal(t, []string{"api.example.com"}, config.AllowedHosts)
	assert.True(t, config.AllowLocalhost)
	assert.False(t, config.AllowPrivateNetworks)
	assert.Equal(t, "Custom error message", config.ErrorMessage)
	assert.True(t, config.LogBlockedAttempts)
}

func TestSecureInternalEndpoints_NilConfig(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := SecureInternalEndpoints(nil)
	router.Use(middleware)
	router.GET("/internal", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with localhost (should be allowed with default config)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/internal", http.NoBody)
	req.RemoteAddr = "127.0.0.1:12345"
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
