package handlers

import "github.com/gin-gonic/gin"

// Simple proxy example
// @Route("GET", "/api/test-proxy")
// @Proxy(target="http://httpbin.org", path="/get")
func TestProxy(c *gin.Context) {
	// Simple proxy to httpbin.org
}
