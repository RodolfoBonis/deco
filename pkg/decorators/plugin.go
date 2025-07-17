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
	// Automatic logging plugin
	RegisterParserHook(func(routes []*RouteMeta) error {
		LogVerbose("Plugin de logging: %d routes processadas", len(routes))
		for _, route := range routes {
			LogVerbose("  - %s %s -> %s", route.Method, route.Path, route.FuncName)
		}
		return nil
	})

	// Plugin that adds default imports
	RegisterGeneratorHook(func(data *GenData) error {
		// Ensure necessary imports
		requiredImports := []string{
			`deco "github.com/RodolfoBonis/deco"`,
		}

		// If the generated package is deco, add handlers import
		if data.PackageName == "deco" && len(data.Routes) > 0 {
			// Automatically detect handlers path based on the first route
			firstRoute := data.Routes[0]
			if firstRoute.PackageName == "handlers" {
				// Detect current directory to build relative import
				wd, err := os.Getwd()
				if err == nil {
					// Check if go.mod exists to extract module name
					if modName := getModuleName(wd); modName != "" {
						// Build handlers path based on working directory
						handlerImport := ""

						// Remove part of module path from working directory to get relative path
						if strings.Contains(wd, modName) {
							// Extract part after module name
							parts := strings.Split(wd, modName)
							if len(parts) > 1 {
								relativePath := strings.TrimPrefix(parts[1], "/")
								if relativePath != "" {
									handlerImport = fmt.Sprintf(`handlers "%s/%s/handlers"`, modName, relativePath)
								} else {
									handlerImport = fmt.Sprintf(`handlers "%s/handlers"`, modName)
								}
							}
						} else {
							// Fallback - assume default structure
							handlerImport = fmt.Sprintf(`handlers "%s/handlers"`, modName)
						}

						if handlerImport != "" {
							requiredImports = append(requiredImports, handlerImport)
						}
					}
				}
			}
		}

		for _, imp := range requiredImports {
			found := false
			for _, existing := range data.Imports {
				if existing == imp {
					found = true
					break
				}
			}
			if !found {
				data.Imports = append(data.Imports, imp)
			}
		}

		LogVerbose("Plugin de imports: %d imports configurados", len(data.Imports))
		return nil
	})
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
