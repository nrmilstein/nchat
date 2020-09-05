package middlewares

import (
	"github.com/gin-gonic/gin"
)

func JSONContentType() gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
    c.Next()
  }
}
