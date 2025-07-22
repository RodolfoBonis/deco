// Tests for logging logic in gin-decorators framework
package decorators

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogLevel_Constants(t *testing.T) {
	// Test log level constants
	assert.Equal(t, LogLevel(0), LogLevelSilent)
	assert.Equal(t, LogLevel(1), LogLevelNormal)
	assert.Equal(t, LogLevel(2), LogLevelVerbose)

	// Test ordering
	assert.True(t, LogLevelSilent < LogLevelNormal)
	assert.True(t, LogLevelNormal < LogLevelVerbose)
}

func TestSetLogLevel(t *testing.T) {
	// Test setting different log levels
	SetLogLevel(LogLevelSilent)
	assert.Equal(t, LogLevelSilent, GetLogLevel())

	SetLogLevel(LogLevelNormal)
	assert.Equal(t, LogLevelNormal, GetLogLevel())

	SetLogLevel(LogLevelVerbose)
	assert.Equal(t, LogLevelVerbose, GetLogLevel())
}

func TestGetLogLevel(t *testing.T) {
	// Remove  to avoid race conditions

	// Save current level
	originalLevel := GetLogLevel()

	// Test setting and getting level
	SetLogLevel(LogLevelSilent)
	level := GetLogLevel()
	assert.Equal(t, LogLevelSilent, level)

	// Test another level
	SetLogLevel(LogLevelVerbose)
	level = GetLogLevel()
	assert.Equal(t, LogLevelVerbose, level)

	// Test normal level
	SetLogLevel(LogLevelNormal)
	level = GetLogLevel()
	assert.Equal(t, LogLevelNormal, level)

	// Restore original level
	SetLogLevel(originalLevel)
}

func TestSetVerbose(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Test enabling verbose mode
	SetVerbose(true)
	assert.Equal(t, LogLevelVerbose, GetLogLevel())

	// Test disabling verbose mode
	SetVerbose(false)
	assert.Equal(t, LogLevelNormal, GetLogLevel())
}

func TestLogVerbose_WithVerboseEnabled(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Enable verbose logging
	SetLogLevel(LogLevelVerbose)

	// LogVerbose should work when verbose is enabled
	// Note: We can't easily test the actual output without capturing stdout
	// But we can test that the function doesn't panic
	assert.NotPanics(t, func() {
		LogVerbose("Test verbose message: %s", "test")
	})
}

func TestLogVerbose_WithVerboseDisabled(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Disable verbose logging
	SetLogLevel(LogLevelNormal)

	// LogVerbose should not output when verbose is disabled
	// But the function should not panic
	assert.NotPanics(t, func() {
		LogVerbose("Test verbose message: %s", "test")
	})
}

func TestLogNormal_WithNormalEnabled(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Enable normal logging
	SetLogLevel(LogLevelNormal)

	// LogNormal should work when normal logging is enabled
	assert.NotPanics(t, func() {
		LogNormal("Test normal message: %s", "test")
	})
}

func TestLogNormal_WithSilentLevel(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Set to silent level
	SetLogLevel(LogLevelSilent)

	// LogNormal should not output when silent
	// But the function should not panic
	assert.NotPanics(t, func() {
		LogNormal("Test normal message: %s", "test")
	})
}

func TestLogSilent_AlwaysWorks(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Test with different log levels
	testLevels := []LogLevel{LogLevelSilent, LogLevelNormal, LogLevelVerbose}

	for _, level := range testLevels {
		t.Run(fmt.Sprintf("level_%d", level), func(t *testing.T) {
			SetLogLevel(level)

			// LogSilent should always work regardless of level
			assert.NotPanics(t, func() {
				LogSilent("Test silent message: %s", "test")
			})
		})
	}
}

func TestLogger_ThreadSafety(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Test concurrent access to logger
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			// Concurrently set and get log levels
			SetLogLevel(LogLevelVerbose)
			_ = GetLogLevel()
			SetVerbose(true)
			_ = GetLogLevel()
			SetVerbose(false)
			_ = GetLogLevel()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and maintain consistency
	assert.NotPanics(t, func() {
		level := GetLogLevel()
		assert.True(t, level >= LogLevelSilent && level <= LogLevelVerbose)
	})
}

func TestLogLevel_Comparison(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Test log level comparisons
	assert.True(t, LogLevelNormal >= LogLevelSilent)
	assert.True(t, LogLevelVerbose >= LogLevelSilent)

	assert.True(t, LogLevelVerbose >= LogLevelNormal)

	assert.False(t, LogLevelSilent >= LogLevelNormal)
	assert.False(t, LogLevelSilent >= LogLevelVerbose)
	assert.False(t, LogLevelNormal >= LogLevelVerbose)
}

func TestLogging_Integration(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Test integration of all logging functions
	SetLogLevel(LogLevelSilent)
	assert.Equal(t, LogLevelSilent, GetLogLevel())

	// Silent level - only LogSilent should work
	assert.NotPanics(t, func() {
		LogSilent("Silent message")
		LogNormal("Normal message")   // Should not output
		LogVerbose("Verbose message") // Should not output
	})

	// Normal level
	SetLogLevel(LogLevelNormal)
	assert.Equal(t, LogLevelNormal, GetLogLevel())

	assert.NotPanics(t, func() {
		LogSilent("Silent message")
		LogNormal("Normal message")   // Should work
		LogVerbose("Verbose message") // Should not output
	})

	// Verbose level
	SetLogLevel(LogLevelVerbose)
	assert.Equal(t, LogLevelVerbose, GetLogLevel())

	assert.NotPanics(t, func() {
		LogSilent("Silent message")
		LogNormal("Normal message")   // Should work
		LogVerbose("Verbose message") // Should work
	})
}

func TestSetVerbose_Integration(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	// Test SetVerbose integration with logging functions
	SetVerbose(true)
	assert.Equal(t, LogLevelVerbose, GetLogLevel())

	// All logging functions should work
	assert.NotPanics(t, func() {
		LogSilent("Silent message")
		LogNormal("Normal message")
		LogVerbose("Verbose message")
	})

	SetVerbose(false)
	assert.Equal(t, LogLevelNormal, GetLogLevel())

	// Only LogSilent and LogNormal should work
	assert.NotPanics(t, func() {
		LogSilent("Silent message")
		LogNormal("Normal message")
		LogVerbose("Verbose message") // Should not output
	})
}

func TestLogging_FormatStrings(t *testing.T) {
	// Remove  to avoid race conditions with log level changes

	SetLogLevel(LogLevelVerbose)

	// Test various format strings
	testCases := []struct {
		format string
		args   []interface{}
	}{
		{"Simple message", nil},
		{"Message with %s", []interface{}{"string"}},
		{"Message with %d", []interface{}{42}},
		{"Message with %s and %d", []interface{}{"test", 123}},
		{"Message with %% literal", nil},
		{"", nil}, // Empty format
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("format_%s", tc.format), func(t *testing.T) {
			assert.NotPanics(t, func() {
				LogSilent(tc.format, tc.args...)
				LogNormal(tc.format, tc.args...)
				LogVerbose(tc.format, tc.args...)
			})
		})
	}
}
