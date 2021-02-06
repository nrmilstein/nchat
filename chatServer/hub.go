package chatServer

import (
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
)

type Hub struct {
	clientsMutex sync.RWMutex
	clients      map[int]clientGroup
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int]clientGroup),
	}
}

func (hub *Hub) relayMessage(clt *client, msgData *wsMsgRequestData) (*wsMsgData, error) {
	db := db.GetDb()

	sender := clt.user

	var recipient models.User
	result := db.Take(&recipient, &models.User{Username: msgData.Username})
	if result.Error != nil {
		return nil, result.Error
	}

	newMessage, conversation, err := models.CreateMessage(sender, &recipient, msgData.Body)
	if err != nil {
		return nil, err
	}

	newMsgData := &wsMsgData{
		Message: wsMsgMessage{
			Id:             newMessage.ID,
			ConversationId: newMessage.ConversationID,
			SenderId:       newMessage.UserID,
			Body:           newMessage.Body,
			CreatedAt:      newMessage.CreatedAt,
		},
		Conversation: wsMsgConversation{
			Id:        conversation.ID,
			CreatedAt: conversation.CreatedAt,
			ConversationPartner: wsMsgConversationPartner{
				Id:       sender.ID,
				Username: sender.Username,
				Name:     sender.Name,
			},
		},
	}

	hub.clientsMutex.RLock()
	defer hub.clientsMutex.RUnlock()

	newMsgNotification := wsNotification{
		Type:   "notification",
		Method: "newMessage",
		Data:   newMsgData,
	}
	hub.clients[recipient.ID].broadcastNotification(&newMsgNotification)
	hub.clients[sender.ID].broadcastNotificationExceptToSelf(&newMsgNotification, clt)

	return newMsgData, nil
}

func (hub *Hub) AddClient(clt *client) {
	hub.clientsMutex.Lock()
	defer hub.clientsMutex.Unlock()

	_, keyExists := hub.clients[clt.user.ID]
	if !keyExists {
		hub.clients[clt.user.ID] = make(clientGroup)
	}
	hub.clients[clt.user.ID].addClient(clt)

	if gin.IsDebugging() {
		log.Println("added client")
		log.Println(hub.clients)
	}
}

func (hub *Hub) RemoveClient(clt *client) {
	hub.clientsMutex.Lock()
	defer hub.clientsMutex.Unlock()

	hub.clients[clt.user.ID].removeClient(clt)
	if len(hub.clients[clt.user.ID]) == 0 {
		delete(hub.clients, clt.user.ID)
	}

	if gin.IsDebugging() {
		log.Println("removed client")
		log.Println(hub.clients)
	}
}
