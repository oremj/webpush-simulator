package pushclient

import "sync"

var DefaultClientPool = NewClientPool()

type ClientPool struct {
	pool sync.Pool
}

func NewClientPool() *ClientPool {
	pool := sync.Pool{
		New: func() interface{} {
			return New("")
		},
	}
	return &ClientPool{
		pool: pool,
	}
}

func (c *ClientPool) Get() *Client {
	return c.pool.Get().(*Client)
}

func (c *ClientPool) Put(client *Client) {
	c.pool.Put(client)
}
