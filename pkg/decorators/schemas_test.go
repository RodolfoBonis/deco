package decorators

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseStructFields(t *testing.T) {
	// Create a simple struct type
	structType := &ast.StructType{
		Fields: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{{Name: "Name"}},
					Type:  &ast.Ident{Name: "string"},
					Tag:   &ast.BasicLit{Value: "`json:\"name\"`"},
				},
				{
					Names: []*ast.Ident{{Name: "Age"}},
					Type:  &ast.Ident{Name: "int"},
					Tag:   &ast.BasicLit{Value: "`json:\"age\"`"},
				},
			},
		},
	}

	fields := parseStructFields(structType)
	assert.Len(t, fields, 2)
	assert.Equal(t, "Name", fields[0].Name)
	assert.Equal(t, "string", fields[0].Type)
	assert.Equal(t, "name", fields[0].JSONTag)
	assert.Equal(t, "Age", fields[1].Name)
	assert.Equal(t, "int", fields[1].Type)
	assert.Equal(t, "age", fields[1].JSONTag)
}

func TestExtractTypeString(t *testing.T) {
	// Test with basic types
	assert.Equal(t, "string", extractTypeString(&ast.Ident{Name: "string"}))
	assert.Equal(t, "int", extractTypeString(&ast.Ident{Name: "int"}))
	assert.Equal(t, "bool", extractTypeString(&ast.Ident{Name: "bool"}))

	// Test with pointer types
	assert.Equal(t, "*string", extractTypeString(&ast.StarExpr{X: &ast.Ident{Name: "string"}}))
	assert.Equal(t, "*User", extractTypeString(&ast.StarExpr{X: &ast.Ident{Name: "User"}}))

	// Test with array types
	assert.Equal(t, "[]string", extractTypeString(&ast.ArrayType{Elt: &ast.Ident{Name: "string"}}))
	assert.Equal(t, "[]int", extractTypeString(&ast.ArrayType{Elt: &ast.Ident{Name: "int"}}))

	// Test with map types
	assert.Equal(t, "map[string]interface{}", extractTypeString(&ast.MapType{
		Key:   &ast.Ident{Name: "string"},
		Value: &ast.InterfaceType{},
	}))
}

func TestExtractJSONTag(t *testing.T) {
	// Test with valid JSON tag
	jsonName := extractJSONTag("`json:\"name,omitempty\"`")
	assert.Equal(t, "name", jsonName)

	// Test with JSON tag without omitempty
	jsonName = extractJSONTag("`json:\"age\"`")
	assert.Equal(t, "age", jsonName)

	// Test with no JSON tag
	jsonName = extractJSONTag("`validate:\"required\"`")
	assert.Equal(t, "", jsonName)
}

func TestExtractValidateTag(t *testing.T) {
	// Test with validate tag
	constraints := extractValidateTag("`validate:\"required,email\"`")
	assert.Equal(t, "required,email", constraints)

	// Test with no validate tag
	constraints = extractValidateTag("`json:\"name\"`")
	assert.Equal(t, "", constraints)
}

func TestConvertEntityToSchema(t *testing.T) {
	entity := &EntityMeta{
		Name: "User",
		Fields: []FieldMeta{
			{Name: "Name", Type: "string", JSONTag: "name"},
			{Name: "Age", Type: "int", JSONTag: "age"},
		},
	}

	schema := convertEntityToSchema(entity)
	assert.Equal(t, "User", schema.Name)
	assert.Len(t, schema.Properties, 2)
	assert.Equal(t, "string", schema.Properties["name"].Type)
	assert.Equal(t, "integer", schema.Properties["age"].Type)
}

func TestGetFieldNameForJSON(t *testing.T) {
	field := &FieldMeta{
		Name:    "Name",
		JSONTag: "name,omitempty",
	}

	jsonName := getFieldNameForJSON(field)
	assert.Equal(t, "name,omitempty", jsonName)
}

func TestMapGoTypeToOpenAPIType(t *testing.T) {
	assert.Equal(t, "string", mapGoTypeToOpenAPIType("string"))
	assert.Equal(t, "integer", mapGoTypeToOpenAPIType("int"))
	assert.Equal(t, "integer", mapGoTypeToOpenAPIType("int64"))
	assert.Equal(t, "number", mapGoTypeToOpenAPIType("float64"))
	assert.Equal(t, "boolean", mapGoTypeToOpenAPIType("bool"))
	assert.Equal(t, "array", mapGoTypeToOpenAPIType("[]string"))
	assert.Equal(t, "object", mapGoTypeToOpenAPIType("map[string]interface{}"))
	assert.Equal(t, "object", mapGoTypeToOpenAPIType("User"))
}

func TestGetOpenAPIFormat(t *testing.T) {
	assert.Equal(t, "", getOpenAPIFormat("time.Time"))
	assert.Equal(t, "", getOpenAPIFormat("uuid.UUID"))
	assert.Equal(t, "", getOpenAPIFormat("string"))
	assert.Equal(t, "", getOpenAPIFormat("int"))
}

func TestIsFieldRequired(t *testing.T) {
	required := isFieldRequired("required")
	assert.True(t, required)

	required = isFieldRequired("email")
	assert.False(t, required)
}

func TestResolvePropertyReferences(t *testing.T) {
	property := &PropertyInfo{
		Type: "User",
	}
	resolvePropertyReferences(property)
	assert.Equal(t, "", property.Ref)
}

func TestExtractValidationConstraints(t *testing.T) {
	property := &PropertyInfo{}
	extractValidationConstraints("required,email,min=1", property)
	assert.NotNil(t, property)
}

func TestExtractMarkersFromComment(t *testing.T) {
	comment := `// @Schema(name="User", description="User entity")
// @Response(code=200, type="User")`

	markers := extractMarkersFromComment(comment)
	assert.Len(t, markers, 2)

	// Check that both markers are present (order doesn't matter)
	markerNames := make(map[string]bool)
	for _, marker := range markers {
		markerNames[marker.Name] = true
	}
	assert.True(t, markerNames["Schema"])
	assert.True(t, markerNames["Response"])
}

func TestParseArgumentsFromString(t *testing.T) {
	args := parseArgumentsFromString(`name="User", description="User entity"`)
	assert.Contains(t, args, "name=\"User\"")
	assert.Contains(t, args, "description=\"User entity\"")
}

func TestGetSchema(t *testing.T) {
	// Clear schemas first
	ClearSchemas()

	// Register a schema
	RegisterSchema(&SchemaInfo{
		Name: "User",
		Properties: map[string]*PropertyInfo{
			"name": {Type: "string"},
			"age":  {Type: "integer"},
		},
	})

	// Get the schema
	schema := GetSchema("User")
	assert.NotNil(t, schema)
	assert.Equal(t, "User", schema.Name)
	assert.Len(t, schema.Properties, 2)
}

func TestClearSchemas(t *testing.T) {
	// Register a schema
	RegisterSchema(&SchemaInfo{Name: "Test"})

	// Clear schemas
	ClearSchemas()

	// Verify schema is cleared
	schema := GetSchema("Test")
	assert.Nil(t, schema)
}
