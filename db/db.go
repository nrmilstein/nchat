package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"neal-chat/utils"
)

var db *sql.DB

type PsqlInfo struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

func GetDb() *sql.DB {
	return db
}

func InitDb(connectionStr string) {
	var err error
	db, err = sql.Open("postgres", connectionStr)
	utils.Check(err)
}

func InitDbStruct(pi PsqlInfo) {
	connectionStr := fmt.Sprintf("host=%s port=%d user=%s password=%s "+
		"dbname=%s sslmode=disable",
		pi.Host, pi.Port, pi.User, pi.Password, pi.Dbname)

	InitDb(connectionStr)
}

func CloseDb() {
	db.Close()
}
