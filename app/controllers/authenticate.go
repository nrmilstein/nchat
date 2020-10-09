package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"gorm.io/gorm"

	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
)

func PostAuthenticate(c *gin.Context) {
	db := db.GetDb()
	invalidCredError := utils.AppError{"Invalid email/password.", 1, nil}

	var params struct {
		Email    string `json:"email" binding:"required"`
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

	email, password := strings.ToLower(params.Email), params.Password
	hashedPassword := models.HashPassword(password)

	var user models.User
	readUserResult := db.Take(&user, &models.User{Email: email, Password: hashedPassword})

	if errors.Is(readUserResult.Error, gorm.ErrRecordNotFound) {
		c.AbortWithError(http.StatusUnauthorized, invalidCredError)
		return
	} else if readUserResult.Error != nil {
		utils.AbortErrServer(c)
		return
	}

	randBytes := make([]byte, 18)
	_, err = rand.Read(randBytes)
	if err != nil {
		utils.AbortErrServer(c)
		return
	}
	authKey := base64.URLEncoding.EncodeToString(randBytes)

	session := models.Session{
		Key:    authKey,
		UserID: user.ID,
	}
	createUserResult := db.Create(&session)

	if createUserResult.Error != nil {
		utils.AbortErrServer(c)
		return
	}
	if createUserResult.RowsAffected == 0 {
		utils.AbortErrServer(c)
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{
		"authKey": authKey,
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
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
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(userJson))
}
