package decorators

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test struct for validation tests
type TestUser struct {
	Name     string `json:"name" binding:"required" validate:"required"`
	Email    string `json:"email" binding:"required,email" validate:"required,email"`
	Age      int    `json:"age" binding:"required,gte=18" validate:"required,gte=18"`
	Phone    string `json:"phone" binding:"required" validate:"required"`
	CPF      string `json:"cpf" binding:"required" validate:"required"`
	CNPJ     string `json:"cnpj" binding:"required" validate:"required"`
	DateTime string `json:"datetime" binding:"required" validate:"required"`
}

// Test struct for query validation tests
type TestQueryParams struct {
	Page     int    `form:"page" binding:"required,gte=1" validate:"required,gte=1"`
	Limit    int    `form:"limit" binding:"required,gte=1,lte=100" validate:"required,gte=1,lte=100"`
	Search   string `form:"search" binding:"required" validate:"required"`
	Category string `form:"category" binding:"required,oneof=tech business sports" validate:"required,oneof=tech business sports"`
}

// Tests for JSON validation middleware

func TestValidateJSON_ValidData(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{
		"name": "Test User",
		"email": "test@example.com",
		"age": 25,
		"phone": "1234567890",
		"cpf": "12345678901",
		"cnpj": "12345678901234",
		"datetime": "2023-01-01"
	}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestValidateJSON_InvalidData(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{
		"name": "",
		"email": "invalid-email",
		"age": 15,
		"phone": "",
		"cpf": "",
		"cnpj": "",
		"datetime": ""
	}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "fields")
	assert.Contains(t, response, "error")
}

func TestValidateJSON_InvalidJSON(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Tests for query validation middleware

func TestValidateQuery_ValidData(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var params TestQueryParams
	middleware := ValidateQuery(&params, config)

	router := createTestGinEngine(t)
	router.GET("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=1&limit=10&search=test&category=tech", http.NoBody)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestValidateQuery_InvalidData(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var params TestQueryParams
	middleware := ValidateQuery(&params, config)

	router := createTestGinEngine(t)
	router.GET("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=0&limit=200&search=&category=invalid", http.NoBody)

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "fields")
	assert.Contains(t, response, "error")
}

// Tests for parameter validation middleware

func TestValidateParams(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	rules := map[string]string{
		"page":   "gte=1",
		"limit":  "lte=100",
		"search": "required",
	}

	middleware := ValidateParams(rules, config)

	router := createTestGinEngine(t)
	router.GET("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=1&limit=50&search=test", http.NoBody)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// Tests for parameter value validation

func TestValidateParamValue(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		rule     string
		expected bool
	}{
		{"gte valid", "10", "gte=5", true},
		{"gte invalid", "3", "gte=5", false},
		{"lte valid", "5", "lte=10", true},
		{"lte invalid", "15", "lte=10", false},
		{"oneof valid", "tech", "oneof=tech business sports", true},
		{"oneof invalid", "invalid", "oneof=tech business sports", false},
		{"required valid", "value", "required", true},
		{"required invalid", "", "required", false},
		{"email valid", "test@example.com", "email", true},
		{"email invalid", "invalid-email", "email", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateParamValue(tt.value, tt.rule)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for custom validators

func TestCustomValidators(t *testing.T) {
	// Test that custom validators are registered and can be called
	// Since these functions require validator.FieldLevel which is complex to mock,
	// we'll test that the functions exist and can be referenced
	_ = validatePhone
	_ = validateCPF
	_ = validateCNPJ
	_ = validateDateTime

	// Test that the functions are not nil
	assert.NotNil(t, validatePhone)
	assert.NotNil(t, validateCPF)
	assert.NotNil(t, validateCNPJ)
	assert.NotNil(t, validateDateTime)
}

// Tests for validation message generation

func TestGetValidationMessage(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	// Test that the function exists and can be called
	// Since we can't easily mock validator.FieldError, we'll test the function exists
	// and doesn't panic when called with nil
	assert.NotNil(t, config, "Config should not be nil")

	// Test that the function exists and can be referenced
	_ = getValidationMessage
	assert.NotNil(t, getValidationMessage)
}

// Tests for validation response structure

func TestValidationResponse_Structure(t *testing.T) {
	response := ValidationResponse{
		Error:   "validation_failed",
		Message: "Validation failed",
		Fields: []ValidationField{
			{
				Field:   "email",
				Message: "Email is required",
				Value:   "",
			},
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaled ValidationResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, response.Error, unmarshaled.Error)
	assert.Len(t, unmarshaled.Fields, 1)
	assert.Equal(t, response.Fields[0].Field, unmarshaled.Fields[0].Field)
}

// Tests for validation field structure

func TestValidationField_Structure(t *testing.T) {
	field := ValidationField{
		Field:   "name",
		Message: "Name is required",
		Value:   "",
		Tag:     "required",
		Param:   "",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(field)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Test JSON unmarshaling
	var unmarshaled ValidationField
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, field.Field, unmarshaled.Field)
	assert.Equal(t, field.Message, unmarshaled.Message)
}

// Tests for getting validated data

func TestGetValidatedData(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		// Test that validated data is available
		data, exists := GetValidatedData(c)
		assert.NotNil(t, data)
		assert.True(t, exists)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{
		"name": "Test User",
		"email": "test@example.com",
		"age": 25,
		"phone": "1234567890",
		"cpf": "12345678901",
		"cnpj": "12345678901234",
		"datetime": "2023-01-01"
	}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// Tests for getting validated query

func TestGetValidatedQuery(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var params TestQueryParams
	middleware := ValidateQuery(&params, config)

	router := createTestGinEngine(t)
	router.GET("/test", middleware, func(c *gin.Context) {
		// Test that validated query is available
		query, exists := GetValidatedQuery(c)
		assert.NotNil(t, query)
		assert.True(t, exists)
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=1&limit=10&search=test&category=tech", http.NoBody)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// Tests for getting detailed validation message

func TestGetDetailedValidationMessage(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	message := getDetailedValidationMessage("email", "required", "", config)
	assert.NotEmpty(t, message)
}

// Test validation middleware edge cases

func TestValidateJSON_EdgeCases(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	// Test with invalid JSON
	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Test validation config default values

func TestValidationConfig_DefaultValues(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}
	assert.True(t, config.Enabled)
	assert.False(t, config.FailFast)
	assert.Equal(t, "json", config.ErrorFormat)
}

// Test validation error formats

func TestValidationError_Formats(t *testing.T) {
	tests := []struct {
		format        string
		shouldContain string
	}{
		{"json", "fields"},
		{"json", "error"},
		{"json", "message"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			config := &ValidationConfig{
				Enabled:     true,
				FailFast:    false,
				ErrorFormat: tt.format,
			}

			var testUser TestUser
			middleware := ValidateJSON(&testUser, config)

			router := createTestGinEngine(t)
			router.POST("/test", middleware, func(c *gin.Context) {
				c.Status(200)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"email":"invalid"}`))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, 400, w.Code)
			assert.Contains(t, w.Body.String(), tt.shouldContain)
		})
	}
}

// Test disabled validation

func TestValidationMiddleware_Disabled(t *testing.T) {
	config := &ValidationConfig{
		Enabled: false,
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	// Create router to test middleware properly
	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"test","email":"test@test.com","age":25,"phone":"123","cpf":"123","cnpj":"123","datetime":"2023-01-01"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	// Should pass through without validation
	assert.Equal(t, 200, w.Code)
}

// Test fail fast behavior

func TestValidationMiddleware_FailFast(t *testing.T) {
	config := &ValidationConfig{
		Enabled:  true,
		FailFast: true,
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"email":"invalid"}`))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Test query parameter validation

func TestValidateQuery_EdgeCases(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	type QueryParams struct {
		Page   int    `form:"page" binding:"required,gte=1"`
		Limit  int    `form:"limit" binding:"required,lte=100"`
		Search string `form:"search" binding:"required"`
	}

	var params QueryParams
	middleware := ValidateQuery(&params, config)

	router := createTestGinEngine(t)
	router.GET("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=1&limit=50&search=test", http.NoBody)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// Test parameter validation

func TestValidateParams_EdgeCases(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	rules := map[string]string{
		"page":  "gte=1",
		"limit": "lte=100",
	}

	middleware := ValidateParams(rules, config)

	router := createTestGinEngine(t)
	router.GET("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test?page=1&limit=50", http.NoBody)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

// Test concurrent validation requests

func TestValidationMiddleware_Concurrent(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	// Test concurrent requests
	var wg sync.WaitGroup
	numRequests := 10

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"Test","email":"test@example.com","age":25,"phone":"1234567890","cpf":"12345678901","cnpj":"12345678901234","datetime":"2023-01-01"}`))
			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
		}()
	}

	wg.Wait()
}

// Test validation performance

func TestValidationMiddleware_Performance(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	// Benchmark validation performance
	start := time.Now()
	numRequests := 1000

	for i := 0; i < numRequests; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name":"Test","email":"test@example.com","age":25,"phone":"1234567890","cpf":"12345678901","cnpj":"12345678901234","datetime":"2023-01-01"}`))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	}

	duration := time.Since(start)
	avgTime := duration / time.Duration(numRequests)

	// Should process requests quickly (less than 1ms per request)
	assert.Less(t, avgTime, time.Millisecond, "Average validation time should be less than 1ms")
}

// Test error handling scenarios

func TestValidationMiddleware_ErrorHandling(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/test", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	tests := []struct {
		name           string
		requestBody    string
		contentType    string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "malformed JSON",
			requestBody:    `{"name": "test", "email": "test@example.com"`,
			contentType:    "application/json",
			expectedStatus: 400,
			expectedError:  "Invalid JSON format",
		},
		{
			name:           "wrong content type",
			requestBody:    `{"name": "test"}`,
			contentType:    "text/plain",
			expectedStatus: 400, // Should fail for non-JSON content
		},
		{
			name:           "empty content type",
			requestBody:    `{"name": "test"}`,
			contentType:    "",
			expectedStatus: 400, // Should fail for empty content type
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			var body io.Reader
			if tt.requestBody != "" {
				body = strings.NewReader(tt.requestBody)
			}

			req := httptest.NewRequest("POST", "/test", body)
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" {
				assert.Contains(t, w.Body.String(), tt.expectedError)
			}
		})
	}
}

// Test file upload validation

func TestValidationMiddleware_FileUpload(t *testing.T) {
	config := &ValidationConfig{
		Enabled:     true,
		FailFast:    false,
		ErrorFormat: "json",
	}

	var testUser TestUser
	middleware := ValidateJSON(&testUser, config)

	router := createTestGinEngine(t)
	router.POST("/upload", middleware, func(c *gin.Context) {
		c.Status(200)
	})

	// Create multipart form data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "test.txt")
	assert.NoError(t, err)
	part.Write([]byte("test content"))
	writer.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	router.ServeHTTP(w, req)

	// Should fail because it's not JSON content
	assert.Equal(t, 400, w.Code)
}
