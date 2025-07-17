package decorators

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileWatcher monitors handler files and automatically regenerates code
type FileWatcher struct {
	config       *Config
	watcher      *fsnotify.Watcher
	watchedFiles map[string]bool
	debouncer    *Debouncer
	isRunning    bool
	mu           sync.RWMutex
}

// Debouncer prevents too frequent regenerations
type Debouncer struct {
	duration time.Duration
	timer    *time.Timer
	mu       sync.Mutex
}

// NewFileWatcher creates a new file watcher
func NewFileWatcher(config *Config) (*FileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("error creating watcher: %v", err)
	}

	return &FileWatcher{
		config:       config,
		watcher:      watcher,
		watchedFiles: make(map[string]bool),
		debouncer:    NewDebouncer(500 * time.Millisecond), // 500ms debounce
		isRunning:    false,
	}, nil
}

// NewDebouncer creates a new debouncer
func NewDebouncer(duration time.Duration) *Debouncer {
	return &Debouncer{
		duration: duration,
	}
}

// Debounce executes the function after a delay, canceling previous executions
func (d *Debouncer) Debounce(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.duration, fn)
}

// Start starts file watching
func (fw *FileWatcher) Start() error {
	if !fw.config.Dev.Watch {
		return nil // Watching disabled
	}

	fw.mu.Lock()
	defer fw.mu.Unlock()

	if fw.isRunning {
		return nil // Already running
	}

	// Discover files to monitor
	wd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("error getting current directory: %v", err)
	}

	handlerFiles, err := fw.config.DiscoverHandlers(wd)
	if err != nil {
		return fmt.Errorf("error discovering handlers: %v", err)
	}

	// Add files to watcher
	for _, file := range handlerFiles {
		if err := fw.addFile(file); err != nil {
			log.Printf("⚠️  Error monitoring file %s: %v", file, err)
		}
	}

	// Add directories to watcher
	if err := fw.addDirectories(wd); err != nil {
		return fmt.Errorf("error monitoring directories: %v", err)
	}

	fw.isRunning = true

	// Start goroutine to process events
	go fw.watchEvents()

	LogNormal("👀 Monitoring %d files", len(fw.watchedFiles))
	return nil
}

// Stop file watching
func (fw *FileWatcher) Stop() error {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	if !fw.isRunning {
		return nil
	}

	fw.isRunning = false

	if fw.debouncer.timer != nil {
		fw.debouncer.timer.Stop()
	}

	err := fw.watcher.Close()
	if err != nil {
		return fmt.Errorf("error closing watcher: %v", err)
	}

	log.Println("🛑 File watcher stopped")
	return nil
}

// addFile adds a file to the watcher
func (fw *FileWatcher) addFile(file string) error {
	absPath, err := filepath.Abs(file)
	if err != nil {
		return err
	}

	if fw.watchedFiles[absPath] {
		return nil // Already being monitored
	}

	dir := filepath.Dir(absPath)
	if err := fw.watcher.Add(dir); err != nil {
		return err
	}

	fw.watchedFiles[absPath] = true
	return nil
}

// addDirectories adds directories to the watcher
func (fw *FileWatcher) addDirectories(rootDir string) error {
	// Monitor directories based on include patterns
	for _, pattern := range fw.config.Handlers.Include {
		// Extract base directory from pattern
		dir := extractBaseDir(pattern)
		if dir != "" {
			fullPath := filepath.Join(rootDir, dir)
			if err := fw.watcher.Add(fullPath); err != nil {
				log.Printf("⚠️  Error monitoring directory %s: %v", fullPath, err)
			}
		}
	}

	return nil
}

// extractBaseDir extracts the base directory from a pattern
func extractBaseDir(pattern string) string {
	// Remove wildcards and return base directory
	parts := strings.Split(pattern, "**")
	if len(parts) > 0 {
		base := strings.TrimSuffix(parts[0], "/")
		if base != "" && !strings.Contains(base, "*") {
			return base
		}
	}

	// Fallback: try to extract first directory without wildcards
	parts = strings.Split(pattern, "/")
	for _, part := range parts {
		if !strings.Contains(part, "*") && part != "" {
			return part
		}
	}

	return ""
}

// watchEvents processes file watcher events
func (fw *FileWatcher) watchEvents() {
	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			if fw.shouldProcessEvent(event) {
				log.Printf("📁 File modified: %s", event.Name)
				fw.debouncer.Debounce(func() {
					if err := fw.regenerateCode(); err != nil {
						log.Printf("❌ Error in automatic regeneration: %v", err)
					}
				})
			}

		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("❌ File watcher error: %v", err)
		}
	}
}

// shouldProcessEvent checks if we should process the event
func (fw *FileWatcher) shouldProcessEvent(event fsnotify.Event) bool {
	// Ignore temporary/irrelevant events
	if strings.HasSuffix(event.Name, "~") ||
		strings.HasSuffix(event.Name, ".tmp") ||
		strings.HasSuffix(event.Name, ".swp") ||
		strings.Contains(event.Name, ".git/") {
		return false
	}

	// Process only .go files that match patterns
	if !strings.HasSuffix(event.Name, ".go") {
		return false
	}

	// Check if file matches include patterns
	wd, err := filepath.Abs(".")
	if err != nil {
		return false
	}

	relPath, err := filepath.Rel(wd, event.Name)
	if err != nil {
		return false
	}

	// Check if it matches configured patterns
	return fw.matchesIncludePatterns(relPath) && !fw.matchesExcludePatterns(relPath)
}

// matchesIncludePatterns checks if file matches include patterns
func (fw *FileWatcher) matchesIncludePatterns(relPath string) bool {
	relPath = filepath.ToSlash(relPath)
	for _, pattern := range fw.config.Handlers.Include {
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
		// Verify patterns with **
		if strings.Contains(pattern, "**") {
			if fw.matchesGlobPattern(relPath, pattern) {
				return true
			}
		}
	}
	return false
}

// matchesExcludePatterns checks if file matches exclude patterns
func (fw *FileWatcher) matchesExcludePatterns(relPath string) bool {
	relPath = filepath.ToSlash(relPath)
	for _, pattern := range fw.config.Handlers.Exclude {
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
		// Verify patterns with **
		if strings.Contains(pattern, "**") {
			if fw.matchesGlobPattern(relPath, pattern) {
				return true
			}
		}
	}
	return false
}

// matchesGlobPattern checks if path matches glob pattern with **
func (fw *FileWatcher) matchesGlobPattern(path, pattern string) bool {
	regex, err := globToRegex(pattern)
	if err != nil {
		return false
	}
	return regex.MatchString(path)
}

// regenerateCode automatically regenerates the code
func (fw *FileWatcher) regenerateCode() error {
	log.Println("🔄 Automatically regenerating code...")

	// Use default configuration for regeneration
	outputPath := "./.deco/init_decorators.go"
	packageName := "deco"

	wd, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	// Discover updated handlers
	handlerFiles, err := fw.config.DiscoverHandlers(wd)
	if err != nil {
		return fmt.Errorf("error discovering handlers: %v", err)
	}

	if len(handlerFiles) == 0 {
		log.Println("⚠️  No handlers found for regeneration")
		return nil
	}

	// Generate code with configuration
	rootDir := findCommonRoot(handlerFiles)
	if err := GenerateInitFileWithConfig(rootDir, outputPath, packageName, fw.config); err != nil {
		return fmt.Errorf("error in generation: %v", err)
	}

	log.Printf("✅ Code regenerated automatically: %s", outputPath)
	return nil
}

// findCommonRoot finds the common root directory of a file list
func findCommonRoot(files []string) string {
	if len(files) == 0 {
		return "."
	}

	// Take the first file as base
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

// IsRunning returns whether the watcher is running
func (fw *FileWatcher) IsRunning() bool {
	fw.mu.RLock()
	defer fw.mu.RUnlock()
	return fw.isRunning
}
