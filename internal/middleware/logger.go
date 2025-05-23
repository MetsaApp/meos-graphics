package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"meos-graphics/internal/logger"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.InfoLogger.Printf("%s %s %d %v %s",
			clientIP,
			method,
			statusCode,
			latency,
			path,
		)
	}
}
