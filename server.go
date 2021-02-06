package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"

	"github.com/gin-contrib/static"
	"github.com/nrmilstein/nchat/app/controllers"
	"github.com/nrmilstein/nchat/app/middlewares"
	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/chatServer"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
)

func main() {
	dynoEnv := os.Getenv("DYNO")
	if dynoEnv == "" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("Error: no environment variale DATABASE_URL found.")
	}
	db.InitDb(databaseUrl)

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

	allowedHosts := []string{"nchat-app.herokuapp.com"}
	router.Use(middlewares.Secure(allowedHosts, gin.IsDebugging()))

	hub := chatServer.NewHub()

	api := router.Group("/api/v1")
	{
		api.Use(middlewares.JSONContentType())
		api.Use(middlewares.ErrorHandler())
		api.POST("/users", controllers.PostUsers)
		api.POST("/demoUsers", controllers.PostDemoUsers)
		api.GET("/users/:username", controllers.GetUser)
		api.POST("/authenticate", controllers.PostAuthenticate)
		api.GET("/authenticate", controllers.GetAuthenticate)
		api.GET("/conversations", controllers.GetConversations)
		api.GET("/conversations/:id", controllers.GetConversation)
		api.GET("/chat", controllers.GetChat(hub))
	}

	router.Use(static.Serve("/", static.LocalFile("./nchat-web", true)))
	router.Use(static.Serve("/accounts/login", static.LocalFile("./nchat-web", true)))
	router.Use(static.Serve("/accounts/get-started", static.LocalFile("./nchat-web", true)))

	router.NoRoute(controllers.NoRoute)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	router.Run(":" + port)
}
