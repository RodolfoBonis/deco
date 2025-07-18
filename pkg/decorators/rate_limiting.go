package decorators

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter interface for different rate limiting implementations
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error)
	Reset(ctx context.Context, key string) error
}

// MemoryRateLimiter local in-memory implementation
type MemoryRateLimiter struct {
	buckets map[string]*TokenBucket
}

// TokenBucket represents a token bucket
type TokenBucket struct {
	tokens     int
	lastRefill time.Time
	limit      int
	window     time.Duration
}

// RedisRateLimiter distributed implementation with Redis
type RedisRateLimiter struct {
	client *redis.Client
}

// RateLimitResponse response when rate limit is exceeded
type RateLimitResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	Limit      int    `json:"limit"`
	Remaining  int    `json:"remaining"`
	RetryAfter int    `json:"retry_after"`
}

// KeyGeneratorFunc function to generate rate limiting keys
type KeyGeneratorFunc func(c *gin.Context) string

// Default key generation functions
var (
	IPKeyGenerator = func(c *gin.Context) string {
		return "ratelimit:ip:" + c.ClientIP()
	}

	UserKeyGenerator = func(c *gin.Context) string {
		userID := c.GetString("user_id")
		if userID == "" {
			return "ratelimit:anonymous:" + c.ClientIP()
		}
		return "ratelimit:user:" + userID
	}

	EndpointKeyGenerator = func(c *gin.Context) string {
		return fmt.Sprintf("ratelimit:endpoint:%s:%s:%s", c.Request.Method, c.FullPath(), c.ClientIP())
	}
)

// NewMemoryRateLimiter creates an in-memory rate limiter
func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
}

// Allow checks if the request can proceed (in-memory implementation)
func (m *MemoryRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, retryAfter time.Duration, err error) {
	// Use context for timeout and cancellation
	select {
	case <-ctx.Done():
		return false, 0, 0, ctx.Err()
	default:
	}

	now := time.Now()

	bucket, exists := m.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			tokens:     limit - 1,
			lastRefill: now,
			limit:      limit,
			window:     window,
		}
		m.buckets[key] = bucket
		return true, limit - 1, 0, nil
	}

	// Calculate how many tokens should be added
	elapsed := now.Sub(bucket.lastRefill)
	if elapsed >= window {
		// Complete bucket reset
		bucket.tokens = limit
		bucket.lastRefill = now
	} else {
		// Add tokens proportionally
		tokensToAdd := int(elapsed * time.Duration(limit) / window)
		bucket.tokens = minValue(bucket.limit, bucket.tokens+tokensToAdd)
		if tokensToAdd > 0 {
			bucket.lastRefill = now
		}
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true, bucket.tokens, 0, nil
	}

	// Calculate time until next token
	timeUntilNextToken := window - elapsed
	return false, 0, timeUntilNextToken, nil
}

// Reset clears the bucket for a key (in-memory implementation)
func (m *MemoryRateLimiter) Reset(ctx context.Context, key string) error {
	// Use context for timeout and cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	delete(m.buckets, key)
	return nil
}

// NewRedisRateLimiter creates a distributed rate limiter with Redis
func NewRedisRateLimiter(config RedisConfig) (*RedisRateLimiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Address,
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	return &RedisRateLimiter{client: client}, nil
}

// Allow checks if the request can proceed (Redis implementation)
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, retryAfter time.Duration, err error) {
	// Use context for timeout and cancellation
	select {
	case <-ctx.Done():
		return false, 0, 0, ctx.Err()
	default:
	}

	// Lua script for atomic rate limiting operation
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local current_time = tonumber(ARGV[3])
		
		-- Get current information
		local bucket = redis.call('HMGET', key, 'count', 'reset_time')
		local count = tonumber(bucket[1]) or 0
		local reset_time = tonumber(bucket[2]) or current_time
		
		-- If window time has passed, reset
		if current_time >= reset_time then
			count = 0
			reset_time = current_time + window
		end
		
		-- Check if request can be made
		if count >= limit then
			local retry_after = reset_time - current_time
			return {0, count, retry_after}
		end
		
		-- Increment counter
		count = count + 1
		redis.call('HMSET', key, 'count', count, 'reset_time', reset_time)
		redis.call('EXPIRE', key, math.ceil(window))
		
		local remaining = limit - count
		return {1, remaining, 0}
	`

	now := time.Now().Unix()
	windowSeconds := int64(window.Seconds())

	result, err := r.client.Eval(ctx, script, []string{key}, limit, windowSeconds, now).Result()
	if err != nil {
		return false, 0, 0, fmt.Errorf("redis rate limiting error: %v", err)
	}

	values := result.([]interface{})
	allowed = values[0].(int64) == 1
	remaining = int(values[1].(int64))
	retryAfter = time.Duration(values[2].(int64)) * time.Second

	return allowed, remaining, retryAfter, nil
}

// Reset clears the bucket for a key (Redis implementation)
func (r *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	// Use context for timeout and cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return r.client.Del(ctx, key).Err()
}

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(config *RateLimitConfig, keyGen KeyGeneratorFunc) gin.HandlerFunc {
	var limiter RateLimiter
	var err error

	// Choose implementation based on configuration
	if config.Type == "redis" {
		redisConfig := DefaultConfig().Redis
		limiter, err = NewRedisRateLimiter(redisConfig)
		if err != nil {
			// Fallback to memory if Redis fails
			limiter = NewMemoryRateLimiter()
		}
	} else {
		limiter = NewMemoryRateLimiter()
	}

	return func(c *gin.Context) {
		if !config.Enabled {
			c.Next()
			return
		}

		// Generate key for this client/endpoint
		key := keyGen(c)

		// Check rate limit
		allowed, remaining, retryAfter, err := limiter.Allow(
			c.Request.Context(),
			key,
			config.DefaultRPS,
			time.Minute, // 1 minute window
		)

		if err != nil {
			// In case of error, allow request (fail-open)
			c.Next()
			return
		}

		// Add informative headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(config.DefaultRPS))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10))

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))

			response := RateLimitResponse{
				Error:      "rate_limit_exceeded",
				Message:    "Request rate exceeded. Please try again later.",
				Limit:      config.DefaultRPS,
				Remaining:  0,
				RetryAfter: int(retryAfter.Seconds()),
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, response)
			return
		}

		c.Next()
	}
}

// RateLimitByIP rate limiting middleware by IP
func RateLimitByIP(config *RateLimitConfig) gin.HandlerFunc {
	return RateLimitMiddleware(config, IPKeyGenerator)
}

// RateLimitByUser rate limiting middleware by user
func RateLimitByUser(config *RateLimitConfig) gin.HandlerFunc {
	return RateLimitMiddleware(config, UserKeyGenerator)
}

// RateLimitByEndpoint rate limiting middleware by endpoint
func RateLimitByEndpoint(config *RateLimitConfig) gin.HandlerFunc {
	return RateLimitMiddleware(config, EndpointKeyGenerator)
}

// CustomRateLimit customizable rate limiting middleware
func CustomRateLimit(limit int, window time.Duration, keyGen KeyGeneratorFunc, rateLimiterType string) gin.HandlerFunc {
	config := &RateLimitConfig{
		Enabled:    true,
		Type:       rateLimiterType,
		DefaultRPS: limit,
	}

	// Create specific limiter based on type
	var limiter RateLimiter
	if rateLimiterType == "redis" {
		redisConfig := DefaultConfig().Redis
		if redisLimiter, err := NewRedisRateLimiter(redisConfig); err == nil {
			limiter = redisLimiter
		} else {
			limiter = NewMemoryRateLimiter()
		}
	} else {
		limiter = NewMemoryRateLimiter()
	}

	return func(c *gin.Context) {
		if !config.Enabled {
			c.Next()
			return
		}

		// Generate key for this client/endpoint
		key := keyGen(c)

		// Use context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Check rate limit using the custom window
		allowed, remaining, retryAfter, err := limiter.Allow(
			ctx,
			key,
			limit,
			window, // Use the custom window parameter
		)

		if err != nil {
			// In case of error, allow request (fail-open)
			c.Next()
			return
		}

		// Add informative headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Window", window.String())

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))

			response := RateLimitResponse{
				Error:      "rate_limit_exceeded",
				Message:    fmt.Sprintf("Request rate exceeded. Limit: %d per %v", limit, window),
				Limit:      limit,
				Remaining:  0,
				RetryAfter: int(retryAfter.Seconds()),
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, response)
			return
		}

		c.Next()
	}
}

// ParseRateLimitArgs parses @RateLimit decorator arguments
func ParseRateLimitArgs(args []string) (limit int, window time.Duration, rateLimiterType string, keyGen KeyGeneratorFunc) {
	limit = 100                // default
	window = time.Minute       // default
	rateLimiterType = "memory" // default
	keyGen = IPKeyGenerator    // default

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

			switch key {
			case "limit", "rps":
				if parsed, err := strconv.Atoi(value); err == nil {
					limit = parsed
				}
			case "window":
				if parsed, err := time.ParseDuration(value); err == nil {
					window = parsed
				}
			case "type":
				rateLimiterType = value
			case "key", "by":
				switch value {
				case "ip":
					keyGen = IPKeyGenerator
				case "user":
					keyGen = UserKeyGenerator
				case "endpoint":
					keyGen = EndpointKeyGenerator
				}
			}
		}
	}

	return limit, window, rateLimiterType, keyGen
}

// createRateLimitMiddlewareInternal creates rate limiting middleware (for markers.go)
func createRateLimitMiddlewareInternal(args []string) gin.HandlerFunc {
	limit, window, rateLimiterType, keyGen := ParseRateLimitArgs(args)

	// Create specific limiter
	var limiter RateLimiter
	if rateLimiterType == "redis" {
		redisConfig := DefaultConfig().Redis
		if redisLimiter, err := NewRedisRateLimiter(redisConfig); err == nil {
			limiter = redisLimiter
		} else {
			limiter = NewMemoryRateLimiter()
		}
	} else {
		limiter = NewMemoryRateLimiter()
	}

	return func(c *gin.Context) {
		key := keyGen(c)

		allowed, remaining, retryAfter, err := limiter.Allow(
			c.Request.Context(),
			key,
			limit,
			window,
		)

		if err != nil {
			c.Next()
			return
		}

		// Informative headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Window", window.String())

		if !allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(retryAfter.Seconds()), 10))

			response := RateLimitResponse{
				Error:      "rate_limit_exceeded",
				Message:    fmt.Sprintf("Request rate exceeded. Limit: %d per %v", limit, window),
				Limit:      limit,
				Remaining:  0,
				RetryAfter: int(retryAfter.Seconds()),
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, response)
			return
		}

		c.Next()
	}
}

// minValue helper function
func minValue(a, b int) int {
	if a < b {
		return a
	}
	return b
}
