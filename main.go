package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// NewSingleHostReverseProxyWithRewrite creates a reverse proxy with path rewriting
func NewSingleHostReverseProxyWithRewrite(target *url.URL, pathPrefix string) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = strings.TrimPrefix(req.URL.Path, pathPrefix)
		req.Host = target.Host // This preserves the original Host header
	}
	return &httputil.ReverseProxy{Director: director}
}

func TokenAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.GetHeader("clientToken")
		adminToken := c.GetHeader("adminToken")

		// Check if the tokens are present
		if clientToken == "" || adminToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API token required"})
			return
		}

		// Additional token validation logic can be added here

		c.Next() // Proceed to the next handler if validation is successful
	}
}

func main() {
	router := gin.Default()

	// Apply the middleware globally
	router.Use(TokenAuthMiddleware())

	// Group routes under /gateway/route
	gatewayRoutes := router.Group("/gateway/route")

	// Setup for the identity service
	identityServiceURL, _ := url.Parse("http://localhost:8000")
	identityProxy := NewSingleHostReverseProxyWithRewrite(identityServiceURL, "/gateway/route/identity")
	gatewayRoutes.Any("/identity/*any", gin.WrapH(identityProxy))

	// Setup for the content service
	contentServiceURL, _ := url.Parse("http://localhost:9000")
	contentProxy := NewSingleHostReverseProxyWithRewrite(contentServiceURL, "/gateway/route/content")
	gatewayRoutes.Any("/content/*any", gin.WrapH(contentProxy))

	// Start the Gin server
	router.Run(":8080")
}
