package chatServer

import (
	"context"
	"time"

	"github.com/nrmilstein/nchat/app/models"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type client struct {
	hub  *Hub
	user *models.User
	send chan *wsMsgNotification
}

func newClient(hub *Hub, user *models.User) *client {
	return &client{
		hub:  hub,
		user: user,
		send: make(chan *wsMsgNotification),
	}
}

func (clt *client) serveChatMessages(connection *websocket.Conn, ctx context.Context) error {
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
		heartbeat, cancel := context.WithTimeout(ctx, time.Second*30)
		defer cancel()

		select { // TODO: rearange these, put ctx.Done() at the top
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			return err
		case <-heartbeat.Done():
			pingTimeout, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			err := connection.Ping(pingTimeout)
			if err != nil {
				return err
			}
		case msgRequest := <-msgRequests:
			msgSent := clt.hub.relayMessage(clt.user, msgRequest)
			if msgSent != nil {
				wsjson.Write(ctx, connection, msgSent)
			}
		case msgResponse := <-clt.send:
			wsjson.Write(ctx, connection, msgResponse)
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

func (cltGroup clientGroup) broadcastMessage(msg *wsMsgNotification) {
	for clt := range cltGroup {
		clt.send <- msg
	}
}
