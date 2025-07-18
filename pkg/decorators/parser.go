package decorators

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// Regex to extract route: @Route("METHOD", "path")
	routeRegex = regexp.MustCompile(`@Route\s*\(\s*"([^"]+)"\s*,\s*"([^"]+)"\s*\)`)
)

// ParseDirectory analyzes a directory and extracts route metadata
func ParseDirectory(rootDir string) ([]*RouteMeta, error) {
	var routes []*RouteMeta
	var parseErrors []ValidationError

	// Parse directory
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, rootDir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing do directory %s: %v", rootDir, err)
	}

	// Process each package
	for pkgName, pkg := range pkgs {

		// Process each file in package
		for fileName, file := range pkg.Files {

			fileRoutes, errs := parseFileWithValidation(fset, fileName, file, pkgName)

			routes = append(routes, fileRoutes...)
			parseErrors = append(parseErrors, errs...)
		}
	}

	// Report any parsing errors found
	if len(parseErrors) > 0 {
		return routes, &MultipleValidationError{Errors: parseErrors}
	}

	// Process middlewares for each route
	for _, route := range routes {
		if err := processMiddlewares(route); err != nil {
			return nil, fmt.Errorf("error processing middlewares para %s: %v", route.FuncName, err)
		}
	}

	return routes, nil
}

// parseFileWithValidation analyzes a specific file and validates decorators
func parseFileWithValidation(fset *token.FileSet, fileName string, file *ast.File, pkgName string) ([]*RouteMeta, []ValidationError) {
	var routes []*RouteMeta
	var parseErrors []ValidationError

	// Process each declaration in the file
	for _, decl := range file.Decls {
		// Look for functions
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			route, err := parseFunctionWithValidation(fset, fileName, funcDecl, pkgName)
			if route != nil {
				routes = append(routes, route)
			}
			if err != nil {
				parseErrors = append(parseErrors, *err)
			}
		}

		// Look for structs with @Schema annotations
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			entity := parseEntityFromStruct(fset, fileName, genDecl, pkgName)
			if entity != nil {
				// Convert entity to schema and register it
				schema := convertEntityToSchema(entity)
				RegisterSchema(schema)
				LogVerbose("Schema detected and registered: %s", schema.Name)
			}
		}
	}

	return routes, parseErrors
}

// parseFunctionWithValidation analyzes a function and extracts metadata with validation
func parseFunctionWithValidation(fset *token.FileSet, fileName string, funcDecl *ast.FuncDecl, pkgName string) (*RouteMeta, *ValidationError) {
	// Check if it has comments
	if funcDecl.Doc == nil {
		return nil, nil
	}

	// Join all comments
	var comments []string
	for _, comment := range funcDecl.Doc.List {
		comments = append(comments, comment.Text)
	}
	commentText := strings.Join(comments, "\n")

	// Check if has any decorator annotations
	if !hasDecoratorAnnotations(commentText) {
		return nil, nil
	}

	// Validate decorator syntax first
	if err := validateDecoratorSyntax(fset, fileName, funcDecl, commentText); err != nil {
		return nil, err
	}

	// Extract all markers with validation first
	markers, err := extractMarkersWithValidation(fset, fileName, funcDecl, commentText)
	if err != nil {
		return nil, err
	}

	// Look for @Route
	routeMatches := routeRegex.FindStringSubmatch(commentText)

	if len(routeMatches) != 3 {
		if strings.Contains(commentText, "@Route") {
			pos := fset.Position(funcDecl.Pos())
			return nil, &ValidationError{
				File:    filepath.Base(fileName),
				Line:    pos.Line,
				Message: fmt.Sprintf("Invalid @Route syntax in function %s. Use: @Route(\"METHOD\", \"/path\")", funcDecl.Name.Name),
				Code:    "INVALID_ROUTE_SYNTAX",
			}
		}

		// Markers already extracted above

		// Check if it's a WebSocket handler without route
		hasWebSocketWithArgs := false
		for _, marker := range markers {
			if marker.Name == "WebSocket" && len(marker.Args) > 0 {
				hasWebSocketWithArgs = true
				break
			}
		}

		// If it has @WebSocket with args but no @Route, create a WebSocket-only meta
		if hasWebSocketWithArgs {
			route := &RouteMeta{
				Method:      "", // No HTTP method for pure WebSocket handlers
				Path:        "", // No HTTP path for pure WebSocket handlers
				FuncName:    funcDecl.Name.Name,
				PackageName: pkgName,
				FileName:    filepath.Base(fileName),
				Markers:     markers,
			}
			return route, nil
		}
		return nil, nil // Not a handler
	}

	method := routeMatches[1]
	path := routeMatches[2]
	funcName := funcDecl.Name.Name

	// Validate method
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	if !contains(validMethods, method) {
		pos := fset.Position(funcDecl.Pos())
		return nil, &ValidationError{
			File:    filepath.Base(fileName),
			Line:    pos.Line,
			Message: fmt.Sprintf("Invalid HTTP method '%s' in function %s. Valid methods: %v", method, funcName, validMethods),
			Code:    "INVALID_HTTP_METHOD",
		}
	}

	// Validate path
	if !strings.HasPrefix(path, "/") {
		pos := fset.Position(funcDecl.Pos())
		return nil, &ValidationError{
			File:    filepath.Base(fileName),
			Line:    pos.Line,
			Message: fmt.Sprintf("Invalid path '%s' in function %s. Path must start with '/'", path, funcName),
			Code:    "INVALID_PATH",
		}
	}

	// Markers already extracted above

	route := &RouteMeta{
		Method:      method,
		Path:        path,
		FuncName:    funcName,
		PackageName: pkgName,
		FileName:    filepath.Base(fileName),
		Markers:     markers,
	}

	return route, nil
}

// hasDecoratorAnnotations checks if comment text contains any decorator annotations
func hasDecoratorAnnotations(commentText string) bool {
	decorators := []string{"@Route", "@Middleware", "@Response", "@RequestBody", "@Schema", "@Summary", "@Description", "@Tag", "@Validate", "@WebSocket", "@WebSocketStats"}
	for _, decorator := range decorators {
		if strings.Contains(commentText, decorator) {
			return true
		}
	}
	return false
}

// validateDecoratorSyntax validates the overall syntax of decorators in comments
func validateDecoratorSyntax(fset *token.FileSet, fileName string, funcDecl *ast.FuncDecl, commentText string) *ValidationError {
	pos := fset.Position(funcDecl.Pos())

	// Check for common syntax errors
	lines := strings.Split(commentText, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			trimmed = strings.TrimSpace(trimmed[2:])
		}

		// Check for malformed decorators
		if strings.HasPrefix(trimmed, "@") {
			// Check for missing parentheses in decorators that require them
			if (strings.Contains(trimmed, "@Route") || strings.Contains(trimmed, "@Response") ||
				strings.Contains(trimmed, "@RequestBody") || strings.Contains(trimmed, "@Middleware")) &&
				!strings.Contains(trimmed, "(") {
				return &ValidationError{
					File:    filepath.Base(fileName),
					Line:    pos.Line + i,
					Message: fmt.Sprintf("Malformed decorator: '%s'. Missing parentheses", trimmed),
					Code:    "MALFORMED_DECORATOR",
				}
			}

			// Check for unmatched quotes
			quoteCount := strings.Count(trimmed, "\"")
			if quoteCount%2 != 0 {
				return &ValidationError{
					File:    filepath.Base(fileName),
					Line:    pos.Line + i,
					Message: fmt.Sprintf("Unmatched quotes in: '%s'", trimmed),
					Code:    "UNMATCHED_QUOTES",
				}
			}

			// Check for unmatched parentheses
			openParen := strings.Count(trimmed, "(")
			closeParen := strings.Count(trimmed, ")")
			if openParen != closeParen {
				return &ValidationError{
					File:    filepath.Base(fileName),
					Line:    pos.Line + i,
					Message: fmt.Sprintf("Unmatched parentheses in: '%s'", trimmed),
					Code:    "UNMATCHED_PARENTHESES",
				}
			}
		}
	}

	return nil
}

// extractMarkersWithValidation extracts all markers from a comment with validation
func extractMarkersWithValidation(fset *token.FileSet, fileName string, funcDecl *ast.FuncDecl, commentText string) ([]MarkerInstance, *ValidationError) {
	var markers []MarkerInstance
	pos := fset.Position(funcDecl.Pos())

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
				args, err := parseArgumentsWithValidation(match[1], name)
				if err != nil {
					return nil, &ValidationError{
						File:    filepath.Base(fileName),
						Line:    pos.Line,
						Message: fmt.Sprintf("Error in @%s decorator arguments: %s", name, err.Error()),
						Code:    "INVALID_ARGUMENTS",
					}
				}
				marker.Args = args
			}

			markers = append(markers, marker)
		}
	}

	return markers, nil
}

// parseArgumentsWithValidation converts argument string to slice with validation
func parseArgumentsWithValidation(argsStr, decoratorName string) ([]string, error) {
	if argsStr == "" {
		return nil, nil
	}

	var args []string
	parts := strings.Split(argsStr, ",")
	for _, part := range parts {
		arg := strings.TrimSpace(part)
		if arg != "" {
			// Remove quotes if present
			if (strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"")) ||
				(strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'")) {
				arg = arg[1 : len(arg)-1]
			}

			// Validate argument is not empty after processing
			if arg == "" {
				return nil, fmt.Errorf("empty argument found")
			}

			args = append(args, arg)
		}
	}

	// Validate argument count for specific decorators
	if err := validateArgumentCount(decoratorName, args); err != nil {
		return nil, err
	}

	return args, nil
}

// validateArgumentCount validates the number of arguments for specific decorators
func validateArgumentCount(decoratorName string, args []string) error {
	switch decoratorName {
	case "Route":
		if len(args) != 2 {
			return fmt.Errorf("@Route requires exactly 2 arguments (method, path), found %d", len(args))
		}
	case "Response":
		if len(args) == 0 {
			return fmt.Errorf("@Response requires at least 1 argument (status code)")
		}
	case "RequestBody":
		if len(args) == 0 {
			return fmt.Errorf("@RequestBody requires at least 1 argument (type)")
		}
	}
	return nil
}

// contains checks if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// MultipleValidationError represents multiple validation errors
type MultipleValidationError struct {
	Errors []ValidationError
}

func (e *MultipleValidationError) Error() string {
	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "\n")
}

// parseArguments converts string of arguments to slice
func parseArguments(argsStr string) []string {
	if argsStr == "" {
		return nil
	}

	var args []string
	parts := strings.Split(argsStr, ",")
	for _, part := range parts {
		arg := strings.TrimSpace(part)
		if arg != "" {
			args = append(args, arg)
		}
	}

	return args
}

// processMiddlewares generates middleware calls for a route
func processMiddlewares(route *RouteMeta) error {
	var middlewareCalls []string
	var middlewareInfo []MiddlewareInfo
	var parameters []ParameterInfo
	var tags []string
	var responses []ResponseInfo // Changed to []ResponseInfo
	var groupInfo *GroupInfo

	// Process each marker
	for _, marker := range route.Markers {
		switch marker.Name {
		case "Auth", "Cache", "RateLimit", "Metrics", "CORS", "WebSocketStats", "Proxy", "Security":
			// Traditional middlewares
			call := generateMiddlewareCall(marker)
			if call != "" {
				middlewareCalls = append(middlewareCalls, call)

				// Add middleware information
				info := MiddlewareInfo{
					Name:        marker.Name,
					Args:        parseArgsToMap(marker.Args),
					Description: getMiddlewareDescription(marker.Name),
				}
				middlewareInfo = append(middlewareInfo, info)
			}

		case "WebSocket":
			// WebSocket can be both middleware and handler registration
			call := generateMiddlewareCall(marker)
			if call != "" {
				middlewareCalls = append(middlewareCalls, call)

				// Add middleware information
				info := MiddlewareInfo{
					Name:        marker.Name,
					Args:        parseArgsToMap(marker.Args),
					Description: getMiddlewareDescription(marker.Name),
				}
				middlewareInfo = append(middlewareInfo, info)
			}

			// If has args, register as WebSocket handler
			if len(marker.Args) > 0 {
				for _, arg := range marker.Args {
					messageType := strings.Trim(arg, `"' `)
					if messageType != "" {
						route.WebSocketHandlers = append(route.WebSocketHandlers, messageType)
					}
				}
			}

		case "Group":
			// Process grupo
			if len(marker.Args) > 0 {
				groupName := strings.Trim(marker.Args[0], `"`)
				groupInfo = GetGroup(groupName)
				if groupInfo == nil {
					// Create group if it does not exist
					prefix := "/" + strings.ToLower(groupName)
					description := fmt.Sprintf("Grupo %s", groupName)
					if len(marker.Args) > 1 {
						prefix = strings.Trim(marker.Args[1], `"`)
					}
					if len(marker.Args) > 2 {
						description = strings.Trim(marker.Args[2], `"`)
					}
					groupInfo = RegisterGroup(groupName, prefix, description)
				}
			}

		case "Param":
			// Process parameter: @Param(name="id", type="string", location="path", required=true, description="User ID")
			param := parseParameterInfo(marker.Args)
			if param.Name != "" {
				parameters = append(parameters, param)
			}

		case "Tag":
			// Add tag
			if len(marker.Args) > 0 {
				tag := strings.Trim(marker.Args[0], `"`)
				tags = append(tags, tag)
			}

		case "Response":
			// Process response: @Response(code="200", description="Success", type="UserResponse")
			response := parseResponseInfo(marker.Args)
			if response.Code != "" && response.Description != "" {
				responses = append(responses, response)
			}

		case "Description":
			// Description will be processed at route level
			if len(marker.Args) > 0 {
				route.Description = strings.Trim(marker.Args[0], `"`)
			}

		case "Summary":
			// Summary will be processed at route level
			if len(marker.Args) > 0 {
				route.Summary = strings.Trim(marker.Args[0], `"`)
			}
		}
	}

	route.MiddlewareCalls = middlewareCalls
	route.MiddlewareInfo = middlewareInfo
	route.Parameters = parameters
	route.Tags = tags
	route.Responses = responses
	route.Group = groupInfo

	return nil
}

// parseArgsToMap converts arguments to map[string]interface{}
func parseArgsToMap(args []string) map[string]interface{} {
	result := make(map[string]interface{})

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
			result[key] = value
		} else {
			// Argument without key, use as "value"
			result["value"] = strings.Trim(arg, `"`)
		}
	}

	return result
}

// parseParameterInfo converts arguments to ParameterInfo
func parseParameterInfo(args []string) ParameterInfo {
	param := ParameterInfo{}

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)

			switch key {
			case "name":
				param.Name = value
			case "type":
				param.Type = value
			case "location":
				param.Location = value
			case "required":
				param.Required = value == "true"
			case "description":
				param.Description = value
			case "example":
				param.Example = value
			}
		}
	}

	return param
}

// parseResponseInfo converts arguments to ResponseInfo
func parseResponseInfo(args []string) ResponseInfo {
	response := ResponseInfo{}

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), `"`)

			switch key {
			case "code":
				response.Code = value
			case "description":
				response.Description = value
			case "type":
				response.Type = value
			case "example":
				response.Example = value
			}
		}
	}

	return response
}

// getMiddlewareDescription returns default description for middlewares
func getMiddlewareDescription(name string) string {
	descriptions := map[string]string{
		"Auth":           "Middleware de autenticação e autorização",
		"Cache":          "Middleware de cache de responses",
		"RateLimit":      "Middleware de limitação de taxa",
		"Metrics":        "Middleware de coleta de métricas",
		"CORS":           "Middleware de Cross-Origin Resource Sharing",
		"WebSocket":      "Middleware de upgrade para conexão WebSocket",
		"WebSocketStats": "Middleware de estatísticas WebSocket",
		"Proxy":          "Middleware de proxy reverso com service discovery e load balancing",
	}

	if desc, exists := descriptions[name]; exists {
		return desc
	}
	return fmt.Sprintf("Middleware %s", name)
}

// generateMiddlewareCall generates Go call for a middleware
func generateMiddlewareCall(marker MarkerInstance) string {
	switch marker.Name {
	case "Auth":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateAuthMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateAuthMiddleware("")`

	case "Cache":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateCacheMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateCacheMiddleware("duration=5m")`

	case "RateLimit":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateRateLimitMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateRateLimitMiddleware("limit=100,window=1m")`

	case "Metrics":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateMetricsMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateMetricsMiddleware("")`

	case "CORS":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateCORSMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateCORSMiddleware("")`

	case "WebSocket":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateWebSocketMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateWebSocketMiddleware("")`

	case "WebSocketStats":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateWebSocketStatsMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateWebSocketStatsMiddleware("")`

	case "Proxy":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateProxyMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateProxyMiddleware("")`

	case "Security":
		if len(marker.Args) > 0 {
			return fmt.Sprintf(`deco.CreateSecurityMiddleware(%q)`, strings.Join(marker.Args, ","))
		}
		return `deco.CreateSecurityMiddleware("")`
	}

	return ""
}

// CreateAuthMiddleware creates auth middleware (wrapper for generation)
func CreateAuthMiddleware(args string) func(c *gin.Context) {
	argsSlice := parseArguments(args)
	config := GetMarkers()["Auth"]
	return config.Factory(argsSlice)
}

// CreateCacheMiddleware creates cache middleware (wrapper for generation)
func CreateCacheMiddleware(args string) func(c *gin.Context) {
	argsSlice := parseArguments(args)
	config := GetMarkers()["Cache"]
	return config.Factory(argsSlice)
}

// CreateRateLimitMiddleware creates rate limit middleware (wrapper for generation)
func CreateRateLimitMiddleware(args string) func(c *gin.Context) {
	argsSlice := parseArguments(args)
	config := GetMarkers()["RateLimit"]
	return config.Factory(argsSlice)
}

// CreateMetricsMiddleware creates metrics middleware (wrapper for generation)
func CreateMetricsMiddleware(args string) func(c *gin.Context) {
	argsSlice := parseArguments(args)
	config := GetMarkers()["Metrics"]
	return config.Factory(argsSlice)
}

// CreateCORSMiddleware creates CORS middleware (wrapper for generation)
func CreateCORSMiddleware(args string) func(c *gin.Context) {
	argsSlice := parseArguments(args)
	config := GetMarkers()["CORS"]
	return config.Factory(argsSlice)
}

// CreateWebSocketMiddleware creates WebSocket middleware (wrapper for generation)
func CreateWebSocketMiddleware(args string) gin.HandlerFunc {
	argsSlice := parseArguments(args)
	config := GetMarkers()["WebSocket"]
	return config.Factory(argsSlice)
}

// CreateWebSocketStatsMiddleware creates WebSocket stats middleware (wrapper for generation)
func CreateWebSocketStatsMiddleware(args string) gin.HandlerFunc {
	argsSlice := parseArguments(args)
	config := GetMarkers()["WebSocketStats"]
	return config.Factory(argsSlice)
}

// CreateProxyMiddleware creates proxy middleware (wrapper for generation)
func CreateProxyMiddleware(args string) gin.HandlerFunc {
	argsSlice := parseArguments(args)
	config := GetMarkers()["Proxy"]
	return config.Factory(argsSlice)
}

// CreateSecurityMiddleware creates security middleware (wrapper for generation)
func CreateSecurityMiddleware(args string) gin.HandlerFunc {
	argsSlice := parseArguments(args)
	config := GetMarkers()["Security"]
	return config.Factory(argsSlice)
}
