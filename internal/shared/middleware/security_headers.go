package middleware

import "github.com/gin-gonic/gin"

// SecurityHeaders applies essential HTTP security headers to all responses.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking via iframes
		c.Header("X-Frame-Options", "DENY")

		// Control referrer information leakage
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Legacy XSS filter protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict Content Security Policy for API
		c.Header("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		c.Next()
	}
}
