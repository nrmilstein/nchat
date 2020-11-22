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

	tim, timPassword, err1 := getDemoUser("tim", "Tim Talkalot")
	sarah, _, err2 := getDemoUser("sarah", "Sarah Smiley")
	nick, _, err3 := getDemoUser("nick", "Nick NewMessage")
	victoria, _, err4 := getDemoUser("victoria", "Victoria Chatterbox")

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		utils.AbortErrServer(c)
		return
	}

	demoUsers := []*models.User{tim, sarah, nick, victoria}

	for _, user := range demoUsers {
		result := db.Create(user)
		if result.Error != nil {
			utils.AbortErrServer(c)
			return
		}
	}

	session, _, err := models.CreateSession(tim.Username, timPassword)
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	models.CreateMessage(victoria, tim, "Hey Tim! How are you?")
	models.CreateMessage(tim, victoria, "Not too shabby Victoria. How about you?")
	models.CreateMessage(victoria, tim, "Pretty good, pretty good.")
	models.CreateMessage(victoria, tim, "Say, I have a question for you...")
	models.CreateMessage(victoria, tim, "Do you ever feel like you don't exist? Like you're just "+
		"a demo account in some web application?")
	models.CreateMessage(tim, victoria, "Maybe...why do you ask?")
	models.CreateMessage(victoria, tim, "No reason really. It's just, I've felt very...artificial "+
		"lately.")
	models.CreateMessage(victoria, tim, "Like I was just created a few seconds ago.")
	models.CreateMessage(tim, victoria, "Victoria, do you really think that if we were demo "+
		"accounts, they would give us these normal names like Victoria Chatterbox?")
	models.CreateMessage(victoria, tim, "You're right, Tim. If we were demo accounts, they'd "+
		"probably name us something crazy like Talky McMessageFace.")
	models.CreateMessage(tim, victoria, "Exactly. There's no way that could be us.")
	models.CreateMessage(victoria, tim, "Gee, thanks Tim! I feel loads better already.")
	models.CreateMessage(tim, victoria, "You're welcome! Have I ever told you about the great features "+
		"of nchat?")
	models.CreateMessage(victoria, tim, "No...why do you bring it up?")
	models.CreateMessage(tim, victoria, "No reason! Just something I feel like saying.")
	models.CreateMessage(tim, victoria, "Anyways, gotta go!")
	models.CreateMessage(victoria, tim, "Bye!")

	models.CreateMessage(nick, tim, "Hey Tim, do you have those files I mentioned?")
	models.CreateMessage(tim, nick, "Yes! I can get them to you by tomorrow.")
	models.CreateMessage(nick, tim, "Awesome! Don't you just love business?")
	models.CreateMessage(tim, nick, "Business rules!")

	models.CreateMessage(tim, sarah, "Hey Sarah! How's it going?")
	models.CreateMessage(sarah, tim, "Great! Thanks Tim.")
	models.CreateMessage(sarah, tim, "Don't you love using nchat?")
	models.CreateMessage(tim, sarah, "I sure do. This conversation doesn't seem scripted at all.")
	models.CreateMessage(sarah, tim, "I know! It's like we just said all this naturally.")

	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{
		"authKey": session.Key,
		"user": gin.H{
			"id":       tim.ID,
			"username": tim.Username,
			"name":     tim.Name,
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
