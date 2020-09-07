package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var ErrForbidden = AppError{"Forbidden", -403, nil}
var ErrInternalServer = AppError{"Internal server error", -500, nil}

type AppError struct {
	Message string
	Code    int
	Data    interface{}
}

//type FailResponse struct {
//HttpStatus int
//Data interface{}
//}

func (e AppError) Error() string {
	return e.Message
}

//func (e FailResponse) Error() string {
//return e.Message
//}

func AbortErrServer(c *gin.Context) {
	c.AbortWithError(http.StatusInternalServerError, ErrInternalServer)
}

func AbortErrForbidden(c *gin.Context) {
	c.AbortWithError(http.StatusForbidden, ErrForbidden)
}

func Check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
