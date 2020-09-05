package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"neal-chat/utils"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Errors != nil && len(c.Errors) > 0 {
			err := c.Errors[0].Err
			switch err.(type) {
			case utils.AppError:
				appErr := err.(utils.AppError)
				c.JSON(c.Writer.Status(), utils.ErrorResponse(appErr.Message, appErr.Code, appErr.Data))
			//case FailResponse:
			//c.JSON(err.HttpStatus, failResponse(err.Data))
			default:
				c.JSON(http.StatusInternalServerError,
					utils.ErrorResponse("500: An internal error was encountered.", 1, nil))
			}
		}
	}
}
