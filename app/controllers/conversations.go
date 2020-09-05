package controllers

import (
	"net/http"
	"database/sql"
	"strconv"

	"github.com/gin-gonic/gin"

	"neal-chat/app/models"
	"neal-chat/utils"
	"neal-chat/db"
)

func GetConversations(c *gin.Context) {
	db := db.GetDb()
	user, err := models.GetUserFromRequest(c)
	utils.Check(err)

	rows, err := db.Query(
		"SELECT conversations.id, conversations.created, users.id, users.email, users.name, "+
			"users.created "+
			"FROM "+
			"(SELECT conversations_users.conversation_id AS id "+
			"FROM conversations_users "+
			"WHERE conversations_users.user_id = $1) AS my_conversations "+
			"JOIN conversations_users "+
			"ON my_conversations.id = conversations_users.conversation_id "+
			"JOIN users ON conversations_users.user_id = users.id "+
			"JOIN conversations ON conversations_users.conversation_id = conversations.id "+
			"WHERE users.id != $1",
		user.Id)
	defer rows.Close()
	utils.Check(err)

	conversations := []*models.Conversation{}
	for rows.Next() {
		conversation := new(models.Conversation)
		otherUser := new(models.User)

		err = rows.Scan(&conversation.Id, &conversation.Created,
			&otherUser.Id, &otherUser.Name, &otherUser.Email, &otherUser.Created)
		utils.Check(err)

		conversation.Users = []*models.User{user, otherUser}
		conversations = append(conversations, conversation)
	}
	utils.Check(rows.Err())

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func GetConversation(c *gin.Context) {
	db := db.GetDb()

	conversationId, err := strconv.Atoi(c.Param("id"))
	utils.Check(err)

	user, err := models.GetUserFromRequest(c)
	utils.Check(err)

	participants, err := db.Query(
		"SELECT conversations.id, conversations.created, "+
			"users.id, users.email, users.name, users.created "+
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
	utils.Check(err)

	conversation := new(models.Conversation)
	for participants.Next() {
		participant := new(models.User)
		err = participants.Scan(&conversation.Id, &conversation.Created,
			&participant.Id, &participant.Email, &participant.Name, &participant.Created)
		utils.Check(err)

		conversation.Users = append(conversation.Users, participant)
	}
	utils.Check(participants.Err())

	if len(conversation.Users) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	messages, err := db.Query(
		"SELECT messages.id, messages.sent, messages.body, users.id "+
			"FROM conversations_users "+
			"JOIN conversations ON conversations.id = conversations_users.conversation_id "+
			"JOIN messages ON messages.conversation_id = conversations.id "+
			"JOIN users ON messages.user_id = users.id "+
			"WHERE "+
			"conversations_users.user_id = $1 AND "+
			"conversations_users.conversation_id = $2 "+
			"ORDER BY messages.sent ASC ",
		user.Id,
		conversationId)
	defer messages.Close()
	utils.Check(err)

	for messages.Next() {
		message := new(models.Message)
		message.User = new(models.User)
		err = messages.Scan(&message.Id, &message.Sent, &message.Body,
			&message.User.Id)
		utils.Check(err)

		//message.Conversation = conversation
		conversation.Messages = append(conversation.Messages, message)
	}
	utils.Check(messages.Err())

	c.JSON(http.StatusOK, gin.H{"conversation": conversation})
}

func PostConversations(c *gin.Context) {
	db := db.GetDb()

	user, err := models.GetUserFromRequest(c)
	utils.Check(err)

	otherUser := new(models.User)

	conversation := new(models.Conversation)
	message := new(models.Message)
	conversation.Messages = append(conversation.Messages, message)

	var json struct {
		UserId  int    `json:"userId" binding:"required"`
		Message string `json:"message" binding:"required"`
	}
	err = c.ShouldBindJSON(&json)
	utils.Check(err)
	otherUser.Id, message.Body = json.UserId, json.Message

	if otherUser.Id == user.Id {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot create conversation with self"})
		return
	}

	err = db.QueryRow(
		"SELECT users.email, users.name, users.created FROM users WHERE users.id = $1",
		otherUser.Id,
	).Scan(&otherUser.Email, &otherUser.Name, &otherUser.Created)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Target user does not exist"})
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
		utils.Check(err)
		c.JSON(http.StatusConflict, gin.H{"error": "conversation already exists"})
		return
	}

	tx, err := db.Begin()
	utils.Check(err)

	res, err := tx.Exec(
		"SET CONSTRAINTS " +
			"conversations_users_conversation_id_fkey, conversations_users_user_id_fkey " +
			"DEFERRED",
	)
	if err != nil {
		tx.Rollback()
		utils.Check(err)
	}

	err = tx.QueryRow(
		"INSERT "+
			"INTO conversations (created) "+
			"VALUES (CURRENT_TIMESTAMP) "+
			"RETURNING conversations.id, conversations.created",
	).Scan(&conversation.Id, &conversation.Created)
	if err != nil {
		tx.Rollback()
		utils.Check(err)
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
		utils.Check(err)
	}
	rowsAffected, err := res.RowsAffected()
	utils.Check(err)
	if rowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not add conversation"})
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
		utils.Check(err)
	}

	err = tx.Commit()
	utils.Check(err)

	c.JSON(http.StatusCreated, gin.H{
		"success": "conversation added", "conversation": conversation,
	})
}

func PostConversation(c *gin.Context) {
	db := db.GetDb()

	conversationId, err := strconv.Atoi(c.Param("id"))
	utils.Check(err)

	user, err := models.GetUserFromRequest(c)
	utils.Check(err)

	var conversationsUsersId int
	err = db.QueryRow(
		"SELECT conversations_users.id "+
			"FROM conversations_users "+
			"WHERE conversations_users.conversation_id = $1 "+
			"AND conversations_users.user_id = $2",
		conversationId,
		user.Id).Scan(&conversationsUsersId)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	message := new(models.Message)

	var json struct {
		Message string `json:"message" binding:"required"`
	}
	err = c.ShouldBindJSON(&json)
	utils.Check(err)
	message.Body = json.Message

	err = db.QueryRow(
		"INSERT INTO messages(conversation_id, user_id, sent, body) "+
			"VALUES($1, $2, CURRENT_TIMESTAMP, $3) RETURNING messages.id",
		conversationId,
		user.Id,
		message.Body,
	).Scan(&message.Id)
	utils.Check(err)

	c.JSON(http.StatusCreated, gin.H{"success": "message sent", "message": message})
}
