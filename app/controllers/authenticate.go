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
	invalidCredError := utils.AppError{"invalid email/password", 3, nil}

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
		c.AbortWithError(http.StatusForbidden, invalidCredError)
		return
	default:
		c.AbortWithError(http.StatusBadRequest, utils.AppError{"Could not parse request body", 4, nil})
		return
	}
	email, password := params.Email, params.Password

	user := new(models.User)
	hashedPassword := models.HashPassword(password)
	err = db.QueryRow(
		"SELECT id, email, name FROM users WHERE email = $1 AND password = $2",
		email, hashedPassword).Scan(&user.Id, &user.Email, &user.Name)
	if err == sql.ErrNoRows {
		c.AbortWithError(http.StatusForbidden, invalidCredError)
		return
	} else if err != nil {
		utils.Check(err)
	}

	randBytes := make([]byte, 18)
	_, err = rand.Read(randBytes)
	utils.Check(err)
	authKey := base64.URLEncoding.EncodeToString(randBytes)

	res, err := db.Exec(
		"INSERT INTO auth_keys (auth_key, user_id, created, accessed) "+
			"VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		authKey, user.Id)
	utils.Check(err)
	rowCount, err := res.RowsAffected()
	utils.Check(err)
	if rowCount == 0 {
		c.AbortWithError(http.StatusInternalServerError,
			utils.AppError{"could not authenticate", 4, nil})
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
