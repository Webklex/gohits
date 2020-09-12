package server

import (
	"../utils/log"
	"github.com/gorilla/websocket"
	"time"
)

// Client is a middleman between the websocket connection and the node.
type Client struct {
	Server *Server

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan *Message

	received chan *Message
}

func NewClient(server *Server, conn *websocket.Conn) *Client {
	c := &Client{
		Server: server,
		Conn:   conn,
		send:   make(chan *Message, 256),
	}
	return c
}

func (c *Client) NewMessage() *Message {
	return NewMessage(c, []byte(""))
}

func (c *Client) Send(message *Message) {
	message.Client = c
	select {
	case <-c.send:
		return
	default:
	}

	c.send <- message
	return
}

func (c *Client) SendBytes(payload []byte) {
	message := NewMessage(c, payload)
	c.Send(message)
	return
}

func (c *Client) SendString(payload string) {
	c.SendBytes([]byte(payload))
	return
}

func (c *Client) GetSend() chan *Message {
	return c.send
}

func (c *Client) IsClosed(ch <-chan *Message) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

func (c *Client) Close() error {
	if !c.IsClosed(c.send) && c.send != nil {
		close(c.send)
	}
	if !c.IsClosed(c.received) && c.received != nil {
		close(c.received)
	}

	return c.Conn.Close()
}

func (c *Client) Listen() {
	// CONN -----> NODE
	go c.writePump()

	// NODE <----- CONN
	go c.readPump()
}

// readPump pumps messages from the websocket connection to the node.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.Server.unregister <- c
		_ = c.Close()
	}()
	c.Conn.SetReadLimit(c.Server.Config.MaxMessageSize)

	_ = c.Conn.SetReadDeadline(time.Now().Add(c.Server.Config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		if err := c.Conn.SetReadDeadline(time.Now().Add(c.Server.Config.PongWait)); err != nil {
			log.Error("error: ", err)
			_ = c.Close()
		}
		return nil
	})

	for {
		_, text, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("error: ", err)
				return
			}
			break
		}

		message := NewMessage(c, text)
		c.Server.messageHandler(message)
	}
}

// writePump pumps messages from the node to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(c.Server.Config.PingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(c.Server.Config.WriteWait))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				log.Error("Socket closed")
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message.Payload); err != nil {
				log.Error("Socket error: ", err)
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(c.Server.Config.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Error("Socket error: ", err)
				return
			}
		}
	}
}
