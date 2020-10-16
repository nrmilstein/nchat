package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
	"gorm.io/gorm"
)

func PostUsers(c *gin.Context) {
	db := db.GetDb()

	var params struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Name     string `json:"name" binding:"required"`
	}

	err := c.ShouldBindJSON(&params)
	switch err.(type) {
	case nil:
	case *json.SyntaxError:
		c.AbortWithError(http.StatusBadRequest,
			utils.AppError{"JSON syntax error.", 1, nil})
		return
	case validator.ValidationErrors:
		c.AbortWithError(http.StatusUnauthorized,
			utils.AppError{"Missing parameters.", 2, nil})
		return
	default:
		c.AbortWithError(http.StatusBadRequest,
			utils.AppError{"Could not parse request body.", 3, nil})
		return
	}

	email, password, name := strings.ToLower(params.Email), params.Password, params.Name

	if strings.TrimSpace(email) == "" ||
		strings.TrimSpace(password) == "" ||
		strings.TrimSpace(name) == "" {
		c.AbortWithError(http.StatusBadRequest, utils.AppError{"Fields cannot be empty.", 5, nil})
		return
	}

	var user models.User
	readUserResult := db.Take(&user, &models.User{Email: email})
	if readUserResult.Error != gorm.ErrRecordNotFound {
		if readUserResult.Error == nil {
			c.AbortWithError(http.StatusConflict,
				utils.AppError{"Email already registered.", 6, nil})
			return
		} else {
			utils.AbortErrServer(c)
			return
		}
	}

	hashedPassword := models.HashPassword(password)

	newUser := models.User{
		Email:    email,
		Password: hashedPassword,
		Name:     name,
	}
	createUserResult := db.Create(&newUser)
	if createUserResult.Error != nil {
		utils.AbortErrServer(c)
		return
	}

	newUserJson := gin.H{
		"id":    newUser.ID,
		"email": newUser.Email,
		"name:": newUser.Name,
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{"user": newUserJson}))
}
