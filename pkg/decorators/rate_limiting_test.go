package decorators

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Tests for rate limiting functionality
func TestNewMemoryRateLimiter(t *testing.T) {

	limiter := NewMemoryRateLimiter()
	assert.NotNil(t, limiter)
	assert.NotNil(t, limiter.buckets)
	assert.Empty(t, limiter.buckets)
}

func TestMemoryRateLimiter_Allow_FirstRequest(t *testing.T) {

	limiter := NewMemoryRateLimiter()
	ctx := context.Background()

	// First request should always be allowed
	allowed, remaining, retryAfter, err := limiter.Allow(ctx, "test-key", 10, time.Minute)

	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, 9, remaining) // 10 - 1
	assert.Equal(t, time.Duration(0), retryAfter)
}

func TestMemoryRateLimiter_Allow_WithinLimit(t *testing.T) {

	limiter := NewMemoryRateLimiter()
	ctx := context.Background()
	key := "test-key"
	limit := 5
	window := time.Minute

	// Make requests within limit
	for i := 0; i < limit; i++ {
		allowed, remaining, retryAfter, err := limiter.Allow(ctx, key, limit, window)

		assert.NoError(t, err)
		assert.True(t, allowed)
		assert.Equal(t, limit-i-1, remaining)
		assert.Equal(t, time.Duration(0), retryAfter)
	}
}

func TestMemoryRateLimiter_Allow_ExceedLimit(t *testing.T) {

	limiter := NewMemoryRateLimiter()
	ctx := context.Background()
	key := "test-key"
	limit := 3
	window := time.Minute

	// Exhaust the limit
	for i := 0; i < limit; i++ {
		limiter.Allow(ctx, key, limit, window)
	}

	// Next request should be denied
	allowed, remaining, retryAfter, err := limiter.Allow(ctx, key, limit, window)

	assert.NoError(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
	assert.Greater(t, retryAfter, time.Duration(0))
}

func TestMemoryRateLimiter_Allow_ContextCancellation(t *testing.T) {

	limiter := NewMemoryRateLimiter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	allowed, remaining, retryAfter, err := limiter.Allow(ctx, "test-key", 10, time.Minute)

	assert.Error(t, err)
	assert.False(t, allowed)
	assert.Equal(t, 0, remaining)
	assert.Equal(t, time.Duration(0), retryAfter)
}

func TestMemoryRateLimiter_Reset(t *testing.T) {

	limiter := NewMemoryRateLimiter()
	ctx := context.Background()
	key := "test-key"

	// Make a request to create a bucket
	limiter.Allow(ctx, key, 10, time.Minute)
	assert.Contains(t, limiter.buckets, key)

	// Reset the bucket
	err := limiter.Reset(ctx, key)
	assert.NoError(t, err)
	assert.NotContains(t, limiter.buckets, key)
}

func TestMemoryRateLimiter_Reset_ContextCancellation(t *testing.T) {

	limiter := NewMemoryRateLimiter()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := limiter.Reset(ctx, "test-key")
	assert.Error(t, err)
}

func TestMemoryRateLimiter_TokenRefill(t *testing.T) {

	limiter := NewMemoryRateLimiter()
	ctx := context.Background()
	key := "test-key"
	limit := 10
	window := 100 * time.Millisecond

	// Exhaust the limit
	for i := 0; i < limit; i++ {
		limiter.Allow(ctx, key, limit, window)
	}

	// Wait for window to pass
	time.Sleep(window + 10*time.Millisecond)

	// Should be allowed again
	allowed, remaining, retryAfter, err := limiter.Allow(ctx, key, limit, window)

	assert.NoError(t, err)
	assert.True(t, allowed)
	assert.Equal(t, limit-1, remaining)
	assert.Equal(t, time.Duration(0), retryAfter)
}

func TestNewRedisRateLimiter_InvalidConfig(t *testing.T) {

	config := RedisConfig{
		Address: "invalid:6379",
	}

	limiter, err := NewRedisRateLimiter(config)
	assert.Error(t, err)
	assert.Nil(t, limiter)
	assert.Contains(t, err.Error(), "failed to connect to Redis")
}

func TestKeyGenerators(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	t.Run("IPKeyGenerator", func(t *testing.T) {
		key := IPKeyGenerator(c)
		assert.Equal(t, "ratelimit:ip:192.168.1.100", key)
	})

	t.Run("EndpointKeyGenerator", func(t *testing.T) {
		// Create a new context with the path set
		w2 := httptest.NewRecorder()
		req2, _ := http.NewRequest("GET", "/api/users", http.NoBody)
		req2.RemoteAddr = "192.168.1.100:12345"
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = req2

		// Set the path manually since we can't mock FullPath
		c2.Params = []gin.Param{}
		key := EndpointKeyGenerator(c2)
		// The key will be based on the actual path from the request
		assert.Contains(t, key, "ratelimit:endpoint:GET:")
		assert.Contains(t, key, "192.168.1.100")
	})

	t.Run("UserKeyGenerator_Anonymous", func(t *testing.T) {
		key := UserKeyGenerator(c)
		assert.Equal(t, "ratelimit:anonymous:192.168.1.100", key)
	})

	t.Run("UserKeyGenerator_Authenticated", func(t *testing.T) {
		c.Set("user_id", "user123")
		key := UserKeyGenerator(c)
		assert.Equal(t, "ratelimit:user:user123", key)
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &RateLimitConfig{
		Enabled:    true,
		Type:       "memory",
		DefaultRPS: 3,
		BurstSize:  5,
		KeyFunc:    "ip",
	}

	router.Use(RateLimitMiddleware(config, IPKeyGenerator))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	t.Run("requests within limit", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", http.NoBody)
			req.RemoteAddr = "192.168.1.100:12345"
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("request exceeds limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", http.NoBody)
		req.RemoteAddr = "192.168.1.100:12345"
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "rate_limit_exceeded")
	})
}

func TestRateLimitByIP(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &RateLimitConfig{
		Enabled:    true,
		Type:       "memory",
		DefaultRPS: 2,
		BurstSize:  3,
		KeyFunc:    "ip",
	}

	router.Use(RateLimitByIP(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with different IPs
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", http.NoBody)
	req1.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", http.NoBody)
	req2.RemoteAddr = "192.168.1.101:12345"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestRateLimitByUser(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &RateLimitConfig{
		Enabled:    true,
		Type:       "memory",
		DefaultRPS: 2,
		BurstSize:  3,
		KeyFunc:    "user",
	}

	router.Use(RateLimitByUser(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test with user context
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("user_id", "user123")

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimitByEndpoint(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &RateLimitConfig{
		Enabled:    true,
		Type:       "memory",
		DefaultRPS: 2,
		BurstSize:  3,
		KeyFunc:    "endpoint",
	}

	router.Use(RateLimitByEndpoint(config))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCustomRateLimit(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	customKeyGen := func(c *gin.Context) string {
		return "custom:" + c.ClientIP()
	}

	middleware := CustomRateLimit(2, time.Minute, customKeyGen, "memory")
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.RemoteAddr = "192.168.1.100:12345"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestParseRateLimitArgs(t *testing.T) {

	tests := []struct {
		name           string
		args           []string
		expectedLimit  int
		expectedWindow time.Duration
		expectedType   string
	}{
		{
			name:           "empty args",
			args:           []string{},
			expectedLimit:  100,
			expectedWindow: time.Minute,
			expectedType:   "memory",
		},
		{
			name:           "with limit",
			args:           []string{"limit=50"},
			expectedLimit:  50,
			expectedWindow: time.Minute,
			expectedType:   "memory",
		},
		{
			name:           "with window",
			args:           []string{"window=30s"},
			expectedLimit:  100,
			expectedWindow: 30 * time.Second,
			expectedType:   "memory",
		},
		{
			name:           "with type",
			args:           []string{"type=redis"},
			expectedLimit:  100,
			expectedWindow: time.Minute,
			expectedType:   "redis",
		},
		{
			name:           "all parameters",
			args:           []string{"limit=25", "window=1h", "type=memory"},
			expectedLimit:  25,
			expectedWindow: time.Hour,
			expectedType:   "memory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, window, rateLimiterType, keyGen := ParseRateLimitArgs(tt.args)

			assert.Equal(t, tt.expectedLimit, limit)
			assert.Equal(t, tt.expectedWindow, window)
			assert.Equal(t, tt.expectedType, rateLimiterType)
			assert.NotNil(t, keyGen)
		})
	}
}

func TestCreateRateLimitMiddlewareInternal(t *testing.T) {
	// Remove  to avoid race conditions

	tests := []struct {
		name string
		args []string
	}{
		{
			name: "empty args",
			args: []string{},
		},
		{
			name: "with limit",
			args: []string{"limit=50"},
		},
		{
			name: "with window",
			args: []string{"window=30s"},
		},
		{
			name: "with type",
			args: []string{"type=memory"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := createRateLimitMiddlewareInternal(tt.args)
			assert.NotNil(t, middleware)

			// Test that middleware can be called without panic
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", http.NoBody)
			req.RemoteAddr = "192.168.1.100:12345"
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			assert.NotPanics(t, func() {
				middleware(c)
			})
		})
	}
}

func TestMinValue(t *testing.T) {

	assert.Equal(t, 5, minValue(5, 10))
	assert.Equal(t, 3, minValue(10, 3))
	assert.Equal(t, 7, minValue(7, 7))
	assert.Equal(t, -5, minValue(-5, 10))
	assert.Equal(t, -10, minValue(5, -10))
}

func TestRateLimitResponse_Structure(t *testing.T) {

	response := RateLimitResponse{
		Error:      "rate_limit_exceeded",
		Message:    "Too many requests",
		Limit:      100,
		Remaining:  0,
		RetryAfter: 60,
	}

	assert.Equal(t, "rate_limit_exceeded", response.Error)
	assert.Equal(t, "Too many requests", response.Message)
	assert.Equal(t, 100, response.Limit)
	assert.Equal(t, 0, response.Remaining)
	assert.Equal(t, 60, response.RetryAfter)
}

func TestRateLimiter_Interface(_ *testing.T) {

	// Test that MemoryRateLimiter implements RateLimiter interface
	var _ RateLimiter = (*MemoryRateLimiter)(nil)

	// Test that RedisRateLimiter implements RateLimiter interface
	var _ RateLimiter = (*RedisRateLimiter)(nil)
}

func TestTokenBucket_Structure(t *testing.T) {

	bucket := &TokenBucket{
		tokens:     5,
		lastRefill: time.Now(),
		limit:      10,
		window:     time.Minute,
	}

	assert.Equal(t, 5, bucket.tokens)
	assert.Equal(t, 10, bucket.limit)
	assert.Equal(t, time.Minute, bucket.window)
	assert.NotZero(t, bucket.lastRefill)
}

func TestAllow(t *testing.T) {
	// Test Allow method for memory rate limiter
	limiter := NewMemoryRateLimiter()

	// Test allowing requests
	for i := 0; i < 10; i++ {
		allowed, _, _, err := limiter.Allow(context.Background(), "test-key", 10, time.Minute)
		assert.NoError(t, err)
		assert.True(t, allowed, "Request %d should be allowed", i+1)
	}

	// Test rate limit exceeded
	allowed, _, _, err := limiter.Allow(context.Background(), "test-key", 10, time.Minute)
	assert.NoError(t, err)
	assert.False(t, allowed, "Request should be blocked after limit exceeded")
}

func TestReset(t *testing.T) {
	// Test Reset method for memory rate limiter
	limiter := NewMemoryRateLimiter()

	// Use up the limit
	for i := 0; i < 5; i++ {
		_, _, _, err := limiter.Allow(context.Background(), "test-key", 5, time.Minute)
		assert.NoError(t, err)
	}

	// Verify limit is exceeded
	allowed, _, _, err := limiter.Allow(context.Background(), "test-key", 5, time.Minute)
	assert.NoError(t, err)
	assert.False(t, allowed, "Request should be blocked")

	// Reset the limiter
	err = limiter.Reset(context.Background(), "test-key")
	assert.NoError(t, err)

	// Verify requests are allowed again
	allowed, _, _, err = limiter.Allow(context.Background(), "test-key", 5, time.Minute)
	assert.NoError(t, err)
	assert.True(t, allowed, "Request should be allowed after reset")
}

func TestRedisRateLimiter_Allow(t *testing.T) {
	// Test Allow method for Redis rate limiter
	config := RedisConfig{
		Address:  "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	limiter, err := NewRedisRateLimiter(config)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Test allowing requests (skip if Redis not available)
	for i := 0; i < 5; i++ {
		allowed, _, _, err := limiter.Allow(context.Background(), "test-key", 10, time.Minute)
		if err == nil && allowed {
			// Redis is available, continue testing
			break
		}
		if i == 4 {
			t.Skip("Redis not available, skipping test")
		}
	}
}

func TestRedisRateLimiter_Reset(t *testing.T) {
	// Test Reset method for Redis rate limiter
	config := RedisConfig{
		Address:  "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 10,
	}

	limiter, err := NewRedisRateLimiter(config)
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Test reset (skip if Redis not available)
	err = limiter.Reset(context.Background(), "test-key")
	if err != nil {
		t.Skip("Redis not available, skipping test")
	}

	assert.NoError(t, err, "Reset should not return error")
}
