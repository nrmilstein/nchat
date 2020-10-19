package chatServer

import "time"

type wsMsgNotification struct {
	Type   string                `json:"type"`
	Method string                `json:"method"`
	Data   wsMsgNotificationData `json:"data"`
}

type wsMsgNotificationData struct {
	Message wsMsgNotificationMessage `json:"message"`
}

type wsMsgNotificationMessage struct {
	Id             int       `json:"id"`
	ConversationId int       `json:"conversationId"`
	SenderId       int       `json:"senderId"`
	Body           string    `json:"body"`
	Created        time.Time `json:"sent"`
}

type wsMsgRequest struct {
	Id     int              `json:"id"`
	Type   string           `json:"type"`
	Method string           `json:"method"`
	Data   wsMsgRequestData `json:"data"`
}

type wsMsgRequestData struct {
	Email string `json:"email"`
	Body  string `json:"body"`
}

type wsMsgSuccessResponse struct {
	Id     int                      `json:"id"`
	Type   string                   `json:"type"`
	Status string                   `json:"status"`
	Data   wsMsgSuccessResponseData `json:"data"`
}

type wsMsgSuccessResponseData struct {
	Message wsMsgSuccessResponseMessage `json:"message"`
}

type wsMsgSuccessResponseMessage struct {
	Id             int       `json:"id"`
	ConversationId int       `json:"conversationId"`
	SenderId       int       `json:"senderId"`
	Body           string    `json:"body"`
	Created        time.Time `json:"sent"`
}
