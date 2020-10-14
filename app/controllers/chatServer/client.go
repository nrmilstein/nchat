package chatServer

import (
	"context"

	"github.com/nrmilstein/nchat/app/models"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type client struct {
	hub  *Hub
	user *models.User
	send chan *wsMsgResponse
}

func newClient(hub *Hub, user *models.User) *client {
	return &client{
		hub:  hub,
		user: user,
		send: make(chan *wsMsgResponse),
	}
}

func (clt *client) serveChatMessages(connection *websocket.Conn, ctx context.Context) error {
	// ctx, cancel := context.WithCancel(ctx)

	msgRequests := make(chan *wsMsgRequest)
	errs := make(chan error)

	go func() {
		for ctx.Err() == nil {
			var msgRequest wsMsgRequest
			// TODO: do I need to use a mutex to prevent reading and writing to the connection at the same
			//  time?
			err := wsjson.Read(ctx, connection, &msgRequest)
			if err != nil {
				errs <- err
				return
			}
			msgRequests <- &msgRequest
		}
	}()

	for {
		select {
		case msgRequest := <-msgRequests:
			msgSent := clt.hub.relayMessage(clt.user, msgRequest)
			if msgSent != nil {
				wsjson.Write(ctx, connection, msgSent) // TODO: implement proper ACKs
			}
		case msgResponse := <-clt.send:
			wsjson.Write(ctx, connection, msgResponse)
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

func (cltGroup clientGroup) broadcastMessage(msg *wsMsgResponse) {
	for clt := range cltGroup {
		clt.send <- msg
	}
}
