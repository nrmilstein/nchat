package utils

import (
	"github.com/gin-gonic/gin"
)

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
