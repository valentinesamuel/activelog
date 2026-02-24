package websocket

import (
	"encoding/json"
	"log"
	"time"

	gorillaws "github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 54 * time.Second // must be less than pongWait
	maxMessageSize = 512
)

// Client is a middleman between the WebSocket connection and the hub
type Client struct {
	hub    *Hub
	conn   *gorillaws.Conn
	userID int
	send   chan Message
}

// newClient creates a new Client
func newClient(hub *Hub, conn *gorillaws.Conn, userID int) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		userID: userID,
		send:   make(chan Message, 256),
	}
}

// readPump pumps messages from the WebSocket connection to the hub.
// The application runs readPump in a goroutine per connection.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if gorillaws.IsUnexpectedCloseError(err, gorillaws.CloseGoingAway, gorillaws.CloseAbnormalClosure) {
				log.Printf("websocket read error for user %d: %v", c.userID, err)
			}
			break
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
// The application runs writePump in a goroutine per connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(gorillaws.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(gorillaws.TextMessage)
			if err != nil {
				return
			}

			if err := json.NewEncoder(w).Encode(msg); err != nil {
				log.Printf("websocket write error for user %d: %v", c.userID, err)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(gorillaws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
