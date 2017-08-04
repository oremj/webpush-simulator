package simulator

import (
	"sync"

	"github.com/oremj/webpush-simulator/endpoint"
)

type endpoints struct {
	mu sync.RWMutex

	eps map[string]endpoint.Endpoint
}

func newEndpoints() endpoints {
	return endpoints{
		eps: make(map[string]endpoint.Endpoint),
	}
}

func (e *endpoints) Add(chanID string, ep endpoint.Endpoint) {
	e.mu.Lock()
	e.eps[chanID] = ep
	e.mu.Unlock()
}

func (e *endpoints) Remove(chanID string) {
	e.mu.Lock()
	delete(e.eps, chanID)
	e.mu.Unlock()
}

func (e *endpoints) GetRandom() (chanID string, ep endpoint.Endpoint, ok bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for k, v := range e.eps {
		return k, v, true
	}
	return
}
