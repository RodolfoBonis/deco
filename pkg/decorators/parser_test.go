package decorators

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyDecoratorDetection(t *testing.T) {
	// Test source code with @Proxy decorator
	source := `package handlers

import "github.com/gin-gonic/gin"

// @Route("GET", "/api/test")
// @Proxy(target="http://httpbin.org", path="/get")
func TestProxy(c *gin.Context) {
	// Test function
}
`

	// Parse the source code
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", source, parser.ParseComments)
	assert.NoError(t, err)

	// Check if decorator annotations are detected
	hasDecorators := hasDecoratorAnnotations(source)
	assert.True(t, hasDecorators, "Decorator annotations should be detected")

	// Find function declaration
	var funcDecl *ast.FuncDecl
	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			funcDecl = fd
			break
		}
	}
	assert.NotNil(t, funcDecl, "Function declaration should be found")

	// Check if Proxy decorator is detected
	markers, validationErr := extractMarkersWithValidation(fset, "test.go", funcDecl, source)
	assert.Nil(t, validationErr, "Should not have validation errors")
	assert.Len(t, markers, 1, "Should detect 1 decorator: @Proxy")

	// Check if Proxy marker is present
	proxyFound := false
	for _, marker := range markers {
		if marker.Name == "Proxy" {
			proxyFound = true
			break
		}
	}
	assert.True(t, proxyFound, "Proxy decorator should be detected")
}
