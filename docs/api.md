package decorators // import "github.com/RodolfoBonis/deco/pkg/decorators"


VARIABLES

var (
	URLCacheKey = func(c *gin.Context) string {
		return fmt.Sprintf("cache:url:%s:%s", c.Request.Method, c.Request.URL.String())
	}

	UserURLCacheKey = func(c *gin.Context) string {
		userID := c.GetString("user_id")
		if userID == "" {
			userID = "anonymous"
		}
		return fmt.Sprintf("cache:user:%s:url:%s:%s", userID, c.Request.Method, c.Request.URL.String())
	}

	EndpointCacheKey = func(c *gin.Context) string {
		return fmt.Sprintf("cache:endpoint:%s:%s", c.Request.Method, c.FullPath())
	}
)
    Default cache key generation functions

var (
	IPKeyGenerator = func(c *gin.Context) string {
		return "ratelimit:ip:" + c.ClientIP()
	}

	UserKeyGenerator = func(c *gin.Context) string {
		userID := c.GetString("user_id")
		if userID == "" {
			return "ratelimit:anonymous:" + c.ClientIP()
		}
		return "ratelimit:user:" + userID
	}

	EndpointKeyGenerator = func(c *gin.Context) string {
		return fmt.Sprintf("ratelimit:endpoint:%s:%s:%s", c.Request.Method, c.FullPath(), c.ClientIP())
	}
)
    Default key generation functions

var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(_ *http.Request) bool {

		return true
	},
}
    WebSocketUpgrader configuration for connection upgrade WebSocket


FUNCTIONS

func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue)
    AddSpanAttributes adds attributes to current span

func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue)
    AddSpanEvent adds event to current span

func BroadcastHandler(conn *WebSocketConnection, message *WebSocketMessage) error
    BroadcastHandler handler for broadcast

func CacheByEndpoint(config *CacheConfig) gin.HandlerFunc
    CacheByEndpoint cache middleware by endpoint

func CacheByURL(config *CacheConfig) gin.HandlerFunc
    CacheByURL cache middleware by URL

func CacheByUserURL(config *CacheConfig) gin.HandlerFunc
    CacheByUserURL cache middleware by user and URL

func CacheMiddleware(config *CacheConfig, keyGen CacheKeyFunc) gin.HandlerFunc
    CacheMiddleware creates cache middleware

func CacheStatsHandler(store CacheStore) gin.HandlerFunc
    CacheStatsHandler handler for cache statistics

func ClearSchemas()
    ClearSchemas clears all registered schemas (useful for testing)

func CreateAuthMiddleware(args string) func(c *gin.Context)
    CreateAuthMiddleware creates auth middleware (wrapper for generation)

func CreateCORSMiddleware(args string) func(c *gin.Context)
    CreateCORSMiddleware creates CORS middleware (wrapper for generation)

func CreateCacheMiddleware(args string) func(c *gin.Context)
    CreateCacheMiddleware creates cache middleware (wrapper for generation)

func CreateMetricsMiddleware(args string) func(c *gin.Context)
    CreateMetricsMiddleware creates metrics middleware (wrapper for generation)

func CreateRateLimitMiddleware(args string) func(c *gin.Context)
    CreateRateLimitMiddleware creates rate limit middleware (wrapper for
    generation)

func CreateWebSocketHandler(config *WebSocketConfig) gin.HandlerFunc
    CreateWebSocketHandler creates handler for WebSocket connections

func CreateWebSocketMiddleware(args string) gin.HandlerFunc
    CreateWebSocketMiddleware creates WebSocket middleware (wrapper for
    generation)

func CreateWebSocketStatsMiddleware(args string) gin.HandlerFunc
    CreateWebSocketStatsMiddleware creates WebSocket stats middleware (wrapper
    for generation)

func CustomCache(ttl time.Duration, keyGen CacheKeyFunc, cacheType string) gin.HandlerFunc
    CustomCache customizable cache middleware

func CustomRateLimit(limit int, _ time.Duration, keyGen KeyGeneratorFunc, rateLimiterType string) gin.HandlerFunc
    CustomRateLimit customizable rate limiting middleware

func Default() *gin.Engine
    Default creates a gin.Engine with all registered routes

func DocsHandler(c *gin.Context)
    DocsHandler serves the HTML documentation page

func DocsJSONHandler(c *gin.Context)
    DocsJSONHandler serves documentation in JSON/OpenAPI format

func EchoHandler(conn *WebSocketConnection, message *WebSocketMessage) error
    EchoHandler echo handler for testing

func GenerateClientSDKs(config *ClientSDKConfig) error
    GenerateClientSDKs generates client SDKs for multiple languages

func GenerateFromTemplate(rootDir, templatePath, outputPath, pkgName string) error
    GenerateFromTemplate generates code using custom template

func GenerateInitFile(rootDir, outputPath, pkgName string) error
    GenerateInitFile generates the init_decorators.go file for production

func GenerateInitFileWithConfig(rootDir, outputPath, pkgName string, config *Config) error
    GenerateInitFileWithConfig generates file with specific configuration

func GetGroups() map[string]*GroupInfo
    GetGroups returns all registered groups

func GetMarkers() map[string]MarkerConfig
    GetMarkers returns all registered markers

func GetMinifiedTemplate() string
    GetMinifiedTemplate returns minified template for generation

func GetSchemas() map[string]*SchemaInfo
    GetSchemas returns all registered schemas

func GetValidatedData(c *gin.Context) (interface{}, bool)
    GetValidatedData extracts validated data from context

func GetValidatedQuery(c *gin.Context) (interface{}, bool)
    GetValidatedQuery extracts validated query from context

func GetWebSocketInfo(config WebSocketConfig) map[string]interface{}
    GetWebSocketInfo returns information about WebSocket

func HealthCheckHandler() gin.HandlerFunc
    HealthCheckHandler health check handler with metrics

func HealthCheckWithTracing() gin.HandlerFunc
    HealthCheckWithTracing instrumented health check

func InstrumentedHandler(handlerName string, handler gin.HandlerFunc) gin.HandlerFunc
    InstrumentedHandler wrapper to instrument custom handlers

func InvalidateCacheHandler(store CacheStore) gin.HandlerFunc
    InvalidateCacheHandler handler to invalidate cache

func JoinGroupHandler(conn *WebSocketConnection, message *WebSocketMessage) error
    JoinGroupHandler handler to join group

func LeaveGroupHandler(conn *WebSocketConnection, message *WebSocketMessage) error
    LeaveGroupHandler handler to leave group

func LogNormal(format string, args ...interface{})
    LogNormal imprime log em modo normal e verbose

func LogSilent(format string, args ...interface{})
    LogSilent always prints log (used for important errors)

func LogVerbose(format string, args ...interface{})
    LogVerbose imprime log apenas em modo verbose

func MetricsMiddleware(config *MetricsConfig) gin.HandlerFunc
    MetricsMiddleware main middleware for metrics collection

func MinifyCode(inputPath, outputPath string, enabled bool) error
    MinifyCode minifies Go code by removing comments and unnecessary spaces

func OpenAPIJSONHandler(config *Config) gin.HandlerFunc
    OpenAPIJSONHandler serves OpenAPI 3.0 documentation in JSON

func OpenAPIYAMLHandler(config *Config) gin.HandlerFunc
    OpenAPIYAMLHandler serves OpenAPI 3.0 documentation in YAML

func PrometheusHandler() gin.HandlerFunc
    PrometheusHandler returns Prometheus handler

func RateLimitByEndpoint(config *RateLimitConfig) gin.HandlerFunc
    RateLimitByEndpoint rate limiting middleware by endpoint

func RateLimitByIP(config *RateLimitConfig) gin.HandlerFunc
    RateLimitByIP rate limiting middleware by IP

func RateLimitByUser(config *RateLimitConfig) gin.HandlerFunc
    RateLimitByUser rate limiting middleware by user

func RateLimitMiddleware(config *RateLimitConfig, keyGen KeyGeneratorFunc) gin.HandlerFunc
    RateLimitMiddleware creates rate limiting middleware

func RecordCacheHit(cacheType, keyType string)
    RecordCacheHit registra hit de cache

func RecordCacheMiss(cacheType, keyType string)
    RecordCacheMiss registra miss de cache

func RecordCacheSize(cacheType string, size float64)
    RecordCacheSize registra tamanho do cache

func RecordMiddlewareError(middleware, errorType string)
    RecordMiddlewareError records middleware error

func RecordMiddlewareTime(middleware, endpoint string, duration time.Duration)
    RecordMiddlewareTime records middleware execution time

func RecordRateLimitExceeded(endpoint, limitType string)
    RecordRateLimitExceeded registra rate limit excedido

func RecordRateLimitHit(endpoint, limitType string)
    RecordRateLimitHit records rate limit check

func RecordValidationError(validationType, field string)
    RecordValidationError records validation error

func RecordValidationTime(validationType string, duration time.Duration)
    RecordValidationTime records validation time

func RegisterDefaultHandlers()
    RegisterDefaultHandlers registers default handlers

func RegisterDefaultWebSocketHandlers()
    RegisterDefaultWebSocketHandlers is a public alias for
    RegisterDefaultHandlers

func RegisterGeneratorHook(h GeneratorHook)
    RegisterGeneratorHook registers a generation hook

func RegisterMarker(config MarkerConfig)
    RegisterMarker registers a new marker in the framework

func RegisterParserHook(h ParserHook)
    RegisterParserHook registers a parsing hook

func RegisterRoute(method, path string, handlers ...gin.HandlerFunc)
    RegisterRoute registers a new route in the framework

func RegisterRouteWithMeta(entry *RouteEntry)
    RegisterRouteWithMeta registers a route with complete metadata

func RegisterSchema(schema *SchemaInfo)
    RegisterSchema registers a new schema in the framework

func RegisterWebSocketHandler(messageType string, handler WebSocketHandler)
    RegisterWebSocketHandler allows applications to register custom WebSocket
    handlers

func SaveConfig(config *Config, configPath string) error
    SaveConfig saves configuration to file

func SetLogLevel(level LogLevel)
    SetLogLevel defines logging level globally

func SetSpanError(ctx context.Context, err error)
    SetSpanError marca span como error

func SetVerbose(verbose bool)
    SetVerbose ativa/desativa logs verbose

func SpanFromContext(ctx context.Context) trace.Span
    SpanFromContext extracts span from context

func StartSpan(ctx context.Context, name string) (context.Context, trace.Span)
    StartSpan starts a new span

func SwaggerRedirectHandler(c *gin.Context)
    SwaggerRedirectHandler redirects to swagger UI (convenience endpoint)

func SwaggerUIHandler(_ *Config) gin.HandlerFunc
    SwaggerUIHandler serves Swagger UI interface for API documentation

func TraceCacheOperation(ctx context.Context, operation, cacheType, key string) (context.Context, trace.Span)
    TraceCacheOperation instruments cache operations

func TraceMiddleware(middlewareName string) gin.HandlerFunc
    TraceMiddleware instrumenta middleware individual

func TraceRateLimitOperation(ctx context.Context, operation, limitType string, allowed bool) (context.Context, trace.Span)
    TraceRateLimitOperation instruments rate limit operations

func TraceValidationOperation(ctx context.Context, validationType string, fieldCount int) (context.Context, trace.Span)
    TraceValidationOperation instruments validation operations

func TraceWebSocketOperation(ctx context.Context, operation, connectionID string) (context.Context, trace.Span)
    TraceWebSocketOperation instruments WebSocket operations

func TracingMiddleware(config *TelemetryConfig) gin.HandlerFunc
    TracingMiddleware main tracing middleware

func TracingStatsHandler() gin.HandlerFunc
    TracingStatsHandler handler for tracing statistics

func ValidateGeneration(generatedPath string) error
    ValidateGeneration validates if the generated file is correct

func ValidateJSON(target interface{}, config *ValidationConfig) gin.HandlerFunc
    ValidateJSON middleware for automatic JSON validation

func ValidateParams(rules map[string]string, _ *ValidationConfig) gin.HandlerFunc
    ValidateParams middleware for path parameter validation

func ValidateQuery(target interface{}, config *ValidationConfig) gin.HandlerFunc
    ValidateQuery middleware for query parameter validation

func ValidateStruct(config *ValidationConfig) gin.HandlerFunc
    ValidateStruct middleware for automatic struct validation

func WebSocketHandlerWrapper(_ WebSocketHandler) gin.HandlerFunc
    WebSocketHandlerWrapper converts WebSocketHandler to gin.HandlerFunc for
    documentation

func WebSocketStatsHandler() gin.HandlerFunc
    WebSocketStatsHandler handler for WebSocket statistics


TYPES

type CacheConfig struct {
	Type        string `yaml:"type"` // "memory", "redis"
	DefaultTTL  string `yaml:"default_ttl"`
	MaxSize     int    `yaml:"max_size,omitempty"`
	Compression bool   `yaml:"compression"`
}
    CacheConfig cache system configuration

type CacheEntry struct {
	Data      []byte            `json:"data"`
	Headers   map[string]string `json:"headers"`
	Status    int               `json:"status"`
	ExpiresAt time.Time         `json:"expires_at"`
}
    CacheEntry represents a cache entry

type CacheKeyFunc func(c *gin.Context) string
    CacheKeyFunc function to generate cache key

func ParseCacheArgs(args []string) (time.Duration, string, CacheKeyFunc)
    ParseCacheArgs parses @Cache decorator arguments

type CacheStats struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Sets      int64   `json:"sets"`
	Deletes   int64   `json:"deletes"`
	Evictions int64   `json:"evictions"`
	Size      int64   `json:"size"`
	MaxSize   int64   `json:"max_size"`
	HitRate   float64 `json:"hit_rate"`
}
    CacheStats cache statistics

type CacheStore interface {
	Get(ctx context.Context, key string) (*CacheEntry, error)
	Set(ctx context.Context, key string, entry *CacheEntry, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	Stats() CacheStats
}
    CacheStore interface for different cache implementations

type ClientSDKConfig struct {
	Enabled     bool     `yaml:"enabled"`
	OutputDir   string   `yaml:"output_dir"`
	Languages   []string `yaml:"languages"` // "go", "python", "javascript", "typescript"
	PackageName string   `yaml:"package_name"`
	ModuleName  string   `yaml:"module_name,omitempty"`
}
    ClientSDKConfig SDK generation configuration

type Config struct {
	Version    string           `yaml:"version"`
	Handlers   HandlersConfig   `yaml:"handlers"`
	Generate   GenerationConfig `yaml:"generation"`
	Dev        DevConfig        `yaml:"dev"`
	Prod       ProdConfig       `yaml:"prod"`
	Redis      RedisConfig      `yaml:"redis,omitempty"`
	Cache      CacheConfig      `yaml:"cache,omitempty"`
	RateLimit  RateLimitConfig  `yaml:"rate_limit,omitempty"`
	Metrics    MetricsConfig    `yaml:"metrics,omitempty"`
	OpenAPI    OpenAPIConfig    `yaml:"openapi,omitempty"`
	Validation ValidationConfig `yaml:"validation,omitempty"`
	WebSocket  WebSocketConfig  `yaml:"websocket,omitempty"`
	Telemetry  TelemetryConfig  `yaml:"telemetry,omitempty"`
	ClientSDK  ClientSDKConfig  `yaml:"client_sdk,omitempty"`
}
    Config framework configuration structure

func DefaultConfig() *Config
    DefaultConfig returns default configuration

func LoadConfig(configPath string) (*Config, error)
    LoadConfig loads configuration from file

func (c *Config) DiscoverHandlers(rootDir string) ([]string, error)
    DiscoverHandlers discovers handler files based on configuration

func (c *Config) Validate() error
    Validate validates the configuration

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}
    Contact contact information

type Debouncer struct {
	// Has unexported fields.
}
    Debouncer prevents too frequent regenerations

func NewDebouncer(duration time.Duration) *Debouncer
    NewDebouncer creates a new debouncer

func (d *Debouncer) Debounce(fn func())
    Debounce executes the function after a delay, canceling previous executions

type DevConfig struct {
	AutoDiscover bool `yaml:"auto_discover"`
	Watch        bool `yaml:"watch"`
}
    DevConfig configuration for development mode

type Discriminator struct {
	PropertyName string            `json:"propertyName"`
	Mapping      map[string]string `json:"mapping,omitempty"`
}
    Discriminator discriminator for polymorphism

type DocsData struct {
	Routes           []RouteEntry
	TotalRoutes      int
	UniqueMethods    int
	TotalMiddlewares int
}
    DocsData structure to pass data to documentation template

type Encoding struct {
	ContentType   string            `json:"contentType,omitempty"`
	Headers       map[string]Header `json:"headers,omitempty"`
	Style         string            `json:"style,omitempty"`
	Explode       bool              `json:"explode,omitempty"`
	AllowReserved bool              `json:"allowReserved,omitempty"`
}
    Encoding encoding

type EntityMeta struct {
	Name        string           `json:"name"`
	PackageName string           `json:"package_name"`
	FileName    string           `json:"file_name"`
	Markers     []MarkerInstance `json:"markers"`
	Fields      []FieldMeta      `json:"fields"`

	// Documentation information
	Description string                 `json:"description"`
	Example     map[string]interface{} `json:"example,omitempty"`
}
    EntityMeta represents metadata of an entity/struct extracted from comments

type Example struct {
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}
    Example exemplo

type ExternalDocs struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}
    ExternalDocs external documentation

type FieldMeta struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	JSONTag     string      `json:"json_tag"`
	Description string      `json:"description"`
	Example     interface{} `json:"example,omitempty"`
	Validation  string      `json:"validation,omitempty"` // from validate tags
}
    FieldMeta represents metadata of a struct field

type FileWatcher struct {
	// Has unexported fields.
}
    FileWatcher monitors handler files and automatically regenerates code

var GlobalWatcher *FileWatcher
    GlobalWatcher global file watcher instance

func NewFileWatcher(config *Config) (*FileWatcher, error)
    NewFileWatcher creates a new file watcher

func (fw *FileWatcher) IsRunning() bool
    IsRunning returns whether the watcher is running

func (fw *FileWatcher) Start() error
    Start starts file watching

func (fw *FileWatcher) Stop() error
    Stop file watching

type FrameworkStats struct {
	TotalRoutes       int            `json:"total_routes"`
	UniqueMiddlewares int            `json:"unique_middlewares"`
	PackagesScanned   int            `json:"packages_scanned"`
	BuildMode         string         `json:"build_mode"` // "development" ou "production"
	GeneratedAt       string         `json:"generated_at"`
	Methods           map[string]int `json:"methods"` // GET: 5, POST: 3, etc.
}
    FrameworkStats statistics do framework

type GenData struct {
	PackageName string                 // nome do pacote de destino
	Routes      []*RouteMeta           // routes to be generated
	Imports     []string               // necessary imports
	Metadata    map[string]interface{} // additional plugin data
	GeneratedAt string                 // generation timestamp
}
    GenData data passed to generation template

type GenerationConfig struct {
	Template string `yaml:"template,omitempty"`
}
    GenerationConfig configuration for code generation

type GeneratorHook func(data *GenData) error
    GeneratorHook executed before code generation

func GetGeneratorHooks() []GeneratorHook
    GetGeneratorHooks returns all generator hooks (for testing)

type GoSDKGenerator struct{}
    GoSDKGenerator generator for Go

func (g *GoSDKGenerator) Generate(spec *OpenAPISpec, config *ClientSDKConfig) error
    Generate creates a Go client SDK from the OpenAPI specification

func (g *GoSDKGenerator) GetFileExtension() string
    GetFileExtension retorna a extensão de arquivo para a linguagem.

func (g *GoSDKGenerator) GetLanguage() string
    GetLanguage retorna a linguagem de programação usada.

type GroupInfo struct {
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Description string `json:"description"`
}
    GroupInfo represents information of a route group

func GetGroup(name string) *GroupInfo
    GetGroup returns information of a group

func RegisterGroup(name, prefix, description string) *GroupInfo
    RegisterGroup registers a new route group

type HandlersConfig struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}
    HandlersConfig configuration for handlers discovery

type Header struct {
	Description     string               `json:"description,omitempty"`
	Required        bool                 `json:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty"`
	Schema          *OpenAPISchema       `json:"schema,omitempty"`
	Example         interface{}          `json:"example,omitempty"`
	Examples        map[string]Example   `json:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty"`
}
    Header header

type JavaScriptSDKGenerator struct{}
    JavaScriptSDKGenerator generator for JavaScript

func (j *JavaScriptSDKGenerator) Generate(spec *OpenAPISpec, config *ClientSDKConfig) error
    Generate creates a JavaScript client SDK from the OpenAPI specification

func (j *JavaScriptSDKGenerator) GetFileExtension() string
    GetFileExtension retorna a extensão de arquivo para a linguagem.

func (j *JavaScriptSDKGenerator) GetLanguage() string
    GetLanguage retorna a linguagem de programação usada.

type KeyGeneratorFunc func(c *gin.Context) string
    KeyGeneratorFunc function to generate rate limiting keys

func ParseRateLimitArgs(args []string) (limit int, window time.Duration, rateLimiterType string, keyGen KeyGeneratorFunc)
    ParseRateLimitArgs parses @RateLimit decorator arguments

type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}
    License license information

type Link struct {
	OperationRef string                 `json:"operationRef,omitempty"`
	OperationID  string                 `json:"operationId,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	RequestBody  interface{}            `json:"requestBody,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Server       *OpenAPIServer         `json:"server,omitempty"`
}
    Link link to other operations

type LogLevel int
    LogLevel defines logging level

const (
	// LogLevelSilent indicates that no logs should be produced
	LogLevelSilent LogLevel = iota
	// LogLevelNormal indicates normal logging level
	LogLevelNormal
	// LogLevelVerbose indicates verbose logging level
	LogLevelVerbose
)
func GetLogLevel() LogLevel
    GetLogLevel returns current logging level

type Logger struct {
	// Has unexported fields.
}
    Logger controla o logging do framework

type MarkerConfig struct {
	Name        string                              // Marker name (ex: "Auth")
	Pattern     *regexp.Regexp                      // Regex to detect the marker
	Factory     func(args []string) gin.HandlerFunc // Factory to create middleware
	Description string                              // Marker description
}
    MarkerConfig configuration of a marker

type MarkerFactory func(args []string) gin.HandlerFunc
    MarkerFactory function that creates a middleware based on arguments

type MarkerInstance struct {
	Name string   // Auth, Cache, etc.
	Args []string // parsed arguments
	Raw  string   // original comment text
}
    MarkerInstance represents a marker instance found

type MediaType struct {
	Schema   *OpenAPISchema      `json:"schema,omitempty"`
	Example  interface{}         `json:"example,omitempty"`
	Examples map[string]Example  `json:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty"`
}
    MediaType media type

type MemoryCache struct {
	// Has unexported fields.
}
    MemoryCache in-memory cache implementation

func NewMemoryCache(maxSize int) *MemoryCache
    NewMemoryCache creates a new in-memory cache

func (m *MemoryCache) Clear(_ context.Context) error
    Clear clears entire cache (in-memory implementation)

func (m *MemoryCache) Delete(_ context.Context, key string) error
    Delete removes cache entry (in-memory implementation)

func (m *MemoryCache) Get(_ context.Context, key string) (*CacheEntry, error)
    Get retrieves cache entry (in-memory implementation)

func (m *MemoryCache) Set(_ context.Context, key string, entry *CacheEntry, ttl time.Duration) error
    Set stores cache entry (in-memory implementation)

func (m *MemoryCache) Stats() CacheStats
    Stats returns cache statistics (in-memory implementation)

type MemoryRateLimiter struct {
	// Has unexported fields.
}
    MemoryRateLimiter local in-memory implementation

func NewMemoryRateLimiter() *MemoryRateLimiter
    NewMemoryRateLimiter creates an in-memory rate limiter

func (m *MemoryRateLimiter) Allow(_ context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, retryAfter time.Duration, err error)
    Allow checks if the request can proceed (in-memory implementation)

func (m *MemoryRateLimiter) Reset(_ context.Context, key string) error
    Reset clears the bucket for a key (in-memory implementation)

type MetricsCollector struct {
	// Has unexported fields.
}
    MetricsCollector collects custom metrics

func InitMetrics(config *MetricsConfig) *MetricsCollector
    InitMetrics initializes metrics system

type MetricsConfig struct {
	Enabled   bool      `yaml:"enabled"`
	Endpoint  string    `yaml:"endpoint"`
	Namespace string    `yaml:"namespace"`
	Subsystem string    `yaml:"subsystem"`
	Buckets   []float64 `yaml:"buckets,omitempty"`
}
    MetricsConfig Prometheus configuration

type MetricsInfo struct {
	Enabled   bool     `json:"enabled"`
	Endpoint  string   `json:"endpoint"`
	Namespace string   `json:"namespace"`
	Subsystem string   `json:"subsystem"`
	Metrics   []string `json:"metrics"`
}
    MetricsInfo information about available metrics

func GetMetricsInfo(config *MetricsConfig) MetricsInfo
    GetMetricsInfo returns information about metrics

type MiddlewareInfo struct {
	Name        string                 `json:"name"`
	Args        map[string]interface{} `json:"args"`
	Order       int                    `json:"order"`
	Description string                 `json:"description"`
}
    MiddlewareInfo information about middlewares aplicados

type MultipleValidationError struct {
	Errors []ValidationError
}
    MultipleValidationError represents multiple validation errors

func (e *MultipleValidationError) Error() string

type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}
    OAuthFlow fluxo OAuth2

type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty"`
}
    OAuthFlows fluxos OAuth2

type OpenAPIComponents struct {
	Schemas         map[string]*OpenAPISchema     `json:"schemas,omitempty"`
	Responses       map[string]OpenAPIResponse    `json:"responses,omitempty"`
	Parameters      map[string]OpenAPIParameter   `json:"parameters,omitempty"`
	Examples        map[string]Example            `json:"examples,omitempty"`
	RequestBodies   map[string]OpenAPIRequestBody `json:"requestBodies,omitempty"`
	Headers         map[string]Header             `json:"headers,omitempty"`
	SecuritySchemes map[string]SecurityScheme     `json:"securitySchemes,omitempty"`
	Links           map[string]Link               `json:"links,omitempty"`
	Callbacks       map[string]interface{}        `json:"callbacks,omitempty"`
}
    OpenAPIComponents reusable components

type OpenAPIConfig struct {
	Version     string                 `yaml:"version"`
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description"`
	Host        string                 `yaml:"host"`
	BasePath    string                 `yaml:"base_path"`
	Schemes     []string               `yaml:"schemes"`
	Contact     map[string]interface{} `yaml:"contact,omitempty"`
	License     map[string]interface{} `yaml:"license,omitempty"`
	Security    []map[string][]string  `yaml:"security,omitempty"`
}
    OpenAPIConfig OpenAPI documentation configuration

type OpenAPIInfo struct {
	Title          string   `json:"title"`
	Description    string   `json:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty"`
	License        *License `json:"license,omitempty"`
	Version        string   `json:"version"`
}
    OpenAPIInfo basic API information

type OpenAPIOperation struct {
	Tags        []string                   `json:"tags,omitempty"`
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	OperationID string                     `json:"operationId,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty"`
	RequestBody *OpenAPIRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
	Callbacks   map[string]interface{}     `json:"callbacks,omitempty"`
	Deprecated  bool                       `json:"deprecated,omitempty"`
	Security    []SecurityRequirement      `json:"security,omitempty"`
	Servers     []OpenAPIServer            `json:"servers,omitempty"`
	Extensions  map[string]interface{}     `json:"-"`
}
    OpenAPIOperation individual operation

type OpenAPIParameter struct {
	Name            string               `json:"name"`
	In              string               `json:"in"` // query, header, path, cookie
	Description     string               `json:"description,omitempty"`
	Required        bool                 `json:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty"`
	Schema          *OpenAPISchema       `json:"schema,omitempty"`
	Example         interface{}          `json:"example,omitempty"`
	Examples        map[string]Example   `json:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty"`
}
    OpenAPIParameter operation parameter

type OpenAPIPath map[string]*OpenAPIOperation
    OpenAPIPath operations available on a path

type OpenAPIRequestBody struct {
	Description string               `json:"description,omitempty"`
	Content     map[string]MediaType `json:"content"`
	Required    bool                 `json:"required,omitempty"`
}
    OpenAPIRequestBody corpo da request

type OpenAPIResponse struct {
	Description string               `json:"description"`
	Headers     map[string]Header    `json:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Links       map[string]Link      `json:"links,omitempty"`
}
    OpenAPIResponse operation response

type OpenAPISchema struct {
	Type                 string                    `json:"type,omitempty"`
	AllOf                []*OpenAPISchema          `json:"allOf,omitempty"`
	OneOf                []*OpenAPISchema          `json:"oneOf,omitempty"`
	AnyOf                []*OpenAPISchema          `json:"anyOf,omitempty"`
	Not                  *OpenAPISchema            `json:"not,omitempty"`
	Items                *OpenAPISchema            `json:"items,omitempty"`
	Properties           map[string]*OpenAPISchema `json:"properties,omitempty"`
	AdditionalProperties interface{}               `json:"additionalProperties,omitempty"`
	Description          string                    `json:"description,omitempty"`
	Format               string                    `json:"format,omitempty"`
	Default              interface{}               `json:"default,omitempty"`
	Title                string                    `json:"title,omitempty"`
	MultipleOf           float64                   `json:"multipleOf,omitempty"`
	Maximum              float64                   `json:"maximum,omitempty"`
	ExclusiveMaximum     bool                      `json:"exclusiveMaximum,omitempty"`
	Minimum              float64                   `json:"minimum,omitempty"`
	ExclusiveMinimum     bool                      `json:"exclusiveMinimum,omitempty"`
	MaxLength            int                       `json:"maxLength,omitempty"`
	MinLength            int                       `json:"minLength,omitempty"`
	Pattern              string                    `json:"pattern,omitempty"`
	MaxItems             int                       `json:"maxItems,omitempty"`
	MinItems             int                       `json:"minItems,omitempty"`
	UniqueItems          bool                      `json:"uniqueItems,omitempty"`
	MaxProperties        int                       `json:"maxProperties,omitempty"`
	MinProperties        int                       `json:"minProperties,omitempty"`
	Required             []string                  `json:"required,omitempty"`
	Enum                 []interface{}             `json:"enum,omitempty"`
	Example              interface{}               `json:"example,omitempty"`
	Nullable             bool                      `json:"nullable,omitempty"`
	ReadOnly             bool                      `json:"readOnly,omitempty"`
	WriteOnly            bool                      `json:"writeOnly,omitempty"`
	XML                  *XML                      `json:"xml,omitempty"`
	ExternalDocs         *ExternalDocs             `json:"externalDocs,omitempty"`
	Deprecated           bool                      `json:"deprecated,omitempty"`
	Discriminator        *Discriminator            `json:"discriminator,omitempty"`
	Ref                  string                    `json:"$ref,omitempty"`
}
    OpenAPISchema data schema

type OpenAPIServer struct {
	URL         string                    `json:"url"`
	Description string                    `json:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty"`
}
    OpenAPIServer server information

type OpenAPISpec struct {
	OpenAPI      string                 `json:"openapi"`
	Info         OpenAPIInfo            `json:"info"`
	Servers      []OpenAPIServer        `json:"servers,omitempty"`
	Paths        map[string]OpenAPIPath `json:"paths"`
	Components   *OpenAPIComponents     `json:"components,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty"`
	Tags         []OpenAPITag           `json:"tags,omitempty"`
	ExternalDocs *ExternalDocs          `json:"externalDocs,omitempty"`
}
    OpenAPISpec complete OpenAPI 3.0 specification structure

func GenerateOpenAPISpec(config *Config) *OpenAPISpec
    GenerateOpenAPISpec generates complete OpenAPI 3.0 specification

type OpenAPITag struct {
	Name         string        `json:"name"`
	Description  string        `json:"description,omitempty"`
	ExternalDocs *ExternalDocs `json:"externalDocs,omitempty"`
}
    OpenAPITag tag for grouping

type ParameterInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`     // string, int, bool, etc.
	Location    string `json:"location"` // query, path, body, header
	Required    bool   `json:"required"`
	Description string `json:"description"`
	Example     string `json:"example"`
}
    ParameterInfo represents information of a route parameter

type ParserHook func(routes []*RouteMeta) error
    ParserHook executed after parsing routes

func GetParserHooks() []ParserHook
    GetParserHooks returns all parser hooks (for testing)

type ParserStats struct {
	FilesProcessed  int               `json:"files_processed"`
	RoutesFound     int               `json:"routes_found"`
	MarkersApplied  int               `json:"markers_applied"`
	Errors          []ValidationError `json:"errors"`
	Warnings        []ValidationError `json:"warnings"`
	ProcessingTime  string            `json:"processing_time"`
	SourceDirectory string            `json:"source_directory"`
}
    ParserStats statistics do processo de parsing

type ProdConfig struct {
	Validate bool `yaml:"validate"`
	Minify   bool `yaml:"minify"`
}
    ProdConfig configuration for production mode

type PropertyInfo struct {
	Name        string        `json:"name"`
	Type        string        `json:"type,omitempty"`
	Format      string        `json:"format,omitempty"`
	Description string        `json:"description,omitempty"`
	Example     interface{}   `json:"example,omitempty"`
	Required    bool          `json:"required"`
	Enum        []string      `json:"enum,omitempty"`
	MinLength   *int          `json:"min_length,omitempty"`
	MaxLength   *int          `json:"max_length,omitempty"`
	Minimum     *float64      `json:"minimum,omitempty"`
	Maximum     *float64      `json:"maximum,omitempty"`
	Items       *PropertyInfo `json:"items,omitempty"` // For array types
	Ref         string        `json:"$ref,omitempty"`  // For schema references
}
    PropertyInfo information about a schema property

type PythonSDKGenerator struct{}
    PythonSDKGenerator generator for Python

func (p *PythonSDKGenerator) Generate(spec *OpenAPISpec, config *ClientSDKConfig) error
    Generate creates a Python client SDK from the OpenAPI specification

func (p *PythonSDKGenerator) GetFileExtension() string
    GetFileExtension retorna a extensão de arquivo para a linguagem.

func (p *PythonSDKGenerator) GetLanguage() string
    GetLanguage retorna a linguagem de programação usada.

type RateLimitConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Type       string `yaml:"type"` // "memory", "redis"
	DefaultRPS int    `yaml:"default_rps"`
	BurstSize  int    `yaml:"burst_size"`
	KeyFunc    string `yaml:"key_func"` // "ip", "user", "custom"
}
    RateLimitConfig rate limiting configuration

type RateLimitResponse struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	Limit      int    `json:"limit"`
	Remaining  int    `json:"remaining"`
	RetryAfter int    `json:"retry_after"`
}
    RateLimitResponse response when rate limit is exceeded

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int, time.Duration, error)
	Reset(ctx context.Context, key string) error
}
    RateLimiter interface for different rate limiting implementations

type RedisCache struct {
	// Has unexported fields.
}
    RedisCache Redis cache implementation

func NewRedisCache(config RedisConfig, prefix string) (*RedisCache, error)
    NewRedisCache creates a new Redis cache

func (r *RedisCache) Clear(ctx context.Context) error
    Clear clears entire cache (Redis implementation)

func (r *RedisCache) Delete(ctx context.Context, key string) error
    Delete removes cache entry (Redis implementation)

func (r *RedisCache) Get(ctx context.Context, key string) (*CacheEntry, error)
    Get retrieves cache entry (Redis implementation)

func (r *RedisCache) Set(ctx context.Context, key string, entry *CacheEntry, ttl time.Duration) error
    Set stores cache entry (Redis implementation)

func (r *RedisCache) Stats() CacheStats
    Stats returns cache statistics (Redis implementation)

type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Address  string `yaml:"address"`
	Password string `yaml:"password,omitempty"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}
    RedisConfig Redis configuration

type RedisRateLimiter struct {
	// Has unexported fields.
}
    RedisRateLimiter distributed implementation with Redis

func NewRedisRateLimiter(config RedisConfig) (*RedisRateLimiter, error)
    NewRedisRateLimiter creates a distributed rate limiter with Redis

func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, retryAfter time.Duration, err error)
    Allow checks if the request can proceed (Redis implementation)

func (r *RedisRateLimiter) Reset(_ context.Context, key string) error
    Reset clears the bucket for a key (Redis implementation)

type ResponseInfo struct {
	Code        string `json:"code"`        // HTTP status code (200, 404, etc.)
	Description string `json:"description"` // Response description
	Type        string `json:"type"`        // Schema type name (UserResponse, ErrorResponse, etc.)
	Example     string `json:"example"`     // Response example
}
    ResponseInfo represents information of a route response

type Route struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Handler     gin.HandlerFunc   `json:"-"`
	Middlewares []gin.HandlerFunc `json:"-"`
}
    Route represents a route extracted from parsing

type RouteEntry struct {
	Method            string            `json:"method"`
	Path              string            `json:"path"`
	Handler           gin.HandlerFunc   `json:"-"`
	Middlewares       []gin.HandlerFunc `json:"-"`
	FuncName          string            `json:"func_name"`
	PackageName       string            `json:"package_name"`
	FileName          string            `json:"file_name"`
	Description       string            `json:"description"`
	Summary           string            `json:"summary"`
	Tags              []string          `json:"tags"`
	MiddlewareInfo    []MiddlewareInfo  `json:"middleware_info"`
	Parameters        []ParameterInfo   `json:"parameters"`
	Group             *GroupInfo        `json:"group,omitempty"`
	Responses         []ResponseInfo    `json:"responses,omitempty"`         // Updated to use ResponseInfo
	WebSocketHandlers []string          `json:"websocketHandlers,omitempty"` // WebSocket message types this function handles
}
    RouteEntry represents complete information about a route

func GetRoutes() []RouteEntry
    GetRoutes returns all registered routes (used for documentation)

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
    RouteMeta represents metadata of a route extracted from comments

func ParseDirectory(rootDir string) ([]*RouteMeta, error)
    ParseDirectory analyzes a directory and extracts route metadata

type SDKGenerator interface {
	Generate(spec *OpenAPISpec, config *ClientSDKConfig) error
	GetLanguage() string
	GetFileExtension() string
}
    SDKGenerator interface for different SDK generators

type SDKManager struct {
	// Has unexported fields.
}
    SDKManager manages SDK generation

func NewSDKManager(config *ClientSDKConfig) *SDKManager
    NewSDKManager creates new SDK manager

func (sm *SDKManager) GenerateSDKs(spec *OpenAPISpec) error
    GenerateSDKs generates SDKs for all configured languages

func (sm *SDKManager) RegisterGenerator(language string, generator SDKGenerator)
    RegisterGenerator registers new generator

type SchemaInfo struct {
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Type        string                   `json:"type"` // "object", "array", etc.
	Properties  map[string]*PropertyInfo `json:"properties,omitempty"`
	Required    []string                 `json:"required,omitempty"`
	Example     interface{}              `json:"example,omitempty"`
	PackageName string                   `json:"package_name"`
	FileName    string                   `json:"file_name"`
}
    SchemaInfo information about a registered schema/entity

func GetSchema(name string) *SchemaInfo
    GetSchema returns a specific schema by name

type SecurityRequirement map[string][]string
    SecurityRequirement security requirement

type SecurityScheme struct {
	Type             string      `json:"type"`
	Description      string      `json:"description,omitempty"`
	Name             string      `json:"name,omitempty"`
	In               string      `json:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty"`
}
    SecurityScheme security scheme

type ServerVariable struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}
    ServerVariable server variable

type TelemetryConfig struct {
	Enabled        bool    `yaml:"enabled"`
	ServiceName    string  `yaml:"service_name"`
	ServiceVersion string  `yaml:"service_version"`
	Environment    string  `yaml:"environment"`
	Endpoint       string  `yaml:"endpoint"`
	Insecure       bool    `yaml:"insecure"`
	SampleRate     float64 `yaml:"sample_rate"`
}
    TelemetryConfig OpenTelemetry configuration

type TelemetryManager struct {
	// Has unexported fields.
}
    TelemetryManager manages OpenTelemetry configuration and instrumentation

func InitTelemetry(config *TelemetryConfig) (*TelemetryManager, error)
    InitTelemetry initializes OpenTelemetry

func (tm *TelemetryManager) Shutdown(ctx context.Context) error
    Shutdown finaliza telemetria

type TokenBucket struct {
	// Has unexported fields.
}
    TokenBucket represents a token bucket

type TracingInfo struct {
	Enabled        bool              `json:"enabled"`
	ServiceName    string            `json:"service_name"`
	ServiceVersion string            `json:"service_version"`
	Environment    string            `json:"environment"`
	Endpoint       string            `json:"endpoint"`
	SampleRate     float64           `json:"sample_rate"`
	Attributes     map[string]string `json:"attributes"`
}
    TracingInfo information about tracing for documentation

func GetTracingInfo(config *TelemetryConfig) TracingInfo
    GetTracingInfo returns information about tracing configuration

type TypeScriptSDKGenerator struct{}
    TypeScriptSDKGenerator generator for TypeScript

func (t *TypeScriptSDKGenerator) Generate(spec *OpenAPISpec, config *ClientSDKConfig) error
    Generate creates a TypeScript client SDK from the OpenAPI specification

func (t *TypeScriptSDKGenerator) GetFileExtension() string
    GetFileExtension retorna a extensão de arquivo para a linguagem.

func (t *TypeScriptSDKGenerator) GetLanguage() string
    GetLanguage retorna a linguagem de programação usada.

type ValidationConfig struct {
	Enabled       bool     `yaml:"enabled"`
	FailFast      bool     `yaml:"fail_fast"`
	CustomTags    []string `yaml:"custom_tags,omitempty"`
	ErrorFormat   string   `yaml:"error_format"`
	TranslateFunc string   `yaml:"translate_func,omitempty"`
}
    ValidationConfig validation configuration

type ValidationError struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
	Code    string `json:"code"`
}
    ValidationError validation error during parsing or generation

func (e ValidationError) Error() string

type ValidationField struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Tag     string `json:"tag"`
	Message string `json:"message"`
	Param   string `json:"param,omitempty"`
}
    ValidationField field-specific error

type ValidationResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Fields  []ValidationField      `json:"fields,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}
    ValidationResponse validation error response

type WebSocketConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ReadBuffer   int    `yaml:"read_buffer"`
	WriteBuffer  int    `yaml:"write_buffer"`
	CheckOrigin  bool   `yaml:"check_origin"`
	Compression  bool   `yaml:"compression"`
	PingInterval string `yaml:"ping_interval"`
	PongTimeout  string `yaml:"pong_timeout"`
}
    WebSocketConfig WebSocket configuration

type WebSocketConnection struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	Hub      *WebSocketHub
	UserID   string
	Groups   map[string]bool
	Metadata map[string]interface{}
	// Has unexported fields.
}
    WebSocketConnection represents a WebSocket connection

type WebSocketHandler func(conn *WebSocketConnection, message *WebSocketMessage) error
    WebSocketHandler handler type for WebSocket messages

type WebSocketHub struct {
	// Has unexported fields.
}
    WebSocketHub manages WebSocket connections

func GetWebSocketHub() *WebSocketHub
    GetWebSocketHub returns the default WebSocket hub for direct access

func InitWebSocket(config WebSocketConfig) *WebSocketHub
    InitWebSocket initializes the WebSocket system

func (h *WebSocketHub) Broadcast(message *WebSocketMessage)
    Broadcast sends message to all connections

func (h *WebSocketHub) JoinGroup(connID, groupName string) error
    JoinGroup adds connection to a group

func (h *WebSocketHub) LeaveGroup(connID, groupName string) error
    LeaveGroup removes connection from a group

func (h *WebSocketHub) SendToConnection(connID string, message *WebSocketMessage)
    SendToConnection sends message to specific connection

func (h *WebSocketHub) SendToGroup(groupName string, message *WebSocketMessage)
    SendToGroup sends message to group

type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      interface{}            `json:"data"`
	Sender    string                 `json:"sender,omitempty"`
	Target    string                 `json:"target,omitempty"` // ID da specific connection
	Group     string                 `json:"group,omitempty"`  // Nome do grupo
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
    WebSocketMessage represents a WebSocket message

func (m *WebSocketMessage) ToJSON() string
    ToJSON converts message to JSON

type WebSocketRouter struct {
	// Has unexported fields.
}
    WebSocketRouter router for WebSocket messages

func (r *WebSocketRouter) HandleMessage(conn *WebSocketConnection, message *WebSocketMessage)
    HandleMessage processes message using registered handlers

func (r *WebSocketRouter) RegisterHandler(messageType string, handler WebSocketHandler)
    RegisterHandler registers handler for message type

type XML struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Prefix    string `json:"prefix,omitempty"`
	Attribute bool   `json:"attribute,omitempty"`
	Wrapped   bool   `json:"wrapped,omitempty"`
}
    XML metadata

