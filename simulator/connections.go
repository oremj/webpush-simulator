package simulator

import (
	"sync"

	"github.com/oremj/webpush-simulator/pushclient"
)

type connections struct {
	mu sync.RWMutex

	conns map[*pushclient.Conn]*pushclient.Client
}

func newConnections() connections {
	return connections{
		conns: make(map[*pushclient.Conn]*pushclient.Client),
	}
}

func (c *connections) Add(conn *pushclient.Conn, client *pushclient.Client) {
	c.mu.Lock()
	c.conns[conn] = client
	c.mu.Unlock()
}

func (c *connections) Remove(conn *pushclient.Conn) {
	c.mu.Lock()
	delete(c.conns, conn)
	c.mu.Unlock()
}

func (c *connections) Get(conn *pushclient.Conn) *pushclient.Client {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conns[conn]
}

func (c *connections) GetRandom() (*pushclient.Conn, *pushclient.Client) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for k, v := range c.conns {
		return k, v
	}
	return nil, nil
}
