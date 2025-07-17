// Package decorators fornece um framework baseado em anotações para Gin
package decorators

// Re-exportar as principais funções e tipos para facilitar o uso
import "github.com/RodolfoBonis/deco/pkg/decorators"

// Re-exportar funções principais
var (
	// Funções de registro
	RegisterRoute         = decorators.RegisterRoute
	RegisterRouteWithMeta = decorators.RegisterRouteWithMeta
	RegisterGroup         = decorators.RegisterGroup
	Default               = decorators.Default
	GetRoutes             = decorators.GetRoutes
	GetGroups             = decorators.GetGroups

	// Funções de markers
	RegisterMarker = decorators.RegisterMarker
	GetMarkers     = decorators.GetMarkers

	// Hooks
	RegisterParserHook    = decorators.RegisterParserHook
	RegisterGeneratorHook = decorators.RegisterGeneratorHook

	// Funções de middleware
	CreateAuthMiddleware           = decorators.CreateAuthMiddleware
	CreateCacheMiddleware          = decorators.CreateCacheMiddleware
	CreateRateLimitMiddleware      = decorators.CreateRateLimitMiddleware
	CreateMetricsMiddleware        = decorators.CreateMetricsMiddleware
	CreateCORSMiddleware           = decorators.CreateCORSMiddleware
	CreateWebSocketMiddleware      = decorators.CreateWebSocketMiddleware
	CreateWebSocketStatsMiddleware = decorators.CreateWebSocketStatsMiddleware

	// WebSocket functions
	RegisterWebSocketHandler         = decorators.RegisterWebSocketHandler
	RegisterDefaultWebSocketHandlers = decorators.RegisterDefaultWebSocketHandlers
	GetWebSocketHub                  = decorators.GetWebSocketHub
	WebSocketHandlerWrapper          = decorators.WebSocketHandlerWrapper
)

// Re-exportar tipos principais
type (
	// RouteEntry representa uma rota registrada
	RouteEntry = decorators.RouteEntry

	// RouteMeta contém metadata extraída dos comentários
	RouteMeta = decorators.RouteMeta

	// MarkerConfig configuration of a marker personalizado
	MarkerConfig = decorators.MarkerConfig

	// MarkerInstance instância de um marker encontrado
	MarkerInstance = decorators.MarkerInstance

	// GenData dados para templates de geração
	GenData = decorators.GenData

	// MiddlewareInfo information about middlewares
	MiddlewareInfo = decorators.MiddlewareInfo

	// ParameterInfo informações de parâmetros
	ParameterInfo = decorators.ParameterInfo

	// GroupInfo informações de grupos
	GroupInfo = decorators.GroupInfo

	// Hooks
	// ParserHook is an alias for decorators.ParserHook. Represents a hook for custom parsing logic.
	ParserHook = decorators.ParserHook
	// GeneratorHook represents a hook for custom generation logic
	GeneratorHook = decorators.GeneratorHook
)
