package server

import (
	"compress/gzip"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// compressionMiddleware adds gzip compression for responses
func (s *Server) compressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if client accepts gzip
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// Skip compression for small responses or certain content types
		if c.Writer.Size() < 1024 {
			c.Next()
			return
		}

		// Set compression headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// Create gzip writer
		gz := gzip.NewWriter(c.Writer)
		defer gz.Close()

		// Replace the writer
		c.Writer = &gzipWriter{c.Writer, gz}
		c.Next()
	}
}

// gzipWriter wraps gin.ResponseWriter with gzip compression
type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// cacheMiddleware adds cache headers for static content
func (s *Server) cacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set cache headers for static endpoints
		path := c.Request.URL.Path

		if path == "/health" || path == "/metrics" {
			// Don't cache dynamic endpoints
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		} else if strings.HasPrefix(path, "/api/") {
			// Short cache for API responses
			c.Header("Cache-Control", "public, max-age=60")
		} else {
			// Longer cache for static content
			c.Header("Cache-Control", "public, max-age=3600")
		}

		c.Next()
	}
}

// securityMiddleware adds security headers
func (s *Server) securityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Only add HSTS in production with HTTPS
		if s.config.Server.Mode == "release" && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// contentTypeMiddleware ensures proper content types
func (s *Server) contentTypeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set default content type for API endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Content-Type", "application/json; charset=utf-8")
		}

		c.Next()
	}
}

// timeoutMiddleware adds request timeout handling
func (s *Server) timeoutMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add timeout context based on endpoint
		path := c.Request.URL.Path

		var timeoutHeader string
		if strings.Contains(path, "/webhook/") {
			// Longer timeout for plugin execution
			timeoutHeader = "120"
		} else {
			// Standard timeout for other endpoints
			timeoutHeader = "30"
		}

		c.Header("X-Timeout-Seconds", timeoutHeader)
		c.Next()
	}
}

// performanceHeaders adds performance-related headers
func (s *Server) performanceHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add performance hints
		c.Header("X-DNS-Prefetch-Control", "on")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// Add server information
		c.Header("Server", "webhook-bridge/2.0")

		c.Next()
	}
}

// enablePerformanceOptimizations adds all performance middleware
func (s *Server) enablePerformanceOptimizations(router *gin.Engine) {
	// Add performance middleware
	router.Use(s.compressionMiddleware())
	router.Use(s.cacheMiddleware())
	router.Use(s.securityMiddleware())
	router.Use(s.contentTypeMiddleware())
	router.Use(s.timeoutMiddleware())
	router.Use(s.performanceHeaders())
}

// getResponseSize returns the response size from headers
func getResponseSize(c *gin.Context) int64 {
	if sizeStr := c.GetHeader("Content-Length"); sizeStr != "" {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			return size
		}
	}
	return int64(c.Writer.Size())
}
