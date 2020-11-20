package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"

	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
)

func PostDemoUsers(c *gin.Context) {
	db := db.GetDb()

	var params struct {
		Password string `json:"password" binding:"required"`
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

	password := params.Password
	hashedPassword := models.HashPassword(password)

	john, err1 := getDemoUser("john", "John Doe")
	john.Password = hashedPassword

	tim, err2 := getDemoUser("tim", "Tim Miller")
	sarah, err3 := getDemoUser("sarah", "Sarah Smith")
	victoria, err4 := getDemoUser("victoria", "Victoria Jenson")

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

	models.CreateMessage(victoria, john, "Hey John! How's it going?")
	models.CreateMessage(john, victoria, "Not too shabby Victoria. How about you?")
	models.CreateMessage(victoria, john, "Pretty good, pretty good.")

	models.CreateMessage(sarah, john, "Hey John, do you have those files I mentioned?")
	models.CreateMessage(john, sarah, "Yes! I can get them to you by tomorrow.")
	models.CreateMessage(sarah, john, "Awesome! Don't you just love business?")
	models.CreateMessage(john, sarah, "Business rules!")

	models.CreateMessage(john, tim, "Hey Tim! How's it going?")
	models.CreateMessage(tim, john, "Great! Thanks John.")
	models.CreateMessage(tim, john, "Don't you love using nchat?")
	models.CreateMessage(john, tim, "I sure do. This conversation doesn't seem scripted at all.")
	models.CreateMessage(tim, john, "I know! It's like we just said all this naturally.")

	demoUserJson := gin.H{
		"id":       john.ID,
		"username": john.Username,
		"name:":    john.Name,
	}
	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{"user": demoUserJson}))
}

func getDemoUser(username string, name string) (*models.User, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}
	token := base64.URLEncoding.EncodeToString(randBytes)
	demoUsername := strings.ToLower("demo_" + username + "_" + token)

	demoUser := &models.User{
		Username: demoUsername,
		Password: models.HashPassword(token + "80fc201fb6ac4035ebb7ffe9ec61520522e3cc47"),
		Name:     name,
	}
	return demoUser, nil
}
