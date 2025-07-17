package decorators

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// escapeGoString escapes a string for safe use in generated Go code
func escapeGoString(s string) string {
	// Handle empty strings
	if s == "" {
		return `""`
	}

	// Replace problematic characters that could break the template
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `\`, `\\`)

	// Use strconv.Quote to properly escape the string
	return strconv.Quote(s)
}

// GenerateInitFile generates the init_decorators.go file for production
func GenerateInitFile(rootDir, outputPath, pkgName string) error {
	return GenerateInitFileWithConfig(rootDir, outputPath, pkgName, nil)
}

// GenerateInitFileWithConfig generates file with specific configuration
func GenerateInitFileWithConfig(rootDir, outputPath, pkgName string, config *Config) error {
	// Parse source directory
	routes, err := ParseDirectory(rootDir)
	if err != nil {
		return fmt.Errorf("error in parsing do directory %s: %v", rootDir, err)
	}

	// Run hooks de parsing
	if err := executeParserHooks(routes); err != nil {
		return fmt.Errorf("error nos parser hooks: %v", err)
	}

	// Prepare data for generation
	genData := &GenData{
		PackageName: pkgName,
		Routes:      routes,
		Imports: []string{
			`decorators "github.com/RodolfoBonis/deco/pkg/decorators"`,
		},
		Metadata: make(map[string]interface{}),
	}

	// Add generation timestamp
	genData.Metadata["generated_at"] = time.Now().Format(time.RFC3339)

	// Run generation hooks
	if err := executeGeneratorHooks(genData); err != nil {
		return fmt.Errorf("error nos generator hooks: %v", err)
	}

	// Use default configuration if not provided
	if config == nil {
		config = DefaultConfig()
	}

	// Choose template based on minification configuration
	var tmplContent string
	if config.Prod.Minify {
		tmplContent = GetMinifiedTemplate()
	} else {
		tmplContent = getInitTemplate()
	}

	// Create template with string escaping function
	tmpl, err := template.New("init_decorators").Funcs(template.FuncMap{
		"escapeString": escapeGoString,
	}).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("error processing template: %v", err)
	}

	// Create output directory if necessary
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	// If using .deco folder, create .gitignore for it
	if strings.Contains(outputDir, ".deco") {
		gitignorePath := filepath.Join(outputDir, ".gitignore")
		if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
			gitignoreContent := `# Files generateds automatically pelo gin-decorators
*.go
!.gitignore

# Files de cache e temporários
*.tmp
*.cache
`
			if err := os.WriteFile(gitignorePath, []byte(gitignoreContent), 0644); err != nil {
				fmt.Printf("⚠️  Warning: could not criar .gitignore em %s: %v\n", outputDir, err)
			} else {
				fmt.Printf("📝 Created .gitignore em %s\n", outputDir)
			}
		}
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", outputPath, err)
	}
	defer outputFile.Close()

	// Run template
	if err := tmpl.Execute(outputFile, genData); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	// Validation if enabled
	if config.Prod.Validate {
		if err := ValidateGeneration(outputPath); err != nil {
			return fmt.Errorf("validation failed: %v", err)
		}
		LogVerbose("File validado com success")
	}

	// Count statistics
	wsHandlerCount := 0
	middlewareCount := 0
	for _, route := range routes {
		wsHandlerCount += len(route.WebSocketHandlers)
		for _, marker := range route.Markers {
			if marker.Name != "Route" && marker.Name != "Summary" && marker.Name != "Description" &&
				marker.Name != "Tag" && marker.Name != "Response" && marker.Name != "RequestBody" &&
				marker.Name != "Schema" && marker.Name != "Group" && marker.Name != "Param" {
				middlewareCount++
			}
		}
	}

	// Essential logs only
	LogNormal("Code generated: %d routes, %d websockets, %d middlewares processed", len(routes), wsHandlerCount, middlewareCount)

	// Detailed logs only in verbose mode
	LogVerbose("✅ File generated successfully: %s", outputPath)
	LogVerbose("🚀 Works in DEV and PROD automatically!")
	LogVerbose("📊 Statistics:")
	LogVerbose("   - %d routes processed", len(routes))
	LogVerbose("   - %d imports added", len(genData.Imports))
	LogVerbose("   - Package: %s", pkgName)
	if config.Prod.Minify {
		LogVerbose("📦 Code minified for production")
	}
	if config.Prod.Validate {
		LogVerbose("🔍 Syntax validation enabled")
	}
	if strings.Contains(outputPath, ".deco") {
		LogVerbose("📁 Files organizados na pasta .deco")
	}

	return nil
}

// getInitTemplate returns the default template for code generation
func getInitTemplate() string {
	return `// Code generated by gin-decorators; DO NOT EDIT.
// This file is automatically generated and works in both dev and prod modes.
package {{ .PackageName }}

import (
	"github.com/gin-gonic/gin"
{{- range .Imports }}
	{{ . }}
{{- end }}
)

func init() {
{{- range .Routes }}
{{- if and .Method .Path }}
	// {{ .Method }} {{ .Path }} -> {{ .FuncName }}
	{{- if .Description }}
	// {{ .Description }}
	{{- end }}
	deco.RegisterRouteWithMeta(deco.RouteEntry{
		Method:      "{{ .Method }}",
		Path:        "{{ .Path }}",
		Handler:     {{ if eq $.PackageName "deco" }}handlers.{{ .FuncName }}{{ else }}{{ .FuncName }}{{ end }},
		{{- if .MiddlewareCalls }}
		Middlewares: []gin.HandlerFunc{
			{{- range .MiddlewareCalls }}
			{{ . }},
			{{- end }}
		},
		{{- end }}
		FuncName:    "{{ .FuncName }}",
		PackageName: "{{ .PackageName }}",
		{{- if .Description }}
		Description: {{ escapeString .Description }},
		{{- end }}
		{{- if .Summary }}
		Summary:     {{ escapeString .Summary }},
		{{- end }}
		{{- if .Tags }}
		Tags:        []string{
			{{- range .Tags }}
			"{{ . }}",
			{{- end }}
		},
		{{- end }}
		{{- if .MiddlewareInfo }}
		MiddlewareInfo: []deco.MiddlewareInfo{
			{{- range .MiddlewareInfo }}
			{
				Name:        {{ escapeString .Name }},
				Description: {{ escapeString .Description }},
				Args: map[string]interface{}{
					{{- range $key, $value := .Args }}
					{{ escapeString $key }}: {{ escapeString $value }},
					{{- end }}
				},
			},
			{{- end }}
		},
		{{- end }}
		{{- if .Parameters }}
		Parameters: []deco.ParameterInfo{
			{{- range .Parameters }}
			{
				Name:        {{ escapeString .Name }},
				Type:        {{ escapeString .Type }},
				Location:    {{ escapeString .Location }},
				Required:    {{ .Required }},
				Description: {{ escapeString .Description }},
				Example:     {{ escapeString .Example }},
			},
			{{- end }}
		},
		{{- end }}
		{{- if .Group }}
		Group: &deco.GroupInfo{
			Name:        {{ escapeString .Group.Name }},
			Prefix:      {{ escapeString .Group.Prefix }},
			Description: {{ escapeString .Group.Description }},
		},
		{{- end }}
		{{- if .Responses }}
		Responses: []decorators.ResponseInfo{
			{{- range .Responses }}
			{
				Code:        {{ escapeString .Code }},
				Description: {{ escapeString .Description }},
				Type:        {{ escapeString .Type }},
				Example:     {{ escapeString .Example }},
			},
			{{- end }}
		},
		{{- end }}
	})
{{- else if .WebSocketHandlers }}
	// WebSocket-only handlers for {{ .FuncName }}
	{{- $funcName := .FuncName }}
	{{- range .WebSocketHandlers }}
	deco.RegisterWebSocketHandler("{{ . }}", {{ if eq $.PackageName "deco" }}handlers.{{ $funcName }}{{ else }}{{ $funcName }}{{ end }})
	{{- end }}
	
	// Register WebSocket handlers as routes for documentation
	deco.RegisterRouteWithMeta(deco.RouteEntry{
		Method:      "WS",
		Path:        "/ws/{{ .FuncName }}",
		Handler:     deco.WebSocketHandlerWrapper({{ if eq $.PackageName "deco" }}handlers.{{ .FuncName }}{{ else }}{{ .FuncName }}{{ end }}),
		FuncName:    "{{ .FuncName }}",
		PackageName: "{{ .PackageName }}",
		{{- if .Description }}
		Description: {{ escapeString .Description }},
		{{- end }}
		{{- if .Summary }}
		Summary:     {{ escapeString .Summary }},
		{{- end }}
		{{- if .Tags }}
		Tags:        []string{
			{{- range .Tags }}
			"{{ . }}",
			{{- end }}
		},
		{{- end }}
		{{- if .MiddlewareInfo }}
		MiddlewareInfo: []deco.MiddlewareInfo{
			{{- range .MiddlewareInfo }}
			{
				Name:        {{ escapeString .Name }},
				Description: {{ escapeString .Description }},
				Args: map[string]interface{}{
					{{- range $key, $value := .Args }}
					{{ escapeString $key }}: {{ escapeString $value }},
					{{- end }}
				},
			},
			{{- end }}
		},
		{{- end }}
		{{- if .Group }}
		Group: &deco.GroupInfo{
			Name:        {{ escapeString .Group.Name }},
			Prefix:      {{ escapeString .Group.Prefix }},
			Description: {{ escapeString .Group.Description }},
		},
		{{- end }}
		WebSocketHandlers: []string{
			{{- range .WebSocketHandlers }}
			"{{ . }}",
			{{- end }}
		},
	})
{{- end }}
{{- end }}

	// Initialize WebSocket default handlers
	deco.RegisterDefaultWebSocketHandlers()
}

// Metadata generated automatically
var GeneratedMetadata = map[string]interface{}{
	"routes_count": {{ len .Routes }},
	"generated_at": "{{ .GeneratedAt }}",
	"package_name": "{{ .PackageName }}",
}
`
}

// GenerateFromTemplate generates code using custom template
func GenerateFromTemplate(rootDir, templatePath, outputPath, pkgName string) error {
	// Parse source directory
	routes, err := ParseDirectory(rootDir)
	if err != nil {
		return fmt.Errorf("error in parsing: %v", err)
	}

	// Run hooks
	if err := executeParserHooks(routes); err != nil {
		return err
	}

	genData := &GenData{
		PackageName: pkgName,
		Routes:      routes,
		Imports:     []string{},
		Metadata:    make(map[string]interface{}),
	}

	if err := executeGeneratorHooks(genData); err != nil {
		return err
	}

	// Load template customizado
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("error reading template %s: %v", templatePath, err)
	}

	tmpl, err := template.New("custom").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("error processing template: %v", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer outputFile.Close()

	// Run template
	return tmpl.Execute(outputFile, genData)
}

// ValidateGeneration validates if the generated file is correct
func ValidateGeneration(generatedPath string) error {
	// Check if file exists
	if _, err := os.Stat(generatedPath); os.IsNotExist(err) {
		return fmt.Errorf("generated file not found: %s", generatedPath)
	}

	// Check if file is not empty
	info, err := os.Stat(generatedPath)
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		return fmt.Errorf("generated file is empty: %s", generatedPath)
	}

	// Complete Go syntax validation
	if err := validateGoSyntax(generatedPath); err != nil {
		return fmt.Errorf("syntax error no generated file: %v", err)
	}

	// Structural validation
	if err := validateStructure(generatedPath); err != nil {
		return fmt.Errorf("structural error no generated file: %v", err)
	}

	return nil
}

// validateGoSyntax validates the Go syntax of the file
func validateGoSyntax(filePath string) error {
	fset := token.NewFileSet()

	// Parse file
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("syntax error: %v", err)
	}

	// Check if there are parsing errors
	if file == nil {
		return fmt.Errorf("file could not be parsed")
	}

	// Validate basic AST structure
	if file.Name == nil {
		return fmt.Errorf("package declaration not found")
	}

	return nil
}

// validateStructure validates the expected structure of the generated file
func validateStructure(filePath string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	var hasInitFunc bool
	var hasImports bool
	var hasRegistrations bool

	// Verify imports
	if len(file.Imports) > 0 {
		hasImports = true
	}

	// Verify declarations
	for _, decl := range file.Decls {
		// Verify init function
		if fnDecl, ok := decl.(*ast.FuncDecl); ok {
			if fnDecl.Name.Name == "init" {
				hasInitFunc = true

				// Check if there are route registrations
				ast.Inspect(fnDecl, func(n ast.Node) bool {
					if callExpr, ok := n.(*ast.CallExpr); ok {
						if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
							if selExpr.Sel.Name == "RegisterRouteWithMeta" {
								hasRegistrations = true
							}
						}
					}
					return true
				})
			}
		}
	}

	// Validate minimum expected structure
	if !hasImports {
		return fmt.Errorf("necessary imports not founds")
	}

	if !hasInitFunc {
		return fmt.Errorf("init() function not found")
	}

	if !hasRegistrations {
		return fmt.Errorf("route registrations not founds na init() function")
	}

	return nil
}
