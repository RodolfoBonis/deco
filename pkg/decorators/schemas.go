package decorators

import (
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

// global schemas registry
var schemas = make(map[string]*SchemaInfo)

// RegisterSchema registers a new schema in the framework
func RegisterSchema(schema *SchemaInfo) {
	if schema.Name != "" {
		schemas[schema.Name] = schema
		LogVerbose("Schema registered: %s", schema.Name)
	}
}

// GetSchemas returns all registered schemas
func GetSchemas() map[string]*SchemaInfo {
	return schemas
}

// GetSchema returns a specific schema by name
func GetSchema(name string) *SchemaInfo {
	return schemas[name]
}

// ClearSchemas clears all registered schemas (useful for testing)
func ClearSchemas() {
	schemas = make(map[string]*SchemaInfo)
}

// parseEntityFromStruct extracts entity metadata from a struct declaration
func parseEntityFromStruct(fset *token.FileSet, fileName string, structDecl *ast.GenDecl, pkgName string) *EntityMeta {
	if structDecl.Doc == nil {
		return nil
	}

	// Join all comments
	var comments []string
	for _, comment := range structDecl.Doc.List {
		comments = append(comments, comment.Text)
	}
	commentText := strings.Join(comments, "\n")

	// Look for @Schema marker
	schemaRegex := regexp.MustCompile(`@Schema\s*\(([^)]*)\)`)
	if !schemaRegex.MatchString(commentText) {
		return nil // Not a schema struct
	}

	// Find the struct type
	for _, spec := range structDecl.Specs {
		if typeSpec, ok := spec.(*ast.TypeSpec); ok {
			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				entity := &EntityMeta{
					Name:        typeSpec.Name.Name,
					PackageName: pkgName,
					FileName:    fileName,
					Markers:     extractMarkersFromComment(commentText),
					Fields:      parseStructFields(structType),
				}

				// Extract description from markers
				for _, marker := range entity.Markers {
					if marker.Name == "Description" && len(marker.Args) > 0 {
						entity.Description = strings.Trim(marker.Args[0], `"`)
					}
				}

				return entity
			}
		}
	}

	return nil
}

// parseStructFields extracts field information from struct
func parseStructFields(structType *ast.StructType) []FieldMeta {
	var fields []FieldMeta

	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if !name.IsExported() {
				continue // Skip unexported fields
			}

			fieldMeta := FieldMeta{
				Name: name.Name,
				Type: extractTypeString(field.Type),
			}

			// Extract JSON tag
			if field.Tag != nil {
				tagValue := field.Tag.Value
				if jsonTag := extractJSONTag(tagValue); jsonTag != "" {
					fieldMeta.JsonTag = jsonTag
				}

				// Extract validation tags
				if validateTag := extractValidateTag(tagValue); validateTag != "" {
					fieldMeta.Validation = validateTag
				}
			}

			// Extract field comment/description
			if field.Comment != nil {
				var comments []string
				for _, comment := range field.Comment.List {
					comments = append(comments, strings.TrimPrefix(comment.Text, "//"))
				}
				fieldMeta.Description = strings.TrimSpace(strings.Join(comments, " "))
			}

			fields = append(fields, fieldMeta)
		}
	}

	return fields
}

// extractTypeString converts ast.Expr to string representation
func extractTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + extractTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + extractTypeString(t.Elt)
	case *ast.SelectorExpr:
		pkg := extractTypeString(t.X)
		return pkg + "." + t.Sel.Name
	case *ast.MapType:
		return "map[" + extractTypeString(t.Key) + "]" + extractTypeString(t.Value)
	default:
		return "interface{}"
	}
}

// extractJSONTag extracts JSON tag from struct tag
func extractJSONTag(tag string) string {
	jsonRegex := regexp.MustCompile(`json:"([^"]*)"`)
	matches := jsonRegex.FindStringSubmatch(tag)
	if len(matches) > 1 {
		return strings.Split(matches[1], ",")[0] // Get field name, ignore omitempty etc
	}
	return ""
}

// extractValidateTag extracts validation tag from struct tag
func extractValidateTag(tag string) string {
	validateRegex := regexp.MustCompile(`validate:"([^"]*)"`)
	matches := validateRegex.FindStringSubmatch(tag)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// convertEntityToSchema converts EntityMeta to SchemaInfo
func convertEntityToSchema(entity *EntityMeta) *SchemaInfo {
	schema := &SchemaInfo{
		Name:        entity.Name,
		Description: entity.Description,
		Type:        "object",
		Properties:  make(map[string]*PropertyInfo),
		PackageName: entity.PackageName,
		FileName:    entity.FileName,
	}

	var required []string

	for _, field := range entity.Fields {
		propInfo := &PropertyInfo{
			Name:        getFieldNameForJSON(field),
			Type:        mapGoTypeToOpenAPIType(field.Type),
			Description: field.Description,
		}

		// Set format if applicable
		if format := getOpenAPIFormat(field.Type); format != "" {
			propInfo.Format = format
		}

		// Handle array types
		if strings.HasPrefix(field.Type, "[]") {
			itemType := strings.TrimPrefix(field.Type, "[]")
			propInfo.Items = &PropertyInfo{
				Type: mapGoTypeToOpenAPIType(itemType),
			}

			// Set format for array items if applicable
			if format := getOpenAPIFormat(itemType); format != "" {
				propInfo.Items.Format = format
			}

			// Store the raw item type for later reference resolution
			propInfo.Items.Name = itemType // Use Name field to store original type
		}

		// Check if field is required based on validation tags
		if isFieldRequired(field.Validation) {
			required = append(required, propInfo.Name)
			propInfo.Required = true
		}

		// Extract validation constraints
		extractValidationConstraints(field.Validation, propInfo)

		schema.Properties[propInfo.Name] = propInfo
	}

	if len(required) > 0 {
		schema.Required = required
	}

	return schema
}

// getFieldNameForJSON returns the field name to use in JSON (considers json tag)
func getFieldNameForJSON(field FieldMeta) string {
	if field.JsonTag != "" && field.JsonTag != "-" {
		return field.JsonTag
	}
	// Convert field name to camelCase if no json tag
	return strings.ToLower(field.Name[:1]) + field.Name[1:]
}

// mapGoTypeToOpenAPIType maps Go types to OpenAPI types
func mapGoTypeToOpenAPIType(goType string) string {
	switch {
	case goType == "string":
		return "string"
	case goType == "int" || goType == "int32" || goType == "int64":
		return "integer"
	case goType == "float32" || goType == "float64":
		return "number"
	case goType == "bool":
		return "boolean"
	case strings.HasPrefix(goType, "[]"):
		return "array"
	case strings.HasPrefix(goType, "map["):
		return "object"
	case strings.HasPrefix(goType, "*"):
		// Pointer type - recursively map the underlying type
		return mapGoTypeToOpenAPIType(strings.TrimPrefix(goType, "*"))
	default:
		return "object"
	}
}

// getOpenAPIFormat returns OpenAPI format for specific Go types
func getOpenAPIFormat(goType string) string {
	switch goType {
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "float32":
		return "float"
	case "float64":
		return "double"
	default:
		return ""
	}
}

// isFieldRequired checks if field is required based on validation tag
func isFieldRequired(validation string) bool {
	return strings.Contains(validation, "required")
}

// resolveSchemaReferences resolves schema references in all registered schemas
// This should be called after all schemas have been registered
func resolveSchemaReferences() {
	registeredSchemas := GetSchemas()

	for _, schema := range registeredSchemas {
		for _, property := range schema.Properties {
			resolvePropertyReferences(property)
		}
	}
}

// resolvePropertyReferences resolves references in a single property
func resolvePropertyReferences(prop *PropertyInfo) {
	// Check if this property has items (is an array)
	if prop.Items != nil && prop.Items.Name != "" {
		itemTypeName := prop.Items.Name

		// Check if the item type is a registered schema
		if registeredSchema := findSchemaByName(itemTypeName); registeredSchema != nil {
			// Replace with schema reference
			prop.Items = &PropertyInfo{
				Ref: fmt.Sprintf("#/components/schemas/%s", itemTypeName),
			}
		}
	}
}

// extractValidationConstraints extracts validation constraints and sets them in PropertyInfo
func extractValidationConstraints(validation string, prop *PropertyInfo) {
	if validation == "" {
		return
	}

	// Extract min/max length for strings
	if minRegex := regexp.MustCompile(`min=(\d+)`); minRegex.MatchString(validation) {
		if matches := minRegex.FindStringSubmatch(validation); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				if prop.Type == "string" {
					prop.MinLength = &val
				} else if prop.Type == "integer" || prop.Type == "number" {
					min := float64(val)
					prop.Minimum = &min
				}
			}
		}
	}

	if maxRegex := regexp.MustCompile(`max=(\d+)`); maxRegex.MatchString(validation) {
		if matches := maxRegex.FindStringSubmatch(validation); len(matches) > 1 {
			if val, err := strconv.Atoi(matches[1]); err == nil {
				if prop.Type == "string" {
					prop.MaxLength = &val
				} else if prop.Type == "integer" || prop.Type == "number" {
					max := float64(val)
					prop.Maximum = &max
				}
			}
		}
	}

	// Extract enum values
	if enumRegex := regexp.MustCompile(`oneof=([^,\s]+)`); enumRegex.MatchString(validation) {
		if matches := enumRegex.FindStringSubmatch(validation); len(matches) > 1 {
			enumValues := strings.Split(matches[1], " ")
			prop.Enum = enumValues
		}
	}
}

// extractMarkersFromComment extracts markers from comment text (reused from parser.go)
func extractMarkersFromComment(commentText string) []MarkerInstance {
	var markers []MarkerInstance

	// Look for each registered marker
	for name, config := range GetMarkers() {
		matches := config.Pattern.FindAllStringSubmatch(commentText, -1)
		for _, match := range matches {
			marker := MarkerInstance{
				Name: name,
				Raw:  match[0],
			}

			// Extract arguments if they exist
			if len(match) > 1 && match[1] != "" {
				marker.Args = parseArgumentsFromString(match[1])
			}

			markers = append(markers, marker)
		}
	}

	return markers
}

// parseArgumentsFromString converts argument string to slice
func parseArgumentsFromString(argsStr string) []string {
	if argsStr == "" {
		return nil
	}

	var args []string
	parts := strings.Split(argsStr, ",")
	for _, part := range parts {
		arg := strings.TrimSpace(part)
		if arg != "" {
			// Remove surrounding quotes if present
			if (strings.HasPrefix(arg, `"`) && strings.HasSuffix(arg, `"`)) ||
				(strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'")) {
				arg = arg[1 : len(arg)-1]
			}
			args = append(args, arg)
		}
	}

	return args
}
