package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
			"conversationPartner": gin.H{
				"id":       conversationPartner.ID,
				"username": conversationPartner.Username,
				"name":     conversationPartner.Name,
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
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Order("messages.created_at ASC")
		}).
		Association("Conversations").Find(&conversations, models.Conversation{ID: conversationIdParam})

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.AbortErrServer(c)
		return
	}

	if len(conversations) == 0 {
		c.AbortWithError(http.StatusNotFound,
			utils.AppError{"Conversation not found.", 1, nil})
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
			"id":       conversationPartner.ID,
			"username": conversationPartner.Username,
			"name":     conversationPartner.Name,
		},
		"messages": messagesJson,
	}

	c.JSON(http.StatusOK,
		utils.SuccessResponse(gin.H{"conversation": conversationJson}))
}
