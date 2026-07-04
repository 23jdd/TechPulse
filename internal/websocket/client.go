package websocket

import (
	"net/http"
	"time"

	gws "github.com/gorilla/websocket"
)

var upgrader = gws.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

type Client struct {
	hub  *Hub
	conn *gws.Conn
	send chan Event
}

func Serve(hub *Hub, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	client := &Client{hub: hub, conn: conn, send: make(chan Event, 16)}
	hub.register <- client
	go client.writePump()
	go client.readPump()
	return nil
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()
	for {
		if _, _, err := c.conn.NextReader(); err != nil {
			return
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case event, ok := <-c.send:
			if !ok {
				_ = c.conn.WriteMessage(gws.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteJSON(event); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteJSON(Event{Type: "worker_status", Message: "alive", Time: time.Now()}); err != nil {
				return
			}
		}
	}
}
