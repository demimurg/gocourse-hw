package midware

import (
	"encoding/json"
	"sync"

	"github.com/madmaxeatfax/homeworks/Microservices/proto"
)

// easy to add, read and delete channel (uses only keys)
type tunnels map[chan *proto.Event]bool

type Logger struct {
	sync.RWMutex
	Tunnels tunnels
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

	m.log.Tunnels = make(tunnels, 0)

	return &m, nil
}

func (m *middleware) Logger() *Logger {
	return &m.log
}

func (m *middleware) Close() {
	for waiter := range m.log.Tunnels {
		close(waiter)
	}
	m.log.Tunnels = make(tunnels, 0) // to empty
}
