package midware

import (
	"encoding/json"
	"sync"

	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

type Logger struct {
	sync.RWMutex
	Tunnels []chan *proto.Event
}

type middleware struct {
	acl map[string][]string // map{consumer: [method1, method2...]}
	log Logger
}

func New(ACLdata string) (*middleware, error) {
	m := middleware{}

	m.acl = make(map[string][]string, 0)
	err := json.Unmarshal([]byte(ACLdata), &m.acl)
	if err != nil {
		return nil, err
	}

	m.log.Tunnels = make([]chan *proto.Event, 0)

	return &m, nil
}

func (m *middleware) Logger() *Logger {
	return &m.log
}

func (m *middleware) Close() {
	for _, waiter := range m.log.Tunnels {
		close(waiter)
	}
	m.log.Tunnels = m.log.Tunnels[:0] //empty slice
}
