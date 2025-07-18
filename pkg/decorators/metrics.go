package decorators

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsCollector collects custom metrics
type MetricsCollector struct {
	// HTTP request metrics
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpRequestSize     *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec
	httpActiveRequests  *prometheus.GaugeVec

	// Middleware metrics
	middlewareExecutionTime *prometheus.HistogramVec
	middlewareErrors        *prometheus.CounterVec

	// Cache metrics
	cacheHits   *prometheus.CounterVec
	cacheMisses *prometheus.CounterVec
	cacheSize   *prometheus.GaugeVec

	// Rate limiting metrics
	rateLimitHits     *prometheus.CounterVec
	rateLimitExceeded *prometheus.CounterVec

	// Validation metrics
	validationErrors *prometheus.CounterVec
	validationTime   *prometheus.HistogramVec

	// System metrics
	gorutines       prometheus.Gauge
	memoryAllocated prometheus.Gauge
}

// DefaultMetricsCollector global instance default
var defaultMetricsCollector *MetricsCollector

// InitMetrics initializes metrics system
func InitMetrics(config *MetricsConfig) *MetricsCollector {
	collector := &MetricsCollector{
		// HTTP metrics
		httpRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status", "handler"},
		),

		httpRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_request_duration_seconds",
				Help:      "Duration of HTTP requests in seconds",
				Buckets:   config.Buckets,
			},
			[]string{"method", "endpoint", "status"},
		),

		httpRequestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_request_size_bytes",
				Help:      "Size of HTTP requests in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 5),
			},
			[]string{"method", "endpoint"},
		),

		httpResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_response_size_bytes",
				Help:      "Size of HTTP responses in bytes",
				Buckets:   prometheus.ExponentialBuckets(100, 10, 5),
			},
			[]string{"method", "endpoint", "status"},
		),

		httpActiveRequests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "http_active_requests",
				Help:      "Number of active HTTP requests",
			},
			[]string{"method", "endpoint"},
		),

		// Middleware metrics
		middlewareExecutionTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "middleware_execution_time_seconds",
				Help:      "Time spent executing middlewares",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"middleware", "endpoint"},
		),

		middlewareErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "middleware_errors_total",
				Help:      "Total number of middleware errors",
			},
			[]string{"middleware", "error_type"},
		),

		// Cache metrics
		cacheHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"cache_type", "key_type"},
		),

		cacheMisses: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"cache_type", "key_type"},
		),

		cacheSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "cache_size",
				Help:      "Current cache size",
			},
			[]string{"cache_type"},
		),

		// Rate limiting metrics
		rateLimitHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "rate_limit_hits_total",
				Help:      "Total number of rate limit checks",
			},
			[]string{"endpoint", "limit_type"},
		),

		rateLimitExceeded: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "rate_limit_exceeded_total",
				Help:      "Total number of rate limit exceeded",
			},
			[]string{"endpoint", "limit_type"},
		),

		// Validation metrics
		validationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "validation_errors_total",
				Help:      "Total number of validation errors",
			},
			[]string{"validation_type", "field"},
		),

		validationTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "validation_time_seconds",
				Help:      "Time spent validating requests",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"validation_type"},
		),

		// System metrics
		gorutines: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "goroutines",
				Help:      "Number of goroutines",
			},
		),

		memoryAllocated: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: config.Namespace,
				Subsystem: config.Subsystem,
				Name:      "memory_allocated_bytes",
				Help:      "Memory allocated in bytes",
			},
		),
	}

	// Register metrics
	prometheus.MustRegister(
		collector.httpRequestsTotal,
		collector.httpRequestDuration,
		collector.httpRequestSize,
		collector.httpResponseSize,
		collector.httpActiveRequests,
		collector.middlewareExecutionTime,
		collector.middlewareErrors,
		collector.cacheHits,
		collector.cacheMisses,
		collector.cacheSize,
		collector.rateLimitHits,
		collector.rateLimitExceeded,
		collector.validationErrors,
		collector.validationTime,
		collector.gorutines,
		collector.memoryAllocated,
	)

	defaultMetricsCollector = collector
	return collector
}

// MetricsMiddleware main middleware for metrics collection
func MetricsMiddleware(config *MetricsConfig) gin.HandlerFunc {
	if !config.Enabled {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	// Initialize coletor se not existir
	if defaultMetricsCollector == nil {
		InitMetrics(config)
	}

	return func(c *gin.Context) {
		start := time.Now()

		// Increment active requests
		endpoint := getEndpointPattern(c)
		method := c.Request.Method

		defaultMetricsCollector.httpActiveRequests.WithLabelValues(method, endpoint).Inc()

		// Register tamanho da request
		if c.Request.ContentLength > 0 {
			defaultMetricsCollector.httpRequestSize.WithLabelValues(method, endpoint).Observe(float64(c.Request.ContentLength))
		}

		// Capturar response
		writer := &metricsResponseWriter{
			ResponseWriter: c.Writer,
			size:           0,
		}
		c.Writer = writer

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)
		status := strconv.Itoa(writer.status)

		// Register metrics
		defaultMetricsCollector.httpRequestsTotal.WithLabelValues(method, endpoint, status, "unknown").Inc()
		defaultMetricsCollector.httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
		defaultMetricsCollector.httpResponseSize.WithLabelValues(method, endpoint, status).Observe(float64(writer.size))

		// Decrement active requests
		defaultMetricsCollector.httpActiveRequests.WithLabelValues(method, endpoint).Dec()
	}
}

// metricsResponseWriter wrapper to capture response size
type metricsResponseWriter struct {
	gin.ResponseWriter
	size   int
	status int
}

func (w *metricsResponseWriter) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.size += size
	return size, err
}

func (w *metricsResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// getEndpointPattern extracts endpoint pattern
func getEndpointPattern(c *gin.Context) string {
	// Use FullPath() if available, otherwise use Path
	if fullPath := c.FullPath(); fullPath != "" {
		return fullPath
	}
	return c.Request.URL.Path
}

// RecordCacheHit registra hit de cache
func RecordCacheHit(cacheType, keyType string) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.cacheHits.WithLabelValues(cacheType, keyType).Inc()
	}
}

// RecordCacheMiss registra miss de cache
func RecordCacheMiss(cacheType, keyType string) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.cacheMisses.WithLabelValues(cacheType, keyType).Inc()
	}
}

// RecordCacheSize registra tamanho do cache
func RecordCacheSize(cacheType string, size float64) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.cacheSize.WithLabelValues(cacheType).Set(size)
	}
}

// RecordRateLimitHit records rate limit check
func RecordRateLimitHit(endpoint, limitType string) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.rateLimitHits.WithLabelValues(endpoint, limitType).Inc()
	}
}

// RecordRateLimitExceeded registra rate limit excedido
func RecordRateLimitExceeded(endpoint, limitType string) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.rateLimitExceeded.WithLabelValues(endpoint, limitType).Inc()
	}
}

// RecordValidationError records validation error
func RecordValidationError(validationType, field string) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.validationErrors.WithLabelValues(validationType, field).Inc()
	}
}

// RecordValidationTime records validation time
func RecordValidationTime(validationType string, duration time.Duration) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.validationTime.WithLabelValues(validationType).Observe(duration.Seconds())
	}
}

// RecordMiddlewareTime records middleware execution time
func RecordMiddlewareTime(middleware, endpoint string, duration time.Duration) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.middlewareExecutionTime.WithLabelValues(middleware, endpoint).Observe(duration.Seconds())
	}
}

// RecordMiddlewareError records middleware error
func RecordMiddlewareError(middleware, errorType string) {
	if defaultMetricsCollector != nil {
		defaultMetricsCollector.middlewareErrors.WithLabelValues(middleware, errorType).Inc()
	}
}

// PrometheusHandler returns Prometheus handler
func PrometheusHandler() gin.HandlerFunc {
	handler := promhttp.Handler()
	return func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthCheckHandler health check handler with metrics
func HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Basic health information
		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"service":   "gin-decorators",
		}

		// Add metrics if available
		if defaultMetricsCollector != nil {
			// Get some basic metrics via registry
			metricFamilies, err := prometheus.DefaultGatherer.Gather()
			if err == nil {
				metrics := make(map[string]interface{})

				for _, mf := range metricFamilies {
					if strings.Contains(mf.GetName(), "http_requests_total") {
						var total float64
						for _, metric := range mf.GetMetric() {
							total += metric.GetCounter().GetValue()
						}
						metrics["total_requests"] = total
					}
				}

				health["metrics"] = metrics
			}
		}

		c.JSON(http.StatusOK, health)
	}
}

// createMetricsMiddleware creates metrics middleware with customizable settings via args
func createMetricsMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Metrics

	// Parse custom settings from args
	for _, arg := range args {
		if strings.HasPrefix(arg, "namespace=") {
			v := strings.TrimPrefix(arg, "namespace=")
			config.Namespace = v
		}
		if strings.HasPrefix(arg, "subsystem=") {
			v := strings.TrimPrefix(arg, "subsystem=")
			config.Subsystem = v
		}
		if strings.HasPrefix(arg, "endpoint=") {
			v := strings.TrimPrefix(arg, "endpoint=")
			config.Endpoint = v
		}
		if strings.HasPrefix(arg, "enabled=") {
			v := strings.TrimPrefix(arg, "enabled=")
			if enabled, err := strconv.ParseBool(v); err == nil {
				config.Enabled = enabled
			}
		}
	}

	return MetricsMiddleware(&config)
}

// MetricsInfo information about available metrics
type MetricsInfo struct {
	Enabled   bool     `json:"enabled"`
	Endpoint  string   `json:"endpoint"`
	Namespace string   `json:"namespace"`
	Subsystem string   `json:"subsystem"`
	Metrics   []string `json:"metrics"`
}

// GetMetricsInfo returns information about metrics
func GetMetricsInfo(config *MetricsConfig) MetricsInfo {
	metrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"http_request_size_bytes",
		"http_response_size_bytes",
		"http_active_requests",
		"middleware_execution_time_seconds",
		"middleware_errors_total",
		"cache_hits_total",
		"cache_misses_total",
		"cache_size",
		"rate_limit_hits_total",
		"rate_limit_exceeded_total",
		"validation_errors_total",
		"validation_time_seconds",
		"goroutines",
		"memory_allocated_bytes",
	}

	return MetricsInfo{
		Enabled:   config.Enabled,
		Endpoint:  config.Endpoint,
		Namespace: config.Namespace,
		Subsystem: config.Subsystem,
		Metrics:   metrics,
	}
}
