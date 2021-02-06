package controllers

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nrmilstein/nchat/app/models"
	"github.com/nrmilstein/nchat/chatServer"
	"github.com/nrmilstein/nchat/utils"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func GetChat(hub *chatServer.Hub) func(*gin.Context) {
	return func(c *gin.Context) {
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

		clt := chatServer.NewClient(hub, user)
		hub.AddClient(clt)
		defer hub.RemoveClient(clt)

		err = clt.ServeChatMessages(connection, request.Context())

		log.Println(err)
		connection.Close(websocket.StatusNormalClosure, "")
	}
}

func handleAuthMessage(connection *websocket.Conn, ctx context.Context) (*models.User, error) {
	var authRequest chatServer.WsAuthRequest
	err := wsjson.Read(ctx, connection, &authRequest)
	if err != nil {
		return nil, err
	}

	authKey := authRequest.Data.AuthKey
	user, err := models.GetUserFromKey(authKey)
	if err != nil {
		return nil, err
	}

	authResponse := chatServer.WsAuthSuccessResponse{
		Id:     authRequest.Id,
		Type:   "response",
		Status: "success",
	}

	wsjson.Write(ctx, connection, authResponse)
	return user, nil
}
