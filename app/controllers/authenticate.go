package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"

	"neal-chat/app/models"
	"neal-chat/db"
	"neal-chat/utils"
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
	case validator.ValidationErrors: // TODO: make this case work
		c.AbortWithError(http.StatusUnauthorized, invalidCredError)
		return
	default:
		c.AbortWithError(http.StatusBadRequest,
			utils.AppError{"Could not parse request body", 3, nil})
		return
	}

	email, password := params.Email, params.Password

	user := new(models.User)
	hashedPassword := models.HashPassword(password)

	err = db.QueryRow(
		"SELECT id, email, name FROM users WHERE email = $1 AND password = $2",
		email, hashedPassword).Scan(&user.Id, &user.Email, &user.Name)
	if err == sql.ErrNoRows {
		c.AbortWithError(http.StatusUnauthorized, invalidCredError)
		return
	} else if err != nil {
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

	res, err := db.Exec(
		"INSERT INTO auth_keys (auth_key, user_id, created, accessed) "+
			"VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		authKey, user.Id)
	if err != nil {
		utils.AbortErrServer(c)
		return
	}
	rowCount, err := res.RowsAffected()
	if err != nil || rowCount == 0 {
		utils.AbortErrServer(c)
		return
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{
		"auth_key": authKey,
		"user": gin.H{
			"id":    user.Id,
			"email": user.Email,
			"name":  user.Name,
		},
	}))
}
