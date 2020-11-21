package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AbortErrServer(c *gin.Context) {
	c.AbortWithError(http.StatusInternalServerError, ErrInternalServer)
}

func AbortErrForbidden(c *gin.Context) {
	c.AbortWithError(http.StatusForbidden, ErrForbidden)
}

//type FailResponse struct {
//HttpStatus int
//Data interface{}
//}

//func (e FailResponse) Error() string {
//return e.Message
//}

func Check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
