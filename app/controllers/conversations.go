package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
	"gorm.io/gorm"
)

func GetConversations(c *gin.Context) {
	db := db.GetDb()

	user, err := models.GetUserFromRequest(c)
	if err != nil {
		utils.AbortErrForbidden(c)
		return
	}

	var conversations []models.Conversation

	readConversationsResult := db.Model(&user).Preload("Users", "ID <> ?", user.ID).
		Association("Conversations").Find(&conversations)

	if readConversationsResult != nil {
		utils.AbortErrServer(c)
		return
	}

	conversationsJson := []gin.H{}
	for _, conversation := range conversations {
		conversationPartner := conversation.Users[0]
		conversationsJson = append(conversationsJson, gin.H{
			"id":      conversation.ID,
			"created": conversation.CreatedAt,
			"otherUser": gin.H{
				"id":    conversationPartner.ID,
				"email": conversationPartner.Email,
				"name":  conversationPartner.Name,
			},
		})
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(gin.H{"conversations": conversationsJson}))
}

func GetConversation(c *gin.Context) {
	db := db.GetDb()

	user, err := models.GetUserFromRequest(c)
	if err != nil {
		utils.AbortErrForbidden(c)
		return
	}

	conversationIdParam, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	var conversations []models.Conversation
	err = db.Model(&user).
		Preload("Users", "ID <> ?", user.ID).
		Preload("Messages").
		Association("Conversations").Find(&conversations, models.Conversation{ID: conversationIdParam})

	if len(conversations) == 0 {
		c.AbortWithError(http.StatusNotFound,
			utils.AppError{"Conversation not found.", 1, nil})
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.AbortErrServer(c)
		return
	}

	conversation := conversations[0]
	conversationPartner := conversation.Users[0]

	var messagesJson []gin.H

	for _, message := range conversation.Messages {
		messagesJson = append(messagesJson, gin.H{
			"id":       message.ID,
			"senderId": message.UserID,
			"sent":     message.CreatedAt,
			"body":     message.Body,
		})
	}

	conversationJson := gin.H{
		"id":      conversation.ID,
		"created": conversation.CreatedAt,
		"conversationPartner": gin.H{
			"id":    conversationPartner.ID,
			"email": conversationPartner.Email,
			"name":  conversationPartner.Name,
		},
		"messages": messagesJson,
	}

	c.JSON(http.StatusOK,
		utils.SuccessResponse(gin.H{"conversation": conversationJson}))
}

func PostConversations(c *gin.Context) {
	db := db.GetDb()

	user, err := models.GetUserFromRequest(c)
	if err != nil {
		utils.AbortErrForbidden(c)
		return
	}

	otherUser := new(models.User)
	newMessage := new(models.Message)

	var params struct {
		Email   string `json:"email" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	err = c.ShouldBindJSON(&params)
	switch err.(type) {
	case nil:
	case *json.SyntaxError:
		c.AbortWithError(http.StatusBadRequest,
			utils.AppError{"JSON syntax error.", 2, nil})
		return
	case validator.ValidationErrors:
		c.AbortWithError(http.StatusUnauthorized,
			utils.AppError{"Missing parameters.", 3, nil})
		return
	default:
		c.AbortWithError(http.StatusBadRequest,
			utils.AppError{"Could not parse request body.", 4, nil})
		return
	}

	otherUser.Email, newMessage.Body = strings.ToLower(params.Email), params.Message

	if otherUser.Email == user.Email {
		c.AbortWithError(http.StatusConflict,
			utils.AppError{"Cannot create conversation with self.", 1, nil})
		return
	}

	readOtherUserResult := db.Take(&otherUser, &models.User{Email: otherUser.Email})

	if errors.Is(readOtherUserResult.Error, gorm.ErrRecordNotFound) {
		c.AbortWithError(http.StatusUnprocessableEntity,
			utils.AppError{"Recipient user does not exist", 5, nil})
		return
	} else if readOtherUserResult != nil {
		utils.AbortErrServer(c)
		return
	}

	var myConversations []models.Conversation
	var otherUsers []models.User
	err = db.Model(&user).Association("Conversations").Find(&myConversations)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.AbortErrServer(c)
		return
	}

	err = db.Model(&myConversations).
		Where(&models.User{ID: otherUser.ID}).Association("Users").Find(&otherUsers)
	if len(otherUsers) > 0 {
		c.AbortWithError(http.StatusConflict,
			utils.AppError{"Conversation already exists.", 6, nil})
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.AbortErrServer(c)
		return
	}

	newMessage.UserID = user.ID
	newConversation := models.Conversation{
		Messages: []models.Message{
			*newMessage,
		},
		Users: []models.User{
			*user,
			*otherUser,
		},
	}
	createConversationResult := db.Create(&newConversation)
	if createConversationResult.Error != nil {
		utils.AbortErrServer(c)
		return
	}

	conversationJson := gin.H{
		"id":      newConversation.ID,
		"created": newConversation.CreatedAt,
		"messages": []gin.H{
			{
				"id":       newMessage.ID,
				"senderId": user.ID,
				"sent":     newMessage.CreatedAt,
				"body":     newMessage.Body,
			},
		},
		"conversationPartner": gin.H{
			"id":    otherUser.ID,
			"email": otherUser.Email,
			"name":  otherUser.Name,
		},
	}

	c.JSON(http.StatusCreated,
		utils.SuccessResponse(gin.H{"conversation": conversationJson}))
}

func PostConversation(c *gin.Context) {
	db := db.GetDb()

	user, err := models.GetUserFromRequest(c)
	if err != nil {
		utils.AbortErrForbidden(c)
		return
	}

	conversationIdParam, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	var params struct {
		Message string `json:"message" binding:"required"`
	}

	err = c.ShouldBindJSON(&params)
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

	var conversations []models.Conversation
	db.Model(&user).Association("Conversations").
		Find(&conversations, models.Conversation{ID: conversationIdParam})

	if len(conversations) == 0 {
		c.AbortWithError(http.StatusNotFound, utils.AppError{"Conversation does not exist.", 1, nil})
		return
	}

	newMessage := models.Message{
		UserID:         user.ID,
		ConversationID: conversationIdParam,
		Body:           params.Message,
	}
	createMessageResult := db.Create(&newMessage)
	if createMessageResult.Error != nil {
		utils.AbortErrServer(c)
		return
	}

	messageJson := gin.H{
		"id":             newMessage.ID,
		"conversationId": newMessage.ConversationID,
		"senderId":       newMessage.UserID,
		"sent":           newMessage.CreatedAt,
		"body":           newMessage.Body,
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{"message": messageJson}))
}
