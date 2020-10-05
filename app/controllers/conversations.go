package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"

	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
)

func GetConversations(c *gin.Context) {
	db := db.GetDb()

	user, err := models.GetUserFromRequest(c)
	if err != nil {
		utils.AbortErrForbidden(c)
		return
	}

	rows, err := db.Query(
		"SELECT my_conversations.id, users.id, users.email, users.name "+
			"FROM "+
			"(SELECT conversations_users.conversation_id AS id "+
			"FROM conversations_users "+
			"WHERE conversations_users.user_id = $1) AS my_conversations "+
			"JOIN conversations_users "+
			"ON my_conversations.id = conversations_users.conversation_id "+
			"JOIN users ON conversations_users.user_id = users.id "+
			"WHERE users.id != $1",
		user.Id)
	defer rows.Close()
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	conversations := []gin.H{}
	for rows.Next() {
		conversation := new(models.Conversation)
		otherUser := new(models.User)

		err = rows.Scan(&conversation.Id, &otherUser.Id, &otherUser.Email, &otherUser.Name)
		if err != nil {
			utils.AbortErrServer(c)
			return
		}

		conversations = append(conversations, gin.H{
			"id": conversation.Id,
			"users": []gin.H{
				{
					"id":    user.Id,
					"email": user.Email,
					"name":  user.Name,
				},
				{
					"id":    otherUser.Id,
					"email": otherUser.Email,
					"name":  otherUser.Name,
				},
			},
		})
	}
	if err = rows.Err(); err != nil {
		utils.AbortErrServer(c)
		return
	}

	c.JSON(http.StatusOK, utils.SuccessResponse(gin.H{"conversations": conversations}))
}

func GetConversation(c *gin.Context) {
	db := db.GetDb()

	user, err := models.GetUserFromRequest(c)
	if err != nil {
		utils.AbortErrForbidden(c)
		return
	}

	conversationId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	participants, err := db.Query(
		"SELECT conversations.id, conversations.created, "+
			"users.id, users.email, users.name "+
			"FROM "+
			"(SELECT conversations_users.conversation_id AS id "+
			"FROM conversations_users "+
			"WHERE conversations_users.user_id = $1 AND "+
			"conversations_users.conversation_id = $2) AS my_conversations "+
			"JOIN conversations ON conversations.id = my_conversations.id "+
			"JOIN conversations_users "+
			"ON conversations.id = conversations_users.conversation_id "+
			"JOIN users ON conversations_users.user_id = users.id",
		user.Id,
		conversationId)
	defer participants.Close()
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	conversation := new(models.Conversation)
	for participants.Next() {
		participant := new(models.User)
		err = participants.Scan(&conversation.Id, &conversation.Created,
			&participant.Id, &participant.Email, &participant.Name)
		if err != nil {
			utils.AbortErrServer(c)
			return
		}

		conversation.Users = append(conversation.Users, participant)
	}
	if err = participants.Err(); err != nil {
		utils.AbortErrServer(c)
		return
	}

	if len(conversation.Users) == 0 {
		c.AbortWithError(http.StatusNotFound,
			utils.AppError{"Conversation not found.", 1, nil})
		return
	}

	messages, err := db.Query(
		"SELECT messages.id, messages.sent, messages.body, users.id "+
			"FROM messages "+
			"JOIN users ON messages.user_id = users.id "+
			"WHERE messages.conversation_id = $1 "+
			"ORDER BY messages.sent ASC ",
		conversationId)
	defer messages.Close()
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	for messages.Next() {
		message := new(models.Message)
		message.User = new(models.User)
		err = messages.Scan(&message.Id, &message.Sent, &message.Body,
			&message.User.Id)
		if err != nil {
			utils.AbortErrServer(c)
			return
		}

		conversation.Messages = append(conversation.Messages, message)
	}
	if err = messages.Err(); err != nil {
		utils.AbortErrServer(c)
		return
	}

	usersJson := []gin.H{}
	for i := range conversation.Users {
		usersJson = append(usersJson, gin.H{
			"id":    conversation.Users[i].Id,
			"email": conversation.Users[i].Email,
			"name":  conversation.Users[i].Name,
		})
	}

	messagesJson := []gin.H{}
	for i := range conversation.Messages {
		messagesJson = append(messagesJson, gin.H{
			"id":     conversation.Messages[i].Id,
			"userId": conversation.Messages[i].User.Id,
			"sent":   conversation.Messages[i].Sent,
			"body":   conversation.Messages[i].Body,
		})
	}

	conversationJson := gin.H{
		"id":       conversation.Id,
		"created":  conversation.Created,
		"users":    usersJson,
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

	conversation := new(models.Conversation)
	message := new(models.Message)
	conversation.Messages = append(conversation.Messages, message)

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

	otherUser.Email, message.Body = strings.ToLower(params.Email), params.Message

	if otherUser.Email == user.Email {
		c.AbortWithError(http.StatusConflict,
			utils.AppError{"Cannot create conversation with self.", 1, nil})
		return
	}

	err = db.QueryRow(
		"SELECT users.id, users.name FROM users WHERE users.email = $1",
		otherUser.Email,
	).Scan(&otherUser.Id, &otherUser.Name)
	if err == sql.ErrNoRows {
		c.AbortWithError(http.StatusUnprocessableEntity,
			utils.AppError{"Recipient user does not exist", 5, nil})
		return
	}

	err = db.QueryRow(
		"SELECT conversations_users.id "+
			"FROM "+
			"(SELECT conversations_users.conversation_id AS id "+
			"FROM conversations_users "+
			"WHERE conversations_users.user_id = $1) AS my_conversations "+
			"JOIN conversations_users "+
			"ON my_conversations.id = conversations_users.conversation_id "+
			"WHERE conversations_users.user_id = $2",
		user.Id,
		otherUser.Id,
	).Scan(&conversation.Id)
	if err != sql.ErrNoRows {
		if err == nil {
			c.AbortWithError(http.StatusConflict,
				utils.AppError{"Conversation already exists.", 6, nil})
			return
		} else {
			utils.AbortErrServer(c)
			return
		}
	}

	tx, err := db.Begin()
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	res, err := tx.Exec(
		"SET CONSTRAINTS " +
			"conversations_users_conversation_id_fkey, conversations_users_user_id_fkey " +
			"DEFERRED",
	)
	if err != nil {
		tx.Rollback()
		utils.AbortErrServer(c)
		return
	}

	err = tx.QueryRow(
		"INSERT "+
			"INTO conversations (created) "+
			"VALUES (CURRENT_TIMESTAMP) "+
			"RETURNING conversations.id, conversations.created",
	).Scan(&conversation.Id, &conversation.Created)
	if err != nil {
		tx.Rollback()
		utils.AbortErrServer(c)
		return
	}

	res, err = tx.Exec(
		"INSERT "+
			"INTO conversations_users (conversation_id, user_id) "+
			"VALUES "+
			"($1, $2), "+
			"($1, $3)",
		conversation.Id,
		user.Id,
		otherUser.Id,
	)
	if err != nil {
		tx.Rollback()
		utils.AbortErrServer(c)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		tx.Rollback()
		utils.AbortErrServer(c)
		return
	}

	err = tx.QueryRow(
		"INSERT "+
			"INTO messages (conversation_id, user_id, sent, body) "+
			"VALUES ($1, $2, CURRENT_TIMESTAMP, $3) "+
			"RETURNING messages.id, messages.sent",
		conversation.Id,
		user.Id,
		message.Body,
	).Scan(&message.Id, &message.Sent)
	if err != nil {
		tx.Rollback()
		utils.AbortErrServer(c)
		return
	}

	err = tx.Commit()
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	conversationJson := gin.H{
		"id":      conversation.Id,
		"created": conversation.Created,
		"messages": []gin.H{
			{
				"id":     message.Id,
				"userId": user.Id,
				"sent":   message.Sent,
				"body":   message.Body,
			},
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

	conversationId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	message := new(models.Message)
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

	message.Body = params.Message

	var conversationsUsersId int
	err = db.QueryRow(
		"SELECT conversations_users.id "+
			"FROM conversations_users "+
			"WHERE conversations_users.conversation_id = $1 "+
			"AND conversations_users.user_id = $2",
		conversationId,
		user.Id).Scan(&conversationsUsersId)
	if err == sql.ErrNoRows {
		c.AbortWithError(http.StatusNotFound, utils.AppError{"Conversation does not exist.", 1, nil})
		return
	}

	err = db.QueryRow(
		"INSERT INTO messages(conversation_id, user_id, sent, body) "+
			"VALUES($1, $2, CURRENT_TIMESTAMP, $3) RETURNING messages.id, messages.sent",
		conversationId,
		user.Id,
		message.Body,
	).Scan(&message.Id, &message.Sent)
	if err != nil {
		utils.AbortErrServer(c)
		return
	}

	messageJson := gin.H{
		"id":             message.Id,
		"conversationId": conversationId,
		"userId":         user.Id,
		"sent":           message.Sent,
		"body":           message.Body,
	}

	c.JSON(http.StatusCreated, utils.SuccessResponse(gin.H{"message": messageJson}))
}
