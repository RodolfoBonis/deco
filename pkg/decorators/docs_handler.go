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
	totalProxies := 0

	for i := range routes {
		route := &routes[i]
		methodsMap[route.Method] = true
		totalMiddlewares += len(route.MiddlewareInfo)
		for _, mw := range route.MiddlewareInfo {
			uniqueMiddlewares[mw.Name] = true
			// Count proxy middlewares
			if mw.Name == "Proxy" {
				totalProxies++
			}
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
		TotalProxies      int
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
		TotalProxies:      totalProxies,
	}

	htmlTemplate := `
<!DOCTYPE html>
<html lang="pt-BR">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>gin-decorators - Route Documentation</title>
    <style>
        :root {
            --mascot-blue: #40B0C0;
            --mascot-cream: #F5E5C0;
            --mascot-green: #66CC33;
            --mascot-brown: #A0522D;
            --dark-bg: #1a1a1a;
            --dark-surface: #2d2d2d;
            --dark-surface-hover: #3a3a3a;
            --dark-border: #404040;
            --text-primary: #ffffff;
            --text-secondary: #b0b0b0;
            --text-muted: #808080;
        }

        * {
            box-sizing: border-box;
        }

        html, body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            margin: 0;
            padding: 0;
            background: var(--dark-bg);
            color: var(--text-primary);
            line-height: 1.6;
            height: auto;
            min-height: 100vh;
            overflow: auto;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            background: var(--dark-surface);
            border-radius: 16px;
            box-shadow: 0 8px 32px rgba(0,0,0,0.3);
            margin-top: 20px;
            margin-bottom: 20px;
            height: auto;
            min-height: auto;
            overflow: auto;
        }

        .header {
            background: linear-gradient(135deg, var(--mascot-blue) 0%, #2a8a9a 100%);
            color: white;
            padding: 40px 30px;
            text-align: center;
            position: relative;
            overflow: hidden;
        }

        .header::before {
            content: '';
            position: absolute;
            top: -50%;
            left: -50%;
            width: 200%;
            height: 200%;
            background: radial-gradient(circle, rgba(102, 204, 51, 0.1) 0%, transparent 70%);
            animation: float 6s ease-in-out infinite;
        }

        @keyframes float {
            0%, 100% { transform: translateY(0px) rotate(0deg); }
            50% { transform: translateY(-20px) rotate(180deg); }
        }

        .header h1 {
            margin: 0;
            font-size: 3rem;
            font-weight: 700;
            text-shadow: 0 2px 4px rgba(0,0,0,0.3);
            position: relative;
            z-index: 1;
        }

        .header p {
            margin: 10px 0 0 0;
            opacity: 0.95;
            font-size: 1.1rem;
            position: relative;
            z-index: 1;
        }

        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 20px;
            background: var(--dark-surface);
            padding: 30px;
            border-bottom: 1px solid var(--dark-border);
        }

        .stat {
            text-align: center;
            background: var(--dark-surface-hover);
            padding: 20px;
            border-radius: 12px;
            border: 1px solid var(--dark-border);
            transition: all 0.3s ease;
            position: relative;
            overflow: hidden;
        }

        .stat::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(64, 176, 192, 0.1), transparent);
            transition: left 0.5s ease;
        }

        .stat:hover::before {
            left: 100%;
        }

        .stat:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(64, 176, 192, 0.2);
            border-color: var(--mascot-blue);
        }

        .stat-number {
            font-size: 2.5rem;
            font-weight: 700;
            color: var(--mascot-blue);
            margin-bottom: 8px;
            text-shadow: 0 2px 4px rgba(0,0,0,0.3);
        }

        .stat-label {
            color: var(--text-secondary);
            font-size: 0.9rem;
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .view-controls {
            padding: 25px 30px;
            background: var(--dark-surface-hover);
            border-bottom: 1px solid var(--dark-border);
            display: flex;
            gap: 12px;
            flex-wrap: wrap;
            align-items: center;
        }

        .view-toggle {
            background: var(--dark-surface);
            color: var(--text-primary);
            border: 1px solid var(--dark-border);
            padding: 10px 18px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 0.9rem;
            font-weight: 500;
            transition: all 0.3s ease;
            position: relative;
            overflow: hidden;
        }

        .view-toggle::before {
            content: '';
            position: absolute;
            top: 0;
            left: -100%;
            width: 100%;
            height: 100%;
            background: linear-gradient(90deg, transparent, rgba(102, 204, 51, 0.2), transparent);
            transition: left 0.3s ease;
        }

        .view-toggle:hover::before {
            left: 100%;
        }

        .view-toggle:hover {
            background: var(--mascot-blue);
            color: white;
            border-color: var(--mascot-blue);
            transform: translateY(-1px);
        }

        .view-toggle.active {
            background: var(--mascot-green);
            color: white;
            border-color: var(--mascot-green);
            box-shadow: 0 4px 12px rgba(102, 204, 51, 0.3);
        }

        .collapse-section {
            margin: 20px 30px;
            border: 1px solid var(--dark-border);
            border-radius: 12px;
            overflow: hidden;
            background: var(--dark-surface);
            transition: all 0.3s ease;
        }

        .collapse-section:hover {
            border-color: var(--mascot-blue);
            box-shadow: 0 4px 20px rgba(64, 176, 192, 0.1);
        }

        .collapse-header {
            background: var(--dark-surface-hover);
            padding: 20px 25px;
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: space-between;
            border-bottom: 1px solid var(--dark-border);
            transition: all 0.3s ease;
        }

        .collapse-header:hover {
            background: var(--dark-surface);
        }

        .collapse-header h3 {
            margin: 0;
            color: var(--text-primary);
            font-size: 1.2rem;
            font-weight: 600;
        }

        .collapse-icon {
            font-size: 1.2rem;
            transition: transform 0.3s ease;
            color: var(--mascot-blue);
        }

        .collapse-icon.collapsed {
            transform: rotate(-90deg);
        }

        .collapse-content {
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.4s cubic-bezier(0.4, 0, 0.2, 1);
        }

        .collapse-content.expanded {
            max-height: fit-content;
        }

        .collapse-routes {
            padding: 0;
        }

        .route {
            border-bottom: 1px solid var(--dark-border);
            padding: 25px 30px;
            transition: all 0.3s ease;
            background: var(--dark-surface);
        }

        .route:hover {
            background: var(--dark-surface-hover);
            transform: translateX(4px);
        }

        .route:last-child {
            border-bottom: none;
        }

        .route-header {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
            flex-wrap: wrap;
            gap: 12px;
        }

        .method {
            padding: 8px 16px;
            border-radius: 8px;
            font-weight: 700;
            font-size: 0.85rem;
            min-width: 80px;
            text-align: center;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.2);
        }

        .method-GET { background: linear-gradient(135deg, #4CAF50, #45a049); color: white; }
        .method-POST { background: linear-gradient(135deg, #2196F3, #1976D2); color: white; }
        .method-PUT { background: linear-gradient(135deg, #FF9800, #F57C00); color: white; }
        .method-DELETE { background: linear-gradient(135deg, #F44336, #D32F2F); color: white; }
        .method-PATCH { background: linear-gradient(135deg, #9C27B0, #7B1FA2); color: white; }
        .method-WS { background: linear-gradient(135deg, var(--mascot-blue), #2a8a9a); color: white; }

        .path {
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            font-size: 1.1rem;
            font-weight: 600;
            color: var(--mascot-cream);
            flex: 1;
            background: var(--dark-bg);
            padding: 8px 12px;
            border-radius: 6px;
            border: 1px solid var(--dark-border);
        }

        .handler {
            color: var(--text-muted);
            font-size: 0.9rem;
            font-style: italic;
            background: var(--dark-surface-hover);
            padding: 6px 10px;
            border-radius: 6px;
        }

        .route-tags {
            margin-bottom: 15px;
        }

        .tag {
            display: inline-block;
            background: linear-gradient(135deg, var(--mascot-blue), #2a8a9a);
            color: white;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 0.75rem;
            margin-right: 8px;
            margin-bottom: 6px;
            font-weight: 500;
            box-shadow: 0 2px 6px rgba(64, 176, 192, 0.3);
        }

        .middlewares {
            margin: 15px 0;
            display: flex;
            flex-wrap: wrap;
            gap: 12px;
            position: relative;
            z-index: 1;
        }

        .middleware {
            background: var(--dark-surface-hover);
            color: var(--text-secondary);
            padding: 16px;
            border-radius: 12px;
            font-size: 0.85rem;
            position: relative;
            border: 1px solid var(--dark-border);
            transition: all 0.3s ease;
            min-width: 250px;
            flex: 1;
            max-width: 400px;
            height: auto;
            min-height: auto;
            overflow: visible;
        }

        .middleware:hover {
            background: var(--dark-surface);
            border-color: var(--mascot-blue);
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(64, 176, 192, 0.15);
        }

        .middleware-header {
            display: flex;
            align-items: center;
            margin-bottom: 12px;
            font-weight: 600;
            font-size: 0.9rem;
            color: var(--text-primary);
        }

        .middleware-icon {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            margin-right: 8px;
            flex-shrink: 0;
        }

        .middleware-auth .middleware-icon { background: #FF9800; }
        .middleware-cache .middleware-icon { background: #4CAF50; }
        .middleware-ratelimit .middleware-icon { background: #F44336; }
        .middleware-metrics .middleware-icon { background: #2196F3; }
        .middleware-cors .middleware-icon { background: #9C27B0; }
        .middleware-proxy .middleware-icon { background: var(--mascot-green); }

        .middleware-args {
            display: flex;
            flex-direction: column;
            gap: 6px;
        }

        .arg-item {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            padding: 8px 10px;
            background: var(--dark-bg);
            border-radius: 6px;
            border: 1px solid var(--dark-border);
            font-size: 0.75rem;
            gap: 8px;
        }

        .arg-key {
            color: var(--mascot-cream);
            font-weight: 500;
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
        }

        .arg-value {
            color: var(--mascot-blue);
            font-weight: 600;
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            background: rgba(64, 176, 192, 0.1);
            padding: 4px 8px;
            border-radius: 4px;
            border: 1px solid rgba(64, 176, 192, 0.2);
            word-break: break-all;
            max-width: 200px;
            overflow-wrap: break-word;
            white-space: normal;
            display: inline-block;
        }

        .arg-separator {
            color: var(--text-muted);
            margin: 0 4px;
        }

        .description {
            color: var(--text-secondary);
            font-size: 0.9rem;
            margin-top: 12px;
            font-style: italic;
            background: var(--dark-surface-hover);
            padding: 12px;
            border-radius: 8px;
            border-left: 4px solid var(--mascot-cream);
        }

        .empty-state {
            text-align: center;
            padding: 80px;
            color: var(--text-muted);
        }

        .empty-state h3 {
            color: var(--text-secondary);
            margin-bottom: 10px;
        }

        .json-link {
            position: fixed;
            bottom: 80px;
            right: 80px;
            background: linear-gradient(135deg, var(--mascot-blue), #2a8a9a);
            color: white;
            padding: 15px 25px;
            border-radius: 30px;
            text-decoration: none;
            box-shadow: 0 8px 25px rgba(64, 176, 192, 0.4);
            transition: all 0.3s ease;
            font-weight: 600;
            z-index: 1000;
            pointer-events: auto;
        }

        .json-link:hover {
            background: linear-gradient(135deg, var(--mascot-green), #5bbf2a);
            transform: translateY(-2px);
            box-shadow: 0 12px 35px rgba(102, 204, 51, 0.4);
            text-decoration: none;
            color: white;
        }

        .routes {
            padding: 0;
        }

        /* Responsive design */
        @media (max-width: 768px) {
            .container {
                margin: 10px;
                border-radius: 12px;
            }
            
            .header h1 {
                font-size: 2rem;
            }
            
            .stats {
                grid-template-columns: repeat(2, 1fr);
                gap: 15px;
                padding: 20px;
            }
            
            .stat-number {
                font-size: 2rem;
            }
            
            .view-controls {
                flex-direction: column;
                align-items: stretch;
            }
            
            .route-header {
                flex-direction: column;
                align-items: flex-start;
            }
            
            .path {
                width: 100%;
            }
        }

        /* Scrollbar styling */
        ::-webkit-scrollbar {
            width: 8px;
        }

        ::-webkit-scrollbar-track {
            background: var(--dark-bg);
        }

        ::-webkit-scrollbar-thumb {
            background: var(--mascot-blue);
            border-radius: 4px;
        }

        ::-webkit-scrollbar-thumb:hover {
            background: var(--mascot-green);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎨 gin-decorators</h1>
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
            <div class="stat">
                <div class="stat-number">{{.TotalProxies}}</div>
                <div class="stat-label">Proxies processed</div>
            </div>
        </div>
        
        <div class="view-controls">
            <button class="view-toggle active" onclick="switchView('tags')">🏷️ Por Tags</button>
            <button class="view-toggle" onclick="switchView('groups')">📁 Por Grupos</button>
            <button class="view-toggle" onclick="switchView('all')">📄 Todas as Rotas</button>
            <div style="margin-left: auto;">
                <button class="view-toggle" onclick="expandAll()" style="background: var(--mascot-green);">🔽 Expandir Tudo</button>
                <button class="view-toggle" onclick="collapseAll()" style="background: var(--mascot-brown);">🔼 Colapsar Tudo</button>
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
                                    <div class="middleware middleware-{{.Name | lower}}">
                                        <div class="middleware-header">
                                            <div class="middleware-icon"></div>
                                            <strong>{{.Name}}</strong>
                                        </div>
                                        {{if .Args}}
                                        <div class="middleware-args">
                                            {{range $key, $value := .Args}}
                                            <div class="arg-item">
                                                <span class="arg-key">{{$key}}</span>
                                                <span class="arg-separator">:</span>
                                                <span class="arg-value">{{$value}}</span>
                                            </div>
                                            {{end}}
                                        </div>
                                        {{end}}
                                    </div>
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
                                <div class="middleware middleware-{{.Name | lower}}">
                                    <div class="middleware-header">
                                        <div class="middleware-icon"></div>
                                        <strong>{{.Name}}</strong>
                                    </div>
                                    {{if .Args}}
                                    <div class="middleware-args">
                                        {{range $key, $value := .Args}}
                                        <div class="arg-item">
                                            <span class="arg-key">{{$key}}</span>
                                            <span class="arg-separator">:</span>
                                            <span class="arg-value">{{$value}}</span>
                                        </div>
                                        {{end}}
                                    </div>
                                    {{end}}
                                </div>
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
                                    <div class="middleware middleware-{{.Name | lower}}">
                                        <div class="middleware-header">
                                            <div class="middleware-icon"></div>
                                            <strong>{{.Name}}</strong>
                                        </div>
                                        {{if .Args}}
                                        <div class="middleware-args">
                                            {{range $key, $value := .Args}}
                                            <div class="arg-item">
                                                <span class="arg-key">{{$key}}</span>
                                                <span class="arg-separator">:</span>
                                                <span class="arg-value">{{$value}}</span>
                                            </div>
                                            {{end}}
                                        </div>
                                        {{end}}
                                    </div>
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
                                <div class="middleware middleware-{{.Name | lower}}">
                                    <div class="middleware-header">
                                        <div class="middleware-icon"></div>
                                        <strong>{{.Name}}</strong>
                                    </div>
                                    {{if .Args}}
                                    <div class="middleware-args">
                                        {{range $key, $value := .Args}}
                                        <div class="arg-item">
                                            <span class="arg-key">{{$key}}</span>
                                            <span class="arg-separator">:</span>
                                            <span class="arg-value">{{$value}}</span>
                                        </div>
                                        {{end}}
                                    </div>
                                    {{end}}
                                </div>
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
                            <div class="middleware middleware-{{.Name | lower}}">
                                <div class="middleware-header">
                                    <div class="middleware-icon"></div>
                                    <strong>{{.Name}}</strong>
                                </div>
                                {{if .Args}}
                                <div class="middleware-args">
                                    {{range $key, $value := .Args}}
                                    <div class="arg-item">
                                        <span class="arg-key">{{$key}}</span>
                                        <span class="arg-separator">:</span>
                                        <span class="arg-value">{{$value}}</span>
                                    </div>
                                    {{end}}
                                </div>
                                {{end}}
                            </div>
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
