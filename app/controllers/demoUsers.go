package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
)

func PostDemoUsers(c *gin.Context) {
	db := db.GetDb()

	john, johnPassword, err1 := getDemoUser("john", "John Doe")
	tim, _, err2 := getDemoUser("tim", "Tim Miller")
	sarah, _, err3 := getDemoUser("sarah", "Sarah Smith")
	victoria, _, err4 := getDemoUser("victoria", "Victoria Jenson")

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		utils.AbortErrServer(c)
		return
	}

	demoUsers := []*models.User{john, tim, sarah, victoria}

	for _, user := range demoUsers {
		result := db.Create(user)
		if result.Error != nil {
			utils.AbortErrServer(c)
			return
		}
	}

	session, _, err := models.CreateSession(john.Username, johnPassword)
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	models.CreateMessage(victoria, john, "Hey John! How's it going?")
	models.CreateMessage(john, victoria, "Not too shabby Victoria. How about you?")
	models.CreateMessage(victoria, john, "Pretty good, pretty good.")
	models.CreateMessage(victoria, john, "Say, I have a question for you...")
	models.CreateMessage(victoria, john, "Do you ever feel like you don't exist? Like you're just the"+
		"manifestation")

	models.CreateMessage(sarah, john, "Hey John, do you have those files I mentioned?")
	models.CreateMessage(john, sarah, "Yes! I can get them to you by tomorrow.")
	models.CreateMessage(sarah, john, "Awesome! Don't you just love business?")
	models.CreateMessage(john, sarah, "Business rules!")

	models.CreateMessage(john, tim, "Hey Tim! How's it going?")
	models.CreateMessage(tim, john, "Great! Thanks John.")
	models.CreateMessage(tim, john, "Don't you love using nchat?")
	models.CreateMessage(john, tim, "I sure do. This conversation doesn't seem scripted at all.")
	models.CreateMessage(tim, john, "I know! It's like we just said all this naturally.")

	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{
		"authKey": session.Key,
		"user": gin.H{
			"id":       john.ID,
			"username": john.Username,
			"name":     john.Name,
		},
	}))
}

func getDemoUser(username string, name string) (*models.User, string, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, "", err
	}
	token := base64.URLEncoding.EncodeToString(randBytes)
	demoUsername := strings.ToLower("demo_" + username + "_" + token)
	demoPassword := token + "80fc201fb6ac4035ebb7ffe9ec61520522e3cc47"

	demoUser := &models.User{
		Username: demoUsername,
		Password: models.HashPassword(demoPassword),
		Name:     name,
	}
	return demoUser, demoPassword, nil
}
