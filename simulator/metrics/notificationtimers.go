package metrics

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type notifTimersData struct {
	mu     sync.Mutex
	timers map[string]*prometheus.Timer
}

var notifTimers = notifTimersData{
	timers: make(map[string]*prometheus.Timer),
}

func (n *notifTimersData) add(channelID string) {
	n.mu.Lock()
	n.timers[channelID] = prometheus.NewTimer(notificationDuration)
	n.mu.Unlock()
}

func (n *notifTimersData) del(channelID string) {
	n.mu.Lock()
	delete(n.timers, channelID)
	n.mu.Unlock()
}

func (n *notifTimersData) finish(channelID string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	t, ok := n.timers[channelID]
	if !ok {
		log.Printf("Got notification for unknown ChannelID: %s", channelID)
		return
	}
	t.ObserveDuration()
}
