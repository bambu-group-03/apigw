package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

// TokenAuthMiddleware is a middleware function for token authentication
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

// getServiceURL fetches and parses a service URL from the environment
func getServiceURL(envVarName string) (*url.URL, error) {
	urlString := os.Getenv(envVarName)
	if urlString == "" {
		return nil, fmt.Errorf("%s environment variable not set", envVarName)
	}
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("invalid URL in %s: %v", envVarName, err)
	}
	return parsedURL, nil
}

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Fetch and parse service URLs
	identityServiceURL, err := getServiceURL("IDENTITY_SOCIALIZER_URL")
	if err != nil {
		log.Fatal(err)
	}
	contentServiceURL, err := getServiceURL("CONTENT_DISCOVERY_URL")
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.SetTrustedProxies(nil)
	// Apply the middleware globally
	router.Use(TokenAuthMiddleware())

	// Group routes under /gateway/route
	gatewayRoutes := router.Group("/gateway/route")

	// Setup for the identity service
	identityProxy := NewSingleHostReverseProxyWithRewrite(identityServiceURL, "/gateway/route/identity")
	gatewayRoutes.Any("/identity/*any", gin.WrapH(identityProxy))

	// Setup for the content service
	contentProxy := NewSingleHostReverseProxyWithRewrite(contentServiceURL, "/gateway/route/content")
	gatewayRoutes.Any("/content/*any", gin.WrapH(contentProxy))

	println("🔗 identityServiceURL:", identityServiceURL.String())
	println("🔗 contentServiceURL:", contentServiceURL.String())
	// Start the Gin server
	router.Run(":8080")
}
