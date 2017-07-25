package pushclient

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/oremj/webpush-simulator/pushclient/messages"
)

type ConnState int

const (
	UnRegistered ConnState = iota
	Registered
)

type Handler interface {
	HandleHello(*Conn, *messages.HelloResp)
	HandleRegister(*Conn, *messages.RegisterResp)
	HandleNotification(*Conn, *messages.RegisterResp)
}

type Conn struct {
	ws *websocket.Conn

	state   ConnState
	Handler Handler

	UAID string
}

func NewConn(ws *websocket.Conn, handler Handler) *Conn {
	return &Conn{
		ws:      ws,
		Handler: handler,
	}
}

func (c *Conn) Hello(req *messages.Hello) error {
	if err := c.ws.WriteJSON(req); err != nil {
		return fmt.Errorf("WriteJSON(): %v", err)
	}

	resp := new(messages.HelloResp)
	if err := c.ws.ReadJSON(&resp); err != nil {
		return fmt.Errorf("ReadJSON(): %v", err)
	}

	c.UAID = resp.UAID
	c.Handler.HandleHello(c, resp)

	c.state = Registered
	return nil
}

func (c *Conn) Register() error {
	c.ws.WriteJSON(messages.NewRegister())
	return nil
}

func (c *Conn) Loop() error {
	if c.state == UnRegistered {
		if err := c.Hello(messages.NewHello()); err != nil {
			return fmt.Errorf("Hello(): %v", err)
		}
	}

	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			return fmt.Errorf("ReadMessage(): %v", err)
		}
		switch messages.MessageType(msg) {
		case messages.TypeRegister:
			resp := new(messages.RegisterResp)
			if err := json.Unmarshal(msg, resp); err != nil {
				log.Printf("Unmarshal(%s): %v", msg, err)
				continue
			}
			c.Handler.HandleRegister(c, resp)
		case messages.TypeHello:
			log.Printf("Unexpected hello: %s", msg)
		case messages.TypeNotification:
			fmt.Println("Notification")
		default:
			log.Printf("Unknown message: %s", msg)
		}
	}
	return nil
}
