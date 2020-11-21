package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"

	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/utils"
)

func PostAuthenticate(c *gin.Context) {
	invalidCredError := utils.AppError{"Invalid username/password.", 1, nil}

	var params struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	err := c.ShouldBindJSON(&params)
	switch err.(type) {
	case nil:
	case *json.SyntaxError:
		c.AbortWithError(http.StatusBadRequest, utils.AppError{"JSON syntax error", 2, nil})
		return
	case validator.ValidationErrors:
		c.AbortWithError(http.StatusUnauthorized, invalidCredError)
		return
	default:
		c.AbortWithError(http.StatusBadRequest,
			utils.AppError{"Could not parse request body", 3, nil})
		return
	}

	username, password := strings.ToLower(params.Username), params.Password
	session, user, err := models.CreateSession(username, password)

	if err != nil {
		if errors.Is(err, models.ErrInvalidCred) {
			c.AbortWithError(http.StatusUnauthorized, invalidCredError)
		} else {
			utils.AbortErrServer(c)
		}
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{
		"authKey": session.Key,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
		},
	}))
}

func GetAuthenticate(c *gin.Context) {
	user, err := models.GetUserFromRequest(c)
	if err != nil {
		utils.AbortErrForbidden(c)
		return
	}

	userJson := gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"name":     user.Name,
		},
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(userJson))
}
