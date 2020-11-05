package controllers

import (
	"errors"
	"net/http"
	"sort"
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

	readConversationsResult := db.Model(&user).
		Preload("Users", "ID <> ?", user.ID).
		Preload("Messages", func(db *gorm.DB) *gorm.DB {
			return db.Select("DISTINCT ON (conversation_id) *").Order("conversation_id, created_at DESC")
		}).
		Association("Conversations").Find(&conversations)

	if readConversationsResult != nil {
		utils.AbortErrServer(c)
		return
	}

	sort.Slice(conversations, func(i, j int) bool {
		if len(conversations[i].Messages) == 0 {
			return true
		} else if len(conversations[j].Messages) == 0 {
			return false
		}
		return conversations[i].Messages[0].CreatedAt.After(conversations[j].Messages[0].CreatedAt)
	})

	conversationsJson := []gin.H{}
	for _, conversation := range conversations {
		firstMessage := conversation.Messages[0]
		conversationPartner := conversation.Users[0]

		conversationsJson = append(conversationsJson, gin.H{
			"id":      conversation.ID,
			"created": conversation.CreatedAt,
			"conversationPartner": gin.H{
				"id":       conversationPartner.ID,
				"username": conversationPartner.Username,
				"name":     conversationPartner.Name,
			},
			"messages": []gin.H{
				{
					"id":       firstMessage.ID,
					"senderId": firstMessage.UserID,
					"sent":     firstMessage.CreatedAt,
					"body":     firstMessage.Body,
				},
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
