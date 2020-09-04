package main

import (
	"errors"
	"log"
	"net/http"
	"time"
	//"os"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	_ "github.com/heroku/x/hmetrics/onload"
	_ "github.com/lib/pq"
	"github.com/kr/pretty"
)

var db *sql.DB

type User struct {
	Id      int
	Email   string
	Name    string
	Created time.Time
}

type Conversation struct {
	Id       int
	Created  time.Time
	Users    []*User
	Messages []*Message
}

type Message struct {
	Id           int
	Conversation *Conversation
	User         *User
	Sent         time.Time
	Body         string
}

type PsqlInfo struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

type AppError struct {
  Message string
  Code int
  Data interface{}
}

//type FailResponse struct {
  //HttpStatus int
  //Data interface{}
//}

func (e AppError ) Error() string {
  return e.Message
}

//func (e FailResponse) Error() string {
  //return e.Message
//}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func successResponse(data interface{}) gin.H {
	return gin.H{
		"status": "success",
		"data":   data,
	}
}

func failResponse(data interface{}) gin.H {
	return gin.H{
		"status": "fail",
		"data":   data,
	}
}

func errorResponse(message string, code int, data interface{}) gin.H {
	response := gin.H{
		"status":  "error",
		"message": message,
		"code":    code,
	}
	if data != nil {
		response["data"] = data
	}
	return response
}

func ErrorHandler() gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Next()

    if c.Errors != nil && len(c.Errors) > 0 {
      err := c.Errors[0].Err
      switch err.(type) {
      case AppError:
        appErr := err.(AppError)
        c.JSON(c.Writer.Status(), errorResponse(appErr.Message, appErr.Code, appErr.Data))
      //case FailResponse:
        //c.JSON(err.HttpStatus, failResponse(err.Data))
      default:
        c.JSON(http.StatusInternalServerError,
          errorResponse("500: An internal error was encountered.", 1, nil))
      }
    }
  }
}

func JSONContentType() gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
    c.Next()
  }
}

func setupDb(pi PsqlInfo) *sql.DB {
	connectionStr := fmt.Sprintf("host=%s port=%d user=%s password=%s "+
		"dbname=%s sslmode=disable",
		pi.Host, pi.Port, pi.User, pi.Password, pi.Dbname)

	db, err := sql.Open("postgres", connectionStr)
	check(err)
	return db
}

func hashPassword(str string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(str)))
}

func getUserFromKey(key string) (*User, error) {
	user := new(User)
	err := db.QueryRow(
		"SELECT users.id, users.email, users.name, users.created "+
			"FROM auth_keys JOIN users ON auth_keys.user_id = users.id "+
			"WHERE auth_key = $1",
		key,
	).Scan(&user.Id, &user.Email, &user.Name, &user.Created)
	if err == sql.ErrNoRows {
		return nil, errors.New("No user found")
	}
	check(err)
	return user, nil
}

func getUserFromRequest(c *gin.Context) (*User, error) {
	return getUserFromKey(c.GetHeader("X-API-Key"))
}

func postUsers(c *gin.Context) {
	var json struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&json)
	check(err)
	email, password := json.Email, json.Password

	var id int
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)
	if err != sql.ErrNoRows {
		check(err)
		c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
		return
	}

	hashedPassword := hashPassword(password)

	var userId int
	var created time.Time
	err = db.QueryRow(
		"INSERT INTO users(email, password, created) "+
			"VALUES($1, $2, CURRENT_TIMESTAMP) "+
			"RETURNING users.id, users.created",
		email,
		hashedPassword,
	).Scan(&userId, &created)
	check(err)
	addedUser := User{userId, email, "", created}
	c.JSON(http.StatusCreated, gin.H{"success": "user added", "user": addedUser})
}

func postAuthenticate(c *gin.Context) {
	invalidCredError := AppError{"invalid email/password", 3, nil}

	var params struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	err := c.ShouldBindJSON(&params)
	switch err.(type) {
	case nil:
	case *json.SyntaxError:
		c.AbortWithError(http.StatusBadRequest, AppError{"JSON syntax error", 2, nil})
		return
	case validator.ValidationErrors: // TODO: make this case work
		c.AbortWithError(http.StatusForbidden, invalidCredError)
		return
	default:
		c.AbortWithError(http.StatusBadRequest, AppError{"Could not parse request body", 4, nil})
		return
	}
	email, password := params.Email, params.Password

	user := new(User)
	hashedPassword := hashPassword(password)
	err = db.QueryRow(
		"SELECT id, email, name FROM users WHERE email = $1 AND password = $2",
		email, hashedPassword).Scan(&user.Id, &user.Email, &user.Name)
	if err == sql.ErrNoRows {
		c.AbortWithError(http.StatusForbidden, invalidCredError)
		return
	} else if err != nil {
		check(err)
	}

	randBytes := make([]byte, 18)
	_, err = rand.Read(randBytes)
	check(err)
	authKey := base64.URLEncoding.EncodeToString(randBytes)

	res, err := db.Exec(
		"INSERT INTO auth_keys (auth_key, user_id, created, accessed) "+
			"VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		authKey, user.Id)
	check(err)
	rowCount, err := res.RowsAffected()
	check(err)
	if rowCount == 0 {
		c.AbortWithError(http.StatusInternalServerError,
      AppError{"could not authenticate", 4, nil})
		return
	}

	c.JSON(http.StatusCreated, successResponse(gin.H{
		"auth_key": authKey,
		"user": gin.H{
			"id":    user.Id,
			"email": user.Email,
			"name":  user.Name,
		},
	}))
}

func getConversations(c *gin.Context) {
	user, err := getUserFromRequest(c)
	check(err)

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
	check(err)

	conversations := []*Conversation{}
	for rows.Next() {
		conversation := new(Conversation)
		otherUser := new(User)

		err = rows.Scan(&conversation.Id, &conversation.Created,
			&otherUser.Id, &otherUser.Name, &otherUser.Email, &otherUser.Created)
		check(err)

		conversation.Users = []*User{user, otherUser}
		conversations = append(conversations, conversation)
	}
	check(rows.Err())

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func getConversationsId(c *gin.Context) {
	conversationId, err := strconv.Atoi(c.Param("id"))
	check(err)

	user, err := getUserFromRequest(c)
	check(err)

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
	check(err)

	conversation := new(Conversation)
	for participants.Next() {
		participant := new(User)
		err = participants.Scan(&conversation.Id, &conversation.Created,
			&participant.Id, &participant.Email, &participant.Name, &participant.Created)
		check(err)

		conversation.Users = append(conversation.Users, participant)
	}
	check(participants.Err())

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
	check(err)

	for messages.Next() {
		message := new(Message)
		message.User = new(User)
		err = messages.Scan(&message.Id, &message.Sent, &message.Body,
			&message.User.Id)
		check(err)

		//message.Conversation = conversation
		conversation.Messages = append(conversation.Messages, message)
	}
	check(messages.Err())

	c.JSON(http.StatusOK, gin.H{"conversation": conversation})
}

func postConversations(c *gin.Context) {
	user, err := getUserFromRequest(c)
	check(err)

	otherUser := new(User)

	conversation := new(Conversation)
	message := new(Message)
	conversation.Messages = append(conversation.Messages, message)

	var json struct {
		UserId  int    `json:"userId" binding:"required"`
		Message string `json:"message" binding:"required"`
	}
	err = c.ShouldBindJSON(&json)
	check(err)
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
		check(err)
		c.JSON(http.StatusConflict, gin.H{"error": "conversation already exists"})
		return
	}

	tx, err := db.Begin()
	check(err)

	res, err := tx.Exec(
		"SET CONSTRAINTS " +
			"conversations_users_conversation_id_fkey, conversations_users_user_id_fkey " +
			"DEFERRED",
	)
	if err != nil {
		tx.Rollback()
		check(err)
	}

	err = tx.QueryRow(
		"INSERT "+
			"INTO conversations (created) "+
			"VALUES (CURRENT_TIMESTAMP) "+
			"RETURNING conversations.id, conversations.created",
	).Scan(&conversation.Id, &conversation.Created)
	if err != nil {
		tx.Rollback()
		check(err)
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
		check(err)
	}
	rowsAffected, err := res.RowsAffected()
	check(err)
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
		check(err)
	}

	err = tx.Commit()
	check(err)

	c.JSON(http.StatusCreated, gin.H{
		"success": "conversation added", "conversation": conversation,
	})
}

func postConversationsId(c *gin.Context) {
	conversationId, err := strconv.Atoi(c.Param("id"))
	check(err)

	user, err := getUserFromRequest(c)
	check(err)

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

	message := new(Message)

	var json struct {
		Message string `json:"message" binding:"required"`
	}
	err = c.ShouldBindJSON(&json)
	check(err)
	message.Body = json.Message

	err = db.QueryRow(
		"INSERT INTO messages(conversation_id, user_id, sent, body) "+
			"VALUES($1, $2, CURRENT_TIMESTAMP, $3) RETURNING messages.id",
		conversationId,
		user.Id,
		message.Body,
	).Scan(&message.Id)
	check(err)

	c.JSON(http.StatusCreated, gin.H{"success": "message sent", "message": message})
}

func main() {
  if false {
    pretty.Println()
  }

	psqlInfo := PsqlInfo{
		Host:     "localhost",
		Port:     5432,
		User:     "nrmilstein",
		Password: "password",
		Dbname:   "neal_chat",
	}

	db = setupDb(psqlInfo)
	defer db.Close()

	err := db.Ping()
	check(err)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
  router.Use(JSONContentType())
  router.Use(ErrorHandler())
	//router.Static("/static", "static")

	router.POST("/users", postUsers)
	router.POST("/authenticate", postAuthenticate)
	router.GET("/conversations", getConversations)
	router.GET("/conversations/:id", getConversationsId)
	router.POST("/conversations", postConversations)
	router.POST("/conversations/:id", postConversationsId)

	router.Run(":" + strconv.Itoa(5000))
}
