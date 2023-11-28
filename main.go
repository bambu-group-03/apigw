package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

var (
	adminOnlyRoutes = []string{
		// Add your admin-only routes here
		// "/gateway/route/identity/megahypersecret",
	}
	ctx = context.Background()
)

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

		for _, adminRoute := range adminOnlyRoutes {
			if strings.HasPrefix(c.Request.URL.Path, adminRoute) {
				if adminToken == "" {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Admin token required"})
					return
				}
				break
			}
		}

		if clientToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "API token required"})
			return
		}

		c.Next()
	}
}

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

func CacheMiddleware(rdb *redis.Client, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		cachedResponse, err := rdb.Get(ctx, path).Result()
		if err == nil {
			c.Writer.WriteString(cachedResponse)
			c.Abort()
			return
		}

		c.Next()

		response, exists := c.Get("response")
		if exists {
			rdb.Set(ctx, path, response, ttl)
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	identityServiceURL, err := getServiceURL("IDENTITY_SOCIALIZER_URL")
	if err != nil {
		log.Fatal(err)
	}
	contentServiceURL, err := getServiceURL("CONTENT_DISCOVERY_URL")
	if err != nil {
		log.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Use(TokenAuthMiddleware())
	router.Use(CacheMiddleware(rdb, 30*time.Second))

	gatewayRoutes := router.Group("/gateway/route")
	identityProxy := NewSingleHostReverseProxyWithRewrite(identityServiceURL, "/gateway/route/identity")
	gatewayRoutes.Any("/identity/*any", gin.WrapH(identityProxy))
	contentProxy := NewSingleHostReverseProxyWithRewrite(contentServiceURL, "/gateway/route/content")
	gatewayRoutes.Any("/content/*any", gin.WrapH(contentProxy))

	println("🔗 identityServiceURL:", identityServiceURL.String())
	println("🔗 contentServiceURL:", contentServiceURL.String())
	router.Run(":8080")
}
