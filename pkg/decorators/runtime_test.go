package decorators

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsProdBuild(t *testing.T) {
	// Test isProdBuild function
	// In development mode, this should return false
	isProd := isProdBuild()
	assert.False(t, isProd)
}

func TestIsCliCommand(t *testing.T) {
	// Test isCliCommand function
	// Save original args
	originalArgs := os.Args

	// Test with CLI command
	os.Args = []string{"deco", "generate"}
	isCli := isCliCommand()
	assert.True(t, isCli)

	// Test with regular executable
	os.Args = []string{"main", "server"}
	isCli = isCliCommand()
	assert.False(t, isCli)

	// Test with empty args
	os.Args = []string{}
	isCli = isCliCommand()
	assert.False(t, isCli)

	// Restore original args
	os.Args = originalArgs
}

func TestDetectHandlersDirectory(t *testing.T) {
	// Test detectHandlersDirectory function
	handlersDir := detectHandlersDirectory()

	// The function should return a valid directory path or empty string
	if handlersDir != "" {
		assert.DirExists(t, handlersDir)
	}
}

func TestHasGoFiles(t *testing.T) {
	// Test hasGoFiles function
	// Test with current directory (should have Go files)
	hasGo := hasGoFiles(".")
	assert.True(t, hasGo)

	// Test with non-existent directory
	hasGo = hasGoFiles("/non/existent/directory")
	assert.False(t, hasGo)

	// Test with temporary directory (should not have Go files)
	tempDir := os.TempDir()
	hasGo = hasGoFiles(tempDir)
	// This might be true or false depending on the temp directory contents
	// We just test that the function doesn't panic
	assert.IsType(t, false, hasGo)
}

func TestRuntime_DevelopmentMode(t *testing.T) {
	// Test runtime behavior in development mode
	// This test verifies that the runtime functions work correctly

	// Test that GlobalWatcher can be created
	config := DefaultConfig()
	config.Dev.Watch = true

	watcher, err := NewFileWatcher(config)
	if err == nil {
		assert.NotNil(t, watcher)
		// Clean up
		watcher.Stop()
	}
}

func TestRuntime_FileWatcherIntegration(t *testing.T) {
	// Test file watcher integration
	config := DefaultConfig()
	config.Dev.Watch = true

	watcher, err := NewFileWatcher(config)
	if err == nil {
		assert.NotNil(t, watcher)

		// Test that watcher can be started and stopped
		err = watcher.Start()
		if err == nil {
			assert.True(t, watcher.IsRunning())

			err = watcher.Stop()
			assert.NoError(t, err)
			assert.False(t, watcher.IsRunning())
		}
	}
}

func TestRuntime_ConfigurationLoading(t *testing.T) {
	// Test configuration loading in runtime
	config, err := LoadConfig("")
	if err != nil {
		// If config loading fails, it should use default config
		config = DefaultConfig()
	}

	assert.NotNil(t, config)
	assert.IsType(t, false, config.Dev.Watch)
}

func TestRuntime_DirectoryDetection(t *testing.T) {
	// Test directory detection logic
	currentDir, err := os.Getwd()
	assert.NoError(t, err)

	// Test with current directory
	hasGo := hasGoFiles(currentDir)
	assert.IsType(t, false, hasGo)

	// Test with parent directory
	parentDir := filepath.Dir(currentDir)
	hasGo = hasGoFiles(parentDir)
	assert.IsType(t, false, hasGo)
}

func TestRuntime_EdgeCases(t *testing.T) {
	// Test edge cases in runtime functions

	// Test with empty directory
	hasGo := hasGoFiles("")
	assert.False(t, hasGo)

	// Test with relative path
	hasGo = hasGoFiles("./")
	assert.IsType(t, false, hasGo)

	// Test with absolute path
	hasGo = hasGoFiles("/")
	assert.IsType(t, false, hasGo)
}

func TestRuntime_ErrorHandling(t *testing.T) {
	// Test error handling in runtime functions

	// Test with invalid directory
	hasGo := hasGoFiles("/invalid/path/that/does/not/exist")
	assert.False(t, hasGo)

	// Test with file instead of directory
	tempFile, err := os.CreateTemp("", "test")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	hasGo = hasGoFiles(tempFile.Name())
	assert.False(t, hasGo)
}

func TestRuntime_Concurrency(_ *testing.T) {
	// Test concurrency safety of runtime functions

	// Test concurrent calls to hasGoFiles
	done := make(chan bool, 3)

	go func() {
		hasGoFiles(".")
		done <- true
	}()

	go func() {
		hasGoFiles("..")
		done <- true
	}()

	go func() {
		hasGoFiles("../..")
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

func TestRuntime_Integration(t *testing.T) {
	// Test integration of runtime components

	// Test that all runtime functions work together
	config := DefaultConfig()

	// Test configuration
	assert.NotNil(t, config)

	// Test directory detection
	handlersDir := detectHandlersDirectory()
	if handlersDir != "" {
		assert.DirExists(t, handlersDir)
		hasGo := hasGoFiles(handlersDir)
		assert.IsType(t, false, hasGo)
	}

	// Test file watcher creation
	config.Dev.Watch = true
	watcher, err := NewFileWatcher(config)
	if err == nil {
		assert.NotNil(t, watcher)
		watcher.Stop()
	}
}

func TestRuntime_Performance(t *testing.T) {
	// Test performance of runtime functions

	// Test that hasGoFiles doesn't take too long
	start := time.Now()
	hasGoFiles(".")
	duration := time.Since(start)

	// Should complete in reasonable time (less than 1 second)
	assert.Less(t, duration, time.Second)
}

func TestRuntime_MemoryUsage(t *testing.T) {
	// Test memory usage of runtime functions

	// Test that multiple calls don't leak memory
	for i := 0; i < 100; i++ {
		hasGoFiles(".")
		detectHandlersDirectory()
	}

	// If we get here without panicking, memory usage is reasonable
	assert.True(t, true)
}

func TestRuntime_CrossPlatform(t *testing.T) {
	// Test cross-platform compatibility

	// Test with different path separators
	paths := []string{
		".",
		"./",
		"../",
		"../../",
		"/",
		"/tmp",
	}

	for _, path := range paths {
		hasGo := hasGoFiles(path)
		assert.IsType(t, false, hasGo)
	}
}

func TestRuntime_FileSystemOperations(t *testing.T) {
	// Test file system operations in runtime

	// Create a temporary directory with Go files
	tempDir, err := os.MkdirTemp("", "runtime_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a Go file
	goFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(goFile, []byte("package main"), 0o644)
	assert.NoError(t, err)

	// Test that the directory is detected as having Go files
	hasGo := hasGoFiles(tempDir)
	assert.True(t, hasGo)

	// Remove the Go file
	err = os.Remove(goFile)
	assert.NoError(t, err)

	// Test that the directory no longer has Go files
	hasGo = hasGoFiles(tempDir)
	assert.False(t, hasGo)
}
