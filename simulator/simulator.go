package simulator

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oremj/webpush-simulator/pushclient"
	"github.com/oremj/webpush-simulator/pushclient/messages"
)

type Simulator struct {
	pushURL        string
	connectionsSem *semaphore
	connectingSem  *semaphore

	connections connections
}

func New(options Options) *Simulator {
	return &Simulator{
		pushURL:        options.PushUrl,
		connectingSem:  newSemaphore(options.Connections),
		connectionsSem: newSemaphore(options.ConcurrentConnections),
		connections:    connections{},
	}
}

func (s *Simulator) connect() (*pushclient.Conn, error) {
	ws, _, err := websocket.DefaultDialer.Dial(s.pushURL, nil)
	if err != nil {
		return nil, err
	}
	conn := pushclient.NewConn(ws)
	return conn, err
}

func (s *Simulator) hello(conn *pushclient.Conn, client *pushclient.Client) error {
	hello := messages.NewHello(client.UAID, client.Channels().List())

	resp, err := conn.Hello(hello)
	if err != nil {
		return fmt.Errorf("conn.hello(%v): %v", hello, err)
	}

	client.UAID = resp.UAID
	client.Channels().Set(resp.ChannelIDs)
	return nil
}

func (s *Simulator) balance() {
	for {
		s.connectionsSem.Acquire()
		s.connectingSem.Acquire()
		go func() {
			defer s.connectionsSem.Release()

			conn, err := s.connect()
			if err != nil {
				log.Printf("connect(): %v", err)

			}

			client := pushclient.DefaultClientPool.Get()
			defer pushclient.DefaultClientPool.Put(client)

			err = s.hello(conn, client)
			s.connectingSem.Release()
			if err != nil {
				log.Println(err)
				return
			}

			s.connections.Add(conn, client)
			defer s.connections.Remove(conn)

			for {
				msg, err, ok := conn.ReadMessage()
				if !ok {
					return
				}
				if err != nil {
					log.Printf("conn.ReadMessage(): %v", err)
					return
				}
				switch val := msg.(type) {
				case messages.RegisterResp:
					client.Channels().Add(val.ChannelID)
				case messages.NotificationResp:
					log.Printf("Notification: %v", val)
				case []byte:
					log.Printf("Unknown message: %v", val)
				}
			}
		}()
	}
}

func (s *Simulator) chaos() {
	killConnectionTicker := time.Tick(1 * time.Second)
	registerTicker := time.Tick(1 * time.Second)
	notifyTicker := time.Tick(1 * time.Second)
	for {
		select {
		case <-killConnectionTicker:
			conn, _ := s.connections.GetRandom()
			conn.Close()
		case <-registerTicker:
			conn, _ := s.connections.GetRandom()
			conn.Register()
		case <-notifyTicker:
		}
	}
}

func (s *Simulator) Run() {
	go s.chaos()
	s.balance()
}
