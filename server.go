package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"

	"neal-chat/app/controllers"
	"neal-chat/app/middlewares"
	"neal-chat/db"
	"neal-chat/utils"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "5000"
	}

	psqlInfo := db.PsqlInfo{
		Host:     "localhost",
		Port:     5432,
		User:     "nrmilstein",
		Password: "password",
		Dbname:   "neal_chat",
	}

	db.InitDb(psqlInfo)
	defer db.CloseDb()

	err := db.GetDb().Ping()
	utils.Check(err)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middlewares.JSONContentType())
	router.Use(middlewares.ErrorHandler())
	//router.Static("/static", "static")

	router.POST("/users", controllers.PostUsers)
	router.POST("/authenticate", controllers.PostAuthenticate)
	router.GET("/conversations", controllers.GetConversations)
	router.GET("/conversations/:id", controllers.GetConversation)
	router.POST("/conversations", controllers.PostConversations)
	router.POST("/conversations/:id", controllers.PostConversation)

	router.Run(":" + port)
}
