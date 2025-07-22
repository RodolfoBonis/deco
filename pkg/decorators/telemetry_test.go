package decorators

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Tests for telemetry functionality

func TestInitTelemetry_Disabled(t *testing.T) {
	config := &TelemetryConfig{
		Enabled: false,
	}

	manager, err := InitTelemetry(config)
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.False(t, manager.config.Enabled)
}

func TestInitTelemetry_Enabled(t *testing.T) {
	config := &TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4318",
		Insecure:       true,
		SampleRate:     1.0,
	}

	manager, err := InitTelemetry(config)
	// This might fail if OTLP endpoint is not available, but we can test the structure
	if err == nil {
		assert.NotNil(t, manager)
		assert.True(t, manager.config.Enabled)
		assert.Equal(t, "test-service", manager.config.ServiceName)
		assert.Equal(t, "1.0.0", manager.config.ServiceVersion)
		assert.Equal(t, "test", manager.config.Environment)
	} else {
		// If it fails, it should be due to connection issues, not configuration
		assert.Contains(t, err.Error(), "error creating exporter")
	}
}

func TestTelemetryManager_Shutdown(t *testing.T) {
	manager := &TelemetryManager{
		config: TelemetryConfig{Enabled: false},
	}

	ctx := context.Background()
	err := manager.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestTracingMiddleware_Disabled(t *testing.T) {
	// Remove  to avoid race conditions with gin.SetMode()

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &TelemetryConfig{
		Enabled: false,
	}

	middleware := TracingMiddleware(config)
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTracingMiddleware_Enabled(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4318",
		Insecure:       true,
		SampleRate:     1.0,
	}

	middleware := TracingMiddleware(config)
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	req.Header.Set("X-Request-ID", "test-request-id")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	spanCtx, span := StartSpan(ctx, "test-span")

	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)

	// Clean up
	if span != nil {
		span.End()
	}
}

func TestSpanFromContext(t *testing.T) {
	ctx := context.Background()
	span := SpanFromContext(ctx)

	// Should return a no-op span when no span is in context
	assert.NotNil(t, span)
}

func TestAddSpanAttributes(t *testing.T) {
	ctx := context.Background()

	// This should not panic even without a real span
	assert.NotPanics(t, func() {
		AddSpanAttributes(ctx)
	})
}

func TestAddSpanEvent(t *testing.T) {
	ctx := context.Background()

	// This should not panic even without a real span
	assert.NotPanics(t, func() {
		AddSpanEvent(ctx, "test-event")
	})
}

func TestSetSpanError(t *testing.T) {
	ctx := context.Background()

	// This should not panic even without a real span
	assert.NotPanics(t, func() {
		SetSpanError(ctx, assert.AnError)
	})
}

func TestTraceMiddleware(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	middleware := TraceMiddleware("test-middleware")
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTraceCacheOperation(t *testing.T) {
	ctx := context.Background()
	spanCtx, span := TraceCacheOperation(ctx, "get", "memory", "test-key")

	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)

	// Clean up
	if span != nil {
		span.End()
	}
}

func TestTraceRateLimitOperation(t *testing.T) {
	ctx := context.Background()
	spanCtx, span := TraceRateLimitOperation(ctx, "check", "ip", true)

	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)

	// Clean up
	if span != nil {
		span.End()
	}
}

func TestTraceValidationOperation(t *testing.T) {
	ctx := context.Background()
	spanCtx, span := TraceValidationOperation(ctx, "json", 5)

	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)

	// Clean up
	if span != nil {
		span.End()
	}
}

func TestTraceWebSocketOperation(t *testing.T) {
	ctx := context.Background()
	spanCtx, span := TraceWebSocketOperation(ctx, "connect", "conn-123")

	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)

	// Clean up
	if span != nil {
		span.End()
	}
}

func TestGetTracingInfo(t *testing.T) {
	config := &TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4318",
		SampleRate:     0.5,
	}

	info := GetTracingInfo(config)
	assert.True(t, info.Enabled)
	assert.Equal(t, "test-service", info.ServiceName)
	assert.Equal(t, "1.0.0", info.ServiceVersion)
	assert.Equal(t, "test", info.Environment)
	assert.Equal(t, "localhost:4318", info.Endpoint)
	assert.Equal(t, 0.5, info.SampleRate)
	assert.NotNil(t, info.Attributes)
}

func TestTracingStatsHandler(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := TracingStatsHandler()
	router.GET("/tracing/stats", handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tracing/stats", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateTelemetryMiddleware(t *testing.T) {
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
			name: "with service name",
			args: []string{"service=test-service"},
		},
		{
			name: "with endpoint",
			args: []string{"endpoint=localhost:4318"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := createTelemetryMiddleware(tt.args)
			assert.NotNil(t, middleware)

			// Test that middleware can be called without panic
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/test", http.NoBody)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			assert.NotPanics(t, func() {
				middleware(c)
			})
		})
	}
}

func TestHealthCheckWithTracing(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := HealthCheckWithTracing()
	router.GET("/health", handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInstrumentedHandler(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	originalHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}

	instrumentedHandler := InstrumentedHandler("test-handler", originalHandler)
	router.GET("/test", instrumentedHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTelemetryConfig_Structure(t *testing.T) {
	config := &TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "production",
		Endpoint:       "localhost:4318",
		Insecure:       true,
		SampleRate:     0.5,
	}

	assert.True(t, config.Enabled)
	assert.Equal(t, "test-service", config.ServiceName)
	assert.Equal(t, "1.0.0", config.ServiceVersion)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "localhost:4318", config.Endpoint)
	assert.True(t, config.Insecure)
	assert.Equal(t, 0.5, config.SampleRate)
}

func TestTracingInfo_Structure(t *testing.T) {
	info := TracingInfo{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4318",
		SampleRate:     1.0,
		Attributes: map[string]string{
			"version": "1.0.0",
			"env":     "test",
		},
	}

	assert.True(t, info.Enabled)
	assert.Equal(t, "test-service", info.ServiceName)
	assert.Equal(t, "1.0.0", info.ServiceVersion)
	assert.Equal(t, "test", info.Environment)
	assert.Equal(t, "localhost:4318", info.Endpoint)
	assert.Equal(t, 1.0, info.SampleRate)
	assert.Equal(t, "1.0.0", info.Attributes["version"])
	assert.Equal(t, "test", info.Attributes["env"])
}

func TestTelemetryManager_Structure(t *testing.T) {
	manager := &TelemetryManager{
		config: TelemetryConfig{
			Enabled:     true,
			ServiceName: "test-service",
		},
	}

	assert.True(t, manager.config.Enabled)
	assert.Equal(t, "test-service", manager.config.ServiceName)
}

func TestTracingMiddleware_WithUserContext(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4318",
		Insecure:       true,
		SampleRate:     1.0,
	}

	middleware := TracingMiddleware(config)
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.Set("user_id", "user123")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTracingMiddleware_WithErrors(t *testing.T) {
	// Remove  to avoid race conditions

	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := &TelemetryConfig{
		Enabled:        true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Endpoint:       "localhost:4318",
		Insecure:       true,
		SampleRate:     1.0,
	}

	middleware := TracingMiddleware(config)
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "test error"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", http.NoBody)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
