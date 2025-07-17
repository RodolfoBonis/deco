package decorators

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config framework configuration structure
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

// HandlersConfig configuration for handlers discovery
type HandlersConfig struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

// GenerationConfig configuration for code generation
type GenerationConfig struct {
	Template string `yaml:"template,omitempty"`
}

// DevConfig configuration for development mode
type DevConfig struct {
	AutoDiscover bool `yaml:"auto_discover"`
	Watch        bool `yaml:"watch"`
}

// ProdConfig configuration for production mode
type ProdConfig struct {
	Validate bool `yaml:"validate"`
	Minify   bool `yaml:"minify"`
}

// RedisConfig Redis configuration
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Address  string `yaml:"address"`
	Password string `yaml:"password,omitempty"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// CacheConfig cache system configuration
type CacheConfig struct {
	Type        string `yaml:"type"` // "memory", "redis"
	DefaultTTL  string `yaml:"default_ttl"`
	MaxSize     int    `yaml:"max_size,omitempty"`
	Compression bool   `yaml:"compression"`
}

// RateLimitConfig rate limiting configuration
type RateLimitConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Type       string `yaml:"type"` // "memory", "redis"
	DefaultRPS int    `yaml:"default_rps"`
	BurstSize  int    `yaml:"burst_size"`
	KeyFunc    string `yaml:"key_func"` // "ip", "user", "custom"
}

// MetricsConfig Prometheus configuration
type MetricsConfig struct {
	Enabled   bool      `yaml:"enabled"`
	Endpoint  string    `yaml:"endpoint"`
	Namespace string    `yaml:"namespace"`
	Subsystem string    `yaml:"subsystem"`
	Buckets   []float64 `yaml:"buckets,omitempty"`
}

// OpenAPIConfig OpenAPI documentation configuration
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

// ValidationConfig validation configuration
type ValidationConfig struct {
	Enabled       bool     `yaml:"enabled"`
	FailFast      bool     `yaml:"fail_fast"`
	CustomTags    []string `yaml:"custom_tags,omitempty"`
	ErrorFormat   string   `yaml:"error_format"`
	TranslateFunc string   `yaml:"translate_func,omitempty"`
}

// WebSocketConfig WebSocket configuration
type WebSocketConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ReadBuffer   int    `yaml:"read_buffer"`
	WriteBuffer  int    `yaml:"write_buffer"`
	CheckOrigin  bool   `yaml:"check_origin"`
	Compression  bool   `yaml:"compression"`
	PingInterval string `yaml:"ping_interval"`
	PongTimeout  string `yaml:"pong_timeout"`
}

// TelemetryConfig OpenTelemetry configuration
type TelemetryConfig struct {
	Enabled        bool    `yaml:"enabled"`
	ServiceName    string  `yaml:"service_name"`
	ServiceVersion string  `yaml:"service_version"`
	Environment    string  `yaml:"environment"`
	Endpoint       string  `yaml:"endpoint"`
	Insecure       bool    `yaml:"insecure"`
	SampleRate     float64 `yaml:"sample_rate"`
}

// ClientSDKConfig SDK generation configuration
type ClientSDKConfig struct {
	Enabled     bool     `yaml:"enabled"`
	OutputDir   string   `yaml:"output_dir"`
	Languages   []string `yaml:"languages"` // "go", "python", "javascript", "typescript"
	PackageName string   `yaml:"package_name"`
	ModuleName  string   `yaml:"module_name,omitempty"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0",
		Handlers: HandlersConfig{
			Include: []string{
				"handlers/*.go",
				"handlers/**/*.go",
				"features/*/handlers/**/*.go",
				"internal/*/handlers/**/*.go",
				"pkg/*/handlers/**/*.go",
				"app/*/handlers/**/*.go",
				"src/handlers/**/*.go",
			},
			Exclude: []string{
				"**/*_test.go",
				"**/mock_*.go",
				"**/mocks/**/*.go",
				"vendor/**",
				".git/**",
				"node_modules/**",
				"**/*.pb.go",
				".deco/**",
			},
		},
		Generate: GenerationConfig{},
		Dev: DevConfig{
			AutoDiscover: true,
			Watch:        false,
		},
		Prod: ProdConfig{
			Validate: true,
			Minify:   false,
		},
		Redis: RedisConfig{
			Enabled:  false,
			Address:  "localhost:6379",
			DB:       0,
			PoolSize: 10,
		},
		Cache: CacheConfig{
			Type:        "memory",
			DefaultTTL:  "1h",
			MaxSize:     1000,
			Compression: false,
		},
		RateLimit: RateLimitConfig{
			Enabled:    false,
			Type:       "memory",
			DefaultRPS: 100,
			BurstSize:  200,
			KeyFunc:    "ip",
		},
		Metrics: MetricsConfig{
			Enabled:   false,
			Endpoint:  "/metrics",
			Namespace: "gin_decorators",
			Subsystem: "api",
			Buckets:   []float64{0.1, 0.3, 1.2, 5.0},
		},
		OpenAPI: OpenAPIConfig{
			Version:     "3.0.0",
			Title:       "API Documentation",
			Description: "Generated API documentation",
			Host:        "localhost:8080",
			BasePath:    "/api",
			Schemes:     []string{"http", "https"},
		},
		Validation: ValidationConfig{
			Enabled:     true,
			FailFast:    false,
			ErrorFormat: "json",
		},
		WebSocket: WebSocketConfig{
			Enabled:      false,
			ReadBuffer:   1024,
			WriteBuffer:  1024,
			CheckOrigin:  false,
			Compression:  false,
			PingInterval: "54s",
			PongTimeout:  "60s",
		},
		Telemetry: TelemetryConfig{
			Enabled:        false,
			ServiceName:    "gin-decorators-app",
			ServiceVersion: "1.0.0",
			Environment:    "development",
			Endpoint:       "http://localhost:4317",
			Insecure:       true,
			SampleRate:     1.0,
		},
		ClientSDK: ClientSDKConfig{
			Enabled:     false,
			OutputDir:   "./sdk",
			Languages:   []string{"go"},
			PackageName: "client",
		},
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = findConfigFile()
	}

	if configPath == "" {
		// Use default configuration if file not found
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading file de configuration %s: %v", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing da configuration: %v", err)
	}

	// Apply defaults for unspecified fields
	applyDefaults(&config)

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error serializing configuration: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("error saving configuration: %v", err)
	}

	return nil
}

// findConfigFile searches for configuration file in default locations
func findConfigFile() string {
	candidates := []string{
		".deco.yaml",
		".gin-decorators.yml",
		"gin-decorators.yaml",
		"gin-decorators.yml",
		".config/gin-decorators.yaml",
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// applyDefaults applies default values for unspecified fields
func applyDefaults(config *Config) {
	defaults := DefaultConfig()

	if config.Version == "" {
		config.Version = defaults.Version
	}

	if len(config.Handlers.Include) == 0 {
		config.Handlers.Include = defaults.Handlers.Include
	}

	if len(config.Handlers.Exclude) == 0 {
		config.Handlers.Exclude = defaults.Handlers.Exclude
	}

	// Apply defaults for Redis
	if config.Redis.Address == "" {
		config.Redis = defaults.Redis
	}

	// Apply defaults for Cache
	if config.Cache.Type == "" {
		config.Cache = defaults.Cache
	}

	// Apply defaults for RateLimit
	if config.RateLimit.Type == "" {
		config.RateLimit = defaults.RateLimit
	}

	// Apply defaults for Metrics
	if config.Metrics.Endpoint == "" {
		config.Metrics = defaults.Metrics
	}

	// Apply defaults for OpenAPI
	if config.OpenAPI.Version == "" {
		config.OpenAPI = defaults.OpenAPI
	}

	// Apply defaults for Validation
	if config.Validation.ErrorFormat == "" {
		config.Validation = defaults.Validation
	}

	// Apply defaults for WebSocket
	if config.WebSocket.ReadBuffer == 0 {
		config.WebSocket = defaults.WebSocket
	}

	// Apply defaults for Telemetry
	if config.Telemetry.ServiceName == "" {
		config.Telemetry = defaults.Telemetry
	}

	// Apply defaults for ClientSDK
	if config.ClientSDK.OutputDir == "" {
		config.ClientSDK = defaults.ClientSDK
	}
}

// DiscoverHandlers discovers handler files based on configuration
func (c *Config) DiscoverHandlers(rootDir string) ([]string, error) {
	var handlerFiles []string

	// Compile exclusion patterns
	excludePatterns, err := compilePatterns(c.Handlers.Exclude)
	if err != nil {
		return nil, fmt.Errorf("error compiling exclusion patterns: %v", err)
	}

	// Process each inclusion pattern
	for _, includePattern := range c.Handlers.Include {
		files, err := findFilesByPattern(rootDir, includePattern, excludePatterns)
		if err != nil {
			return nil, fmt.Errorf("error processing pattern '%s': %v", includePattern, err)
		}
		handlerFiles = append(handlerFiles, files...)
	}

	// Remove duplicates
	return removeDuplicates(handlerFiles), nil
}

// findFilesByPattern finds files that match the pattern
func findFilesByPattern(rootDir, pattern string, excludePatterns []*regexp.Regexp) ([]string, error) {
	var matchedFiles []string

	// Convert glob pattern to regex if necessary
	patternRegex, err := globToRegex(pattern)
	if err != nil {
		return nil, err
	}

	err = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Ignore access errors
		}

		// Skip directories that match excludes
		if d.IsDir() {
			for _, excludePattern := range excludePatterns {
				if excludePattern.MatchString(path) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if file matches pattern
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return nil
		}

		// Normalize path to always use /
		relPath = filepath.ToSlash(relPath)

		if patternRegex.MatchString(relPath) {
			// Check if not excluded
			excluded := false
			for _, excludePattern := range excludePatterns {
				if excludePattern.MatchString(relPath) {
					excluded = true
					break
				}
			}

			if !excluded {
				matchedFiles = append(matchedFiles, path)
			}
		}

		return nil
	})

	return matchedFiles, err
}

// globToRegex converts glob pattern to regex
func globToRegex(pattern string) (*regexp.Regexp, error) {
	// First, handle ** specially
	pattern = strings.ReplaceAll(pattern, "**", "⭐⭐") // temporary placeholder

	// Escape characters especiais do regex
	escaped := regexp.QuoteMeta(pattern)

	// Convert glob wildcards to regex
	escaped = strings.ReplaceAll(escaped, `⭐⭐`, `.*`)    // ** = qualquer coisa (incluindo / e zero chars)
	escaped = strings.ReplaceAll(escaped, `\*`, `[^/]*`) // * = qualquer coisa exceto /
	escaped = strings.ReplaceAll(escaped, `\?`, `[^/]`)  // ? = um caractere exceto /

	// Add anchors
	escaped = `^` + escaped + `$`

	return regexp.Compile(escaped)
}

// compilePatterns compiles pattern list to regex
func compilePatterns(patterns []string) ([]*regexp.Regexp, error) {
	var compiled []*regexp.Regexp

	for _, pattern := range patterns {
		regex, err := globToRegex(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern '%s': %v", pattern, err)
		}
		compiled = append(compiled, regex)
	}

	return compiled, nil
}

// removeDuplicates removes duplicate files from list
func removeDuplicates(files []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, file := range files {
		if !seen[file] {
			seen[file] = true
			unique = append(unique, file)
		}
	}

	return unique
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}

	if len(c.Handlers.Include) == 0 {
		return fmt.Errorf("at least one pattern de include is required")
	}

	// Validate patterns
	for _, pattern := range c.Handlers.Include {
		if _, err := globToRegex(pattern); err != nil {
			return fmt.Errorf("invalid include pattern '%s': %v", pattern, err)
		}
	}

	for _, pattern := range c.Handlers.Exclude {
		if _, err := globToRegex(pattern); err != nil {
			return fmt.Errorf("invalid exclude pattern '%s': %v", pattern, err)
		}
	}

	return nil
}
