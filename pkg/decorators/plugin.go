package decorators

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RouteMeta represents metadata of a route extracted from comments
type RouteMeta struct {
	Method          string           // GET, POST, etc.
	Path            string           // /api/users
	FuncName        string           // GetUsers
	PackageName     string           // handlers
	FileName        string           // user_handlers.go
	Markers         []MarkerInstance // found marker instances
	MiddlewareCalls []string         // generated middleware calls

	// Documentation information
	Description       string           `json:"description"`
	Summary           string           `json:"summary"`
	Tags              []string         `json:"tags"`
	MiddlewareInfo    []MiddlewareInfo `json:"middlewareInfo"`
	Parameters        []ParameterInfo  `json:"parameters"`
	Group             *GroupInfo       `json:"group,omitempty"`
	Responses         []ResponseInfo   `json:"responses,omitempty"`         // Updated to use ResponseInfo
	WebSocketHandlers []string         `json:"websocketHandlers,omitempty"` // WebSocket message types this function handles
}

// MarkerInstance represents a marker instance found
type MarkerInstance struct {
	Name string   // Auth, Cache, etc.
	Args []string // parsed arguments
	Raw  string   // original comment text
}

// GenData data passed to generation template
type GenData struct {
	PackageName string                 // nome do pacote de destino
	Routes      []*RouteMeta           // routes to be generated
	Imports     []string               // necessary imports
	Metadata    map[string]interface{} // additional plugin data
	GeneratedAt string                 // generation timestamp
}

// Hooks for extensibility
type (
	// ParserHook executed after parsing routes
	ParserHook func(routes []*RouteMeta) error

	// GeneratorHook executed before code generation
	GeneratorHook func(data *GenData) error
)

// registries globais de hooks
var (
	parserHooks    []ParserHook
	generatorHooks []GeneratorHook
)

// RegisterParserHook registers a parsing hook
func RegisterParserHook(h ParserHook) {
	parserHooks = append(parserHooks, h)
	LogVerbose("Parser hook registrado")
}

// RegisterGeneratorHook registers a generation hook
func RegisterGeneratorHook(h GeneratorHook) {
	generatorHooks = append(generatorHooks, h)
	LogVerbose("Generator hook registrado")
}

// executeParserHooks executes all parsing hooks
func executeParserHooks(routes []*RouteMeta) error {
	for i, hook := range parserHooks {
		if err := hook(routes); err != nil {
			return err
		}
		LogVerbose("Parser hook %d executed successfully", i+1)
	}
	return nil
}

// executeGeneratorHooks executes all generation hooks
func executeGeneratorHooks(data *GenData) error {
	for i, hook := range generatorHooks {
		if err := hook(data); err != nil {
			return err
		}
		LogVerbose("Generator hook %d executed successfully", i+1)
	}
	return nil
}

// GetParserHooks returns all parser hooks (for testing)
func GetParserHooks() []ParserHook {
	return parserHooks
}

// GetGeneratorHooks returns all generator hooks (for testing)
func GetGeneratorHooks() []GeneratorHook {
	return generatorHooks
}

// Example plugin that adds automatic logging
func init() {
	registerLoggingPlugin()
	registerImportsPlugin()
}

// registerLoggingPlugin registers the logging plugin
func registerLoggingPlugin() {
	RegisterParserHook(func(routes []*RouteMeta) error {
		LogVerbose("Plugin de logging: %d routes processadas", len(routes))
		for _, route := range routes {
			LogVerbose("  - %s %s -> %s", route.Method, route.Path, route.FuncName)
		}
		return nil
	})
}

// registerImportsPlugin registers the imports plugin
func registerImportsPlugin() {
	RegisterGeneratorHook(func(data *GenData) error {
		requiredImports := getRequiredImports(data)
		addMissingImports(data, requiredImports)
		LogVerbose("Plugin de imports: %d imports configurados", len(data.Imports))
		return nil
	})
}

// getRequiredImports returns the list of required imports
func getRequiredImports(data *GenData) []string {
	requiredImports := []string{
		`deco "github.com/RodolfoBonis/deco"`,
	}

	if shouldAddHandlersImport(data) {
		if handlerImport := buildHandlersImport(data); handlerImport != "" {
			requiredImports = append(requiredImports, handlerImport)
		}
	}

	return requiredImports
}

// shouldAddHandlersImport checks if handlers import should be added
func shouldAddHandlersImport(data *GenData) bool {
	return data.PackageName == "deco" && len(data.Routes) > 0
}

// buildHandlersImport builds the handlers import path
func buildHandlersImport(data *GenData) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	modName := getModuleName(wd)
	if modName == "" {
		return ""
	}

	// Use the actual package name from the first route
	if len(data.Routes) > 0 {
		actualPackageName := data.Routes[0].PackageName
		return buildImportPath(wd, modName, actualPackageName)
	}

	return ""
}

// buildImportPath builds the import path based on working directory and module name
func buildImportPath(wd, modName, packageName string) string {
	if !strings.Contains(wd, modName) {
		return fmt.Sprintf(`%s "%s/%s"`, packageName, modName, packageName)
	}

	parts := strings.Split(wd, modName)
	if len(parts) <= 1 {
		return fmt.Sprintf(`%s "%s/%s"`, packageName, modName, packageName)
	}

	relativePath := strings.TrimPrefix(parts[1], "/")
	if relativePath == "" {
		return fmt.Sprintf(`%s "%s/%s"`, packageName, modName, packageName)
	}

	return fmt.Sprintf(`%s "%s/%s/%s"`, packageName, modName, relativePath, packageName)
}

// addMissingImports adds missing imports to the data
func addMissingImports(data *GenData, requiredImports []string) {
	for _, imp := range requiredImports {
		if !containsImport(data.Imports, imp) {
			data.Imports = append(data.Imports, imp)
		}
	}
}

// containsImport checks if an import already exists
func containsImport(imports []string, imp string) bool {
	for _, existing := range imports {
		if existing == imp {
			return true
		}
	}
	return false
}

// getModuleName extracts module name from go.mod
func getModuleName(dir string) string {
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			file, err := os.Open(goModPath)
			if err != nil {
				return ""
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "module ") {
					file.Close()
					return strings.TrimSpace(strings.TrimPrefix(line, "module"))
				}
			}
			file.Close()
			return ""
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
