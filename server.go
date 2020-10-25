package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"

	"github.com/gin-contrib/static"
	"github.com/nrmilstein/nchat/app/controllers"
	"github.com/nrmilstein/nchat/app/controllers/chatServer"
	"github.com/nrmilstein/nchat/app/middlewares"
	"github.com/nrmilstein/nchat/app/models"
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

	db.GetDb()

	err := db.GetDb().AutoMigrate(
		&models.Session{},
		&models.Conversation{},
		&models.Message{},
		&models.User{},
	)
	utils.Check(err)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	hub := chatServer.NewHub()

	api := router.Group("/api/v1")
	{
		api.Use(middlewares.JSONContentType())
		api.Use(middlewares.ErrorHandler())
		api.POST("/users", controllers.PostUsers)
		api.GET("/users/:email", controllers.GetUser)
		api.POST("/authenticate", controllers.PostAuthenticate)
		api.GET("/authenticate", controllers.GetAuthenticate)
		api.GET("/conversations", controllers.GetConversations)
		api.GET("/conversations/:id", controllers.GetConversation)
		api.GET("/chat", hub.GetChat)
	}

	router.Use(static.Serve("/", static.LocalFile("./nchat-web/build", true)))

	router.NoRoute(controllers.NoRoute)

	router.Run(":" + port)
}
