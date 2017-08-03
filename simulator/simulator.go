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
	concurrentSem  *semaphore

	connections connections
}

func New(options Options) *Simulator {
	log.Printf("Simulating: %d connections %d concurrent", options.Connections, options.ConcurrentConnections)
	return &Simulator{
		pushURL:        options.PushUrl,
		concurrentSem:  newSemaphore(options.ConcurrentConnections),
		connectionsSem: newSemaphore(options.Connections),
		connections:    connections{},
	}
}

func (s *Simulator) connect(client *pushclient.Client) (*pushclient.Conn, error) {
	ws, _, err := websocket.DefaultDialer.Dial(s.pushURL, nil)
	if err != nil {
		return nil, err
	}
	conn := pushclient.NewConn(ws)

	hello := messages.NewHello(client.UAID, client.Channels().List())

	resp, err := conn.Hello(hello)
	if err != nil {
		return nil, fmt.Errorf("conn.hello(%v): %v", hello, err)
	}

	client.UAID = resp.UAID
	client.Channels().Set(resp.ChannelIDs)
	return conn, nil
}

func (s *Simulator) balance() {
	for {
		s.connectionsSem.Acquire()
		s.concurrentSem.Acquire()
		go func() {
			defer s.connectionsSem.Release()

			client := pushclient.DefaultClientPool.Get()
			defer pushclient.DefaultClientPool.Put(client)

			conn, err := s.connect(client)
			s.concurrentSem.Release()
			if err != nil {
				log.Printf("connect(): %v", err)
				return

			}
			log.Printf("Connections: %d In progress: %d", s.connectionsSem.Count()-s.concurrentSem.Count(), s.concurrentSem.Count())

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
					log.Printf("Registration Response: %v", val)
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
	killConnectionTicker := time.Tick(3 * time.Second)
	registerTicker := time.Tick(10 * time.Second)
	notifyTicker := time.Tick(30 * time.Second)
	for {
		select {
		case <-killConnectionTicker:
			log.Println("Killing a connection")
			conn, _ := s.connections.GetRandom()
			conn.Close()
		case <-registerTicker:
			log.Println("Registering a channel")
			conn, _ := s.connections.GetRandom()
			conn.Register()
		case <-notifyTicker:
			log.Println("Notifying a channel")
		}
	}
}

func (s *Simulator) Run() {
	go s.chaos()
	s.balance()
}
