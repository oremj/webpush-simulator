package pushclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/oremj/webpush-simulator/pushclient/messages"
)

type ConnState int

const (
	UnRegistered ConnState = iota
	Registered
	Closing
)

type readMsg struct {
	messageType int
	data        []byte
	err         error
}

type Conn struct {
	ws *websocket.Conn

	state   ConnState
	Handler Handler

	doneLoop  chan bool
	writeChan chan interface{}
}

func NewConn(ws *websocket.Conn, handler Handler) *Conn {
	conn := &Conn{
		ws:        ws,
		Handler:   handler,
		doneLoop:  make(chan bool),
		writeChan: make(chan interface{}, 100),
	}

	go conn.loop()
	return conn
}

func (c *Conn) loop() {
	defer c.close()
	for {
		select {
		case msg := <-c.writeChan:
			if err := c.ws.WriteJSON(msg); err != nil {
				log.Printf("ws.WriteJSON(%s): %v", msg, err)
				return
			}
		case <-c.doneLoop:
			return
		}
	}
}

func (c *Conn) close() {
	if err := c.ws.Close(); err != nil {
		log.Printf("ws.Close(): %v", err)
	}
}

func (c *Conn) Close() {
	c.state = Closing
	select {
	case <-c.doneLoop:
	default:
		close(c.doneLoop)
	}
}

func (c *Conn) WriteJSON(v interface{}) error {
	select {
	case c.writeChan <- v:
		return nil
	default:
		return errors.New("Dropping msg, write channel full.")
	}
}

func (c *Conn) Hello(req messages.Hello) (messages.HelloResp, error) {
	resp := messages.HelloResp{}
	if err := c.WriteJSON(req); err != nil {
		return resp, fmt.Errorf("WriteJSON(): %v", err)
	}

	if err := c.ws.ReadJSON(&resp); err != nil {
		return resp, fmt.Errorf("ReadJSON(): %v", err)
	}

	return resp, nil
}

func (c *Conn) Register() error {
	c.WriteJSON(messages.NewRegister())
	return nil
}

func (c *Conn) runHandlers(msg []byte) {
	switch messages.MessageType(msg) {
	case messages.TypeRegister:
		resp := messages.RegisterResp{}
		if err := json.Unmarshal(msg, &resp); err != nil {
			log.Printf("Unmarshal(%s): %v", msg, err)
			return
		}
		c.Handler.HandleRegister(c, resp)
	case messages.TypeNotification:
		resp := messages.NotificationResp{}
		if err := json.Unmarshal(msg, &resp); err != nil {
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
	for {
		_, msg, err := c.ws.ReadMessage()
		if c.state == Closing {
			return nil
		}
		if err != nil {
			return fmt.Errorf("ReadMessage(): %v", err)
		}

		c.runHandlers(msg)
	}
	return nil
}
