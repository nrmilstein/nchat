package controllers

import (
	"context"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/utils"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type wSAuthMessage struct {
	AuthKey string `json:"authKey"`
}

type wSMessage struct {
	SenderId   int    `json:"senderId"`
	ReceiverId int    `json:"receiverId"`
	Body       string `json:"body"`
}

type wSMessageFromClient struct {
	ReceiverId int    `json:"receiverId"`
	Body       string `json:"body"`
}

type wSMessageToClient struct {
	SenderId int    `json:"senderId"`
	Body     string `json:"body"`
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
	defer connection.Close(websocket.StatusInternalError, "The sky is falling.")

	user, err := readAuthMessage(connection, request.Context())
	if err != nil {
		connection.Close(4003, "Authorization failed.")
		return
	}

	clt := &client{
		hub:          hub,
		userId:       user.ID,
		msgsToClient: make(chan *wSMessageToClient),
	}
	hub.addClient(clt)
	defer hub.removeClient(clt)

	err = clt.serveChatMessages(connection, request.Context())

	log.Println(err)
	connection.Close(websocket.StatusNormalClosure, "")
}

func readAuthMessage(connection *websocket.Conn, ctx context.Context) (*models.User, error) {
	var authMessage wSAuthMessage
	err := wsjson.Read(ctx, connection, &authMessage)
	if err != nil {
		return nil, err
	}
	authKey := authMessage.AuthKey

	user, err := models.GetUserFromKey(authKey)
	return user, err
}

func (hub *Hub) addClient(clt *client) {
	hub.clientsMutex.Lock()
	defer hub.clientsMutex.Unlock()

	_, keyExists := hub.clients[clt.userId]
	if !keyExists {
		hub.clients[clt.userId] = make(clientGroup)
	}
	hub.clients[clt.userId].addClient(clt)
	log.Println("added client")
	log.Println(hub.clients)
}

func (hub *Hub) removeClient(clt *client) {
	hub.clientsMutex.Lock()
	defer hub.clientsMutex.Unlock()

	hub.clients[clt.userId].removeClient(clt)
	if len(hub.clients[clt.userId]) == 0 {
		delete(hub.clients, clt.userId)
	}
	log.Println("removed client")
	log.Println(hub.clients)
}

func (hub *Hub) relayMessage(msg *wSMessage) {
	if msg.ReceiverId == msg.SenderId {
		return
	}
	hub.clientsMutex.RLock()
	defer hub.clientsMutex.RUnlock()

	msgToClient := &wSMessageToClient{
		SenderId: msg.SenderId,
		Body:     msg.Body,
	}

	hub.clients[msg.ReceiverId].broadcastMessage(msgToClient)
	log.Println("Sent message")
	log.Println(msg)
}

type client struct {
	hub          *Hub
	userId       int
	msgsToClient chan *wSMessageToClient
}

func (clt *client) serveChatMessages(connection *websocket.Conn, ctx context.Context) error {
	// ctx, cancel := context.WithCancel(ctx)

	msgsFromClient := make(chan *wSMessageFromClient)
	errs := make(chan error)
	go func() {
		for ctx.Err() == nil {
			var msgFromClient wSMessageFromClient
			err := wsjson.Read(ctx, connection, &msgFromClient)
			if err != nil {
				errs <- err
				return
			}
			msgsFromClient <- &msgFromClient
		}
	}()

	for {
		select {
		case msgFromClient := <-msgsFromClient:
			go clt.hub.relayMessage(&wSMessage{
				SenderId:   clt.userId,
				ReceiverId: msgFromClient.ReceiverId,
				Body:       msgFromClient.Body,
			})
		case msgToClient := <-clt.msgsToClient:
			wsjson.Write(ctx, connection, msgToClient)
		case err := <-errs:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type clientGroup map[*client]bool

func (cltGroup clientGroup) addClient(clt *client) {
	cltGroup[clt] = true
}

func (cltGroup clientGroup) removeClient(clt *client) {
	delete(cltGroup, clt)
}

func (cltGroup clientGroup) broadcastMessage(msg *wSMessageToClient) {
	for clt := range cltGroup {
		clt.msgsToClient <- msg
	}
}
