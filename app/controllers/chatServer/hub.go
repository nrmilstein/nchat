package chatServer

import (
	"context"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
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

func (hub *Hub) GetChat(c *gin.Context) {
	writer := c.Writer
	request := c.Request

	originPatterns := []string{}
	if gin.IsDebugging() {
		originPatterns = []string{"localhost:3000"}
	}

	acceptOptions := &websocket.AcceptOptions{
		Subprotocols:   []string{"nchat"},
		OriginPatterns: originPatterns,
	}

	connection, err := websocket.Accept(writer, request, acceptOptions)
	if err != nil {
		utils.AbortErrServer(c)
		return
	}
	defer connection.Close(websocket.StatusInternalError, "Internal server error.")

	user, err := handleAuthMessage(connection, request.Context())
	if err != nil {
		connection.Close(4003, "Authorization failed.")
		return
	}

	clt := newClient(hub, user)
	hub.addClient(clt)
	defer hub.removeClient(clt)

	err = clt.serveChatMessages(connection, request.Context())

	log.Println(err)
	connection.Close(websocket.StatusNormalClosure, "")
}

func handleAuthMessage(connection *websocket.Conn, ctx context.Context) (*models.User, error) {
	var authRequest wsAuthRequest
	err := wsjson.Read(ctx, connection, &authRequest)
	if err != nil {
		return nil, err
	}

	authKey := authRequest.Data.AuthKey
	user, err := models.GetUserFromKey(authKey)
	if err != nil {
		return nil, err
	}

	authResponse := wsAuthSuccessResponse{
		Id:     authRequest.Id,
		Type:   "response",
		Status: "success",
	}

	wsjson.Write(ctx, connection, authResponse)
	return user, nil
}

func (hub *Hub) relayMessage(sender *models.User, msgRequest *wsMsgRequest) *wsMsgSuccessResponse {
	db := db.GetDb()

	var recipient models.User
	result := db.Take(&recipient, &models.User{Username: msgRequest.Data.Username})
	if result.Error != nil {
		return nil
	}

	newMessage, conversation, err := models.CreateMessage(sender, &recipient, msgRequest.Data.Body)
	if err != nil {
		return nil
	}

	msgNotificaiton := &wsMsgNotification{
		Type:   "notification",
		Method: "newMessage",
		Data: wsMsgNotificationData{
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
		},
	}

	hub.clientsMutex.RLock()
	defer hub.clientsMutex.RUnlock()

	hub.clients[recipient.ID].broadcastMessage(msgNotificaiton)

	msgResponse := &wsMsgSuccessResponse{
		Id:     msgRequest.Id,
		Type:   "response",
		Status: "success",
		Data: wsMsgSuccessResponseData{
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
					Id:       recipient.ID,
					Username: recipient.Username,
					Name:     recipient.Name,
				},
			},
		},
	}
	return msgResponse
}

func (hub *Hub) addClient(clt *client) {
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

func (hub *Hub) removeClient(clt *client) {
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
