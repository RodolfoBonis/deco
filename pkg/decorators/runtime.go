//go:build !prod
// +build !prod

package decorators

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// GlobalWatcher global file watcher instance
var GlobalWatcher *FileWatcher

// init executes automatic parsing in development (when not production build)
func init() {
	if isProdBuild() {
		return
	}

	// Do not execute runtime parsing if we are executing CLI commands
	if isCliCommand() {
		return
	}

	LogVerbose("gin-decorators: Development mode detected, starting automatic parsing...")

	// Load configuration to determine if watcher should start
	config, err := LoadConfig("")
	if err != nil {
		LogVerbose("gin-decorators: Error loading config, using default: %v", err)
		config = DefaultConfig()
	}

	// Detect handlers directory automatically
	handlersDir := detectHandlersDirectory()
	if handlersDir == "" {
		LogVerbose("gin-decorators: No 'handlers' directory found, skipping automatic parsing")
		return
	}

	// Parse routes
	routes, err := ParseDirectory(handlersDir)
	if err != nil {
		LogSilent("gin-decorators: Error in automatic parsing: %v", err)
		return
	}

	// Run hooks de parsing
	if err := executeParserHooks(routes); err != nil {
		LogSilent("gin-decorators: Error nos parser hooks: %v", err)
		return
	}

	// Register each found route
	for _, route := range routes {
		// Simulate route registration (in development we don't have real handlers)
		// In production this would be done by generated code
		LogVerbose("gin-decorators: Route found %s %s -> %s (middlewares: %d)",
			route.Method, route.Path, route.FuncName, len(route.MiddlewareCalls))
	}

	LogVerbose("gin-decorators: Automatic parsing completed - %d routes processed", len(routes))

	// Start file watcher if enabled
	if config.Dev.Watch {
		GlobalWatcher, err = NewFileWatcher(config)
		if err != nil {
			LogSilent("gin-decorators: Error creating file watcher: %v", err)
			return
		}

		if err := GlobalWatcher.Start(); err != nil {
			LogSilent("gin-decorators: Error starting file watcher: %v", err)
			return
		}

		LogVerbose("gin-decorators: File watching active - code will be regenerated automatically")
	}
}

// isProdBuild checks if we are in production build
func isProdBuild() bool {
	// In development, we assume it is not production
	// The build tag //go:build !prod already handles this, but I add extra verification
	return false
}

// isCliCommand checks if we are executing a CLI command
func isCliCommand() bool {
	// Check executable name
	if len(os.Args) > 0 {
		execName := filepath.Base(os.Args[0])
		if strings.Contains(execName, "deco") {
			return true
		}
	}

	// Check arguments that indicate CLI command
	for _, arg := range os.Args[1:] {
		if arg == "init" || arg == "-init" || arg == "--init" ||
			arg == "version" || arg == "-version" || arg == "--version" ||
			arg == "help" || arg == "-help" || arg == "--help" {
			return true
		}
	}

	return false
}

// detectHandlersDirectory tries to automatically find handlers directory
func detectHandlersDirectory() string {
	// Possible handler directories in order of preference
	candidates := []string{
		"./handlers",
		"./internal/handlers",
		"./pkg/handlers",
		"./app/handlers",
		"./src/handlers",
		"../handlers", // caso estejamos em subdirectory
		"../../handlers",
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		LogSilent("gin-decorators: Error getting current directory: %v", err)
		return ""
	}

	LogVerbose("gin-decorators: Looking for handlers starting from: %s", wd)

	// Check each candidate
	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}

		// Check if directory exists
		if info, err := os.Stat(absPath); err == nil && info.IsDir() {
			// Check if it contains .go files
			if hasGoFiles(absPath) {
				LogVerbose("gin-decorators: Directory handlers found: %s", absPath)
				return absPath
			}
		}
	}

	// Try to search recursively in current directory
	var foundDir string
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "handler") {
			if hasGoFiles(path) {
				foundDir = path
				return filepath.SkipDir // stop search
			}
		}

		return nil
	}); err != nil {
		LogSilent("gin-decorators: Error walking directory: %v", err)
	}

	if foundDir != "" {
		absPath, _ := filepath.Abs(foundDir)
		log.Printf("gin-decorators: Directory handlers found via search: %s", absPath)
		return absPath
	}

	return ""
}

// hasGoFiles checks if a directory contains .go files
func hasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
			return true
		}
	}

	return false
}
