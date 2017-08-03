package main

import (
	"flag"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/oremj/webpush-simulator/pushclient"
	"github.com/oremj/webpush-simulator/pushclient/messages"
)

var concurrentConnectionLimit int
var connectionLimit int

var testURL string

func init() {
	flag.IntVar(&concurrentConnectionLimit, "concurrency", 10, "how many concurrent connection attempts")
	flag.IntVar(&connectionLimit, "connections", 100, "how many connections to establish")
	flag.StringVar(&testURL, "url", "wss://autopush.stage.mozaws.net/", "url to test against")
}

func sendNotification(endpoint string) {
	req, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		log.Printf("NewRequest(%s): %s", endpoint, err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Do(%s): %s", endpoint, err)
		return
	}
	defer resp.Body.Close()
}

type conns struct {
	conns    map[string]*pushclient.Conn
	inactive []*pushclient.Conn

	mu sync.Mutex
}

func (c *conns) Add(uaid string, conn *pushclient.Conn) {
	c.mu.Lock()
	c.conns[uaid] = conn
	c.mu.Unlock()
}

func (c *conns) Remove(uaid string) {
	c.mu.Lock()
	if conn, ok := c.conns[uaid]; ok {
		c.inactive = append(c.inactive, conn)
	}
	delete(c.conns, uaid)
	c.mu.Unlock()
}

func (c *conns) PopInactive() *pushclient.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.inactive) == 0 {
		return nil
	}

	item := c.inactive[len(c.inactive)-1]
	c.inactive = c.inactive[:len(c.inactive)-1]
	return item
}

func (c *conns) RandomActive() *pushclient.Conn {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, v := range c.conns {
		return v
	}
	return nil
}

type endpoints struct {
	endpoints []string

	mu sync.Mutex
}

func (e *endpoints) Add(ep string) {
	e.mu.Lock()
	e.endpoints = append(e.endpoints, ep)
	defer e.mu.Unlock()
}

func (e *endpoints) Random() string {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(e.endpoints) < 1 {
		return ""
	}
	return e.endpoints[rand.Intn(len(e.endpoints))]
}

type pushHandler struct {
	connLim chan bool

	concurrentLim chan bool

	conns     *conns
	endpoints *endpoints
}

func (p *pushHandler) waitConnLimits() {
	p.connLim <- true
	p.concurrentLim <- true
}

func (p *pushHandler) signalError() {
	<-p.concurrentLim
	<-p.connLim
}

func (p *pushHandler) signalConnStarted() {
	<-p.concurrentLim
}

func (p *pushHandler) signalConnClosed() {
	<-p.connLim
}

func (p *pushHandler) connKiller() {
	ticker := time.Tick(1 * time.Second)
	for range ticker {
		conn := p.conns.RandomActive()
		if conn == nil {
			continue
		}
		conn.Close()
	}
}

func (p *pushHandler) randomlyRegister() {
	ticker := time.Tick(1 * time.Second)
	for range ticker {
		conn := p.conns.RandomActive()
		if conn == nil {
			continue
		}
		conn.Register()
	}
}

func (p *pushHandler) randomlyNotify() {
	ticker := time.Tick(1 * time.Second)
	for range ticker {
		endpoint := p.endpoints.Random()
		if endpoint == "" {
			continue
		}
		sendNotification(endpoint)
	}
}

func (p *pushHandler) balanceConns() {
	for {
		p.waitConnLimits()

		go func() {
			defer p.signalConnClosed()
			ws, _, err := websocket.DefaultDialer.Dial(testURL, nil)
			if err != nil {
				log.Println("Connecting to websocket: ", err)
				p.signalError()
				return
			}

			client := pushclient.DefaultClientPool.Get()
			defer pushclient.DefaultClientPool.Put(client)

			conn := pushclient.NewConn(ws, p)
			defer conn.Close()

			hello := messages.NewHello(client.UAID, client.Channels().List())

			resp, err := conn.Hello(hello)
			p.signalConnStarted()
			if err != nil {
				log.Printf("Hello(%v): %v", hello, err)
				return
			}

			client.UAID = resp.UAID
			client.Channels().Set(resp.ChannelIDs)

			if err := conn.Loop(); err != nil {
				log.Printf("Loop(): %v", err)
				return
			}
		}()
	}
}

func (p *pushHandler) HandleRegister(conn *pushclient.Conn, msg messages.RegisterResp) {
	p.endpoints.Add(msg.PushEndpoint)
}
func (p *pushHandler) HandleNotification(conn *pushclient.Conn, msg messages.NotificationResp) {
}

func main() {
	flag.Parse()
	handler := &pushHandler{
		connLim:       make(chan bool, connectionLimit),
		concurrentLim: make(chan bool, concurrentConnectionLimit),
		conns: &conns{
			conns: make(map[string]*pushclient.Conn),
		},
		endpoints: &endpoints{
			endpoints: make([]string, 0),
		},
	}

	go handler.connKiller()
	go handler.randomlyRegister()
	go handler.randomlyNotify()
	handler.balanceConns()
}
