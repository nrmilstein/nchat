package chatServer

import "time"

type wsMsgNotification struct {
	Type   string                `json:"type"`
	Method string                `json:"method"`
	Data   wsMsgNotificationData `json:"data"`
}

type wsMsgNotificationData struct {
	Message      wsMsgMessage      `json:"message"`
	Conversation wsMsgConversation `json:"conversation"`
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
	Message      wsMsgMessage      `json:"message"`
	Conversation wsMsgConversation `json:"conversation"`
}
type wsMsgMessage struct {
	Id             int       `json:"id"`
	ConversationId int       `json:"conversationId"`
	SenderId       int       `json:"senderId"`
	Body           string    `json:"body"`
	CreatedAt      time.Time `json:"sent"`
}

type wsMsgConversation struct {
	Id                  int                      `json:"id"`
	CreatedAt           time.Time                `json:"created"`
	ConversationPartner wsMsgConversationPartner `json:"conversationPartner"`
}

type wsMsgConversationPartner struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
