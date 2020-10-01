package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"

	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
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

	var id int
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)
	if err != sql.ErrNoRows {
		if err == nil {
			c.AbortWithError(http.StatusConflict,
				utils.AppError{"Email already registered.", 4, nil})
			return
		} else {
			utils.AbortErrServer(c)
			return
		}
	}

	hashedPassword := models.HashPassword(password)

	var (
		newUserId int
		created   time.Time
	)
	err = db.QueryRow(
		"INSERT INTO users(email, password, name, created) "+
			"VALUES($1, $2, $3, CURRENT_TIMESTAMP) "+
			"RETURNING users.id, users.created",
		email,
		hashedPassword,
		name,
	).Scan(&newUserId, &created)
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	newUserJson := gin.H{
		"id":      newUserId,
		"email":   email,
		"name:":   name,
		"created": created,
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{"user": newUserJson}))
}
