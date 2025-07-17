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

	// Calculate statistics
	methodsMap := make(map[string]bool)
	totalMiddlewares := 0
	uniqueMiddlewares := make(map[string]bool)
	totalWebSockets := 0

	for i := range routes {
		route := &routes[i]
		methodsMap[route.Method] = true
		totalMiddlewares += len(route.MiddlewareInfo)
		for _, mw := range route.MiddlewareInfo {
			uniqueMiddlewares[mw.Name] = true
		}
		// Count WebSocket handlers
		totalWebSockets += len(route.WebSocketHandlers)
	}

	// Organize routes by tags and groups
	routesByTag := make(map[string][]RouteEntry)
	routesByGroup := make(map[string][]RouteEntry)
	untaggedRoutes := []RouteEntry{}
	ungroupedRoutes := []RouteEntry{}

	for i := range routes {
		route := &routes[i]
		// Group by tags
		if len(route.Tags) > 0 {
			for _, tag := range route.Tags {
				routesByTag[tag] = append(routesByTag[tag], *route)
			}
		} else {
			untaggedRoutes = append(untaggedRoutes, *route)
		}

		// Group by groups
		if route.Group != nil {
			routesByGroup[route.Group.Name] = append(routesByGroup[route.Group.Name], *route)
		} else {
			ungroupedRoutes = append(ungroupedRoutes, *route)
		}
	}

	// Create data structure for template
	data := struct {
		Routes            []RouteEntry
		RoutesByTag       map[string][]RouteEntry
		RoutesByGroup     map[string][]RouteEntry
		UntaggedRoutes    []RouteEntry
		UngroupedRoutes   []RouteEntry
		Groups            map[string]*GroupInfo
		TotalRoutes       int
		UniqueMethods     int
		TotalMiddlewares  int
		UniqueMiddlewares int
		TotalWebSockets   int
	}{
		Routes:            routes,
		RoutesByTag:       routesByTag,
		RoutesByGroup:     routesByGroup,
		UntaggedRoutes:    untaggedRoutes,
		UngroupedRoutes:   ungroupedRoutes,
		Groups:            groups,
		TotalRoutes:       len(routes),
		UniqueMethods:     len(methodsMap),
		TotalMiddlewares:  totalMiddlewares,
		UniqueMiddlewares: len(uniqueMiddlewares),
		TotalWebSockets:   totalWebSockets,
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
        .method-WS { background: #e6f3ff; color: #0066cc; }
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
        
        /* Collapse styles */
        .collapse-section {
            margin: 20px 0;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            overflow: hidden;
        }
        .collapse-header {
            background: #f8f9fa;
            padding: 15px 20px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: space-between;
            border-bottom: 1px solid #e9ecef;
            transition: background-color 0.2s;
        }
        .collapse-header:hover {
            background: #e9ecef;
        }
        .collapse-header h3 {
            margin: 0;
            color: #495057;
            font-size: 1.1rem;
        }
        .collapse-icon {
            font-size: 1.2rem;
            transition: transform 0.3s;
        }
        .collapse-icon.collapsed {
            transform: rotate(-90deg);
        }
        .collapse-content {
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.3s ease-out;
        }
        .collapse-content.expanded {
            max-height: 2000px;
        }
        .collapse-routes {
            padding: 0;
        }
        .collapse-routes .route {
            border-bottom: 1px solid #e9ecef;
            margin: 0;
        }
        .collapse-routes .route:last-child {
            border-bottom: none;
        }
        .view-toggle {
            background: #007bff;
            color: white;
            border: none;
            padding: 8px 16px;
            border-radius: 4px;
            cursor: pointer;
            margin: 10px 0;
            font-size: 0.9rem;
        }
        .view-toggle:hover {
            background: #0056b3;
        }
        .view-toggle.active {
            background: #28a745;
        }
        .view-controls {
            padding: 20px 30px;
            background: #f8f9fa;
            border-bottom: 1px solid #e9ecef;
            display: flex;
            gap: 10px;
            flex-wrap: wrap;
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
            <div class="stat">
                <div class="stat-number">{{.TotalWebSockets}}</div>
                <div class="stat-label">WebSocket handlers</div>
            </div>
        </div>
        
        <div class="view-controls">
            <button class="view-toggle active" onclick="switchView('tags')">📋 Por Tags</button>
            <button class="view-toggle" onclick="switchView('groups')">📁 Por Grupos</button>
            <button class="view-toggle" onclick="switchView('all')">📄 Todas as Rotas</button>
            <div style="margin-left: auto;">
                <button class="view-toggle" onclick="expandAll()" style="background: #28a745;">🔽 Expandir Tudo</button>
                <button class="view-toggle" onclick="collapseAll()" style="background: #6c757d;">🔼 Colapsar Tudo</button>
            </div>
        </div>
        
        <!-- Routes by Tags -->
        <div id="tags-view" class="view-content">
            {{if .RoutesByTag}}
                {{range $tag, $routes := .RoutesByTag}}
                <div class="collapse-section">
                    <div class="collapse-header" onclick="toggleCollapse('tag-{{$tag}}')">
                        <h3>🏷️ {{$tag}} ({{len $routes}} rotas)</h3>
                        <span class="collapse-icon" id="icon-tag-{{$tag}}">▼</span>
                    </div>
                    <div class="collapse-content" id="content-tag-{{$tag}}">
                        <div class="collapse-routes">
                            {{range $routes}}
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
                        </div>
                    </div>
                </div>
                {{end}}
            {{end}}
            
            {{if .UntaggedRoutes}}
            <div class="collapse-section">
                <div class="collapse-header" onclick="toggleCollapse('untagged')">
                    <h3>🏷️ Sem Tags ({{len .UntaggedRoutes}} rotas)</h3>
                    <span class="collapse-icon" id="icon-untagged">▼</span>
                </div>
                <div class="collapse-content" id="content-untagged">
                    <div class="collapse-routes">
                        {{range .UntaggedRoutes}}
                        <div class="route">
                            <div class="route-header">
                                <span class="method method-{{.Method}}">{{.Method}}</span>
                                <span class="path">{{.Path}}</span>
                                <span class="handler">{{.FuncName}}</span>
                            </div>
                            
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
                    </div>
                </div>
            </div>
            {{end}}
        </div>
        
        <!-- Routes by Groups -->
        <div id="groups-view" class="view-content" style="display: none;">
            {{if .RoutesByGroup}}
                {{range $group, $routes := .RoutesByGroup}}
                <div class="collapse-section">
                    <div class="collapse-header" onclick="toggleCollapse('group-{{$group}}')">
                        <h3>📁 {{$group}} ({{len $routes}} rotas)</h3>
                        <span class="collapse-icon" id="icon-group-{{$group}}">▼</span>
                    </div>
                    <div class="collapse-content" id="content-group-{{$group}}">
                        <div class="collapse-routes">
                            {{range $routes}}
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
                        </div>
                    </div>
                </div>
                {{end}}
            {{end}}
            
            {{if .UngroupedRoutes}}
            <div class="collapse-section">
                <div class="collapse-header" onclick="toggleCollapse('ungrouped')">
                    <h3>📁 Sem Grupo ({{len .UngroupedRoutes}} rotas)</h3>
                    <span class="collapse-icon" id="icon-ungrouped">▼</span>
                </div>
                <div class="collapse-content" id="content-ungrouped">
                    <div class="collapse-routes">
                        {{range .UngroupedRoutes}}
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
                    </div>
                </div>
            </div>
            {{end}}
        </div>
        
        <!-- All Routes -->
        <div id="all-view" class="view-content" style="display: none;">
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
    </div>
    
    <a href="/decorators/docs.json" class="json-link">📄 JSON</a>
    
    <script>
        // Toggle collapse functionality
        function toggleCollapse(id) {
            const content = document.getElementById('content-' + id);
            const icon = document.getElementById('icon-' + id);
            
            if (content.classList.contains('expanded')) {
                content.classList.remove('expanded');
                icon.classList.add('collapsed');
            } else {
                content.classList.add('expanded');
                icon.classList.remove('collapsed');
            }
        }
        
        // Switch between different views
        function switchView(view) {
            // Hide all views
            document.getElementById('tags-view').style.display = 'none';
            document.getElementById('groups-view').style.display = 'none';
            document.getElementById('all-view').style.display = 'none';
            
            // Show selected view
            document.getElementById(view + '-view').style.display = 'block';
            
            // Update button states
            document.querySelectorAll('.view-toggle').forEach(btn => {
                btn.classList.remove('active');
            });
            event.target.classList.add('active');
        }
        
        // Expand all collapses in current view
        function expandAll() {
            const currentView = document.querySelector('.view-content[style*="block"]') || document.getElementById('tags-view');
            const collapses = currentView.querySelectorAll('.collapse-content');
            const icons = currentView.querySelectorAll('.collapse-icon');
            
            collapses.forEach(collapse => {
                collapse.classList.add('expanded');
            });
            icons.forEach(icon => {
                icon.classList.remove('collapsed');
            });
        }
        
        // Collapse all in current view
        function collapseAll() {
            const currentView = document.querySelector('.view-content[style*="block"]') || document.getElementById('tags-view');
            const collapses = currentView.querySelectorAll('.collapse-content');
            const icons = currentView.querySelectorAll('.collapse-icon');
            
            collapses.forEach(collapse => {
                collapse.classList.remove('expanded');
            });
            icons.forEach(icon => {
                icon.classList.add('collapsed');
            });
        }
        
        // Initialize - expand first collapse in each view
        document.addEventListener('DOMContentLoaded', function() {
            // Expand first collapse in tags view
            const firstTagCollapse = document.querySelector('#tags-view .collapse-content');
            if (firstTagCollapse) {
                firstTagCollapse.classList.add('expanded');
                const firstIcon = document.querySelector('#tags-view .collapse-icon');
                if (firstIcon) {
                    firstIcon.classList.remove('collapsed');
                }
            }
        });
    </script>
</body>
</html>
`

	tmpl, err := template.New("docs").Funcs(template.FuncMap{
		"lower": strings.ToLower,
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
