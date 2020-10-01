package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"

	"github.com/nrmilstein/nchat/app/controllers"
	"github.com/nrmilstein/nchat/app/middlewares"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		psqlInfo := db.PsqlInfo{
			Host:     "localhost",
			Port:     5432,
			User:     "nrmilstein",
			Password: "password",
			Dbname:   "nchat",
		}
		db.InitDbStruct(psqlInfo)
	} else {
		db.InitDb(databaseUrl)
	}
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
	router.GET("/authenticate", controllers.GetAuthenticate)
	router.GET("/conversations", controllers.GetConversations)
	router.GET("/conversations/:id", controllers.GetConversation)
	router.POST("/conversations", controllers.PostConversations)
	router.POST("/conversations/:id", controllers.PostConversation)

	router.Run(":" + port)
}
