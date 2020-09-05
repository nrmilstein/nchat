package controllers

import (
	"net/http"
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"

	"neal-chat/app/models"
	"neal-chat/utils"
	"neal-chat/db"
)

func PostUsers(c *gin.Context) {
	db := db.GetDb()

	var params struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&params)
	utils.Check(err)
	email, password := params.Email, params.Password

	var id int
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)
	if err != sql.ErrNoRows {
		utils.Check(err)
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hashedPassword := models.HashPassword(password)

	var userId int
	var created time.Time
	err = db.QueryRow(
		"INSERT INTO users(email, password, created) "+
			"VALUES($1, $2, CURRENT_TIMESTAMP) "+
			"RETURNING users.id, users.created",
		email,
		hashedPassword,
	).Scan(&userId, &created)
	utils.Check(err)
	addedUser := models.User{userId, email, "", created}
	c.JSON(http.StatusCreated, gin.H{"success": "user added", "user": addedUser})
}
