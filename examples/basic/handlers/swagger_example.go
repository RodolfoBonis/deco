package handlers

import (
	"github.com/gin-gonic/gin"
)

// SwaggerUIExample demonstrates how to serve Swagger UI documentation
// @Route("GET", "/swagger")
// @SwaggerUI()
// @Description("Interactive Swagger UI for API documentation and testing")
// @Summary("Swagger UI Interface")
// @Tag("Documentation")
func SwaggerUIExample(c *gin.Context) {
	// This handler will automatically serve the Swagger UI interface
	// The interface will load the OpenAPI specification from /decorators/openapi.json
	c.Header("Content-Type", "text/html")
}

// OpenAPIJSONExample demonstrates how to serve OpenAPI JSON specification
// @Route("GET", "/api-spec")
// @OpenAPIJSON()
// @Description("OpenAPI 3.0 specification in JSON format")
// @Summary("API Specification JSON")
// @Tag("Documentation")
func OpenAPIJSONExample(c *gin.Context) {
	// This handler will automatically serve the OpenAPI JSON specification
	c.Header("Content-Type", "application/json")
}

// DocumentationRedirect provides a convenient redirect to documentation
// @Route("GET", "/docs")
// @Description("Redirect to main documentation")
// @Summary("Documentation Home")
// @Tag("Documentation")
func DocumentationRedirect(c *gin.Context) {
	c.Redirect(302, "/swagger")
}
