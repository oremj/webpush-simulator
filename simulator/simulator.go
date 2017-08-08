package simulator

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oremj/webpush-simulator/endpoint"
	"github.com/oremj/webpush-simulator/pushclient"
	"github.com/oremj/webpush-simulator/pushclient/messages"
	"github.com/oremj/webpush-simulator/simulator/metrics"
)

type Simulator struct {
	pushURL        string
	connectionsSem *semaphore
	concurrentSem  *semaphore

	connections connections
	endpoints   endpoints
}

func New(options Options) *Simulator {
	log.Printf("Simulating: %d connections %d concurrent", options.Connections, options.ConcurrentConnections)
	return &Simulator{
		pushURL:        options.PushUrl,
		concurrentSem:  newSemaphore(options.ConcurrentConnections),
		connectionsSem: newSemaphore(options.Connections),
		connections:    newConnections(),
		endpoints:      newEndpoints(),
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
				metrics.Error("connect")
				return
			}
			log.Printf("Connections: %d In progress: %d", s.connectionsSem.Count()-s.concurrentSem.Count(), s.concurrentSem.Count())

			s.connections.Add(conn, client)
			metrics.ConnectionStarted()
			defer func() {
				s.connections.Remove(conn)
				metrics.ConnectionEnded()
			}()

			for {
				msg, err, ok := conn.ReadMessage()
				if !ok {
					return
				}
				if err != nil {
					log.Printf("conn.ReadMessage(): %v", err)
					metrics.Error("read_message")
					return
				}
				switch val := msg.(type) {
				case messages.RegisterResp:
					log.Printf("Registration Response: %v", val)
					metrics.RegistrationRecv.Inc()
					client.Channels().Add(val.ChannelID)
					ep := endpoint.New(
						val.PushEndpoint,
						client.Keys().PubKey(),
						client.Keys().AuthKey())
					s.endpoints.Add(val.ChannelID, ep)
				case messages.NotificationResp:
					log.Printf("Notification: %v", val)
					metrics.NotificationRecv(val.ChannelID)
				case []byte:
					log.Printf("Unknown message: %v", val)
					metrics.Error("unknown_msg")
				}
			}
		}()
	}
}

func (s *Simulator) chaos() {
	killConnectionTicker := time.Tick(3 * time.Second)
	registerTicker := time.Tick(1 * time.Second)
	notifyTicker := time.Tick(3 * time.Second)
	for {
		select {
		case <-killConnectionTicker:
			log.Println("Killing a connection")
			conn, _ := s.connections.GetRandom()
			conn.Close()
		case <-registerTicker:
			log.Println("Registering a channel")
			conn, _ := s.connections.GetRandom()
			if conn == nil {
				continue
			}
			conn.Register()
			metrics.RegistrationSent.Inc()
		case <-notifyTicker:
			channelID, ep, ok := s.endpoints.GetRandom()
			if !ok {
				continue
			}
			log.Printf("Notifying a channel: %v", ep)
			metrics.NotificationSent(channelID)
			if err := ep.Notify("test"); err != nil {
				log.Printf("Notify(): %v", err)
				metrics.NotificationCancel(channelID)
				continue
			}
		}
	}
}

func (s *Simulator) Run() {
	go s.chaos()
	s.balance()
}
