package decorators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// SDKGenerator interface for different SDK generators
type SDKGenerator interface {
	Generate(spec *OpenAPISpec, config *ClientSDKConfig) error
	GetLanguage() string
	GetFileExtension() string
}

// GoSDKGenerator generator for Go
type GoSDKGenerator struct{}

// PythonSDKGenerator generator for Python
type PythonSDKGenerator struct{}

// JavaScriptSDKGenerator generator for JavaScript
type JavaScriptSDKGenerator struct{}

// TypeScriptSDKGenerator generator for TypeScript
type TypeScriptSDKGenerator struct{}

// SDKManager manages SDK generation
type SDKManager struct {
	generators map[string]SDKGenerator
	config     ClientSDKConfig
}

// NewSDKManager creates new SDK manager
func NewSDKManager(config *ClientSDKConfig) *SDKManager {
	manager := &SDKManager{
		generators: make(map[string]SDKGenerator),
		config:     *config,
	}

	// Register generatores
	manager.RegisterGenerator("go", &GoSDKGenerator{})
	manager.RegisterGenerator("python", &PythonSDKGenerator{})
	manager.RegisterGenerator("javascript", &JavaScriptSDKGenerator{})
	manager.RegisterGenerator("typescript", &TypeScriptSDKGenerator{})

	return manager
}

// RegisterGenerator registers new generator
func (sm *SDKManager) RegisterGenerator(language string, generator SDKGenerator) {
	sm.generators[language] = generator
}

// GenerateSDKs generates SDKs for all configured languages
func (sm *SDKManager) GenerateSDKs(spec *OpenAPISpec) error {
	if !sm.config.Enabled {
		return nil
	}

	// Create output directory
	if err := os.MkdirAll(sm.config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("error creating output directory: %v", err)
	}

	// Generate for each language
	for _, language := range sm.config.Languages {
		if generator, exists := sm.generators[language]; exists {
			fmt.Printf("Gerando SDK para %s...\n", language)
			if err := generator.Generate(spec, &sm.config); err != nil {
				return fmt.Errorf("error ao gerar SDK para %s: %v", language, err)
			}
		} else {
			fmt.Printf("Generator not found para linguagem: %s\n", language)
		}
	}

	return nil
}

// Go SDK Generator

// GetLanguage retorna a linguagem de programação usada.
func (g *GoSDKGenerator) GetLanguage() string {
	return "go"
}

// GetFileExtension retorna a extensão de arquivo para a linguagem.
func (g *GoSDKGenerator) GetFileExtension() string {
	return ".go"
}

// Generate creates a Go client SDK from the OpenAPI specification
func (g *GoSDKGenerator) Generate(spec *OpenAPISpec, config *ClientSDKConfig) error {
	outputDir := filepath.Join(config.OutputDir, "go")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	// Template do cliente Go
	tmpl := `// Package {{.PackageName}} provides a client for the {{.ServiceName}} API
// Generated automatically by gin-decorators on {{.GeneratedAt}}
package {{.PackageName}}

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents the API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	UserAgent  string
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: "{{.PackageName}}-go-client/1.0.0",
	}
}

// SetAPIKey sets the API key for authentication
func (c *Client) SetAPIKey(apiKey string) {
	c.APIKey = apiKey
}

// SetTimeout sets the HTTP client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.HTTPClient.Timeout = timeout
}

{{range .Endpoints}}
// {{.FunctionName}} {{.Description}}
func (c *Client) {{.FunctionName}}(ctx context.Context{{.ParametersSignature}}) ({{.ReturnType}}, error) {
	{{.URLConstruction}}
	
	{{.RequestBody}}
	
	req, err := http.NewRequestWithContext(ctx, "{{.Method}}", url, {{.RequestBodyVar}})
	if err != nil {
		return {{.ZeroValue}}, fmt.Errorf("error creating request: %w", err)
	}

	{{.Headers}}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return {{.ZeroValue}}, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return {{.ZeroValue}}, fmt.Errorf("API error: %s", resp.Status)
	}

	{{.ResponseHandling}}
}
{{end}}

// Error represents an API error
type Error struct {
	Code    int    ` + "`json:\"code\"`" + `
	Message string ` + "`json:\"message\"`" + `
}

func (e Error) Error() string {
	return fmt.Sprintf("API error %d: %s", e.Code, e.Message)
}
`

	data := g.prepareTemplateData(spec, config)
	return g.executeTemplate(tmpl, data, filepath.Join(outputDir, "client.go"))
}

func (g *GoSDKGenerator) prepareTemplateData(spec *OpenAPISpec, config *ClientSDKConfig) map[string]interface{} {
	endpoints := make([]map[string]interface{}, 0)

	for path, pathItem := range spec.Paths {
		for method, operation := range pathItem {
			endpoint := map[string]interface{}{
				"FunctionName":        g.generateFunctionName(method, path),
				"Description":         operation.Summary,
				"Method":              strings.ToUpper(method),
				"Path":                path,
				"ParametersSignature": g.generateParametersSignature(operation.Parameters),
				"URLConstruction":     g.generateURLConstruction(path, operation.Parameters),
				"RequestBody":         g.generateRequestBody(operation.RequestBody),
				"RequestBodyVar":      g.getRequestBodyVar(operation.RequestBody),
				"Headers":             g.generateHeaders(),
				"ReturnType":          g.generateReturnType(operation.Responses),
				"ZeroValue":           g.generateZeroValue(operation.Responses),
				"ResponseHandling":    g.generateResponseHandling(operation.Responses),
			}
			endpoints = append(endpoints, endpoint)
		}
	}

	return map[string]interface{}{
		"PackageName": config.PackageName,
		"ServiceName": spec.Info.Title,
		"GeneratedAt": time.Now().Format("2006-01-02 15:04:05"),
		"Endpoints":   endpoints,
	}
}

func (g *GoSDKGenerator) generateFunctionName(method, path string) string {
	// Convert method and path to function name
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var name strings.Builder

	// Use cases.Title instead of deprecated strings.Title
	caser := cases.Title(language.English)
	name.WriteString(caser.String(strings.ToLower(method)))

	for _, part := range parts {
		if !strings.HasPrefix(part, "{") {
			name.WriteString(caser.String(part))
		}
	}

	return name.String()
}

func (g *GoSDKGenerator) generateParametersSignature(params []OpenAPIParameter) string {
	parts := make([]string, 0, len(params))
	for _, param := range params {
		goType := g.convertTypeToGo(param.Schema.Type)
		parts = append(parts, fmt.Sprintf(", %s %s", param.Name, goType))
	}
	return strings.Join(parts, "")
}

func (g *GoSDKGenerator) generateURLConstruction(path string, params []OpenAPIParameter) string {
	// Replace path parameters and build query
	code := fmt.Sprintf("url := c.BaseURL + %q", path)

	// Replace path parameters
	for _, param := range params {
		if param.In == "path" {
			code = strings.ReplaceAll(code, "{"+param.Name+"}", fmt.Sprintf("\" + %s + \"", param.Name))
		}
	}

	// Add query parameters
	queryParams := make([]string, 0)
	for _, param := range params {
		if param.In == "query" {
			queryParams = append(queryParams, param.Name)
		}
	}

	if len(queryParams) > 0 {
		code += "\n\tvalues := url.Values{}\n"
		for _, param := range queryParams {
			code += fmt.Sprintf("\tvalues.Set(%q, %s)\n", param, param)
		}
		code += "\turl += \"?\" + values.Encode()"
	}

	return code
}

func (g *GoSDKGenerator) generateRequestBody(body *OpenAPIRequestBody) string {
	if body == nil {
		return "var body io.Reader"
	}
	return `jsonBody, _ := json.Marshal(requestBody)
	body := bytes.NewBuffer(jsonBody)`
}

func (g *GoSDKGenerator) getRequestBodyVar(body *OpenAPIRequestBody) string {
	if body == nil {
		return "nil"
	}
	return "body"
}

func (g *GoSDKGenerator) generateHeaders() string {
	return `req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer " + c.APIKey)
	}`
}

func (g *GoSDKGenerator) generateReturnType(_ map[string]OpenAPIResponse) string {
	return "interface{}" // Simplificado
}

func (g *GoSDKGenerator) generateZeroValue(_ map[string]OpenAPIResponse) string {
	return "nil"
}

func (g *GoSDKGenerator) generateResponseHandling(_ map[string]OpenAPIResponse) string {
	return `body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return result, nil`
}

func (g *GoSDKGenerator) convertTypeToGo(openAPIType string) string {
	switch openAPIType {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "array":
		return "[]interface{}"
	default:
		return "interface{}"
	}
}

func (g *GoSDKGenerator) executeTemplate(tmplStr string, data interface{}, outputPath string) error {
	tmpl, err := template.New("client").Parse(tmplStr)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// Python SDK Generator

// GetLanguage retorna a linguagem de programação usada.
func (p *PythonSDKGenerator) GetLanguage() string {
	return "python"
}

// GetFileExtension retorna a extensão de arquivo para a linguagem.
func (p *PythonSDKGenerator) GetFileExtension() string {
	return ".py"
}

// Generate creates a Python client SDK from the OpenAPI specification
func (p *PythonSDKGenerator) Generate(spec *OpenAPISpec, config *ClientSDKConfig) error {
	outputDir := filepath.Join(config.OutputDir, "python")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	// Template do cliente Python
	tmpl := `"""
{{.ServiceName}} API Client
Generated automatically by gin-decorators on {{.GeneratedAt}}
"""

import requests
import json
from typing import Dict, Any, Optional
from urllib.parse import urljoin, urlencode


class {{.ClassName}}:
    """Client for {{.ServiceName}} API"""
    
    def __init__(self, base_url: str, api_key: Optional[str] = None):
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'User-Agent': '{{.PackageName}}-python-client/1.0.0'
        })
        
        if api_key:
            self.session.headers['Authorization'] = f'Bearer {api_key}'
    
    def set_api_key(self, api_key: str):
        """Set API key for authentication"""
        self.api_key = api_key
        self.session.headers['Authorization'] = f'Bearer {api_key}'
    
{{range .Endpoints}}
    def {{.FunctionName}}(self{{.ParametersSignature}}) -> Dict[str, Any]:
        """{{.Description}}"""
        {{.URLConstruction}}
        
        {{.RequestBody}}
        
        response = self.session.{{.Method}}(url{{.RequestBodyParam}})
        response.raise_for_status()
        
        return response.json()
{{end}}


class APIError(Exception):
    """API Error Exception"""
    
    def __init__(self, message: str, status_code: int = None):
        self.message = message
        self.status_code = status_code
        super().__init__(self.message)
`

	data := p.prepareTemplateData(spec, config)
	return p.executeTemplate(tmpl, data, filepath.Join(outputDir, "client.py"))
}

func (p *PythonSDKGenerator) prepareTemplateData(spec *OpenAPISpec, config *ClientSDKConfig) map[string]interface{} {
	// Use cases.Title instead of deprecated strings.Title
	caser := cases.Title(language.English)
	className := caser.String(config.PackageName) + "Client"
	endpoints := make([]map[string]interface{}, 0)

	for path, pathItem := range spec.Paths {
		for method, operation := range pathItem {
			endpoint := map[string]interface{}{
				"FunctionName":        p.generateFunctionName(method, path),
				"Description":         operation.Summary,
				"Method":              strings.ToLower(method),
				"Path":                path,
				"ParametersSignature": p.generateParametersSignature(operation.Parameters),
				"URLConstruction":     p.generateURLConstruction(path, operation.Parameters),
				"RequestBody":         p.generateRequestBody(operation.RequestBody),
				"RequestBodyParam":    p.getRequestBodyParam(operation.RequestBody),
			}
			endpoints = append(endpoints, endpoint)
		}
	}

	return map[string]interface{}{
		"PackageName": config.PackageName,
		"ClassName":   className,
		"ServiceName": spec.Info.Title,
		"GeneratedAt": time.Now().Format("2006-01-02 15:04:05"),
		"Endpoints":   endpoints,
	}
}

func (p *PythonSDKGenerator) generateFunctionName(method, path string) string {
	// Convert to snake_case
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var name strings.Builder
	name.WriteString(strings.ToLower(method))

	for _, part := range parts {
		if !strings.HasPrefix(part, "{") {
			name.WriteString("_")
			name.WriteString(strings.ToLower(part))
		}
	}

	return name.String()
}

func (p *PythonSDKGenerator) generateParametersSignature(params []OpenAPIParameter) string {
	parts := make([]string, 0, len(params))
	for _, param := range params {
		pythonType := p.convertTypeToPython(param.Schema.Type)
		parts = append(parts, fmt.Sprintf(", %s: %s", param.Name, pythonType))
	}
	return strings.Join(parts, "")
}

func (p *PythonSDKGenerator) generateURLConstruction(path string, params []OpenAPIParameter) string {
	code := fmt.Sprintf("url = f'{self.base_url}%s'", path)

	// Replace path parameters
	for _, param := range params {
		if param.In == "path" {
			code = strings.ReplaceAll(code, "{"+param.Name+"}", fmt.Sprintf("{%s}", param.Name))
		}
	}

	// Add query parameters
	queryParams := make([]string, 0)
	for _, param := range params {
		if param.In == "query" {
			queryParams = append(queryParams, param.Name)
		}
	}

	if len(queryParams) > 0 {
		code += "\n        params = {}\n"
		for _, param := range queryParams {
			code += fmt.Sprintf("        params['%s'] = %s\n", param, param)
		}
		code += "        url += '?' + urlencode(params)"
	}

	return code
}

func (p *PythonSDKGenerator) generateRequestBody(body *OpenAPIRequestBody) string {
	if body == nil {
		return ""
	}
	return "json_data = json.dumps(request_body)"
}

func (p *PythonSDKGenerator) getRequestBodyParam(body *OpenAPIRequestBody) string {
	if body == nil {
		return ""
	}
	return ", json=request_body"
}

func (p *PythonSDKGenerator) convertTypeToPython(openAPIType string) string {
	switch openAPIType {
	case "string":
		return "str"
	case "integer":
		return "int"
	case "number":
		return "float"
	case "boolean":
		return "bool"
	case "array":
		return "list"
	default:
		return "Any"
	}
}

func (p *PythonSDKGenerator) executeTemplate(tmplStr string, data interface{}, outputPath string) error {
	tmpl, err := template.New("client").Parse(tmplStr)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// JavaScript SDK Generator

// GetLanguage retorna a linguagem de programação usada.
func (j *JavaScriptSDKGenerator) GetLanguage() string {
	return "javascript"
}

// GetFileExtension retorna a extensão de arquivo para a linguagem.
func (j *JavaScriptSDKGenerator) GetFileExtension() string {
	return ".js"
}

// Generate creates a JavaScript client SDK from the OpenAPI specification
func (j *JavaScriptSDKGenerator) Generate(spec *OpenAPISpec, config *ClientSDKConfig) error {
	outputDir := filepath.Join(config.OutputDir, "javascript")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	// Template JavaScript
	tmpl := `/**
 * {{.ServiceName}} API Client
 * Generated automatically by gin-decorators on {{.GeneratedAt}}
 */

class {{.ClassName}} {
    constructor(baseURL, apiKey = null) {
        this.baseURL = baseURL.replace(/\/$/, '');
        this.apiKey = apiKey;
        this.defaultHeaders = {
            'Content-Type': 'application/json',
            'User-Agent': '{{.PackageName}}-js-client/1.0.0'
        };
        
        if (apiKey) {
            this.defaultHeaders['Authorization'] = ` + "`Bearer ${apiKey}`" + `;
        }
    }
    
    setApiKey(apiKey) {
        this.apiKey = apiKey;
        this.defaultHeaders['Authorization'] = ` + "`Bearer ${apiKey}`" + `;
    }

{{range .Endpoints}}
    async {{.FunctionName}}({{.ParametersSignature}}) {
        {{.URLConstruction}}
        
        const options = {
            method: '{{.Method}}',
            headers: this.defaultHeaders{{.RequestBody}}
        };
        
        const response = await fetch(url, options);
        
        if (!response.ok) {
            throw new Error(` + "`API Error: ${response.status} ${response.statusText}`" + `);
        }
        
        return await response.json();
    }
{{end}}
}

module.exports = {{.ClassName}};
`

	data := j.prepareTemplateData(spec, config)
	return j.executeTemplate(tmpl, data, filepath.Join(outputDir, "client.js"))
}

func (j *JavaScriptSDKGenerator) prepareTemplateData(spec *OpenAPISpec, config *ClientSDKConfig) map[string]interface{} {
	// Use cases.Title instead of deprecated strings.Title
	caser := cases.Title(language.English)
	className := caser.String(config.PackageName) + "Client"
	endpoints := make([]map[string]interface{}, 0)

	for path, pathItem := range spec.Paths {
		for method, operation := range pathItem {
			endpoint := map[string]interface{}{
				"FunctionName":        j.generateFunctionName(method, path),
				"Method":              strings.ToUpper(method),
				"ParametersSignature": j.generateParametersSignature(operation.Parameters),
				"URLConstruction":     j.generateURLConstruction(path, operation.Parameters),
				"RequestBody":         j.generateRequestBody(operation.RequestBody),
			}
			endpoints = append(endpoints, endpoint)
		}
	}

	return map[string]interface{}{
		"PackageName": config.PackageName,
		"ClassName":   className,
		"ServiceName": spec.Info.Title,
		"GeneratedAt": time.Now().Format("2006-01-02 15:04:05"),
		"Endpoints":   endpoints,
	}
}

func (j *JavaScriptSDKGenerator) generateFunctionName(method, path string) string {
	// Convert to camelCase
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var name strings.Builder
	name.WriteString(strings.ToLower(method))

	// Use cases.Title instead of deprecated strings.Title
	caser := cases.Title(language.English)
	for _, part := range parts {
		if !strings.HasPrefix(part, "{") {
			name.WriteString(caser.String(part))
		}
	}

	return name.String()
}

func (j *JavaScriptSDKGenerator) generateParametersSignature(params []OpenAPIParameter) string {
	parts := make([]string, 0, len(params))
	for _, param := range params {
		parts = append(parts, param.Name)
	}
	return strings.Join(parts, ", ")
}

func (j *JavaScriptSDKGenerator) generateURLConstruction(path string, params []OpenAPIParameter) string {
	code := fmt.Sprintf("let url = `${this.baseURL}%s`;", path)

	// Replace path parameters
	for _, param := range params {
		if param.In == "path" {
			code = strings.ReplaceAll(code, "{"+param.Name+"}", "${"+param.Name+"}")
		}
	}

	// Add query parameters
	queryParams := make([]string, 0)
	for _, param := range params {
		if param.In == "query" {
			queryParams = append(queryParams, param.Name)
		}
	}

	if len(queryParams) > 0 {
		code += "\n        const params = new URLSearchParams();\n"
		for _, param := range queryParams {
			code += fmt.Sprintf("        params.append('%s', %s);\n", param, param)
		}
		code += "        url += '?' + params.toString();"
	}

	return code
}

func (j *JavaScriptSDKGenerator) generateRequestBody(body *OpenAPIRequestBody) string {
	if body == nil {
		return ""
	}
	return ",\n            body: JSON.stringify(requestBody)"
}

func (j *JavaScriptSDKGenerator) executeTemplate(tmplStr string, data interface{}, outputPath string) error {
	tmpl, err := template.New("client").Parse(tmplStr)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// TypeScript SDK Generator

// GetLanguage retorna a linguagem de programação usada.
func (t *TypeScriptSDKGenerator) GetLanguage() string {
	return "typescript"
}

// GetFileExtension retorna a extensão de arquivo para a linguagem.
func (t *TypeScriptSDKGenerator) GetFileExtension() string {
	return ".ts"
}

// Generate creates a TypeScript client SDK from the OpenAPI specification
func (t *TypeScriptSDKGenerator) Generate(spec *OpenAPISpec, config *ClientSDKConfig) error {
	outputDir := filepath.Join(config.OutputDir, "typescript")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}

	// TypeScript template (similar to JavaScript but with types)
	tmpl := `/**
 * {{.ServiceName}} API Client
 * Generated automatically by gin-decorators on {{.GeneratedAt}}
 */

export class {{.ClassName}} {
    private baseURL: string;
    private apiKey: string | null;
    private defaultHeaders: Record<string, string>;

    constructor(baseURL: string, apiKey: string | null = null) {
        this.baseURL = baseURL.replace(/\/$/, '');
        this.apiKey = apiKey;
        this.defaultHeaders = {
            'Content-Type': 'application/json',
            'User-Agent': '{{.PackageName}}-ts-client/1.0.0'
        };
        
        if (apiKey) {
            this.defaultHeaders['Authorization'] = ` + "`Bearer ${apiKey}`" + `;
        }
    }
    
    setApiKey(apiKey: string): void {
        this.apiKey = apiKey;
        this.defaultHeaders['Authorization'] = ` + "`Bearer ${apiKey}`" + `;
    }

{{range .Endpoints}}
    async {{.FunctionName}}({{.ParametersSignature}}): Promise<any> {
        {{.URLConstruction}}
        
        const options: RequestInit = {
            method: '{{.Method}}',
            headers: this.defaultHeaders{{.RequestBody}}
        };
        
        const response = await fetch(url, options);
        
        if (!response.ok) {
            throw new Error(` + "`API Error: ${response.status} ${response.statusText}`" + `);
        }
        
        return await response.json();
    }
{{end}}
}

export class APIError extends Error {
    public statusCode?: number;
    
    constructor(message: string, statusCode?: number) {
        super(message);
        this.statusCode = statusCode;
        this.name = 'APIError';
    }
}
`

	data := t.prepareTemplateData(spec, config)
	return t.executeTemplate(tmpl, data, filepath.Join(outputDir, "client.ts"))
}

func (t *TypeScriptSDKGenerator) prepareTemplateData(spec *OpenAPISpec, config *ClientSDKConfig) map[string]interface{} {
	// Use cases.Title instead of deprecated strings.Title
	caser := cases.Title(language.English)
	className := caser.String(config.PackageName) + "Client"
	endpoints := make([]map[string]interface{}, 0)

	for path, pathItem := range spec.Paths {
		for method, operation := range pathItem {
			endpoint := map[string]interface{}{
				"FunctionName":        t.generateFunctionName(method, path),
				"Method":              strings.ToUpper(method),
				"ParametersSignature": t.generateParametersSignature(operation.Parameters),
				"URLConstruction":     t.generateURLConstruction(path, operation.Parameters),
				"RequestBody":         t.generateRequestBody(operation.RequestBody),
			}
			endpoints = append(endpoints, endpoint)
		}
	}

	return map[string]interface{}{
		"PackageName": config.PackageName,
		"ClassName":   className,
		"ServiceName": spec.Info.Title,
		"GeneratedAt": time.Now().Format("2006-01-02 15:04:05"),
		"Endpoints":   endpoints,
	}
}

func (t *TypeScriptSDKGenerator) generateFunctionName(method, path string) string {
	// Similar to JavaScript
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var name strings.Builder
	name.WriteString(strings.ToLower(method))

	// Use cases.Title instead of deprecated strings.Title
	caser := cases.Title(language.English)
	for _, part := range parts {
		if !strings.HasPrefix(part, "{") {
			name.WriteString(caser.String(part))
		}
	}

	return name.String()
}

func (t *TypeScriptSDKGenerator) generateParametersSignature(params []OpenAPIParameter) string {
	parts := make([]string, 0, len(params))
	for _, param := range params {
		tsType := t.convertTypeToTypeScript(param.Schema.Type)
		parts = append(parts, fmt.Sprintf("%s: %s", param.Name, tsType))
	}
	return strings.Join(parts, ", ")
}

func (t *TypeScriptSDKGenerator) generateURLConstruction(path string, params []OpenAPIParameter) string {
	// Similar to JavaScript
	code := fmt.Sprintf("let url = `${this.baseURL}%s`;", path)

	for _, param := range params {
		if param.In == "path" {
			code = strings.ReplaceAll(code, "{"+param.Name+"}", "${"+param.Name+"}")
		}
	}

	queryParams := make([]string, 0)
	for _, param := range params {
		if param.In == "query" {
			queryParams = append(queryParams, param.Name)
		}
	}

	if len(queryParams) > 0 {
		code += "\n        const params = new URLSearchParams();\n"
		for _, param := range queryParams {
			code += fmt.Sprintf("        params.append('%s', %s.toString());\n", param, param)
		}
		code += "        url += '?' + params.toString();"
	}

	return code
}

func (t *TypeScriptSDKGenerator) generateRequestBody(body *OpenAPIRequestBody) string {
	if body == nil {
		return ""
	}
	return ",\n            body: JSON.stringify(requestBody)"
}

func (t *TypeScriptSDKGenerator) convertTypeToTypeScript(openAPIType string) string {
	switch openAPIType {
	case "string":
		return "string"
	case "integer":
		return "number"
	case "number":
		return "number"
	case "boolean":
		return "boolean"
	case "array":
		return "any[]"
	default:
		return "any"
	}
}

func (t *TypeScriptSDKGenerator) executeTemplate(tmplStr string, data interface{}, outputPath string) error {
	tmpl, err := template.New("client").Parse(tmplStr)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

// GenerateClientSDKs generates client SDKs for multiple languages
func GenerateClientSDKs(config *ClientSDKConfig) error {
	if !config.Enabled {
		return nil
	}

	// Generate spec OpenAPI
	defaultConfig := DefaultConfig()
	spec := GenerateOpenAPISpec(defaultConfig)

	// Create manager and generate SDKs
	manager := NewSDKManager(config)
	return manager.GenerateSDKs(spec)
}
