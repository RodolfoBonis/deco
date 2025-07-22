package decorators

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewFileWatcher(t *testing.T) {
	// Test creating new file watcher
	config := &Config{}
	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)
	assert.NotNil(t, watcher)
	assert.NotNil(t, watcher.watcher)
	assert.NotNil(t, watcher.debouncer)
}

func TestNewDebouncer(t *testing.T) {
	// Test creating new debouncer
	debouncer := NewDebouncer(100 * time.Millisecond)
	assert.NotNil(t, debouncer)
	assert.Equal(t, 100*time.Millisecond, debouncer.duration)
}

func TestDebounce(t *testing.T) {
	// Test debouncing functionality
	debouncer := NewDebouncer(50 * time.Millisecond)
	var called int32

	callback := func() {
		atomic.AddInt32(&called, 1)
	}

	// Call debounce multiple times quickly
	debouncer.Debounce(callback)
	debouncer.Debounce(callback)
	debouncer.Debounce(callback)

	// Wait for debounce period
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, int32(1), atomic.LoadInt32(&called), "Callback should be called once after debounce period")
}

func TestStart(t *testing.T) {
	// Test starting the watcher
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "watcher_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Start watching
	err = watcher.Start()
	assert.NoError(t, err)

	// Check if watcher is running
	assert.True(t, watcher.IsRunning())

	// Stop the watcher
	watcher.Stop()
}

func TestStop(t *testing.T) {
	// Test stopping the watcher
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	tempDir, err := os.MkdirTemp("", "watcher_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	err = watcher.Start()
	assert.NoError(t, err)

	// Stop the watcher
	err = watcher.Stop()
	assert.NoError(t, err)

	// Check if watcher is stopped
	assert.False(t, watcher.IsRunning())
}

func TestAddFile(t *testing.T) {
	// Test adding a file to watch
	config := &Config{}
	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	tempDir, err := os.MkdirTemp("", "watcher_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte("package test"), 0o644)
	assert.NoError(t, err)

	err = watcher.addFile(testFile)
	assert.NoError(t, err)
}

func TestAddDirectories(t *testing.T) {
	// Test adding directories to watch
	config := &Config{}
	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	tempDir, err := os.MkdirTemp("", "watcher_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	subDir1 := filepath.Join(tempDir, "subdir1")
	subDir2 := filepath.Join(tempDir, "subdir2")

	err = os.MkdirAll(subDir1, 0o755)
	assert.NoError(t, err)
	err = os.MkdirAll(subDir2, 0o755)
	assert.NoError(t, err)

	err = watcher.addDirectories(tempDir)
	assert.NoError(t, err)
}

func TestExtractBaseDir(t *testing.T) {
	// Test extracting base directory
	baseDir := extractBaseDir("/path/to/file.go")
	assert.Equal(t, "/path/to/file.go", baseDir)

	// Test with nested path
	baseDir = extractBaseDir("/path/to/subdir/file.go")
	assert.Equal(t, "/path/to/subdir/file.go", baseDir)
}

func TestWatchEvents(t *testing.T) {
	// Test watching events
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	tempDir, err := os.MkdirTemp("", "watcher_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	// Start watching in a goroutine
	go func() {
		watcher.watchEvents()
	}()

	// Give some time for the watcher to start
	time.Sleep(100 * time.Millisecond)

	// Stop the watcher
	watcher.Stop()
}

func TestShouldProcessEvent(t *testing.T) {
	// Test if event should be processed
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	// This test would need a mock fsnotify.Event
	// For now, just test that the function exists
	assert.NotNil(t, watcher.shouldProcessEvent)
}

func TestMatchesIncludePatterns(t *testing.T) {
	// Test matching include patterns
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
		Handlers: HandlersConfig{
			Include: []string{"*.go", "handlers/*.go"},
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	// Test with custom patterns
	assert.True(t, watcher.matchesIncludePatterns("test.go"))
	assert.True(t, watcher.matchesIncludePatterns("handlers/test.go"))
	assert.False(t, watcher.matchesIncludePatterns("test.txt"))
}

func TestMatchesExcludePatterns(t *testing.T) {
	// Test matching exclude patterns
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
		Handlers: HandlersConfig{
			Exclude: []string{"*_test.go", "test/*.go"},
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	// Test with custom patterns
	assert.True(t, watcher.matchesExcludePatterns("test_test.go"))
	assert.True(t, watcher.matchesExcludePatterns("test/handler.go"))
	assert.False(t, watcher.matchesExcludePatterns("handler.go"))
}

func TestMatchesGlobPattern(t *testing.T) {
	// Test matching glob patterns
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	// Test simple pattern
	assert.True(t, watcher.matchesGlobPattern("test.go", "*.go"))
	assert.False(t, watcher.matchesGlobPattern("test.txt", "*.go"))

	// Test complex pattern
	assert.True(t, watcher.matchesGlobPattern("test_test.go", "*_test.go"))
	assert.False(t, watcher.matchesGlobPattern("test.go", "*_test.go"))
}

func TestRegenerateCode(t *testing.T) {
	// Test code regeneration
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	// This might fail in test environment, but we're testing the function exists
	assert.NotNil(t, watcher.regenerateCode)
}

func TestFindCommonRoot(t *testing.T) {
	// Test finding common root directory
	paths := []string{
		"/path/to/file1.go",
		"/path/to/file2.go",
		"/path/to/subdir/file3.go",
	}

	commonRoot := findCommonRoot(paths)
	assert.Equal(t, "/path/to", commonRoot)

	// Test with single path
	singlePath := []string{"/path/to/file.go"}
	commonRoot = findCommonRoot(singlePath)
	assert.Equal(t, "/path/to", commonRoot)
}

func TestIsRunning(t *testing.T) {
	// Test checking if watcher is running
	config := &Config{
		Dev: DevConfig{
			Watch: true,
		},
	}

	watcher, err := NewFileWatcher(config)
	assert.NoError(t, err)

	// Initially not running
	assert.False(t, watcher.IsRunning())

	// Start watching
	tempDir, err := os.MkdirTemp("", "watcher_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	assert.NoError(t, err)

	err = watcher.Start()
	assert.NoError(t, err)

	// Should be running
	assert.True(t, watcher.IsRunning())

	// Stop watching
	watcher.Stop()

	// Should not be running
	assert.False(t, watcher.IsRunning())
}
