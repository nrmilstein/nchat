package chatServer

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/db"
	"github.com/nrmilstein/nchat/utils"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type wSAuthMessage struct {
	AuthKey string `json:"authKey"`
}

type wsMsgRequest struct {
	Email string `json:"email"`
	Body  string `json:"body"`
}

type wsMsgResponse struct {
	Id             int       `json:"id"`
	ConversationId int       `json:"conversationId"`
	SenderId       int       `json:"senderId"`
	Body           string    `json:"body"`
	Created        time.Time `json:"sent"`
}

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
	acceptOptions := &websocket.AcceptOptions{
		Subprotocols:   []string{"nchat"},
		OriginPatterns: []string{"localhost:3000"},
	}
	connection, err := websocket.Accept(writer, request, acceptOptions)
	if err != nil {
		utils.AbortErrServer(c)
		return
	}
	defer connection.Close(websocket.StatusInternalError, "Internal server error.")

	authKey, err := readAuthMessage(connection, request.Context())
	if err != nil {
		connection.Close(4003, "Authorization failed.")
		return
	}
	user, err := models.GetUserFromKey(authKey)
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

func readAuthMessage(connection *websocket.Conn, ctx context.Context) (string, error) {
	var authMessage wSAuthMessage
	err := wsjson.Read(ctx, connection, &authMessage)
	if err != nil {
		return "", err
	}
	return authMessage.AuthKey, err
}

func (hub *Hub) relayMessage(sender *models.User, msgRequest *wsMsgRequest) *wsMsgResponse {
	db := db.GetDb()

	var recipient models.User
	result := db.Take(&recipient, &models.User{Email: msgRequest.Email})
	if result.Error != nil {
		return nil
	}

	newMessage, err := models.CreateMessage(sender, &recipient, msgRequest.Body)
	if err != nil {
		return nil
	}

	msgResponse := &wsMsgResponse{
		Id:             newMessage.ID,
		ConversationId: newMessage.ConversationID,
		SenderId:       newMessage.UserID,
		Body:           newMessage.Body,
		Created:        newMessage.CreatedAt,
	}

	hub.clientsMutex.RLock()
	defer hub.clientsMutex.RUnlock()

	hub.clients[recipient.ID].broadcastMessage(msgResponse)
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
	log.Println("added client")
	log.Println(hub.clients)
}

func (hub *Hub) removeClient(clt *client) {
	hub.clientsMutex.Lock()
	defer hub.clientsMutex.Unlock()

	hub.clients[clt.user.ID].removeClient(clt)
	if len(hub.clients[clt.user.ID]) == 0 {
		delete(hub.clients, clt.user.ID)
	}
	log.Println("removed client")
	log.Println(hub.clients)
}
