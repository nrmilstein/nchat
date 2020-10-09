package db

import (
	"fmt"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/nrmilstein/nchat/utils"
)

var db *gorm.DB

type PsqlInfo struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
}

func GetDb() *gorm.DB {
	return db
}

func InitDb(connectionStr string) {
	var err error
	db, err = gorm.Open(postgres.Open(connectionStr),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	utils.Check(err)
}

func InitDbStruct(pi PsqlInfo) {
	connectionStr := fmt.Sprintf("host=%s port=%d user=%s password=%s "+
		"dbname=%s sslmode=disable",
		pi.Host, pi.Port, pi.User, pi.Password, pi.Dbname)

	InitDb(connectionStr)
}
