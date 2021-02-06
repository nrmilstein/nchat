package chatServer

import (
	"context"
	"errors"
	"time"

	"github.com/nrmilstein/nchat/app/models"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var ErrRequestMethodNotFound = errors.New("WebSocket request method not found.")

type client struct {
	hub  *Hub
	user *models.User
	send chan *wsNotification
}

func NewClient(hub *Hub, user *models.User) *client {
	return &client{
		hub:  hub,
		user: user,
		send: make(chan *wsNotification),
	}
}

func (clt *client) ServeChatMessages(connection *websocket.Conn, ctx context.Context) error {
	requests := make(chan *wsRequest)
	errs := make(chan error)

	go func() {
		for ctx.Err() == nil {
			var request wsRequest
			err := wsjson.Read(ctx, connection, &request)
			if err != nil {
				errs <- err
				return
			}
			requests <- &request
		}
	}()

	for {
		heartbeat, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()

		select {
		case err := <-errs:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-heartbeat.Done():
			pingTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			err := connection.Ping(pingTimeout)
			if err != nil {
				return err
			}
		case request := <-requests:
			go func() {
				response, err := clt.handleWsRequest(ctx, request)

				if err == nil {
					wsjson.Write(ctx, connection, *response)
				}
			}()
		case notification := <-clt.send:
			wsjson.Write(ctx, connection, notification)
		}
	}
}

func (clt *client) handleWsRequest(ctx context.Context, request *wsRequest) (*wsSuccessResponse, error) {
	switch request.Method {
	case "sendMessage":
		msgRequestData := &wsMsgRequestData{
			Username: request.Data["username"],
			Body:     request.Data["body"],
		}

		msgResponseData, err := clt.hub.relayMessage(clt, msgRequestData)
		if err != nil {
			return nil, err
		}

		response := &wsSuccessResponse{
			Id:     request.Id,
			Type:   "response",
			Status: "success",
			Data:   msgResponseData,
		}
		return response, nil
	default:
		return nil, ErrRequestMethodNotFound
	}
}

type clientGroup map[*client]bool

func (cltGroup clientGroup) addClient(clt *client) {
	cltGroup[clt] = true
}

func (cltGroup clientGroup) removeClient(clt *client) {
	delete(cltGroup, clt)
}

func (cltGroup clientGroup) broadcastNotification(notification *wsNotification) {
	for clt := range cltGroup {
		clt.send <- notification
	}
}

func (cltGroup clientGroup) broadcastNotificationExceptToSelf(
	notification *wsNotification, self *client) {
	for clt := range cltGroup {
		if clt != self {
			clt.send <- notification
		}
	}
}
