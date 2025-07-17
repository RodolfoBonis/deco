package decorators

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// DocsData structure to pass data to documentation template
type DocsData struct {
	Routes           []RouteEntry
	TotalRoutes      int
	UniqueMethods    int
	TotalMiddlewares int
}

// DocsHandler serves the HTML documentation page
func DocsHandler(c *gin.Context) {
	routes := GetRoutes()
	groups := GetGroups()

	// Calcular statistics
	methodsMap := make(map[string]bool)
	totalMiddlewares := 0
	uniqueMiddlewares := make(map[string]bool)

	for _, route := range routes {
		methodsMap[route.Method] = true
		totalMiddlewares += len(route.MiddlewareInfo)
		for _, mw := range route.MiddlewareInfo {
			uniqueMiddlewares[mw.Name] = true
		}
	}

	// Create data structure for template
	data := struct {
		Routes            []RouteEntry
		Groups            map[string]*GroupInfo
		TotalRoutes       int
		UniqueMethods     int
		TotalMiddlewares  int
		UniqueMiddlewares int
	}{
		Routes:            routes,
		Groups:            groups,
		TotalRoutes:       len(routes),
		UniqueMethods:     len(methodsMap),
		TotalMiddlewares:  totalMiddlewares,
		UniqueMiddlewares: len(uniqueMiddlewares),
	}

	htmlTemplate := `
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>gin-decorators - Route Documentation</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            margin: 0;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        .header h1 {
            margin: 0;
            font-size: 2.5rem;
        }
        .header p {
            margin: 10px 0 0 0;
            opacity: 0.9;
        }
        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            background: #f8f9fa;
            padding: 20px;
            border-bottom: 1px solid #e9ecef;
        }
        .stat {
            text-align: center;
        }
        .stat-number {
            font-size: 2rem;
            font-weight: bold;
            color: #495057;
        }
        .stat-label {
            color: #6c757d;
            font-size: 0.9rem;
        }
        .groups-section {
            padding: 20px 30px;
            background: #fff3cd;
            border-bottom: 1px solid #e9ecef;
        }
        .groups-section h3 {
            margin: 0 0 15px 0;
            color: #856404;
        }
        .group-item {
            display: inline-block;
            background: #ffc107;
            color: #212529;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 0.85rem;
            margin-right: 10px;
            margin-bottom: 5px;
        }
        .routes {
            padding: 0;
        }
        .route {
            border-bottom: 1px solid #e9ecef;
            padding: 20px 30px;
            transition: background-color 0.2s;
        }
        .route:hover {
            background-color: #f8f9fa;
        }
        .route:last-child {
            border-bottom: none;
        }
        .route-header {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
            flex-wrap: wrap;
            gap: 10px;
        }
        .method {
            padding: 6px 16px;
            border-radius: 6px;
            font-weight: bold;
            font-size: 0.85rem;
            min-width: 70px;
            text-align: center;
        }
        .method-GET { background: #d4edda; color: #155724; }
        .method-POST { background: #cce5ff; color: #004085; }
        .method-PUT { background: #fff3cd; color: #856404; }
        .method-DELETE { background: #f8d7da; color: #721c24; }
        .method-PATCH { background: #e2e3e5; color: #383d41; }
        .path {
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 1.1rem;
            font-weight: 500;
            color: #495057;
            flex: 1;
        }
        .handler {
            color: #6c757d;
            font-size: 0.9rem;
            font-style: italic;
        }
        .route-tags {
            margin-bottom: 10px;
        }
        .tag {
            display: inline-block;
            background: #e7f3ff;
            color: #0066cc;
            padding: 2px 8px;
            border-radius: 10px;
            font-size: 0.75rem;
            margin-right: 6px;
        }
        .middlewares {
            margin: 10px 0;
        }
        .middleware {
            display: inline-block;
            background: #e9ecef;
            color: #495057;
            padding: 6px 12px;
            border-radius: 16px;
            font-size: 0.8rem;
            margin-right: 8px;
            margin-bottom: 6px;
            position: relative;
        }
        .middleware-auth { background: #fff3cd; color: #856404; }
        .middleware-cache { background: #d1ecf1; color: #0c5460; }
        .middleware-ratelimit { background: #f8d7da; color: #721c24; }
        .middleware-metrics { background: #d4edda; color: #155724; }
        .middleware-cors { background: #e2e3e5; color: #383d41; }
        .middleware-args {
            font-size: 0.7rem;
            opacity: 0.8;
            display: block;
            margin-top: 2px;
        }

        .description {
            color: #6c757d;
            font-size: 0.85rem;
            margin-top: 10px;
            font-style: italic;
        }
        .empty-state {
            text-align: center;
            padding: 60px;
            color: #6c757d;
        }
        .json-link {
            position: fixed;
            bottom: 20px;
            right: 20px;
            background: #007bff;
            color: white;
            padding: 12px 20px;
            border-radius: 25px;
            text-decoration: none;
            box-shadow: 0 4px 12px rgba(0,123,255,0.3);
            transition: background-color 0.2s;
        }
        .json-link:hover {
            background: #0056b3;
            text-decoration: none;
            color: white;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎯 gin-decorators</h1>
            <p>Automatic documentation of routes</p>
        </div>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-number">{{.TotalRoutes}}</div>
                <div class="stat-label">Routes registradas</div>
            </div>
            <div class="stat">
                <div class="stat-number">{{.UniqueMethods}}</div>
                <div class="stat-label">Unique methods</div>
            </div>
            <div class="stat">
                <div class="stat-number">{{.TotalMiddlewares}}</div>
                <div class="stat-label">Applied middlewares</div>
            </div>
            <div class="stat">
                <div class="stat-number">{{.UniqueMiddlewares}}</div>
                <div class="stat-label">Middleware types</div>
            </div>
        </div>
        
        {{if .Groups}}
        <div class="groups-section">
            <h3>📁 Route Groups</h3>
            {{range .Groups}}
            <div class="group-item">
                <strong>{{.Name}}</strong> → {{.Prefix}}
                {{if .Description}}<br><small>{{.Description}}</small>{{end}}
            </div>
            {{end}}
        </div>
        {{end}}
        
        <div class="routes">
            {{if .Routes}}
                {{range .Routes}}
                <div class="route">
                    <div class="route-header">
                        <span class="method method-{{.Method}}">{{.Method}}</span>
                        <span class="path">{{.Path}}</span>
                        <span class="handler">{{.FuncName}}</span>
                    </div>
                    
                    {{if .Tags}}
                    <div class="route-tags">
                        {{range .Tags}}
                        <span class="tag">{{.}}</span>
                        {{end}}
                    </div>
                    {{end}}
                    
                    {{if .Description}}
                    <div class="description">{{.Description}}</div>
                    {{end}}
                    
                    {{if .MiddlewareInfo}}
                    <div class="middlewares">
                        {{range .MiddlewareInfo}}
                        <span class="middleware middleware-{{.Name | lower}}">
                            <strong>{{.Name}}</strong>
                            {{if .Args}}
                            <span class="middleware-args">
                                {{range $key, $value := .Args}}{{$key}}: {{$value}} {{end}}
                            </span>
                            {{end}}
                        </span>
                        {{end}}
                    </div>
                    {{end}}
                    

                </div>
                {{end}}
            {{else}}
                <div class="empty-state">
                    <h3>No registered routes</h3>
                    <p>Add @Route annotations to your handlers to see them here.</p>
                </div>
            {{end}}
        </div>
    </div>
    
    <a href="/decorators/docs.json" class="json-link">📄 JSON</a>
</body>
</html>
`

	tmpl, err := template.New("docs").Funcs(template.FuncMap{
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
	}).Parse(htmlTemplate)
	if err != nil {
		c.JSON(500, gin.H{"error": "Error processing template"})
		return
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(c.Writer, data); err != nil {
		c.JSON(500, gin.H{"error": "Error rendering template"})
		return
	}
}

// DocsJSONHandler serves documentation in JSON/OpenAPI format
func DocsJSONHandler(c *gin.Context) {
	// Use default configuration if not provided
	config := DefaultConfig()
	spec := GenerateOpenAPISpec(config)
	c.JSON(http.StatusOK, spec)
}

// Removed - RouteInfo now in types.go
