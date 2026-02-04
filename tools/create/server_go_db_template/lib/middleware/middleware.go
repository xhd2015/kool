package middleware

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/cookie"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/routehelp"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/lib/trace"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/service/session"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/service/user"
	"github.com/xhd2015/kool/tools/create/server_go_db_template/types"
)

// Trace injects a trace ID into the request context
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if trace ID is provided in header
		traceID := trace.TraceID(c.GetHeader("X-Trace-ID"))
		if traceID == "" {
			traceID = trace.NewTraceID()
		}

		// Set trace ID in context
		ctx := trace.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(ctx)

		// Set trace ID in response header
		c.Header("X-Trace-ID", traceID.String())

		c.Next()
	}
}

// Recovery returns a custom recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.RecoveryWithWriter(gin.DefaultErrorWriter, func(c *gin.Context, err any) {
		c.AbortWithStatus(http.StatusInternalServerError)
		c.Writer.WriteString(fmt.Sprintf(`{"code":%d, "msg":%q}`, 1, fmt.Sprintf("panic: %v", err)))
	})
}

// CORS returns a CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Trace-ID")
		c.Header("Access-Control-Expose-Headers", "X-Trace-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// StaticFiles serves static files for non-API requests
func StaticFiles(staticDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := c.Request.URL.RequestURI()

		// Only handle non-API requests
		if strings.HasPrefix(uri, "/api/") {
			c.Next()
			return
		}

		// Serve static files
		possibleFile := filepath.Join(staticDir, uri)
		stat, statErr := os.Stat(possibleFile)
		if statErr != nil || stat.IsDir() {
			possibleFile = filepath.Join(staticDir, "index.html")
		}
		c.Header("Cache-Control", "no-cache, must-revalidate")
		c.File(possibleFile)
		c.Abort()
	}
}

// AuthConfig holds configuration for the Auth middleware
type AuthConfig struct {
	// PublicPaths are paths that don't require authentication
	PublicPaths []string
}

// Auth handles authentication
func Auth(cfg AuthConfig) gin.HandlerFunc {
	publicPaths := make(map[string]bool)
	for _, path := range cfg.PublicPaths {
		publicPaths[path] = true
	}

	return func(c *gin.Context) {
		uriPath := c.Request.URL.Path

		// Skip auth for public endpoints
		if publicPaths[uriPath] {
			c.Next()
			return
		}

		// Extract user ID from token
		var userID types.UserID

		cookieHeader := c.Request.Header.Get("Cookie")
		token := cookie.GetToken(cookieHeader)
		if token == "" {
			// Try Authorization header
			authHeader := c.Request.Header.Get("Authorization")
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}

		if token != "" {
			// Validate token using user service
			var err error
			userID, err = user.QueryUserByAuthToken(c, token)
			if err != nil {
				routehelp.AbortWithErrCode(c, http.StatusUnauthorized, err)
				return
			}
		}

		if userID <= 0 {
			routehelp.AbortWithErrCode(c, http.StatusUnauthorized, fmt.Errorf("invalid user id"))
			return
		}

		// Set session
		session.SetGin(c, &session.Session{
			UserID: userID,
		})

		c.Next()
	}
}
