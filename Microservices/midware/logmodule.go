package midware

import (
	"sync"
)

type data struct {
	Consumer, Method string
}

// easy to add, delete and iterate on channels
// (uses only keys)
type tunnels map[chan data]bool

type logmod struct {
	sync.RWMutex
	tunnels tunnels
}

func (l *logmod) share(consumer, method string) error {
	logData := data{
		Consumer: consumer,
		Method:   method,
	}

	l.RLock()
	for waiter := range l.tunnels {
		waiter <- logData
	}
	l.RUnlock()

	return nil
}

type Logger interface {
	NewTunnel() chan data
	DeleteTunnel(chan data)
}

func (l *logmod) NewTunnel() chan data {
	ch := make(chan data)

	l.Lock()
	l.tunnels[ch] = true
	l.Unlock()

	return ch
}

func (l *logmod) DeleteTunnel(ch chan data) {
	l.Lock()
	delete(l.tunnels, ch)
	l.Unlock()
}
