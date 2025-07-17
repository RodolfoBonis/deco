package decorators

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// CacheEntry represents a cache entry
type CacheEntry struct {
	Data      []byte            `json:"data"`
	Headers   map[string]string `json:"headers"`
	Status    int               `json:"status"`
	ExpiresAt time.Time         `json:"expires_at"`
}

// CacheStore interface for different cache implementations
type CacheStore interface {
	Get(ctx context.Context, key string) (*CacheEntry, error)
	Set(ctx context.Context, key string, entry *CacheEntry, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Stats() CacheStats
}

// CacheStats cache statistics
type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Sets      int64   `json:"sets"`
	Deletes   int64   `json:"deletes"`
	Evictions int64   `json:"evictions"`
	Size      int64   `json:"size"`
	MaxSize   int64   `json:"max_size"`
	HitRate   float64 `json:"hit_rate"`
}

// MemoryCache in-memory cache implementation
type MemoryCache struct {
	mu      sync.RWMutex
	data    map[string]*CacheEntry
	maxSize int
	stats   CacheStats
}

// RedisCache Redis cache implementation
type RedisCache struct {
	client *redis.Client
	prefix string
	stats  CacheStats
}

// CacheKeyFunc function to generate cache key
type CacheKeyFunc func(c *gin.Context) string

// Default cache key generation functions
var (
	URLCacheKey = func(c *gin.Context) string {
		return fmt.Sprintf("cache:url:%s:%s", c.Request.Method, c.Request.URL.String())
	}

	UserURLCacheKey = func(c *gin.Context) string {
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}
		return fmt.Sprintf("cache:user:%s:url:%s:%s", userID, c.Request.Method, c.Request.URL.String())
	}

	EndpointCacheKey = func(c *gin.Context) string {
		return fmt.Sprintf("cache:endpoint:%s:%s", c.Request.Method, c.FullPath())
	}
)

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxSize int) *MemoryCache {
	return &MemoryCache{
		data:    make(map[string]*CacheEntry),
		maxSize: maxSize,
		stats:   CacheStats{MaxSize: int64(maxSize)},
	}
}

// Get retrieves cache entry (in-memory implementation)
func (m *MemoryCache) Get(ctx context.Context, key string) (*CacheEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entry, exists := m.data[key]
	if !exists {
		m.stats.Misses++
		m.updateHitRate()
		return nil, nil
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		m.mu.RUnlock()
		m.mu.Lock()
		delete(m.data, key)
		m.stats.Evictions++
		m.mu.Unlock()
		m.mu.RLock()

		m.stats.Misses++
		m.updateHitRate()
		return nil, nil
	}

	m.stats.Hits++
	m.updateHitRate()
	return entry, nil
}

// Set stores cache entry (in-memory implementation)
func (m *MemoryCache) Set(ctx context.Context, key string, entry *CacheEntry, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check size limit
	if len(m.data) >= m.maxSize {
		// Simple LRU: remove oldest entry
		var oldestKey string
		var oldestTime time.Time = time.Now()

		for k, v := range m.data {
			if v.ExpiresAt.Before(oldestTime) {
				oldestTime = v.ExpiresAt
				oldestKey = k
			}
		}

		if oldestKey != "" {
			delete(m.data, oldestKey)
			m.stats.Evictions++
		}
	}

	entry.ExpiresAt = time.Now().Add(ttl)
	m.data[key] = entry
	m.stats.Sets++
	m.stats.Size = int64(len(m.data))

	return nil
}

// Delete removes cache entry (in-memory implementation)
func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[key]; exists {
		delete(m.data, key)
		m.stats.Deletes++
		m.stats.Size = int64(len(m.data))
	}

	return nil
}

// Clear clears entire cache (in-memory implementation)
func (m *MemoryCache) Clear(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = make(map[string]*CacheEntry)
	m.stats.Size = 0

	return nil
}

// Stats returns cache statistics (in-memory implementation)
func (m *MemoryCache) Stats() CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := m.stats
	stats.Size = int64(len(m.data))
	return stats
}

// updateHitRate updates hit rate
func (m *MemoryCache) updateHitRate() {
	total := m.stats.Hits + m.stats.Misses
	if total > 0 {
		m.stats.HitRate = float64(m.stats.Hits) / float64(total) * 100
	}
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(config RedisConfig, prefix string) (*RedisCache, error) {
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

	return &RedisCache{
		client: client,
		prefix: prefix,
		stats:  CacheStats{},
	}, nil
}

// Get retrieves cache entry (Redis implementation)
func (r *RedisCache) Get(ctx context.Context, key string) (*CacheEntry, error) {
	fullKey := r.prefix + key

	data, err := r.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			r.stats.Misses++
			r.updateHitRate()
			return nil, nil
		}
		return nil, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("error deserializing cache: %v", err)
	}

	// Check if expired (double verification)
	if time.Now().After(entry.ExpiresAt) {
		r.client.Del(ctx, fullKey)
		r.stats.Misses++
		r.stats.Evictions++
		r.updateHitRate()
		return nil, nil
	}

	r.stats.Hits++
	r.updateHitRate()
	return &entry, nil
}

// Set stores cache entry (Redis implementation)
func (r *RedisCache) Set(ctx context.Context, key string, entry *CacheEntry, ttl time.Duration) error {
	fullKey := r.prefix + key

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("error serializing cache: %v", err)
	}

	if err := r.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return err
	}

	r.stats.Sets++
	return nil
}

// Delete removes cache entry (Redis implementation)
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.prefix + key

	result := r.client.Del(ctx, fullKey)
	if result.Val() > 0 {
		r.stats.Deletes++
	}

	return result.Err()
}

// Clear clears entire cache (Redis implementation)
func (r *RedisCache) Clear(ctx context.Context) error {
	pattern := r.prefix + "*"

	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// Stats returns cache statistics (Redis implementation)
func (r *RedisCache) Stats() CacheStats {
	// For Redis, some statistics may be limited
	stats := r.stats

	// Try to get Redis information
	ctx := context.Background()
	info, err := r.client.Info(ctx, "memory").Result()
	if err == nil {
		// Basic parsing of memory information
		lines := strings.Split(info, "\r\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "used_memory:") {
				if size, err := strconv.ParseInt(strings.TrimPrefix(line, "used_memory:"), 10, 64); err == nil {
					stats.Size = size
				}
			}
		}
	}

	r.updateHitRate()
	stats.HitRate = r.stats.HitRate

	return stats
}

// updateHitRate updates hit rate (Redis)
func (r *RedisCache) updateHitRate() {
	total := r.stats.Hits + r.stats.Misses
	if total > 0 {
		r.stats.HitRate = float64(r.stats.Hits) / float64(total) * 100
	}
}

// CacheMiddleware creates cache middleware
func CacheMiddleware(config *CacheConfig, keyGen CacheKeyFunc) gin.HandlerFunc {
	var store CacheStore
	var err error

	// Choose implementation based on configuration
	if config.Type == "redis" {
		redisConfig := DefaultConfig().Redis
		store, err = NewRedisCache(redisConfig, "gin_decorators:")
		if err != nil {
			// Fallback to memory if Redis fails
			store = NewMemoryCache(config.MaxSize)
		}
	} else {
		store = NewMemoryCache(config.MaxSize)
	}

	// Parse default TTL
	defaultTTL, err := time.ParseDuration(config.DefaultTTL)
	if err != nil {
		defaultTTL = 5 * time.Minute
	}

	return func(c *gin.Context) {
		// Only cache GET methods by default
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Generate cache key
		key := keyGen(c)

		// Try to retrieve from cache
		ctx := c.Request.Context()
		entry, err := store.Get(ctx, key)
		if err == nil && entry != nil {
			// Cache hit - return cached response
			for headerKey, headerValue := range entry.Headers {
				c.Header(headerKey, headerValue)
			}
			c.Header("X-Cache", "HIT")
			c.Header("X-Cache-Key", generateCacheKeyHash(key))

			c.Data(entry.Status, c.GetHeader("Content-Type"), entry.Data)
			c.Abort()
			return
		}

		// Cache miss - continue processing
		c.Header("X-Cache", "MISS")
		c.Header("X-Cache-Key", generateCacheKeyHash(key))

		// Capture response
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
			headers:        make(map[string]string),
		}
		c.Writer = writer

		c.Next()

		// Store in cache if response is successful
		if writer.status >= 200 && writer.status < 300 {
			entry := &CacheEntry{
				Data:    writer.body,
				Headers: writer.headers,
				Status:  writer.status,
			}

			if err := store.Set(ctx, key, entry, defaultTTL); err != nil {
				// Log error but don't fail the request
				log.Printf("Failed to store cache entry: %v", err)
			}
		}
	}
}

// responseWriter wrapper to capture response
type responseWriter struct {
	gin.ResponseWriter
	body    []byte
	headers map[string]string
	status  int
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body = append(w.body, data...)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Header() http.Header {
	// Capture important headers
	header := w.ResponseWriter.Header()
	for key, values := range header {
		if len(values) > 0 {
			w.headers[key] = values[0]
		}
	}
	return header
}

// CacheByURL cache middleware by URL
func CacheByURL(config *CacheConfig) gin.HandlerFunc {
	return CacheMiddleware(config, URLCacheKey)
}

// CacheByUserURL cache middleware by user and URL
func CacheByUserURL(config *CacheConfig) gin.HandlerFunc {
	return CacheMiddleware(config, UserURLCacheKey)
}

// CacheByEndpoint cache middleware by endpoint
func CacheByEndpoint(config *CacheConfig) gin.HandlerFunc {
	return CacheMiddleware(config, EndpointCacheKey)
}

// CustomCache customizable cache middleware
func CustomCache(ttl time.Duration, keyGen CacheKeyFunc, cacheType string) gin.HandlerFunc {
	config := &CacheConfig{
		Type:       cacheType,
		DefaultTTL: ttl.String(),
		MaxSize:    1000,
	}

	return CacheMiddleware(config, keyGen)
}

// ParseCacheArgs parses @Cache decorator arguments
func ParseCacheArgs(args []string) (time.Duration, string, CacheKeyFunc) {
	duration := 5 * time.Minute // default
	cacheType := "memory"       // default
	keyGen := URLCacheKey       // default

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)

			switch key {
			case "duration", "ttl":
				if parsed, err := time.ParseDuration(value); err == nil {
					duration = parsed
				}
			case "type":
				cacheType = value
			case "key", "by":
				switch value {
				case "url":
					keyGen = URLCacheKey
				case "user":
					keyGen = UserURLCacheKey
				case "endpoint":
					keyGen = EndpointCacheKey
				}
			}
		}
	}

	return duration, cacheType, keyGen
}

// generateCacheKeyHash generates MD5 hash of the key for headers
func generateCacheKeyHash(key string) string {
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", hash)[:8] // First 8 characters
}

// CacheStatsHandler handler for cache statistics
func CacheStatsHandler(store CacheStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := store.Stats()
		c.JSON(http.StatusOK, gin.H{
			"cache_stats": stats,
		})
	}
}

// InvalidateCacheHandler handler to invalidate cache
func InvalidateCacheHandler(store CacheStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Query("key")

		if key == "" {
			// Clear entire cache
			if err := store.Clear(c.Request.Context()); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to clear cache",
					"details": err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Cache completely cleared",
			})
			return
		}

		// Clear specific key
		if err := store.Delete(c.Request.Context(), key); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to invalidate cache",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Cache invalidated for key: %s", key),
		})
	}
}
