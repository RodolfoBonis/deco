package decorators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasDecoratorAnnotations(t *testing.T) {
	// Test with decorator annotations
	comment := "// @Route(\"GET\", \"/users\")"
	assert.True(t, hasDecoratorAnnotations(comment))

	// Test without decorator annotations
	comment = "// This is a regular comment"
	assert.False(t, hasDecoratorAnnotations(comment))
}

func TestContains(t *testing.T) {
	// Test contains function
	slice := []string{"a", "b", "c"}
	assert.True(t, contains(slice, "a"))
	assert.True(t, contains(slice, "b"))
	assert.False(t, contains(slice, "d"))
}

func TestParseArguments(t *testing.T) {
	// Test parsing arguments
	args := parseArguments(`name="User", description="User entity"`)
	assert.Contains(t, args, "name=\"User\"")
	assert.Contains(t, args, "description=\"User entity\"")
}

func TestParseArgsToMap(t *testing.T) {
	// Test parsing arguments to map
	args := []string{"name=User", "description=User entity"}
	result := parseArgsToMap(args)
	assert.Equal(t, "User", result["name"])
	assert.Equal(t, "User entity", result["description"])
}

func TestParseParameterInfo(t *testing.T) {
	// Test parsing parameter info
	args := []string{"name=id", "type=int", "required=true"}
	param := parseParameterInfo(args)
	assert.Equal(t, "id", param.Name)
	assert.Equal(t, "int", param.Type)
	assert.True(t, param.Required)
}

func TestParseResponseInfo(t *testing.T) {
	// Test parsing response info
	args := []string{"code=200", "type=User"}
	response := parseResponseInfo(args)
	assert.Equal(t, "200", response.Code)
	assert.Equal(t, "User", response.Type)
}

func TestGetMiddlewareDescription(t *testing.T) {
	// Test getting middleware description
	descriptions := map[string]string{
		"Cache":     "cache",
		"RateLimit": "limitação",
		"Auth":      "autenticação",
	}

	for name, expected := range descriptions {
		desc := getMiddlewareDescription(name)
		assert.Contains(t, desc, expected)
	}
}

func TestGenerateMiddlewareCall(t *testing.T) {
	// Test generating middleware call
	marker := MarkerInstance{Name: "Cache", Args: []string{"ttl=5m"}}
	call := generateMiddlewareCall(marker)
	assert.Contains(t, call, "Cache")
	assert.Contains(t, call, "ttl=5m")
}
