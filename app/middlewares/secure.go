package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

func Secure(allowedHosts []string, isDevelopment bool) gin.HandlerFunc {
	secureMiddleware := secure.New(secure.Options{
		AllowedHosts:    allowedHosts,
		SSLRedirect:     true,
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
		FrameDeny:       true,
		IsDevelopment:   isDevelopment,
	})

	return func(c *gin.Context) {
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			c.Abort()
			return
		}

		// Avoid header rewrite if response is a redirection.
		if status := c.Writer.Status(); status > 300 && status < 399 {
			c.Abort()
		}
	}
}
