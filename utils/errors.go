package utils

import (
	"log"
)

type AppError struct {
	Message string
	Code    int
	Data    interface{}
}

//type FailResponse struct {
//HttpStatus int
//Data interface{}
//}

func (e AppError) Error() string {
	return e.Message
}

//func (e FailResponse) Error() string {
//return e.Message
//}

func Check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
