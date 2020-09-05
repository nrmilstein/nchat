package utils

import (
	"log"
	
	"github.com/gin-gonic/gin"
)

type AppError struct {
  Message string
  Code int
  Data interface{}
}

//type FailResponse struct {
  //HttpStatus int
  //Data interface{}
//}

func (e AppError ) Error() string {
  return e.Message
}

//func (e FailResponse) Error() string {
  //return e.Message
//}


func Check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func SuccessResponse(data interface{}) gin.H {
	return gin.H{
		"status": "success",
		"data":   data,
	}
}

func FailResponse(data interface{}) gin.H {
	return gin.H{
		"status": "fail",
		"data":   data,
	}
}

func ErrorResponse(message string, code int, data interface{}) gin.H {
	response := gin.H{
		"status":  "error",
		"message": message,
		"code":    code,
	}
	if data != nil {
		response["data"] = data
	}
	return response
}
