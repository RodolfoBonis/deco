// Tests for decorators types and structures
package decorators

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		error    ValidationError
		expected string
	}{
		{
			name: "should format error with file and line",
			error: ValidationError{
				File:    "test.go",
				Line:    42,
				Message: "Invalid syntax",
				Code:    "SYNTAX_ERROR",
			},
			expected: "test.go:42 - Invalid syntax",
		},
		{
			name: "should format error without line",
			error: ValidationError{
				File:    "test.go",
				Line:    0,
				Message: "General error",
				Code:    "GENERAL_ERROR",
			},
			expected: "test.go - General error",
		},
		{
			name: "should handle empty file name",
			error: ValidationError{
				File:    "",
				Line:    10,
				Message: "Empty file error",
				Code:    "EMPTY_FILE",
			},
			expected: ":10 - Empty file error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.error.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRoute_Structure(t *testing.T) {
	route := &Route{
		Method:      "GET",
		Path:        "/test",
		Handler:     nil,
		Middlewares: []gin.HandlerFunc{},
	}

	assert.Equal(t, "GET", route.Method)
	assert.Equal(t, "/test", route.Path)
	assert.Nil(t, route.Handler)
	assert.Empty(t, route.Middlewares)
}

func TestMiddlewareInfo_Structure(t *testing.T) {
	middleware := &MiddlewareInfo{
		Name:        "auth",
		Args:        map[string]interface{}{"role": "admin"},
		Order:       1,
		Description: "Authentication middleware",
	}

	assert.Equal(t, "auth", middleware.Name)
	assert.Equal(t, "admin", middleware.Args["role"])
	assert.Equal(t, 1, middleware.Order)
	assert.Equal(t, "Authentication middleware", middleware.Description)
}

func TestFrameworkStats_Structure(t *testing.T) {
	stats := &FrameworkStats{
		TotalRoutes:       10,
		UniqueMiddlewares: 5,
		PackagesScanned:   3,
		BuildMode:         "development",
		GeneratedAt:       "2024-01-01T00:00:00Z",
		Methods:           map[string]int{"GET": 5, "POST": 3, "PUT": 2},
	}

	assert.Equal(t, 10, stats.TotalRoutes)
	assert.Equal(t, 5, stats.UniqueMiddlewares)
	assert.Equal(t, 3, stats.PackagesScanned)
	assert.Equal(t, "development", stats.BuildMode)
	assert.Equal(t, "2024-01-01T00:00:00Z", stats.GeneratedAt)
	assert.Equal(t, 5, stats.Methods["GET"])
	assert.Equal(t, 3, stats.Methods["POST"])
	assert.Equal(t, 2, stats.Methods["PUT"])
}

func TestParserStats_Structure(t *testing.T) {
	errors := []ValidationError{
		{File: "test.go", Line: 10, Message: "Error 1", Code: "ERR1"},
		{File: "test2.go", Line: 20, Message: "Error 2", Code: "ERR2"},
	}

	warnings := []ValidationError{
		{File: "test.go", Line: 15, Message: "Warning 1", Code: "WARN1"},
	}

	stats := &ParserStats{
		FilesProcessed:  5,
		RoutesFound:     10,
		MarkersApplied:  15,
		Errors:          errors,
		Warnings:        warnings,
		ProcessingTime:  "1.5s",
		SourceDirectory: "/test/path",
	}

	assert.Equal(t, 5, stats.FilesProcessed)
	assert.Equal(t, 10, stats.RoutesFound)
	assert.Equal(t, 15, stats.MarkersApplied)
	assert.Len(t, stats.Errors, 2)
	assert.Len(t, stats.Warnings, 1)
	assert.Equal(t, "1.5s", stats.ProcessingTime)
	assert.Equal(t, "/test/path", stats.SourceDirectory)
}

func TestSchemaInfo_Structure(t *testing.T) {
	properties := map[string]*PropertyInfo{
		"name": {
			Name:        "name",
			Type:        "string",
			Description: "User name",
			Required:    true,
		},
		"age": {
			Name:        "age",
			Type:        "integer",
			Description: "User age",
			Required:    false,
		},
	}

	schema := &SchemaInfo{
		Name:        "User",
		Description: "User entity",
		Type:        "object",
		Properties:  properties,
		Required:    []string{"name"},
		PackageName: "models",
		FileName:    "user.go",
	}

	assert.Equal(t, "User", schema.Name)
	assert.Equal(t, "User entity", schema.Description)
	assert.Equal(t, "object", schema.Type)
	assert.Len(t, schema.Properties, 2)
	assert.Len(t, schema.Required, 1)
	assert.Equal(t, "name", schema.Required[0])
	assert.Equal(t, "models", schema.PackageName)
	assert.Equal(t, "user.go", schema.FileName)

	// Test properties
	nameProp := schema.Properties["name"]
	assert.Equal(t, "name", nameProp.Name)
	assert.Equal(t, "string", nameProp.Type)
	assert.Equal(t, "User name", nameProp.Description)
	assert.True(t, nameProp.Required)

	ageProp := schema.Properties["age"]
	assert.Equal(t, "age", ageProp.Name)
	assert.Equal(t, "integer", ageProp.Type)
	assert.Equal(t, "User age", ageProp.Description)
	assert.False(t, ageProp.Required)
}

func TestPropertyInfo_Validation(t *testing.T) {
	tests := []struct {
		name     string
		property *PropertyInfo
		validate func(*testing.T, *PropertyInfo)
	}{
		{
			name: "should handle string property with validation",
			property: &PropertyInfo{
				Name:        "email",
				Type:        "string",
				Format:      "email",
				Description: "User email",
				Required:    true,
				MinLength:   intPtr(5),
				MaxLength:   intPtr(100),
			},
			validate: func(t *testing.T, prop *PropertyInfo) {
				assert.Equal(t, "email", prop.Name)
				assert.Equal(t, "string", prop.Type)
				assert.Equal(t, "email", prop.Format)
				assert.True(t, prop.Required)
				assert.Equal(t, 5, *prop.MinLength)
				assert.Equal(t, 100, *prop.MaxLength)
			},
		},
		{
			name: "should handle numeric property with validation",
			property: &PropertyInfo{
				Name:        "score",
				Type:        "number",
				Description: "User score",
				Required:    false,
				Minimum:     float64Ptr(0.0),
				Maximum:     float64Ptr(100.0),
			},
			validate: func(t *testing.T, prop *PropertyInfo) {
				assert.Equal(t, "score", prop.Name)
				assert.Equal(t, "number", prop.Type)
				assert.False(t, prop.Required)
				assert.Equal(t, 0.0, *prop.Minimum)
				assert.Equal(t, 100.0, *prop.Maximum)
			},
		},
		{
			name: "should handle enum property",
			property: &PropertyInfo{
				Name:        "status",
				Type:        "string",
				Description: "User status",
				Required:    true,
				Enum:        []string{"active", "inactive", "pending"},
			},
			validate: func(t *testing.T, prop *PropertyInfo) {
				assert.Equal(t, "status", prop.Name)
				assert.Equal(t, "string", prop.Type)
				assert.True(t, prop.Required)
				assert.Len(t, prop.Enum, 3)
				assert.Contains(t, prop.Enum, "active")
				assert.Contains(t, prop.Enum, "inactive")
				assert.Contains(t, prop.Enum, "pending")
			},
		},
		{
			name: "should handle array property",
			property: &PropertyInfo{
				Name:        "tags",
				Type:        "array",
				Description: "User tags",
				Required:    false,
				Items: &PropertyInfo{
					Name: "tag",
					Type: "string",
				},
			},
			validate: func(t *testing.T, prop *PropertyInfo) {
				assert.Equal(t, "tags", prop.Name)
				assert.Equal(t, "array", prop.Type)
				assert.False(t, prop.Required)
				assert.NotNil(t, prop.Items)
				assert.Equal(t, "tag", prop.Items.Name)
				assert.Equal(t, "string", prop.Items.Type)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.property)
		})
	}
}

func TestEntityMeta_Structure(t *testing.T) {
	markers := []MarkerInstance{
		{Name: "Schema", Args: []string{"User entity"}},
		{Name: "Example", Args: []string{`{"name":"John","age":30}`}},
	}

	fields := []FieldMeta{
		{
			Name:        "ID",
			Type:        "int",
			JSONTag:     "id",
			Description: "User ID",
			Validation:  "required",
		},
		{
			Name:        "Name",
			Type:        "string",
			JSONTag:     "name",
			Description: "User name",
			Validation:  "required,min=2",
		},
	}

	entity := &EntityMeta{
		Name:        "User",
		PackageName: "models",
		FileName:    "user.go",
		Markers:     markers,
		Fields:      fields,
		Description: "User entity for authentication",
		Example: map[string]interface{}{
			"id":   1,
			"name": "John Doe",
		},
	}

	assert.Equal(t, "User", entity.Name)
	assert.Equal(t, "models", entity.PackageName)
	assert.Equal(t, "user.go", entity.FileName)
	assert.Len(t, entity.Markers, 2)
	assert.Len(t, entity.Fields, 2)
	assert.Equal(t, "User entity for authentication", entity.Description)
	assert.Equal(t, 1, entity.Example["id"])
	assert.Equal(t, "John Doe", entity.Example["name"])

	// Test markers
	assert.Equal(t, "Schema", entity.Markers[0].Name)
	assert.Equal(t, "User entity", entity.Markers[0].Args[0])
	assert.Equal(t, "Example", entity.Markers[1].Name)

	// Test fields
	assert.Equal(t, "ID", entity.Fields[0].Name)
	assert.Equal(t, "int", entity.Fields[0].Type)
	assert.Equal(t, "id", entity.Fields[0].JSONTag)
	assert.Equal(t, "required", entity.Fields[0].Validation)

	assert.Equal(t, "Name", entity.Fields[1].Name)
	assert.Equal(t, "string", entity.Fields[1].Type)
	assert.Equal(t, "name", entity.Fields[1].JSONTag)
	assert.Equal(t, "required,min=2", entity.Fields[1].Validation)
}

func TestFieldMeta_Structure(t *testing.T) {
	field := &FieldMeta{
		Name:        "Email",
		Type:        "string",
		JSONTag:     "email",
		Description: "User email address",
		Example:     "user@example.com",
		Validation:  "required,email",
	}

	assert.Equal(t, "Email", field.Name)
	assert.Equal(t, "string", field.Type)
	assert.Equal(t, "email", field.JSONTag)
	assert.Equal(t, "User email address", field.Description)
	assert.Equal(t, "user@example.com", field.Example)
	assert.Equal(t, "required,email", field.Validation)
}

// Helper functions for tests
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
