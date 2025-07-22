package decorators

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

// MinifyCode minifies Go code by removing comments and unnecessary spaces
func MinifyCode(inputPath, outputPath string, enabled bool) error {
	if !enabled {
		// If minification is disabled, just copy the file
		return copyFile(inputPath, outputPath)
	}

	// Read file original
	content, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Apply minification
	minifiedContent, err := minifyGoCode(string(content))
	if err != nil {
		return fmt.Errorf("error in minification: %v", err)
	}

	// Escrever file minificado
	if err := os.WriteFile(outputPath, []byte(minifiedContent), 0o600); err != nil {
		return fmt.Errorf("error ao escrever file minificado: %v", err)
	}

	return nil
}

// minifyGoCode minifies Go code while maintaining functionality
func minifyGoCode(code string) (string, error) {
	// Method 1: Use AST to remove comments and format
	minified, err := minifyWithAST(code)
	if err != nil {
		// Fallback: Simple regex-based minification
		return minifyWithRegex(code), nil
	}

	// Apply additional minification
	return applyAdditionalMinification(minified), nil
}

// minifyWithAST uses AST to remove comments and reformat
func minifyWithAST(code string) (string, error) {
	fset := token.NewFileSet()

	// Parse code
	file, err := parser.ParseFile(fset, "", code, parser.ParseComments)
	if err != nil {
		return "", err
	}

	// Remove comments (except build and generation comments)
	file.Comments = filterComments(file.Comments)

	// Reformat code
	var buf strings.Builder
	if err := format.Node(&buf, fset, file); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// filterComments filters comments keeping only essential ones
func filterComments(comments []*ast.CommentGroup) []*ast.CommentGroup {
	var filtered []*ast.CommentGroup

	for _, group := range comments {
		for _, comment := range group.List {
			text := comment.Text

			// Keep important comments
			if strings.Contains(text, "Code generated") ||
				strings.Contains(text, "DO NOT EDIT") ||
				strings.Contains(text, "//go:build") ||
				strings.Contains(text, "+build") {
				filtered = append(filtered, group)
				break
			}
		}
	}

	return filtered
}

// minifyWithRegex simple regex-based minification application
func minifyWithRegex(code string) string {
	lines := strings.Split(code, "\n")
	minified := make([]string, 0, len(lines))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Pular linhas vazias
		if trimmed == "" {
			continue
		}

		// Keep important comments
		if strings.HasPrefix(trimmed, "//") {
			if strings.Contains(trimmed, "Code generated") ||
				strings.Contains(trimmed, "DO NOT EDIT") ||
				strings.Contains(trimmed, "go:build") ||
				strings.Contains(trimmed, "+build") {
				minified = append(minified, line)
			}
			continue
		}

		// Keep important block comments
		if strings.HasPrefix(trimmed, "/*") &&
			(strings.Contains(trimmed, "Code generated") ||
				strings.Contains(trimmed, "DO NOT EDIT")) {
			minified = append(minified, line)
			continue
		}

		// Add code line
		minified = append(minified, line)
	}

	return strings.Join(minified, "\n")
}

// applyAdditionalMinification applies additional minifications
func applyAdditionalMinification(code string) string {
	// Remove multiple consecutive empty lines
	multipleNewlines := regexp.MustCompile(`\n\s*\n\s*\n`)
	code = multipleNewlines.ReplaceAllString(code, "\n\n")

	// Remove excess spaces in lines
	lines := strings.Split(code, "\n")
	cleaned := make([]string, 0, len(lines))

	for _, line := range lines {
		// Keep indentation but clean unnecessary spaces at the end
		cleaned = append(cleaned, strings.TrimRight(line, " \t"))
	}

	// Compact imports when possible
	code = strings.Join(cleaned, "\n")
	code = compactImports(code)

	return code
}

// compactImports compacts import section when possible
func compactImports(code string) string {
	lines := strings.Split(code, "\n")
	var result []string
	inImports := false
	var importLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "import (" {
			inImports = true
			result = append(result, line)
			continue
		}

		if inImports && trimmed == ")" {
			// Process imports coletados
			if len(importLines) > 0 {
				// Remove empty lines between imports
				var compactImports []string
				for _, imp := range importLines {
					if strings.TrimSpace(imp) != "" {
						compactImports = append(compactImports, imp)
					}
				}
				result = append(result, compactImports...)
			}
			result = append(result, line)
			inImports = false
			importLines = nil
			continue
		}

		if inImports {
			importLines = append(importLines, line)
		} else {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// copyFile copies file when minification is disabled
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0o600)
}

// GetMinifiedTemplate returns minified template for generation
func GetMinifiedTemplate() string {
	return `// Code generated by gin-decorators; DO NOT EDIT.
package {{ .PackageName }}
import ("github.com/gin-gonic/gin"
{{- range .Imports }}
{{ . }}
{{- end }}
)
func init() {
{{- range .Routes }}
deco.RegisterRouteWithMeta(deco.RouteEntry{Method:"{{ .Method }}",Path:"{{ .Path }}",Handler:{{ if eq $.PackageName "deco" }}{{ .PackageName }}.{{ .FuncName }}{{ else }}{{ .FuncName }}{{ end }},
{{- if .MiddlewareCalls }}
Middlewares:[]gin.HandlerFunc{
{{- range .MiddlewareCalls }}
{{ . }},
{{- end }}
},
{{- end }}
FuncName:"{{ .FuncName }}",PackageName:"{{ .PackageName }}",
{{- if .Description }}
Description:"{{ .Description }}",
{{- end }}
{{- if .Summary }}
Summary:"{{ .Summary }}",
{{- end }}
{{- if .Tags }}
Tags:[]string{
{{- range .Tags }}
"{{ . }}",
{{- end }}
},
{{- end }}
{{- if .MiddlewareInfo }}
MiddlewareInfo:[]deco.MiddlewareInfo{
{{- range .MiddlewareInfo }}
{Name:"{{ .Name }}",Description:"{{ .Description }}",Args:map[string]interface{}{
{{- range $key, $value := .Args }}
"{{ $key }}":"{{ $value }}",
{{- end }}
}},
{{- end }}
},
{{- end }}
{{- if .Parameters }}
Parameters:[]deco.ParameterInfo{
{{- range .Parameters }}
{Name:"{{ .Name }}",Type:"{{ .Type }}",Location:"{{ .Location }}",Required:{{ .Required }},Description:"{{ .Description }}",Example:"{{ .Example }}"},
{{- end }}
},
{{- end }}
{{- if .Group }}
Group:&deco.GroupInfo{Name:"{{ .Group.Name }}",Prefix:"{{ .Group.Prefix }}",Description:"{{ .Group.Description }}"},
{{- end }}
{{- if .Responses }}
Responses:[]decorators.ResponseInfo{
{{- range .Responses }}
{Code:"{{ .Code }}",Description:"{{ .Description }}",Type:"{{ .Type }}",Example:"{{ .Example }}"},
{{- end }}
},
{{- end }}
})
{{- end }}
}
var GeneratedMetadata=map[string]interface{}{"routes_count":{{ len .Routes }},"generated_at":"{{ .GeneratedAt }}","package_name":"{{ .PackageName }}"}
`
}
