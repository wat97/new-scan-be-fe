package middleware

import (
	"log"
	"os"
	"scandata/config"
	"time"

	"github.com/gin-gonic/gin"
)

var securityLogger *log.Logger

func InitSecurityLogger(cfg *config.Config) {
	if !cfg.EnableSecurityLog {
		return
	}

	file, err := os.OpenFile("security.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Warning: Could not open security.log: %v", err)
		securityLogger = log.New(os.Stdout, "[SECURITY] ", log.LstdFlags)
		return
	}

	securityLogger = log.New(file, "", log.LstdFlags)
}

func LogSecurityEvent(eventType, ip, username, details string) {
	if securityLogger != nil {
		securityLogger.Printf("[%s] IP=%s User=%s Details=%s", eventType, ip, username, details)
	}
}

// SecurityLoggerMiddleware logs security-related events
func SecurityLoggerMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// Log after request
		if cfg.EnableSecurityLog && securityLogger != nil {
			latency := time.Since(start)
			status := c.Writer.Status()

			// Log failed authentication attempts
			if status == 401 || status == 403 {
				username := ""
				if u, exists := c.Get("username"); exists {
					username = u.(string)
				}

				LogSecurityEvent(
					"AUTH_FAILURE",
					c.ClientIP(),
					username,
					c.Request.URL.Path,
				)
			}

			// Log suspicious activity (too many requests)
			if status == 429 {
				LogSecurityEvent(
					"RATE_LIMIT",
					c.ClientIP(),
					"",
					c.Request.URL.Path,
				)
			}

			// Log all admin actions
			if status >= 200 && status < 300 {
				if role, exists := c.Get("role"); exists && role == "admin" {
					if c.Request.Method != "GET" {
						username := ""
						if u, exists := c.Get("username"); exists {
							username = u.(string)
						}
						LogSecurityEvent(
							"ADMIN_ACTION",
							c.ClientIP(),
							username,
							c.Request.Method+" "+c.Request.URL.Path,
						)
					}
				}
			}

			// Log slow requests (potential DoS)
			if latency > 5*time.Second {
				LogSecurityEvent(
					"SLOW_REQUEST",
					c.ClientIP(),
					"",
					c.Request.URL.Path,
				)
			}
		}
	}
}
