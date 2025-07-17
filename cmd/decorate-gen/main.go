package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	decorators "github.com/RodolfoBonis/deco/pkg/decorators"
	"github.com/fsnotify/fsnotify"
)

func main() {
	// Check for init command before flag parsing
	if len(os.Args) > 1 && os.Args[1] == "init" {
		verbose := contains(os.Args, "-v") || contains(os.Args, "--verbose")
		if err := handleInitCommand(verbose); err != nil {
			log.Fatalf("❌ Error in init command: %v", err)
		}
		return
	}

	// Check for dev command for hot reload
	if len(os.Args) > 1 && os.Args[1] == "dev" {
		verbose := contains(os.Args, "-v") || contains(os.Args, "--verbose")
		port := "8080"
		if len(os.Args) > 2 && strings.HasPrefix(os.Args[2], "--port=") {
			port = strings.TrimPrefix(os.Args[2], "--port=")
		}
		if err := handleDevCommand(verbose, port); err != nil {
			log.Fatalf("❌ Error in dev command: %v", err)
		}
		return
	}

	var (
		// Main flags
		configPath   = flag.String("config", "", "Configuration file path")
		rootDir      = flag.String("root", "", "Root directory to search for handlers (overrides config)")
		outputPath   = flag.String("out", "", "Output file path (overrides config)")
		packageName  = flag.String("pkg", "", "Package name for the generated file (overrides config)")
		templatePath = flag.String("template", "", "Path to custom template (overrides config)")
		validate     = flag.Bool("validate", true, "Validate generated file")
		verbose      = flag.Bool("v", false, "Verbose output")
		version      = flag.Bool("version", false, "Show version")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "deco Code Generator v1.0.0\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [command]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  init                 Create .deco.yaml configuration file\n")
		fmt.Fprintf(os.Stderr, "  generate (default)   Generate code based on configuration\n")
		fmt.Fprintf(os.Stderr, "  dev                  Start development server with hot reload\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s init                                    # Create default configuration\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s                                         # Use .deco.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -config custom.yaml                     # Use custom configuration\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -root ./handlers -out ./init.go -pkg handlers  # Legacy mode\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s dev                                     # Development mode with hot reload\n", os.Args[0])
	}

	flag.Parse()

	// Version command
	if *version {
		fmt.Println("deco Code Generator v1.0.0")
		fmt.Println("Code generator for the deco framework")
		return
	}

	// Configure logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(0)
	}

	// Generate command (default)
	if err := handleGenerateCommand(*configPath, *rootDir, *outputPath, *packageName, *templatePath, *validate, *verbose); err != nil {
		log.Fatalf("❌ Generation error: %v", err)
	}
}

// handleInitCommand executes the initialization command
func handleInitCommand(verbose bool) error {
	configFile := ".deco.yaml"

	// Check if it already exists
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("⚠️  File %s already exists.\n", configFile)
		fmt.Print("Do you want to overwrite? (y/N): ")

		var response string
		if _, err := fmt.Scanln(&response); err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return nil
		}

		if response != "y" && response != "Y" && response != "yes" {
			fmt.Println("❌ Operation cancelled.")
			return nil
		}
	}

	// Create default configuration
	config := decorators.DefaultConfig()

	if err := decorators.SaveConfig(config, configFile); err != nil {
		return fmt.Errorf("error saving configuration: %v", err)
	}

	fmt.Printf("✅ Configuration file created: %s\n\n", configFile)

	if verbose {
		fmt.Println("📋 Configuration created with:")
		fmt.Printf("   - %d include patterns\n", len(config.Handlers.Include))
		fmt.Printf("   - %d exclude patterns\n", len(config.Handlers.Exclude))
		fmt.Printf("   - Fixed output: ./.deco/init_decorators.go\n")
		fmt.Printf("   - Fixed package: deco\n")
		fmt.Println("\n🔧 Configurable features:")
		fmt.Printf("   - Redis: %v (Address: %s)\n", config.Redis.Enabled, config.Redis.Address)
		fmt.Printf("   - Cache: %s (TTL: %s)\n", config.Cache.Type, config.Cache.DefaultTTL)
		fmt.Printf("   - Rate Limiting: %v (RPS: %d)\n", config.RateLimit.Enabled, config.RateLimit.DefaultRPS)
		fmt.Printf("   - Metrics/Prometheus: %v (Endpoint: %s)\n", config.Metrics.Enabled, config.Metrics.Endpoint)
		fmt.Printf("   - OpenAPI: %s (%s)\n", config.OpenAPI.Version, config.OpenAPI.Title)
		fmt.Printf("   - Validation: %v (Format: %s)\n", config.Validation.Enabled, config.Validation.ErrorFormat)
		fmt.Printf("   - WebSockets: %v (Buffers: %d/%d)\n", config.WebSocket.Enabled, config.WebSocket.ReadBuffer, config.WebSocket.WriteBuffer)
		fmt.Printf("   - Telemetry/OpenTelemetry: %v (Service: %s)\n", config.Telemetry.Enabled, config.Telemetry.ServiceName)
		fmt.Printf("   - Client SDK: %v (Languages: %v)\n", config.ClientSDK.Enabled, config.ClientSDK.Languages)
	}

	// 🚀 NEW: Automatically generate initial code
	fmt.Println("🔍 Checking for existing handlers...")

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("⚠️  Error getting current directory: %v\n", err)
		return printNextSteps()
	}

	handlerFiles, err := config.DiscoverHandlers(wd)
	if err != nil {
		fmt.Printf("⚠️  Error discovering handlers: %v\n", err)
		return printNextSteps()
	}

	if len(handlerFiles) == 0 {
		fmt.Println("📁 No handlers found yet.")
		return printNextSteps()
	}

	fmt.Printf("✨ Found %d handlers! Generating initial code...\n", len(handlerFiles))

	if verbose {
		fmt.Println("📋 Handlers found:")
		for _, file := range handlerFiles {
			fmt.Printf("   - %s\n", file)
		}
	}

	// Run initial generation
	if err := handleGenerateCommand(configFile, "", "", "", "", true, verbose); err != nil {
		fmt.Printf("⚠️  Error in initial generation: %v\n", err)
		return printNextSteps()
	}

	fmt.Println("\n🎉 Project initialized successfully!")
	fmt.Println("📁 Generated file: ./.deco/init_decorators.go")
	fmt.Println("\n🚀 Next steps:")
	fmt.Println("   1. Import the generated package in your main.go:")
	fmt.Println("      import _ \"yourmodule/.deco\"")
	fmt.Println("   2. Run: go run main.go")
	fmt.Println("   3. Access: http://localhost:8080/decorators/docs")
	fmt.Println("\n💡 To add new routes:")
	fmt.Println("   - Create handlers with @Route decorators")
	fmt.Println("   - Run: deco --config .deco.yaml")
	fmt.Println("\n🔥 For development with hot reload:")
	fmt.Println("   - Run: deco dev")
	fmt.Println("\n📖 For more information: https://github.com/RodolfoBonis/deco")

	return nil
}

// printNextSteps prints instructions when automatic generation is not possible
func printNextSteps() error {
	fmt.Println("\n🔧 Next steps:")
	fmt.Println("   1. Create handlers with @Route decorators")
	fmt.Println("   2. Run: deco --config .deco.yaml")
	fmt.Println("   3. Import the generated package in your main.go:")
	fmt.Println("      import _ \"yourmodule/.deco\"")
	fmt.Println("   4. Run: go run main.go")
	fmt.Println("\n📖 For more information: https://github.com/RodolfoBonis/deco")
	return nil
}

// handleGenerateCommand executes generation command
func handleGenerateCommand(configPath, rootDir, outputPath, packageName, templatePath string, validate, verbose bool) error {
	startTime := time.Now()

	// Load configuration
	config, err := decorators.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %v", err)
	}

	// Override configuration with flags if provided
	if rootDir != "" {
		// Legacy mode: use root flag
		if verbose {
			log.Printf("🔧 Using legacy mode with -root: %s", rootDir)
		}
		return handleLegacyMode(rootDir, outputPath, packageName, templatePath, validate, verbose, startTime)
	}

	// Use configuration for discovery
	if verbose {
		log.Printf("📋 Using configuration for discovery")
		log.Printf("   - Include patterns: %v", config.Handlers.Include)
		log.Printf("   - Exclude patterns: %v", config.Handlers.Exclude)
	}

	// Discover handlers based on configuration
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	handlerFiles, err := config.DiscoverHandlers(wd)
	if err != nil {
		return fmt.Errorf("error discovering handlers: %v", err)
	}

	if len(handlerFiles) == 0 {
		log.Printf("⚠️  No handlers found with configured patterns")
		log.Printf("💡 Tip: Run 'deco init' to generate default configuration")
		return nil
	}

	if verbose {
		log.Printf("🔍 Handlers found (%d):", len(handlerFiles))
		for _, file := range handlerFiles {
			log.Printf("   - %s", file)
		}
	}

	// Force use of .deco folder in root (not customizable)
	finalOutput := "./.deco/init_decorators.go"
	finalPackage := "deco"

	// Ignore user output and package configurations
	if outputPath != "" && verbose {
		log.Printf("⚠️  Ignoring -out: always uses ./.deco/init_decorators.go")
	}
	if packageName != "" && verbose {
		log.Printf("⚠️  Ignoring -pkg: always uses package deco")
	}

	finalTemplate := config.Generate.Template
	if templatePath != "" {
		finalTemplate = templatePath
	}

	// Final logs
	if verbose {
		log.Printf("📄 Output file: %s", finalOutput)
		log.Printf("�� Package name: %s", finalPackage)
		if finalTemplate != "" {
			log.Printf("🎨 Custom template: %s", finalTemplate)
		}
	}

	// Generate using configuration-based discovery
	err = generateFromFilesWithConfig(handlerFiles, finalOutput, finalPackage, finalTemplate, validate, verbose, config)
	if err != nil {
		return err
	}

	// Final statistics only in verbose mode
	duration := time.Since(startTime)
	if verbose {
		log.Printf("✅ Generation completed in %v", duration)
		log.Printf("📁 File created: %s", finalOutput)
	}

	return nil
}

// handleLegacyMode executes generation in legacy mode (compatibility)
func handleLegacyMode(rootDir, outputPath, packageName, templatePath string, validate, verbose bool, startTime time.Time) error {
	// Validate legacy arguments
	if rootDir == "" {
		return fmt.Errorf("root directory is required (-root)")
	}

	// Force use of .deco folder in root (not customizable)
	if outputPath != "" && verbose {
		log.Printf("⚠️  Ignoring -out: always uses ./.deco/init_decorators.go")
	}
	if packageName != "" && verbose {
		log.Printf("⚠️  Ignoring -pkg: always uses package deco")
	}

	outputPath = "./.deco/init_decorators.go"
	packageName = "deco"

	// Convert to absolute paths
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return fmt.Errorf("error resolving root path: %v", err)
	}

	absOutputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return fmt.Errorf("error resolving output path: %v", err)
	}

	// Verify if root directory exists
	if _, err := os.Stat(absRootDir); os.IsNotExist(err) {
		return fmt.Errorf("root directory not found: %s", absRootDir)
	}

	// Initial log
	if verbose {
		log.Printf("🔍 Analyzing directory: %s", absRootDir)
		log.Printf("📄 Output file: %s", absOutputPath)
		log.Printf("📦 Package name: %s", packageName)
	}

	// Generate file using legacy method
	var genErr error
	if templatePath != "" {
		// Use custom template
		absTemplatePath, err := filepath.Abs(templatePath)
		if err != nil {
			return fmt.Errorf("error resolving template path: %v", err)
		}

		if verbose {
			log.Printf("🎨 Using custom template: %s", absTemplatePath)
		}

		genErr = decorators.GenerateFromTemplate(absRootDir, absTemplatePath, absOutputPath, packageName)
	} else {
		// Load configuration for legacy mode
		config, configErr := decorators.LoadConfig("")
		if configErr != nil {
			config = decorators.DefaultConfig()
		}

		// Use default template with configuration
		genErr = decorators.GenerateInitFileWithConfig(absRootDir, absOutputPath, packageName, config)
	}

	if genErr != nil {
		return genErr
	}

	// Validation is now done automatically within GenerateInitFileWithConfig if enabled
	// Manual validation only if not done automatically
	if validate {
		// Verify if validation was already done automatically
		config, _ := decorators.LoadConfig("")
		if config == nil || !config.Prod.Validate {
			if verbose {
				log.Printf("✅ Validating generated file...")
			}

			if err := decorators.ValidateGeneration(absOutputPath); err != nil {
				enhancedErr := enhanceErrorWithSourceInfo(err, ".deco.yaml")
				return fmt.Errorf("validation failed: %v", enhancedErr)
			}

			if verbose {
				log.Printf("✅ File validated successfully")
			}
		}
	}

	// Final statistics
	duration := time.Since(startTime)

	log.Printf("✅ Generation completed in %v", duration)
	log.Printf("📁 File created: %s", absOutputPath)

	// Show next steps
	if verbose {
		log.Printf("\n📋 Next steps:")
		log.Printf("   1. Run: go build -tags prod")
		log.Printf("   2. File %s will be used in production", filepath.Base(absOutputPath))
	}

	return nil
}

// generateFromFilesWithConfig generates code with specific configuration
func generateFromFilesWithConfig(handlerFiles []string, outputPath, packageName, templatePath string, _, verbose bool, config *decorators.Config) error {
	// Extract common directory from files to use as root
	rootDir := findCommonRoot(handlerFiles)

	if verbose {
		log.Printf("📁 Root directory detected: %s", rootDir)
	}

	// Use default configuration if not provided
	if config == nil {
		config = decorators.DefaultConfig()
	}

	// Use default generation in root directory
	if templatePath != "" {
		return decorators.GenerateFromTemplate(rootDir, templatePath, outputPath, packageName)
	}

	return decorators.GenerateInitFileWithConfig(rootDir, outputPath, packageName, config)
}

// findCommonRoot finds the common root directory of a file list
func findCommonRoot(files []string) string {
	if len(files) == 0 {
		return "."
	}

	// Get the first file as base
	common := filepath.Dir(files[0])

	// Find common prefix with all others
	for _, file := range files[1:] {
		dir := filepath.Dir(file)

		// Find common prefix between common and dir
		for !strings.HasPrefix(dir, common) {
			common = filepath.Dir(common)
			if common == "." || common == "/" {
				break
			}
		}
	}

	return common
}

// contains checks if slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// handleDevCommand executes hot reload development server
func handleDevCommand(verbose bool, port string) error {
	// Configure logging based on verbose flag
	decorators.SetVerbose(verbose)

	fmt.Printf("🔥 Starting development server with hot reload on port %s...\n", port)

	configFile := ".deco.yaml"

	// Verify if config exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("⚠️  File .deco.yaml not found. Run 'deco init' first.")
		return nil
	}

	// Load configuration
	config, err := decorators.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("error loading configuration: %v", err)
	}

	// Perform initial generation
	if verbose {
		fmt.Println("🔄 Generating initial code...")
	}
	if err := handleGenerateCommand(configFile, "", "", "", "", true, verbose); err != nil {
		return fmt.Errorf("error in initial generation: %v", err)
	}

	// Setup to capture signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Channel for communication with watcher
	reloadChan := make(chan bool, 1)
	errorChan := make(chan error, 1)

	// Create DevServer
	devServer := &DevServer{
		Port:       port,
		Verbose:    verbose,
		Config:     config,
		ConfigFile: configFile,
		ReloadChan: reloadChan,
		ErrorChan:  errorChan,
		SigChan:    sigChan,
	}

	// Start server
	if err := devServer.Start(); err != nil {
		return err
	}

	// Start file watcher
	if err := devServer.StartWatcher(); err != nil {
		return err
	}

	if verbose {
		fmt.Println("👀 Monitoring changes automatically...")
		fmt.Println("📝 Edit your handlers - automatic regeneration and restart!")
		fmt.Println("⏹️  Ctrl+C to stop")
	}

	// Main dev server loop
	return devServer.Run()
}

// DevServer manages the development server with hot reload
type DevServer struct {
	Port       string
	Verbose    bool
	Config     *decorators.Config
	ConfigFile string
	ReloadChan chan bool
	ErrorChan  chan error
	SigChan    chan os.Signal

	serverCmd    *exec.Cmd
	watcher      *decorators.FileWatcher
	isRunning    bool
	restartCount int
}

// Start starts the server for the first time
func (ds *DevServer) Start() error {
	return ds.startServer()
}

// StartWatcher starts the file watcher
func (ds *DevServer) StartWatcher() error {
	// Ensure watch is enabled
	ds.Config.Dev.Watch = true

	watcher, err := decorators.NewFileWatcher(ds.Config)
	if err != nil {
		return fmt.Errorf("error creating file watcher: %v", err)
	}

	ds.watcher = watcher

	// Start watcher with custom callback
	return ds.startWatcherWithCallback()
}

// startWatcherWithCallback starts watcher with callback for hot reload
func (ds *DevServer) startWatcherWithCallback() error {
	// Discover files to monitor
	wd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	handlerFiles, err := ds.Config.DiscoverHandlers(wd)
	if err != nil {
		return fmt.Errorf("error discovering handlers: %v", err)
	}

	if len(handlerFiles) == 0 {
		fmt.Println("⚠️  No handlers found to monitor")
		return nil
	}

	// Monitor handler directories
	monitoredDirs := make(map[string]bool)
	for _, file := range handlerFiles {
		dir := filepath.Dir(file)
		if !monitoredDirs[dir] {
			monitoredDirs[dir] = true
		}
	}

	// DO NOT monitor .deco directory to avoid infinite loop

	if ds.Verbose {
		fmt.Printf("🔍 Monitoring directories: %v\n", getKeys(monitoredDirs))
	}

	// Start watcher
	if err := ds.watcher.Start(); err != nil {
		return err
	}

	// Goroutine to process file watching events
	go ds.watchFiles()

	return nil
}

// watchFiles processes file watcher events
func (ds *DevServer) watchFiles() {
	// Debouncing to avoid multiple regenerations
	var debounceTimer *time.Timer
	debounceDuration := 500 * time.Millisecond

	regenerate := func() {
		if ds.Verbose {
			fmt.Println("🔄 Changes detected, regenerating...")
		}

		// Regenerate code
		if err := handleGenerateCommand(ds.ConfigFile, "", "", "", "", true, false); err != nil {
			// Enhanced error reporting with source file information
			enhancedErr := enhanceErrorWithSourceInfo(err, ds.ConfigFile)
			fmt.Printf("❌ Error in regeneration: %v\n", enhancedErr)
			ds.ErrorChan <- enhancedErr
			return
		}

		if ds.Verbose {
			fmt.Println("✅ Code regenerated, restarting server...")
		}

		// Signal reload
		select {
		case ds.ReloadChan <- true:
		default:
			// Channel full, ignore
		}
	}

	// Use real file watching system
	// Discover and monitor files
	wd, _ := filepath.Abs(".")
	handlerFiles, err := ds.Config.DiscoverHandlers(wd)
	if err != nil {
		fmt.Printf("❌ Error discovering handlers: %v\n", err)
		return
	}

	// Monitor handler directories
	monitoredDirs := make(map[string]bool)
	for _, file := range handlerFiles {
		dir := filepath.Dir(file)
		if !monitoredDirs[dir] {
			monitoredDirs[dir] = true
		}
	}

	// DO NOT monitor .deco directory to avoid infinite loop

	if ds.Verbose {
		fmt.Printf("🔍 Monitoring directories: %v\n", getKeys(monitoredDirs))
	}

	// Create a custom watcher using fsnotify
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("❌ Error creating fsnotify watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	// Add directories to watcher
	for dir := range monitoredDirs {
		if err := watcher.Add(dir); err != nil {
			fmt.Printf("⚠️ Error monitoring directory %s: %v\n", dir, err)
		} else if ds.Verbose {
			fmt.Printf("👀 Monitoring directory: %s\n", dir)
		}
	}

	// Event loop
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			// Filter only relevant .go files
			if ds.shouldProcessEvent(event) {
				if ds.Verbose {
					fmt.Printf("📁 Modified: %s\n", event.Name)
				}

				// Debounce
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDuration, regenerate)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("❌ Error in file watcher: %v\n", err)
		}
	}
}

// shouldProcessEvent checks if we should process the file event
func (ds *DevServer) shouldProcessEvent(event fsnotify.Event) bool {
	// Ignore irrelevant events
	if event.Op&fsnotify.Chmod == fsnotify.Chmod {
		return false // Ignore permission changes
	}

	// Process only .go files
	if !strings.HasSuffix(event.Name, ".go") {
		return false
	}

	// Ignore temporary files
	if strings.HasSuffix(event.Name, "~") ||
		strings.HasSuffix(event.Name, ".tmp") ||
		strings.HasSuffix(event.Name, ".swp") ||
		strings.Contains(event.Name, ".git/") {
		return false
	}

	eventPath, err := filepath.Abs(event.Name)
	if err != nil {
		return false
	}

	// DO NOT process .deco/init_decorators.go file to avoid infinite loop
	initDecoratorsPath, err := filepath.Abs("./.deco/init_decorators.go")
	if err == nil && eventPath == initDecoratorsPath {
		if ds.Verbose {
			fmt.Printf("⏭️  Ignoring .deco/init_decorators.go (generated file)\n")
		}
		return false
	}

	// Verify if the file is in the list of monitored handlers
	wd, err := filepath.Abs(".")
	if err != nil {
		return false
	}

	handlerFiles, err := ds.Config.DiscoverHandlers(wd)
	if err != nil {
		return false
	}

	// Verify if the modified file is one of the handlers
	for _, handlerFile := range handlerFiles {
		handlerPath, err := filepath.Abs(handlerFile)
		if err != nil {
			continue
		}
		if eventPath == handlerPath {
			return true
		}
	}

	return false
}

// getKeys returns the keys of a map as slice
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Run executes the main dev server loop
func (ds *DevServer) Run() error {
	for {
		select {
		case <-ds.ReloadChan:
			// Server restart
			if err := ds.restartServer(); err != nil {
				fmt.Printf("❌ Error restarting server: %v\n", err)
			}

		case err := <-ds.ErrorChan:
			fmt.Printf("⚠️  Error in dev server: %v\n", err)

		case <-ds.SigChan:
			fmt.Println("\n🛑 Stopping development server...")
			return ds.Stop()
		}
	}
}

// startServer starts the server process
func (ds *DevServer) startServer() error {
	// Verify if port is free before trying to start
	if !ds.isPortFree() {
		return fmt.Errorf("port :%s is already in use", ds.Port)
	}

	cmd := exec.Command("go", "run", "main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%s", ds.Port))

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error starting server: %v", err)
	}

	ds.serverCmd = cmd
	ds.isRunning = true
	ds.restartCount++

	fmt.Printf("✅ Server started (PID: %d, restart #%d)\n", cmd.Process.Pid, ds.restartCount)

	// Goroutine to monitor the process
	go func() {
		if err := cmd.Wait(); err != nil {
			if ds.isRunning {
				// Process died unexpectedly
				fmt.Printf("⚠️  Server stopped unexpectedly: %v\n", err)
				ds.isRunning = false
			}
		}
	}()

	// Wait a bit for the server to initialize
	// We don't need to check the port again as it may be in the process of binding
	time.Sleep(2 * time.Second) // Increased to 2s to give Gin time to bind

	return nil
}

// restartServer restarts the server gracefully
func (ds *DevServer) restartServer() error {
	fmt.Println("🔄 Restarting server...")

	// Stop current server if running
	if err := ds.stopServer(); err != nil {
		fmt.Printf("⚠️  Error stopping server: %v\n", err)
	}

	// Wait for port to become free
	if err := ds.waitForPortFree(); err != nil {
		return fmt.Errorf("error waiting for port to be free: %v", err)
	}

	// Start new server
	return ds.startServer()
}

// stopServer stops the current server robustly
func (ds *DevServer) stopServer() error {
	if ds.serverCmd == nil || ds.serverCmd.Process == nil {
		return nil
	}

	pid := ds.serverCmd.Process.Pid
	if ds.Verbose {
		fmt.Printf("🛑 Stopping server (PID: %d)...\n", pid)
	}

	// Mark as not running to avoid "stopped unexpectedly" logs
	ds.isRunning = false

	// Try graceful shutdown with SIGINT (Go responds better to this)
	if err := ds.serverCmd.Process.Signal(syscall.SIGINT); err != nil {
		if ds.Verbose {
			fmt.Printf("⚠️  SIGINT failed: %v, trying SIGTERM...\n", err)
		}
		if err := ds.serverCmd.Process.Signal(syscall.SIGTERM); err != nil {
			if ds.Verbose {
				fmt.Printf("⚠️  SIGTERM failed: %v, using SIGKILL...\n", err)
			}
			if err := ds.serverCmd.Process.Kill(); err != nil {
				if ds.Verbose {
					fmt.Printf("⚠️  SIGKILL failed: %v\n", err)
				}
			}
		}
	}

	// Wait for process to terminate with extended timeout
	done := make(chan error, 1)
	go func() {
		done <- ds.serverCmd.Wait()
	}()

	select {
	case err := <-done:
		if ds.Verbose {
			fmt.Printf("✅ Server stopped gracefully (PID: %d)\n", pid)
		}
		return err
	case <-time.After(10 * time.Second): // Extended timeout to 10s
		if ds.Verbose {
			fmt.Printf("⏰ Timeout waiting, forcing kill (PID: %d)...\n", pid)
		}
		if err := ds.serverCmd.Process.Kill(); err != nil {
			if ds.Verbose {
				fmt.Printf("⚠️  Force kill failed: %v\n", err)
			}
		}
		<-done // Wait for Kill to finish
		return nil
	}
}

// waitForPortFree waits for the port to become available
func (ds *DevServer) waitForPortFree() error {
	maxAttempts := 20 // 20 attempts = 2 seconds maximum
	for i := 0; i < maxAttempts; i++ {
		if ds.isPortFree() {
			if ds.Verbose && i > 0 {
				fmt.Printf("✅ Port :%s freed after %d attempts (%.1fs)\n", ds.Port, i+1, float64(i+1)*0.1)
			}
			return nil
		}
		if ds.Verbose && i == 0 {
			fmt.Printf("⏳ Waiting for port :%s to become free...\n", ds.Port)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// If we got here, try to force kill processes on the port
	if ds.Verbose {
		fmt.Printf("⚠️  Timeout waiting for port :%s, trying forced kill...\n", ds.Port)
	}

	// Try to kill processes using the port (macOS/Linux)
	ds.killProcessesOnPort()

	// Try a few more times after forced kill
	for i := 0; i < 5; i++ {
		if ds.isPortFree() {
			if ds.Verbose {
				fmt.Printf("✅ Port :%s freed after forced kill\n", ds.Port)
			}
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for port :%s to become free (tried forced kill)", ds.Port)
}

// killProcessesOnPort tries to kill processes using the port
func (ds *DevServer) killProcessesOnPort() {
	// Validate port to prevent command injection
	if ds.Port == "" || !isValidPort(ds.Port) {
		if ds.Verbose {
			fmt.Printf("⚠️  Invalid port: %s\n", ds.Port)
		}
		return
	}

	// Command to find and kill processes on the port (works on macOS and Linux)
	// #nosec G204
	cmd := exec.Command("sh", "-c", fmt.Sprintf("lsof -ti :%s | xargs -r kill -9", ds.Port))
	if err := cmd.Run(); err != nil && ds.Verbose {
		fmt.Printf("⚠️  Could not force kill on port :%s: %v\n", ds.Port, err)
	}
}

// isValidPort validates if the port string is safe for command execution
func isValidPort(port string) bool {
	// Check if port is numeric and within valid range
	if port == "" {
		return false
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	return portNum > 0 && portNum <= 65535
}

// isPortFree checks if the port is available
func (ds *DevServer) isPortFree() bool {
	// Try to bind to the port to check if it's free
	addr := fmt.Sprintf(":%s", ds.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false // Port occupied
	}
	listener.Close()
	return true // Port free
}

// Stop stops the dev server
func (ds *DevServer) Stop() error {
	fmt.Println("🛑 Stopping dev server...")

	// Stop watcher
	if ds.watcher != nil {
		if err := ds.watcher.Stop(); err != nil {
			fmt.Printf("⚠️  Error stopping watcher: %v\n", err)
		}
	}

	// Stop server using robust method
	if err := ds.stopServer(); err != nil {
		fmt.Printf("⚠️  Error stopping server: %v\n", err)
	}

	fmt.Println("✅ Dev server stopped.")
	return nil
}

// enhanceErrorWithSourceInfo attempts to map errors from generated files back to source files
func enhanceErrorWithSourceInfo(err error, configFile string) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Handle decorator validation errors directly
	if multiErr, ok := err.(*decorators.MultipleValidationError); ok {
		var messages []string
		for _, valErr := range multiErr.Errors {
			messages = append(messages, valErr.Error())
		}
		return fmt.Errorf("❌ Decorator errors found:\n%s", strings.Join(messages, "\n"))
	}

	// Handle single validation error
	if valErr, ok := err.(*decorators.ValidationError); ok {
		return fmt.Errorf("❌ Decorator error: %s", valErr.Error())
	}

	// Check if it's a parsing error during generation
	if strings.Contains(errStr, "./.deco/init_decorators.go") {
		// Extract line information if available
		if strings.Contains(errStr, "syntax error") || strings.Contains(errStr, "expected") {
			// Load config to get handler files
			config, configErr := decorators.LoadConfig(configFile)
			if configErr != nil {
				return fmt.Errorf("❌ Decorator syntax error - check your handler files")
			}

			// Get working directory
			wd, _ := filepath.Abs(".")
			handlerFiles, _ := config.DiscoverHandlers(wd)

			if len(handlerFiles) > 0 {
				// Try to extract line number and map to source file
				lineInfo := extractLineInfoFromError(errStr)
				if lineInfo != "" {
					return fmt.Errorf("❌ Decorator syntax error %s - check: %s", lineInfo, getFirstFewFiles(handlerFiles))
				}
				return fmt.Errorf("❌ Decorator syntax error - check: %s", getFirstFewFiles(handlerFiles))
			}

			return fmt.Errorf("❌ Decorator syntax error - check your handler files")
		}

		// Handle compilation errors
		if strings.Contains(errStr, "cannot use") || strings.Contains(errStr, "undefined") {
			return fmt.Errorf("❌ Compilation error - check your decorator syntax and imports")
		}
	}

	// Handle other Go compilation errors
	if strings.Contains(errStr, "go build") || strings.Contains(errStr, "go run") {
		if strings.Contains(errStr, "expected") {
			return fmt.Errorf("❌ Go syntax error - check parentheses, quotes and commas in decorators")
		}
		if strings.Contains(errStr, "undefined") {
			return fmt.Errorf("❌ Error: function or type not defined - check imports and function names")
		}
	}

	return err
}

// extractLineInfoFromError extracts line information from error messages
func extractLineInfoFromError(errStr string) string {
	// Look for patterns like ":123:" or ":123:123:"
	re := regexp.MustCompile(`:\d+:\d*:`)
	match := re.FindString(errStr)
	if match != "" {
		return match
	}
	return ""
}

// getFirstFewFiles returns a simple comma-separated list of the first few files
func getFirstFewFiles(files []string) string {
	if len(files) == 0 {
		return "no handler files found"
	}

	if len(files) == 1 {
		return filepath.Base(files[0])
	}

	if len(files) <= 3 {
		var names []string
		for _, file := range files {
			names = append(names, filepath.Base(file))
		}
		return strings.Join(names, ", ")
	}

	// More than 3 files, show first 2 and "..."
	return fmt.Sprintf("%s, %s, ... (%d more)",
		filepath.Base(files[0]),
		filepath.Base(files[1]),
		len(files)-2)
}
