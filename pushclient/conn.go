package pushclient

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/oremj/webpush-simulator/pushclient/messages"
)

type ConnState int

const (
	UnRegistered ConnState = iota
	Registered
	Closing
)

type Handler interface {
	HandleHello(*Conn, *messages.HelloResp)
	HandleRegister(*Conn, *messages.RegisterResp)
	HandleNotification(*Conn, *messages.NotificationResp)
}

type Conn struct {
	ws *websocket.Conn

	state   ConnState
	Handler Handler

	UAID       string
	ChannelIDs []string

	mu      sync.Mutex
	closeMu sync.Mutex
}

func NewConn(ws *websocket.Conn, handler Handler) *Conn {
	return &Conn{
		ws:      ws,
		Handler: handler,
	}
}

func (c *Conn) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ws.WriteJSON(v)
}

func (c *Conn) Hello(req *messages.Hello) error {
	if err := c.WriteJSON(req); err != nil {
		return fmt.Errorf("WriteJSON(): %v", err)
	}

	resp := new(messages.HelloResp)
	if err := c.ws.ReadJSON(&resp); err != nil {
		return fmt.Errorf("ReadJSON(): %v", err)
	}

	c.UAID = resp.UAID
	c.ChannelIDs = resp.ChannelIDs
	c.Handler.HandleHello(c, resp)

	c.state = Registered
	return nil
}

func (c *Conn) Register() error {
	c.WriteJSON(messages.NewRegister())
	return nil
}

func (c *Conn) Close() error {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()
	c.state = Closing
	return c.ws.Close()
}

func (c *Conn) runHandlers(msg []byte) {
	switch messages.MessageType(msg) {
	case messages.TypeRegister:
		resp := new(messages.RegisterResp)
		if err := json.Unmarshal(msg, resp); err != nil {
			log.Printf("Unmarshal(%s): %v", msg, err)
			return
		}
		c.ChannelIDs = append(c.ChannelIDs, resp.ChannelID)
		c.Handler.HandleRegister(c, resp)
	case messages.TypeNotification:
		resp := new(messages.NotificationResp)
		if err := json.Unmarshal(msg, resp); err != nil {
			log.Printf("Unmarshal(%s): %v", msg, err)
			return
		}

		ack := messages.NewAck()
		ack.Add(resp.ChannelID, resp.Version)
		c.WriteJSON(ack)

		c.Handler.HandleNotification(c, resp)
	case messages.TypeHello:
		log.Printf("Unexpected hello: %s", msg)
	default:
		log.Printf("Unknown message: %s", msg)
	}
}

func (c *Conn) Loop() error {
	if c.state == UnRegistered {
		if err := c.Hello(messages.NewHello()); err != nil {
			return fmt.Errorf("Hello(): %v", err)
		}
	}

	for {
		_, msg, err := c.ws.ReadMessage()
		if c.state == Closing {
			return nil
		}
		if err != nil {
			return fmt.Errorf("ReadMessage(): %v", err)
		}

		c.closeMu.Lock()
		c.runHandlers(msg)
		c.closeMu.Unlock()
	}
	return nil
}
