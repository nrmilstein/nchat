package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nrmilstein/nchat/utils"
)

func NoRoute(c *gin.Context) {
	c.AbortWithError(http.StatusNotFound, utils.AppError{"Resource not found.", -404, nil})
}
