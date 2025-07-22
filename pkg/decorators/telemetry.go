package decorators

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// TelemetryManager manages OpenTelemetry configuration and instrumentation
type TelemetryManager struct {
	tracer   trace.Tracer
	config   TelemetryConfig
	provider *sdktrace.TracerProvider
}

// TracingInfo information about tracing for documentation
type TracingInfo struct {
	Enabled        bool              `json:"enabled"`
	ServiceName    string            `json:"service_name"`
	ServiceVersion string            `json:"service_version"`
	Environment    string            `json:"environment"`
	Endpoint       string            `json:"endpoint"`
	SampleRate     float64           `json:"sample_rate"`
	Attributes     map[string]string `json:"attributes"`
}

// defaultTelemetryManager global instance
var (
	defaultTelemetryManager *TelemetryManager
	telemetryMutex          sync.RWMutex
)

// InitTelemetry initializes OpenTelemetry
func InitTelemetry(config *TelemetryConfig) (*TelemetryManager, error) {
	if !config.Enabled {
		return &TelemetryManager{config: *config}, nil
	}

	// Configure resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating resource: %v", err)
	}

	// Configure exporter OTLP
	var opts []otlptracehttp.Option
	opts = append(opts, otlptracehttp.WithEndpoint(config.Endpoint))
	if config.Insecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	exporter, err := otlptracehttp.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("error creating exporter: %v", err)
	}

	// Configure trace provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SampleRate)),
	)

	// Configure propagation
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Create tracer
	tracer := otel.Tracer("gin-decorators")

	manager := &TelemetryManager{
		tracer:   tracer,
		config:   *config,
		provider: provider,
	}

	telemetryMutex.Lock()
	defaultTelemetryManager = manager
	telemetryMutex.Unlock()
	return manager, nil
}

// Shutdown finaliza telemetria
func (tm *TelemetryManager) Shutdown(ctx context.Context) error {
	if tm.provider != nil {
		return tm.provider.Shutdown(ctx)
	}
	return nil
}

// TracingMiddleware main tracing middleware
func TracingMiddleware(config *TelemetryConfig) gin.HandlerFunc {
	if !config.Enabled {
		return gin.HandlerFunc(func(c *gin.Context) {
			c.Next()
		})
	}

	// Initialize if necessary
	telemetryMutex.RLock()
	manager := defaultTelemetryManager
	telemetryMutex.RUnlock()

	if manager == nil {
		telemetryMutex.Lock()
		// Double-check after acquiring lock
		if defaultTelemetryManager == nil {
			var err error
			manager, err = InitTelemetry(config)
			if err != nil {
				// Log error and continue without tracing
				fmt.Printf("Error ao inicializar telemetria: %v\n", err)
				telemetryMutex.Unlock()
				return gin.HandlerFunc(func(c *gin.Context) {
					c.Next()
				})
			}
			defaultTelemetryManager = manager
		} else {
			manager = defaultTelemetryManager
		}
		telemetryMutex.Unlock()
	}

	return func(c *gin.Context) {
		// Extract contexto de tracing dos headers
		ctx := otel.GetTextMapPropagator().Extract(
			c.Request.Context(),
			propagation.HeaderCarrier(c.Request.Header),
		)

		// Create span
		spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
		if c.FullPath() == "" {
			spanName = fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
		}

		ctx, span := manager.tracer.Start(ctx, spanName)
		defer span.End()

		// Add atributos ao span
		span.SetAttributes(
			semconv.HTTPMethod(c.Request.Method),
			semconv.HTTPTarget(c.Request.URL.Path),
			semconv.HTTPRoute(c.FullPath()),
			semconv.HTTPScheme(c.Request.URL.Scheme),
			attribute.String("http.host", c.Request.Host),
			semconv.HTTPUserAgent(c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		// Add headers customizados
		if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
			span.SetAttributes(attribute.String("http.request.id", requestID))
		}

		if userID := c.GetString("user_id"); userID != "" {
			span.SetAttributes(attribute.String("user.id", userID))
		}

		// Update context in request
		c.Request = c.Request.WithContext(ctx)

		// Inject tracing headers in response
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(c.Writer.Header()))

		// Continue processing
		c.Next()

		// Add response information
		span.SetAttributes(
			semconv.HTTPStatusCode(c.Writer.Status()),
			attribute.Int("http.response.size", c.Writer.Size()),
		)

		// Define span status based on HTTP code
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, http.StatusText(c.Writer.Status()))
		} else {
			span.SetStatus(codes.Ok, "")
		}

		// Add errors se houver
		if len(c.Errors) > 0 {
			span.SetStatus(codes.Error, c.Errors.String())
			span.SetAttributes(attribute.String("error.message", c.Errors.String()))
		}
	}
}

// StartSpan starts a new span
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	telemetryMutex.RLock()
	manager := defaultTelemetryManager
	telemetryMutex.RUnlock()

	if manager == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return manager.tracer.Start(ctx, name)
}

// SpanFromContext extracts span from context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// AddSpanAttributes adds attributes to current span
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds event to current span
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetSpanError marca span como error
func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetStatus(codes.Error, err.Error())
		span.SetAttributes(attribute.String("error.message", err.Error()))
	}
}

// TraceMiddleware instrumenta middleware individual
func TraceMiddleware(middlewareName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if defaultTelemetryManager == nil {
			c.Next()
			return
		}

		ctx, span := StartSpan(c.Request.Context(), fmt.Sprintf("middleware.%s", middlewareName))
		defer span.End()

		span.SetAttributes(
			attribute.String("middleware.name", middlewareName),
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", c.FullPath()),
		)

		c.Request = c.Request.WithContext(ctx)
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		span.SetAttributes(
			attribute.Float64("middleware.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		if len(c.Errors) > 0 {
			SetSpanError(ctx, c.Errors.Last())
		}
	}
}

// TraceCacheOperation instruments cache operations
func TraceCacheOperation(ctx context.Context, operation, cacheType, key string) (context.Context, trace.Span) {
	if defaultTelemetryManager == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := StartSpan(ctx, fmt.Sprintf("cache.%s", operation))
	span.SetAttributes(
		attribute.String("cache.operation", operation),
		attribute.String("cache.type", cacheType),
		attribute.String("cache.key", key),
	)

	return ctx, span
}

// TraceRateLimitOperation instruments rate limit operations
func TraceRateLimitOperation(ctx context.Context, operation, limitType string, allowed bool) (context.Context, trace.Span) {
	if defaultTelemetryManager == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := StartSpan(ctx, fmt.Sprintf("ratelimit.%s", operation))
	span.SetAttributes(
		attribute.String("ratelimit.operation", operation),
		attribute.String("ratelimit.type", limitType),
		attribute.Bool("ratelimit.allowed", allowed),
	)

	return ctx, span
}

// TraceValidationOperation instruments validation operations
func TraceValidationOperation(ctx context.Context, validationType string, fieldCount int) (context.Context, trace.Span) {
	if defaultTelemetryManager == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := StartSpan(ctx, fmt.Sprintf("validation.%s", validationType))
	span.SetAttributes(
		attribute.String("validation.type", validationType),
		attribute.Int("validation.field_count", fieldCount),
	)

	return ctx, span
}

// TraceWebSocketOperation instruments WebSocket operations
func TraceWebSocketOperation(ctx context.Context, operation, connectionID string) (context.Context, trace.Span) {
	if defaultTelemetryManager == nil {
		return ctx, trace.SpanFromContext(ctx)
	}

	ctx, span := StartSpan(ctx, fmt.Sprintf("websocket.%s", operation))
	span.SetAttributes(
		attribute.String("websocket.operation", operation),
		attribute.String("websocket.connection_id", connectionID),
	)

	return ctx, span
}

// GetTracingInfo returns information about tracing configuration
func GetTracingInfo(config *TelemetryConfig) TracingInfo {
	info := TracingInfo{
		Enabled:        config.Enabled,
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		Environment:    config.Environment,
		Endpoint:       config.Endpoint,
		SampleRate:     config.SampleRate,
		Attributes:     make(map[string]string),
	}

	// Add atributos default
	info.Attributes["service.name"] = config.ServiceName
	info.Attributes["service.version"] = config.ServiceVersion
	info.Attributes["deployment.environment"] = config.Environment

	return info
}

// TracingStatsHandler handler for tracing statistics
func TracingStatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := map[string]interface{}{
			"enabled": false,
		}

		if defaultTelemetryManager != nil {
			stats["enabled"] = defaultTelemetryManager.config.Enabled
			stats["service_name"] = defaultTelemetryManager.config.ServiceName
			stats["service_version"] = defaultTelemetryManager.config.ServiceVersion
			stats["environment"] = defaultTelemetryManager.config.Environment
			stats["sample_rate"] = defaultTelemetryManager.config.SampleRate
		}

		c.JSON(http.StatusOK, gin.H{
			"tracing_stats": stats,
		})
	}
}

// createTelemetryMiddleware creates telemetry middleware with customizable settings via args
func createTelemetryMiddleware(args []string) gin.HandlerFunc {
	config := DefaultConfig().Telemetry

	// Parse custom settings from args
	for _, arg := range args {
		if strings.HasPrefix(arg, "sampleRate=") {
			v := strings.TrimPrefix(arg, "sampleRate=")
			if rate, err := strconv.ParseFloat(v, 64); err == nil && rate >= 0 && rate <= 1 {
				config.SampleRate = rate
			}
		}
		if strings.HasPrefix(arg, "serviceName=") {
			v := strings.TrimPrefix(arg, "serviceName=")
			config.ServiceName = v
		}
		if strings.HasPrefix(arg, "environment=") {
			v := strings.TrimPrefix(arg, "environment=")
			config.Environment = v
		}
		if strings.HasPrefix(arg, "endpoint=") {
			v := strings.TrimPrefix(arg, "endpoint=")
			config.Endpoint = v
		}
	}

	return TracingMiddleware(&config)
}

// HealthCheckWithTracing instrumented health check
func HealthCheckWithTracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := StartSpan(c.Request.Context(), "health_check")
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		// Verify components
		components := map[string]string{
			"server": "healthy",
		}

		if defaultTelemetryManager != nil {
			components["tracing"] = "healthy"
			span.SetAttributes(attribute.String("health.tracing", "enabled"))
		} else {
			components["tracing"] = "disabled"
			span.SetAttributes(attribute.String("health.tracing", "disabled"))
		}

		span.SetAttributes(
			attribute.String("health.status", "healthy"),
			attribute.Int("health.components", len(components)),
		)

		c.JSON(http.StatusOK, gin.H{
			"status":     "healthy",
			"timestamp":  time.Now().Unix(),
			"components": components,
			"trace_id":   span.SpanContext().TraceID().String(),
		})
	}
}

// InstrumentedHandler wrapper to instrument custom handlers
func InstrumentedHandler(handlerName string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := StartSpan(c.Request.Context(), fmt.Sprintf("handler.%s", handlerName))
		defer span.End()

		span.SetAttributes(
			attribute.String("handler.name", handlerName),
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", c.FullPath()),
		)

		c.Request = c.Request.WithContext(ctx)
		start := time.Now()

		handler(c)

		duration := time.Since(start)
		span.SetAttributes(
			attribute.Float64("handler.duration_ms", float64(duration.Nanoseconds())/1e6),
			attribute.Int("http.status_code", c.Writer.Status()),
		)

		if len(c.Errors) > 0 {
			SetSpanError(ctx, c.Errors.Last())
		}
	}
}
