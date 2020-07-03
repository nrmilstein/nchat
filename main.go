package main

import (
  "errors"
  "log"
	"net/http"
	//"os"
  "database/sql"
  "fmt"
  "strconv"
  "crypto/sha256"
  "crypto/rand"
  "encoding/base64"

	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
  _ "github.com/lib/pq"
)

var db *sql.DB

func check(err error) {
  if err != nil {
    log.Panic(err)
  }
}

type PsqlInfo struct {
    Host      string
    Port      int
    User      string
    Password  string
    Dbname    string
}

func setupDb(pi PsqlInfo) *sql.DB {
  connectionStr := fmt.Sprintf("host=%s port=%d user=%s " +
      "password=%s dbname=%s sslmode=disable",
    pi.Host, pi.Port, pi.User, pi.Password, pi.Dbname)

  db, err := sql.Open("postgres", connectionStr)
  check(err)
  return db
}

func hashPassword(str string) string {
  return fmt.Sprintf("%x", sha256.Sum256([]byte(str)))
}

type AuthenticatedUser struct {
  UserId int
  Email string
}

func authenticate(key string) (*AuthenticatedUser, error) {
  var userId int
  var email string
  err := db.QueryRow("SELECT user_id, email FROM auth_keys WHERE auth_key = $1", key).
      Scan(&userId, &email)
  if err == sql.ErrNoRows {
    return nil, errors.New("No user found")
  }
  check(err)
  return &AuthenticatedUser{userId, email}, nil
}

type PostUserData struct {
  Email string `json:"email" binding:"required"`
  Password string `json:"password" binding:"required"`
}

func postUsers(c *gin.Context) {
  var json PostUserData
  err := c.BindJSON(&json)
  check(err)
  email, password := json.Email, json.Password
  hashedPassword := hashPassword(password)

  var id int
  err2 := db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&id)
  if err2 == nil {
    c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
    return
  } else if err2 != sql.ErrNoRows {
    check(err2)
  }

  res, err3 := db.Exec("INSERT INTO users(email, password, created) " +
      "VALUES($1, $2, CURRENT_TIMESTAMP)",
    email, hashedPassword)
  check(err3)
  rowsAffected, err4 := res.RowsAffected()
  check(err4)
  if rowsAffected == 0 {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "user couldnot be added"})
    return
  }
  c.JSON(http.StatusOK, gin.H{"success": "user added"})
}

type postAuthenticateData struct {
  Email string `json:"email" binding:"required"`
  Password string `json:"password" binding:"required"`
}

func postAuthenticate(c *gin.Context) {
  var json postAuthenticateData
  err := c.BindJSON(&json)
  check(err)
  email, password := json.Email, json.Password
  hashedPassword := hashPassword(password)

  var userId int
  err2 := db.QueryRow("SELECT id FROM users WHERE email = $1 AND password = $2",
      email, hashedPassword).Scan(&userId)
  if err2 == sql.ErrNoRows {
    c.JSON(http.StatusForbidden, gin.H{"error": "invalid email/password"})
    return
  } else if err2 != nil {
    check(err2)
  }

	randBytes := make([]byte, 18)
	_, err3 := rand.Read(randBytes)
  check(err3)
  authKey := base64.URLEncoding.EncodeToString(randBytes)

  res, err4 := db.Exec("INSERT INTO auth_keys (auth_key, user_id, created, accessed) " + 
      "VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", authKey, userId)
  check(err4)
  rowCount, err5 := res.RowsAffected()
  check(err5)
  if rowCount < 1 {
    c.JSON(http.StatusInternalServerError, gin.H{"error": "could not authenticate"})
    return
  }

  c.JSON(http.StatusOK, gin.H{"auth_key": authKey})
}

func main() {
  psqlInfo := PsqlInfo {
    Host:      "localhost",
    Port:      5432,
    User:      "nrmilstein",
    Password:  "password",
    Dbname:    "neal_chat",
  }

  db = setupDb(psqlInfo)
  defer db.Close()

  err := db.Ping()
  check(err)

	router := gin.New()
	router.Use(gin.Logger())
	//router.Static("/static", "static")

  router.POST("/users", postUsers)
  router.POST("/authenticate", postAuthenticate)

	router.Run(":" + strconv.Itoa(5000))
}
