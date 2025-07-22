package decorators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterParserHook(t *testing.T) {
	// Clear existing hooks
	parserHooks = nil

	// Test registering a parser hook
	hook := func(_ []*RouteMeta) error {
		return nil
	}

	RegisterParserHook(hook)
	assert.Len(t, parserHooks, 1)
	// Cannot compare functions directly in Go
	assert.NotNil(t, parserHooks[0])
}

func TestRegisterGeneratorHook(t *testing.T) {
	// Clear existing hooks
	generatorHooks = nil

	// Test registering a generator hook
	hook := func(_ *GenData) error {
		return nil
	}

	RegisterGeneratorHook(hook)
	assert.Len(t, generatorHooks, 1)
	// Cannot compare functions directly in Go
	assert.NotNil(t, generatorHooks[0])
}

func TestExecuteParserHooks(t *testing.T) {
	// Clear existing hooks
	parserHooks = nil

	// Test executing parser hooks
	hook1Called := false
	hook2Called := false

	hook1 := func(_ []*RouteMeta) error {
		hook1Called = true
		return nil
	}

	hook2 := func(_ []*RouteMeta) error {
		hook2Called = true
		return nil
	}

	RegisterParserHook(hook1)
	RegisterParserHook(hook2)

	routes := []*RouteMeta{
		{Method: "GET", Path: "/test", FuncName: "TestHandler"},
	}

	err := executeParserHooks(routes)
	assert.NoError(t, err)
	assert.True(t, hook1Called)
	assert.True(t, hook2Called)
}

func TestExecuteParserHooks_Error(t *testing.T) {
	// Clear existing hooks
	parserHooks = nil

	// Test executing parser hooks with error
	hook := func(_ []*RouteMeta) error {
		return assert.AnError
	}

	RegisterParserHook(hook)

	routes := []*RouteMeta{
		{Method: "GET", Path: "/test", FuncName: "TestHandler"},
	}

	err := executeParserHooks(routes)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestExecuteGeneratorHooks(t *testing.T) {
	// Clear existing hooks
	generatorHooks = nil

	// Test executing generator hooks
	hook1Called := false
	hook2Called := false

	hook1 := func(_ *GenData) error {
		hook1Called = true
		return nil
	}

	hook2 := func(_ *GenData) error {
		hook2Called = true
		return nil
	}

	RegisterGeneratorHook(hook1)
	RegisterGeneratorHook(hook2)

	data := &GenData{
		PackageName: "test",
		Routes: []*RouteMeta{
			{Method: "GET", Path: "/test", FuncName: "TestHandler"},
		},
	}

	err := executeGeneratorHooks(data)
	assert.NoError(t, err)
	assert.True(t, hook1Called)
	assert.True(t, hook2Called)
}

func TestExecuteGeneratorHooks_Error(t *testing.T) {
	// Clear existing hooks
	generatorHooks = nil

	// Test executing generator hooks with error
	hook := func(_ *GenData) error {
		return assert.AnError
	}

	RegisterGeneratorHook(hook)

	data := &GenData{
		PackageName: "test",
		Routes: []*RouteMeta{
			{Method: "GET", Path: "/test", FuncName: "TestHandler"},
		},
	}

	err := executeGeneratorHooks(data)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestGetParserHooks(t *testing.T) {
	// Clear existing hooks
	parserHooks = nil

	// Test getting parser hooks
	hook := func(_ []*RouteMeta) error {
		return nil
	}

	RegisterParserHook(hook)
	hooks := GetParserHooks()
	assert.Len(t, hooks, 1)
	// Cannot compare functions directly in Go
	assert.NotNil(t, hooks[0])
}

func TestGetGeneratorHooks(t *testing.T) {
	// Clear existing hooks
	generatorHooks = nil

	// Test getting generator hooks
	hook := func(_ *GenData) error {
		return nil
	}

	RegisterGeneratorHook(hook)
	hooks := GetGeneratorHooks()
	assert.Len(t, hooks, 1)
	// Cannot compare functions directly in Go
	assert.NotNil(t, hooks[0])
}

func TestGetRequiredImports(t *testing.T) {
	// Test getting required imports
	data := &GenData{
		Routes: []*RouteMeta{
			{Method: "GET", Path: "/test", FuncName: "TestHandler"},
		},
	}

	imports := getRequiredImports(data)
	assert.NotNil(t, imports)
	assert.Contains(t, imports, "deco \"github.com/RodolfoBonis/deco\"")
}

func TestShouldAddHandlersImport(t *testing.T) {
	// Test should add handlers import
	data := &GenData{
		Routes: []*RouteMeta{
			{Method: "GET", Path: "/test", FuncName: "TestHandler"},
		},
	}

	shouldAdd := shouldAddHandlersImport(data)
	assert.IsType(t, false, shouldAdd)
}

func TestBuildHandlersImport(t *testing.T) {
	// Test building handlers import
	importPath := buildHandlersImport()
	assert.Contains(t, importPath, "handlers")
}

func TestBuildImportPath(t *testing.T) {
	// Test building import path
	path := buildImportPath("test", "handler")
	assert.Contains(t, path, "handler")
}

func TestAddMissingImports(t *testing.T) {
	// Test adding missing imports
	data := &GenData{
		Imports: []string{"fmt"},
	}

	requiredImports := []string{"os", "net/http"}

	addMissingImports(data, requiredImports)
	assert.Contains(t, data.Imports, "fmt")
	assert.Contains(t, data.Imports, "os")
	assert.Contains(t, data.Imports, "net/http")
}

func TestContainsImport(t *testing.T) {
	// Test contains import
	imports := []string{"fmt", "os", "net/http"}

	assert.True(t, containsImport(imports, "fmt"))
	assert.True(t, containsImport(imports, "os"))
	assert.False(t, containsImport(imports, "encoding/json"))
}

func TestGetModuleName(t *testing.T) {
	// Test getting module name
	moduleName := getModuleName(".")
	// Just test that it doesn't panic
	assert.IsType(t, "", moduleName)
}

func TestGenData_Structure(t *testing.T) {
	// Test GenData structure
	data := &GenData{
		PackageName: "test",
		Routes: []*RouteMeta{
			{Method: "GET", Path: "/test", FuncName: "TestHandler"},
		},
		Imports:     []string{"import1"},
		Metadata:    map[string]interface{}{"key": "value"},
		GeneratedAt: "2023-01-01",
	}

	assert.Equal(t, "test", data.PackageName)
	assert.Len(t, data.Routes, 1)
	assert.Len(t, data.Imports, 1)
	assert.Len(t, data.Metadata, 1)
	assert.Equal(t, "2023-01-01", data.GeneratedAt)
}

func TestRouteMeta_Structure(t *testing.T) {
	// Test RouteMeta structure
	route := &RouteMeta{
		Method:            "GET",
		Path:              "/test",
		FuncName:          "TestHandler",
		PackageName:       "test",
		FileName:          "test.go",
		Markers:           []MarkerInstance{},
		MiddlewareCalls:   []string{},
		Description:       "Test route",
		Summary:           "Test summary",
		Tags:              []string{"test"},
		MiddlewareInfo:    []MiddlewareInfo{},
		Parameters:        []ParameterInfo{},
		Group:             &GroupInfo{},
		Responses:         []ResponseInfo{},
		WebSocketHandlers: []string{},
	}

	assert.Equal(t, "GET", route.Method)
	assert.Equal(t, "/test", route.Path)
	assert.Equal(t, "TestHandler", route.FuncName)
	assert.Equal(t, "test", route.PackageName)
	assert.Equal(t, "test.go", route.FileName)
	assert.Equal(t, "Test route", route.Description)
	assert.Equal(t, "Test summary", route.Summary)
	assert.Len(t, route.Tags, 1)
	assert.Len(t, route.Markers, 0)
	assert.Len(t, route.MiddlewareCalls, 0)
}

func TestMarkerInstance_Structure(t *testing.T) {
	// Test MarkerInstance structure
	marker := MarkerInstance{
		Name: "Auth",
		Args: []string{"required"},
		Raw:  "// @Auth(required)",
	}

	assert.Equal(t, "Auth", marker.Name)
	assert.Len(t, marker.Args, 1)
	assert.Equal(t, "required", marker.Args[0])
	assert.Equal(t, "// @Auth(required)", marker.Raw)
}

func TestPluginHooks_Integration(t *testing.T) {
	// Test plugin hooks integration
	parserHooks = nil
	generatorHooks = nil

	// Register hooks
	parserCalled := false
	generatorCalled := false

	parserHook := func(_ []*RouteMeta) error {
		parserCalled = true
		return nil
	}

	generatorHook := func(_ *GenData) error {
		generatorCalled = true
		return nil
	}

	RegisterParserHook(parserHook)
	RegisterGeneratorHook(generatorHook)

	// Execute hooks
	routes := []*RouteMeta{
		{Method: "GET", Path: "/test", FuncName: "TestHandler"},
	}

	data := &GenData{
		PackageName: "test",
		Routes:      routes,
	}

	err1 := executeParserHooks(routes)
	err2 := executeGeneratorHooks(data)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.True(t, parserCalled)
	assert.True(t, generatorCalled)
}

func TestPluginHooks_ErrorHandling(t *testing.T) {
	// Test plugin hooks error handling
	parserHooks = nil
	generatorHooks = nil

	// Register error hooks
	parserHook := func(_ []*RouteMeta) error {
		return assert.AnError
	}

	generatorHook := func(_ *GenData) error {
		return assert.AnError
	}

	RegisterParserHook(parserHook)
	RegisterGeneratorHook(generatorHook)

	// Execute hooks
	routes := []*RouteMeta{
		{Method: "GET", Path: "/test", FuncName: "TestHandler"},
	}

	data := &GenData{
		PackageName: "test",
		Routes:      routes,
	}

	err1 := executeParserHooks(routes)
	err2 := executeGeneratorHooks(data)

	assert.Error(t, err1)
	assert.Error(t, err2)
	assert.Equal(t, assert.AnError, err1)
	assert.Equal(t, assert.AnError, err2)
}
