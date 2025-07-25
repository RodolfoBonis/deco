package decorators

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// OpenAPISpec complete OpenAPI 3.0 specification structure
type OpenAPISpec struct {
	OpenAPI      string                 `json:"openapi"`
	Info         OpenAPIInfo            `json:"info"`
	Servers      []OpenAPIServer        `json:"servers,omitempty"`
	Paths        map[string]OpenAPIPath `json:"paths"`
	Components   *OpenAPIComponents     `json:"components,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty"`
	Tags         []OpenAPITag           `json:"tags,omitempty"`
	ExternalDocs *ExternalDocs          `json:"externalDocs,omitempty"`
}

// OpenAPIInfo basic API information
type OpenAPIInfo struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Version        string   `json:"version"`
}

// Contact contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// License license information
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// OpenAPIServer server information
type OpenAPIServer struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}

// ServerVariable server variable
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

// OpenAPIPath operations available on a path
type OpenAPIPath map[string]*OpenAPIOperation

// OpenAPIOperation individual operation
type OpenAPIOperation struct {
	Tags        []string                   `json:"tags,omitempty"`
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	OperationID string                     `json:"operationId,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
	Callbacks   map[string]interface{}     `json:"callbacks,omitempty"`
	Deprecated  bool                       `json:"deprecated,omitempty"`
	Security    []SecurityRequirement      `json:"security,omitempty"`
	Servers     []OpenAPIServer            `json:"servers,omitempty"`
	Extensions  map[string]interface{}     `json:"-"`
}

// OpenAPIParameter operation parameter
type OpenAPIParameter struct {
	Name            string               `json:"name"`
	In              string               `json:"in"` // query, header, path, cookie
	Description     string               `json:"description,omitempty"`
	Required        bool                 `json:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty"`
	Schema          *OpenAPISchema       `json:"schema,omitempty"`
	Example         interface{}          `json:"example,omitempty"`
	Examples        map[string]Example   `json:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty"`
}

// OpenAPIRequestBody corpo da request
type OpenAPIRequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}

// OpenAPIResponse operation response
type OpenAPIResponse struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Links       map[string]Link      `json:"links,omitempty"`
}

// MediaType media type
type MediaType struct {
	Schema   *OpenAPISchema      `json:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty"`
}

// OpenAPISchema data schema
type OpenAPISchema struct {
	Type                 string                    `json:"type,omitempty"`
	AllOf                []*OpenAPISchema          `json:"allOf,omitempty"`
	OneOf                []*OpenAPISchema          `json:"oneOf,omitempty"`
	AnyOf                []*OpenAPISchema          `json:"anyOf,omitempty"`
	Not                  *OpenAPISchema            `json:"not,omitempty"`
	Items                *OpenAPISchema            `json:"items,omitempty"`
	Properties           map[string]*OpenAPISchema `json:"properties,omitempty"`
	AdditionalProperties interface{}               `json:"additionalProperties,omitempty"`
	Description          string                    `json:"description,omitempty"`
	Format               string                    `json:"format,omitempty"`
	Default              interface{}               `json:"default,omitempty"`
	Title                string                    `json:"title,omitempty"`
	MultipleOf           float64                   `json:"multipleOf,omitempty"`
	Maximum              float64                   `json:"maximum,omitempty"`
	ExclusiveMaximum     bool                      `json:"exclusiveMaximum,omitempty"`
	Minimum              float64                   `json:"minimum,omitempty"`
	ExclusiveMinimum     bool                      `json:"exclusiveMinimum,omitempty"`
	MaxLength            int                       `json:"maxLength,omitempty"`
	MinLength            int                       `json:"minLength,omitempty"`
	Pattern              string                    `json:"pattern,omitempty"`
	MaxItems             int                       `json:"maxItems,omitempty"`
	MinItems             int                       `json:"minItems,omitempty"`
	UniqueItems          bool                      `json:"uniqueItems,omitempty"`
	MaxProperties        int                       `json:"maxProperties,omitempty"`
	MinProperties        int                       `json:"minProperties,omitempty"`
	Required             []string                  `json:"required,omitempty"`
	Enum                 []interface{}             `json:"enum,omitempty"`
	Example              interface{}               `json:"example,omitempty"`
	Nullable             bool                      `json:"nullable,omitempty"`
	ReadOnly             bool                      `json:"readOnly,omitempty"`
	WriteOnly            bool                      `json:"writeOnly,omitempty"`
	XML                  *XML                      `json:"xml,omitempty"`
	ExternalDocs         *ExternalDocs             `json:"externalDocs,omitempty"`
	Deprecated           bool                      `json:"deprecated,omitempty"`
	Discriminator        *Discriminator            `json:"discriminator,omitempty"`
	Ref                  string                    `json:"$ref,omitempty"`
}

// OpenAPIComponents reusable components
type OpenAPIComponents struct {
	Schemas         map[string]*OpenAPISchema     `json:"schemas,omitempty"`
	Responses       map[string]OpenAPIResponse    `json:"responses,omitempty"`
	Parameters      map[string]OpenAPIParameter   `json:"parameters,omitempty"`
	Examples        map[string]Example            `json:"examples,omitempty"`
	RequestBodies   map[string]OpenAPIRequestBody `json:"requestBodies,omitempty"`
	Headers         map[string]Header             `json:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme     `json:"securitySchemes,omitempty"`
	Links           map[string]Link               `json:"links,omitempty"`
	Callbacks       map[string]interface{}        `json:"callbacks,omitempty"`
}

// SecurityRequirement security requirement
type SecurityRequirement map[string][]string

// SecurityScheme security scheme
type SecurityScheme struct {
	Type             string      `json:"type"`
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}

// OAuthFlows fluxos OAuth2
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}

// OAuthFlow fluxo OAuth2
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

// OpenAPITag tag for grouping
type OpenAPITag struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}

// ExternalDocs external documentation
type ExternalDocs struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

// Example exemplo
type Example struct {
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}

// Header header
type Header struct {
	Description     string               `json:"description,omitempty"`
	Required        bool                 `json:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty"`
	Schema          *OpenAPISchema       `json:"schema,omitempty"`
	Example         interface{}          `json:"example,omitempty"`
	Examples        map[string]Example   `json:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty"`
}

// Link link to other operations
type Link struct {
	OperationRef string                 `json:"operationRef,omitempty"`
	OperationID  string                 `json:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Server       *OpenAPIServer         `json:"server,omitempty"`
}

// Encoding encoding
type Encoding struct {
	ContentType   string            `json:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty"`
	Style         string            `json:"style,omitempty"`
	Explode       bool              `json:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty"`
}

// XML metadata
type XML struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty"`
}

// Discriminator discriminator for polymorphism
type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}

// GenerateOpenAPISpec generates complete OpenAPI 3.0 specification
func GenerateOpenAPISpec(config *Config) *OpenAPISpec {
	routes := GetRoutes()
	groups := GetGroups()

	spec := createBaseSpec(config)
	configureSpecInfo(spec, config)
	configureSpecServers(spec, config)
	configureSpecSecurity(spec, config)
	configureSpecComponents(spec)
	configureSpecTags(spec, groups)
	configureSpecPaths(spec, routes)

	return spec
}

func createBaseSpec(config *Config) *OpenAPISpec {
	info := getSpecInfo(config)

	return &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info:    info,
		Paths:   make(map[string]OpenAPIPath),
		Components: &OpenAPIComponents{
			Schemas:         make(map[string]*OpenAPISchema),
			Responses:       make(map[string]OpenAPIResponse),
			Parameters:      make(map[string]OpenAPIParameter),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
		Tags: make([]OpenAPITag, 0),
	}
}

func getSpecInfo(config *Config) OpenAPIInfo {
	title := "gin-decorators API"
	description := "API generated by gin-decorators framework"
	version := "1.0.0"

	if config != nil {
		if config.OpenAPI.Title != "" {
			title = config.OpenAPI.Title
		}
		if config.OpenAPI.Description != "" {
			description = config.OpenAPI.Description
		}
		if config.OpenAPI.Version != "" {
			version = config.OpenAPI.Version
		}
	}

	return OpenAPIInfo{
		Title:       title,
		Description: description,
		Version:     version,
	}
}

func configureSpecInfo(spec *OpenAPISpec, config *Config) {
	if config == nil {
		return
	}

	configureContact(spec, config)
	configureLicense(spec, config)
}

func configureContact(spec *OpenAPISpec, config *Config) {
	if len(config.OpenAPI.Contact) == 0 {
		return
	}

	contact := config.OpenAPI.Contact
	spec.Info.Contact = &Contact{}

	if name, ok := contact["name"].(string); ok {
		spec.Info.Contact.Name = name
	}
	if url, ok := contact["url"].(string); ok {
		spec.Info.Contact.URL = url
	}
	if email, ok := contact["email"].(string); ok {
		spec.Info.Contact.Email = email
	}
}

func configureLicense(spec *OpenAPISpec, config *Config) {
	if len(config.OpenAPI.License) == 0 {
		return
	}

	license := config.OpenAPI.License
	spec.Info.License = &License{}

	if name, ok := license["name"].(string); ok {
		spec.Info.License.Name = name
	}
	if url, ok := license["url"].(string); ok {
		spec.Info.License.URL = url
	}
}

func configureSpecServers(spec *OpenAPISpec, config *Config) {
	if config == nil || config.OpenAPI.Host == "" {
		return
	}

	for _, scheme := range config.OpenAPI.Schemes {
		spec.Servers = append(spec.Servers, OpenAPIServer{
			URL:         fmt.Sprintf("%s://%s%s", scheme, config.OpenAPI.Host, config.OpenAPI.BasePath),
			Description: fmt.Sprintf("Server %s", strings.ToUpper(scheme)),
		})
	}
}

func configureSpecSecurity(spec *OpenAPISpec, config *Config) {
	if config == nil || len(config.OpenAPI.Security) == 0 {
		return
	}

	for _, secReq := range config.OpenAPI.Security {
		spec.Security = append(spec.Security, SecurityRequirement(secReq))
	}
}

func configureSpecComponents(spec *OpenAPISpec) {
	addDefaultSecuritySchemes(spec.Components)
	addRegisteredSchemas(spec.Components)
}

func configureSpecTags(spec *OpenAPISpec, groups map[string]*GroupInfo) {
	for _, group := range groups {
		spec.Tags = append(spec.Tags, OpenAPITag{
			Name:        group.Name,
			Description: group.Description,
		})
	}
}

func configureSpecPaths(spec *OpenAPISpec, routes []RouteEntry) {
	for i := range routes {
		route := &routes[i]
		path := route.Path

		if spec.Paths[path] == nil {
			spec.Paths[path] = make(OpenAPIPath)
		}

		operation := convertRouteToOperation(route, spec.Components)
		spec.Paths[path][strings.ToLower(route.Method)] = operation
	}
}

// convertRouteToOperation converts RouteEntry to OpenAPIOperation
func convertRouteToOperation(route *RouteEntry, components *OpenAPIComponents) *OpenAPIOperation {
	operation := &OpenAPIOperation{
		Summary:     route.Summary,
		Description: route.Description,
		OperationID: generateOperationID(route),
		Responses:   make(map[string]OpenAPIResponse),
		Extensions:  make(map[string]interface{}),
	}

	// Add tags
	if route.Group != nil {
		operation.Tags = append(operation.Tags, route.Group.Name)
	}
	operation.Tags = append(operation.Tags, route.Tags...)

	// Separate body parameters from other parameters
	var bodyParams []ParameterInfo
	var otherParams []ParameterInfo

	for _, param := range route.Parameters {
		if param.Location == "body" {
			bodyParams = append(bodyParams, param)
		} else {
			otherParams = append(otherParams, param)
		}
	}

	// Process non-body parameters
	for _, param := range otherParams {
		operation.Parameters = append(operation.Parameters, convertToOpenAPIParameter(&param, components))
	}

	// Process request body if there are body parameters
	if len(bodyParams) > 0 {
		operation.RequestBody = createRequestBodyFromParameters(bodyParams, components)
	}

	// Process responses with schema support
	if len(route.Responses) > 0 {
		for _, response := range route.Responses {
			apiResponse := createResponseWithSchemaAndType(response, components)
			operation.Responses[response.Code] = apiResponse
		}
	} else {
		// Default response
		defaultResponse := ResponseInfo{
			Code:        "200",
			Description: "Success",
			Type:        "",
		}
		operation.Responses["200"] = createResponseWithSchemaAndType(defaultResponse, components)
	}

	// Add middleware information as extension
	if len(route.MiddlewareInfo) > 0 {
		middlewares := make([]map[string]interface{}, 0)
		for _, mw := range route.MiddlewareInfo {
			middlewares = append(middlewares, map[string]interface{}{
				"name":        mw.Name,
				"description": mw.Description,
				"args":        mw.Args,
			})
		}
		operation.Extensions["x-middlewares"] = middlewares
	}

	// Add rate limiting if present
	for _, mw := range route.MiddlewareInfo {
		if mw.Name == "RateLimit" {
			operation.Extensions["x-rate-limit"] = mw.Args
		}
	}

	return operation
}

// createRequestBodyFromParameters creates an OpenAPIRequestBody from a slice of ParameterInfo
func createRequestBodyFromParameters(params []ParameterInfo, _ *OpenAPIComponents) *OpenAPIRequestBody {
	if len(params) == 0 {
		return nil
	}

	requestBody := &OpenAPIRequestBody{
		Content:  make(map[string]MediaType),
		Required: true,
	}

	// Check if any parameter references an existing schema
	for _, param := range params {
		schemaRef := findSchemaByName(param.Type)
		if schemaRef != nil {
			// Reference existing schema
			requestBody.Content["application/json"] = MediaType{
				Schema: &OpenAPISchema{
					Ref: fmt.Sprintf("#/components/schemas/%s", param.Type),
				},
			}
			requestBody.Description = param.Description
		} else {
			// Create inline schema
			mediaType := MediaType{
				Schema: convertTypeToSchema(param.Type),
			}
			if param.Example != "" {
				mediaType.Example = param.Example
			}
			requestBody.Content["application/json"] = mediaType
		}
	}

	return requestBody
}

// createResponseWithSchemaAndType creates an OpenAPIResponse using ResponseInfo
func createResponseWithSchemaAndType(responseInfo ResponseInfo, _ *OpenAPIComponents) OpenAPIResponse {
	response := OpenAPIResponse{
		Description: responseInfo.Description,
		Content:     make(map[string]MediaType),
	}

	// If a specific type is provided, use it
	if responseInfo.Type != "" {
		if schema := findSchemaByName(responseInfo.Type); schema != nil {
			response.Content["application/json"] = MediaType{
				Schema: &OpenAPISchema{
					Ref: fmt.Sprintf("#/components/schemas/%s", responseInfo.Type),
				},
			}
		} else {
			// If schema not found, create inline schema with the type
			response.Content["application/json"] = MediaType{
				Schema: convertTypeToSchema(responseInfo.Type),
			}
		}
	} else {
		// Fall back to the old logic for automatic schema detection
		var schemaName string
		switch responseInfo.Code {
		case "200", "201":
			// Try to find common response schemas
			if schema := findSchemaByPattern("Response"); schema != nil {
				schemaName = schema.Name
			} else if schema := findSchemaByPattern("UserResponse"); schema != nil {
				schemaName = schema.Name
			}
		case "400", "401", "403", "404", "500":
			// Try to find error response schema
			if schema := findSchemaByPattern("ErrorResponse"); schema != nil {
				schemaName = schema.Name
			} else if schema := findSchemaByPattern("Error"); schema != nil {
				schemaName = schema.Name
			}
		}

		if schemaName != "" {
			response.Content["application/json"] = MediaType{
				Schema: &OpenAPISchema{
					Ref: fmt.Sprintf("#/components/schemas/%s", schemaName),
				},
			}
		} else {
			// Create a generic response schema
			response.Content["application/json"] = MediaType{
				Schema: &OpenAPISchema{
					Type: "object",
					Properties: map[string]*OpenAPISchema{
						"message": {Type: "string"},
						"data":    {Type: "object"},
					},
				},
			}
		}
	}

	// Add example if provided
	if responseInfo.Example != "" {
		mediaType := response.Content["application/json"]
		mediaType.Example = responseInfo.Example
		response.Content["application/json"] = mediaType
	}

	return response
}

// findSchemaByName finds a registered schema by exact name
func findSchemaByName(name string) *SchemaInfo {
	schemas := GetSchemas()
	return schemas[name]
}

// findSchemaByPattern finds a registered schema by name pattern
func findSchemaByPattern(pattern string) *SchemaInfo {
	schemas := GetSchemas()

	// First try exact match
	if schema, exists := schemas[pattern]; exists {
		return schema
	}

	// Then try pattern match
	for name, schema := range schemas {
		if strings.Contains(name, pattern) {
			return schema
		}
	}
	return nil
}

// convertToOpenAPIParameter converts ParameterInfo to OpenAPIParameter
func convertToOpenAPIParameter(param *ParameterInfo, _ *OpenAPIComponents) OpenAPIParameter {
	openAPIParam := OpenAPIParameter{
		Name:        param.Name,
		In:          param.Location,
		Description: param.Description,
		Required:    param.Required,
		Schema:      convertTypeToSchema(param.Type),
	}

	if param.Example != "" {
		openAPIParam.Example = param.Example
	}

	return openAPIParam
}

// convertTypeToSchema converts Go type to OpenAPI Schema
func convertTypeToSchema(goType string) *OpenAPISchema {
	schema := &OpenAPISchema{}

	switch goType {
	case "string":
		schema.Type = "string"
	case "int", "int32", "int64":
		schema.Type = "integer"
		if goType == "int64" {
			schema.Format = "int64"
		} else {
			schema.Format = "int32"
		}
	case "float32", "float64":
		schema.Type = "number"
		if goType == "float32" {
			schema.Format = "float"
		} else {
			schema.Format = "double"
		}
	case "bool", "boolean":
		schema.Type = "boolean"
	case "time.Time":
		schema.Type = "string"
		schema.Format = "date-time"
	case "uuid.UUID":
		schema.Type = "string"
		schema.Format = "uuid"
	case "[]string":
		schema.Type = "array"
		schema.Items = &OpenAPISchema{Type: "string"}
	case "[]int":
		schema.Type = "array"
		schema.Items = &OpenAPISchema{Type: "integer", Format: "int32"}
	default:
		// Complex or custom types
		if strings.HasPrefix(goType, "[]") {
			schema.Type = "array"
			itemType := strings.TrimPrefix(goType, "[]")

			// Check if the array item type is a registered schema
			if registeredSchema := findSchemaByName(itemType); registeredSchema != nil {
				schema.Items = &OpenAPISchema{
					Ref: fmt.Sprintf("#/components/schemas/%s", itemType),
				}
			} else {
				schema.Items = convertTypeToSchema(itemType)
			}
		} else {
			// Check if this is a registered schema
			if registeredSchema := findSchemaByName(goType); registeredSchema != nil {
				schema.Ref = fmt.Sprintf("#/components/schemas/%s", goType)
			} else {
				schema.Type = "object"
			}
		}
	}

	return schema
}

// generateOperationID generates unique ID for the operation
func generateOperationID(route *RouteEntry) string {
	// Clean characters especiais do path
	cleanPath := regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(route.Path, "")

	// Use cases.Title instead of deprecated strings.Title
	caser := cases.Title(language.English)
	operationID := strings.ToLower(route.Method) + caser.String(cleanPath)

	if route.FuncName != "" {
		operationID = route.FuncName
	}

	return operationID
}

// addDefaultSecuritySchemes adds default security schemes
func addDefaultSecuritySchemes(components *OpenAPIComponents) {
	components.SecuritySchemes["BearerAuth"] = SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "Authentication via Bearer Token (JWT)",
	}

	components.SecuritySchemes["ApiKeyAuth"] = SecurityScheme{
		Type:        "apiKey",
		In:          "header",
		Name:        "X-API-Key",
		Description: "Authentication via API Key in header",
	}

	components.SecuritySchemes["BasicAuth"] = SecurityScheme{
		Type:        "http",
		Scheme:      "basic",
		Description: "HTTP Basic Authentication",
	}

	components.SecuritySchemes["OAuth2"] = SecurityScheme{
		Type:        "oauth2",
		Description: "OAuth 2.0",
		Flows: &OAuthFlows{
			AuthorizationCode: &OAuthFlow{
				AuthorizationURL: "/oauth/authorize",
				TokenURL:         "/oauth/token",
				Scopes: map[string]string{
					"read":  "Read permission",
					"write": "Write permission",
					"admin": "Administrative permissions",
				},
			},
		},
	}
}

// addRegisteredSchemas adds schemas registered via RegisterSchema to OpenAPI components
func addRegisteredSchemas(components *OpenAPIComponents) {
	// First resolve schema references now that all schemas are registered
	resolveSchemaReferences()

	registeredSchemas := GetSchemas()

	for _, schemaInfo := range registeredSchemas {
		openAPISchema := convertSchemaInfoToOpenAPISchema(schemaInfo)
		components.Schemas[schemaInfo.Name] = openAPISchema
		LogVerbose("Added schema to OpenAPI spec: %s", schemaInfo.Name)
	}
}

// convertSchemaInfoToOpenAPISchema converts SchemaInfo to OpenAPISchema
func convertSchemaInfoToOpenAPISchema(schemaInfo *SchemaInfo) *OpenAPISchema {
	schema := &OpenAPISchema{
		Type:        schemaInfo.Type,
		Description: schemaInfo.Description,
		Properties:  make(map[string]*OpenAPISchema),
		Required:    schemaInfo.Required,
	}

	// Add example if provided
	if schemaInfo.Example != nil {
		schema.Example = schemaInfo.Example
	}

	// Convert properties
	for propName, propInfo := range schemaInfo.Properties {
		propSchema := &OpenAPISchema{
			Type:        propInfo.Type,
			Description: propInfo.Description,
		}

		// Handle schema reference
		if propInfo.Ref != "" {
			propSchema = &OpenAPISchema{
				Ref: propInfo.Ref,
			}
		} else {
			// Set format if available
			if propInfo.Format != "" {
				propSchema.Format = propInfo.Format
			}

			// Set example if available
			if propInfo.Example != nil {
				propSchema.Example = propInfo.Example
			}

			// Handle array items
			if propInfo.Items != nil {
				if propInfo.Items.Ref != "" {
					propSchema.Items = &OpenAPISchema{
						Ref: propInfo.Items.Ref,
					}
				} else {
					propSchema.Items = &OpenAPISchema{
						Type:   propInfo.Items.Type,
						Format: propInfo.Items.Format,
					}
				}
			}

			// Set validation constraints
			if propInfo.MinLength != nil {
				propSchema.MinLength = *propInfo.MinLength
			}
			if propInfo.MaxLength != nil {
				propSchema.MaxLength = *propInfo.MaxLength
			}
			if propInfo.Minimum != nil {
				propSchema.Minimum = *propInfo.Minimum
			}
			if propInfo.Maximum != nil {
				propSchema.Maximum = *propInfo.Maximum
			}
			if len(propInfo.Enum) > 0 {
				for _, enumVal := range propInfo.Enum {
					propSchema.Enum = append(propSchema.Enum, enumVal)
				}
			}
		}

		schema.Properties[propName] = propSchema
	}

	return schema
}

// OpenAPIJSONHandler serves OpenAPI 3.0 documentation in JSON
func OpenAPIJSONHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		spec := GenerateOpenAPISpec(config)
		c.JSON(http.StatusOK, spec)
	}
}

// OpenAPIYAMLHandler serves OpenAPI 3.0 documentation in YAML
func OpenAPIYAMLHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		spec := GenerateOpenAPISpec(config)
		c.YAML(http.StatusOK, spec)
	}
}

// SwaggerUIHandler creates Swagger UI handler with customizable settings via config
func SwaggerUIHandler(_ *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use config to customize Swagger UI settings
		swaggerURL := "/decorators/openapi.json"
		// Don't add BasePath for internal endpoints

		// Customize Swagger UI HTML based on config
		htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '{{SWAGGER_URL}}',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                docExpansion: "list",
                filter: true,
                showRequestHeaders: true,
                tryItOutEnabled: true,
                requestInterceptor: function(request) {
                    // Add custom headers if needed
                    return request;
                }
            });
        };
    </script>
</body>
</html>`

		// Replace placeholder with actual URL
		html := strings.Replace(htmlTemplate, "{{SWAGGER_URL}}", swaggerURL, 1)

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	}
}

// SwaggerRedirectHandler redirects to swagger UI (convenience endpoint)
func SwaggerRedirectHandler(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/decorators/swagger-ui")
}
