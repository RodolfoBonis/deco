package decorators

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Route represents a route extracted from parsing
type Route struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Handler     gin.HandlerFunc   `json:"-"`
	Middlewares []gin.HandlerFunc `json:"-"`
}

// MiddlewareInfo information about middlewares aplicados
type MiddlewareInfo struct {
	Name        string                 `json:"name"`
	Args        map[string]interface{} `json:"args"`
	Order       int                    `json:"order"`
	Description string                 `json:"description"`
}

// FrameworkStats statistics do framework
type FrameworkStats struct {
	TotalRoutes       int            `json:"total_routes"`
	UniqueMiddlewares int            `json:"unique_middlewares"`
	PackagesScanned   int            `json:"packages_scanned"`
	BuildMode         string         `json:"build_mode"` // "development" ou "production"
	GeneratedAt       string         `json:"generated_at"`
	Methods           map[string]int `json:"methods"` // GET: 5, POST: 3, etc.
}

// ValidationError validation error during parsing or generation
type ValidationError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d - %s", e.File, e.Line, e.Message)
	}
	return fmt.Sprintf("%s - %s", e.File, e.Message)
}

// ParserStats statistics do processo de parsing
type ParserStats struct {
	FilesProcessed  int               `json:"files_processed"`
	RoutesFound     int               `json:"routes_found"`
	MarkersApplied  int               `json:"markers_applied"`
	Errors          []ValidationError `json:"errors"`
	Warnings        []ValidationError `json:"warnings"`
	ProcessingTime  string            `json:"processing_time"`
	SourceDirectory string            `json:"source_directory"`
}

// SchemaInfo information about a registered schema/entity
type SchemaInfo struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Type        string                   `json:"type"` // "object", "array", etc.
	Properties  map[string]*PropertyInfo `json:"properties,omitempty"`
	Required    []string                 `json:"required,omitempty"`
	Example     interface{}              `json:"example,omitempty"`
	PackageName string                   `json:"package_name"`
	FileName    string                   `json:"file_name"`
}

// PropertyInfo information about a schema property
type PropertyInfo struct {
	Name        string        `json:"name"`
	Type        string        `json:"type,omitempty"`
	Format      string        `json:"format,omitempty"`
	Description string        `json:"description,omitempty"`
	Example     interface{}   `json:"example,omitempty"`
	Required    bool          `json:"required"`
	Enum        []string      `json:"enum,omitempty"`
	MinLength   *int          `json:"min_length,omitempty"`
	MaxLength   *int          `json:"max_length,omitempty"`
	Minimum     *float64      `json:"minimum,omitempty"`
	Maximum     *float64      `json:"maximum,omitempty"`
	Items       *PropertyInfo `json:"items,omitempty"` // For array types
	Ref         string        `json:"$ref,omitempty"`  // For schema references
}

// EntityMeta represents metadata of an entity/struct extracted from comments
type EntityMeta struct {
	Name        string           `json:"name"`
	PackageName string           `json:"package_name"`
	FileName    string           `json:"file_name"`
	Markers     []MarkerInstance `json:"markers"`
	Fields      []FieldMeta      `json:"fields"`

	// Documentation information
	Description string                 `json:"description"`
	Example     map[string]interface{} `json:"example,omitempty"`
}

// FieldMeta represents metadata of a struct field
type FieldMeta struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	JSONTag     string      `json:"json_tag"`
	Description string      `json:"description"`
	Example     interface{} `json:"example,omitempty"`
	Validation  string      `json:"validation,omitempty"` // from validate tags
}
