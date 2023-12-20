package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// TrustProxyHeader middleware to extract the client IP from the X-Forwarded-For header
func TrustProxyHeader() gin.HandlerFunc {
	return func(c *gin.Context) {
		xForwardedFor := c.GetHeader("X-Forwarded-For")
		if xForwardedFor != "" {
			clientIP := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
			if strings.Contains(clientIP, ":") {
				clientIP = strings.Split(clientIP, ":")[0]
			}
			c.Set("ClientIP", clientIP)

			port := strings.Split(c.Request.RemoteAddr, ":")[1]
			c.Request.RemoteAddr = fmt.Sprintf("%s:%s", clientIP, port)
		} else {
			c.Set("ClientIP", c.ClientIP())
		}
		c.Next()
	}
}