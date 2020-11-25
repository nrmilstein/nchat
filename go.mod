module github.com/nrmilstein/nchat

// +heroku goVersion go1.15
go 1.14

require (
	github.com/gin-contrib/static v0.0.0-20200916080430-d45d9a37d28e
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/heroku/x v0.0.24
	github.com/lib/pq v1.7.0
	github.com/unrolled/secure v1.0.1
	gorm.io/driver/postgres v1.0.2
	gorm.io/gorm v1.20.2
	nhooyr.io/websocket v1.8.6
)
