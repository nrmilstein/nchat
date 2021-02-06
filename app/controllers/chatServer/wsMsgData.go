package chatServer

import "time"

type wsMsgRequestData struct {
	Username string `json:"username"`
	Body     string `json:"body"`
}

type wsMsgData struct {
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
	Id       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}
