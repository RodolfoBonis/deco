package decorators

import (
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

var (
	ginModeMutex sync.Mutex
	ginModeSet   bool
)

// setupGinTestMode sets Gin to test mode in a thread-safe way
func setupGinTestMode(_ *testing.T) {
	ginModeMutex.Lock()
	defer ginModeMutex.Unlock()

	if !ginModeSet {
		gin.SetMode(gin.TestMode)
		ginModeSet = true
	}
}

// createTestGinContext creates a test Gin context with proper setup
// This function is kept for future use in more complex test scenarios

// createTestGinEngine creates a test Gin engine with proper setup
func createTestGinEngine(t *testing.T) *gin.Engine {
	setupGinTestMode(t)
	return gin.New()
}
